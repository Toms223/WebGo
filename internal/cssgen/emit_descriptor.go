package main

import "fmt"

func emitDescriptor(m *model) []byte {
	b := source("css")
	for _, d := range m.descriptors {
		param := "Value"
		if d.tier == 1 {
			param = d.goName + "Descriptor"
			markerInterface(b, param, "Value", descriptorMarker(d.goName))
		}
		args, value := signature(param, d.shape)
		fmt.Fprintf(b, "\nfunc %s(%s) Descriptor {\n\treturn declareDescriptor(%q, %s)\n}\n", d.goName, args, d.css, value)
	}
	return b.Bytes()
}
