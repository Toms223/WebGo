package main

import (
	"fmt"
	"slices"
	"strings"
)

var reservedRule = []string{"Name", "Selector", "Set", "Media", "On", "Initially", "CSS", "Err", "Descendant", "Child", "Next", "Sibling", "And", "Nest", "Bind", "Container", "Supports", "Self", "CSSFor"}

var reservedCSS = []string{
	"Add", "An", "Angle", "Calc", "CalcValue", "Ch", "Clamp", "Cm", "ColorLiteral", "Custom",
	"Deg", "Div", "Dpcm", "Dpi", "Dppx", "Dvh", "Dvw", "Em", "Even", "Ex",
	"Expr", "Fr", "Fraction", "Grad", "HSL", "HSLA", "Hex", "Ident", "IdentValue", "In",
	"Int", "Integer", "Length", "Lvh", "Lvw", "Max", "Min", "Mm", "Ms", "Mul",
	"Nth", "Num", "Number", "Odd", "Pc", "Pct", "Percentage", "Property", "Pt", "Px",
	"Q", "RGB", "RGBA", "Rad", "Ratio", "RatioValue", "Raw", "RawValue", "Rem", "Resolution",
	"S", "String", "StringValue", "Sub", "Svh", "Svw", "Time", "Turn", "Url", "UrlValue",
	"Value", "Var", "VarValue", "Vh", "Vmax", "Vmin", "Vw",
	"Descriptor", "Format", "FormatValue", "SyntaxValue", "SyntaxRaw", "SyntaxLength", "SyntaxNumber", "SyntaxPercentage",
	"SyntaxLengthPercentage", "SyntaxColor", "SyntaxImage", "SyntaxUrl", "SyntaxInteger", "SyntaxAngle", "SyntaxTime",
	"SyntaxResolution", "SyntaxTransformFunction", "SyntaxTransformList", "SyntaxCustomIdent", "SyntaxAny",
}

var reservedMedia = []string{"Query", "Screen", "Print", "All", "Not", "ErrBadQuery"}

var reservedContainer = []string{"Query", "Named", "Not", "ErrBadQuery"}

func goName(css string) string {
	name := strings.TrimPrefix(strings.TrimSuffix(strings.TrimLeft(css, ":"), "()"), "-")
	var b strings.Builder
	for piece := range strings.SplitSeq(strings.ToLower(name), "-") {
		if piece != "" {
			b.WriteString(strings.ToUpper(piece[:1]) + piece[1:])
		}
	}
	return b.String()
}

func keywordType(goName string) string {
	return "k" + goName
}

type namespace struct {
	owner    string
	reserved []string
	claimed  map[string]string
}

func newNamespace(owner string, reserved []string) *namespace {
	return &namespace{owner: owner, reserved: reserved, claimed: map[string]string{}}
}

func (n *namespace) claim(item, ident string) error {
	if slices.Contains(n.reserved, ident) {
		return fmt.Errorf("cssgen: %s maps to %s, which %s already declares", item, ident, n.owner)
	}
	if other, taken := n.claimed[ident]; taken {
		return fmt.Errorf("cssgen: %s and %s both map to %s in %s", other, item, ident, n.owner)
	}
	n.claimed[ident] = item
	return nil
}
