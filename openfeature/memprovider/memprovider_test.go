package memprovider

import (
	"fmt"
	"slices"
	"testing"

	"github.com/Toms223/WebGo/openfeature"
)

type nopLog struct{}

func (nopLog) Error(...any) {}
func (nopLog) Warn(...any)  {}
func (nopLog) Info(...any)  {}
func (nopLog) Debug(...any) {}

func TestTypeMismatchUsesJSTypeof(t *testing.T) {
	p := New(
		Flag{Key: "n", Variants: map[string]any{"a": 3}, DefaultVariant: "a"},
		Flag{Key: "b", Variants: map[string]any{"a": true}, DefaultVariant: "a"},
		Flag{Key: "z", Variants: map[string]any{"a": nil}, DefaultVariant: "a"},
	)
	if r := p.ResolveNumber("n", 1.5, openfeature.Context{}, nopLog{}); r.ErrorCode != "" || r.Value != 3 {
		t.Errorf("int variant vs float64 default: %+v", r)
	}
	if r := p.ResolveString("b", "s", openfeature.Context{}, nopLog{}); r.ErrorCode != openfeature.CodeTypeMismatch {
		t.Errorf("bool vs string: %+v", r)
	}
	if r := p.ResolveObject("z", map[string]any{}, openfeature.Context{}, nopLog{}); r.ErrorCode != "" || r.Value != nil {
		t.Errorf("null is an object: %+v", r)
	}
	if r := p.ResolveNumber("z", 1, openfeature.Context{}, nopLog{}); r.ErrorCode != openfeature.CodeTypeMismatch {
		t.Errorf("null vs number: %+v", r)
	}
}

func TestFlagsChangedOrder(t *testing.T) {
	v := map[string]any{"a": true}
	f := func(k string) Flag { return Flag{Key: k, Variants: v, DefaultVariant: "a"} }
	p := New(f("old1"), f("keep"), f("old2"))
	var got []string
	openfeature.Reset()
	defer openfeature.Reset()
	if err := openfeature.SetProviderAndWait(p); err != nil {
		t.Fatal(err)
	}
	openfeature.AddHandler(openfeature.EventConfigurationChanged, func(d openfeature.EventDetails) { got = d.FlagsChanged })
	if err := p.PutConfiguration(f("z"), f("keep"), f("a"), f("z")); err != nil {
		t.Fatal(err)
	}
	if want := []string{"z", "keep", "a", "old1", "old2"}; !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDuplicateKeyLastWins(t *testing.T) {
	p := New(
		Flag{Key: "k", Variants: map[string]any{"a": "first"}, DefaultVariant: "a"},
		Flag{Key: "k", Variants: map[string]any{"a": "last"}, DefaultVariant: "a"},
	)
	if r := p.ResolveString("k", "", openfeature.Context{}, nopLog{}); r.Value != "last" {
		t.Errorf("got %+v", r)
	}
}

type recLog struct {
	nopLog
	lines []string
}

func (l *recLog) Error(a ...any) { l.lines = append(l.lines, "error:"+fmt.Sprint(a...)) }
func (l *recLog) Debug(a ...any) { l.lines = append(l.lines, "debug:"+fmt.Sprint(a...)) }

func TestNumericKindsAndDisabled(t *testing.T) {
	p := New(
		Flag{Key: "u", Variants: map[string]any{"a": uint(4)}, DefaultVariant: "a"},
		Flag{Key: "f", Variants: map[string]any{"a": float32(2.5)}, DefaultVariant: "a"},
		Flag{Key: "d", Variants: map[string]any{"a": "s"}, DefaultVariant: "a", Disabled: true},
	)
	c := openfeature.Context{}
	if r := p.ResolveNumber("u", 1, c, nopLog{}); r.Value != 4 || r.ErrorCode != "" {
		t.Errorf("uint: %+v", r)
	}
	if r := p.ResolveNumber("f", 1, c, nopLog{}); r.Value != 2.5 || r.ErrorCode != "" {
		t.Errorf("float32: %+v", r)
	}
	if r := p.ResolveBoolean("d", true, c, nopLog{}); r.Value != true || r.Reason != openfeature.ReasonDisabled || r.ErrorCode != "" {
		t.Errorf("disabled: %+v", r)
	}
}

func TestVariantEdgesAndLogs(t *testing.T) {
	p := New(
		Flag{Key: "e", Variants: map[string]any{"a": "A"}},
		Flag{Key: "u", Variants: map[string]any{"a": "A"}, DefaultVariant: "a", ContextEvaluator: func(openfeature.Context) string { return "ghost" }},
	)
	c := openfeature.Context{}
	if r := p.ResolveString("e", "x", c, nopLog{}); r.Value != "" || r.Variant != "" || r.Reason != openfeature.ReasonStatic || r.ErrorCode != "" {
		t.Errorf("empty default variant: %+v", r)
	}
	l := &recLog{}
	if r := p.ResolveString("u", "x", c, l); r.ErrorCode != openfeature.CodeGeneral || r.ErrorMessage != "no value associated with variant provided for u found" {
		t.Errorf("unknown variant: %+v", r)
	}
	p.ResolveString("nope", "x", c, l)
	if r := p.ResolveString("toString", "x", c, l); r.ErrorCode != openfeature.CodeGeneral {
		t.Errorf("inherited key: %+v", r)
	}
	want := []string{
		"error:no value associated with variant provided for u found",
		"debug:no flag found with key nope",
		"error:no value associated with variant provided for toString found",
	}
	if !slices.Equal(l.lines, want) {
		t.Errorf("logs %q, want %q", l.lines, want)
	}
}
