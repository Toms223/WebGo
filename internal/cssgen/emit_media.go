package main

import (
	"bytes"
	"fmt"
)

func emitMedia(m *model) []byte {
	b := source("media", featureImports(m.features)...)
	fmt.Fprintf(b, "\nconst rangeSyntax = %t\n", m.rangeSyntax)
	writeFeatures(b, m.features, mediaMarker, func(f mediaFeature, kind string) {
		fmt.Fprintf(b, "func %sBetween(lo, hi %s) Query {\n", f.goName, kind)
		fmt.Fprintf(b, "\tif rangeSyntax {\n\t\treturn features(\"(\" + lo.CSS() + %q + hi.CSS() + \")\")\n\t}\n", " <= "+f.css+" <= ")
		fmt.Fprintf(b, "\treturn features(%q+lo.CSS()+\")\", %q+hi.CSS()+\")\")\n}\n", "(min-"+f.css+": ", "(max-"+f.css+": ")
	})
	return b.Bytes()
}

func featureImports(features []mediaFeature) []string {
	if len(features) == 0 {
		return nil
	}
	return []string{cssImport}
}

func writeFeatures(b *bytes.Buffer, features []mediaFeature, marker func(string) string, between func(f mediaFeature, kind string)) {
	for _, f := range features {
		if f.rangeKind == "" {
			value := f.goName + "Value"
			markerInterface(b, value, "css.Value", marker(f.goName))
			fmt.Fprintf(b, "\nfunc %s(v %s) Query { return features(%q + v.CSS() + \")\") }\n", f.goName, value, "("+f.css+": ")
			continue
		}
		kind := "css." + f.rangeKind
		fmt.Fprintf(b, "\nfunc Min%s(v %s) Query { return features(%q + v.CSS() + \")\") }\n", f.goName, kind, "(min-"+f.css+": ")
		fmt.Fprintf(b, "func Max%s(v %s) Query { return features(%q + v.CSS() + \")\") }\n", f.goName, kind, "(max-"+f.css+": ")
		fmt.Fprintf(b, "func %s(v %s) Query { return features(%q + v.CSS() + \")\") }\n", f.goName, kind, "("+f.css+": ")
		between(f, kind)
	}
}
