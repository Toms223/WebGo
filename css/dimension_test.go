package css

import "testing"

func TestDimensionsFormatShortest(t *testing.T) {
	cases := []struct {
		value Value
		want  string
	}{
		{Px(8), "8px"},
		{Rem(0.9), "0.9rem"},
		{Rem(1), "1rem"},
		{Px(0), "0px"},
		{Px(-20), "-20px"},
		{Pct(50), "50%"},
		{Num(1.5), "1.5"},
		{Int(3), "3"},
		{Deg(90), "90deg"},
		{Turn(0.25), "0.25turn"},
		{Ms(250), "250ms"},
		{S(1), "1s"},
		{Dppx(2), "2dppx"},
		{Fr(1), "1fr"},
		{Ratio(16, 9), "16/9"},
		{Vmin(10), "10vmin"},
		{Q(4), "4Q"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestEveryUnitConstructor(t *testing.T) {
	cases := []struct {
		value Value
		want  string
	}{
		{Px(1), "1px"},
		{Rem(1), "1rem"},
		{Em(1), "1em"},
		{Ch(1), "1ch"},
		{Ex(1), "1ex"},
		{Vw(1), "1vw"},
		{Vh(1), "1vh"},
		{Vmin(1), "1vmin"},
		{Vmax(1), "1vmax"},
		{Dvh(1), "1dvh"},
		{Svh(1), "1svh"},
		{Lvh(1), "1lvh"},
		{Dvw(1), "1dvw"},
		{Svw(1), "1svw"},
		{Lvw(1), "1lvw"},
		{Cm(1), "1cm"},
		{Mm(1), "1mm"},
		{In(1), "1in"},
		{Pt(1), "1pt"},
		{Pc(1), "1pc"},
		{Q(1), "1Q"},
		{Deg(1), "1deg"},
		{Rad(1), "1rad"},
		{Grad(1), "1grad"},
		{Turn(1), "1turn"},
		{S(1), "1s"},
		{Ms(1), "1ms"},
		{Dpi(1), "1dpi"},
		{Dpcm(1), "1dpcm"},
		{Dppx(1), "1dppx"},
		{Fr(1), "1fr"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	cases := []struct {
		in   float64
		want string
	}{
		{0, "0"},
		{-2.5, "-2.5"},
		{0.1, "0.1"},
		{1e21, "1000000000000000000000"},
		{1e-7, "0.0000001"},
	}
	for _, c := range cases {
		if got := formatNumber(c.in); got != c.want {
			t.Errorf("formatNumber(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}
