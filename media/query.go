package media

import (
	"cmp"
	"errors"
	"slices"
	"strings"
)

var ErrBadQuery = errors.New("media: query cannot be expressed as one media query")

type Query struct {
	clauses []clause
	err     error
}

type clause struct {
	not       bool
	mediaType string
	features  []string
}

func Screen() Query { return Query{clauses: []clause{{mediaType: "screen"}}} }

func Print() Query { return Query{clauses: []clause{{mediaType: "print"}}} }

func All() Query { return Query{clauses: []clause{{mediaType: "all"}}} }

func features(texts ...string) Query {
	return Query{clauses: []clause{{features: texts}}}
}

func (q Query) And(o Query) Query {
	err := cmp.Or(q.err, o.err)
	if len(q.clauses) == 1 && len(o.clauses) == 1 {
		if merged, ok := q.clauses[0].and(o.clauses[0]); ok {
			return Query{clauses: []clause{merged}, err: err}
		}
	}
	return Query{clauses: q.clauses, err: cmp.Or(err, ErrBadQuery)}
}

func (q Query) Or(o Query) Query {
	return Query{clauses: slices.Concat(q.clauses, o.clauses), err: cmp.Or(q.err, o.err)}
}

func Not(q Query) Query {
	if len(q.clauses) != 1 {
		return Query{clauses: q.clauses, err: cmp.Or(q.err, ErrBadQuery)}
	}
	c := q.clauses[0]
	c.not = !c.not
	return Query{clauses: []clause{c}, err: q.err}
}

func (q Query) CSS() string {
	if len(q.clauses) == 0 {
		return "all"
	}
	parts := make([]string, len(q.clauses))
	for i, c := range q.clauses {
		parts[i] = c.css()
	}
	return strings.Join(parts, ", ")
}

func (q Query) Err() error { return q.err }

func (c clause) and(o clause) (clause, bool) {
	left, leftOK := c.operand()
	right, rightOK := o.operand()
	if !leftOK || !rightOK {
		return clause{}, false
	}
	if left.mediaType != "" && right.mediaType != "" && left.mediaType != right.mediaType {
		return clause{}, false
	}
	return clause{
		mediaType: cmp.Or(left.mediaType, right.mediaType),
		features:  slices.Concat(left.features, right.features),
	}, true
}

func (c clause) operand() (clause, bool) {
	switch {
	case !c.not:
		return c, true
	case c.mediaType == "":
		return clause{features: []string{"(not " + c.condition() + ")"}}, true
	default:
		return clause{}, false
	}
}

func (c clause) condition() string {
	joined := strings.Join(c.features, " and ")
	if len(c.features) > 1 {
		return "(" + joined + ")"
	}
	return joined
}

func (c clause) css() string {
	if c.mediaType == "" {
		if c.not {
			return "not " + c.condition()
		}
		return strings.Join(c.features, " and ")
	}
	head := c.mediaType
	if c.not {
		head = "not " + head
	}
	return strings.Join(append([]string{head}, c.features...), " and ")
}
