package bcd

import (
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func day(year int, month time.Month, date int) time.Time {
	return time.Date(year, month, date, 0, 0, 0, 0, time.UTC)
}

func loadFixture(t *testing.T) *Data {
	t.Helper()
	data, err := Load(filepath.Join("..", "testdata", "bcd"))
	if err != nil {
		t.Fatalf("expected the fixture to load, got %v", err)
	}
	return data
}

func loadDocument(t *testing.T, document string) (*Data, error) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.json"), []byte(document), 0o644); err != nil {
		t.Fatal(err)
	}
	return Load(dir)
}

func propertyDocument(support string) string {
	return `{"css": {"properties": {"x": {"__compat": {"support": {"chrome": ` + support + `}}}}}}`
}

func uniform(releases map[string]Release) map[string]map[string]Release {
	all := make(map[string]map[string]Release, len(Browsers))
	for _, browser := range Browsers {
		all[browser] = maps.Clone(releases)
	}
	return all
}

func everywhere(statements ...Statement) map[string][]Statement {
	support := make(map[string][]Statement, len(Browsers))
	for _, browser := range Browsers {
		support[browser] = statements
	}
	return support
}

func override(support map[string][]Statement, browser string, statements ...Statement) map[string][]Statement {
	support[browser] = statements
	return support
}

func without(support map[string][]Statement, browser string) map[string][]Statement {
	delete(support, browser)
	return support
}

func singleFeature(releases map[string]map[string]Release, support map[string][]Statement) *Data {
	return &Data{Releases: releases, Features: map[string]*Compat{"x": {Support: support}}}
}

func currentReleases() map[string]map[string]Release {
	return uniform(map[string]Release{
		"120": {Date: day(2023, 12, 5), Status: "retired"},
		"130": {Date: day(2024, 10, 15), Status: "current"},
		"131": {Date: day(2024, 11, 12), Status: "beta"},
	})
}

func TestBrowsers(t *testing.T) {
	want := []string{"chrome", "chrome_android", "edge", "firefox", "firefox_android", "safari", "safari_ios"}
	if !slices.Equal(Browsers, want) {
		t.Fatalf("expected %v, got %v", want, Browsers)
	}
}

func TestLatest(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		browser string
		want    string
	}{
		{"chrome", "130"},
		{"chrome_android", "130"},
		{"edge", "130"},
		{"firefox", "131"},
		{"firefox_android", "131"},
		{"safari", "18"},
		{"safari_ios", "18"},
		{"ie", ""},
		{"netscape", ""},
	}

	for _, tc := range cases {
		t.Run(tc.browser, func(t *testing.T) {
			if got := data.Latest(tc.browser); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestLatestFallbacks(t *testing.T) {
	cases := []struct {
		name     string
		releases map[string]Release
		want     string
	}{
		{"versions order numerically", map[string]Release{"9": {Status: "current"}, "100": {Status: "retired"}}, "9"},
		{"the highest of several current releases", map[string]Release{"9": {Status: "current"}, "100": {Status: "current"}, "99": {Status: "current"}}, "100"},
		{"dotted versions order by segment", map[string]Release{"17": {Status: "current"}, "17.5": {Status: "current"}, "17.10": {Status: "current"}}, "17.10"},
		{"a current release beats newer unreleased ones", map[string]Release{"130": {Status: "current"}, "131": {Status: "beta"}, "132": {Status: "nightly"}, "133": {Status: "planned"}}, "130"},
		{"no current release falls back to the highest", map[string]Release{"9": {Status: "retired"}, "100": {Status: "retired"}, "99": {Status: "retired"}}, "100"},
		{"no releases at all", map[string]Release{}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := &Data{Releases: map[string]map[string]Release{"chrome": tc.releases}}
			if got := data.Latest("chrome"); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}

	t.Run("equal versions resolve the same way every time", func(t *testing.T) {
		data := &Data{Releases: map[string]map[string]Release{"chrome": {"18": {Status: "current"}, "18.0": {Status: "current"}}}}
		for range 50 {
			if got := data.Latest("chrome"); got != "18.0" {
				t.Fatalf("expected %q, got %q", "18.0", got)
			}
		}
	})
}

func TestRenderable(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		key  string
		want bool
	}{
		{"css.properties.width", true},
		{"css.properties.anchor-name", false},
		{"css.properties.flagged-thing", false},
		{"css.properties.aliased-thing", false},
		{"css.selectors.popover-open", false},
		{"css.properties.backdrop-filter", true},
		{"css.properties.-webkit-line-clamp", true},
		{"css.properties.legacy-thing", true},
		{"css.at-rules.media.range_syntax", true},
		{"css.selectors.hover", true},
		{"css.properties.missing", false},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			if got := data.Renderable(tc.key); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestRenderableRules(t *testing.T) {
	releases := currentReleases()
	supported := Statement{Added: true, VersionAdded: "1"}
	aliased := Statement{Added: true, VersionAdded: "1", Alternative: "-webkit-x"}

	cases := []struct {
		name    string
		support map[string][]Statement
		want    bool
	}{
		{"supported by every browser", everywhere(supported), true},
		{"always supported", everywhere(Statement{Added: true}), true},
		{"added in the latest release", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "130"}), true},
		{"added after an older release but before the latest", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "125"}), true},
		{"added after the latest current release", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "131"}), false},
		{"not supported", override(everywhere(supported), "firefox", Statement{}), false},
		{"a browser without statements", override(everywhere(supported), "safari_ios"), false},
		{"a browser missing from the support", without(everywhere(supported), "safari_ios"), false},
		{"prefixed support only", override(everywhere(supported), "safari", Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-"}), true},
		{"behind a flag", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", Flagged: true}), false},
		{"a flagged and an unflagged statement", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", Flagged: true}, supported), true},
		{"supported only under an alternative name", override(everywhere(supported), "safari", aliased), false},
		{"supported under an alternative name everywhere", everywhere(aliased), false},
		{"an alternative name next to the real name", override(everywhere(supported), "safari", aliased, supported), true},
		{"an alternative name that is also prefixed", override(everywhere(supported), "safari", Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-", Alternative: "-webkit-x"}), false},
		{"a partial implementation", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", Partial: true}), true},
		{"a preview only", override(everywhere(supported), "firefox", Statement{Added: true, VersionAdded: "preview"}), false},
		{"removed before the latest release", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", VersionRemoved: "125"}), false},
		{"removed in the latest release", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", VersionRemoved: "130"}), false},
		{"removed after the latest release", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", VersionRemoved: "140"}), true},
		{"removed and added back", override(everywhere(supported), "chrome", Statement{Added: true, VersionAdded: "1", VersionRemoved: "125"}, Statement{Added: true, VersionAdded: "128"}), true},
		{"a browser that is not considered", override(everywhere(supported), "ie", Statement{}), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := singleFeature(releases, tc.support)
			if got := data.Renderable("x"); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
			if unsupported := data.FirstUnsupported("x"); (unsupported == "") != tc.want {
				t.Fatalf("expected no unsupported browser exactly when renderable, got %q", unsupported)
			}
		})
	}

	t.Run("a nil entry", func(t *testing.T) {
		data := &Data{Releases: releases, Features: map[string]*Compat{"x": nil}}
		if data.Renderable("x") {
			t.Fatal("expected a nil entry to not be renderable")
		}
	})
}

func TestRenderableUsesEachBrowsersOwnLatestRelease(t *testing.T) {
	releases := currentReleases()
	releases["safari"] = map[string]Release{"18": {Date: day(2024, 9, 16), Status: "current"}}
	supported := Statement{Added: true, VersionAdded: "1"}

	cases := []struct {
		name    string
		support map[string][]Statement
		want    bool
		first   string
	}{
		{"added after that browser's latest release", override(everywhere(supported), "safari", Statement{Added: true, VersionAdded: "19"}), false, "safari"},
		{"added in that browser's latest release", override(everywhere(supported), "safari", Statement{Added: true, VersionAdded: "18"}), true, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := singleFeature(releases, tc.support)
			if got := data.Renderable("x"); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
			if got := data.FirstUnsupported("x"); got != tc.first {
				t.Fatalf("expected %q, got %q", tc.first, got)
			}
		})
	}
}

func TestFirstUnsupported(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		key  string
		want string
	}{
		{"css.properties.anchor-name", "firefox"},
		{"css.properties.flagged-thing", "chrome"},
		{"css.properties.width", ""},
		{"css.properties.nope", "chrome"},
		{"css.properties.aliased-thing", "chrome"},
		{"css.selectors.popover-open", "firefox"},
		{"css.properties.backdrop-filter", ""},
		{"css.at-rules.media.range_syntax", ""},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			if got := data.FirstUnsupported(tc.key); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestFirstUnsupportedRules(t *testing.T) {
	releases := currentReleases()
	supported := Statement{Added: true, VersionAdded: "1"}

	cases := []struct {
		name    string
		support map[string][]Statement
		want    string
	}{
		{"supported by every browser", everywhere(supported), ""},
		{"the earliest unsupported browser in order wins", override(override(everywhere(supported), "safari", Statement{}), "chrome_android", Statement{}), "chrome_android"},
		{"the last browser in order", override(everywhere(supported), "safari_ios", Statement{}), "safari_ios"},
		{"a browser without statements", override(everywhere(supported), "edge"), "edge"},
		{"a browser missing from the support", without(everywhere(supported), "firefox"), "firefox"},
		{"behind a flag", override(everywhere(supported), "firefox_android", Statement{Added: true, VersionAdded: "1", Flagged: true}), "firefox_android"},
		{"only an alternative name", override(everywhere(supported), "safari", Statement{Added: true, VersionAdded: "1", Alternative: "-webkit-x"}), "safari"},
		{"added after the latest current release", override(everywhere(supported), "edge", Statement{Added: true, VersionAdded: "131"}), "edge"},
		{"prefixed support counts", override(everywhere(supported), "safari", Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-"}), ""},
		{"no support at all", map[string][]Statement{}, "chrome"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := singleFeature(releases, tc.support)
			if got := data.FirstUnsupported("x"); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}

	t.Run("a nil entry", func(t *testing.T) {
		data := &Data{Releases: releases, Features: map[string]*Compat{"x": nil}}
		if got := data.FirstUnsupported("x"); got != "chrome" {
			t.Fatalf("expected %q, got %q", "chrome", got)
		}
	})

	t.Run("a missing key", func(t *testing.T) {
		data := &Data{Releases: releases}
		if got := data.FirstUnsupported("x"); got != "chrome" {
			t.Fatalf("expected %q, got %q", "chrome", got)
		}
	})
}

func TestPrefixes(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		name  string
		key   string
		floor time.Time
		want  []string
	}{
		{"a release since the floor lacks unprefixed support", "css.properties.backdrop-filter", day(2024, 4, 1), []string{"-webkit-"}},
		{"every release since the floor is unprefixed", "css.properties.backdrop-filter", day(2024, 10, 1), nil},
		{"a property no browser prefixes", "css.properties.width", day(2024, 4, 1), nil},
		{"the floor date itself counts", "css.properties.backdrop-filter", day(2024, 5, 13), []string{"-webkit-"}},
		{"the day after the last prefixed-only release", "css.properties.backdrop-filter", day(2024, 5, 14), nil},
		{"a floor that predates every release", "css.properties.backdrop-filter", day(2023, 1, 1), []string{"-webkit-"}},
		{"a property that is unsupported without any prefix", "css.properties.anchor-name", day(2024, 4, 1), nil},
		{"a media feature that is never prefixed", "css.at-rules.media.range_syntax", day(2024, 4, 1), nil},
		{"a key that does not exist", "css.properties.missing", day(2024, 4, 1), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := data.Prefixes(tc.key, tc.floor); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %#v, got %#v", tc.want, got)
			}
		})
	}
}

func TestPrefixesRules(t *testing.T) {
	releases := uniform(map[string]Release{
		"5":   {Status: "retired"},
		"9":   {Date: day(2023, 1, 10), Status: "retired"},
		"10":  {Date: day(2024, 6, 1), Status: "retired"},
		"130": {Date: day(2025, 1, 1), Status: "current"},
	})
	releases["ie"] = map[string]Release{"11": {Date: day(2024, 7, 1), Status: "retired"}}

	plain := Statement{Added: true, VersionAdded: "1"}
	modern := Statement{Added: true, VersionAdded: "130"}
	webkit := Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-"}
	moz := Statement{Added: true, VersionAdded: "1", Prefix: "-moz-"}
	floor := day(2024, 1, 1)

	cases := []struct {
		name    string
		support map[string][]Statement
		floor   time.Time
		want    []string
	}{
		{"prefixed until a later unprefixed release", override(everywhere(plain), "safari", modern, webkit), floor, []string{"-webkit-"}},
		{"releases before the floor and undated releases are ignored", override(everywhere(plain), "safari", Statement{Added: true, VersionAdded: "10"}, webkit), floor, nil},
		{"an older release comes in with an older floor", override(everywhere(plain), "safari", Statement{Added: true, VersionAdded: "10"}, webkit), day(2023, 1, 1), []string{"-webkit-"}},
		{"the floor date itself counts", override(everywhere(plain), "safari", modern, webkit), day(2024, 6, 1), []string{"-webkit-"}},
		{"the day after a release excludes it", override(everywhere(plain), "safari", modern, webkit), day(2024, 6, 2), nil},
		{"prefixes are sorted and deduplicated", override(override(override(everywhere(plain), "chrome", modern, webkit), "firefox", modern, moz), "safari", modern, webkit), floor, []string{"-moz-", "-webkit-"}},
		{"several prefixes in one browser", override(everywhere(plain), "firefox", modern, webkit, moz), floor, []string{"-moz-", "-webkit-"}},
		{"a flagged unprefixed statement is not support", override(everywhere(plain), "safari", Statement{Added: true, VersionAdded: "1", Flagged: true}, webkit), floor, []string{"-webkit-"}},
		{"a flagged prefixed statement is not support", override(everywhere(plain), "safari", modern, Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-", Flagged: true}), floor, nil},
		{"an aliased unprefixed statement is not support", override(everywhere(plain), "safari", Statement{Added: true, VersionAdded: "1", Alternative: "-webkit-x"}, webkit), floor, []string{"-webkit-"}},
		{"an aliased prefixed statement is not support", override(everywhere(plain), "safari", modern, Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-", Alternative: "-webkit-x"}), floor, nil},
		{"a prefixed statement removed in the release", override(everywhere(plain), "safari", modern, Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-", VersionRemoved: "10"}), floor, nil},
		{"a prefixed statement removed after the release", override(everywhere(plain), "safari", modern, Statement{Added: true, VersionAdded: "1", Prefix: "-webkit-", VersionRemoved: "11"}), floor, []string{"-webkit-"}},
		{"a prefixed statement added after the release", override(everywhere(plain), "safari", modern, Statement{Added: true, VersionAdded: "20", Prefix: "-webkit-"}), floor, nil},
		{"an unprefixed preview does not count", override(everywhere(plain), "safari", Statement{Added: true, VersionAdded: "preview"}, webkit), floor, []string{"-webkit-"}},
		{"no prefixed statement", override(everywhere(plain), "safari", modern), floor, nil},
		{"an unsupported statement carrying a prefix", override(everywhere(plain), "safari", modern, Statement{Prefix: "-webkit-"}), floor, nil},
		{"a browser that is not considered", override(everywhere(plain), "ie", Statement{Added: true, VersionAdded: "10", Prefix: "-ms-"}), floor, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := singleFeature(releases, tc.support)
			if got := data.Prefixes("x", tc.floor); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("expected %#v, got %#v", tc.want, got)
			}
		})
	}
}

func TestStandard(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		key  string
		want bool
	}{
		{"css.properties.legacy-thing", false},
		{"css.properties.width", true},
		{"css.properties.-webkit-line-clamp", true},
		{"css.properties.missing", false},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			if got := data.Standard(tc.key); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestStandardRules(t *testing.T) {
	data := &Data{Features: map[string]*Compat{
		"standard":                {StandardTrack: true},
		"deprecated":              {StandardTrack: true, Deprecated: true},
		"non-standard":            {},
		"non-standard deprecated": {Deprecated: true},
		"nil":                     nil,
	}}
	cases := []struct {
		key  string
		want bool
	}{
		{"standard", true},
		{"deprecated", false},
		{"non-standard", false},
		{"non-standard deprecated", false},
		{"nil", false},
		{"missing", false},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			if got := data.Standard(tc.key); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestHas(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		key  string
		want bool
	}{
		{"css.properties.width", true},
		{"css.at-rules.media.range_syntax", true},
		{"css.selectors.hover", true},
		{"css.properties.missing", false},
		{"width", false},
		{"css.at-rules.media", true},
		{"css.types.length", false},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			if got := data.Has(tc.key); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestVersionLess(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"9", "18", true},
		{"18", "9", false},
		{"16.4", "17", true},
		{"17", "16.4", false},
		{"99", "104", true},
		{"104", "99", false},
		{"200", "preview", true},
		{"preview", "200", false},
		{"preview", "preview", false},
		{"17", "17", false},
		{"17", "17.0", false},
		{"17.0", "17", false},
		{"17.5", "17.10", true},
		{"17.10", "17.5", false},
		{"1.5", "2", true},
		{"≤79", "80", true},
		{"≤79", "79", false},
		{"", "1", true},
		{"1", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.a+" < "+tc.b, func(t *testing.T) {
			if got := VersionLess(tc.a, tc.b); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestCovers(t *testing.T) {
	cases := []struct {
		name      string
		statement Statement
		version   string
		want      bool
	}{
		{"always supported", Statement{Added: true}, "1", true},
		{"unsupported", Statement{}, "130", false},
		{"unsupported ignores its version", Statement{VersionAdded: "1"}, "130", false},
		{"added at the version", Statement{Added: true, VersionAdded: "18"}, "18", true},
		{"added before the version", Statement{Added: true, VersionAdded: "9"}, "18", true},
		{"added after the version", Statement{Added: true, VersionAdded: "18"}, "17.5", false},
		{"added in a dotted version", Statement{Added: true, VersionAdded: "16.4"}, "16.10", true},
		{"a preview never covers a release", Statement{Added: true, VersionAdded: "preview"}, "200", false},
		{"an alternative name covers nothing", Statement{Added: true, VersionAdded: "1", Alternative: "-webkit-x"}, "130", false},
		{"an always-supported alternative name covers nothing", Statement{Added: true, Alternative: "-webkit-x"}, "130", false},
		{"a ranged version counts as its bound", Statement{Added: true, VersionAdded: "≤79"}, "79", true},
		{"a ranged version does not cover below its bound", Statement{Added: true, VersionAdded: "≤79"}, "78", false},
		{"removed after the version", Statement{Added: true, VersionAdded: "9", VersionRemoved: "20"}, "19", true},
		{"removed at the version", Statement{Added: true, VersionAdded: "9", VersionRemoved: "20"}, "20", false},
		{"removed before the version", Statement{Added: true, VersionAdded: "9", VersionRemoved: "20"}, "130", false},
		{"always supported until its removal", Statement{Added: true, VersionRemoved: "20"}, "19", true},
		{"always supported until its removal, then gone", Statement{Added: true, VersionRemoved: "20"}, "20", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Covers(tc.statement, tc.version); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestLoadDecodesReleases(t *testing.T) {
	data := loadFixture(t)
	blink := map[string]Release{
		"120": {Date: day(2023, 12, 5), Status: "retired"},
		"130": {Date: day(2024, 10, 15), Status: "current"},
	}
	gecko := map[string]Release{
		"120": {Date: day(2023, 11, 21), Status: "retired"},
		"131": {Date: day(2024, 10, 1), Status: "current"},
	}
	webkit := map[string]Release{
		"17":   {Date: day(2023, 9, 18), Status: "retired"},
		"17.5": {Date: day(2024, 5, 13), Status: "retired"},
		"18":   {Date: day(2024, 9, 16), Status: "current"},
	}
	want := map[string]map[string]Release{
		"chrome":          blink,
		"chrome_android":  blink,
		"edge":            blink,
		"firefox":         gecko,
		"firefox_android": gecko,
		"safari":          webkit,
		"safari_ios":      webkit,
	}

	if len(data.Releases) != len(want) {
		t.Fatalf("expected releases for %d browsers, got %d", len(want), len(data.Releases))
	}
	for browser, versions := range want {
		got := data.Releases[browser]
		if len(got) != len(versions) {
			t.Fatalf("%s: expected %d releases, got %d", browser, len(versions), len(got))
		}
		for version, release := range versions {
			if !got[version].Date.Equal(release.Date) || got[version].Status != release.Status {
				t.Errorf("%s %s: expected %v %q, got %v %q", browser, version, release.Date, release.Status, got[version].Date, got[version].Status)
			}
		}
	}
}

func TestLoadDecodesSupport(t *testing.T) {
	data := loadFixture(t)
	webkitPrefixed := []Statement{{VersionAdded: "18", Added: true}, {VersionAdded: "9", Added: true, Prefix: "-webkit-"}}
	cases := []struct {
		key     string
		browser string
		want    []Statement
	}{
		{"css.properties.width", "chrome", []Statement{{VersionAdded: "1", Added: true}}},
		{"css.properties.anchor-name", "chrome", []Statement{{VersionAdded: "125", Added: true}}},
		{"css.properties.anchor-name", "firefox", []Statement{{}}},
		{"css.properties.flagged-thing", "chrome", []Statement{{VersionAdded: "120", Added: true, Flagged: true}}},
		{"css.properties.aliased-thing", "chrome", []Statement{{VersionAdded: "1", Added: true, Alternative: "-webkit-aliased-thing"}}},
		{"css.properties.aliased-thing", "safari_ios", []Statement{{VersionAdded: "1", Added: true, Alternative: "-webkit-aliased-thing"}}},
		{"css.properties.backdrop-filter", "chrome", []Statement{{VersionAdded: "76", Added: true}}},
		{"css.properties.backdrop-filter", "safari", webkitPrefixed},
		{"css.properties.backdrop-filter", "safari_ios", webkitPrefixed},
		{"css.at-rules.media.range_syntax", "safari", []Statement{{VersionAdded: "16.4", Added: true}}},
		{"css.selectors.popover-open", "firefox", []Statement{{}}},
	}

	for _, tc := range cases {
		t.Run(tc.key+" "+tc.browser, func(t *testing.T) {
			got := data.Features[tc.key].Support[tc.browser]
			if !slices.Equal(got, tc.want) {
				t.Fatalf("expected %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestLoadDecodesStatus(t *testing.T) {
	data := loadFixture(t)
	cases := []struct {
		key           string
		standardTrack bool
		deprecated    bool
	}{
		{"css.properties.width", true, false},
		{"css.properties.legacy-thing", true, true},
		{"css.selectors.popover-open", true, false},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			compat := data.Features[tc.key]
			if compat.StandardTrack != tc.standardTrack || compat.Deprecated != tc.deprecated {
				t.Fatalf("expected standard track %v and deprecated %v, got %v and %v", tc.standardTrack, tc.deprecated, compat.StandardTrack, compat.Deprecated)
			}
		})
	}
}

func TestLoadFlattensTheFixture(t *testing.T) {
	data := loadFixture(t)
	want := []string{
		"css.at-rules.container",
		"css.at-rules.container.scroll-state_queries",
		"css.at-rules.container.style_queries_for_custom_properties",
		"css.at-rules.font-face",
		"css.at-rules.font-face.ascent-override",
		"css.at-rules.font-face.font-display",
		"css.at-rules.font-face.font-family",
		"css.at-rules.font-face.font-weight",
		"css.at-rules.font-face.src",
		"css.at-rules.import",
		"css.at-rules.import.layer",
		"css.at-rules.import.supports",
		"css.at-rules.keyframes",
		"css.at-rules.layer",
		"css.at-rules.media",
		"css.at-rules.media.-webkit-device-pixel-ratio",
		"css.at-rules.media.aspect-ratio",
		"css.at-rules.media.grid",
		"css.at-rules.media.hover",
		"css.at-rules.media.orientation",
		"css.at-rules.media.prefers-color-scheme",
		"css.at-rules.media.range_syntax",
		"css.at-rules.media.resolution",
		"css.at-rules.media.width",
		"css.at-rules.property",
		"css.at-rules.property.inherits",
		"css.at-rules.property.initial-value",
		"css.at-rules.property.syntax",
		"css.at-rules.supports",
		"css.at-rules.supports.selector",
		"css.properties.-webkit-line-clamp",
		"css.properties.aliased-thing",
		"css.properties.anchor-name",
		"css.properties.backdrop-filter",
		"css.properties.border-width",
		"css.properties.cursor",
		"css.properties.display",
		"css.properties.flagged-thing",
		"css.properties.font-family",
		"css.properties.legacy-thing",
		"css.properties.padding",
		"css.properties.padding-top",
		"css.properties.width",
		"css.selectors.before",
		"css.selectors.default",
		"css.selectors.dir",
		"css.selectors.first",
		"css.selectors.hover",
		"css.selectors.lang",
		"css.selectors.left",
		"css.selectors.nesting",
		"css.selectors.not",
		"css.selectors.nth-child",
		"css.selectors.nth-child.of_selector_list",
		"css.selectors.popover-open",
	}
	slices.Sort(want)

	if got := slices.Sorted(maps.Keys(data.Features)); !slices.Equal(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestLoadCollectsEveryAtRule(t *testing.T) {
	data := loadFixture(t)
	for _, key := range []string{
		"css.at-rules.container",
		"css.at-rules.container.style_queries_for_custom_properties",
		"css.at-rules.font-face.src",
		"css.at-rules.property.syntax",
		"css.at-rules.layer",
		"css.at-rules.supports.selector",
		"css.at-rules.import.layer",
		"css.at-rules.keyframes",
		"css.selectors.nesting",
	} {
		if !data.Has(key) {
			t.Errorf("Has(%q) = false, want true", key)
		}
	}

	for key, want := range map[string]bool{
		"css.at-rules.font-face.ascent-override":                     false,
		"css.at-rules.container.scroll-state_queries":                false,
		"css.at-rules.container":                                     true,
		"css.at-rules.container.style_queries_for_custom_properties": true,
		"css.selectors.nesting":                                      true,
	} {
		if got := data.Renderable(key); got != want {
			t.Errorf("Renderable(%q) = %v, want %v", key, got, want)
		}
	}
}

func TestLoadKeepsOnlyTheConsideredBrowsers(t *testing.T) {
	data := loadFixture(t)
	want := slices.Sorted(slices.Values(Browsers))

	if got := slices.Sorted(maps.Keys(data.Releases)); !slices.Equal(got, want) {
		t.Fatalf("expected releases for %v, got %v", want, got)
	}
	if got := slices.Sorted(maps.Keys(data.Features["css.properties.width"].Support)); !slices.Equal(got, want) {
		t.Fatalf("expected support for %v, got %v", want, got)
	}
}

func TestLoadDecodesSupportStatements(t *testing.T) {
	cases := []struct {
		name    string
		support string
		want    []Statement
	}{
		{"a version", `{"version_added": "130"}`, []Statement{{VersionAdded: "130", Added: true}}},
		{"always supported", `{"version_added": true}`, []Statement{{Added: true}}},
		{"not supported", `{"version_added": false}`, []Statement{{}}},
		{"unknown support", `{"version_added": null}`, []Statement{{}}},
		{"no version", `{}`, []Statement{{}}},
		{"a preview version", `{"version_added": "preview"}`, []Statement{{VersionAdded: "preview", Added: true}}},
		{"a ranged version", `{"version_added": "≤79"}`, []Statement{{VersionAdded: "≤79", Added: true}}},
		{"a removal version", `{"version_added": "18", "version_removed": "25"}`, []Statement{{VersionAdded: "18", VersionRemoved: "25", Added: true}}},
		{"a removal in an unknown version", `{"version_added": "18", "version_removed": true}`, []Statement{{VersionAdded: "18"}}},
		{"no removal", `{"version_added": "18", "version_removed": false}`, []Statement{{VersionAdded: "18", Added: true}}},
		{"a prefix", `{"version_added": "9", "prefix": "-webkit-"}`, []Statement{{VersionAdded: "9", Added: true, Prefix: "-webkit-"}}},
		{"a flag", `{"version_added": "120", "flags": [{"type": "preference", "name": "x"}]}`, []Statement{{VersionAdded: "120", Added: true, Flagged: true}}},
		{"no flags", `{"version_added": "120", "flags": []}`, []Statement{{VersionAdded: "120", Added: true}}},
		{"a partial implementation", `{"version_added": "120", "partial_implementation": true}`, []Statement{{VersionAdded: "120", Added: true, Partial: true}}},
		{"unrelated fields", `{"version_added": "120", "notes": ["a"], "impl_url": "b"}`, []Statement{{VersionAdded: "120", Added: true}}},
		{"an alternative name", `{"version_added": "1", "alternative_name": "-webkit-aliased-thing"}`, []Statement{{VersionAdded: "1", Added: true, Alternative: "-webkit-aliased-thing"}}},
		{"an array of statements", `[{"version_added": "18"}, {"version_added": "9", "prefix": "-webkit-"}]`, []Statement{{VersionAdded: "18", Added: true}, {VersionAdded: "9", Added: true, Prefix: "-webkit-"}}},
		{"an empty array", `[]`, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := loadDocument(t, propertyDocument(tc.support))
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if got := data.Features["css.properties.x"].Support["chrome"]; !slices.Equal(got, tc.want) {
				t.Fatalf("expected %+v, got %+v", tc.want, got)
			}
		})
	}
}

func TestLoadFlattensNestedFeatures(t *testing.T) {
	data, err := loadDocument(t, `{
		"api": {"Document": {"__compat": {}}},
		"css": {
			"types": {"length": {"__compat": {}}},
			"at-rules": {
				"import": {"__compat": {}},
				"media": {
					"__compat": {},
					"hover": {"__compat": {}},
					"outer": {"inner": {"__compat": {}}}
				}
			},
			"properties": {
				"width": {"__compat": {}, "fit-content": {"__compat": {}}},
				"group": {"child": {"__compat": {}}}
			},
			"selectors": {"hover": {"__compat": {}}}
		}
	}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	want := []string{
		"css.at-rules.import",
		"css.at-rules.media",
		"css.at-rules.media.hover",
		"css.at-rules.media.outer.inner",
		"css.properties.group.child",
		"css.properties.width",
		"css.properties.width.fit-content",
		"css.selectors.hover",
	}
	if got := slices.Sorted(maps.Keys(data.Features)); !slices.Equal(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestLoadDecodesAReleaseWithoutADate(t *testing.T) {
	data, err := loadDocument(t, `{"browsers": {"safari": {"releases": {"preview": {"status": "nightly"}}}}}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	release, ok := data.Releases["safari"]["preview"]
	if !ok || !release.Date.IsZero() || release.Status != "nightly" {
		t.Fatalf("expected an undated nightly release, got %+v", release)
	}
}

func TestLoadAcceptsAnEmptyDocument(t *testing.T) {
	data, err := loadDocument(t, `{}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(data.Features) != 0 || len(data.Releases) != 0 {
		t.Fatalf("expected nothing to be loaded, got %d features and %d browsers", len(data.Features), len(data.Releases))
	}
	if data.Has("css.properties.width") || data.Renderable("css.properties.width") || data.Latest("chrome") != "" {
		t.Fatal("expected an empty answer from empty data")
	}
}

func TestLoadRejectsMalformedData(t *testing.T) {
	cases := []struct {
		name     string
		document string
		want     string
	}{
		{"a document that is not JSON", `{`, "data.json"},
		{"properties that are not an object", `{"css": {"properties": []}}`, "data.json"},
		{"a feature that is not an object", `{"css": {"properties": {"x": 5}}}`, "css.properties.x"},
		{"a nested feature that is not an object", `{"css": {"selectors": {"x": {"y": "z"}}}}`, "css.selectors.x.y"},
		{"a compat that is not an object", `{"css": {"properties": {"x": {"__compat": 5}}}}`, "css.properties.x"},
		{"support that is a string", propertyDocument(`"mirror"`), "css.properties.x"},
		{"a version that is a number", propertyDocument(`{"version_added": 5}`), "css.properties.x"},
		{"a version that is not a version", propertyDocument(`{"version_added": "soon"}`), "css.properties.x"},
		{"an empty version", propertyDocument(`{"version_added": ""}`), "css.properties.x"},
		{"an alternative name that is a number", propertyDocument(`{"version_added": "1", "alternative_name": 5}`), "css.properties.x"},
		{"a removal that is a number", propertyDocument(`{"version_added": "1", "version_removed": 5}`), "css.properties.x"},
		{"a removal that is not a version", propertyDocument(`{"version_added": "1", "version_removed": "later"}`), "css.properties.x"},
		{"a browser that is not an object", `{"browsers": {"chrome": []}}`, "browsers.chrome"},
		{"a release date that is not a date", `{"browsers": {"chrome": {"releases": {"130": {"release_date": "yesterday"}}}}}`, "browsers.chrome.releases.130"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := loadDocument(t, tc.document)
			if err == nil {
				t.Fatalf("expected an error, got %+v", data)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected the error to mention %q, got %v", tc.want, err)
			}
		})
	}
}

func TestLoadIgnoresMalformedDataOfBrowsersItDoesNotConsider(t *testing.T) {
	data, err := loadDocument(t, `{
		"browsers": {"ie": 5},
		"css": {"properties": {"x": {"__compat": {"support": {"ie": 5, "chrome": {"version_added": "1"}}}}}}
	}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got := data.Features["css.properties.x"].Support["chrome"]; !slices.Equal(got, []Statement{{VersionAdded: "1", Added: true}}) {
		t.Fatalf("expected the considered browser to load, got %+v", got)
	}
}

func TestLoadRejectsAMissingFile(t *testing.T) {
	if _, err := Load(t.TempDir()); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected %v, got %v", fs.ErrNotExist, err)
	}
}
