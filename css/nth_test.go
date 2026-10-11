package css

import "testing"

func TestNth(t *testing.T) {
	cases := []struct {
		value Nth
		want  string
	}{
		{An(2, 1), "2n+1"},
		{An(2, -1), "2n-1"},
		{An(0, 3), "3"},
		{An(1, 0), "n"},
		{An(-1, 3), "-n+3"},
		{An(2, 0), "2n"},
		{An(1, 3), "n+3"},
		{An(-1, 0), "-n"},
		{An(-2, -1), "-2n-1"},
		{An(0, 0), "0"},
		{Odd, "odd"},
		{Even, "even"},
	}
	for _, c := range cases {
		if got := c.value.CSS(); got != c.want {
			t.Errorf("CSS() = %q, want %q", got, c.want)
		}
	}
}
