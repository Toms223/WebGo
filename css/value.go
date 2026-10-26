// Package css provides typed CSS values, one function per CSS property and one per at-rule descriptor.
// A property function accepts only the value kinds its grammar allows, plus Var and Raw, so a wrong
// value fails to compile.
package css

import "strings"

// Value is any CSS value. CSS returns its text as written into a stylesheet.
type Value interface{ CSS() string }

// Property is one declaration, made by a property function or Custom; only this package makes them.
// Important returns a copy written with !important.
type Property interface {
	Name() string
	Value() string
	Prefixes() []string
	Important() Property
	property()
}

type declaration struct {
	name      string
	value     string
	prefixes  []string
	important bool
}

func (d declaration) Name() string { return d.name }
func (d declaration) Value() string {
	if d.important {
		return d.value + " !important"
	}
	return d.value
}
func (d declaration) Prefixes() []string { return d.prefixes }
func (d declaration) property()          {}

func (d declaration) Important() Property {
	d.important = true
	return d
}

func declare(name, value string, prefixes ...string) Property {
	return declaration{name: name, value: value, prefixes: prefixes}
}

// Custom declares the custom property name, adding the leading -- when missing. Read it back with Var.
func Custom(name string, v Value) Property {
	return declare(dashed(name), valueText(v))
}

func dashed(name string) string {
	if strings.HasPrefix(name, "--") {
		return name
	}
	return "--" + name
}

func valueText(v Value) string {
	if v == nil {
		return ""
	}
	return v.CSS()
}

func join[V Value](first V, rest []V, sep string) string {
	var b strings.Builder
	b.WriteString(valueText(first))
	for _, v := range rest {
		b.WriteString(sep)
		b.WriteString(valueText(v))
	}
	return b.String()
}
