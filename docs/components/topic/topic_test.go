package topic

import (
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/docs/components/doc"
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

func text(s string) func() []app.UI {
	return func() []app.UI { return []app.UI{app.P().Text(s)} }
}

var sectioned = Config{
	Key:   "styling",
	Path:  "/styling",
	Title: "Styling",
	Lead:  "Typed CSS.",
	Intro: text("Lead paragraph."),
	Sections: []Section{
		{Title: "Overview", Body: text("Overview body.")},
		{Title: "Values", Body: text("Values body.")},
		{Title: "CSS()", Body: text("CSS body.")},
	},
}

func mustParse(t *testing.T, raw string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parsing %q failed: %v", raw, err)
	}
	return parsed
}

func mount(t *testing.T, cfg Config, raw string) (*Screen, app.Context, app.TestEngine) {
	t.Helper()
	ctx, engine := newContext(t)
	if raw != "" {
		ctx.Page().ReplaceURL(mustParse(t, raw))
	}
	screen, err := New(cfg)
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	screen.OnMount(ctx)
	engine.ConsumeAll()
	return screen, ctx, engine
}

func TestNewRejectsAnEmptyKey(t *testing.T) {
	if _, err := New(Config{Path: "/markdown", Title: "Markdown"}); err == nil {
		t.Error("expected an empty key to be rejected")
	}
}

func TestRouteIsDerivedFromPath(t *testing.T) {
	screen, err := New(Config{Key: "markdown", Path: "/markdown", Title: "Markdown"})
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	if got := screen.Route(); got != "/markdown" {
		t.Errorf("expected /markdown, got %s", got)
	}
}

func TestContentIsLoadingBeforeMount(t *testing.T) {
	screen, err := New(Config{Key: "markdown", Path: "/markdown", Title: "Markdown"})
	if err != nil {
		t.Fatalf("building the screen failed: %v", err)
	}
	if out := app.HTMLString(screen.Content()); !strings.Contains(out, "topic-loading") {
		t.Errorf("expected the loading state before mount, got %s", out)
	}
}

func TestOnActivateReceivesTheRoute(t *testing.T) {
	var seen string
	cfg := sectioned
	cfg.OnActivate = func(ctx app.Context, route string) { seen = route }
	mount(t, cfg, "")
	if seen != "/styling" {
		t.Errorf("expected the activation hook to receive /styling, got %q", seen)
	}
}

func TestLayoutBodyReplacesRatherThanAppends(t *testing.T) {
	layout := NewLayout(sectioned)
	layout.Body(app.Div().Class("first"))
	layout.Body(app.Div().Class("second"))
	if len(layout.body) != 1 {
		t.Fatalf("expected Body to replace its content, found %d children", len(layout.body))
	}
}

var _ page.Screen = (*Screen)(nil)

func TestSectionsGetUniqueSlugs(t *testing.T) {
	screen, _, _ := mount(t, sectioned, "/styling")
	if want := []string{"overview", "values", "css"}; !slices.Equal(screen.Sections(), want) {
		t.Errorf("sections = %v, want %v", screen.Sections(), want)
	}
	layout := NewLayout(Config{Sections: []Section{{Title: "CSS"}, {Title: "CSS"}, {Title: "???"}}})
	var slugs []string
	for _, s := range layout.sections {
		slugs = append(slugs, s.Slug)
	}
	if want := []string{"css", "css-2", "section-2"}; !slices.Equal(slugs, want) {
		t.Errorf("slugs = %v, want %v", slugs, want)
	}
}

func TestMountSelectsTheFirstSection(t *testing.T) {
	screen, _, _ := mount(t, sectioned, "/styling")
	if got := screen.model.State().Active; got != "overview" {
		t.Errorf("expected the first section to be active, got %q", got)
	}
}

func TestFragmentSelectsTheSection(t *testing.T) {
	screen, _, _ := mount(t, sectioned, "/styling#values")
	if got := screen.model.State().Active; got != "values" {
		t.Errorf("expected the fragment to select values, got %q", got)
	}
}

func TestUnknownFragmentFallsBackToTheFirstSection(t *testing.T) {
	screen, _, _ := mount(t, sectioned, "/styling#missing")
	if got := screen.model.State().Active; got != "overview" {
		t.Errorf("expected an unknown fragment to fall back to overview, got %q", got)
	}
}

func TestDrawRendersOnlyTheActiveSection(t *testing.T) {
	screen, _, _ := mount(t, sectioned, "/styling")
	out := app.HTMLString(screen.Content())
	for _, want := range []string{
		`class="topic"`, `id="topic"`,
		`<h1 class="topic-title">Styling</h1>`,
		`class="topic-lead">Typed CSS.</p>`,
		"Lead paragraph.",
		`class="rail"`, `class="rail-count">1 of 3`,
		`class="section-heading"`, `>Overview</h2>`,
		"Overview body.",
		`class="pager"`, "Next", "Values",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the topic, got %s", want, out)
		}
	}
	for _, unwanted := range []string{"Values body.", "Previous", "turn-b"} {
		if strings.Contains(out, unwanted) {
			t.Errorf("did not expect %q in the first section, got %s", unwanted, out)
		}
	}
}

func TestDrawWithoutSectionsRendersTheIntroOnly(t *testing.T) {
	screen, _, _ := mount(t, Config{Key: "start", Path: "/start", Title: "Start", Intro: text("Plain text.")}, "")
	out := app.HTMLString(screen.Content())
	if strings.Contains(out, `class="rail"`) || strings.Contains(out, `class="pager"`) {
		t.Errorf("expected no rail or pager without sections, got %s", out)
	}
	if !strings.Contains(out, "Plain text.") {
		t.Errorf("expected the intro in the topic, got %s", out)
	}
}

func TestActivateUpdatesActiveAndAlternatesTheTurn(t *testing.T) {
	screen, ctx, engine := mount(t, sectioned, "/styling")
	screen.navigate = func(app.Context, *url.URL) {}

	screen.Activate(ctx, "values")
	engine.ConsumeAll()

	if got := screen.model.State().Active; got != "values" {
		t.Fatalf("expected Activate to select values, got %q", got)
	}
	out := app.HTMLString(screen.Content())
	for _, want := range []string{"Values body.", "turn-b", "Previous", "Overview", "Next", "CSS()"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the values section, got %s", want, out)
		}
	}
	if strings.Contains(out, "Overview body.") {
		t.Errorf("expected only the values section to render, got %s", out)
	}

	screen.Activate(ctx, "css")
	engine.ConsumeAll()
	if out := app.HTMLString(screen.Content()); strings.Contains(out, "turn-b") || strings.Contains(out, "Next") {
		t.Errorf("expected the third switch to drop turn-b and the last section to have no Next, got %s", out)
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
			screen, ctx, engine := mount(t, sectioned, test.browser)
			var visited []string
			screen.navigate = func(ctx app.Context, destination *url.URL) {
				visited = append(visited, destination.String())
			}

			screen.Activate(ctx, "values")
			engine.ConsumeAll()

			if !slices.Equal(visited, []string{test.want}) {
				t.Errorf("expected one navigation to %s, got %v", test.want, visited)
			}
		})
	}
}

func TestActivateIgnoresAnUnknownSection(t *testing.T) {
	screen, ctx, engine := mount(t, sectioned, "/styling")
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
			screen, ctx, engine := mount(t, sectioned, "/styling")
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

func wideRail(query string, open bool) *Rail {
	layout := NewLayout(Config{Sections: []Section{
		{Title: "Overview", Body: func() []app.UI { return []app.UI{doc.H("Block order"), app.P().Text("x"), doc.H("CSS")} }},
		{Title: "Values", Body: text("v")},
		{Title: "CSS()", Body: text("c")},
	}})
	return &Rail{Sections: layout.sections, Active: 0, wide: true, open: open, query: query}
}

func TestWideRailListsSectionsAndTheActiveHeadings(t *testing.T) {
	out := app.HTMLString(wideRail("", false))
	for _, want := range []string{
		`aria-label="Sections"`, `type="search"`, "In this section",
		`href="#overview"`, `href="#values"`, `href="#css"`,
		`href="#block-order"`, `href="#css-2"`, "aria-current",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the wide rail, got %s", want, out)
		}
	}
	if strings.Contains(out, "rail-summary") {
		t.Errorf("expected no collapse button on a wide screen, got %s", out)
	}
	if count := strings.Count(out, "rail-link-active"); count != 1 {
		t.Errorf("expected exactly one active rail link, found %d in %s", count, out)
	}
}

func TestNarrowRailCollapsesUntilOpened(t *testing.T) {
	closed := wideRail("", false)
	closed.wide = false
	out := app.HTMLString(closed)
	for _, want := range []string{"rail-summary", `aria-expanded="false"`, "1 of 3"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the closed rail, got %s", want, out)
		}
	}
	if strings.Contains(out, `type="search"`) {
		t.Errorf("expected the closed rail to hide its panel, got %s", out)
	}
	opened := wideRail("", true)
	opened.wide = false
	out = app.HTMLString(opened)
	for _, want := range []string{"aria-expanded ", `id="rail-panel"`, `type="search"`, `href="#values"`, "rail-open"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the open rail, got %s", want, out)
		}
	}
}

func TestRailSearchFindsSectionsAndHeadings(t *testing.T) {
	cases := []struct {
		query string
		want  []string
		not   []string
	}{
		{"val", []string{`href="#values"`}, []string{`href="#overview"`}},
		{"block", []string{`href="#block-order"`, "Overview"}, []string{`href="#values"`}},
		{"  CSS ", []string{`href="#css"`, `href="#css-2"`}, nil},
		{"nothing", []string{"No section or heading matches"}, []string{"<a "}},
	}
	for _, test := range cases {
		out := app.HTMLString(wideRail(test.query, false))
		for _, want := range test.want {
			if !strings.Contains(out, want) {
				t.Errorf("query %q: expected %q in %s", test.query, want, out)
			}
		}
		for _, not := range test.not {
			if strings.Contains(out, not) {
				t.Errorf("query %q: did not expect %q in %s", test.query, not, out)
			}
		}
	}
}

func TestRailSearchIsCapped(t *testing.T) {
	var sections []Section
	for i := 0; i < 20; i++ {
		sections = append(sections, Section{Title: "Match " + Slugify(string(rune('a'+i)))})
	}
	rail := &Rail{Sections: slugged(sections), wide: true, query: "match"}
	if got := len(rail.matches()); got != maxResults {
		t.Errorf("expected %d results at most, got %d", maxResults, got)
	}
}

func TestHeadingsGetAnchorsThatSelectTheirSection(t *testing.T) {
	cfg := Config{Key: "a", Path: "/a", Title: "A", Sections: []Section{
		{Title: "Set", Body: func() []app.UI { return []app.UI{doc.H("Set"), doc.H("Get")} }},
		{Title: "Other", Body: func() []app.UI { return []app.UI{doc.H("Get"), doc.H("Set")} }},
	}}
	l := NewLayout(cfg)
	ids := [][]string{}
	for _, s := range l.sections {
		var row []string
		for _, h := range s.Subs {
			row = append(row, h.ID)
		}
		ids = append(ids, row)
	}
	if want := [][]string{{"set-2", "get"}, {"get-2", "set-3"}}; !slices.EqualFunc(ids, want, slices.Equal[[]string]) {
		t.Fatalf("heading ids = %v, want %v", ids, want)
	}
	for fragment, want := range map[string]int{"set": 0, "other": 1, "get": 0, "set-3": 1, "nope": 0} {
		at, known := l.find(fragment)
		if at != want || known != (fragment != "nope") {
			t.Errorf("find(%q) = %d, %v; want section %d", fragment, at, known, want)
		}
	}
	out := app.HTMLString(l.section(1, 0))
	for _, want := range []string{`id="get-2"`, `id="set-3"`, `tabindex="-1"`} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the second section, got %s", want, out)
		}
	}
}

func TestFragmentOfAHeadingSelectsItsSection(t *testing.T) {
	cfg := sectioned
	cfg.Sections = []Section{
		{Title: "Overview", Body: text("o")},
		{Title: "Values", Body: func() []app.UI { return []app.UI{doc.H("Colour"), app.P().Text("v")} }},
	}
	screen, _, _ := mount(t, cfg, "/styling#colour")
	if got := screen.model.State().Active; got != "values" {
		t.Errorf("expected a heading fragment to open its section, got %q", got)
	}
}

func TestPageTitleNamesTheSection(t *testing.T) {
	screen, ctx, engine := mount(t, sectioned, "/styling#values")
	if got := ctx.Page().Title(); got != "Values · Styling · WebGo" {
		t.Errorf("title = %q", got)
	}
	screen.navigate = func(app.Context, *url.URL) {}
	screen.Activate(ctx, "css")
	engine.ConsumeAll()
	if got := ctx.Page().Title(); got != "CSS() · Styling · WebGo" {
		t.Errorf("title after switching = %q", got)
	}
}
