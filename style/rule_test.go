package style

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/container"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/supports"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func TestRuleCSS(t *testing.T) {
	rule := Class("card").Set(css.Custom("--pad", css.Rem(1)), css.Custom("--radius", css.Px(8)))
	want := ".card {\n\t--pad: 1rem;\n\t--radius: 8px;\n}\n"
	if got := rule.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestPseudoAndMediaBlocks(t *testing.T) {
	rule := Class("nav-link").
		Set(css.Custom("--a", css.Px(1))).
		pseudo(":hover", []css.Property{css.Custom("--b", css.Px(2))}).
		Media(media.Screen(), css.Custom("--c", css.Px(3)))
	want := ".nav-link {\n\t--a: 1px;\n}\n\n" +
		".nav-link:hover {\n\t--b: 2px;\n}\n\n" +
		"@media screen {\n\t.nav-link {\n\t\t--c: 3px;\n\t}\n}\n"
	if got := rule.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestSameBlockMergesInCallOrder(t *testing.T) {
	rule := Class("a").
		Set(css.Custom("--one", css.Px(1))).
		pseudo(":hover", []css.Property{css.Custom("--two", css.Px(2))}).
		Set(css.Custom("--three", css.Px(3))).
		pseudo(":hover", []css.Property{css.Custom("--four", css.Px(4))})
	want := []string{
		".a {\n\t--one: 1px;\n\t--three: 3px;\n}",
		".a:hover {\n\t--two: 2px;\n\t--four: 4px;\n}",
	}
	if got := rule.blockTexts(); !slices.Equal(got, want) {
		t.Errorf("blockTexts() = %q, want %q", got, want)
	}

	queries := Class("b").
		Media(media.Screen(), css.Custom("--one", css.Px(1))).
		Set(css.Custom("--two", css.Px(2))).
		Media(media.All(), css.Custom("--three", css.Px(3))).
		Media(media.Screen(), css.Custom("--four", css.Px(4)))
	want = []string{
		".b {\n\t--two: 2px;\n}",
		"@media screen {\n\t.b {\n\t\t--one: 1px;\n\t\t--four: 4px;\n\t}\n}",
		"@media all {\n\t.b {\n\t\t--three: 3px;\n\t}\n}",
	}
	if got := queries.blockTexts(); !slices.Equal(got, want) {
		t.Errorf("blockTexts() = %q, want %q", got, want)
	}
}

func TestBlocksAreOrderedByKind(t *testing.T) {
	nav := Class("nav").Media(media.Screen(), css.Custom("--gap", css.Px(4))).Set(css.Custom("--gap", css.Px(16)))
	want := ".nav {\n\t--gap: 16px;\n}\n\n@media screen {\n\t.nav {\n\t\t--gap: 4px;\n\t}\n}\n"
	if got := nav.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}

	mixed := Class("a").
		Media(media.Print(), css.Custom("--one", css.Px(1))).
		pseudo(":focus", []css.Property{css.Custom("--two", css.Px(2))}).
		Media(media.Screen(), css.Custom("--three", css.Px(3))).
		pseudo(":hover", []css.Property{css.Custom("--four", css.Px(4))}).
		Set(css.Custom("--five", css.Px(5)))
	blocks := []string{
		".a {\n\t--five: 5px;\n}",
		".a:focus {\n\t--two: 2px;\n}",
		".a:hover {\n\t--four: 4px;\n}",
		"@media print {\n\t.a {\n\t\t--one: 1px;\n\t}\n}",
		"@media screen {\n\t.a {\n\t\t--three: 3px;\n\t}\n}",
	}
	if got := mixed.blockTexts(); !slices.Equal(got, blocks) {
		t.Errorf("blockTexts() = %q, want %q", got, blocks)
	}
}

func TestEmptyBlocksAreSkipped(t *testing.T) {
	if got := Class("x").pseudo(":hover", nil).CSS(); got != "" {
		t.Errorf("CSS() = %q, want empty", got)
	}
	if got := Class("x").Set().blockTexts(); len(got) != 0 {
		t.Errorf("blockTexts() = %q, want none", got)
	}
	if got := Class("x").Set(nil, nil).CSS(); got != "" {
		t.Errorf("CSS() of nil properties = %q, want empty", got)
	}
	got := Class("x").Set().pseudo(":hover", []css.Property{css.Custom("--a", css.Px(1))}).Set(nil).CSS()
	if want := ".x:hover {\n\t--a: 1px;\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestPseudoAppliesToSubject(t *testing.T) {
	props := []css.Property{css.Custom("--a", css.Px(1))}
	if got := Class("doc").Descendant(Element("a")).pseudo(":hover", props).CSS(); !strings.HasPrefix(got, ".doc a:hover {") {
		t.Errorf("CSS() = %q, want it to start with %q", got, ".doc a:hover {")
	}
	if got := Class("doc").Child(Any()).pseudo(":first-child", props).CSS(); !strings.HasPrefix(got, ".doc > *:first-child {") {
		t.Errorf("CSS() = %q, want it to start with %q", got, ".doc > *:first-child {")
	}
}

func TestBadMediaQueryLatches(t *testing.T) {
	rule := Class("x").Media(media.Screen().And(media.Print()), css.Custom("--a", css.Px(1)))
	if err := rule.Err(); !errors.Is(err, ErrBadMediaQuery) {
		t.Fatalf("Err() = %v, want %v", err, ErrBadMediaQuery)
	}
	if got := rule.CSS(); got != "" {
		t.Errorf("CSS() = %q, want the block of the bad query left out", got)
	}
	rule.Media(media.Screen(), css.Custom("--b", css.Px(2)))
	if err := rule.Err(); !errors.Is(err, ErrBadMediaQuery) {
		t.Errorf("Err() after a valid query = %v, want %v", err, ErrBadMediaQuery)
	}
	if err := Class("").Media(media.Screen().And(media.Print())).Err(); !errors.Is(err, ErrNoName) {
		t.Errorf("Err() = %v, want the earlier %v", err, ErrNoName)
	}
}

func TestOnAndInitially(t *testing.T) {
	yes := func(app.Action) bool { return true }
	no := func(app.Action) bool { return false }

	rule := Class("a")
	if rule.dynamic() {
		t.Error("dynamic() = true before On")
	}
	rule.On("x", yes)
	if !rule.dynamic() {
		t.Error("dynamic() = false after On")
	}
	rule.On("y", no)
	if len(rule.bindings) != 2 {
		t.Fatalf("len(bindings) = %d, want 2", len(rule.bindings))
	}
	if b := rule.bindings[0]; b.action != "x" || !b.when(app.Action{}) {
		t.Errorf("bindings[0] = %q, want x bound to its condition", b.action)
	}
	if b := rule.bindings[1]; b.action != "y" || b.when(app.Action{}) {
		t.Errorf("bindings[1] = %q, want y bound to its condition", b.action)
	}

	if rule.initially {
		t.Error("initially = true before Initially")
	}
	if !rule.Initially(true).initially {
		t.Error("Initially(true) left initially false")
	}
	if rule.Initially(false).initially {
		t.Error("Initially(false) left initially true")
	}

	if Class("a").On("x", yes).Descendant(Class("b")).dynamic() {
		t.Error("a combinator carried the receiver's bindings")
	}
	derived := Class("a").Set(css.Custom("--a", css.Px(1))).Initially(true).Child(Class("b"))
	if derived.initially {
		t.Error("a combinator carried the receiver's initially")
	}
	if got := derived.CSS(); got != "" {
		t.Errorf("CSS() = %q, want a combinator to carry no blocks", got)
	}
}

func TestBadBindingLatches(t *testing.T) {
	yes := func(app.Action) bool { return true }

	unnamed := Class("a").On("", func(app.Action) bool { return true })
	if err := unnamed.Err(); !errors.Is(err, ErrBadBinding) {
		t.Errorf("Err() after On without an action = %v, want %v", err, ErrBadBinding)
	}
	if unnamed.dynamic() {
		t.Error("dynamic() = true after On without an action")
	}

	unchecked := Class("a").On("x", nil)
	if err := unchecked.Err(); !errors.Is(err, ErrBadBinding) {
		t.Errorf("Err() after On without a predicate = %v, want %v", err, ErrBadBinding)
	}
	if unchecked.dynamic() {
		t.Error("dynamic() = true after On without a predicate")
	}

	unchecked.On("x", yes)
	if err := unchecked.Err(); !errors.Is(err, ErrBadBinding) {
		t.Errorf("Err() after a valid On = %v, want %v", err, ErrBadBinding)
	}
	if len(unchecked.bindings) != 1 || unchecked.bindings[0].action != "x" {
		t.Errorf("len(bindings) = %d, want only the valid binding", len(unchecked.bindings))
	}

	if err := Class("").On("", yes).Err(); !errors.Is(err, ErrNoName) {
		t.Errorf("Err() = %v, want the earlier %v", err, ErrNoName)
	}
}

func TestNestRendersChildrenWithAmpersand(t *testing.T) {
	card := Class("card").
		Set(css.Custom("--p", css.Rem(1))).
		Nest(
			Class("title").Set(css.Custom("--w", css.Int(700))),
			Element("a").Hover(css.Custom("--u", css.Px(1))),
		).
		Nest(Self().Hover().Media(media.MaxWidth(css.Px(600)), css.Custom("--p", css.Rem(0.5))))
	want := ".card {\n\t--p: 1rem;\n\t& .title {\n\t\t--w: 700;\n\t}\n\t& a:hover {\n\t\t--u: 1px;\n\t}\n\t&:hover {\n\t\t@media (max-width: 600px) {\n\t\t\t--p: 0.5rem;\n\t\t}\n\t}\n}\n"
	if err := card.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if got := card.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	if got := card.blockTexts(); len(got) != 1 {
		t.Errorf("blockTexts() = %q, want one text", got)
	}
}

func TestNestRendersOwnBlocksBeforeChildren(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	q := css.Custom("--b", css.Px(2))
	card := Class("card").
		Nest(Class("title").Set(p).Hover(q).Media(media.Screen(), p).Nest(Element("b").Set(q))).
		Media(media.Print(), q).
		Hover(p).
		Set(p)
	want := ".card {\n" +
		"\t--a: 1px;\n" +
		"\t&:hover {\n\t\t--a: 1px;\n\t}\n" +
		"\t@media print {\n\t\t--b: 2px;\n\t}\n" +
		"\t& .title {\n" +
		"\t\t--a: 1px;\n" +
		"\t\t&:hover {\n\t\t\t--b: 2px;\n\t\t}\n" +
		"\t\t@media screen {\n\t\t\t--a: 1px;\n\t\t}\n" +
		"\t\t& b {\n\t\t\t--b: 2px;\n\t\t}\n" +
		"\t}\n" +
		"}\n"
	if got := card.CSS(); got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestNestedCompactFormNeedsALonePseudo(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	cases := []struct {
		name  string
		child *Rule
		want  string
	}{
		{"lone pseudo", Element("a").Hover(p), "\t& a:hover {\n\t\t--a: 1px;\n\t}\n"},
		{"empty base and lone pseudo", Element("a").Set().Hover(p).Focus(), "\t& a:hover {\n\t\t--a: 1px;\n\t}\n"},
		{"base and pseudo", Element("a").Set(p).Hover(p), "\t& a {\n\t\t--a: 1px;\n\t\t&:hover {\n\t\t\t--a: 1px;\n\t\t}\n\t}\n"},
		{"two pseudos", Element("a").Hover(p).Focus(p), "\t& a {\n\t\t&:hover {\n\t\t\t--a: 1px;\n\t\t}\n\t\t&:focus {\n\t\t\t--a: 1px;\n\t\t}\n\t}\n"},
		{"pseudo and media", Element("a").Hover(p).Media(media.Screen(), p), "\t& a {\n\t\t&:hover {\n\t\t\t--a: 1px;\n\t\t}\n\t\t@media screen {\n\t\t\t--a: 1px;\n\t\t}\n\t}\n"},
		{"pseudo and child", Element("a").Hover(p).Nest(Element("b").Set(p)), "\t& a {\n\t\t&:hover {\n\t\t\t--a: 1px;\n\t\t}\n\t\t& b {\n\t\t\t--a: 1px;\n\t\t}\n\t}\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got, want := Class("x").Nest(c.child).CSS(), ".x {\n"+c.want+"}\n"; got != want {
				t.Errorf("CSS() = %q, want %q", got, want)
			}
		})
	}
}

func TestNestedCombinatorStripsTheParent(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	card := Class("card")
	cases := []struct {
		child *Rule
		want  string
	}{
		{card.Child(Element("a")).Set(p), "& > a"},
		{card.Descendant(Element("a")).Set(p), "& a"},
		{card.Next(Element("a")).Set(p), "& + a"},
		{card.Sibling(Element("a")).Set(p), "& ~ a"},
		{Class("cards").Descendant(Element("a")).Set(p), "& .cards a"},
		{card.And(Class("on")).Set(p), "&.on"},
		{card.And(Class("on")).And(Class("big")).Set(p), "&.on.big"},
		{card.And(Class("on")).Descendant(Element("a")).Set(p), "&.on a"},
		{Class("cardx").Set(p), "& .cardx"},
	}
	for _, c := range cases {
		parent := Class("card").Nest(c.child)
		if got := nestedSelector(parent, c.child); got != c.want {
			t.Errorf("nestedSelector(%q, %q) = %q, want %q", parent.Selector(), c.child.Selector(), got, c.want)
		}
	}
	got := Class("card").Nest(card.Child(Element("a")).Set(p)).CSS()
	if want := ".card {\n\t& > a {\n\t\t--a: 1px;\n\t}\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	got = Class("card").Nest(card.And(Class("on")).Set(p)).CSS()
	if want := ".card {\n\t&.on {\n\t\t--a: 1px;\n\t}\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestNestedUnrelatedCombinator(t *testing.T) {
	got := Class("card").Nest(Class("x").Child(Element("a")).Set(css.Custom("--a", css.Px(1)))).CSS()
	if want := ".card {\n\t& .x > a {\n\t\t--a: 1px;\n\t}\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestNestUnderSelfStripsTheRealParent(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	card := Class("card")
	got := Class("card").Nest(Self().Nest(card.Child(Element("a")).Set(p))).CSS()
	if want := ".card {\n\t& > a {\n\t\t--a: 1px;\n\t}\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestNestClaimsTheChild(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	child := Class("title").Set(p)
	first := Class("card").Nest(child)
	if err := first.Err(); err != nil {
		t.Fatalf("first parent Err() = %v, want nil", err)
	}
	if child.owner != first {
		t.Error("Nest did not claim the child")
	}
	second := Class("panel").Nest(child)
	if err := second.Err(); !errors.Is(err, ErrNested) {
		t.Errorf("second parent Err() = %v, want %v", err, ErrNested)
	}
	if len(second.children) != 0 {
		t.Errorf("second parent kept %d children, want none", len(second.children))
	}
	if err := child.Descendant(Class("x")).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("Descendant of a nested rule Err() = %v, want %v", err, ErrNested)
	}
	if err := Class("x").Child(child).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("Child with a nested operand Err() = %v, want %v", err, ErrNested)
	}
	if err := child.And(Class("x")).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("And of a nested rule Err() = %v, want %v", err, ErrNested)
	}
	if err := Element("a").And(child).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("And with a nested operand Err() = %v, want %v", err, ErrNested)
	}

	twice := Class("title")
	if err := Class("card").Nest(twice, twice).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("Err() after nesting one child twice = %v, want %v", err, ErrNested)
	}
	loop := Class("a")
	if err := loop.Nest(loop).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("Err() after nesting a rule in itself = %v, want %v", err, ErrNested)
	}
	outer := Class("outer")
	inner := Class("inner")
	outer.Nest(inner)
	if err := inner.Nest(outer).Err(); !errors.Is(err, ErrNested) {
		t.Errorf("Err() after nesting an ancestor = %v, want %v", err, ErrNested)
	}
	if outer.owner != nil {
		t.Error("nesting an ancestor claimed it")
	}
	if err := Class("card").Nest(nil).Err(); !errors.Is(err, ErrNoName) {
		t.Errorf("Err() after nesting nil = %v, want %v", err, ErrNoName)
	}
}

func TestNestedErrorsReachTheParent(t *testing.T) {
	deep := Class("a").Nest(Class("b").Nest(Class("c").Media(media.Screen().And(media.Print()))))
	if err := deep.Err(); !errors.Is(err, ErrBadMediaQuery) {
		t.Errorf("Err() = %v, want %v", err, ErrBadMediaQuery)
	}
	own := Class("").Nest(Class("b").Media(media.Screen().And(media.Print())))
	if err := own.Err(); !errors.Is(err, ErrNoName) {
		t.Errorf("Err() = %v, want the parent's own %v", err, ErrNoName)
	}
	first := Class("a").Nest(Class(""), Class("c").Media(media.Screen().And(media.Print())))
	if err := first.Err(); !errors.Is(err, ErrNoName) {
		t.Errorf("Err() = %v, want the first child's %v", err, ErrNoName)
	}
}

func TestNestedBindingIsAnError(t *testing.T) {
	yes := func(app.Action) bool { return true }
	if err := Class("a").Nest(Class("b").On("x", yes)).Err(); !errors.Is(err, ErrNestedBinding) {
		t.Errorf("Err() = %v, want %v", err, ErrNestedBinding)
	}
	if err := Class("a").Nest(Class("").On("x", yes)).Err(); !errors.Is(err, ErrNestedBinding) {
		t.Errorf("Err() = %v, want %v before the child's own error", err, ErrNestedBinding)
	}
	if err := Class("a").Nest(Class("b").Nest(Class("c").On("x", yes))).Err(); !errors.Is(err, ErrNestedBinding) {
		t.Errorf("Err() of a grandchild binding = %v, want %v", err, ErrNestedBinding)
	}
	if err := Class("a").On("x", yes).Nest(Class("b")).Err(); err != nil {
		t.Errorf("Err() of a bound parent = %v, want nil", err)
	}
}

func TestSelf(t *testing.T) {
	s := Self()
	if got := s.Name(); got != "" {
		t.Errorf("Name() = %q, want empty", got)
	}
	if got := s.Selector(); got != "&" {
		t.Errorf("Selector() = %q, want %q", got, "&")
	}
	if err := Self().Hover(css.Custom("--a", css.Px(1))).Media(media.Screen(), css.Custom("--a", css.Px(2))).Err(); err != nil {
		t.Errorf("Err() of pseudo and media blocks = %v, want nil", err)
	}
	if err := Self().Set(css.Custom("--a", css.Px(1))).Err(); !errors.Is(err, ErrBadSelector) {
		t.Errorf("Err() of Self().Set = %v, want %v", err, ErrBadSelector)
	}
	if err := Class("a").Nest(Self().Set(css.Custom("--a", css.Px(1)))).Err(); !errors.Is(err, ErrBadSelector) {
		t.Errorf("parent Err() of Self().Set = %v, want %v", err, ErrBadSelector)
	}
	if err := Self().Descendant(Class("a")).Err(); !errors.Is(err, ErrBadSelector) {
		t.Errorf("Err() of Self in a combinator = %v, want %v", err, ErrBadSelector)
	}
	if err := Class("a").And(Self()).Err(); !errors.Is(err, ErrBadSelector) {
		t.Errorf("Err() of Self in And = %v, want %v", err, ErrBadSelector)
	}
	if err := Class("a").Nest(Self().And(Class("on"))).Err(); err == nil {
		t.Error("Err() of Self().And = nil, want an error")
	}
}

func TestSelfRendersInsideTheParent(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	q := css.Custom("--b", css.Px(2))
	got := Class("x").Nest(Self().Focus(p).Media(media.Screen(), q)).CSS()
	want := ".x {\n\t&:focus {\n\t\t--a: 1px;\n\t\t@media screen {\n\t\t\t--b: 2px;\n\t\t}\n\t}\n}\n"
	if got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	got = Class("x").Nest(Self().Media(media.Screen(), q).Hover(p)).CSS()
	want = ".x {\n\t&:hover {\n\t\t--a: 1px;\n\t}\n\t@media screen {\n\t\t--b: 2px;\n\t}\n}\n"
	if got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	got = Class("x").Nest(Self().Hover(p).Media(media.Screen(), q).Media(media.Print(), p)).CSS()
	want = ".x {\n\t&:hover {\n\t\t--a: 1px;\n\t\t@media screen {\n\t\t\t--b: 2px;\n\t\t}\n\t\t@media print {\n\t\t\t--a: 1px;\n\t\t}\n\t}\n}\n"
	if got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestNestWithoutChildrenRendersFlat(t *testing.T) {
	rule := Class("card").
		Set(css.Custom("--a", css.Px(1))).
		Hover(css.Custom("--b", css.Px(2))).
		Media(media.Screen(), css.Custom("--c", css.Px(3))).
		Nest()
	want := []string{
		".card {\n\t--a: 1px;\n}",
		".card:hover {\n\t--b: 2px;\n}",
		"@media screen {\n\t.card {\n\t\t--c: 3px;\n\t}\n}",
	}
	if got := rule.blockTexts(); !slices.Equal(got, want) {
		t.Errorf("blockTexts() = %q, want %q", got, want)
	}
}

func TestNestedPseudoMediaWithoutProps(t *testing.T) {
	if got := Class("x").Nest(Self().Hover()).CSS(); got != "" {
		t.Errorf("CSS() of an empty Self().Hover() = %q, want empty", got)
	}
	if got := Class("x").Nest(Self().Hover().Media(media.Screen())).CSS(); got != "" {
		t.Errorf("CSS() of an empty media under Self().Hover() = %q, want empty", got)
	}
	got := Class("x").Set(css.Custom("--a", css.Px(1))).Nest(Self().Hover(), Class("y")).CSS()
	if want := ".x {\n\t--a: 1px;\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
	s := Self().Hover().Media(media.Screen(), css.Custom("--a", css.Px(1)))
	if len(s.blocks) != 2 || s.blocks[1].under != ":hover" {
		t.Errorf("Media after Hover on Self did not hang under :hover")
	}
	plain := Class("y").Hover().Media(media.Screen(), css.Custom("--a", css.Px(1)))
	if plain.blocks[1].under != "" {
		t.Errorf("Media after Hover on a plain rule hung under %q, want top level", plain.blocks[1].under)
	}
}

func TestContainerAndSupportsBlocks(t *testing.T) {
	grid := css.Display(kw.Grid)
	rule := Class("nav").
		Media(media.Screen(), css.Custom("--m", css.Px(1))).
		Supports(supports.Property(grid), css.Gap(css.Rem(1))).
		Container(container.Named("side").And(container.MinWidth(css.Px(400))), grid).
		Hover(css.Custom("--h", css.Px(2))).
		Set(css.Custom("--b", css.Px(3)))
	want := []string{
		".nav {\n\t--b: 3px;\n}",
		".nav:hover {\n\t--h: 2px;\n}",
		"@container side (min-width: 400px) {\n\t.nav {\n\t\tdisplay: grid;\n\t}\n}",
		"@supports (display: grid) {\n\t.nav {\n\t\tgap: 1rem;\n\t}\n}",
		"@media screen {\n\t.nav {\n\t\t--m: 1px;\n\t}\n}",
	}
	if err := rule.Err(); err != nil {
		t.Fatalf("Err() = %v, want nil", err)
	}
	if got := rule.blockTexts(); !slices.Equal(got, want) {
		t.Errorf("blockTexts() = %q, want %q", got, want)
	}
	if got, want := rule.CSS(), strings.Join(want, "\n\n")+"\n"; got != want {
		t.Errorf("CSS() = %q, want %q", got, want)
	}
}

func TestConditionalBlocksMergeByKindAndCondition(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	q := css.Custom("--b", css.Px(2))
	rule := Class("x").
		Container(container.MinWidth(css.Px(20)), p).
		Media(media.MinWidth(css.Px(20)), p).
		Supports(supports.Selector(":has(a)"), p).
		Container(container.MinWidth(css.Px(20)), q).
		Container(container.MaxWidth(css.Px(20)), q).
		Supports(supports.Selector(":has(a)"), q)
	want := []string{
		"@container (min-width: 20px) {\n\t.x {\n\t\t--a: 1px;\n\t\t--b: 2px;\n\t}\n}",
		"@container (max-width: 20px) {\n\t.x {\n\t\t--b: 2px;\n\t}\n}",
		"@supports selector(:has(a)) {\n\t.x {\n\t\t--a: 1px;\n\t\t--b: 2px;\n\t}\n}",
		"@media (min-width: 20px) {\n\t.x {\n\t\t--a: 1px;\n\t}\n}",
	}
	if got := rule.blockTexts(); !slices.Equal(got, want) {
		t.Errorf("blockTexts() = %q, want %q", got, want)
	}
	if got := Class("x").Container(container.MinWidth(css.Px(20))).Supports(supports.Selector("a")).CSS(); got != "" {
		t.Errorf("CSS() of empty conditional blocks = %q, want empty", got)
	}
}

func TestBadConditionalQueries(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	cases := []struct {
		name string
		rule *Rule
		want error
	}{
		{"container", Class("x").Container(container.Named("").And(container.Named("b")), p), ErrBadContainerQuery},
		{"zero container", Class("x").Container(container.Query{}, p), ErrBadContainerQuery},
		{"supports", Class("x").Supports(supports.Selector(""), p), ErrBadSupportsQuery},
		{"zero supports", Class("x").Supports(supports.Query{}, p), ErrBadSupportsQuery},
		{"first wins", Class("x").Supports(supports.Selector(""), p).Container(container.Query{}, p), ErrBadSupportsQuery},
		{"media first", Class("x").Media(media.Screen().And(media.Print()), p).Supports(supports.Selector(""), p), ErrBadMediaQuery},
		{"valid after bad", Class("x").Container(container.Query{}, p).Container(container.MinWidth(css.Px(1)), p), ErrBadContainerQuery},
		{"nested", Class("a").Nest(Self().Supports(supports.Selector(""), p)), ErrBadSupportsQuery},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.rule.Err(); !errors.Is(err, c.want) {
				t.Errorf("Err() = %v, want %v", err, c.want)
			}
		})
	}
	if got := Class("x").Container(container.Query{}, p).Supports(supports.Selector(""), p).CSS(); got != "" {
		t.Errorf("CSS() = %q, want the blocks of bad queries left out", got)
	}
}

func TestNestedContainerBlock(t *testing.T) {
	p := css.Custom("--a", css.Px(1))
	q := css.Custom("--b", css.Px(2))
	wide := container.MinWidth(css.Px(20))
	grid := supports.Property(css.Display(kw.Grid))
	cases := []struct {
		name string
		rule *Rule
		want string
	}{
		{
			"self container",
			Class("card").Nest(Self().Container(wide, p)),
			".card {\n\t@container (min-width: 20px) {\n\t\t--a: 1px;\n\t}\n}\n",
		},
		{
			"self supports",
			Class("card").Nest(Self().Supports(grid, p)),
			".card {\n\t@supports (display: grid) {\n\t\t--a: 1px;\n\t}\n}\n",
		},
		{
			"self hover container",
			Class("card").Nest(Self().Hover().Container(wide, p)),
			".card {\n\t&:hover {\n\t\t@container (min-width: 20px) {\n\t\t\t--a: 1px;\n\t\t}\n\t}\n}\n",
		},
		{
			"self hover in rank order",
			Class("card").Nest(Self().Hover(q).Media(media.Screen(), p).Supports(grid, p).Container(wide, q)),
			".card {\n\t&:hover {\n\t\t--b: 2px;\n" +
				"\t\t@container (min-width: 20px) {\n\t\t\t--b: 2px;\n\t\t}\n" +
				"\t\t@supports (display: grid) {\n\t\t\t--a: 1px;\n\t\t}\n" +
				"\t\t@media screen {\n\t\t\t--a: 1px;\n\t\t}\n" +
				"\t}\n}\n",
		},
		{
			"parent blocks in rank order",
			Class("card").Media(media.Print(), q).Supports(grid, p).Container(wide, q).Hover(p).Set(p).Nest(Class("title").Set(q)),
			".card {\n\t--a: 1px;\n" +
				"\t&:hover {\n\t\t--a: 1px;\n\t}\n" +
				"\t@container (min-width: 20px) {\n\t\t--b: 2px;\n\t}\n" +
				"\t@supports (display: grid) {\n\t\t--a: 1px;\n\t}\n" +
				"\t@media print {\n\t\t--b: 2px;\n\t}\n" +
				"\t& .title {\n\t\t--b: 2px;\n\t}\n}\n",
		},
		{
			"named child",
			Class("card").Nest(Class("title").Container(wide, p).Supports(grid, q)),
			".card {\n\t& .title {\n" +
				"\t\t@container (min-width: 20px) {\n\t\t\t--a: 1px;\n\t\t}\n" +
				"\t\t@supports (display: grid) {\n\t\t\t--b: 2px;\n\t\t}\n" +
				"\t}\n}\n",
		},
		{
			"named child pseudo and container",
			Class("card").Nest(Element("a").Hover(p).Container(wide, q)),
			".card {\n\t& a {\n" +
				"\t\t&:hover {\n\t\t\t--a: 1px;\n\t\t}\n" +
				"\t\t@container (min-width: 20px) {\n\t\t\t--b: 2px;\n\t\t}\n" +
				"\t}\n}\n",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if err := c.rule.Err(); err != nil {
				t.Fatalf("Err() = %v, want nil", err)
			}
			if got := c.rule.CSS(); got != c.want {
				t.Errorf("CSS() = %q, want %q", got, c.want)
			}
		})
	}
	s := Self().Hover().Container(wide, p).Supports(grid, p)
	if len(s.blocks) != 3 || s.blocks[1].under != ":hover" || s.blocks[2].under != ":hover" {
		t.Errorf("Container and Supports after Hover on Self did not hang under :hover")
	}
	plain := Class("y").Hover().Container(wide, p).Supports(grid, p)
	if plain.blocks[1].under != "" || plain.blocks[2].under != "" {
		t.Errorf("Container and Supports after Hover on a plain rule hung under a pseudo")
	}
}
