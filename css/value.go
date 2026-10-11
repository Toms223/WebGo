package css

import "strings"

type Value interface{ CSS() string }

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
