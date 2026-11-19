package openfeature

import (
	"errors"
	"fmt"
	"slices"
	"testing"
)

func TestCodeOf(t *testing.T) {
	if codeOf(NewError(CodeFlagNotFound, "x")) != CodeFlagNotFound {
		t.Fatal("direct")
	}
	if codeOf(errors.New("x")) != CodeGeneral {
		t.Fatal("plain")
	}
	if codeOf(fmt.Errorf("w: %w", NewError(CodeParseError, "x"))) != CodeParseError {
		t.Fatal("wrapped")
	}
	e := NewError(CodeGeneral, "msg")
	if e.Error() != "msg" || e.Unwrap() != nil {
		t.Fatal("error methods")
	}
}

type fakeJS struct {
	s       string
	stack   string
	isError bool
}

func (f fakeJS) Error() string    { return "msg" }
func (f fakeJS) JSString() string { return f.s }
func (f fakeJS) JSIsError() bool  { return f.isError }
func (f fakeJS) JSField(n string) (string, bool) {
	if n == "stack" && f.stack != "" {
		return f.stack, true
	}
	return "", false
}

type lineLogger struct{ lines []string }

func (l *lineLogger) Error(a ...any) {
	if len(a) == 0 {
		l.lines = append(l.lines, "<none>")
		return
	}
	l.lines = append(l.lines, fmt.Sprint(a[0]))
}
func (l *lineLogger) Warn(...any)  {}
func (l *lineLogger) Info(...any)  {}
func (l *lineLogger) Debug(...any) {}

func TestJSThrownLogLines(t *testing.T) {
	cases := []struct {
		err  fakeJS
		want []string
	}{
		{fakeJS{s: "TypeError: x", stack: "S", isError: true}, []string{"Unhandled error during 'finally' hook: TypeError: x", "S", "S"}},
		{fakeJS{s: "[object Object]"}, []string{"Unhandled error during 'finally' hook: [object Object]", "<none>"}},
	}
	for _, tc := range cases {
		setNum(t, 1)
		l := &lineLogger{}
		c := GetClient()
		c.SetLogger(l)
		c.IntDetails("f", 0, WithHooks(Hook{Finally: func(HookContext, EvaluationDetails[any], HookHints) { panic(&Error{Code: CodeGeneral, Cause: tc.err}) }}))
		if !slices.Equal(l.lines, tc.want) {
			t.Errorf("lines = %q, want %q", l.lines, tc.want)
		}
	}
	if got := jsProp(fakeJS{}, "message", "x"); got != "undefined" {
		t.Errorf("jsProp of absent field = %q", got)
	}
}
