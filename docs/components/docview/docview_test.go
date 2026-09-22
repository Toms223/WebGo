package docview

import (
	"strings"
	"testing"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func render(t *testing.T, source string) string {
	t.Helper()
	return app.HTMLString(Render(source))
}

func TestRenderAppliesHeadingClasses(t *testing.T) {
	out := render(t, "# Title\n\n### Deeper\n")
	if !strings.Contains(out, `<h3 class="doc-heading">`) {
		t.Errorf("expected a top-level heading with the doc-heading class, got %s", out)
	}
	if !strings.Contains(out, `<h4 class="doc-subheading">`) {
		t.Errorf("expected a deeper heading with the doc-subheading class, got %s", out)
	}
}

func TestRenderEmptySourceIsMarked(t *testing.T) {
	out := render(t, "   ")
	if !strings.Contains(out, "doc-empty") {
		t.Errorf("expected an empty source to render the doc-empty marker, got %s", out)
	}
}

func TestRenderKeepsCodeBlocks(t *testing.T) {
	out := render(t, "```go\nfmt.Println()\n```\n")
	if !strings.Contains(out, "<pre>") || !strings.Contains(out, "<code>") {
		t.Errorf("expected a fenced block to render as pre/code, got %s", out)
	}
}

func TestRenderDowngradesUnsafeLinks(t *testing.T) {
	out := render(t, "[click](javascript:alert(1))")
	if strings.Contains(out, "javascript:") {
		t.Errorf("an unsafe destination must not reach the output, got %s", out)
	}
	if !strings.Contains(out, "<span>") {
		t.Errorf("expected an unsafe link to render as a span, got %s", out)
	}
}

func TestRenderKeepsSafeLinks(t *testing.T) {
	out := render(t, "[docs](https://example.com)")
	if !strings.Contains(out, `href="https://example.com"`) {
		t.Errorf("expected a safe destination to survive, got %s", out)
	}
	if !strings.Contains(out, `rel="noopener noreferrer"`) {
		t.Errorf("expected the rel attribute on external links, got %s", out)
	}
}
