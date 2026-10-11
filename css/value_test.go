package css

import (
	"slices"
	"testing"
)

func TestCustomProperty(t *testing.T) {
	p := Custom("--brand", Hex("#fff"))
	if p.Name() != "--brand" {
		t.Errorf("Name() = %q, want %q", p.Name(), "--brand")
	}
	if p.Value() != "#fff" {
		t.Errorf("Value() = %q, want %q", p.Value(), "#fff")
	}
	if p.Prefixes() != nil {
		t.Errorf("Prefixes() = %v, want nil", p.Prefixes())
	}
	if got := Custom("brand", Px(1)).Name(); got != "--brand" {
		t.Errorf("Name() = %q, want %q", got, "--brand")
	}
}

func TestDeclareCarriesPrefixes(t *testing.T) {
	p := declare("backdrop-filter", "blur(4px)", "-webkit-")
	if p.Name() != "backdrop-filter" {
		t.Errorf("Name() = %q, want %q", p.Name(), "backdrop-filter")
	}
	if p.Value() != "blur(4px)" {
		t.Errorf("Value() = %q, want %q", p.Value(), "blur(4px)")
	}
	if want := []string{"-webkit-"}; !slices.Equal(p.Prefixes(), want) {
		t.Errorf("Prefixes() = %v, want %v", p.Prefixes(), want)
	}
}

func TestJoin(t *testing.T) {
	if got := join(Px(1), []Length{Px(2), Px(3)}, " "); got != "1px 2px 3px" {
		t.Errorf("join lengths = %q, want %q", got, "1px 2px 3px")
	}
	if got := join(Ident("a"), []IdentValue{Ident("b")}, ", "); got != "a, b" {
		t.Errorf("join idents = %q, want %q", got, "a, b")
	}
	if got := join(Px(1), nil, " "); got != "1px" {
		t.Errorf("join single = %q, want %q", got, "1px")
	}
}

func TestNilValuesRenderEmpty(t *testing.T) {
	if got := Custom("--x", nil).Value(); got != "" {
		t.Errorf("Custom nil Value() = %q, want empty", got)
	}
	if got := join[Value](nil, nil, ", "); got != "" {
		t.Errorf("join nil = %q, want empty", got)
	}
	if got := join[Value](Px(1), []Value{nil}, ", "); got != "1px, " {
		t.Errorf("join nil rest = %q, want %q", got, "1px, ")
	}
}

func TestImportant(t *testing.T) {
	p := Custom("--a", Px(1)).Important()
	if p.Value() != "1px !important" || p.Name() != "--a" {
		t.Errorf("got %s: %s", p.Name(), p.Value())
	}
	if twice := p.Important().Value(); twice != "1px !important" {
		t.Errorf("Important twice = %q, want %q", twice, "1px !important")
	}
	if plain := Custom("--a", Px(1)).Value(); plain != "1px" {
		t.Errorf("Important mutated the original: %q", plain)
	}
}

func TestImportantKeepsPrefixes(t *testing.T) {
	p := declare("backdrop-filter", "blur(4px)", "-webkit-").Important()
	if want := []string{"-webkit-"}; !slices.Equal(p.Prefixes(), want) {
		t.Errorf("Prefixes() = %v, want %v", p.Prefixes(), want)
	}
	if p.Value() != "blur(4px) !important" {
		t.Errorf("Value() = %q, want %q", p.Value(), "blur(4px) !important")
	}
}
