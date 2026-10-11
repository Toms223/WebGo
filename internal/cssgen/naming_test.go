package main

import (
	"slices"
	"testing"
)

func TestGoName(t *testing.T) {
	cases := []struct {
		css  string
		want string
	}{
		{"background-color", "BackgroundColor"},
		{"-webkit-line-clamp", "WebkitLineClamp"},
		{"x-small", "XSmall"},
		{"::first-line", "FirstLine"},
		{":nth-child()", "NthChild"},
		{"currentColor", "Currentcolor"},
		{"inline-flex", "InlineFlex"},
		{"width", "Width"},
	}
	for _, c := range cases {
		if got := goName(c.css); got != c.want {
			t.Errorf("goName(%q) = %q, want %q", c.css, got, c.want)
		}
	}
}

func TestKeywordType(t *testing.T) {
	if got := keywordType("Default"); got != "kDefault" {
		t.Errorf(`keywordType("Default") = %q, want "kDefault"`, got)
	}
}

func TestReservedNames(t *testing.T) {
	cases := []struct {
		list  string
		names []string
		want  []string
	}{
		{"reservedRule", reservedRule, []string{"Nest", "Bind", "Container", "Supports", "Self", "CSSFor"}},
		{"reservedCSS", reservedCSS, []string{"Descriptor", "SyntaxColor", "Format", "FormatValue", "SyntaxValue", "SyntaxRaw", "SyntaxAny"}},
		{"reservedContainer", reservedContainer, []string{"Query", "Named", "Not", "ErrBadQuery"}},
	}
	for _, c := range cases {
		for _, name := range c.want {
			if !slices.Contains(c.names, name) {
				t.Errorf("%s lacks %s", c.list, name)
			}
		}
	}
}
