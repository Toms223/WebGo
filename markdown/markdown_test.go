package markdown

import (
	"regexp"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
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

func demoStyle() Style {
	return Style{
		Example:      "example",
		ExampleCode:  "example-code",
		ExampleLive:  "example-live",
		ExampleLabel: "example-label",
		Demos: map[string]func() app.UI{
			"card": func() app.UI { return app.Div().Class("demo-card").Text("live") },
		},
	}
}

var betweenTags = regexp.MustCompile(`\s*(<[^>]*>)\s*`)

func renderWith(t *testing.T, source string, style Style) string {
	t.Helper()
	var out strings.Builder
	for _, ui := range RenderMarkdown(source, style) {
		out.WriteString(app.HTMLString(ui))
	}
	return betweenTags.ReplaceAllString(out.String(), "$1")
}

func TestFencedCodeBlockCarriesItsLanguage(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   string
	}{
		{"language", "```go\nx := 1\n```", `<pre class="lang-go">`},
		{"language with an unknown demo", "```go missing\nx := 1\n```", `<pre class="lang-go">`},
		{"no language", "```\nx := 1\n```", "<pre>"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			html := renderWith(t, tc.source, demoStyle())
			if !strings.Contains(html, tc.want) {
				t.Fatalf("expected %q in %s", tc.want, html)
			}
			if strings.Contains(html, "example") {
				t.Fatalf("expected no example wrapper in %s", html)
			}
		})
	}
}

func TestDemoFenceRendersTheLiveElement(t *testing.T) {
	html := renderWith(t, "```go card\nx := 1\n```", demoStyle())
	for _, expected := range []string{
		`<div class="example">`,
		`<div class="example-label">Go</div>`,
		`<div class="example-code"><pre class="lang-go"><code>x := 1</code></pre></div>`,
		`<div class="example-label">Result</div>`,
		`<div class="example-live"><div class="demo-card">live</div></div>`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected %q in %s", expected, html)
		}
	}
}

func TestDemoFenceWithEmptyClassNamesDropsTheClassAttributes(t *testing.T) {
	style := Style{Demos: demoStyle().Demos}
	html := renderWith(t, "```go card\nx := 1\n```", style)
	if strings.Contains(html, `class=""`) {
		t.Fatalf("expected no empty class attributes in %s", html)
	}
	if !strings.Contains(html, `<div class="demo-card">live</div>`) {
		t.Fatalf("expected the demo in %s", html)
	}
}

func TestUnregisteredDemoRendersCodeOnly(t *testing.T) {
	for name, style := range map[string]Style{
		"unregistered name": demoStyle(),
		"nil demos":         {Example: "example"},
	} {
		t.Run(name, func(t *testing.T) {
			html := renderWith(t, "```go other\nx := 1\n```", style)
			if strings.Contains(html, "example") || strings.Contains(html, "demo-card") {
				t.Fatalf("expected the code block alone in %s", html)
			}
			if !strings.Contains(html, `<pre class="lang-go"><code>x := 1</code></pre>`) {
				t.Fatalf("expected the code in %s", html)
			}
		})
	}
}

func TestIndentedCodeBlockStaysPlain(t *testing.T) {
	html := renderWith(t, "text\n\n    x := 1\n", demoStyle())
	if !strings.Contains(html, "<pre><code>x := 1</code></pre>") {
		t.Fatalf("expected a plain pre in %s", html)
	}
}

func TestDemoNames(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []string
	}{
		{"second words in order", "```go card\n```\n```css\n```\n```go accent\n```", []string{"card", "accent"}},
		{"none", "```go\n```", nil},
		{"blank", "", nil},
		{"ignores indented code", "    code\n", nil},
		{"fences inside lists", "- item\n\n  ```go card\n  x\n  ```", []string{"card"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := DemoNames(tc.source)
			if strings.Join(got, ",") != strings.Join(tc.want, ",") {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
