// Package supports builds @supports conditions for style.Rule.Supports and style.IfSupports.
package supports

import (
	"cmp"
	"errors"

	"github.com/Toms223/WebGo/css"
)

// ErrBadQuery is carried by a condition built from a nil or empty argument. Operations on a query keep
// its first error.
var ErrBadQuery = errors.New("supports: query cannot be expressed")

// Query is an @supports condition. And, Or and Not parenthesize joined operands, and the zero Query
// carries ErrBadQuery.
type Query struct {
	text     string
	compound bool
	err      error
}

func Property(p css.Property) Query {
	if p == nil {
		return Query{err: ErrBadQuery}
	}
	return Query{text: "(" + p.Name() + ": " + p.Value() + ")"}
}

func Selector(text string) Query {
	if text == "" {
		return Query{err: ErrBadQuery}
	}
	return Query{text: "selector(" + text + ")"}
}

func FontFormat(f css.FormatValue) Query {
	if f == nil {
		return Query{err: ErrBadQuery}
	}
	return Query{text: "font-format(" + f.CSS() + ")"}
}

func FontTech(t css.Value) Query {
	if t == nil {
		return Query{err: ErrBadQuery}
	}
	return Query{text: "font-tech(" + t.CSS() + ")"}
}

func (q Query) And(o Query) Query { return q.join(" and ", o) }

func (q Query) Or(o Query) Query { return q.join(" or ", o) }

func Not(q Query) Query {
	return Query{text: "not " + q.operand(), compound: true, err: q.Err()}
}

func (q Query) CSS() string { return q.text }

func (q Query) Err() error {
	if q.err == nil && q.text == "" {
		return ErrBadQuery
	}
	return q.err
}

func (q Query) join(op string, o Query) Query {
	return Query{text: q.operand() + op + o.operand(), compound: true, err: cmp.Or(q.Err(), o.Err())}
}

func (q Query) operand() string {
	if q.compound {
		return "(" + q.text + ")"
	}
	return q.text
}
