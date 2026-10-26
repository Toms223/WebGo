package doc

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/example"
)

func panel(source, css string) *Panel {
	return &Panel{Example: example.Example{
		Name:   "card.go",
		Source: source,
		CSS:    css,
		Live:   func() app.UI { return app.Span().Class("demo-x") },
	}}
}

func TestPanelShowsTheGoSourceFirst(t *testing.T) {
	got := html(panel("var Card = 1", ".demo-card {\n}\n"))
	for _, want := range []string{">card.go<", "Go</button>", "CSS</button>", "Result", "demo-x", ">Copy<"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
	if strings.Contains(got, "show") {
		t.Errorf("a short source must not offer to expand: %s", got)
	}
}

func TestPanelCSSTabRenamesTheFileAndCopiesTheCSS(t *testing.T) {
	p := panel("var Card = 1", ".demo-card {\n\tcolor: red;\n}\n")
	p.showCSS = true
	got := html(p)
	for _, want := range []string{">card.css<", `<span class="f">.demo-card {</span>`} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
	if strings.Contains(got, ">card.go<") {
		t.Errorf("expected the file name to follow the tab, got %s", got)
	}
}

func TestPanelClampsLongSourceUntilExpanded(t *testing.T) {
	long := strings.Repeat("x := 1\n", clampLines+5) + "y := 2"
	p := panel(long, "")
	got := html(p)
	for _, want := range []string{"example-code-clamped", "Show all 24 lines"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in %s", want, got)
		}
	}
	p.expanded = true
	got = html(p)
	if strings.Contains(got, "example-code-clamped") || !strings.Contains(got, "Show less") {
		t.Errorf("expected the expanded panel to drop the clamp, got %s", got)
	}
}
