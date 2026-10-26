package app

type Condition interface {
	UI

	ElseIf(expr bool, elem func() UI) Condition

	ElseIfSlice(expr bool, elems func() []UI) Condition

	Else(elem func() UI) Condition

	ElseSlice(elems func() []UI) Condition

	body() []UI
}

// If renders elem() when expr is true, and nothing otherwise. Chain ElseIf and
// Else for alternatives; only the chosen branch's function runs.
func If(expr bool, elem func() UI) Condition {
	return IfSlice(expr, func() []UI {
		return []UI{elem()}
	})
}

func IfSlice(expr bool, elems func() []UI) Condition {
	if !expr {
		return condition{}
	}
	return condition{
		children: FilterUIElems(elems()...),
		matched:  true,
	}
}

type condition struct {
	children []UI
	matched  bool
}

func (c condition) ElseIf(expr bool, elem func() UI) Condition {
	return c.ElseIfSlice(expr, func() []UI {
		return []UI{elem()}
	})
}

func (c condition) ElseIfSlice(expr bool, elems func() []UI) Condition {
	if c.matched || !expr {
		return c
	}

	c.children = FilterUIElems(elems()...)
	c.matched = true
	return c
}

func (c condition) Else(elem func() UI) Condition {
	return c.ElseSlice(func() []UI {
		return []UI{elem()}
	})
}

func (c condition) ElseSlice(elems func() []UI) Condition {
	return c.ElseIfSlice(true, elems)
}

func (c condition) JSValue() Value {
	return nil
}

func (c condition) Mounted() bool {
	return false
}

func (c condition) body() []UI {
	return c.children
}

func (c condition) parent() UI {
	return nil
}

func (c condition) setParent(UI) UI {
	return nil
}
