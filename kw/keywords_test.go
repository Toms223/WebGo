package kw_test

import (
	"reflect"
	"testing"

	"github.com/Toms223/WebGo/kw"
)

func TestKeywordsRender(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{kw.InlineFlex.CSS(), "inline-flex"},
		{kw.Rebeccapurple.CSS(), "rebeccapurple"},
		{kw.Auto.CSS(), "auto"},
		{kw.Dark.CSS(), "dark"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("CSS() = %q, want %q", c.got, c.want)
		}
	}
}

func TestEveryKeywordHasItsOwnType(t *testing.T) {
	if reflect.TypeOf(kw.Auto) == reflect.TypeOf(kw.None) {
		t.Error("kw.Auto and kw.None share a type, so every interface one satisfies the other satisfies too")
	}
}
