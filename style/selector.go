package style

import (
	"cmp"
	"slices"
	"strings"
)

type compound struct {
	element   string
	universal bool
	id        string
	classes   []string
}

type selector struct {
	parts       []compound
	combinators []string
}

// Class returns an empty rule for the selector .name. Pass it to an element's Style method to apply
// the class; an empty name gives ErrNoName.
func Class(name string) *Rule { return named(compound{classes: []string{name}}, name) }

// Id returns an empty rule for the selector #name. An empty name gives ErrNoName.
func Id(name string) *Rule { return named(compound{id: name}, name) }

// Element returns an empty rule for the type selector tag. An empty tag gives ErrNoName.
func Element(tag string) *Rule { return named(compound{element: tag}, tag) }

// Any returns an empty rule for the universal selector *. And drops the * when it joins Any with an id
// or a class.
func Any() *Rule { return &Rule{sel: selector{parts: []compound{{universal: true}}}} }

func named(c compound, name string) *Rule {
	r := &Rule{sel: selector{parts: []compound{c}}}
	if name == "" {
		r.err = ErrNoName
	}
	return r
}

// ClassNames returns the classes of a rule whose selector is classes only, such as .card or
// .card.active, whatever pseudo or conditional blocks it has. It returns nil for a nil rule, Self, and
// any selector with an id, an element, Any or a combinator.
func (r *Rule) ClassNames() []string {
	if r == nil || len(r.sel.parts) != 1 {
		return nil
	}
	c := r.sel.parts[0]
	if c.typed() || c.id != "" || len(c.classes) == 0 {
		return nil
	}
	return slices.Clone(c.classes)
}

// Descendant returns a new rule selecting o anywhere inside r, as in .doc a. Neither operand changes,
// and the new rule starts with no properties or bindings.
func (r *Rule) Descendant(o *Rule) *Rule { return r.combine(" ", o) }

// Child returns a new rule selecting o directly inside r, as in .doc > *.
func (r *Rule) Child(o *Rule) *Rule { return r.combine(" > ", o) }

// Next returns a new rule selecting o immediately after r, as in .card + .card.
func (r *Rule) Next(o *Rule) *Rule { return r.combine(" + ", o) }

// Sibling returns a new rule selecting o anywhere after r, as in .card ~ .card.
func (r *Rule) Sibling(o *Rule) *Rule { return r.combine(" ~ ", o) }

// And returns a new rule matching both selectors on one element, as in .card.active or a.nav-link,
// joining o onto the last compound of r. Joining two elements, two ids, or a chained o gives
// ErrBadSelector.
func (r *Rule) And(o *Rule) *Rule {
	o = cmp.Or(o, &Rule{})
	sel, ok := r.sel.and(o.sel)
	err := operandErr(r, o)
	if !ok {
		err = cmp.Or(err, ErrBadSelector)
	}
	return &Rule{sel: sel, err: err}
}

func (r *Rule) Name() string {
	if len(r.sel.parts) != 1 {
		return ""
	}
	c := r.sel.parts[0]
	switch {
	case c.typed():
		return ""
	case c.id != "" && len(c.classes) == 0:
		return c.id
	case c.id == "" && len(c.classes) == 1:
		return c.classes[0]
	}
	return ""
}

func (r *Rule) Selector() string {
	if r.self {
		return "&"
	}
	return r.sel.text()
}

func (r *Rule) combine(combinator string, o *Rule) *Rule {
	o = cmp.Or(o, &Rule{})
	return &Rule{sel: r.sel.join(combinator, o.sel), err: operandErr(r, o)}
}

func operandErr(r, o *Rule) error {
	return cmp.Or(r.Err(), o.Err(), r.placementErr(), o.placementErr())
}

func (s selector) join(combinator string, o selector) selector {
	between := []string{combinator}
	if len(s.parts) == 0 || len(o.parts) == 0 {
		between = nil
	}
	return selector{
		parts:       slices.Concat(s.parts, o.parts),
		combinators: slices.Concat(s.combinators, between, o.combinators),
	}
}

func (s selector) and(o selector) (selector, bool) {
	merged := selector{parts: slices.Clone(s.parts), combinators: slices.Clone(s.combinators)}
	last := len(merged.parts) - 1
	if last < 0 || len(o.parts) != 1 {
		return merged, false
	}
	c, ok := merged.parts[last].and(o.parts[0])
	if ok {
		merged.parts[last] = c
	}
	return merged, ok
}

func (s selector) text() string {
	var b strings.Builder
	for i, c := range s.parts {
		if i > 0 {
			b.WriteString(s.combinators[i-1])
		}
		b.WriteString(c.text())
	}
	return b.String()
}

func (c compound) and(o compound) (compound, bool) {
	if (c.typed() && o.typed()) || (c.id != "" && o.id != "") {
		return compound{}, false
	}
	return compound{
		element:   cmp.Or(c.element, o.element),
		universal: c.universal || o.universal,
		id:        cmp.Or(c.id, o.id),
		classes:   slices.Concat(c.classes, o.classes),
	}, true
}

func (c compound) typed() bool { return c.element != "" || c.universal }

func (c compound) text() string {
	var b strings.Builder
	b.WriteString(c.element)
	if c.universal && c.id == "" && len(c.classes) == 0 {
		b.WriteString("*")
	}
	if c.id != "" {
		b.WriteString("#" + c.id)
	}
	for _, class := range c.classes {
		b.WriteString("." + class)
	}
	return b.String()
}
