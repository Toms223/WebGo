package media

import (
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
)

func TestGeneratedFeatures(t *testing.T) {
	between := "(min-width: 400px) and (max-width: 600px)"
	if rangeSyntax {
		between = "(400px <= width <= 600px)"
	}
	cases := []struct {
		query Query
		want  string
	}{
		{Screen().And(MaxWidth(css.Px(600))), "screen and (max-width: 600px)"},
		{WidthBetween(css.Px(400), css.Px(600)), between},
		{PrefersColorScheme(kw.Dark), "(prefers-color-scheme: dark)"},
		{MinWidth(css.Rem(30)), "(min-width: 30rem)"},
		{Width(css.Px(320)), "(width: 320px)"},
		{AspectRatio(css.Ratio(16, 9)), "(aspect-ratio: 16/9)"},
		{MinResolution(css.Dppx(2)), "(min-resolution: 2dppx)"},
		{Orientation(kw.Landscape), "(orientation: landscape)"},
		{Hover(kw.None), "(hover: none)"},
		{Not(PrefersColorScheme(kw.Light)), "not (prefers-color-scheme: light)"},
		{PrefersColorScheme(css.Raw("dark")), "(prefers-color-scheme: dark)"},
	}
	for _, c := range cases {
		if err := c.query.Err(); err != nil {
			t.Errorf("%s: Err() = %v", c.want, err)
		}
		if got := c.query.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}
