package supports

import (
	"errors"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
)

func TestSupportsQueryCSS(t *testing.T) {
	grid := Property(css.Display(kw.Grid))
	flex := Property(css.Display(kw.Flex))
	sel := Selector(":has(a)")

	cases := []struct {
		name  string
		query Query
		want  string
	}{
		{"a property", grid, "(display: grid)"},
		{"a selector", sel, "selector(:has(a))"},
		{"a font format", FontFormat(css.Raw("woff2")), "font-format(woff2)"},
		{"a font tech", FontTech(css.Raw("variations")), "font-tech(variations)"},
		{"and", grid.And(sel), "(display: grid) and selector(:has(a))"},
		{"or", grid.Or(flex), "(display: grid) or (display: flex)"},
		{"not", Not(grid), "not (display: grid)"},
		{"and chains", grid.And(flex).And(sel), "((display: grid) and (display: flex)) and selector(:has(a))"},
		{"a conjunction wrapped under or", grid.And(flex).Or(sel), "((display: grid) and (display: flex)) or selector(:has(a))"},
		{"a conjunction wrapped under or on the right", sel.Or(grid.And(flex)), "selector(:has(a)) or ((display: grid) and (display: flex))"},
		{"a disjunction wrapped under and", grid.Or(flex).And(sel), "((display: grid) or (display: flex)) and selector(:has(a))"},
		{"not of a conjunction", Not(grid.And(flex)), "not ((display: grid) and (display: flex))"},
		{"not of a disjunction", Not(grid.Or(flex)), "not ((display: grid) or (display: flex))"},
		{"double negation", Not(Not(grid)), "not (not (display: grid))"},
		{"a negated operand under and", Not(grid).And(sel), "(not (display: grid)) and selector(:has(a))"},
		{"a negated operand under or", sel.Or(Not(grid)), "selector(:has(a)) or (not (display: grid))"},
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

func TestSupportsQueryErrors(t *testing.T) {
	grid := Property(css.Display(kw.Grid))
	first := errors.New("first")
	failed := Query{err: first}

	cases := []struct {
		name  string
		query Query
		want  error
	}{
		{"a nil property", Property(nil), ErrBadQuery},
		{"an empty selector", Selector(""), ErrBadQuery},
		{"a nil font format", FontFormat(nil), ErrBadQuery},
		{"a nil font tech", FontTech(nil), ErrBadQuery},
		{"the zero value", Query{}, ErrBadQuery},
		{"and with the zero value on the left", Query{}.And(grid), ErrBadQuery},
		{"and with the zero value on the right", grid.And(Query{}), ErrBadQuery},
		{"or with the zero value", grid.Or(Query{}), ErrBadQuery},
		{"not of the zero value", Not(Query{}), ErrBadQuery},
		{"and keeps the receiver error", failed.And(Property(nil)), first},
		{"and keeps the argument error", grid.And(failed), first},
		{"or keeps the receiver error", failed.Or(grid), first},
		{"or keeps the argument error", grid.Or(failed), first},
		{"not keeps the error", Not(failed), first},
		{"a valid property", grid, nil},
		{"a valid combination", grid.And(Selector("a")), nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.query.Err(); !errors.Is(got, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
