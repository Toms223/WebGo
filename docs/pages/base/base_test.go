package base

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var entries = []nav.Entry{
	{Title: "Getting Started", Route: "/start"},
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

func captureContext(t *testing.T) (app.Context, app.TestEngine) {
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
	return ctx, engine
}

func elementFor(html, tag, text string) string {
	marker := ">" + text + "</" + tag + ">"
	end := strings.Index(html, marker)
	if end < 0 {
		return ""
	}
	start := strings.LastIndex(html[:end], "<"+tag+" ")
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

	ctx, _ := captureContext(t)
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
	if !mainContent.MatchString(out) {
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

	ctx, _ := captureContext(t)
	shellPage.OnMount(ctx)
	Activate(ctx, "/markdown")

	out := app.HTMLString(shellPage.Render())

	activated := elementFor(out, "a", "Markdown")
	if !strings.Contains(activated, "nav-link-active") {
		t.Errorf("expected the activated route's entry to carry nav-link-active, got %s", activated)
	}

	other := elementFor(out, "a", "Getting Started")
	if strings.Contains(other, "nav-link-active") {
		t.Errorf("expected only the activated route's entry to carry nav-link-active, got %s", other)
	}
}

func TestSidebarRendersTheToggle(t *testing.T) {
	viewModel, err := newShellModel("shell-toggle")
	if err != nil {
		t.Fatalf("building the shell model failed: %v", err)
	}

	out := app.HTMLString(NewLayout(entries).Draw(viewModel))

	button := elementFor(out, "button", "Dark mode")
	if button == "" {
		t.Fatalf("expected a Dark mode button, got %s", out)
	}
	for _, attribute := range []string{` type="button"`, ` class="shell-toggle"`} {
		if !strings.Contains(button, attribute) {
			t.Errorf("expected the toggle to carry%s, got %s", attribute, button)
		}
	}
	if at := strings.Index(out, button); at < strings.Index(out, "<aside") || at > strings.Index(out, "</aside>") {
		t.Errorf("expected the toggle inside the sidebar, got %s", out)
	}
}

func TestToggleFlipsDark(t *testing.T) {
	t.Cleanup(func() { shell.State().Dark = false })

	ctx, engine := captureContext(t)
	var dispatched []any
	ctx.Handle(theme.DarkAction, func(ctx app.Context, a app.Action) {
		dispatched = append(dispatched, a.Value)
	})
	layout := NewLayout(entries)

	layout.toggle(ctx, app.Event{})
	engine.ConsumeAll()

	if !shell.State().Dark {
		t.Fatal("expected the first toggle to turn Dark on")
	}
	if elementFor(app.HTMLString(layout.Draw(shell)), "button", "Light mode") == "" {
		t.Error("expected the re-rendered toggle to read Light mode")
	}

	layout.toggle(ctx, app.Event{})
	engine.ConsumeAll()

	if shell.State().Dark {
		t.Fatal("expected the second toggle to turn Dark off")
	}
	if elementFor(app.HTMLString(layout.Draw(shell)), "button", "Dark mode") == "" {
		t.Error("expected the re-rendered toggle to read Dark mode again")
	}
	if want := []any{true, false}; !slices.Equal(dispatched, want) {
		t.Errorf("dispatched %s values = %v, want %v", theme.DarkAction, dispatched, want)
	}
}

func TestDrawHasSidebarAndContent(t *testing.T) {
	viewModel, err := newShellModel("shell-sidebar")
	if err != nil {
		t.Fatalf("building the shell model failed: %v", err)
	}
	layout := NewLayout(entries)
	layout.Body(app.Div().Class("stub").Text("body"))

	out := app.HTMLString(layout.Draw(viewModel))

	aside := strings.Index(out, `<aside class="side">`)
	asideEnd := strings.Index(out, "</aside>")
	if aside < 0 || asideEnd < aside {
		t.Fatalf("expected an aside.side sidebar, got %s", out)
	}
	sidebar := out[aside:asideEnd]
	for _, want := range []string{
		`<a class="logo" href="/">`,
		`<span class="logo-mark">W</span>`,
		`class="side-menu"`,
		`class="side-nav"`,
		`id="side-nav"`,
		`aria-controls="side-nav"`,
		`aria-expanded="false"`,
		`class="shell-toggle"`,
	} {
		if !strings.Contains(sidebar, want) {
			t.Errorf("expected %q inside the sidebar, got %s", want, sidebar)
		}
	}
	if menu, navAt := strings.Index(sidebar, "side-menu"), strings.Index(sidebar, "<nav "); menu > navAt {
		t.Errorf("expected the menu button before the navigation, got %s", sidebar)
	}

	main := mainContent.FindStringIndex(out)[0]
	if main < asideEnd {
		t.Fatalf("expected main.content after the sidebar, got %s", out)
	}
	if !strings.Contains(out[main:], `<div class="stub">body</div>`) {
		t.Errorf("expected the screen body inside main.content, got %s", out)
	}
	if !strings.HasPrefix(out, `<div class="shell">`) {
		t.Errorf("expected the closed shell to carry only the shell class, got %s", out)
	}
}

var mainContent = regexp.MustCompile(`<main[^>]*class="content"`)

var expandedTrue = regexp.MustCompile(`aria-expanded[ >]`)

func TestMenuToggle(t *testing.T) {
	t.Cleanup(func() { shell.State().Menu = false })

	ctx, engine := captureContext(t)
	layout := NewLayout(entries)

	layout.toggleMenu(ctx, app.Event{})
	engine.ConsumeAll()

	if !shell.State().Menu {
		t.Fatal("expected the first menu toggle to open the menu")
	}
	if out := app.HTMLString(layout.Draw(shell)); !strings.HasPrefix(out, `<div class="shell menu-open">`) {
		t.Errorf("expected the open shell to carry menu-open, got %s", out)
	}
	if out := app.HTMLString(layout.Draw(shell)); !expandedTrue.MatchString(out) || strings.Contains(out, `aria-expanded="false"`) || !strings.Contains(out, `aria-controls="side-nav"`) {
		t.Errorf("expected the open menu button to carry aria-expanded=\"true\" and to control side-nav, got %s", out)
	}

	layout.toggleMenu(ctx, app.Event{})
	engine.ConsumeAll()

	if shell.State().Menu {
		t.Fatal("expected the second menu toggle to close the menu")
	}
	if out := app.HTMLString(layout.Draw(shell)); strings.Contains(out, "menu-open") {
		t.Errorf("expected the closed shell to drop menu-open, got %s", out)
	}
	if out := app.HTMLString(layout.Draw(shell)); !strings.Contains(out, `aria-expanded="false"`) {
		t.Errorf("expected the closed menu button not to be expanded, got %s", out)
	}
}

func TestActivateClosesTheMenu(t *testing.T) {
	t.Cleanup(func() {
		shell.State().Menu = false
		shell.State().Active = ""
	})

	ctx, engine := captureContext(t)
	NewLayout(entries).toggleMenu(ctx, app.Event{})
	engine.ConsumeAll()
	if !shell.State().Menu {
		t.Fatal("expected the menu to be open before navigating")
	}

	Activate(ctx, "/markdown")
	engine.ConsumeAll()

	if shell.State().Menu {
		t.Error("expected activating a route to close the menu")
	}
	if got := shell.State().Active; got != "/markdown" {
		t.Errorf("expected /markdown to be active, got %q", got)
	}
}

func TestDrawRendersTheLandingShellForHome(t *testing.T) {
	t.Cleanup(func() { shell.State().Active = "" })

	ctx, engine := captureContext(t)
	Activate(ctx, HomeRoute)
	engine.ConsumeAll()

	layout := NewLayout(entries)
	layout.Body(app.Div().Class("stub").Text("body"))
	out := app.HTMLString(layout.Draw(shell))

	if !strings.HasPrefix(out, `<div class="landing">`) {
		t.Fatalf("expected the landing shell for the home route, got %s", out)
	}
	for _, want := range []string{
		`<header class="topbar">`,
		`class="logo"`,
		`href="/start"`,
		`>Docs</a>`,
		`<a class="pill" href="` + Repository + `">GitHub</a>`,
		`class="shell-toggle"`,
		`<div class="stub">body</div>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the landing shell, got %s", want, out)
		}
	}
	if strings.Contains(out, `<aside`) {
		t.Errorf("expected no sidebar on the landing page, got %s", out)
	}
}
