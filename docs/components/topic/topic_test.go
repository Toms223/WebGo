package topic

import (
	"errors"
	"net/url"
	"slices"
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
	layout := NewLayout("Markdown", nil)
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

const sectioned = "Lead paragraph.\n\n## Overview\n\nOverview body.\n\n## Values\n\nValues body.\n"

func mountSectioned(t *testing.T, raw string) (*Screen, app.Context, app.TestEngine) {
	t.Helper()
	ctx, engine := newContext(t)
	if raw != "" {
		ctx.Page().ReplaceURL(mustParse(t, raw))
	}
	screen, err := New(Config{
		Key:   "styling",
		Path:  "/styling",
		Title: "Styling",
		URL:   "/data/styling.md",
	})
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	screen.loader = func(string) (string, error) { return sectioned, nil }
	screen.OnMount(ctx)
	engine.ConsumeAll()
	return screen, ctx, engine
}

func TestCommitSplitsSections(t *testing.T) {
	screen, _, _ := mountSectioned(t, "/styling")

	state := screen.model.State()
	if state.Intro != "Lead paragraph.\n\n" {
		t.Errorf("expected the intro before the first section, got %q", state.Intro)
	}
	slugs := make([]string, 0, len(state.Sections))
	for _, section := range state.Sections {
		slugs = append(slugs, section.Slug)
	}
	if want := []string{"overview", "values"}; !slices.Equal(slugs, want) {
		t.Errorf("section slugs = %v, want %v", slugs, want)
	}
	if state.Active != "overview" {
		t.Errorf("expected the first section to be active, got %q", state.Active)
	}
}

func TestFragmentSelectsTheSection(t *testing.T) {
	screen, _, _ := mountSectioned(t, "/styling#values")

	if got := screen.model.State().Active; got != "values" {
		t.Errorf("expected the fragment to select values, got %q", got)
	}
}

func TestUnknownFragmentFallsBackToTheFirstSection(t *testing.T) {
	screen, _, _ := mountSectioned(t, "/styling#missing")

	if got := screen.model.State().Active; got != "overview" {
		t.Errorf("expected an unknown fragment to fall back to overview, got %q", got)
	}
}

func TestDrawRendersOnlyTheActiveSection(t *testing.T) {
	screen, _, _ := mountSectioned(t, "/styling")

	out := app.HTMLString(screen.Content())
	opening := out[:strings.Index(out, ">")+1]
	for _, want := range []string{`class="topic"`, `id="topic"`} {
		if !strings.HasPrefix(opening, "<div ") || !strings.Contains(opening, want) {
			t.Errorf("expected the topic container to carry %s, got %s", want, opening)
		}
	}
	for _, want := range []string{
		`<h2 class="topic-title">Styling</h2>`,
		"Lead paragraph.",
		`<nav class="sections">`,
		`href="#overview"`,
		`href="#values"`,
		"Overview body.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the topic, got %s", want, out)
		}
	}
	if strings.Contains(out, "Values body.") {
		t.Errorf("expected only the active section to render, got %s", out)
	}
	if count := strings.Count(out, "section-link-active"); count != 1 {
		t.Errorf("expected exactly one active section link, found %d in %s", count, out)
	}
	active := linkFor(out, "Overview")
	if !strings.Contains(active, `class="section-link section-link-active"`) {
		t.Errorf("expected the overview link to be active, got %s", active)
	}
}

func TestDrawWithoutSectionsRendersTheWholeSource(t *testing.T) {
	ctx, engine := newContext(t)
	screen := newScreen(t, func(string) (string, error) { return "Plain text.\n\n### Deeper\n", nil })

	screen.OnMount(ctx)
	engine.ConsumeAll()

	out := app.HTMLString(screen.Content())
	if strings.Contains(out, `class="sections"`) {
		t.Errorf("expected no section switch without sections, got %s", out)
	}
	for _, want := range []string{"Plain text.", "Deeper"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the topic, got %s", want, out)
		}
	}
}

func TestActivateUpdatesActive(t *testing.T) {
	screen, ctx, engine := mountSectioned(t, "/styling")

	screen.Activate(ctx, "values")
	engine.ConsumeAll()

	if got := screen.model.State().Active; got != "values" {
		t.Fatalf("expected Activate to select values, got %q", got)
	}
	out := app.HTMLString(screen.Content())
	if !strings.Contains(out, "Values body.") || strings.Contains(out, "Overview body.") {
		t.Errorf("expected only the values section to render, got %s", out)
	}
}

func TestActivateNavigatesToTheFragment(t *testing.T) {
	cases := []struct {
		browser string
		want    string
	}{
		{"/styling", "/styling#values"},
		{"/styling#overview", "/styling#values"},
		{"/WebGo/styling", "/WebGo/styling#values"},
	}

	for _, test := range cases {
		t.Run(test.browser, func(t *testing.T) {
			screen, ctx, engine := mountSectioned(t, test.browser)
			var visited []string
			screen.navigate = func(ctx app.Context, destination *url.URL) {
				visited = append(visited, destination.String())
			}

			screen.Activate(ctx, "values")
			engine.ConsumeAll()

			if !slices.Equal(visited, []string{test.want}) {
				t.Errorf("expected one navigation to %s, got %v", test.want, visited)
			}
			if got := ctx.Page().URL().String(); got != test.browser {
				t.Errorf("expected the current URL to be left for the navigator, got %s", got)
			}
		})
	}
}

func TestActivateIgnoresAnUnknownSection(t *testing.T) {
	screen, ctx, engine := mountSectioned(t, "/styling")

	navigated := false
	screen.navigate = func(app.Context, *url.URL) { navigated = true }

	screen.Activate(ctx, "missing")
	engine.ConsumeAll()

	if navigated {
		t.Error("expected an unknown section not to navigate")
	}

	if got := screen.model.State().Active; got != "overview" {
		t.Errorf("expected an unknown section to leave overview active, got %q", got)
	}
}

func linkFor(html, text string) string {
	marker := ">" + text + "</a>"
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

func TestOnNavFollowsTheFragment(t *testing.T) {
	cases := []struct {
		name     string
		fragment string
		want     string
	}{
		{"a known section is selected", "/styling#values", "values"},
		{"an unknown section leaves the active one", "/styling#missing", "overview"},
		{"no fragment leaves the active one", "/styling", "overview"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			screen, ctx, engine := mountSectioned(t, "/styling")
			navigated := false
			screen.navigate = func(app.Context, *url.URL) { navigated = true }

			ctx.Page().ReplaceURL(mustParse(t, test.fragment))
			screen.OnNav(ctx)
			engine.ConsumeAll()

			if got := screen.model.State().Active; got != test.want {
				t.Errorf("expected %q to be active, got %q", test.want, got)
			}
			if navigated {
				t.Error("expected OnNav to follow the URL without navigating")
			}
		})
	}
}

func TestDrawRendersTheRegisteredDemos(t *testing.T) {
	ctx, engine := newContext(t)
	screen, err := New(Config{
		Key:   "demos",
		Path:  "/demos",
		Title: "Demos",
		URL:   "/data/demos.md",
		Demos: map[string]func() app.UI{
			"box": func() app.UI { return app.Span().Class("demo-box") },
		},
	})
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	screen.loader = func(string) (string, error) {
		return "Lead.\n\n## Box\n\n```go box\nvar Box = 1\n```\n", nil
	}
	screen.OnMount(ctx)
	engine.ConsumeAll()

	out := app.HTMLString(screen.Content())
	for _, want := range []string{`class="example"`, `class="demo-box"`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the topic, got %s", want, out)
		}
	}
}
