package media

import (
	"errors"
	"testing"
)

func TestQueryRenders(t *testing.T) {
	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"a media type", Screen(), "screen"},
		{"the zero value", Query{}, "all"},
		{"a lone feature", features("(hover: hover)"), "(hover: hover)"},
		{"a type adopts a feature on its right", Screen().And(features("(max-width: 600px)")), "screen and (max-width: 600px)"},
		{"a type on the right is adopted", features("(a)").And(Screen()), "screen and (a)"},
		{"two typeless features join", features("(a)").And(features("(b)")), "(a) and (b)"},
		{"a negated feature", Not(features("(hover)")), "not (hover)"},
		{"several negated features", Not(features("(a)", "(b)")), "not ((a) and (b))"},
		{"a negated typed clause", Not(Screen().And(features("(a)"))), "not screen and (a)"},
		{"a negated typeless operand folds into a type", Screen().And(Not(features("(hover)"))), "screen and (not (hover))"},
		{"or appends clauses", Screen().Or(Print()), "screen, print"},
		{"one clause holding several features", features("(a)", "(b)"), "(a) and (b)"},
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

func TestQueryErrors(t *testing.T) {
	cases := []struct {
		name  string
		query Query
		want  error
	}{
		{"two different media types", Screen().And(Print()), ErrBadQuery},
		{"and on a query with several clauses", Screen().Or(Print()).And(features("(a)")), ErrBadQuery},
		{"a type with a feature", Screen().And(features("(a)")), nil},
		{"or keeps the error of an errored query", Screen().And(Print()).Or(Screen()), ErrBadQuery},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Err(); !errors.Is(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestQueryComposition(t *testing.T) {
	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"a negated typeless operand on the left folds into a type", Not(features("(hover)")).And(Screen()), "screen and (not (hover))"},
		{"several negated features fold as one condition", Screen().And(Not(features("(a)", "(b)"))), "screen and (not ((a) and (b)))"},
		{"folding keeps operand order", Not(features("(b)")).And(features("(a)")), "(not (b)) and (a)"},
		{"two negated typeless operands fold into one clause", Not(features("(a)")).And(Not(features("(b)"))), "(not (a)) and (not (b))"},
		{"the same media type merges", Screen().And(Screen()), "screen"},
		{"and chains features", Screen().And(features("(a)")).And(features("(b)")), "screen and (a) and (b)"},
		{"a negated media type", Not(Screen()), "not screen"},
		{"double negation cancels on a typed clause", Not(Not(Screen().And(features("(a)")))), "screen and (a)"},
		{"double negation cancels on a typeless clause", Not(Not(features("(a)"))), "(a)"},
		{"or keeps every clause", Screen().Or(Print()).Or(All()), "screen, print, all"},
		{"or keeps negated clauses", Not(Screen()).Or(features("(a)")), "not screen, (a)"},
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

func TestQueryRejectsCombinationsItCannotExpress(t *testing.T) {
	cases := []struct {
		name  string
		query Query
	}{
		{"not of several clauses", Not(Screen().Or(Print()))},
		{"not of the zero value", Not(Query{})},
		{"and with a negated typed clause on the left", Not(Screen()).And(features("(a)"))},
		{"and with a negated typed clause on the right", features("(a)").And(Not(Screen()))},
		{"and with several clauses on the right", Screen().And(Screen().Or(Print()))},
		{"and with the zero value on the left", Query{}.And(Screen())},
		{"and with the zero value on the right", Screen().And(Query{})},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Err(); !errors.Is(got, ErrBadQuery) {
				t.Fatalf("expected %v, got %v", ErrBadQuery, got)
			}
		})
	}
}

func TestQueryKeepsTheFirstError(t *testing.T) {
	first := errors.New("first")
	failed := Query{err: first}
	later := Screen().And(Print())

	cases := []struct {
		name  string
		query Query
		want  error
	}{
		{"or keeps the receiver error", failed.Or(later), first},
		{"or keeps the receiver error over the argument error", later.Or(failed), ErrBadQuery},
		{"or keeps the argument error", Screen().Or(failed), first},
		{"and keeps the receiver error over a new failure", failed.And(Screen()), first},
		{"and keeps the argument error over a new failure", Screen().And(failed), first},
		{"and keeps the receiver error over the argument error", later.And(failed), ErrBadQuery},
		{"and keeps the error when the operands merge", later.And(features("(a)")), ErrBadQuery},
		{"and keeps the argument error when the operands merge", features("(a)").And(later), ErrBadQuery},
		{"not keeps the error over a new failure", Not(failed), first},
		{"not keeps the error when the clause negates", Not(later), ErrBadQuery},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Err(); !errors.Is(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestQueryOperationsLeaveTheirOperandsUntouched(t *testing.T) {
	screen := Screen()
	hover := features("(hover)")
	both := Screen().Or(Print())

	_ = Not(screen)
	_ = Not(hover)
	_ = screen.And(hover)
	_ = hover.And(screen)
	_ = both.Or(screen)
	_ = screen.Or(both)

	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"a media type", screen, "screen"},
		{"a feature", hover, "(hover)"},
		{"a list of media types", both, "screen, print"},
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

func TestQueryDerivedQueriesDoNotShareStorage(t *testing.T) {
	clauses := Screen().Or(Print()).Or(All())
	conditions := Screen().And(features("(a)")).And(features("(b)")).And(features("(c)"))

	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"the base with several clauses", clauses, "screen, print, all"},
		{"a first clause added to it", clauses.Or(features("(hover)")), "screen, print, all, (hover)"},
		{"a second clause added to it", clauses.Or(features("(pointer)")), "screen, print, all, (pointer)"},
		{"the base with several features", conditions, "screen and (a) and (b) and (c)"},
		{"a first feature added to it", conditions.And(features("(d)")), "screen and (a) and (b) and (c) and (d)"},
		{"a second feature added to it", conditions.And(features("(e)")), "screen and (a) and (b) and (c) and (e)"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.CSS(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
