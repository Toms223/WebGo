package style

import (
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
)

type prefixedProperty struct{ css.Property }

func (prefixedProperty) Name() string       { return "backdrop-filter" }
func (prefixedProperty) Value() string      { return "blur(4px)" }
func (prefixedProperty) Prefixes() []string { return []string{"-webkit-"} }

func TestRenderJoinsRules(t *testing.T) {
	p1 := css.Custom("--one", css.Px(1))
	p2 := css.Custom("--ink", css.Hex("#111"))
	got := render([]Item{Class("a").Set(p1), Class("b").Set(p2)})
	if want := ".a {\n\t--one: 1px;\n}\n\n.b {\n\t--ink: #111;\n}\n"; got != want {
		t.Errorf("render() = %q, want %q", got, want)
	}

	got = render([]Item{
		Class("a").Set(p1).pseudo(":hover", []css.Property{p2}),
		Class("empty"),
		Class("b").Set(p2),
	})
	want := ".a {\n\t--one: 1px;\n}\n\n" +
		".a:hover {\n\t--ink: #111;\n}\n\n" +
		".b {\n\t--ink: #111;\n}\n"
	if got != want {
		t.Errorf("render() = %q, want %q", got, want)
	}
}

func TestPrefixedDeclarationsComeFirst(t *testing.T) {
	cases := []struct {
		got  string
		want string
	}{
		{
			renderDeclarations("backdrop-filter", "blur(4px)", []string{"-webkit-"}, 1),
			"\t-webkit-backdrop-filter: blur(4px);\n\tbackdrop-filter: blur(4px);\n",
		},
		{
			renderDeclarations("backdrop-filter", "blur(4px)", nil, 1),
			"\tbackdrop-filter: blur(4px);\n",
		},
		{
			renderDeclarations("user-select", "none", []string{"-webkit-", "-moz-"}, 0),
			"-webkit-user-select: none;\n-moz-user-select: none;\nuser-select: none;\n",
		},
		{
			Class("glass").Set(prefixedProperty{}).CSS(),
			".glass {\n\t-webkit-backdrop-filter: blur(4px);\n\tbackdrop-filter: blur(4px);\n}\n",
		},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

func TestRenderBlockIndents(t *testing.T) {
	props := []css.Property{css.Custom("--a", css.Px(1)), css.Custom("--b", css.Px(2))}
	if got, want := renderBlock(".a", props, 0), ".a {\n\t--a: 1px;\n\t--b: 2px;\n}"; got != want {
		t.Errorf("renderBlock() = %q, want %q", got, want)
	}
	if got, want := renderBlock(".a", nil, 1), "\t.a {\n\t}"; got != want {
		t.Errorf("renderBlock() = %q, want %q", got, want)
	}
	cases := []struct {
		prelude string
		want    string
	}{
		{"@media " + media.Screen().CSS(), "@media screen {\n\t.a {\n\t\t--a: 1px;\n\t\t--b: 2px;\n\t}\n}"},
		{"@container side", "@container side {\n\t.a {\n\t\t--a: 1px;\n\t\t--b: 2px;\n\t}\n}"},
		{"@supports (display: grid)", "@supports (display: grid) {\n\t.a {\n\t\t--a: 1px;\n\t\t--b: 2px;\n\t}\n}"},
	}
	for _, c := range cases {
		if got := renderConditional(c.prelude, ".a", props); got != c.want {
			t.Errorf("renderConditional(%q) = %q, want %q", c.prelude, got, c.want)
		}
	}
}

func TestWrap(t *testing.T) {
	if got, want := wrap("@media screen", "\t\t--a: 1px;\n", 1), "\t@media screen {\n\t\t--a: 1px;\n\t}\n"; got != want {
		t.Errorf("wrap() = %q, want %q", got, want)
	}
	if got := wrap("@media screen", "", 1); got != "" {
		t.Errorf("wrap() of an empty body = %q, want empty", got)
	}
}

func TestRenderOfNothing(t *testing.T) {
	if got := render(nil); got != "" {
		t.Errorf("render(nil) = %q, want empty", got)
	}
	if got := render([]Item{Class("a"), Class("b").Set()}); got != "" {
		t.Errorf("render() of rules without properties = %q, want empty", got)
	}
}

func TestRenderOrdersLayerOrdersThenImportsFirst(t *testing.T) {
	p := css.Custom("--w", css.Px(1))
	got := render([]Item{Class("a").Set(p), LayerOrder("base"), Import(css.Url("x.css"))})
	want := "@layer base;\n\n@import url(\"x.css\");\n\n.a {\n\t--w: 1px;\n}\n"
	if got != want {
		t.Errorf("render() = %q, want %q", got, want)
	}

	got = render([]Item{
		Class("a").Set(p),
		Import(css.Url("one.css")),
		LayerOrder("base"),
		Keyframes("k").From(p),
		Import(css.Url("two.css")),
		LayerOrder("components"),
		Class("b").Set(p),
	})
	want = "@layer base;\n\n" +
		"@layer components;\n\n" +
		"@import url(\"one.css\");\n\n" +
		"@import url(\"two.css\");\n\n" +
		".a {\n\t--w: 1px;\n}\n\n" +
		"@keyframes k {\n\tfrom {\n\t\t--w: 1px;\n\t}\n}\n\n" +
		".b {\n\t--w: 1px;\n}\n"
	if got != want {
		t.Errorf("render() = %q, want %q", got, want)
	}
}

func TestIndent(t *testing.T) {
	cases := []struct {
		text  string
		depth int
		want  string
	}{
		{".a {\n\t--w: 1px;\n}", 1, "\t.a {\n\t\t--w: 1px;\n\t}"},
		{".a {\n}\n\n.b {\n}", 2, "\t\t.a {\n\t\t}\n\n\t\t.b {\n\t\t}"},
		{".a", 0, ".a"},
		{"", 1, ""},
	}
	for _, c := range cases {
		if got := indent(c.text, c.depth); got != c.want {
			t.Errorf("indent(%q, %d) = %q, want %q", c.text, c.depth, got, c.want)
		}
	}
}

func TestRenderDeclarationsOf(t *testing.T) {
	descs := []css.Descriptor{css.Syntax(css.SyntaxColor), css.Inherits(kw.True)}
	if got, want := renderDeclarationsOf(descs, 1), "\tsyntax: \"<color>\";\n\tinherits: true;\n"; got != want {
		t.Errorf("renderDeclarationsOf() = %q, want %q", got, want)
	}
	if got := renderDeclarationsOf(nil, 1); got != "" {
		t.Errorf("renderDeclarationsOf(nil) = %q, want empty", got)
	}
}
