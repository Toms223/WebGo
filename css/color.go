package css

import (
	"fmt"
	"strings"
)

type ColorLiteral struct{ text string }

func Hex(s string) ColorLiteral {
	if !strings.HasPrefix(s, "#") {
		s = "#" + s
	}
	return ColorLiteral{text: s}
}

func RGB(r, g, b uint8) ColorLiteral {
	return ColorLiteral{text: fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)}
}

func RGBA(r, g, b uint8, a float64) ColorLiteral {
	return ColorLiteral{text: fmt.Sprintf("rgba(%d, %d, %d, %s)", r, g, b, formatNumber(a))}
}

func HSL(h, s, l float64) ColorLiteral {
	return ColorLiteral{text: fmt.Sprintf("hsl(%s, %s, %s)", formatNumber(h), Pct(s).CSS(), Pct(l).CSS())}
}

func HSLA(h, s, l, a float64) ColorLiteral {
	return ColorLiteral{text: fmt.Sprintf("hsla(%s, %s, %s, %s)", formatNumber(h), Pct(s).CSS(), Pct(l).CSS(), formatNumber(a))}
}

func (c ColorLiteral) CSS() string { return c.text }
