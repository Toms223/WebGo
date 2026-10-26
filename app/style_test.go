package app

import (
	"fmt"
	"strings"
	"testing"
)

type testClassRule []string

func (r testClassRule) ClassNames() []string { return r }

func TestStyleAddsClassNames(t *testing.T) {
	cases := []struct {
		name string
		ui   UI
		want string
	}{
		{"one rule", Div().Style(testClassRule{"a"}), `<div class="a"></div>`},
		{"two rules", Div().Style(testClassRule{"a"}, testClassRule{"b", "c"}), `<div class="a b c"></div>`},
		{"after Class", Div().Class("x").Style(testClassRule{"a"}), `<div class="x a"></div>`},
		{"before Class", Div().Style(testClassRule{"a"}).Class("x"), `<div class="a x"></div>`},
		{"empty rule", Div().Style(testClassRule{}), `<div></div>`},
		{"nil rule", Div().Style(nil), `<div></div>`},
	}
	for _, c := range cases {
		if got := HTMLString(c.ui); got != c.want {
			t.Errorf("%s: got %s, want %s", c.name, got, c.want)
		}
	}
}

func TestStyleLogsSkippedRules(t *testing.T) {
	var logged []string
	defer func(l func(string, ...any)) { DefaultLogger = l }(DefaultLogger)
	DefaultLogger = func(format string, v ...any) { logged = append(logged, fmt.Sprintf(format, v...)) }
	Div().Style(testClassRule{}, testClassRule{"a"}, nil)
	if len(logged) != 2 || !strings.Contains(strings.Join(logged, " "), "no class names") {
		t.Fatalf("logged %q, want two skip messages", logged)
	}
}

func TestRawStyleWritesInlineStyle(t *testing.T) {
	if got := HTMLString(Div().RawStyle("color", "red")); got != `<div style="color:red"></div>` {
		t.Fatalf("got %s", got)
	}
	if got := HTMLString(Div().RawStylef("width", "%dpx", 4)); got != `<div style="width:4px"></div>` {
		t.Fatalf("got %s", got)
	}
}
