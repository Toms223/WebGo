package base

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var entries = []nav.Entry{
	{Title: "Getting Started", Route: "/"},
	{Title: "Markdown", Route: "/markdown"},
}

type stubScreen struct {
	route   string
	mounted bool
}

func (s *stubScreen) Content() app.UI {
	return app.Div().Class("stub").Text(s.route)
}

func (s *stubScreen) Route() string { return s.route }

func (s *stubScreen) OnMount(ctx app.Context) page.Screen {
	s.mounted = true
	return s
}

func (s *stubScreen) OnDismount() page.Screen {
	s.mounted = false
	return s
}

func (s *stubScreen) Mounted() bool { return s.mounted }

type contextHost struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (h *contextHost) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *contextHost) Render() app.UI {
	return app.Div().Class("host")
}

func captureContext(t *testing.T) app.Context {
	t.Helper()

	var ctx app.Context
	engine := app.NewTestEngine()
	if err := engine.Load(&contextHost{captured: &ctx}); err != nil {
		t.Fatalf("mounting the context host failed: %v", err)
	}
	if err := engine.Load(&contextHost{Tick: 1}); err != nil {
		t.Fatalf("updating the context host failed: %v", err)
	}
	engine.ConsumeAll()

	if ctx.Src() == nil {
		t.Fatal("the context host was never updated, so no Context was captured")
	}
	return ctx
}

func anchorFor(html, title string) string {
	marker := ">" + title + "</a>"
	end := strings.Index(html, marker)
	if end < 0 {
		return ""
	}
	start := strings.LastIndex(html[:end], "<a ")
	if start < 0 {
		return ""
	}
	return html[start : end+len(marker)]
}

func TestNewRejectsMissingScreens(t *testing.T) {
	if _, err := New(entries, nil); err == nil {
		t.Error("expected a nil screen slice to be rejected")
	}
}

func TestLayoutRendersNavigationAndBody(t *testing.T) {
	layout := NewLayout(entries)
	layout.Body(app.Div().Class("content").Text("body"))

	viewModel, err := newShellModel("shell-render")
	if err != nil {
		t.Fatalf("building the shell model failed: %v", err)
	}

	out := app.HTMLString(layout.Draw(viewModel))
	for _, entry := range entries {
		if !strings.Contains(out, entry.Title) {
			t.Errorf("expected %q in the shell, got %s", entry.Title, out)
		}
	}
	if !strings.Contains(out, "body") {
		t.Errorf("expected the screen body inside the shell, got %s", out)
	}
}

func TestLayoutBodyReplacesRatherThanAppends(t *testing.T) {
	layout := NewLayout(entries)
	layout.Body(app.Div().Class("first"))
	layout.Body(app.Div().Class("second"))

	if len(layout.body) != 1 {
		t.Fatalf("expected Body to replace its content, found %d children", len(layout.body))
	}
}

func TestErrorStateIsRendered(t *testing.T) {
	layout := NewLayout(entries)
	out := app.HTMLString(layout.Error(page.ErrNoRoute))
	if !strings.Contains(out, page.ErrNoRoute.Error()) {
		t.Errorf("expected the error to be described, got %s", out)
	}
}

func TestBaseMountsAndRendersWithoutPanicking(t *testing.T) {
	shellPage, err := New(entries, []page.Screen{&stubScreen{route: "/"}})
	if err != nil {
		t.Fatalf("building the base failed: %v", err)
	}

	ctx := captureContext(t)
	if panicked := func() (panicked bool) {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		shellPage.OnMount(ctx)
		return false
	}(); panicked {
		t.Fatal("expected mounting the base not to panic")
	}

	out := app.HTMLString(shellPage.Render())
	if strings.Contains(out, "shell-loading") {
		t.Errorf("expected the shell to have progressed past its loading state, got %s", out)
	}
	if !strings.Contains(out, "shell-main") {
		t.Errorf("expected the mounted shell's main content wrapper, got %s", out)
	}
	if !strings.Contains(out, `class="stub"`) {
		t.Errorf("expected the selected screen's content inside the shell, got %s", out)
	}
}

func TestActivateUpdatesRenderedNavigationHighlight(t *testing.T) {
	shellPage, err := New(entries, []page.Screen{&stubScreen{route: "/markdown"}})
	if err != nil {
		t.Fatalf("building the base failed: %v", err)
	}

	ctx := captureContext(t)
	shellPage.OnMount(ctx)
	Activate(ctx, "/markdown")

	out := app.HTMLString(shellPage.Render())

	activated := anchorFor(out, "Markdown")
	if !strings.Contains(activated, "nav-link-active") {
		t.Errorf("expected the activated route's entry to carry nav-link-active, got %s", activated)
	}

	other := anchorFor(out, "Getting Started")
	if strings.Contains(other, "nav-link-active") {
		t.Errorf("expected only the activated route's entry to carry nav-link-active, got %s", other)
	}
}
