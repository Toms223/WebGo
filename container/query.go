// Package container builds @container queries for style.Rule.Container from generated size and
// orientation features.
package container

import (
	"cmp"
	"errors"
	"slices"
	"strings"
)

// ErrBadQuery is carried by a combination that cannot be written as one container query. Operations on
// a query keep its first error.
var ErrBadQuery = errors.New("container: query cannot be expressed as one container query")

// Query is a container query, optionally naming the container it targets. The zero Query carries
// ErrBadQuery.
type Query struct {
	name    string
	clauses []clause
	err     error
}

type clause struct {
	not      bool
	features []string
}

// Named targets the container called name; add conditions with And. It must come first, and an empty
// name gives ErrBadQuery.
func Named(name string) Query {
	if name == "" {
		return Query{err: ErrBadQuery}
	}
	return Query{name: name}
}

func features(texts ...string) Query {
	return Query{clauses: []clause{{features: texts}}}
}

// And joins conditions with and. A Named argument or an Or list on either side gives ErrBadQuery.
func (q Query) And(o Query) Query {
	err := cmp.Or(q.Err(), o.Err())
	if o.name != "" || len(q.clauses) > 1 || len(o.clauses) > 1 {
		return Query{name: q.name, clauses: q.clauses, err: cmp.Or(err, ErrBadQuery)}
	}
	merged := q.clauses
	if len(q.clauses) == 1 && len(o.clauses) == 1 {
		merged = []clause{q.clauses[0].and(o.clauses[0])}
	} else if len(o.clauses) == 1 {
		merged = o.clauses
	}
	return Query{name: q.name, clauses: merged, err: err}
}

// Or joins conditions with or. A named query on either side gives ErrBadQuery.
func (q Query) Or(o Query) Query {
	err := cmp.Or(q.Err(), o.Err())
	if q.name != "" || o.name != "" {
		err = cmp.Or(err, ErrBadQuery)
	}
	return Query{name: q.name, clauses: slices.Concat(q.clauses, o.clauses), err: err}
}

// Not negates a single condition; an Or list gives ErrBadQuery.
func Not(q Query) Query {
	if len(q.clauses) != 1 {
		return Query{name: q.name, clauses: q.clauses, err: cmp.Or(q.Err(), ErrBadQuery)}
	}
	c := q.clauses[0]
	c.not = !c.not
	return Query{name: q.name, clauses: []clause{c}, err: q.err}
}

func (q Query) CSS() string {
	switch len(q.clauses) {
	case 0:
		return q.name
	case 1:
		return strings.TrimSpace(q.name + " " + q.clauses[0].css())
	}
	parts := make([]string, len(q.clauses))
	for i, c := range q.clauses {
		parts[i] = c.group()
	}
	return strings.Join(parts, " or ")
}

func (q Query) Err() error {
	if q.err == nil && q.name == "" && len(q.clauses) == 0 {
		return ErrBadQuery
	}
	return q.err
}

func (c clause) and(o clause) clause {
	return clause{features: slices.Concat(c.operand(), o.operand())}
}

func (c clause) operand() []string {
	if c.not {
		return []string{"(not " + c.condition() + ")"}
	}
	return c.features
}

func (c clause) condition() string {
	joined := strings.Join(c.features, " and ")
	if len(c.features) > 1 {
		return "(" + joined + ")"
	}
	return joined
}

func (c clause) css() string {
	if c.not {
		return "not " + c.condition()
	}
	return strings.Join(c.features, " and ")
}

func (c clause) group() string {
	if c.not {
		return "(" + c.css() + ")"
	}
	return c.condition()
}
