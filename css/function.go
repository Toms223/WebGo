package css

import "strings"

// Expr is an operand of Calc, Min, Max and Clamp: a dimension, number, percentage, Var, Calc, or the
// result of Add, Sub, Mul or Div.
type Expr interface{ expr() string }

func (l Length) expr() string     { return l.CSS() }
func (p Percentage) expr() string { return p.CSS() }
func (n Number) expr() string     { return n.CSS() }
func (i Integer) expr() string    { return i.CSS() }
func (a Angle) expr() string      { return a.CSS() }
func (t Time) expr() string       { return t.CSS() }
func (f Fraction) expr() string   { return f.CSS() }
func (v VarValue) expr() string   { return v.CSS() }
func (c CalcValue) expr() string  { return c.CSS() }
func (b binary) expr() string     { return b.text }

type VarValue struct {
	name     string
	fallback []Value
}

// Var refers to the custom property name, adding the leading -- when missing. Fallback values are
// written after it, separated by commas.
func Var(name string, fallback ...Value) VarValue {
	return VarValue{name: dashed(name), fallback: fallback}
}

func (v VarValue) CSS() string {
	if len(v.fallback) == 0 {
		return "var(" + v.name + ")"
	}
	return "var(" + v.name + ", " + join(v.fallback[0], v.fallback[1:], ", ") + ")"
}

type CalcValue struct{ text string }

func (c CalcValue) CSS() string { return c.text }

// Calc wraps e in calc(). Build e with Add, Sub, Mul and Div.
func Calc(e Expr) CalcValue {
	return mathFunction("calc", e)
}

func Min(first Expr, rest ...Expr) CalcValue {
	return mathFunction("min", first, rest...)
}

func Max(first Expr, rest ...Expr) CalcValue {
	return mathFunction("max", first, rest...)
}

func Clamp(lo, preferred, hi Expr) CalcValue {
	return mathFunction("clamp", lo, preferred, hi)
}

func mathFunction(name string, first Expr, rest ...Expr) CalcValue {
	var b strings.Builder
	b.WriteString(name)
	b.WriteByte('(')
	b.WriteString(exprText(first))
	for _, e := range rest {
		b.WriteString(", ")
		b.WriteString(exprText(e))
	}
	b.WriteByte(')')
	return CalcValue{text: b.String()}
}

type binary struct{ text string }

// Add returns a + b.
func Add(a, b Expr) Expr {
	return binary{text: exprText(a) + " + " + exprText(b)}
}

// Sub returns a - b, parenthesizing b when it is itself an operation.
func Sub(a, b Expr) Expr {
	return binary{text: exprText(a) + " - " + operand(b)}
}

// Mul returns a * n, parenthesizing a when it is itself an operation.
func Mul(a Expr, n Number) Expr {
	return binary{text: operand(a) + " * " + n.CSS()}
}

// Div returns a / n, parenthesizing a when it is itself an operation.
func Div(a Expr, n Number) Expr {
	return binary{text: operand(a) + " / " + n.CSS()}
}

func operand(e Expr) string {
	if b, ok := e.(binary); ok {
		return "(" + b.text + ")"
	}
	return exprText(e)
}

func exprText(e Expr) string {
	if e == nil {
		return ""
	}
	return e.expr()
}
