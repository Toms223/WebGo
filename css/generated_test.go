package css_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
)

func TestGeneratedPropertyRenders(t *testing.T) {
	width := css.Width(css.Px(10))
	if width.Name() != "width" || width.Value() != "10px" {
		t.Errorf("Width(Px(10)) = %s: %s, want width: 10px", width.Name(), width.Value())
	}
	if width.Prefixes() != nil {
		t.Errorf("Width(Px(10)).Prefixes() = %v, want nil", width.Prefixes())
	}

	cases := []struct {
		name     string
		property css.Property
		want     string
	}{
		{"keyword", css.Width(kw.Auto), "auto"},
		{"repeated values joined by a space", css.Padding(css.Rem(1), css.Rem(2)), "1rem 2rem"},
		{"keyword with a dash", css.Display(kw.InlineFlex), "inline-flex"},
		{"var", css.Width(css.Var("--w")), "var(--w)"},
		{"calc", css.Width(css.Calc(css.Sub(css.Pct(100), css.Px(2)))), "calc(100% - 2px)"},
		{"named color", css.Color(kw.Rebeccapurple), "rebeccapurple"},
		{"raw", css.Width(css.Raw("fit-content(20em)")), "fit-content(20em)"},
		{"comma list", css.FontFamily(css.String("Inter"), kw.SansSerif), `"Inter", sans-serif`},
		{"comma list behind a keyword alternative", css.TextShadow(css.Raw("0 1px #000"), css.Raw("0 2px #000")), "0 1px #000, 0 2px #000"},
	}
	for _, c := range cases {
		if got := c.property.Value(); got != c.want {
			t.Errorf("%s: Value() = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestGeneratedPropertyCarriesItsPrefixes(t *testing.T) {
	p := css.BackdropFilter(css.Raw("blur(4px)"))
	if p.Name() != "backdrop-filter" || p.Value() != "blur(4px)" {
		t.Errorf("BackdropFilter = %s: %s, want backdrop-filter: blur(4px)", p.Name(), p.Value())
	}
	if want := []string{"-webkit-"}; !slices.Equal(p.Prefixes(), want) {
		t.Errorf("Prefixes() = %v, want %v", p.Prefixes(), want)
	}
}

func TestMarkerAbsent(t *testing.T) {
	flex := reflect.TypeOf(kw.Flex)
	if _, ok := flex.MethodByName("CSSWidth"); ok {
		t.Error("kw.Flex has CSSWidth, so css.Width(kw.Flex) would compile")
	}
	if _, ok := flex.MethodByName("CSSDisplay"); !ok {
		t.Error("kw.Flex lacks CSSDisplay, so css.Display(kw.Flex) would not compile")
	}
}

func TestGeneratedDescriptorRenders(t *testing.T) {
	cases := []struct {
		name       string
		descriptor css.Descriptor
		want       string
	}{
		{"keyword", css.FontDisplay(kw.Swap), "font-display: swap"},
		{"source with a format", css.Src(css.Url("/web/inter.woff2"), css.Format(kw.Woff2)), `src: url("/web/inter.woff2") format("woff2")`},
		{"syntax", css.Syntax(css.SyntaxColor), `syntax: "<color>"`},
		{"inherits", css.Inherits(kw.False), "inherits: false"},
		{"initial value", css.InitialValue(css.Hex("#FF3B30")), "initial-value: #FF3B30"},
	}
	for _, c := range cases {
		if got := c.descriptor.Name() + ": " + c.descriptor.Value(); got != c.want {
			t.Errorf("%s: %q, want %q", c.name, got, c.want)
		}
	}
}

func TestDescriptorMarkerAbsent(t *testing.T) {
	if _, ok := reflect.TypeOf(kw.Flex).MethodByName("DescriptorFontDisplay"); ok {
		t.Error("kw.Flex has DescriptorFontDisplay, so css.FontDisplay(kw.Flex) would compile")
	}
	if _, ok := reflect.TypeOf(kw.Swap).MethodByName("DescriptorFontDisplay"); !ok {
		t.Error("kw.Swap lacks DescriptorFontDisplay, so css.FontDisplay(kw.Swap) would not compile")
	}
}

func TestVarIsNotADescriptorValue(t *testing.T) {
	varType := reflect.TypeOf(css.VarValue{})
	rawType := reflect.TypeOf(css.RawValue{})
	for _, marker := range []string{"DescriptorSyntax", "DescriptorFontDisplay"} {
		if _, ok := varType.MethodByName(marker); ok {
			t.Errorf("css.VarValue has %s, so a descriptor would accept css.Var", marker)
		}
		if _, ok := rawType.MethodByName(marker); !ok {
			t.Errorf("css.RawValue lacks %s, so a descriptor would not accept css.Raw", marker)
		}
	}
	for i := range varType.NumMethod() {
		if name := varType.Method(i).Name; strings.HasPrefix(name, "Descriptor") {
			t.Errorf("css.VarValue has %s", name)
		}
	}
}
