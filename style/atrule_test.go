package style

import (
	"errors"
	"slices"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/supports"

	"github.com/Toms223/WebGo/app"
)

func interFace() *FontFaceRule {
	return FontFace(css.FaceFontFamily(css.String("Inter")), css.Src(css.Url("/web/inter.woff2"), css.Format(kw.Woff2)))
}

func accentProperty() *PropertyRule {
	return Property("--accent", css.Syntax(css.SyntaxColor), css.Inherits(kw.True), css.InitialValue(css.Hex("#FF3B30")))
}

type errorCase struct {
	name string
	err  error
	want error
}

func checkErrors(t *testing.T, cases []errorCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if !errors.Is(c.err, c.want) {
				t.Errorf("Err() = %v, want %v", c.err, c.want)
			}
		})
	}
}

func TestKeyframesCSS(t *testing.T) {
	pulse := Keyframes("pulse").At(css.Pct(0), css.Opacity(css.Num(1))).At(css.Pct(100), css.Opacity(css.Num(0.4)))
	want := "@keyframes pulse {\n\t0% {\n\t\topacity: 1;\n\t}\n\t100% {\n\t\topacity: 0.4;\n\t}\n}\n"
	if got := pulse.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if got := pulse.Name(); got != "pulse" {
		t.Errorf("Name() = %q, want %q", got, "pulse")
	}

	fade := Keyframes("fade").From(css.Custom("--o", css.Num(0))).To(css.Custom("--o", css.Num(1)))
	want = "@keyframes fade {\n\tfrom {\n\t\t--o: 0;\n\t}\n\tto {\n\t\t--o: 1;\n\t}\n}\n"
	if got := fade.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	twice := Keyframes("twice").At(css.Pct(50), css.Custom("--a", css.Px(1))).At(css.Pct(50), css.Custom("--b", css.Px(2)))
	want = "@keyframes twice {\n\t50% {\n\t\t--a: 1px;\n\t}\n\t50% {\n\t\t--b: 2px;\n\t}\n}\n"
	if got := twice.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	sparse := Keyframes("sparse").From().At(css.Pct(50), nil).To(css.Custom("--o", css.Num(1)))
	want = "@keyframes sparse {\n\tto {\n\t\t--o: 1;\n\t}\n}\n"
	if got := sparse.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	if got := Keyframes("none").From().CSS(); got != "" {
		t.Errorf("CSS() of keyframes without properties = %q, want empty", got)
	}
	if got := Keyframes("none").CSS(); got != "" {
		t.Errorf("CSS() of keyframes without frames = %q, want empty", got)
	}
}

func TestKeyframesOnAndInitially(t *testing.T) {
	k := Keyframes("pulse")
	if k.dynamic() {
		t.Error("dynamic() = true before On")
	}
	k.On("x", whenTrue)
	if !k.dynamic() || len(k.bindings) != 1 || k.bindings[0].action != "x" {
		t.Errorf("bindings = %d, want x bound", len(k.bindings))
	}
	if !k.Initially(true).initially {
		t.Error("Initially(true) left initially false")
	}
	if k.Initially(false).initially {
		t.Error("Initially(false) left initially true")
	}
}

func TestKeyframesErrors(t *testing.T) {
	checkErrors(t, []errorCase{
		{"valid", Keyframes("pulse").From(css.Custom("--o", css.Num(0))).Err(), nil},
		{"empty name", Keyframes("").From(css.Custom("--o", css.Num(0))).Err(), ErrNoName},
		{"binding without an action", Keyframes("pulse").On("", whenTrue).Err(), ErrBadBinding},
		{"binding without a condition", Keyframes("pulse").On("x", nil).Err(), ErrBadBinding},
		{"empty name before a bad binding", Keyframes("").On("", whenTrue).Err(), ErrNoName},
		{"stops at the ends", Keyframes("pulse").At(css.Pct(0), css.Opacity(css.Num(0))).At(css.Pct(100), css.Opacity(css.Num(1))).Err(), nil},
		{"stop below 0%", Keyframes("pulse").At(css.Pct(-1), css.Opacity(css.Num(0))).Err(), ErrBadAtRule},
		{"stop above 100%", Keyframes("pulse").At(css.Pct(100.5), css.Opacity(css.Num(0))).Err(), ErrBadAtRule},
		{"empty stop out of range", Keyframes("pulse").At(css.Pct(150)).Err(), ErrBadAtRule},
		{"empty name before a bad stop", Keyframes("").At(css.Pct(-1), css.Opacity(css.Num(0))).Err(), ErrNoName},
	})
	outside := Keyframes("pulse").At(css.Pct(-10), css.Opacity(css.Num(0))).At(css.Pct(50), css.Opacity(css.Num(1)))
	if got, want := outside.CSS(), "@keyframes pulse {\n\t50% {\n\t\topacity: 1;\n\t}\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if k := Keyframes("pulse").On("x", nil); k.dynamic() {
		t.Error("dynamic() = true after On without a predicate")
	}
}

func TestFontFaceCSS(t *testing.T) {
	want := "@font-face {\n\tfont-family: \"Inter\";\n\tsrc: url(\"/web/inter.woff2\") format(\"woff2\");\n}\n"
	if got := interFace().CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	set := FontFace(css.FaceFontFamily(css.String("Inter")), nil).Set(css.Src(css.Url("/web/inter.woff2")), nil).Set(css.FontDisplay(kw.Swap))
	want = "@font-face {\n\tfont-family: \"Inter\";\n\tsrc: url(\"/web/inter.woff2\");\n\tfont-display: swap;\n}\n"
	if got := set.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	if got := FontFace().CSS(); got != "" {
		t.Errorf("CSS() of an empty font face = %q, want empty", got)
	}
}

func TestFontFaceErrors(t *testing.T) {
	family := css.FaceFontFamily(css.String("Inter"))
	src := css.Src(css.Url("/web/inter.woff2"))
	checkErrors(t, []errorCase{
		{"valid", interFace().Err(), nil},
		{"valid through Set", FontFace().Set(src).Set(family).Err(), nil},
		{"no descriptors", FontFace().Err(), ErrBadAtRule},
		{"no src", FontFace(family).Err(), ErrBadAtRule},
		{"no font-family", FontFace(src).Err(), ErrBadAtRule},
		{"only nil descriptors", FontFace(nil, nil).Err(), ErrBadAtRule},
	})
}

func TestPropertyCSS(t *testing.T) {
	want := "@property --accent {\n\tsyntax: \"<color>\";\n\tinherits: true;\n\tinitial-value: #FF3B30;\n}\n"
	if got := accentProperty().CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	undashed := Property("accent", css.Syntax(css.SyntaxColor)).Set(nil, css.Inherits(kw.False))
	if got := undashed.Name(); got != "--accent" {
		t.Errorf("Name() = %q, want %q", got, "--accent")
	}
	want = "@property --accent {\n\tsyntax: \"<color>\";\n\tinherits: false;\n}\n"
	if got := undashed.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if got := accentProperty().Name(); got != "--accent" {
		t.Errorf("Name() = %q, want %q", got, "--accent")
	}
}

func TestPropertyErrors(t *testing.T) {
	syntax := css.Syntax(css.SyntaxColor)
	inherits := css.Inherits(kw.True)
	checkErrors(t, []errorCase{
		{"valid", accentProperty().Err(), nil},
		{"valid through Set", Property("a").Set(inherits).Set(syntax).Err(), nil},
		{"no syntax", Property("a", inherits).Err(), ErrBadAtRule},
		{"no inherits", Property("a", syntax).Err(), ErrBadAtRule},
		{"no descriptors", Property("a").Err(), ErrBadAtRule},
		{"empty name", Property("", syntax, inherits).Err(), ErrNoName},
		{"only dashes", Property("--", syntax, inherits).Err(), ErrNoName},
		{"empty name without descriptors", Property("").Err(), ErrNoName},
	})
}

func TestImportCSS(t *testing.T) {
	grid := supports.Property(css.Display(kw.Grid))
	cases := []struct {
		name string
		rule *ImportRule
		want string
	}{
		{"url", Import(css.Url("x.css")), "@import url(\"x.css\");\n"},
		{"string", Import(css.String("x.css")), "@import \"x.css\";\n"},
		{"layer", Import(css.Url("x.css"), InLayer("base")), "@import url(\"x.css\") layer(base);\n"},
		{"supports", Import(css.Url("x.css"), IfSupports(grid)), "@import url(\"x.css\") supports(display: grid);\n"},
		{"media", Import(css.Url("x.css"), ForMedia(media.Screen())), "@import url(\"x.css\") screen;\n"},
		{
			"all in order",
			Import(css.Url("x.css"), ForMedia(media.Screen()), IfSupports(grid), InLayer("base")),
			"@import url(\"x.css\") layer(base) supports(display: grid) screen;\n",
		},
		{"nil option", Import(css.Url("x.css"), nil), "@import url(\"x.css\");\n"},
		{"nil target", Import(nil, InLayer("base")), ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.rule.CSS(); got != c.want {
				t.Errorf("CSS() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestImportErrors(t *testing.T) {
	checkErrors(t, []errorCase{
		{"valid", Import(css.Url("x.css"), InLayer("base"), IfSupports(supports.Selector("a > b")), ForMedia(media.Print())).Err(), nil},
		{"nil target", Import(nil).Err(), ErrNoName},
		{"url target", Import(css.Url("x.css")).Err(), nil},
		{"string target", Import(css.String("x.css")).Err(), nil},
		{"length target", Import(css.Px(1)).Err(), ErrBadAtRule},
		{"bad target before options", Import(css.Px(1), InLayer("")).Err(), ErrBadAtRule},
		{"empty layer name", Import(css.Url("x.css"), InLayer("")).Err(), ErrNoName},
		{"bad supports query", Import(css.Url("x.css"), IfSupports(supports.Selector(""))).Err(), ErrBadSupportsQuery},
		{"empty supports query", Import(css.Url("x.css"), IfSupports(supports.Query{})).Err(), ErrBadSupportsQuery},
		{"bad media query", Import(css.Url("x.css"), ForMedia(media.Screen().And(media.Print()))).Err(), ErrBadMediaQuery},
		{"first error wins", Import(css.Url("x.css"), ForMedia(media.Screen().And(media.Print())), IfSupports(supports.Query{})).Err(), ErrBadMediaQuery},
	})
}

func TestLayerCSS(t *testing.T) {
	layer := Layer("base", Class("a").Set(css.Custom("--w", css.Px(1))))
	want := "@layer base {\n\t.a {\n\t\t--w: 1px;\n\t}\n}\n"
	if got := layer.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if got := len(layer.blockTexts()); got != 1 {
		t.Errorf("len(blockTexts()) = %d, want 1", got)
	}
	if got := Layer("empty").CSS(); got != "" {
		t.Errorf("CSS() of an empty layer = %q, want empty", got)
	}
	if got := Layer("bare", Class("a")).CSS(); got != "" {
		t.Errorf("CSS() of a layer holding no properties = %q, want empty", got)
	}
}

func TestLayerNestsItems(t *testing.T) {
	var none *Rule
	layer := Layer("base",
		Class("a").Set(css.Custom("--w", css.Px(1))).pseudo(":hover", []css.Property{css.Custom("--w", css.Px(2))}),
		none,
		Keyframes("pulse").From(css.Custom("--o", css.Num(0))),
		Layer("inner", Class("b").Media(media.Screen(), css.Custom("--w", css.Px(3)))),
	)
	want := "@layer base {\n" +
		"\t.a {\n\t\t--w: 1px;\n\t}\n\n" +
		"\t.a:hover {\n\t\t--w: 2px;\n\t}\n\n" +
		"\t@keyframes pulse {\n\t\tfrom {\n\t\t\t--o: 0;\n\t\t}\n\t}\n\n" +
		"\t@layer inner {\n\t\t@media screen {\n\t\t\t.b {\n\t\t\t\t--w: 3px;\n\t\t\t}\n\t\t}\n\t}\n" +
		"}\n"
	if got := layer.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if err := layer.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

func TestLayerRejectsImports(t *testing.T) {
	checkErrors(t, []errorCase{
		{"import", Layer("base", Import(css.Url("x.css"))).Err(), ErrBadAtRule},
		{"layer order", Layer("base", LayerOrder("a")).Err(), ErrBadAtRule},
		{"import in a nested layer", Layer("base", Layer("inner", Import(css.Url("x.css")))).Err(), ErrBadAtRule},
		{"font face and property", Layer("base", interFace(), accentProperty()).Err(), nil},
	})
}

func TestLayerErrors(t *testing.T) {
	checkErrors(t, []errorCase{
		{"valid", Layer("base", Class("a")).Err(), nil},
		{"empty name", Layer("", Class("a")).Err(), ErrNoName},
		{"empty name before a bad item", Layer("", Import(css.Url("x.css"))).Err(), ErrNoName},
		{"bad item", Layer("base", Class("")).Err(), ErrNoName},
		{"bad nested item", Layer("base", Layer("inner", FontFace())).Err(), ErrBadAtRule},
	})
}

func TestLayerOrderCSS(t *testing.T) {
	if got, want := LayerOrder("base", "components").CSS(), "@layer base, components;\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if got, want := LayerOrder("base").CSS(), "@layer base;\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if got := LayerOrder().CSS(); got != "" {
		t.Errorf("CSS() without names = %q, want empty", got)
	}
}

func TestLayerOrderErrors(t *testing.T) {
	checkErrors(t, []errorCase{
		{"valid", LayerOrder("base", "components").Err(), nil},
		{"no names", LayerOrder().Err(), ErrNoName},
		{"empty name", LayerOrder("base", "").Err(), ErrNoName},
	})
}

func TestItemSubjects(t *testing.T) {
	cases := []struct {
		item Item
		want string
	}{
		{Class("a").Descendant(Element("b")), ".a b"},
		{Keyframes("pulse"), "@keyframes pulse"},
		{interFace(), "@font-face"},
		{accentProperty(), "@property --accent"},
		{Import(css.Url("x.css"), InLayer("base")), "@import url(\"x.css\")"},
		{Import(nil), "@import"},
		{Layer("base"), "@layer base"},
		{LayerOrder("base", "components"), "@layer base, components"},
	}
	for _, c := range cases {
		if got := c.item.subject(); got != c.want {
			t.Errorf("subject() = %q, want %q", got, c.want)
		}
	}
}

func TestBoundItems(t *testing.T) {
	yes := func(app.Action) bool { return true }
	items := []Item{Class("a").On("x", yes), Keyframes("k").On("x", yes), Class("b"), Keyframes("j"), interFace()}
	var got []bool
	for _, i := range items {
		got = append(got, isBound(i))
	}
	if want := []bool{true, true, false, false, false}; !slices.Equal(got, want) {
		t.Errorf("isBound() = %v, want %v", got, want)
	}
}

func TestUnwrapped(t *testing.T) {
	cases := []struct {
		given string
		want  string
	}{
		{"(display: grid)", "display: grid"},
		{"(a) and (b)", "(a) and (b)"},
		{"not (a)", "not (a)"},
		{"selector(a > b)", "selector(a > b)"},
		{"(not (a))", "not (a)"},
		{"(content: \"(\")", "(content: \"(\")"},
		{"(a", "(a"},
	}
	for _, c := range cases {
		if got := unwrapped(c.given); got != c.want {
			t.Errorf("unwrapped(%q) = %q, want %q", c.given, got, c.want)
		}
	}
}
