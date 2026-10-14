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

func TestCodeBlocksCarryACopyButtonWithTheRawSource(t *testing.T) {
	got := html(Go("x := 1"), CSS(".a {\n\tcolor: red;\n}\n"), Sh("go get x"))
	for _, want := range []string{`class="codeblock"`, `class="lang-go"`, `class="lang-css"`, `class="lang-sh"`} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
	if count := strings.Count(got, "Copy"); count != 3 {
		t.Errorf("expected one copy button per block, found %d in %s", count, got)
	}
}

func TestCopyButtonLabelFollowsItsState(t *testing.T) {
	if got := html(&Copy{Text: "x"}); !strings.Contains(got, ">Copy</button>") {
		t.Errorf("expected a Copy label, got %s", got)
	}
	if got := html(&Copy{Text: "x", done: true}); !strings.Contains(got, ">Copied</button>") {
		t.Errorf("expected a Copied label, got %s", got)
	}
	if got := html(&Copy{Text: "x", Float: true}); !strings.Contains(got, "copy-float") {
		t.Errorf("expected the floating variant, got %s", got)
	}
}

func TestSubHeadingCarriesItsAnchor(t *testing.T) {
	heading := H("Block order").(*Sub)
	if got := html(heading); strings.Contains(got, "id=") {
		t.Errorf("expected no id before the page assigns one, got %s", got)
	}
	heading.ID = "block-order"
	got := html(heading)
	for _, want := range []string{`id="block-order"`, `tabindex="-1"`, "Block order"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
}

func TestHighlightCSSClassifiesSelectorsPropertiesAndStrings(t *testing.T) {
	source := "@media (max-width: 48rem) {\n\t.a > li {\n\t\tcontent: \"x\";\n\t\tpadding: 1rem;\n\t}\n}\n"
	got := html(HighlightCSS(source)...)
	for _, want := range []string{
		`<span class="k">@media (max-width: 48rem) {</span>`,
		`<span class="f">.a &gt; li {</span>`,
		`<span class="t">content</span>`,
		`<span class="s">&#34;x&#34;</span>`,
		`<span class="t">padding</span>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
}

func TestHighlightCSSKeepsEveryByte(t *testing.T) {
	source := ".a {\n\tcolor: red;\n}\n\n.b:hover {\n\tcontent: \"a: b\";\n}\n"
	var text strings.Builder
	for _, u := range HighlightCSS(source) {
		text.WriteString(app.HTMLString(u))
	}
	flat := regexp.MustCompile(`</?span[^>]*>`).ReplaceAllString(text.String(), "")
	flat = strings.NewReplacer("&#34;", "\"", "&gt;", ">").Replace(flat)
	if flat != source {
		t.Fatalf("got %q, want %q", flat, source)
	}
}
