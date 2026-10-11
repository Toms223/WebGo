package css

import "strconv"

func formatNumber(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

type Number struct{ amount float64 }

func Num(n float64) Number { return Number{amount: n} }

func (n Number) CSS() string { return formatNumber(n.amount) }

type Integer struct{ amount int }

func Int(n int) Integer { return Integer{amount: n} }

func (i Integer) CSS() string { return strconv.Itoa(i.amount) }

type Percentage struct{ amount float64 }

func Pct(n float64) Percentage { return Percentage{amount: n} }

func (p Percentage) CSS() string { return formatNumber(p.amount) + "%" }

type RatioValue struct{ w, h int }

func Ratio(w, h int) RatioValue { return RatioValue{w: w, h: h} }

func (r RatioValue) CSS() string { return strconv.Itoa(r.w) + "/" + strconv.Itoa(r.h) }
