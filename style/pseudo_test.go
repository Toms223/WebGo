package style

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
)

func TestGeneratedPseudo(t *testing.T) {
	cases := []struct {
		rule *Rule
		want []string
	}{
		{Class("a").Hover(css.Width(css.Px(1))), []string{".a:hover {"}},
		{Class("a").Before(css.Content(css.String("→"))), []string{"::before", `content: "→"`}},
		{Class("li").NthChild(css.An(2, 1), css.Width(css.Px(1))), []string{":nth-child(2n+1)"}},
		{Class("a").Dir(kw.Rtl, css.Width(css.Px(1))), []string{":dir(rtl)"}},
		{Class("q").Lang("en", css.Width(css.Px(1))), []string{".q:lang(en) {"}},
		{Class("b").FirstChild(css.Width(css.Px(1))), []string{".b:first-child {"}},
	}
	for _, c := range cases {
		got := c.rule.CSS()
		for _, want := range c.want {
			if !strings.Contains(got, want) {
				t.Errorf("CSS() = %q, want it to contain %q", got, want)
			}
		}
	}
}
