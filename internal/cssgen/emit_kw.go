package main

import "fmt"

func emitKW(m *model) []byte {
	b := source("kw")
	for _, k := range m.keywords {
		typ := keywordType(k.goName)
		fmt.Fprintf(b, "\ntype %s string\n\nfunc (k %s) CSS() string { return string(k) }\n\nconst %s %s = %q\n", typ, typ, k.goName, typ, k.css)
	}
	return b.Bytes()
}

func emitKWMarkers(m *model) []byte {
	properties := map[string]string{}
	for _, p := range m.properties {
		properties[p.css] = p.goName
	}
	features := map[string]string{}
	for _, f := range m.features {
		features[f.css] = f.goName
	}
	containerFeatures := map[string]string{}
	for _, f := range m.containerFeatures {
		containerFeatures[f.css] = f.goName
	}
	b := source("kw")
	for _, k := range m.keywords {
		var markers []string
		for _, name := range k.cssProps {
			markers = append(markers, cssMarker(properties[name]))
		}
		for _, name := range k.mediaFeatures {
			markers = append(markers, mediaMarker(features[name]))
		}
		for _, use := range k.styleUses {
			markers = append(markers, styleMarker(use))
		}
		for _, use := range k.descriptors {
			markers = append(markers, descriptorMarker(use))
		}
		for _, name := range k.containerFeatures {
			markers = append(markers, containerMarker(containerFeatures[name]))
		}
		for _, use := range k.cssUses {
			markers = append(markers, cssMarker(use))
		}
		if len(markers) == 0 {
			continue
		}
		b.WriteString("\n")
		for _, marker := range markers {
			markerMethod(b, keywordType(k.goName), marker)
		}
	}
	return b.Bytes()
}
