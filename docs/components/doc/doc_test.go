package doc

import (
	"regexp"
	"strings"
	"testing"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var indent = regexp.MustCompile(`\n *`)

func html(ui ...app.UI) string {
	var out strings.Builder
	for _, u := range ui {
		out.WriteString(indent.ReplaceAllString(app.HTMLString(u), ""))
	}
	return out.String()
}

func TestInlineTurnsBackticksIntoCodeSpans(t *testing.T) {
	got := html(P("A `Rule` is one selector, see `CSS()`."))
	for _, want := range []string{"<p>A <code>Rule</code>", "is one selector, see <code>CSS()</code>", ".</p>"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
}

func TestInlineKeepsElementsAndSlices(t *testing.T) {
	got := html(Item("see ", A("/styling", "Styling"), " and ", B("this"), []app.UI{app.Text("!")}))
	for _, want := range []string{`<a href="/styling">Styling</a>`, "<strong>this</strong>", "!</li>"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
}

func TestExternalLinksOpenSafely(t *testing.T) {
	got := html(A("https://example.invalid/x", "x"))
	for _, want := range []string{`rel="noopener noreferrer"`, `target="_blank"`} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
}

func TestHighlightClassifiesGo(t *testing.T) {
	got := html(Highlight("func (c *Counter) bump() int { return css.Px(1) } // one")...)
	for _, want := range []string{
		`<span class="k">func</span>`,
		`<span class="t">Counter</span>`,
		`<span class="f">bump</span>`,
		`<span class="t">int</span>`,
		`<span class="f">Px</span>`,
		`<span class="c">// one</span>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
	if strings.Contains(got, `class="t">css<`) {
		t.Errorf("a package name must not read as a type: %s", got)
	}
}

func TestHighlightKeepsEveryByte(t *testing.T) {
	source := "x := \"a\" + `b` + f(1)"
	var text strings.Builder
	for _, u := range Highlight(source) {
		text.WriteString(app.HTMLString(u))
	}
	flat := strings.NewReplacer("<span class=\"k\">", "", "<span class=\"s\">", "", "<span class=\"f\">", "", "<span class=\"t\">", "", "<span class=\"c\">", "", "</span>", "", "&#34;", "\"").Replace(text.String())
	if flat != source {
		t.Fatalf("got %q, want %q", flat, source)
	}
}
