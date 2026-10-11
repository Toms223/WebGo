package css

import "testing"

func TestStringEscapes(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"→", `"→"`},
		{`a"b`, `"a\"b"`},
		{`a\b`, `"a\\b"`},
		{"a\nb", `"a\a b"`},
		{"a\rb", `"a\d b"`},
		{"a\fb", `"a\c b"`},
		{`a\"b`, `"a\\\"b"`},
		{"", `""`},
	}
	for _, c := range cases {
		if got := String(c.in).CSS(); got != c.want {
			t.Errorf("String(%q).CSS() = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestUrlQuotes(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"x.png", `url("x.png")`},
		{`a"b.png`, `url("a\"b.png")`},
		{"a\nb.png", `url("a\a b.png")`},
	}
	for _, c := range cases {
		if got := Url(c.in).CSS(); got != c.want {
			t.Errorf("Url(%q).CSS() = %s, want %s", c.in, got, c.want)
		}
	}
}

func TestIdentAndRawPassThrough(t *testing.T) {
	if got := Ident("fade").CSS(); got != "fade" {
		t.Errorf("Ident CSS() = %q, want %q", got, "fade")
	}
	if got := Raw("1px solid").CSS(); got != "1px solid" {
		t.Errorf("Raw CSS() = %q, want %q", got, "1px solid")
	}
}
