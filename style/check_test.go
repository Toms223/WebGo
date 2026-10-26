package style

import (
	"errors"
	"slices"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
)

func whenTrue(app.Action) bool { return true }

func plainRule(name string) *Rule {
	return Class(name).Set(css.Custom("--w", css.Px(1)))
}

func selectorsOf(items []Item) []string {
	var out []string
	for _, i := range items {
		out = append(out, i.subject())
	}
	return out
}

func TestResolveDescribesEachKind(t *testing.T) {
	cases := []struct {
		name  string
		style Style
		want  Source
	}{
		{"static", Static("/web/app.css"), Source{Kind: StaticStyle, Path: "/web/app.css"}},
		{"static url", Static("https://cdn/x.css"), Source{Kind: StaticStyle, Path: "https://cdn/x.css"}},
		{"static as given", Static("/web/./app.css"), Source{Kind: StaticStyle, Path: "/web/./app.css"}},
		{"generated", Generated("/web/t.css", plainRule("a")), Source{Kind: GeneratedStyle, Path: "/web/t.css", Content: []byte(".a {\n\t--w: 1px;\n}\n")}},
		{"empty generated", Generated("/web/t.css"), Source{Kind: GeneratedStyle, Path: "/web/t.css", Content: []byte{}}},
		{"dynamic", Dynamic(plainRule("a").On("x", whenTrue)), Source{Kind: DynamicStyle}},
		{"nil", nil, Source{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Resolve(c.style)
			if err != nil {
				t.Fatalf("Resolve() error = %v, want nil", err)
			}
			if got.Kind != c.want.Kind || got.Path != c.want.Path || string(got.Content) != string(c.want.Content) {
				t.Errorf("Resolve() = %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestResolveReportsItemErrors(t *testing.T) {
	cases := []struct {
		name  string
		style Style
		want  error
	}{
		{"unnamed in generated", Generated("/web/t.css", Class("")), ErrNoName},
		{"bound in generated", Generated("/web/t.css", plainRule("a").On("x", whenTrue)), ErrDynamicInGenerated},
		{"same selector twice", Generated("/web/t.css", plainRule("a"), plainRule("a")), ErrDuplicateRule},
		{"same selector in blocks of different kinds", Generated("/web/t.css", plainRule("a"), Class("a").pseudo(":hover", []css.Property{css.Custom("--x", css.Px(1))})), ErrDuplicateRule},
		{"unbound in dynamic", Dynamic(plainRule("a")), ErrNotDynamic},
		{"unnamed in dynamic", Dynamic(Class("").On("x", whenTrue)), ErrNoName},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Resolve(c.style)
			if !errors.Is(err, c.want) {
				t.Fatalf("Resolve() error = %v, want %v", err, c.want)
			}
			if got.Kind != 0 || got.Path != "" || got.Content != nil {
				t.Errorf("Resolve() = %+v alongside an error, want the zero Source", got)
			}
		})
	}
}

func TestNilItemsAreIgnored(t *testing.T) {
	var (
		rule      *Rule
		keyframes *KeyframesRule
		face      *FontFaceRule
		property  *PropertyRule
		imported  *ImportRule
		layer     *LayerRule
		order     *LayerOrderRule
	)
	g := Generated("/web/t.css", nil, rule, keyframes, face, property, imported, layer, order, plainRule("a"))
	if got := g.(generated).items; len(got) != 1 {
		t.Errorf("items = %d, want only the rule", len(got))
	}
	d := Dynamic(nil, rule, keyframes, plainRule("b").On("x", whenTrue))
	if got := selectorsOf(d.(dynamic).items); !slices.Equal(got, []string{".b"}) {
		t.Errorf("items = %q, want %q", got, []string{".b"})
	}
	if got := Layer("base", nil, rule, plainRule("a")).items; len(got) != 1 {
		t.Errorf("layer items = %d, want only the rule", len(got))
	}
}

func TestStylesKeepTheirOwnRules(t *testing.T) {
	rules := []Item{plainRule("a"), plainRule("b")}
	generatedStyle := Generated("/web/t.css", rules...)
	dynamicStyle := Dynamic(rules...)
	rules[0], rules[1] = rules[1], rules[0]

	source, err := Resolve(generatedStyle)
	if err != nil {
		t.Fatalf("Resolve() error = %v, want nil", err)
	}
	if got, want := string(source.Content), ".a {\n\t--w: 1px;\n}\n\n.b {\n\t--w: 1px;\n}\n"; got != want {
		t.Errorf("Content = %q, want %q", got, want)
	}
	if got, want := selectorsOf(dynamicItems([]Style{dynamicStyle})), []string{".a", ".b"}; !slices.Equal(got, want) {
		t.Errorf("dynamicItems() = %q, want %q", got, want)
	}
}

func TestDynamicRules(t *testing.T) {
	a := plainRule("a").On("x", whenTrue)
	b := plainRule("b").On("y", whenTrue)
	c := plainRule("c").On("z", whenTrue)
	styles := []Style{
		Static("/web/a.css"),
		Dynamic(a, b),
		Generated("/web/g.css", plainRule("g")),
		Dynamic(c),
		Dynamic(),
	}
	if got, want := dynamicItems(styles), []Item{a, b, c}; !slices.Equal(got, want) {
		t.Errorf("dynamicItems() = %q, want %q", selectorsOf(got), selectorsOf(want))
	}
	if got := dynamicItems(nil); len(got) != 0 {
		t.Errorf("dynamicItems(nil) = %q, want none", selectorsOf(got))
	}
	if got := dynamicItems([]Style{Static("/web/a.css"), Generated("/web/g.css", plainRule("g"))}); len(got) != 0 {
		t.Errorf("dynamicItems() of static and generated styles = %q, want none", selectorsOf(got))
	}
}
