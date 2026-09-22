package topic

import (
	"errors"
	"net/url"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type host struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (h *host) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *host) Render() app.UI { return app.Div().Class("host") }

func newContext(t *testing.T) (app.Context, app.TestEngine) {
	t.Helper()

	engine := app.NewTestEngine()
	var ctx app.Context

	if err := engine.Load(&host{captured: &ctx, Tick: 0}); err != nil {
		t.Fatalf("mounting the host component failed: %v", err)
	}
	if err := engine.Load(&host{Tick: 1}); err != nil {
		t.Fatalf("updating the host component failed: %v", err)
	}
	engine.ConsumeAll()

	if ctx.Src() == nil {
		t.Fatal("the host was never updated, so no Context was captured")
	}
	return ctx, engine
}

func newScreen(t *testing.T, loader func(string) (string, error)) *Screen {
	t.Helper()
	screen, err := New(Config{
		Key:   "markdown",
		Path:  "/markdown",
		Title: "Markdown",
		URL:   "/data/markdown.md",
	})
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	screen.loader = loader
	return screen
}

func TestNewRejectsAnEmptyKey(t *testing.T) {
	if _, err := New(Config{Path: "/markdown", Title: "Markdown", URL: "/data/markdown.md"}); err == nil {
		t.Error("expected an empty key to be rejected")
	}
}

func TestRouteIsDerivedFromPath(t *testing.T) {
	screen := newScreen(t, func(string) (string, error) { return "", nil })
	if got := screen.Route(); got != "/markdown" {
		t.Errorf("expected /markdown, got %s", got)
	}
}

func TestMountLoadsTheSource(t *testing.T) {
	ctx, engine := newContext(t)
	screen := newScreen(t, func(url string) (string, error) {
		if url != "/data/markdown.md" {
			t.Errorf("unexpected url %s", url)
		}
		return "# Markdown\n", nil
	})

	screen.OnMount(ctx)
	engine.ConsumeAll()

	out := app.HTMLString(screen.Content())
	if !strings.Contains(out, "Markdown") {
		t.Errorf("expected the fetched source to be rendered, got %s", out)
	}
	if strings.Contains(out, "topic-loading") {
		t.Errorf("expected the loading state to be gone, got %s", out)
	}
}

func TestMountSurfacesAFetchFailure(t *testing.T) {
	ctx, engine := newContext(t)
	screen := newScreen(t, func(string) (string, error) {
		return "", errors.New("404 Not Found")
	})

	screen.OnMount(ctx)
	engine.ConsumeAll()

	out := app.HTMLString(screen.Content())
	if !strings.Contains(out, "topic-error") {
		t.Errorf("expected a failed fetch to render the error state, got %s", out)
	}
	if !strings.Contains(out, "404 Not Found") {
		t.Errorf("expected the failure to be described, got %s", out)
	}
	if strings.Contains(out, "topic-loading") {
		t.Errorf("a failed fetch must not leave the page loading, got %s", out)
	}
}

func TestContentIsLoadingBeforeMount(t *testing.T) {
	screen := newScreen(t, func(string) (string, error) { return "", nil })
	if out := app.HTMLString(screen.Content()); !strings.Contains(out, "topic-loading") {
		t.Errorf("expected the loading state before mount, got %s", out)
	}
}

func TestOnActivateReceivesTheRoute(t *testing.T) {
	ctx, engine := newContext(t)
	var seen string
	screen, err := New(Config{
		Key:        "markdown",
		Path:       "/markdown",
		Title:      "Markdown",
		URL:        "/data/markdown.md",
		OnActivate: func(ctx app.Context, route string) { seen = route },
	})
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	screen.loader = func(string) (string, error) { return "", nil }

	screen.OnMount(ctx)
	engine.ConsumeAll()

	if seen != "/markdown" {
		t.Errorf("expected the activation hook to receive /markdown, got %q", seen)
	}
}

func TestLayoutBodyReplacesRatherThanAppends(t *testing.T) {
	layout := NewLayout("Markdown")
	layout.Body(app.Div().Class("first"))
	layout.Body(app.Div().Class("second"))

	if len(layout.body) != 1 {
		t.Fatalf("expected Body to replace its content, found %d children", len(layout.body))
	}
}

var _ page.Screen = (*Screen)(nil)

func TestLoadDerivesTheMountFromTheCurrentURL(t *testing.T) {
	ctx, engine := newContext(t)
	ctx.Page().ReplaceURL(mustParse(t, "/WebGo/markdown"))

	var requested string
	screen := newScreen(t, func(url string) (string, error) {
		requested = url
		return "# Markdown\n", nil
	})

	screen.OnMount(ctx)
	engine.ConsumeAll()

	if requested != "/WebGo/data/markdown.md" {
		t.Errorf("expected the mount to be applied, got %q", requested)
	}
}

func TestLoadLeavesTheURLAloneAtTheRoot(t *testing.T) {
	ctx, engine := newContext(t)
	ctx.Page().ReplaceURL(mustParse(t, "/markdown"))

	var requested string
	screen := newScreen(t, func(url string) (string, error) {
		requested = url
		return "# Markdown\n", nil
	})

	screen.OnMount(ctx)
	engine.ConsumeAll()

	if requested != "/data/markdown.md" {
		t.Errorf("expected the url to be unchanged at the root, got %q", requested)
	}
}

func TestMountedURL(t *testing.T) {
	cases := []struct {
		browser string
		route   string
		want    string
	}{
		{"/markdown", "/markdown", "/data/markdown.md"},
		{"/", "/", "/data/markdown.md"},
		{"/WebGo/markdown", "/markdown", "/WebGo/data/markdown.md"},
		{"/WebGo", "/", "/WebGo/data/markdown.md"},
		{"/a/b/markdown", "/markdown", "/a/b/data/markdown.md"},
	}

	for _, test := range cases {
		got := mountedURL(test.browser, test.route, "/data/markdown.md")
		if got != test.want {
			t.Errorf("mountedURL(%q, %q): expected %q, got %q", test.browser, test.route, test.want, got)
		}
	}
}

func mustParse(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parsing %q failed: %v", raw, err)
	}
	return parsed
}
