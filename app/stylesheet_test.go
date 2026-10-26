//go:build !wasm

package app_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
	"github.com/Toms223/WebGo/supports"

	"github.com/Toms223/WebGo/app"
)

func whenTrue(app.Action) bool { return true }

func interFace() *style.FontFaceRule {
	return style.FontFace(css.FaceFontFamily(css.String("Inter")), css.Src(css.Url("/web/inter.woff2"), css.Format(kw.Woff2)))
}

func accentProperty() *style.PropertyRule {
	return style.Property("--accent", css.Syntax(css.SyntaxColor), css.Inherits(kw.True), css.InitialValue(css.Hex("#FF3B30")))
}

func plainRule(name string) *style.Rule {
	return style.Class(name).Set(css.Custom("--w", css.Px(1)))
}

func handlerFor(t *testing.T, styles ...style.Style) *app.Handler {
	t.Helper()
	return &app.Handler{Styles: styles, WebDir: t.TempDir()}
}

func putFile(t *testing.T, file, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, file string) string {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func modTime(t *testing.T, file string) time.Time {
	t.Helper()
	info, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}
	return info.ModTime()
}

func namesIn(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

type buildCase struct {
	name   string
	styles []style.Style
	files  []string
	want   error
	links  []string
}

func runBuildCases(t *testing.T, cases []buildCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := handlerFor(t, c.styles...)
			for _, f := range c.files {
				putFile(t, filepath.Join(h.WebDir, f), "")
			}
			err := h.Build()
			if !errors.Is(err, c.want) {
				t.Fatalf("Build() error = %v, want %v", err, c.want)
			}
			if c.want == nil && !slices.Equal(h.StyleLinks(), c.links) {
				t.Errorf("Styles = %q, want %q", h.StyleLinks(), c.links)
			}
		})
	}
}

func TestStaticVerifiesFile(t *testing.T) {
	h := handlerFor(t, style.Static("/web/app.css"))
	err := h.Build()
	if !errors.Is(err, app.ErrMissingStyle) {
		t.Fatalf("Build() error = %v, want %v", err, app.ErrMissingStyle)
	}
	if !strings.Contains(err.Error(), "app.css") {
		t.Errorf("Build() error = %q, want it to name app.css", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Build() error = %v, want it to keep the cause %v", err, os.ErrNotExist)
	}

	putFile(t, filepath.Join(h.WebDir, "app.css"), "")
	err = h.Build()
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	if want := []string{"/web/app.css"}; !slices.Equal(h.StyleLinks(), want) {
		t.Errorf("Styles = %q, want %q", h.StyleLinks(), want)
	}
}

func TestStaticURLPassesThrough(t *testing.T) {
	links := []string{"https://cdn/x.css", "http://cdn/x.css", "//cdn/x.css", "https://cdn/a/../x.css"}
	for _, link := range links {
		h := handlerFor(t, style.Static(link))
		err := h.Build()
		if err != nil {
			t.Errorf("Build() of %q error = %v, want nil", link, err)
			continue
		}
		if want := []string{link}; !slices.Equal(h.StyleLinks(), want) {
			t.Errorf("Styles = %q, want %q", h.StyleLinks(), want)
		}
		if got := namesIn(t, h.WebDir); len(got) != 0 {
			t.Errorf("web directory of %q holds %q, want it untouched", link, got)
		}
	}
}

func TestLinksAreCleaned(t *testing.T) {
	spellings := []string{"/web/app.css", "/web/./app.css", "/web//app.css", "/web/x/../app.css", "/web/app.css/"}
	for _, spelling := range spellings {
		fromStatic := handlerFor(t, style.Static(spelling))
		putFile(t, filepath.Join(fromStatic.WebDir, "app.css"), "")
		fromGenerated := handlerFor(t, style.Generated(spelling, plainRule("a")))
		for _, h := range []*app.Handler{fromStatic, fromGenerated} {
			err := h.Build()
			if err != nil {
				t.Errorf("Build() of %q error = %v, want nil", spelling, err)
				continue
			}
			if want := []string{"/web/app.css"}; !slices.Equal(h.StyleLinks(), want) {
				t.Errorf("Styles of %q = %q, want %q", spelling, h.StyleLinks(), want)
			}
		}
		if got, want := readFile(t, filepath.Join(fromGenerated.WebDir, "app.css")), ".a {\n\t--w: 1px;\n}\n"; got != want {
			t.Errorf("app.css generated from %q = %q, want %q", spelling, got, want)
		}
	}
}

func TestPathsOutsideWebAreRejected(t *testing.T) {
	site := filepath.Join(t.TempDir(), "site")
	webDir := filepath.Join(site, "web")
	putFile(t, filepath.Join(webDir, "app.css"), "")
	putFile(t, filepath.Join(site, "x.css"), "")
	putFile(t, filepath.Join(site, "secret.css"), "")

	cases := []struct {
		name  string
		style style.Style
		want  error
	}{
		{"static outside web", style.Static("/other/x.css"), app.ErrPathOutsideWeb},
		{"static escaping web", style.Static("/web/../x.css"), app.ErrPathOutsideWeb},
		{"static escaping to an existing file", style.Static("/web/../secret.css"), app.ErrPathOutsideWeb},
		{"static escaping through a subdirectory", style.Static("/web/a/../../x.css"), app.ErrPathOutsideWeb},
		{"static without leading slash", style.Static("web/app.css"), app.ErrPathOutsideWeb},
		{"static sharing only a name prefix", style.Static("/webapp/app.css"), app.ErrPathOutsideWeb},
		{"static naming the web directory", style.Static("/web/"), app.ErrPathOutsideWeb},
		{"static with no path", style.Static(""), app.ErrPathOutsideWeb},
		{"generated escaping web", style.Generated("/web/../../x.css", plainRule("a")), app.ErrPathOutsideWeb},
		{"generated outside web", style.Generated("/other/theme.css", plainRule("a")), app.ErrPathOutsideWeb},
		{"generated at a url", style.Generated("https://cdn/theme.css", plainRule("a")), app.ErrPathOutsideWeb},
		{"generated not css", style.Generated("/web/theme.txt", plainRule("a")), app.ErrNotCSSPath},
		{"generated without extension", style.Generated("/web/theme", plainRule("a")), app.ErrNotCSSPath},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := &app.Handler{Styles: []style.Style{c.style}, WebDir: webDir}
			if err := h.Build(); !errors.Is(err, c.want) {
				t.Errorf("Build() error = %v, want %v", err, c.want)
			}
		})
	}

	if got, want := namesIn(t, filepath.Dir(site)), []string{"site"}; !slices.Equal(got, want) {
		t.Errorf("directory above the site holds %q, want %q", got, want)
	}
	if got, want := namesIn(t, site), []string{"secret.css", "web", "x.css"}; !slices.Equal(got, want) {
		t.Errorf("site directory holds %q, want %q", got, want)
	}
	if got, want := namesIn(t, webDir), []string{"app.css"}; !slices.Equal(got, want) {
		t.Errorf("web directory holds %q, want %q", got, want)
	}
}

func TestGeneratedWritesDeterministically(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/t/theme.css", style.Class("a").Set(css.Custom("--w", css.Px(1)))))
	if err := h.Build(); err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	file := filepath.Join(h.WebDir, "t", "theme.css")
	wantCSS := ".a {\n\t--w: 1px;\n}\n"
	if got := readFile(t, file); got != wantCSS {
		t.Fatalf("theme.css = %q, want %q", got, wantCSS)
	}
	first := modTime(t, file)

	time.Sleep(20 * time.Millisecond)
	err := h.Build()
	if err != nil {
		t.Fatalf("second Build() error = %v, want nil", err)
	}
	if got := readFile(t, file); got != wantCSS {
		t.Errorf("theme.css after a second Build() = %q, want %q", got, wantCSS)
	}
	if second := modTime(t, file); !second.Equal(first) {
		t.Errorf("ModTime after a second Build() = %v, want %v", second, first)
	}
	if want := []string{"/web/t/theme.css"}; !slices.Equal(h.StyleLinks(), want) {
		t.Errorf("Styles = %q, want %q", h.StyleLinks(), want)
	}
}

func TestEmptyGeneratedStylesheetIsStillWritten(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/none.css"), style.Generated("/web/bare.css", style.Class("doc")))
	err := h.Build()
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	for _, name := range []string{"none.css", "bare.css"} {
		if got := readFile(t, filepath.Join(h.WebDir, name)); got != "" {
			t.Errorf("%s = %q, want an empty file", name, got)
		}
	}
	if want := []string{"/web/none.css", "/web/bare.css"}; !slices.Equal(h.StyleLinks(), want) {
		t.Errorf("Styles = %q, want %q", h.StyleLinks(), want)
	}
}

func TestGeneratedValidation(t *testing.T) {
	runBuildCases(t, []buildCase{
		{
			name:   "rule with On",
			styles: []style.Style{style.Generated("/web/t.css", plainRule("a").On("x", whenTrue))},
			want:   style.ErrDynamicInGenerated,
		},
		{
			name:   "unnamed class",
			styles: []style.Style{style.Generated("/web/t.css", style.Class(""))},
			want:   style.ErrNoName,
		},
		{
			name:   "unnamed id",
			styles: []style.Style{style.Generated("/web/t.css", style.Id(""))},
			want:   style.ErrNoName,
		},
		{
			name:   "unnamed element",
			styles: []style.Style{style.Generated("/web/t.css", style.Element(""))},
			want:   style.ErrNoName,
		},
		{
			name:   "two elements joined",
			styles: []style.Style{style.Generated("/web/t.css", style.Element("a").And(style.Element("b")))},
			want:   style.ErrBadSelector,
		},
		{
			name:   "bad media query",
			styles: []style.Style{style.Generated("/web/t.css", style.Class("a").Media(media.Screen().And(media.Print())))},
			want:   style.ErrBadMediaQuery,
		},
		{
			name:   "binding without an action",
			styles: []style.Style{style.Generated("/web/t.css", plainRule("a").On("", whenTrue))},
			want:   style.ErrBadBinding,
		},
		{
			name:   "invalid binding beside a valid one",
			styles: []style.Style{style.Generated("/web/t.css", plainRule("a").On("", whenTrue).On("x", whenTrue))},
			want:   style.ErrBadBinding,
		},
		{
			name:   "same selector twice",
			styles: []style.Style{style.Generated("/web/t.css", plainRule("a"), plainRule("a"))},
			want:   style.ErrDuplicateRule,
		},
		{
			name:   "same selector twice without properties",
			styles: []style.Style{style.Generated("/web/t.css", style.Class("a"), style.Class("a"))},
			want:   style.ErrDuplicateRule,
		},
		{
			name:   "same selector in two stylesheets",
			styles: []style.Style{style.Generated("/web/one.css", plainRule("a")), style.Generated("/web/two.css", plainRule("a"))},
			links:  []string{"/web/one.css", "/web/two.css"},
		},
		{
			name:   "different selectors",
			styles: []style.Style{style.Generated("/web/t.css", plainRule("a"), plainRule("b"), style.Class("a").Descendant(style.Class("b")))},
			links:  []string{"/web/t.css"},
		},
	})
}

func TestDynamicValidation(t *testing.T) {
	runBuildCases(t, []buildCase{
		{
			name:   "rule without On",
			styles: []style.Style{style.Dynamic(plainRule("a"))},
			want:   style.ErrNotDynamic,
		},
		{
			name:   "one rule without On among others",
			styles: []style.Style{style.Dynamic(plainRule("a").On("x", whenTrue), plainRule("b"))},
			want:   style.ErrNotDynamic,
		},
		{
			name:   "unnamed class",
			styles: []style.Style{style.Dynamic(style.Class("").On("x", whenTrue))},
			want:   style.ErrNoName,
		},
		{
			name:   "binding without an action",
			styles: []style.Style{style.Dynamic(plainRule("a").On("", whenTrue))},
			want:   style.ErrBadBinding,
		},
		{
			name:   "binding without a condition",
			styles: []style.Style{style.Dynamic(plainRule("a").On("x", nil))},
			want:   style.ErrBadBinding,
		},
		{
			name:   "rule with On",
			styles: []style.Style{style.Dynamic(plainRule("a").On("x", whenTrue))},
		},
		{
			name:   "no rules",
			styles: []style.Style{style.Dynamic()},
		},
		{
			name:   "same selector twice",
			styles: []style.Style{style.Dynamic(plainRule("a").On("x", whenTrue), plainRule("a").On("y", whenTrue))},
		},
		{
			name: "selector shared with a generated stylesheet",
			styles: []style.Style{
				style.Generated("/web/t.css", plainRule("a")),
				style.Dynamic(plainRule("a").On("x", whenTrue)),
			},
			links: []string{"/web/t.css"},
		},
	})
}

func TestDuplicatePaths(t *testing.T) {
	runBuildCases(t, []buildCase{
		{
			name:   "static twice",
			styles: []style.Style{style.Static("/web/a.css"), style.Static("/web/a.css")},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "static then generated",
			styles: []style.Style{style.Static("/web/a.css"), style.Generated("/web/a.css", plainRule("a"))},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "generated then static",
			styles: []style.Style{style.Generated("/web/a.css", plainRule("a")), style.Static("/web/a.css")},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "generated twice",
			styles: []style.Style{style.Generated("/web/a.css", plainRule("a")), style.Generated("/web/a.css", plainRule("b"))},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "same file spelled with a dot segment",
			styles: []style.Style{style.Static("/web/a.css"), style.Static("/web/./a.css")},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "same file spelled through a parent",
			styles: []style.Style{style.Static("/web/a.css"), style.Generated("/web/x/../a.css", plainRule("a"))},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "same url twice",
			styles: []style.Style{style.Static("https://cdn/x.css"), style.Static("https://cdn/x.css")},
			want:   app.ErrDuplicatePath,
		},
		{
			name:   "different files",
			styles: []style.Style{style.Static("/web/a.css"), style.Static("/web/b.css"), style.Generated("/web/c.css", plainRule("a"))},
			files:  []string{"a.css", "b.css"},
			links:  []string{"/web/a.css", "/web/b.css", "/web/c.css"},
		},
		{
			name:   "same name in different directories",
			styles: []style.Style{style.Generated("/web/x/a.css", plainRule("a")), style.Generated("/web/y/a.css", plainRule("a"))},
			links:  []string{"/web/x/a.css", "/web/y/a.css"},
		},
		{
			name:   "url and file with the same name",
			styles: []style.Style{style.Static("https://cdn/a.css"), style.Static("/web/a.css")},
			files:  []string{"a.css"},
			links:  []string{"https://cdn/a.css", "/web/a.css"},
		},
	})
}

func TestBuildIsIdempotent(t *testing.T) {
	h := &app.Handler{
		Styles: []style.Style{style.Static("/web/app.css")},
		WebDir: t.TempDir(),
	}
	putFile(t, filepath.Join(h.WebDir, "app.css"), "")

	want := []string{"/web/app.css"}
	for call := 1; call <= 3; call++ {
		if err := h.Build(); err != nil {
			t.Fatalf("Build() call %d error = %v, want nil", call, err)
		}
		if !slices.Equal(h.StyleLinks(), want) {
			t.Errorf("links after call %d = %q, want %q", call, h.StyleLinks(), want)
		}
	}

	h.Styles = nil
	if err := h.Build(); err != nil {
		t.Fatalf("Build() without styles error = %v, want nil", err)
	}
	if got := h.StyleLinks(); len(got) != 0 {
		t.Errorf("links without styles = %q, want none", got)
	}
}

func TestDefaultWebDir(t *testing.T) {
	h := &app.Handler{Styles: []style.Style{style.Static("/web/app.css")}}
	err := h.Build()
	if !errors.Is(err, app.ErrMissingStyle) {
		t.Fatalf("Build() error = %v, want %v", err, app.ErrMissingStyle)
	}
	relative := filepath.Join("web", "app.css")
	if got := err.Error(); !strings.Contains(got, relative) || strings.Contains(got, "/"+relative) {
		t.Errorf("Build() error = %q, want it to name the relative path %q", got, relative)
	}
}

func TestPathErrorsNameThePath(t *testing.T) {
	cases := []struct {
		name   string
		styles []style.Style
		want   error
		path   string
	}{
		{"outside web", []style.Style{style.Static("/other/x.css")}, app.ErrPathOutsideWeb, "/other/x.css"},
		{"escaping web", []style.Style{style.Static("/web/../x.css")}, app.ErrPathOutsideWeb, "/web/../x.css"},
		{"generated outside web", []style.Style{style.Generated("/other/t.css", plainRule("a"))}, app.ErrPathOutsideWeb, "/other/t.css"},
		{"not css", []style.Style{style.Generated("/web/theme.txt", plainRule("a"))}, app.ErrNotCSSPath, "/web/theme.txt"},
		{"duplicate url", []style.Style{style.Static("https://cdn/x.css"), style.Static("https://cdn/x.css")}, app.ErrDuplicatePath, "https://cdn/x.css"},
		{"duplicate file", []style.Style{style.Static("/web/a.css"), style.Static("/web/./a.css")}, app.ErrDuplicatePath, "/web/a.css"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := handlerFor(t, c.styles...)
			err := h.Build()
			if !errors.Is(err, c.want) {
				t.Fatalf("Build() error = %v, want %v", err, c.want)
			}
			if want := c.want.Error() + ": " + c.path; err.Error() != want {
				t.Errorf("Build() error = %q, want %q", err, want)
			}
		})
	}
}

func TestRuleErrorsNameTheSelector(t *testing.T) {
	unnamed := style.Class("")
	clash := style.Element("a").And(style.Element("b"))
	badQuery := style.Class("a").Media(media.Screen().And(media.Print()))
	unbound := plainRule("a").On("", whenTrue)
	live := plainRule("live").On("x", whenTrue)
	still := plainRule("still")
	twin := plainRule("twin")

	cases := []struct {
		name  string
		style style.Style
		want  error
		rule  *style.Rule
	}{
		{"unnamed", style.Generated("/web/t.css", unnamed), style.ErrNoName, unnamed},
		{"unjoinable", style.Generated("/web/t.css", clash), style.ErrBadSelector, clash},
		{"bad query", style.Generated("/web/t.css", badQuery), style.ErrBadMediaQuery, badQuery},
		{"bad binding", style.Dynamic(unbound), style.ErrBadBinding, unbound},
		{"dynamic in generated", style.Generated("/web/t.css", live), style.ErrDynamicInGenerated, live},
		{"duplicate", style.Generated("/web/t.css", twin, twin), style.ErrDuplicateRule, twin},
		{"not dynamic", style.Dynamic(still), style.ErrNotDynamic, still},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := handlerFor(t, c.style)
			err := h.Build()
			if !errors.Is(err, c.want) {
				t.Fatalf("Build() error = %v, want %v", err, c.want)
			}
			if want := c.want.Error() + ": " + c.rule.Selector(); err.Error() != want {
				t.Errorf("Build() error = %q, want %q", err, want)
			}
		})
	}
}

func TestUnnamedClassMessageHasTheSelector(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/t.css", style.Class("")))
	err := h.Build()
	if want := style.ErrNoName.Error() + ": ."; err == nil || err.Error() != want {
		t.Errorf("Build() error = %v, want %q", err, want)
	}
}

func TestInvalidHandlerWritesNothing(t *testing.T) {
	cases := []struct {
		name  string
		later style.Style
		want  error
	}{
		{"unnamed rule", style.Generated("/web/bad.css", style.Class("")), style.ErrNoName},
		{"missing static file", style.Static("/web/missing.css"), app.ErrMissingStyle},
		{"duplicate path", style.Generated("/web/ok.css", plainRule("b")), app.ErrDuplicatePath},
		{"outside web", style.Static("/other/x.css"), app.ErrPathOutsideWeb},
		{"dynamic rule without On", style.Dynamic(plainRule("c")), style.ErrNotDynamic},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			h := handlerFor(t, style.Generated("/web/t/ok.css", plainRule("a")), style.Generated("/web/ok.css", plainRule("a")), c.later)
			err := h.Build()
			if !errors.Is(err, c.want) {
				t.Fatalf("Build() error = %v, want %v", err, c.want)
			}
			if got := namesIn(t, h.WebDir); len(got) != 0 {
				t.Errorf("web directory holds %q after a failed Build(), want it empty", got)
			}
			if got := h.StyleLinks(); len(got) != 0 {
				t.Errorf("links after a failed Build() = %q, want none", got)
			}
		})
	}
}

func TestWriteFailureIsReported(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/blocker/theme.css", plainRule("a")))
	putFile(t, filepath.Join(h.WebDir, "blocker"), "")

	err := h.Build()
	if err == nil {
		t.Fatal("Build() error = nil, want the failed write")
	}
	if !strings.HasPrefix(err.Error(), "app: ") {
		t.Errorf("Build() error = %q, want it to start with %q", err, "app: ")
	}
	if got := h.StyleLinks(); len(got) != 0 {
		t.Errorf("links after a failed Build() = %q, want none", got)
	}
}

func TestNilStylesAndRulesAreIgnored(t *testing.T) {
	var none *style.Rule
	h := handlerFor(t,
		nil,
		style.Generated("/web/t.css", none, plainRule("a"), none),
		style.Dynamic(none, plainRule("b").On("x", whenTrue)),
		style.Dynamic(none),
	)
	err := h.Build()
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	if want := []string{"/web/t.css"}; !slices.Equal(h.StyleLinks(), want) {
		t.Errorf("Styles = %q, want %q", h.StyleLinks(), want)
	}
	if got, want := readFile(t, filepath.Join(h.WebDir, "t.css")), ".a {\n\t--w: 1px;\n}\n"; got != want {
		t.Errorf("t.css = %q, want %q", got, want)
	}
}

func TestStylesKeepTheirOwnRules(t *testing.T) {
	rules := []style.Item{plainRule("a"), plainRule("b")}
	generatedStyle := style.Generated("/web/t.css", rules...)
	rules[0], rules[1] = rules[1], rules[0]

	h := handlerFor(t, generatedStyle)
	if err := h.Build(); err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	if got, want := readFile(t, filepath.Join(h.WebDir, "t.css")), ".a {\n\t--w: 1px;\n}\n\n.b {\n\t--w: 1px;\n}\n"; got != want {
		t.Errorf("t.css = %q, want %q", got, want)
	}
}

func TestWebPath(t *testing.T) {
	cases := []struct {
		given string
		want  string
	}{
		{"/web/app.css", filepath.Join("site", "app.css")},
		{"/web/a/b.css", filepath.Join("site", "a", "b.css")},
		{"/web/a/../b.css", filepath.Join("site", "b.css")},
		{"/web/./b.css", filepath.Join("site", "b.css")},
		{"/web//b.css", filepath.Join("site", "b.css")},
		{"/web/b.css/", filepath.Join("site", "b.css")},
		{"/web/../web/b.css", filepath.Join("site", "b.css")},
		{"/other/../web/b.css", ""},
		{"/web/a..b.css", filepath.Join("site", "a..b.css")},
		{"/web/..a/b.css", filepath.Join("site", "..a", "b.css")},
		{"/web/..", ""},
		{"/web/../x.css", ""},
		{"/web/a/../../x.css", ""},
		{"/web/../../x.css", ""},
		{"/web/a/../..", ""},
		{"/web", ""},
		{"/web/", ""},
		{"/web/.", ""},
		{"/webapp/b.css", ""},
		{"web/b.css", ""},
		{"/Web/b.css", ""},
		{"/other/b.css", ""},
		{"https://cdn/b.css", ""},
		{"//cdn/b.css", ""},
		{"/", ""},
		{"", ""},
	}
	for _, c := range cases {
		got, err := app.WebPath("site", c.given)
		if c.want == "" {
			if !errors.Is(err, app.ErrPathOutsideWeb) {
				t.Errorf("app.WebPath(%q) error = %v, want %v", c.given, err, app.ErrPathOutsideWeb)
			}
			if got != "" {
				t.Errorf("app.WebPath(%q) = %q alongside an error, want empty", c.given, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("app.WebPath(%q) error = %v, want nil", c.given, err)
		}
		if got != c.want {
			t.Errorf("app.WebPath(%q) = %q, want %q", c.given, got, c.want)
		}
	}
}

func TestIsURL(t *testing.T) {
	cases := []struct {
		given string
		want  bool
	}{
		{"http://cdn/x.css", true},
		{"https://cdn/x.css", true},
		{"//cdn/x.css", true},
		{"/web/x.css", false},
		{"web/x.css", false},
		{"/web//x.css", false},
		{"https:/cdn/x.css", false},
		{"https//cdn/x.css", false},
		{"ftp://cdn/x.css", false},
		{"httpx://cdn/x.css", false},
		{"x.css", false},
		{"/", false},
		{"", false},
	}
	for _, c := range cases {
		if got := app.IsURL(c.given); got != c.want {
			t.Errorf("app.IsURL(%q) = %v, want %v", c.given, got, c.want)
		}
	}
}

func TestWriteIfChanged(t *testing.T) {
	file := filepath.Join(t.TempDir(), "a", "b", "c.css")
	if err := app.WriteIfChanged(file, []byte("one")); err != nil {
		t.Fatalf("first write error = %v, want nil", err)
	}
	if got := readFile(t, file); got != "one" {
		t.Fatalf("file = %q, want %q", got, "one")
	}

	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(file, past, past); err != nil {
		t.Fatal(err)
	}
	pinned := modTime(t, file)

	if err := app.WriteIfChanged(file, []byte("one")); err != nil {
		t.Fatalf("unchanged write error = %v, want nil", err)
	}
	if got := modTime(t, file); !got.Equal(pinned) {
		t.Errorf("ModTime after identical bytes = %v, want %v", got, pinned)
	}

	if err := app.WriteIfChanged(file, []byte("two")); err != nil {
		t.Fatalf("changed write error = %v, want nil", err)
	}
	if got := readFile(t, file); got != "two" {
		t.Errorf("file after changed bytes = %q, want %q", got, "two")
	}
	if got := modTime(t, file); !got.After(pinned) {
		t.Errorf("ModTime after changed bytes = %v, want it after %v", got, pinned)
	}

	if err := app.WriteIfChanged(file, nil); err != nil {
		t.Fatalf("emptying write error = %v, want nil", err)
	}
	if got := readFile(t, file); got != "" {
		t.Errorf("file after empty bytes = %q, want empty", got)
	}
}

func TestWriteIfChangedReportsFailures(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	putFile(t, blocker, "")
	if err := app.WriteIfChanged(filepath.Join(blocker, "x.css"), []byte("x")); err == nil {
		t.Error("write below a regular file error = nil, want an error")
	}
	if err := app.WriteIfChanged(dir, []byte("x")); err == nil {
		t.Error("write onto a directory error = nil, want an error")
	}
}

func TestGeneratedAcceptsItems(t *testing.T) {
	items := []style.Item{
		plainRule("a"),
		style.Keyframes("pulse").From(css.Opacity(css.Num(0))).To(css.Opacity(css.Num(1))),
		interFace(),
		accentProperty(),
		style.Import(css.Url("/web/base.css"), style.InLayer("base")),
		style.Layer("components", plainRule("b"), style.Layer("inner", plainRule("c"))),
		style.LayerOrder("base", "components"),
	}
	h := handlerFor(t, style.Generated("/web/t.css", items...))
	err := h.Build()
	if err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	if want := []string{"/web/t.css"}; !slices.Equal(h.StyleLinks(), want) {
		t.Errorf("Styles = %q, want %q", h.StyleLinks(), want)
	}
	got := readFile(t, filepath.Join(h.WebDir, "t.css"))
	source, err := style.Resolve(style.Generated("/web/t.css", items...))
	if err != nil {
		t.Fatal(err)
	}
	if want := string(source.Content); got != want {
		t.Errorf("t.css = %q, want %q", got, want)
	}
	if want := "@layer base, components;\n\n@import url(\"/web/base.css\") layer(base);\n\n.a {"; !strings.HasPrefix(got, want) {
		t.Errorf("t.css = %q, want it to start with %q", got, want)
	}
}

func TestGeneratedTakesRulesAndItems(t *testing.T) {
	r1, r2 := plainRule("a"), plainRule("b")
	items := []style.Item{plainRule("a"), plainRule("b")}
	h := handlerFor(t, style.Generated("/web/rules.css", r1, r2), style.Generated("/web/items.css", items...))
	if err := h.Build(); err != nil {
		t.Fatalf("Build() error = %v, want nil", err)
	}
	want := ".a {\n\t--w: 1px;\n}\n\n.b {\n\t--w: 1px;\n}\n"
	for _, name := range []string{"rules.css", "items.css"} {
		if got := readFile(t, filepath.Join(h.WebDir, name)); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestDynamicRejectsFontFace(t *testing.T) {
	err := handlerFor(t, style.Dynamic(interFace())).Build()
	if !errors.Is(err, style.ErrNotDynamic) {
		t.Fatalf("Build() error = %v, want %v", err, style.ErrNotDynamic)
	}
	if want := style.ErrNotDynamic.Error() + ": @font-face"; err.Error() != want {
		t.Errorf("Build() error = %q, want %q", err, want)
	}
}

func TestGeneratedRejectsBoundKeyframes(t *testing.T) {
	err := handlerFor(t, style.Generated("/web/t.css", style.Keyframes("pulse").From(css.Opacity(css.Num(0))).On("x", whenTrue))).Build()
	if !errors.Is(err, style.ErrDynamicInGenerated) {
		t.Fatalf("Build() error = %v, want %v", err, style.ErrDynamicInGenerated)
	}
	if want := style.ErrDynamicInGenerated.Error() + ": @keyframes pulse"; err.Error() != want {
		t.Errorf("Build() error = %q, want %q", err, want)
	}
}

func TestItemValidation(t *testing.T) {
	pulse := func() *style.KeyframesRule { return style.Keyframes("pulse").From(css.Opacity(css.Num(0))) }
	runBuildCases(t, []buildCase{
		{
			name:   "same selector inside and outside a layer",
			styles: []style.Style{style.Generated("/web/t.css", plainRule("a"), style.Layer("base", plainRule("a")))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "same selector in two layers",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("one", plainRule("a")), style.Layer("two", plainRule("a")))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "same selector in a layer and its nested layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("one", plainRule("a"), style.Layer("inner", plainRule("a"))))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "same selector twice inside one layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", plainRule("a"), plainRule("a")))},
			want:   style.ErrDuplicateRule,
		},
		{
			name:   "same selector twice inside one nested layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", style.Layer("inner", plainRule("a"), plainRule("a"))))},
			want:   style.ErrDuplicateRule,
		},
		{
			name:   "bound rule inside a layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", plainRule("a").On("x", whenTrue)))},
			want:   style.ErrDynamicInGenerated,
		},
		{
			name:   "bound keyframes inside a layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", pulse().On("x", whenTrue)))},
			want:   style.ErrDynamicInGenerated,
		},
		{
			name:   "bad rule inside a layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", style.Class("")))},
			want:   style.ErrNoName,
		},
		{
			name:   "import inside a layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", style.Import(css.Url("x.css"))))},
			want:   style.ErrBadAtRule,
		},
		{
			name:   "unnamed keyframes",
			styles: []style.Style{style.Generated("/web/t.css", style.Keyframes(""))},
			want:   style.ErrNoName,
		},
		{
			name:   "font face without src",
			styles: []style.Style{style.Generated("/web/t.css", style.FontFace(css.FaceFontFamily(css.String("Inter"))))},
			want:   style.ErrBadAtRule,
		},
		{
			name:   "property without syntax",
			styles: []style.Style{style.Generated("/web/t.css", style.Property("a", css.Inherits(kw.True)))},
			want:   style.ErrBadAtRule,
		},
		{
			name:   "import with a bad supports query",
			styles: []style.Style{style.Generated("/web/t.css", style.Import(css.Url("x.css"), style.IfSupports(supports.Query{})))},
			want:   style.ErrBadSupportsQuery,
		},
		{
			name:   "empty layer order",
			styles: []style.Style{style.Generated("/web/t.css", style.LayerOrder())},
			want:   style.ErrNoName,
		},
		{
			name:   "same keyframes twice",
			styles: []style.Style{style.Generated("/web/t.css", pulse(), pulse())},
			want:   style.ErrDuplicateRule,
		},
		{
			name:   "same keyframes twice inside one layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", pulse(), pulse()))},
			want:   style.ErrDuplicateRule,
		},
		{
			name:   "same keyframes inside and outside a layer",
			styles: []style.Style{style.Generated("/web/t.css", pulse(), style.Layer("base", pulse()))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "same keyframes in two layers",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("one", pulse()), style.Layer("two", pulse()))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "different keyframes and a class of the same name",
			styles: []style.Style{style.Generated("/web/t.css", pulse(), style.Keyframes("fade").From(css.Opacity(css.Num(0))), plainRule("pulse"))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "bound keyframes in dynamic",
			styles: []style.Style{style.Dynamic(pulse().On("x", whenTrue))},
		},
		{
			name:   "keyframes without On in dynamic",
			styles: []style.Style{style.Dynamic(pulse())},
			want:   style.ErrNotDynamic,
		},
		{
			name:   "unnamed bound keyframes in dynamic",
			styles: []style.Style{style.Dynamic(style.Keyframes("").On("x", whenTrue))},
			want:   style.ErrNoName,
		},
		{
			name:   "layer in dynamic",
			styles: []style.Style{style.Dynamic(style.Layer("base", plainRule("a").On("x", whenTrue)))},
			want:   style.ErrNotDynamic,
		},
		{
			name:   "property in dynamic",
			styles: []style.Style{style.Dynamic(accentProperty())},
			want:   style.ErrNotDynamic,
		},
		{
			name:   "import in dynamic",
			styles: []style.Style{style.Dynamic(style.Import(css.Url("x.css")))},
			want:   style.ErrNotDynamic,
		},
		{
			name:   "layer order in dynamic",
			styles: []style.Style{style.Dynamic(style.LayerOrder("base"))},
			want:   style.ErrNotDynamic,
		},
	})
}

func TestItemErrorsNameTheItem(t *testing.T) {
	clash := style.Element("a").And(style.Element("b"))
	cases := []struct {
		name  string
		style style.Style
		want  string
	}{
		{"font face", style.Generated("/web/t.css", style.FontFace()), style.ErrBadAtRule.Error() + ": @font-face"},
		{"property", style.Generated("/web/t.css", style.Property("accent")), style.ErrBadAtRule.Error() + ": @property --accent"},
		{"import", style.Generated("/web/t.css", style.Import(css.Url("x.css"), style.ForMedia(media.Screen().And(media.Print())))), style.ErrBadMediaQuery.Error() + ": @import url(\"x.css\")"},
		{"layer", style.Generated("/web/t.css", style.Layer("base", style.LayerOrder("a"))), style.ErrBadAtRule.Error() + ": @layer base"},
		{"rule in a layer", style.Generated("/web/t.css", style.Layer("base", clash)), style.ErrBadSelector.Error() + ": " + clash.Selector()},
		{"duplicate in a layer", style.Generated("/web/t.css", style.Layer("base", plainRule("a"), plainRule("a"))), style.ErrDuplicateRule.Error() + ": .a"},
		{"duplicate keyframes", style.Generated("/web/t.css", style.Keyframes("pulse").From(css.Opacity(css.Num(0))), style.Keyframes("pulse").To(css.Opacity(css.Num(1)))), style.ErrDuplicateRule.Error() + ": @keyframes pulse"},
		{"layer order", style.Generated("/web/t.css", style.LayerOrder("a", "")), style.ErrNoName.Error() + ": @layer a, "},
		{"layer in dynamic", style.Dynamic(style.Layer("base")), style.ErrNotDynamic.Error() + ": @layer base"},
		{"keyframes in dynamic", style.Dynamic(style.Keyframes("pulse")), style.ErrNotDynamic.Error() + ": @keyframes pulse"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := handlerFor(t, c.style).Build()
			if err == nil || err.Error() != c.want {
				t.Errorf("Build() error = %v, want %q", err, c.want)
			}
		})
	}
}

func TestDuplicateRuleIgnoresNestedChildren(t *testing.T) {
	runBuildCases(t, []buildCase{
		{
			name:   "nested and top-level selector",
			styles: []style.Style{style.Generated("/web/t.css", style.Class("card").Nest(plainRule("title")), plainRule("title"))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "same selector nested under two parents",
			styles: []style.Style{style.Generated("/web/t.css", style.Class("card").Nest(plainRule("title")), style.Class("panel").Nest(plainRule("title")))},
			links:  []string{"/web/t.css"},
		},
	})
}

func TestNestingValidation(t *testing.T) {
	child := plainRule("title")
	parent := style.Class("card").Nest(child)
	bound := style.Class("panel").Nest(plainRule("x")).On("x", whenTrue)
	boundChild := plainRule("y")
	boundParent := style.Class("bar").On("y", whenTrue).Nest(boundChild)
	runBuildCases(t, []buildCase{
		{
			name:   "nested child at top level",
			styles: []style.Style{style.Generated("/web/t.css", parent, child)},
			want:   style.ErrNested,
		},
		{
			name:   "nested child inside a layer",
			styles: []style.Style{style.Generated("/web/t.css", style.Layer("base", child))},
			want:   style.ErrNested,
		},
		{
			name:   "nested child in Dynamic",
			styles: []style.Style{style.Dynamic(boundParent, boundChild)},
			want:   style.ErrNested,
		},
		{
			name:   "style.Self with properties",
			styles: []style.Style{style.Generated("/web/t.css", style.Self().Set(css.Custom("--w", css.Px(1))))},
			want:   style.ErrBadSelector,
		},
		{
			name:   "style.Self with a pseudo block",
			styles: []style.Style{style.Generated("/web/t.css", style.Self().Hover(css.Custom("--w", css.Px(1))))},
			want:   style.ErrBadSelector,
		},
		{
			name:   "style.Self in Dynamic",
			styles: []style.Style{style.Dynamic(style.Self().Hover(css.Custom("--w", css.Px(1))).On("x", whenTrue))},
			want:   style.ErrBadSelector,
		},
		{
			name:   "nested binding",
			styles: []style.Style{style.Generated("/web/t.css", style.Class("card").Nest(plainRule("t").On("x", whenTrue)))},
			want:   style.ErrNestedBinding,
		},
		{
			name:   "parent with children",
			styles: []style.Style{style.Generated("/web/t.css", style.Class("card").Nest(plainRule("t")))},
			links:  []string{"/web/t.css"},
		},
		{
			name:   "bound parent with children",
			styles: []style.Style{style.Dynamic(bound)},
			links:  nil,
		},
	})
}

func TestNestingErrorsNameTheRule(t *testing.T) {
	child := plainRule("title")
	parent := style.Class("card").Nest(child)
	cases := []struct {
		name    string
		style   style.Style
		want    error
		subject string
	}{
		{"nested", style.Generated("/web/t.css", parent, child), style.ErrNested, ".title"},
		{"self", style.Generated("/web/t.css", style.Self().Hover(css.Custom("--w", css.Px(1)))), style.ErrBadSelector, "&"},
		{"nested binding", style.Generated("/web/t.css", style.Class("card").Nest(plainRule("t").On("x", whenTrue))), style.ErrNestedBinding, ".card"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := handlerFor(t, c.style).Build()
			if want := c.want.Error() + ": " + c.subject; err == nil || err.Error() != want {
				t.Errorf("Build() error = %v, want %q", err, want)
			}
		})
	}
}

func TestGeneratedWritesNestedRules(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/t.css",
		plainRule("a"),
		style.Class("card").Set(css.Custom("--w", css.Px(2))).Nest(plainRule("title")),
	))
	if err := h.Build(); err != nil {
		t.Fatal(err)
	}
	want := ".a {\n\t--w: 1px;\n}\n\n.card {\n\t--w: 2px;\n\t& .title {\n\t\t--w: 1px;\n\t}\n}\n"
	if got := readFile(t, filepath.Join(h.WebDir, "t.css")); got != want {
		t.Errorf("t.css = %q, want %q", got, want)
	}
}
