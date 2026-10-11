package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Toms223/WebGo/internal/cssgen/bcd"
	"github.com/Toms223/WebGo/internal/cssgen/webref"
)

var fixtureFloor = time.Date(2024, time.April, 1, 0, 0, 0, 0, time.UTC)

func loadFixtureModel(t *testing.T) (*model, report) {
	t.Helper()
	w, err := webref.Load(filepath.Join("testdata", "webref"))
	if err != nil {
		t.Fatalf("webref.Load: %v", err)
	}
	b, err := bcd.Load(filepath.Join("testdata", "bcd"))
	if err != nil {
		t.Fatalf("bcd.Load: %v", err)
	}
	m, r, err := build(w, b, fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	return m, r
}

func supportedEverywhere(keys ...string) *bcd.Data {
	data := &bcd.Data{Releases: map[string]map[string]bcd.Release{}, Features: map[string]*bcd.Compat{}}
	for _, browser := range bcd.Browsers {
		data.Releases[browser] = map[string]bcd.Release{"1": {Date: fixtureFloor, Status: "current"}}
	}
	for _, rule := range atRules {
		keys = append(slices.Clip(keys), rule.key)
	}
	for _, key := range keys {
		support := map[string][]bcd.Statement{}
		for _, browser := range bcd.Browsers {
			support[browser] = []bcd.Statement{{VersionAdded: "1", Added: true}}
		}
		data.Features[key] = &bcd.Compat{Support: support, StandardTrack: true}
	}
	return data
}

func syntaxes(grammars map[string]string) map[string]*webref.Property {
	properties := map[string]*webref.Property{}
	for name, syntax := range grammars {
		properties[name] = &webref.Property{Name: name, Values: []string{syntax}}
	}
	return properties
}

func describe[T any](items []T, line func(T) string) []string {
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = line(item)
	}
	return lines
}

func propertyLine(p property) string {
	return fmt.Sprintf("%s %s tier %d %+v prefixes %v", p.css, p.goName, p.tier, p.shape, p.prefixes)
}

func keywordLine(k keyword) string {
	return fmt.Sprintf("%s %s css %v media %v style %v", k.css, k.goName, k.cssProps, k.mediaFeatures, k.styleUses)
}

func featureLine(f mediaFeature) string {
	discrete := "nil"
	if f.discrete != nil {
		discrete = fmt.Sprintf("%+v", *f.discrete)
	}
	return fmt.Sprintf("%s %s range %q discrete %s", f.css, f.goName, f.rangeKind, discrete)
}

func descriptorLine(d descriptor) string {
	return fmt.Sprintf("%s %s %s tier %d %+v", d.atRule, d.css, d.goName, d.tier, d.shape)
}

func usesLine(k keyword) string {
	return fmt.Sprintf("%s descriptors %v container %v css %v", k.css, k.descriptors, k.containerFeatures, k.cssUses)
}

func pseudoLine(p pseudo) string {
	return fmt.Sprintf("%s %s element %v arg %q", p.css, p.goName, p.element, p.arg)
}

func expectLines(t *testing.T, what string, got, want []string) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("%s:\ngot\n\t%s\nwant\n\t%s", what, strings.Join(got, "\n\t"), strings.Join(want, "\n\t"))
	}
}

func TestBuildFromFixtures(t *testing.T) {
	m, r := loadFixtureModel(t)

	expectLines(t, "properties", describe(m.properties, propertyLine), []string{
		"-webkit-line-clamp WebkitLineClamp tier 1 {Kinds:[Integer] Keywords:[none] Min:1 Max:1 Comma:false} prefixes []",
		"backdrop-filter BackdropFilter tier 1 {Kinds:[UrlValue] Keywords:[none] Min:1 Max:-1 Comma:false} prefixes [-webkit-]",
		"border-width BorderWidth tier 1 {Kinds:[Length] Keywords:[medium thick thin] Min:1 Max:4 Comma:false} prefixes []",
		"cursor Cursor tier 1 {Kinds:[] Keywords:[auto default pointer] Min:1 Max:1 Comma:false} prefixes []",
		"display Display tier 1 {Kinds:[] Keywords:[block contents flex flow flow-root grid inline list-item none ruby run-in table] Min:1 Max:2 Comma:false} prefixes []",
		"font-family FontFamily tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:true} prefixes []",
		"padding Padding tier 1 {Kinds:[Length Percentage] Keywords:[] Min:1 Max:4 Comma:false} prefixes []",
		"padding-top PaddingTop tier 1 {Kinds:[Length Percentage] Keywords:[] Min:1 Max:1 Comma:false} prefixes []",
		"width Width tier 1 {Kinds:[Length Percentage] Keywords:[auto max-content min-content stretch] Min:1 Max:1 Comma:false} prefixes []",
	})
	if r.tier1 != 8 || r.tier2 != 1 {
		t.Errorf("tier1, tier2 = %d, %d, want 8, 1", r.tier1, r.tier2)
	}
	wantDropped := map[string][]string{"chrome": {"flagged-thing"}, "firefox": {"anchor-name"}}
	if !reflect.DeepEqual(r.droppedProps, wantDropped) {
		t.Errorf("droppedProps = %v, want %v", r.droppedProps, wantDropped)
	}
	for _, p := range m.properties {
		if p.css == "legacy-thing" {
			t.Error("legacy-thing is a property, but it is deprecated")
		}
	}
	for browser, names := range r.droppedProps {
		if slices.Contains(names, "legacy-thing") {
			t.Errorf("legacy-thing is reported as dropped for %s, but deprecated properties are skipped silently", browser)
		}
	}

	expectLines(t, "keywords", describe(m.keywords, keywordLine), []string{
		"auto Auto css [cursor width] media [] style []",
		"block Block css [display] media [] style []",
		"bold Bold css [] media [] style []",
		"collection Collection css [] media [] style []",
		"contents Contents css [display] media [] style []",
		"currentcolor Currentcolor css [] media [] style []",
		"dark Dark css [] media [prefers-color-scheme] style []",
		"default Default css [cursor] media [] style []",
		"embedded-opentype EmbeddedOpentype css [] media [] style []",
		"fallback Fallback css [] media [] style []",
		"false False css [] media [] style []",
		"flex Flex css [display] media [] style []",
		"flow Flow css [display] media [] style []",
		"flow-root FlowRoot css [display] media [] style []",
		"grid Grid css [display] media [] style []",
		"hover Hover css [] media [hover] style []",
		"inherit Inherit css [-webkit-line-clamp backdrop-filter border-width cursor display padding padding-top width] media [] style []",
		"initial Initial css [-webkit-line-clamp backdrop-filter border-width cursor display padding padding-top width] media [] style []",
		"inline Inline css [display] media [] style []",
		"landscape Landscape css [] media [orientation] style []",
		"light Light css [] media [prefers-color-scheme] style []",
		"list-item ListItem css [display] media [] style []",
		"ltr Ltr css [] media [] style [Dir]",
		"max-content MaxContent css [width] media [] style []",
		"medium Medium css [border-width] media [] style []",
		"min-content MinContent css [width] media [] style []",
		"none None css [-webkit-line-clamp backdrop-filter display] media [hover] style []",
		"normal Normal css [] media [] style []",
		"opentype Opentype css [] media [] style []",
		"optional Optional css [] media [] style []",
		"pointer Pointer css [cursor] media [] style []",
		"portrait Portrait css [] media [orientation] style []",
		"rebeccapurple Rebeccapurple css [] media [] style []",
		"red Red css [] media [] style []",
		"revert Revert css [-webkit-line-clamp backdrop-filter border-width cursor display padding padding-top width] media [] style []",
		"revert-layer RevertLayer css [-webkit-line-clamp backdrop-filter border-width cursor display padding padding-top width] media [] style []",
		"rtl Rtl css [] media [] style [Dir]",
		"ruby Ruby css [display] media [] style []",
		"run-in RunIn css [display] media [] style []",
		"stretch Stretch css [width] media [] style []",
		"svg Svg css [] media [] style []",
		"swap Swap css [] media [] style []",
		"table Table css [display] media [] style []",
		"thick Thick css [border-width] media [] style []",
		"thin Thin css [border-width] media [] style []",
		"transparent Transparent css [] media [] style []",
		"true True css [] media [] style []",
		"truetype Truetype css [] media [] style []",
		"unset Unset css [-webkit-line-clamp backdrop-filter border-width cursor display padding padding-top width] media [] style []",
		"woff Woff css [] media [] style []",
		"woff2 Woff2 css [] media [] style []",
	})

	expectLines(t, "features", describe(m.features, featureLine), []string{
		`aspect-ratio AspectRatio range "RatioValue" discrete nil`,
		`hover Hover range "" discrete {Kinds:[] Keywords:[hover none] Min:1 Max:1 Comma:false}`,
		`orientation Orientation range "" discrete {Kinds:[] Keywords:[landscape portrait] Min:1 Max:1 Comma:false}`,
		`prefers-color-scheme PrefersColorScheme range "" discrete {Kinds:[] Keywords:[dark light] Min:1 Max:1 Comma:false}`,
		`resolution Resolution range "Resolution" discrete nil`,
		`width Width range "Length" discrete nil`,
	})
	if want := []string{"-webkit-device-pixel-ratio", "grid"}; !slices.Equal(r.droppedFeatures, want) {
		t.Errorf("droppedFeatures = %q, want %q", r.droppedFeatures, want)
	}
	if !m.rangeSyntax {
		t.Error("rangeSyntax = false, want true")
	}

	expectLines(t, "pseudos", describe(m.pseudos, pseudoLine), []string{
		`::before Before element true arg ""`,
		`:default Default element false arg ""`,
		`:dir() Dir element false arg "dir"`,
		`:hover Hover element false arg ""`,
		`:lang() Lang element false arg "string"`,
		`:nth-child() NthChild element false arg "nth"`,
	})
	if want := []string{":first", ":left", ":not()", ":popover-open"}; !slices.Equal(r.droppedPseudos, want) {
		t.Errorf("droppedPseudos = %q, want %q", r.droppedPseudos, want)
	}
}

func TestBuildIsDeterministic(t *testing.T) {
	first, firstReport := loadFixtureModel(t)
	for range 5 {
		m, r := loadFixtureModel(t)
		if !reflect.DeepEqual(m, first) || !reflect.DeepEqual(r, firstReport) {
			t.Fatal("two builds from the same data differ")
		}
	}
}

func TestBuildCollectsKeywords(t *testing.T) {
	w := &webref.Data{
		Properties: syntaxes(map[string]string{
			"broken":          "auto | [",
			"color":           "<color>",
			"fill":            "none | currentColor | <color>",
			"font":            "[ <length> / <length> ] | caption | icon",
			"gone":            "auto | ghost",
			"shape-rendering": "auto | crispEdges",
		}),
		NamedColors: []string{"red"},
	}
	b := supportedEverywhere("css.properties.broken", "css.properties.color", "css.properties.fill", "css.properties.font", "css.properties.shape-rendering")
	m, r, err := build(w, b, fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	expectLines(t, "properties", describe(m.properties, propertyLine), []string{
		"broken Broken tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:false} prefixes []",
		"color Color tier 1 {Kinds:[ColorLiteral] Keywords:[] Min:1 Max:1 Comma:false} prefixes []",
		"fill Fill tier 1 {Kinds:[ColorLiteral] Keywords:[currentcolor none] Min:1 Max:1 Comma:false} prefixes []",
		"font Font tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:false} prefixes []",
		"shape-rendering ShapeRendering tier 1 {Kinds:[] Keywords:[auto crispedges] Min:1 Max:1 Comma:false} prefixes []",
	})
	if r.tier1 != 3 || r.tier2 != 2 {
		t.Errorf("tier1, tier2 = %d, %d, want 3, 2", r.tier1, r.tier2)
	}
	if len(r.droppedProps) != 0 {
		t.Errorf("droppedProps = %v, want none: a property missing from BCD is skipped silently", r.droppedProps)
	}
	expectLines(t, "keywords", describe(m.keywords, keywordLine), []string{
		"auto Auto css [shape-rendering] media [] style []",
		"caption Caption css [] media [] style []",
		"crispedges Crispedges css [shape-rendering] media [] style []",
		"currentcolor Currentcolor css [color fill] media [] style []",
		"icon Icon css [] media [] style []",
		"inherit Inherit css [color fill shape-rendering] media [] style []",
		"initial Initial css [color fill shape-rendering] media [] style []",
		"ltr Ltr css [] media [] style []",
		"none None css [fill] media [] style []",
		"red Red css [color fill] media [] style []",
		"revert Revert css [color fill shape-rendering] media [] style []",
		"revert-layer RevertLayer css [color fill shape-rendering] media [] style []",
		"rtl Rtl css [] media [] style []",
		"transparent Transparent css [color fill] media [] style []",
		"unset Unset css [color fill shape-rendering] media [] style []",
	})
}

func TestBuildAcceptsCSSWideKeywordsOnEveryTierOneProperty(t *testing.T) {
	w := &webref.Data{
		Properties: syntaxes(map[string]string{
			"color":  "<color>",
			"cursor": "auto | pointer",
			"font":   "[ <length> / <length> ] | caption",
			"width":  "<length>",
		}),
	}
	b := supportedEverywhere("css.properties.color", "css.properties.cursor", "css.properties.font", "css.properties.width")
	m, _, err := build(w, b, fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	var tierOne []string
	for _, p := range m.properties {
		if p.tier == 1 {
			tierOne = append(tierOne, p.css)
		}
	}
	if want := []string{"color", "cursor", "width"}; !slices.Equal(tierOne, want) {
		t.Fatalf("tier-1 properties = %q, want %q", tierOne, want)
	}
	for _, css := range []string{"inherit", "initial", "revert", "revert-layer", "unset"} {
		i := slices.IndexFunc(m.keywords, func(k keyword) bool { return k.css == css })
		if i < 0 {
			t.Errorf("keyword %s is missing", css)
			continue
		}
		k := m.keywords[i]
		if !slices.Equal(k.cssProps, tierOne) {
			t.Errorf("%s cssProps = %q, want every tier-1 property %q", css, k.cssProps, tierOne)
		}
		if slices.Contains(k.cssProps, "font") {
			t.Errorf("%s is accepted by the tier-2 property font", css)
		}
		if len(k.mediaFeatures)+len(k.styleUses) != 0 {
			t.Errorf("%s has media %q and style %q, want neither", css, k.mediaFeatures, k.styleUses)
		}
	}

	t.Run("without any property", func(t *testing.T) {
		m, _, err := build(&webref.Data{}, supportedEverywhere(), fixtureFloor)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		expectLines(t, "keywords", describe(m.keywords, keywordLine), []string{
			"currentcolor Currentcolor css [] media [] style []",
			"inherit Inherit css [] media [] style []",
			"initial Initial css [] media [] style []",
			"ltr Ltr css [] media [] style []",
			"revert Revert css [] media [] style []",
			"revert-layer RevertLayer css [] media [] style []",
			"rtl Rtl css [] media [] style []",
			"transparent Transparent css [] media [] style []",
			"unset Unset css [] media [] style []",
		})
	})
}

func TestBuildJoinsTierTwoCommaLists(t *testing.T) {
	w := &webref.Data{
		Properties: syntaxes(map[string]string{
			"background":  "<bg-layer>#? , <final-bg-layer>",
			"box-shadow":  "<spread-shadow>#",
			"mixed":       "<a> | <b> <c>",
			"partial":     "none | <x># | <y>",
			"text-shadow": "none | <shadow>#",
		}),
		Values: map[string]string{"shadow": "<length> <length>", "x": "<length> <length>"},
	}
	b := supportedEverywhere("css.properties.background", "css.properties.box-shadow", "css.properties.mixed", "css.properties.partial", "css.properties.text-shadow")
	m, _, err := build(w, b, fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	expectLines(t, "properties", describe(m.properties, propertyLine), []string{
		"background Background tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:true} prefixes []",
		"box-shadow BoxShadow tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:true} prefixes []",
		"mixed Mixed tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:false} prefixes []",
		"partial Partial tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:false} prefixes []",
		"text-shadow TextShadow tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:true} prefixes []",
	})
}

func TestBuildClassifiesMediaFeatures(t *testing.T) {
	w := &webref.Data{
		Features: map[string]string{
			"-webkit-device-pixel-ratio": "<number>",
			"aspect-ratio":               "<ratio>",
			"broken":                     "none |",
			"color":                      "<integer>",
			"device-pixel-ratio":         "<number>",
			"flag":                       "<mq-boolean>",
			"grid":                       "<mq-boolean>",
			"mixed":                      "<color> | none",
			"pointer":                    "none | Coarse | fine",
			"resolution":                 "<resolution> | infinite",
			"segments":                   "<length>+",
			"size":                       "<length-percentage>",
			"tint":                       "<color>",
			"unsupported":                "<length>",
			"width":                      "<length>",
		},
		Values: map[string]string{"mq-boolean": "<integer [0,1]>"},
	}
	var keys []string
	for name := range w.Features {
		if name != "unsupported" {
			keys = append(keys, "css.at-rules.media."+name)
		}
	}
	m, r, err := build(w, supportedEverywhere(keys...), fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	expectLines(t, "features", describe(m.features, featureLine), []string{
		`aspect-ratio AspectRatio range "RatioValue" discrete nil`,
		`color Color range "Integer" discrete nil`,
		`device-pixel-ratio DevicePixelRatio range "Number" discrete nil`,
		`flag Flag range "Integer" discrete nil`,
		`pointer Pointer range "" discrete {Kinds:[] Keywords:[coarse fine none] Min:1 Max:1 Comma:false}`,
		`resolution Resolution range "Resolution" discrete nil`,
		`width Width range "Length" discrete nil`,
	})
	if want := []string{"-webkit-device-pixel-ratio", "broken", "grid", "mixed", "segments", "size", "tint", "unsupported"}; !slices.Equal(r.droppedFeatures, want) {
		t.Errorf("droppedFeatures = %q, want %q", r.droppedFeatures, want)
	}
	expectLines(t, "keywords", describe(m.keywords, keywordLine), []string{
		"coarse Coarse css [] media [pointer] style []",
		"currentcolor Currentcolor css [] media [] style []",
		"fine Fine css [] media [pointer] style []",
		"inherit Inherit css [] media [] style []",
		"initial Initial css [] media [] style []",
		"ltr Ltr css [] media [] style []",
		"none None css [] media [pointer] style []",
		"revert Revert css [] media [] style []",
		"revert-layer RevertLayer css [] media [] style []",
		"rtl Rtl css [] media [] style []",
		"transparent Transparent css [] media [] style []",
		"unset Unset css [] media [] style []",
	})
}

func TestBuildRangeSyntax(t *testing.T) {
	const key = "css.at-rules.media.range_syntax"
	prefixed := supportedEverywhere(key)
	prefixed.Features[key].Support["safari"] = []bcd.Statement{{VersionAdded: "1", Added: true, Prefix: "-webkit-"}}
	cases := []struct {
		name string
		b    *bcd.Data
		want bool
	}{
		{"supported everywhere without a prefix", supportedEverywhere(key), true},
		{"missing from BCD", supportedEverywhere(), false},
		{"prefixed in a release since the floor", prefixed, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, _, err := build(&webref.Data{}, c.b, fixtureFloor)
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if m.rangeSyntax != c.want {
				t.Errorf("rangeSyntax = %v, want %v", m.rangeSyntax, c.want)
			}
		})
	}
}

func TestBuildPseudos(t *testing.T) {
	w := &webref.Data{Selectors: []string{
		">", ":unsupported", ":nth-last-of-type()", ":lang()", ":has()", ":before", "::highlight()", "::before", ":dir()",
		":right", ":blank", ":left", ":first",
	}}
	b := supportedEverywhere(
		"css.selectors.before", "css.selectors.dir", "css.selectors.has", "css.selectors.highlight", "css.selectors.lang", "css.selectors.nth-last-of-type",
		"css.selectors.blank", "css.selectors.first", "css.selectors.left", "css.selectors.right",
	)
	m, r, err := build(w, b, fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}

	expectLines(t, "pseudos", describe(m.pseudos, pseudoLine), []string{
		`::before Before element true arg ""`,
		`::highlight() Highlight element true arg "string"`,
		`:dir() Dir element false arg "dir"`,
		`:lang() Lang element false arg "string"`,
		`:nth-last-of-type() NthLastOfType element false arg "nth"`,
	})
	if want := []string{":before", ":blank", ":first", ":has()", ":left", ":right", ":unsupported"}; !slices.Equal(r.droppedPseudos, want) {
		t.Errorf("droppedPseudos = %q, want %q", r.droppedPseudos, want)
	}
	expectLines(t, "keywords", describe(m.keywords, keywordLine), []string{
		"currentcolor Currentcolor css [] media [] style []",
		"inherit Inherit css [] media [] style []",
		"initial Initial css [] media [] style []",
		"ltr Ltr css [] media [] style [Dir]",
		"revert Revert css [] media [] style []",
		"revert-layer RevertLayer css [] media [] style []",
		"rtl Rtl css [] media [] style [Dir]",
		"transparent Transparent css [] media [] style []",
		"unset Unset css [] media [] style []",
	})
}

func TestBuildRejectsReservedNames(t *testing.T) {
	cases := []struct {
		name string
		w    *webref.Data
		key  string
		want []string
	}{
		{"a pseudo named like a Rule method", &webref.Data{Selectors: []string{":name"}}, "css.selectors.name", []string{`":name"`, "Name"}},
		{"a property named like a css identifier", &webref.Data{Properties: syntaxes(map[string]string{"min": "<length>"})}, "css.properties.min", []string{`"min"`, "Min"}},
		{"a tier-1 interface named like a css type", &webref.Data{Properties: syntaxes(map[string]string{"string": "auto"})}, "css.properties.string", []string{`"string"`, "StringValue"}},
		{"a media feature named like a media identifier", &webref.Data{Features: map[string]string{"screen": "<length>"}}, "css.at-rules.media.screen", []string{`"screen"`, "Screen"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := build(c.w, supportedEverywhere(c.key), fixtureFloor)
			if err == nil {
				t.Fatal("build succeeded, want an error")
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %s", err, want)
				}
			}
		})
	}

	t.Run("a dropped pseudo is not checked", func(t *testing.T) {
		_, r, err := build(&webref.Data{Selectors: []string{":name"}}, supportedEverywhere(), fixtureFloor)
		if err != nil {
			t.Fatalf("build: %v", err)
		}
		if want := []string{":name"}; !slices.Equal(r.droppedPseudos, want) {
			t.Errorf("droppedPseudos = %q, want %q", r.droppedPseudos, want)
		}
	})
}

func TestBuildRejectsDuplicateGoNames(t *testing.T) {
	cases := []struct {
		name string
		w    *webref.Data
		keys []string
		want []string
	}{
		{"two pseudos", &webref.Data{Selectors: []string{":-x", ":x"}}, []string{"css.selectors.-x", "css.selectors.x"}, []string{`":-x"`, `":x"`, "X"}},
		{"two properties", &webref.Data{Properties: syntaxes(map[string]string{"-x-y": "<length>#", "x-y": "<length>#"})}, []string{"css.properties.-x-y", "css.properties.x-y"}, []string{`"-x-y"`, `"x-y"`, "XY"}},
		{"two keywords", &webref.Data{Properties: syntaxes(map[string]string{"z": "-a-b | a-b"})}, []string{"css.properties.z"}, []string{`"-a-b"`, `"a-b"`, "AB"}},
		{"two media features", &webref.Data{Features: map[string]string{"a--b": "<length>", "a-b": "<length>"}}, []string{"css.at-rules.media.a--b", "css.at-rules.media.a-b"}, []string{`"a--b"`, `"a-b"`, "AB"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, _, err := build(c.w, supportedEverywhere(c.keys...), fixtureFloor)
			if err == nil {
				t.Fatal("build succeeded, want an error")
			}
			for _, want := range c.want {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q does not mention %s", err, want)
				}
			}
		})
	}
}

func keywordUses(m *model, words ...string) []string {
	var lines []string
	for _, css := range words {
		i := slices.IndexFunc(m.keywords, func(k keyword) bool { return k.css == css })
		if i < 0 {
			lines = append(lines, css+" missing")
			continue
		}
		lines = append(lines, usesLine(m.keywords[i]))
	}
	return lines
}

func TestDescriptorsFromFixture(t *testing.T) {
	m, r := loadFixtureModel(t)

	expectLines(t, "descriptors", describe(m.descriptors, descriptorLine), []string{
		"font-face font-display FontDisplay tier 1 {Kinds:[] Keywords:[auto block fallback optional swap] Min:1 Max:1 Comma:false}",
		"font-face font-family FaceFontFamily tier 1 {Kinds:[IdentValue StringValue] Keywords:[] Min:1 Max:-1 Comma:false}",
		"font-face font-weight FontWeight tier 1 {Kinds:[Number] Keywords:[auto bold normal] Min:1 Max:2 Comma:false}",
		"font-face src Src tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:false}",
		"property inherits Inherits tier 1 {Kinds:[] Keywords:[false true] Min:1 Max:1 Comma:false}",
		"property initial-value InitialValue tier 2 {Kinds:[] Keywords:[] Min:0 Max:0 Comma:false}",
		"property syntax Syntax tier 1 {Kinds:[StringValue] Keywords:[] Min:1 Max:1 Comma:false}",
	})
	if want := map[string]int{"font-face": 4, "property": 3}; !reflect.DeepEqual(r.descriptors, want) {
		t.Errorf("descriptors per at-rule = %v, want %v", r.descriptors, want)
	}
	if want := []string{"font-face/ascent-override"}; !slices.Equal(r.droppedDescriptors, want) {
		t.Errorf("droppedDescriptors = %q, want %q", r.droppedDescriptors, want)
	}
	expectLines(t, "keyword uses", keywordUses(m, "swap", "auto", "bold", "true", "false", "woff2", "embedded-opentype", "inherit"), []string{
		"swap descriptors [FontDisplay] container [] css []",
		"auto descriptors [FontDisplay FontWeight] container [] css []",
		"bold descriptors [FontWeight] container [] css []",
		"true descriptors [Inherits] container [] css []",
		"false descriptors [Inherits] container [] css []",
		"woff2 descriptors [] container [] css [Format]",
		"embedded-opentype descriptors [] container [] css [Format]",
		"inherit descriptors [] container [] css []",
	})
}

func TestDescriptorNamesYieldToProperties(t *testing.T) {
	w := &webref.Data{
		Properties:  syntaxes(map[string]string{"font-style": "normal | italic"}),
		Descriptors: map[string]map[string]string{"font-face": {"font-style": "auto | normal | italic"}, "property": {"font-style": "auto"}},
	}
	b := supportedEverywhere("css.properties.font-style", "css.at-rules.font-face.font-style", "css.at-rules.property.font-style")
	m, _, err := build(w, b, fixtureFloor)
	if err != nil {
		t.Fatalf("build: %v", err)
	}
	expectLines(t, "descriptors", describe(m.descriptors, descriptorLine), []string{
		"font-face font-style FaceFontStyle tier 1 {Kinds:[] Keywords:[auto italic normal] Min:1 Max:1 Comma:false}",
		"property font-style PropertyFontStyle tier 1 {Kinds:[] Keywords:[auto] Min:1 Max:1 Comma:false}",
	})
}

func TestContainerFeaturesFromFixture(t *testing.T) {
	m, r := loadFixtureModel(t)

	expectLines(t, "container features", describe(m.containerFeatures, featureLine), []string{
		`inline-size InlineSize range "Length" discrete nil`,
		`orientation Orientation range "" discrete {Kinds:[] Keywords:[landscape portrait] Min:1 Max:1 Comma:false}`,
		`width Width range "Length" discrete nil`,
	})
	if want := []string{"scrollable"}; !slices.Equal(r.droppedContainer, want) {
		t.Errorf("droppedContainer = %q, want %q", r.droppedContainer, want)
	}
	expectLines(t, "keyword uses", keywordUses(m, "portrait", "landscape", "top"), []string{
		"portrait descriptors [] container [orientation] css []",
		"landscape descriptors [] container [orientation] css []",
		"top missing",
	})
}

func TestContainerFeatureKeys(t *testing.T) {
	w := &webref.Data{Descriptors: map[string]map[string]string{"container": {
		"fallback":   "none | <dashed-ident>",
		"scrollable": "none | top",
		"stuck":      "none | top",
		"width":      "<length>",
	}}}
	cases := []struct {
		name string
		b    *bcd.Data
		want []string
	}{
		{"only the container at-rule", supportedEverywhere(), []string{"width"}},
		{"with scroll-state queries", supportedEverywhere("css.at-rules.container.scroll-state_queries"), []string{"scrollable", "stuck", "width"}},
		{"with anchor position queries", supportedEverywhere("css.at-rules.container.anchor_position_queries"), []string{"width"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m, _, err := build(w, c.b, fixtureFloor)
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			got := describe(m.containerFeatures, func(f mediaFeature) string { return f.css })
			if !slices.Equal(got, c.want) {
				t.Errorf("container features = %q, want %q", got, c.want)
			}
		})
	}
}

func TestAtRuleVerdicts(t *testing.T) {
	_, r := loadFixtureModel(t)
	want := []string{"container=true", "property=true", "layer=true", "supports=true", "import=true", "font-face=true", "keyframes=true", "nesting=true"}
	got := describe(r.atRules, func(v verdict) string { return fmt.Sprintf("%s=%t", v.name, v.renderable) })
	if !slices.Equal(got, want) {
		t.Errorf("at-rule verdicts = %q, want %q", got, want)
	}

	cases := []struct {
		rule, key string
	}{
		{"layer", "css.at-rules.layer"},
		{"nesting", "css.selectors.nesting"},
	}
	for _, c := range cases {
		t.Run(c.rule, func(t *testing.T) {
			w, err := webref.Load(filepath.Join("testdata", "webref"))
			if err != nil {
				t.Fatal(err)
			}
			b, err := bcd.Load(filepath.Join("testdata", "bcd"))
			if err != nil {
				t.Fatal(err)
			}
			b.Features[c.key].Support["safari"] = []bcd.Statement{{}}
			_, _, err = build(w, b, fixtureFloor)
			if err == nil {
				t.Fatalf("build succeeded with %s unsupported in safari", c.rule)
			}
			if !strings.Contains(err.Error(), c.rule) {
				t.Errorf("error %q does not name %s", err, c.rule)
			}
		})
	}
}
