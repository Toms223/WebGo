package main

import (
	"bytes"
	"fmt"
	"slices"
	"strings"

	"github.com/Toms223/WebGo/internal/cssgen/grammar"
)

var calcKinds = []string{"Length", "Percentage", "Number", "Integer", "Angle", "Time", "Fraction"}

var alwaysAccepted = []string{"RawValue", "VarValue"}

var descriptorAccepted = []string{"RawValue"}

func emitCSS(m *model) []byte {
	b := source("css")
	for _, p := range m.properties {
		param := "Value"
		if p.tier == 1 {
			param = p.goName + "Value"
			markerInterface(b, param, "Value", cssMarker(p.goName))
		}
		args, value := signature(param, p.shape)
		fmt.Fprintf(b, "\nfunc %s(%s) Property {\n\treturn declare(%q, %s%s)\n}\n", p.goName, args, p.css, value, quoted(p.prefixes))
	}
	return b.Bytes()
}

func signature(param string, shape grammar.Shape) (args, value string) {
	if shape.Min != 1 || shape.Max != 1 {
		return fmt.Sprintf("first %s, rest ...%s", param, param), fmt.Sprintf("join(first, rest, %q)", separator(shape.Comma))
	}
	return "v " + param, "v.CSS()"
}

func emitCSSMarkers(m *model) []byte {
	b := source("css")
	for _, p := range m.properties {
		if p.tier == 1 {
			acceptedBy(b, p.shape.Kinds, alwaysAccepted, cssMarker(p.goName))
		}
	}
	b.WriteString("\n")
	for _, f := range m.features {
		if f.rangeKind == "" {
			markerMethod(b, "RawValue", mediaMarker(f.goName))
		}
	}
	for _, d := range m.descriptors {
		if d.tier == 1 {
			acceptedBy(b, d.shape.Kinds, descriptorAccepted, descriptorMarker(d.goName))
		}
	}
	b.WriteString("\n")
	for _, f := range m.containerFeatures {
		if f.rangeKind == "" {
			markerMethod(b, "RawValue", containerMarker(f.goName))
		}
	}
	return b.Bytes()
}

func acceptedBy(b *bytes.Buffer, kinds, always []string, marker string) {
	receivers := slices.Concat(kinds, always)
	if slices.ContainsFunc(kinds, func(kind string) bool { return slices.Contains(calcKinds, kind) }) {
		receivers = append(receivers, "CalcValue")
	}
	b.WriteString("\n")
	for _, receiver := range receivers {
		markerMethod(b, receiver, marker)
	}
}

func separator(comma bool) string {
	if comma {
		return ", "
	}
	return " "
}

func quoted(values []string) string {
	var b strings.Builder
	for _, v := range values {
		fmt.Fprintf(&b, ", %q", v)
	}
	return b.String()
}
