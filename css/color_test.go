package css

import "testing"

func TestHexEnsuresHash(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"#fff", "#fff"},
		{"fff", "#fff"},
		{"#ffffff80", "#ffffff80"},
	}
	for _, c := range cases {
		if got := Hex(c.in).CSS(); got != c.want {
			t.Errorf("Hex(%q).CSS() = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestColorFunctions(t *testing.T) {
	cases := []struct {
		value Value
		want  string
	}{
		{RGB(255, 0, 0), "rgb(255, 0, 0)"},
		{RGBA(255, 0, 0, 0.5), "rgba(255, 0, 0, 0.5)"},
		{RGBA(0, 0, 0, 0), "rgba(0, 0, 0, 0)"},
		{HSL(200, 50, 40), "hsl(200, 50%, 40%)"},
		{HSL(200.5, 12.5, 40), "hsl(200.5, 12.5%, 40%)"},
		{HSLA(200, 50, 40, 0.5), "hsla(200, 50%, 40%, 0.5)"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}
