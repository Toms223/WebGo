package news_feed

import (
	"strings"
	"testing"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func render(t *testing.T, source string) string {
	t.Helper()
	var out strings.Builder
	style := Style{
		Heading:    "heading",
		Subheading: "subheading",
	}
	for _, ui := range RenderMarkdown(source, style) {
		out.WriteString(app.HTMLString(ui))
	}
	return out.String()
}

func TestRenderMarkdownEmitsBlockElements(t *testing.T) {
	cases := []struct {
		name     string
		source   string
		contains []string
	}{
		{"paragraph", "Just words.", []string{"<p>", "Just words."}},
		{"h1 and h2 share the heading style", "# Top\n\n## Second", []string{`<h3 class="heading">`, "Top", "Second"}},
		{"h3 drops to the subheading style", "### Third", []string{`<h4 class="subheading">`, "Third"}},
		{"strong", "**loud**", []string{"<strong>", "loud"}},
		{"emphasis", "*soft*", []string{"<em>", "soft"}},
		{"inline code", "a `snippet` here", []string{"<code>", "snippet"}},
		{"fenced code", "```\nfmt.Println()\n```", []string{"<pre>", "<code>", "fmt.Println()"}},
		{"bullet list", "- one\n- two", []string{"<ul>", "<li>", "one", "two"}},
		{"ordered list", "1. one\n2. two", []string{"<ol>", "<li>", "one"}},
		{"blockquote", "> quoted", []string{"<blockquote>", "quoted"}},
		{"thematic break", "text\n\n---\n\nmore", []string{"<hr>"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			html := render(t, tc.source)
			for _, expected := range tc.contains {
				if !strings.Contains(html, expected) {
					t.Fatalf("expected %q in %s", expected, html)
				}
			}
		})
	}
}

func TestRenderMarkdownKeepsTightListItemsFlat(t *testing.T) {
	html := render(t, "- one\n- two")
	if strings.Contains(html, "<li><p>") {
		t.Fatalf("a tight list item should hold its text directly, got %s", html)
	}
}

func TestRenderMarkdownLinksOpenSafely(t *testing.T) {
	html := render(t, "[docs](https://example.invalid/page)")
	for _, expected := range []string{
		`href="https://example.invalid/page"`,
		`rel="noopener noreferrer"`,
		`target="_blank"`,
		"docs",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected %q in %s", expected, html)
		}
	}
}

func TestRenderMarkdownRefusesDangerousDestinations(t *testing.T) {
	for _, source := range []string{
		"[click](javascript:alert(1))",
		"[click](data:text/html;base64,PHNjcmlwdD4=)",
		"[click](//evil.invalid/path)",
		"[click](vbscript:msgbox)",
	} {
		html := render(t, source)
		if strings.Contains(html, "<a ") || strings.Contains(html, "href") {
			t.Fatalf("expected no anchor for %q, got %s", source, html)
		}
		if !strings.Contains(html, "click") {
			t.Fatalf("expected the link text to survive for %q, got %s", source, html)
		}
	}
}

func TestRenderMarkdownAllowsRelativeAndMailtoDestinations(t *testing.T) {
	for _, source := range []string{"[a](/feed)", "[b](#anchor)", "[c](mailto:x@example.invalid)"} {
		if !strings.Contains(render(t, source), "<a ") {
			t.Fatalf("expected an anchor for %q", source)
		}
	}
}

func TestRenderMarkdownNeverEmitsSourceHTMLAsMarkup(t *testing.T) {
	for _, source := range []string{
		"<script>alert(1)</script>",
		"text <b>bold</b> text",
		`<img src=x onerror=alert(1)>`,
		"<iframe src=\"https://evil.invalid\"></iframe>",
	} {
		html := render(t, source)
		for _, forbidden := range []string{"<script", "<b>", "<img", "<iframe"} {
			if strings.Contains(html, forbidden) {
				t.Fatalf("expected %q to be inert in the output, got %s", forbidden, html)
			}
		}
		if !strings.Contains(html, "&lt;") {
			t.Fatalf("expected the source's angle brackets to be escaped, got %s", html)
		}
	}
}

func TestRenderMarkdownDoesNotEmitImageTags(t *testing.T) {
	html := render(t, "![a diagram](https://example.invalid/i.png)")
	if strings.Contains(html, "<img") {
		t.Fatalf("expected no img tag, got %s", html)
	}
	if !strings.Contains(html, "a diagram") {
		t.Fatalf("expected the alt text to survive, got %s", html)
	}
}

func TestRenderMarkdownJoinsSoftWrappedLines(t *testing.T) {
	html := render(t, "line one\nline two")
	if !strings.Contains(html, "line one ") || !strings.Contains(html, "line two") {
		t.Fatalf("expected a soft break to become a trailing space, got %s", html)
	}
	if strings.Contains(html, "<br>") {
		t.Fatalf("expected a soft break not to become a <br>, got %s", html)
	}
}

func TestRenderMarkdownKeepsHardBreaks(t *testing.T) {
	html := render(t, "line one  \nline two")
	if !strings.Contains(html, "<br>") {
		t.Fatalf("expected a two-space hard break to become a <br>, got %s", html)
	}
}

func TestRenderMarkdownKeepsUnsupportedSyntaxAsContent(t *testing.T) {
	cases := map[string]string{
		"| a | b |\n|---|---|\n| 1 | 2 |": "a",
		"~~gone~~":                        "gone",
		"- [ ] todo":                      "todo",
	}
	for source, expected := range cases {
		html := render(t, source)
		if !strings.Contains(html, expected) {
			t.Fatalf("expected %q to survive %q, got %s", expected, source, html)
		}
	}
}

func TestRenderMarkdownRendersNothingForBlankSource(t *testing.T) {
	for _, source := range []string{"", "   ", "\n\n\t\n"} {
		style := Style{}
		if got := RenderMarkdown(source, style); got != nil {
			t.Fatalf("expected no elements for %q, got %d", source, len(got))
		}
	}
}
