package main

import "fmt"

func emitContainer(m *model) []byte {
	b := source("container", featureImports(m.containerFeatures)...)
	writeFeatures(b, m.containerFeatures, containerMarker, func(f mediaFeature, kind string) {
		fmt.Fprintf(b, "func %sBetween(lo, hi %s) Query {\n\treturn features(\"(\" + lo.CSS() + %q + hi.CSS() + \")\")\n}\n", f.goName, kind, " <= "+f.css+" <= ")
	})
	return b.Bytes()
}
