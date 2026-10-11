package main

import (
	"cmp"
	"fmt"
	"strings"
)

var stringParams = map[string]string{":lang()": "tag"}

func emitPseudo(m *model) []byte {
	var imports []string
	if len(m.pseudos) > 0 {
		imports = append(imports, cssImport)
	}
	b := source("style", imports...)
	b.WriteString("\n")
	for _, p := range m.pseudos {
		var param, text string
		switch p.arg {
		case "":
			fmt.Fprintf(b, "func (r *Rule) %s(props ...css.Property) *Rule { return r.pseudo(%q, props) }\n", p.goName, p.css)
			continue
		case "nth":
			param, text = "n css.Nth", "n.CSS()"
		case "string":
			name := cmp.Or(stringParams[p.css], "name")
			param, text = name+" string", name
		case "dir":
			value := p.goName + "Value"
			markerInterface(b, value, "css.Value", styleMarker(p.goName))
			b.WriteString("\n")
			param, text = "d "+value, "d.CSS()"
		}
		fmt.Fprintf(b, "func (r *Rule) %s(%s, props ...css.Property) *Rule {\n\treturn r.pseudo(%q+%s+\")\", props)\n}\n", p.goName, param, strings.TrimSuffix(p.css, ")"), text)
	}
	return b.Bytes()
}
