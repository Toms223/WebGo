package app

import (
	"fmt"
)

func Text(v any) UI {
	return &text{value: toString(v)}
}

func Textf(format string, v ...any) UI {
	return &text{value: fmt.Sprintf(format, v...)}
}

type text struct {
	jsvalue       Value
	parentElement UI
	value         string
}

func (t *text) JSValue() Value {
	return t.jsvalue
}

func (t *text) Mounted() bool {
	return t.jsvalue != nil
}

func (t *text) parent() UI {
	return t.parentElement
}

func (t *text) setParent(p UI) UI {
	t.parentElement = p
	return t
}
