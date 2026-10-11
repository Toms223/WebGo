package container

import (
	"errors"
	"testing"
)

func TestContainerQueryCSS(t *testing.T) {
	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"a lone name", Named("side"), "side"},
		{"a lone feature", features("(min-width: 400px)"), "(min-width: 400px)"},
		{"a name with a feature", Named("side").And(features("(min-width: 400px)")), "side (min-width: 400px)"},
		{"two features join with and", features("(min-width: 400px)").And(features("(orientation: portrait)")), "(min-width: 400px) and (orientation: portrait)"},
		{"a negated feature", Not(features("(min-width: 400px)")), "not (min-width: 400px)"},
		{"a negated feature after a name", Named("side").And(Not(features("(min-width: 400px)"))), "side not (min-width: 400px)"},
		{"a negated named query keeps the name first", Not(Named("side").And(features("(b)"))), "side not (b)"},
		{"several negated features", Not(features("(a)", "(b)")), "not ((a) and (b))"},
		{"or joins parenthesised clauses", features("(min-width: 400px)").Or(features("(orientation: portrait)")), "(min-width: 400px) or (orientation: portrait)"},
		{"or parenthesises a conjunction", features("(a)", "(b)").Or(features("(c)")), "((a) and (b)) or (c)"},
		{"or parenthesises a negation", Not(features("(a)")).Or(features("(b)")), "(not (a)) or (b)"},
		{"double negation cancels", Not(Not(features("(a)"))), "(a)"},
		{"and chains features", features("(a)").And(features("(b)")).And(features("(c)")), "(a) and (b) and (c)"},
		{"a negated operand folds into a conjunction", Not(features("(a)")).And(features("(b)")), "(not (a)) and (b)"},
		{"or keeps every clause", features("(a)").Or(features("(b)")).Or(features("(c)")), "(a) or (b) or (c)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.CSS(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
			if err := tc.query.Err(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerQueryErrors(t *testing.T) {
	cases := []struct {
		name  string
		query Query
		want  error
	}{
		{"two names", Named("a").And(Named("b")), ErrBadQuery},
		{"the same name twice", Named("a").And(Named("a")), ErrBadQuery},
		{"an empty name", Named(""), ErrBadQuery},
		{"a name on the right of and", features("(a)").And(Named("side")), ErrBadQuery},
		{"a name on the right with a feature", features("(a)").And(Named("side").And(features("(b)"))), ErrBadQuery},
		{"or with a named left side", Named("side").Or(features("(a)")), ErrBadQuery},
		{"or with a named right side", features("(a)").Or(Named("side")), ErrBadQuery},
		{"or with a named side and features", Named("side").And(features("(a)")).Or(features("(b)")), ErrBadQuery},
		{"the zero value", Query{}, ErrBadQuery},
		{"not of a lone name", Not(Named("side")), ErrBadQuery},
		{"not of several clauses", Not(features("(a)").Or(features("(b)"))), ErrBadQuery},
		{"and on a query with several clauses", features("(a)").Or(features("(b)")).And(features("(c)")), ErrBadQuery},
		{"and with several clauses on the right", features("(a)").And(features("(b)").Or(features("(c)"))), ErrBadQuery},
		{"and with the zero value on the left", Query{}.And(features("(a)")), ErrBadQuery},
		{"and with the zero value on the right", features("(a)").And(Query{}), ErrBadQuery},
		{"or with the zero value", features("(a)").Or(Query{}), ErrBadQuery},
		{"a name with a feature", Named("side").And(features("(a)")), nil},
		{"a lone name", Named("side"), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Err(); !errors.Is(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestContainerQueryZeroValueRendersEmpty(t *testing.T) {
	if got := (Query{}).CSS(); got != "" {
		t.Fatalf("expected an empty string, got %q", got)
	}
}

func TestContainerQueryKeepsTheFirstError(t *testing.T) {
	first := errors.New("first")
	failed := Query{err: first}
	later := Named("a").And(Named("b"))

	cases := []struct {
		name  string
		query Query
		want  error
	}{
		{"or keeps the receiver error", failed.Or(features("(a)")), first},
		{"or keeps the argument error", features("(a)").Or(failed), first},
		{"or keeps the receiver error over the argument error", later.Or(failed), ErrBadQuery},
		{"and keeps the receiver error", failed.And(features("(a)")), first},
		{"and keeps the argument error", features("(a)").And(failed), first},
		{"and keeps the receiver error over the argument error", later.And(failed), ErrBadQuery},
		{"and keeps the error when the operands merge", later.And(features("(a)")), ErrBadQuery},
		{"not keeps the error", Not(failed), first},
		{"not keeps the error when the clause negates", Not(later.And(features("(a)"))), ErrBadQuery},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Err(); !errors.Is(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestContainerQueryOperationsLeaveTheirOperandsUntouched(t *testing.T) {
	side := Named("side")
	a := features("(a)")
	both := features("(a)").Or(features("(b)"))

	_ = Not(a)
	_ = side.And(a)
	_ = a.And(a)
	_ = both.Or(a)
	_ = a.Or(both)

	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"a name", side, "side"},
		{"a feature", a, "(a)"},
		{"a list of clauses", both, "(a) or (b)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.CSS(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
			if err := tc.query.Err(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestContainerQueryDerivedQueriesDoNotShareStorage(t *testing.T) {
	clauses := features("(a)").Or(features("(b)")).Or(features("(c)"))
	conditions := features("(a)").And(features("(b)")).And(features("(c)")).And(features("(d)"))

	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"the base with several clauses", clauses, "(a) or (b) or (c)"},
		{"a first clause added to it", clauses.Or(features("(x)")), "(a) or (b) or (c) or (x)"},
		{"a second clause added to it", clauses.Or(features("(y)")), "(a) or (b) or (c) or (y)"},
		{"the base with several features", conditions, "(a) and (b) and (c) and (d)"},
		{"a first feature added to it", conditions.And(features("(x)")), "(a) and (b) and (c) and (d) and (x)"},
		{"a second feature added to it", conditions.And(features("(y)")), "(a) and (b) and (c) and (d) and (y)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.CSS(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
