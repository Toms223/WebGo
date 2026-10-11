package webref

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"
)

func loadFixture(t *testing.T) *Data {
	t.Helper()
	data, err := Load(filepath.Join("..", "testdata", "webref"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return data
}

func loadJSON(t *testing.T, content string) (*Data, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "css.json"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(dir)
}

func TestLoadWebrefMergesValues(t *testing.T) {
	data := loadFixture(t)

	wantValues := map[string][]string{
		"display":            {"[ <display-outside> || <display-inside> ] | list-item | contents | none"},
		"width":              {"auto | <box-size> | <anchor-size()>"},
		"padding":            {"<'padding-top'>{1,4}"},
		"border-width":       {"<line-width>{1,4}"},
		"font-family":        {"[ <family-name> | <generic-family> ]#"},
		"-webkit-line-clamp": {"none | <integer [1,∞]>"},
	}
	for name, want := range wantValues {
		property, ok := data.Properties[name]
		if !ok {
			t.Errorf("Properties[%q] missing", name)
			continue
		}
		if property.Name != name {
			t.Errorf("Properties[%q].Name = %q", name, property.Name)
		}
		if !slices.Equal(property.Values, want) {
			t.Errorf("Properties[%q].Values = %q, want %q", name, property.Values, want)
		}
	}

	if got := data.Features["orientation"]; got != "portrait | landscape" {
		t.Errorf(`Features["orientation"] = %q, want "portrait | landscape"`, got)
	}

	for _, selector := range []string{":hover", "::before", ":nth-child()", ":not()", ":default", ":dir()", ":lang()", ":popover-open"} {
		if !slices.Contains(data.Selectors, selector) {
			t.Errorf("Selectors = %q, missing %q", data.Selectors, selector)
		}
	}

	if want := []string{"red", "rebeccapurple"}; !slices.Equal(data.NamedColors, want) {
		t.Errorf("NamedColors = %q, want %q", data.NamedColors, want)
	}

	wantDefinitions := map[string]string{
		"line-width":      "<length [0,∞]> | thin | medium | thick",
		"display-outside": "block | inline | run-in",
		"display-inside":  "flow | flow-root | table | flex | grid | ruby",
		"box-size":        "<length-percentage [0,∞]> | min-content | max-content | fit-content(<length-percentage [0,∞]>) | stretch",
	}
	for name, want := range wantDefinitions {
		got, ok := data.Values[name]
		if !ok {
			t.Errorf("Values[%q] missing", name)
			continue
		}
		if got != want {
			t.Errorf("Values[%q] = %q, want %q", name, got, want)
		}
	}
}

func TestLoadSkipsPropertiesWithoutSyntax(t *testing.T) {
	data := loadFixture(t)
	if _, ok := data.Properties["-webkit-box-align"]; ok {
		t.Error(`Properties["-webkit-box-align"] present, but the property has no syntax`)
	}
	if _, ok := data.Properties["legacy-thing"]; !ok {
		t.Error(`Properties["legacy-thing"] missing`)
	}
}

func TestLoadFeaturesComeOnlyFromMedia(t *testing.T) {
	data := loadFixture(t)
	for name, want := range map[string]string{
		"width":                "<length>",
		"hover":                "none | hover",
		"prefers-color-scheme": "light | dark",
		"aspect-ratio":         "<ratio>",
	} {
		if got := data.Features[name]; got != want {
			t.Errorf("Features[%q] = %q, want %q", name, got, want)
		}
	}
	if got, ok := data.Features["inline-size"]; ok {
		t.Errorf(`Features["inline-size"] = %q, but only @media descriptors are features`, got)
	}
}

func TestLoadCollectsAtRuleDescriptors(t *testing.T) {
	data := loadFixture(t)
	for rule, want := range map[string]map[string]string{
		"font-face": {"src": "<font-src-list>", "font-display": "auto | block | swap | fallback | optional"},
		"property":  {"syntax": "<string>", "inherits": "true | false"},
		"container": {"width": "<length>", "inline-size": "<length>", "orientation": "portrait | landscape"},
		"media":     {"hover": "none | hover"},
	} {
		for name, syntax := range want {
			if got := data.Descriptors[rule][name]; got != syntax {
				t.Errorf("Descriptors[%q][%q] = %q, want %q", rule, name, got, syntax)
			}
		}
	}
	if !maps.Equal(data.Descriptors["media"], data.Features) {
		t.Errorf("Descriptors[media] = %q, want Features %q", data.Descriptors["media"], data.Features)
	}
	for _, rule := range []string{"keyframes", "layer", "supports", "import"} {
		if got := data.Descriptors[rule]; len(got) != 0 {
			t.Errorf("Descriptors[%q] = %q, want none", rule, got)
		}
	}
}

func TestLoadDescriptorsNeedANameAndASyntax(t *testing.T) {
	data, err := loadJSON(t, `{"atrules": [
		{"name": "@font-face", "descriptors": [
			{"name": "bare"},
			{"syntax": "auto | swap"},
			{"name": "font-display", "syntax": "auto | swap"}
		]}
	]}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]map[string]string{"font-face": {"font-display": "auto | swap"}}
	if !reflect.DeepEqual(data.Descriptors, want) {
		t.Errorf("Descriptors = %q, want %q", data.Descriptors, want)
	}
}

func TestLoadFeaturesNeedASyntax(t *testing.T) {
	data, err := loadJSON(t, `{"atrules": [
		{"name": "@media", "descriptors": [
			{"name": "bare"},
			{"syntax": "none | hover"},
			{"name": "hover", "syntax": "none | hover"}
		]}
	]}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]string{"hover": "none | hover"}
	if !maps.Equal(data.Features, want) {
		t.Errorf("Features = %q, want %q", data.Features, want)
	}
}

func TestLoadValuesCollectTypesAndFunctions(t *testing.T) {
	data, err := loadJSON(t, `{
		"types": [
			{"name": "line-width", "syntax": "<length [0,∞]> | thin | medium | thick"},
			{"name": "angle"},
			{"syntax": "thin | thick"},
			{"name": "named-color", "syntax": "red | blue"}
		],
		"functions": [
			{"name": "anchor-size()", "syntax": "anchor-size( [ <anchor-name> || <anchor-size> ]? , <length-percentage>? )"},
			{"name": "cross-origin()"}
		],
		"properties": [{"name": "width", "syntax": "auto"}],
		"selectors": [{"name": ":hover", "syntax": ":hover"}],
		"atrules": [{"name": "@media", "descriptors": [{"name": "hover", "syntax": "none | hover"}]}]
	}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]string{
		"line-width":    "<length [0,∞]> | thin | medium | thick",
		"named-color":   "red | blue",
		"anchor-size()": "anchor-size( [ <anchor-name> || <anchor-size> ]? , <length-percentage>? )",
	}
	if !maps.Equal(data.Values, want) {
		t.Errorf("Values = %q, want %q", data.Values, want)
	}
}

func TestLoadValuesKeepTheFirstDefinition(t *testing.T) {
	data, err := loadJSON(t, `{
		"types": [
			{"name": "content-list", "syntax": "[ <string> ]+"},
			{"name": "content-list", "for": ["content"], "syntax": "[ <image> ]+"},
			{"name": "pending"},
			{"name": "pending", "syntax": "thin | thick"}
		],
		"functions": [
			{"name": "scale()", "syntax": "scale( <number> )"},
			{"name": "scale()", "for": ["transform"], "syntax": "scale( <number> , <number>? )"}
		]
	}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := map[string]string{
		"content-list": "[ <string> ]+",
		"pending":      "thin | thick",
		"scale()":      "scale( <number> )",
	}
	if !maps.Equal(data.Values, want) {
		t.Errorf("Values = %q, want %q", data.Values, want)
	}
}

func TestLoadSortsAndDeduplicatesSelectors(t *testing.T) {
	data, err := loadJSON(t, `{"selectors": [
		{"name": ":hover"},
		{"name": "::before"},
		{"name": ":active"},
		{"name": ":hover"},
		{"syntax": ":nameless"},
		{"name": ">"}
	]}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"::before", ":active", ":hover", ">"}
	if !slices.Equal(data.Selectors, want) {
		t.Errorf("Selectors = %q, want %q", data.Selectors, want)
	}
}

func TestLoadAccumulatesRepeatedPropertyEntries(t *testing.T) {
	data, err := loadJSON(t, `{"properties": [
		{"name": "display", "syntax": "block | inline"},
		{"name": "width", "syntax": "auto"},
		{"syntax": "none"},
		{"name": "display", "syntax": "grid"}
	]}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []string{"block | inline", "grid"}
	if got := data.Properties["display"].Values; !slices.Equal(got, want) {
		t.Errorf("Properties[display].Values = %q, want %q", got, want)
	}
	if got := len(data.Properties); got != 2 {
		t.Errorf("len(Properties) = %d, want 2", got)
	}
}

func TestLoadNamedColors(t *testing.T) {
	for name, tc := range map[string]struct {
		document string
		want     []string
	}{
		"alternatives keep their order": {
			`{"types": [{"name": "named-color", "syntax": "red | rebeccapurple | blue"}]}`,
			[]string{"red", "rebeccapurple", "blue"},
		},
		"stray separators": {
			`{"types": [{"name": "named-color", "syntax": "| red ||  | blue |"}]}`,
			[]string{"red", "blue"},
		},
		"no syntax": {
			`{"types": [{"name": "named-color"}]}`,
			nil,
		},
		"another type only": {
			`{"types": [{"name": "length-percentage", "syntax": "[ <length> | <percentage> ]"}]}`,
			nil,
		},
	} {
		t.Run(name, func(t *testing.T) {
			data, err := loadJSON(t, tc.document)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if !slices.Equal(data.NamedColors, tc.want) {
				t.Errorf("NamedColors = %q, want %q", data.NamedColors, tc.want)
			}
		})
	}
}

func TestLoadEmptyDocument(t *testing.T) {
	data, err := loadJSON(t, `{}`)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if data.Properties == nil || len(data.Properties) != 0 {
		t.Errorf("Properties = %v, want an empty map", data.Properties)
	}
	if data.Features == nil || len(data.Features) != 0 {
		t.Errorf("Features = %v, want an empty map", data.Features)
	}
	if data.Descriptors == nil || len(data.Descriptors) != 0 {
		t.Errorf("Descriptors = %v, want an empty map", data.Descriptors)
	}
	if data.Values == nil || len(data.Values) != 0 {
		t.Errorf("Values = %v, want an empty map", data.Values)
	}
	if len(data.Selectors) != 0 || len(data.NamedColors) != 0 {
		t.Errorf("Selectors = %q, NamedColors = %q, want both empty", data.Selectors, data.NamedColors)
	}
}

func TestLoadReportsBadInput(t *testing.T) {
	if _, err := Load(t.TempDir()); err == nil {
		t.Error("Load of a directory without css.json returned no error")
	}
	if _, err := loadJSON(t, `{"properties": [`); err == nil {
		t.Error("Load of truncated JSON returned no error")
	}
	if _, err := loadJSON(t, `{"properties": {"name": "width"}}`); err == nil {
		t.Error("Load of a properties object instead of an array returned no error")
	}
}
