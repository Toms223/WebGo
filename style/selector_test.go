package style

import (
	"errors"
	"testing"
)

func TestSelectorsRender(t *testing.T) {
	cases := []struct {
		rule *Rule
		want string
	}{
		{Class("card"), ".card"},
		{Id("nav"), "#nav"},
		{Element("a"), "a"},
		{Any(), "*"},
		{Class("doc").Descendant(Element("a")), ".doc a"},
		{Class("doc").Child(Any()), ".doc > *"},
		{Class("card").Next(Class("card")), ".card + .card"},
		{Class("card").Sibling(Class("card")), ".card ~ .card"},
		{Class("card").And(Class("active")), ".card.active"},
		{Element("a").And(Class("nav-link")), "a.nav-link"},
		{Class("nav-link").And(Element("a")), "a.nav-link"},
		{Class("doc").Child(Any()).And(Class("x")), ".doc > .x"},
		{Class("doc").Child(Element("ul")).Descendant(Element("a")), ".doc > ul a"},
		{Class("a").Child(Class("b").Descendant(Class("c"))), ".a > .b .c"},
		{Class("a").And(Id("b")).And(Element("li")), "li#b.a"},
		{Any().And(Id("main")), "#main"},
	}
	for _, c := range cases {
		if got := c.rule.Selector(); got != c.want {
			t.Errorf("Selector() = %q, want %q", got, c.want)
		}
		if err := c.rule.Err(); err != nil {
			t.Errorf("Err() of %q = %v, want nil", c.want, err)
		}
	}
}

func TestNames(t *testing.T) {
	cases := []struct {
		rule *Rule
		want string
	}{
		{Class("card"), "card"},
		{Id("nav"), "nav"},
		{Element("a"), ""},
		{Any(), ""},
		{Class("a").Descendant(Class("b")), ""},
		{Class("a").And(Class("b")), ""},
		{Any().And(Class("x")), ""},
		{Class("x").And(Any()), ""},
		{Element("a").And(Class("x")), ""},
		{Id("a").And(Class("x")), ""},
	}
	for _, c := range cases {
		if got := c.rule.Name(); got != c.want {
			t.Errorf("Name() of %q = %q, want %q", c.rule.Selector(), got, c.want)
		}
	}
}

func TestSelectorErrors(t *testing.T) {
	cases := []struct {
		name string
		rule *Rule
		want error
	}{
		{"an empty class", Class(""), ErrNoName},
		{"an empty id", Id(""), ErrNoName},
		{"an empty element", Element(""), ErrNoName},
		{"two elements in one compound", Element("a").And(Element("b")), ErrBadSelector},
		{"an element and the universal selector in one compound", Any().And(Element("a")), ErrBadSelector},
		{"two ids in one compound", Id("a").And(Id("b")), ErrBadSelector},
		{"and with an operand of several compounds", Class("a").And(Class("b").Child(Class("c"))), ErrBadSelector},
		{"a combinator carries the receiver error", Class("").Descendant(Class("b")), ErrNoName},
		{"a combinator carries the argument error", Class("a").Child(Id("")), ErrNoName},
		{"a combinator keeps the receiver error over the argument error", Element("a").And(Element("b")).Next(Class("")), ErrBadSelector},
		{"and carries the argument error", Class("a").And(Class("")), ErrNoName},
		{"and keeps an operand error over its own failure", Class("").And(Class("b").Child(Class("c"))), ErrNoName},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rule.Err(); !errors.Is(got, tc.want) {
				t.Errorf("Err() = %v, want %v", got, tc.want)
			}
		})
	}

	receiver := Class("doc")
	_ = receiver.Child(Element("a"))
	if got := receiver.Selector(); got != ".doc" {
		t.Errorf("Selector() of the receiver after Child = %q, want %q", got, ".doc")
	}
}

func TestCombinatorsLeaveTheirOperandsUntouched(t *testing.T) {
	chain := Class("a").Child(Class("b")).Child(Class("c"))
	longer := chain.Child(Class("d"))
	merged := Class("a").And(Class("b")).And(Class("c"))
	cases := []struct {
		rule *Rule
		want string
	}{
		{chain.Descendant(Class("x")), ".a > .b > .c .x"},
		{chain.Next(Class("y")), ".a > .b > .c + .y"},
		{longer.Descendant(Class("x")), ".a > .b > .c > .d .x"},
		{longer.Next(Class("y")), ".a > .b > .c > .d + .y"},
		{chain.And(Class("f")), ".a > .b > .c.f"},
		{chain.Sibling(merged), ".a > .b > .c ~ .a.b.c"},
		{merged.And(Class("d")), ".a.b.c.d"},
		{merged.And(Class("e")), ".a.b.c.e"},
		{merged.And(Element("p").Child(Class("x"))), ".a.b.c"},
		{chain, ".a > .b > .c"},
		{longer, ".a > .b > .c > .d"},
		{merged, ".a.b.c"},
	}
	for _, c := range cases {
		if got := c.rule.Selector(); got != c.want {
			t.Errorf("Selector() = %q, want %q", got, c.want)
		}
	}
}

func TestZeroRule(t *testing.T) {
	var zero Rule
	cases := []struct {
		rule *Rule
		want string
	}{
		{&zero, ""},
		{zero.And(Class("a")), ""},
		{zero.Child(Class("a")).Descendant(Class("b")), ".a .b"},
		{Class("a").Child(&zero).Descendant(Class("b")), ".a .b"},
		{new(Rule).On("", nil), ""},
	}
	for _, c := range cases {
		if got := c.rule.Selector(); got != c.want {
			t.Errorf("Selector() = %q, want %q", got, c.want)
		}
		if err := c.rule.Err(); !errors.Is(err, ErrNoName) {
			t.Errorf("Err() of %q = %v, want %v", c.want, err, ErrNoName)
		}
	}
}

func TestNilOperandsLatchNoName(t *testing.T) {
	cases := []struct {
		name    string
		combine func(*Rule) *Rule
	}{
		{"Descendant", Class("a").Descendant},
		{"Child", Class("a").Child},
		{"Next", Class("a").Next},
		{"Sibling", Class("a").Sibling},
		{"And", Class("a").And},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rule := tc.combine(nil)
			if err := rule.Err(); !errors.Is(err, ErrNoName) {
				t.Errorf("Err() = %v, want %v", err, ErrNoName)
			}
			if got := rule.Selector(); got != ".a" {
				t.Errorf("Selector() = %q, want %q", got, ".a")
			}
		})
	}
}
