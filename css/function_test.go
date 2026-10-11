package css

import "testing"

func TestVar(t *testing.T) {
	cases := []struct {
		value Value
		want  string
	}{
		{Var("--ink"), "var(--ink)"},
		{Var("ink"), "var(--ink)"},
		{Var("--ink", Hex("#000")), "var(--ink, #000)"},
		{Var("--a", Px(1), Pct(2)), "var(--a, 1px, 2%)"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestCalc(t *testing.T) {
	cases := []struct {
		value Value
		want  string
	}{
		{Calc(Sub(Pct(100), Px(20))), "calc(100% - 20px)"},
		{Calc(Add(Pct(100), Px(-20))), "calc(100% + -20px)"},
		{Calc(Mul(Add(Pct(100), Px(2)), Num(2))), "calc((100% + 2px) * 2)"},
		{Calc(Div(Var("--w"), Num(3))), "calc(var(--w) / 3)"},
		{Min(Px(10), Pct(5)), "min(10px, 5%)"},
		{Max(Px(10), Pct(5), Rem(1)), "max(10px, 5%, 1rem)"},
		{Clamp(Rem(1), Vw(2), Rem(3)), "clamp(1rem, 2vw, 3rem)"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestCalcAcceptsEveryExprKind(t *testing.T) {
	cases := []struct {
		expr Expr
		want string
	}{
		{Px(1), "calc(1px)"},
		{Pct(2), "calc(2%)"},
		{Num(3), "calc(3)"},
		{Int(4), "calc(4)"},
		{Deg(5), "calc(5deg)"},
		{S(6), "calc(6s)"},
		{Fr(7), "calc(7fr)"},
		{Var("--x"), "calc(var(--x))"},
		{Min(Px(1), Px(2)), "calc(min(1px, 2px))"},
		{Calc(Px(1)), "calc(calc(1px))"},
	}
	for _, c := range cases {
		if got := Calc(c.expr).CSS(); got != c.want {
			t.Errorf("Calc(...).CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestCalcParenthesisesNestedOperands(t *testing.T) {
	cases := []struct {
		expr Expr
		want string
	}{
		{Sub(Px(10), Sub(Px(5), Px(3))), "calc(10px - (5px - 3px))"},
		{Sub(Px(10), Add(Px(5), Px(3))), "calc(10px - (5px + 3px))"},
		{Sub(Add(Px(10), Px(5)), Px(3)), "calc(10px + 5px - 3px)"},
		{Add(Px(10), Sub(Px(5), Px(3))), "calc(10px + 5px - 3px)"},
		{Div(Sub(Pct(100), Px(2)), Num(2)), "calc((100% - 2px) / 2)"},
		{Mul(Mul(Px(2), Num(3)), Num(4)), "calc((2px * 3) * 4)"},
		{Mul(Calc(Add(Px(1), Px(2))), Num(2)), "calc(calc(1px + 2px) * 2)"},
	}
	for _, c := range cases {
		if got := Calc(c.expr).CSS(); got != c.want {
			t.Errorf("Calc(...).CSS() = %q, want %q", got, c.want)
		}
	}
}

func TestNilOperandsRenderEmpty(t *testing.T) {
	cases := []struct {
		value Value
		want  string
	}{
		{Var("--x", nil), "var(--x, )"},
		{Calc(nil), "calc()"},
		{Calc(Add(nil, Px(1))), "calc( + 1px)"},
		{Min(nil, nil), "min(, )"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}
