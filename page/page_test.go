package page

import (
	"errors"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/model"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type appState struct {
	User string
	N    int
}

type host struct {
	app.Compo
	Tick     int
	captured *app.Context
	child    app.UI
}

func (h *host) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *host) Render() app.UI {
	if h.child == nil {
		return app.Div().Class("host")
	}
	return h.child
}

func newContext(t *testing.T) (app.Context, app.TestEngine) {
	t.Helper()
	return mountInTree(t, nil)
}

func mountInTree(t *testing.T, child app.UI) (app.Context, app.TestEngine) {
	t.Helper()

	engine := app.NewTestEngine()
	var ctx app.Context

	if err := engine.Load(&host{captured: &ctx, child: child, Tick: 0}); err != nil {
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

type recordingLayout struct {
	name         string
	loadingCalls int
	drawCalls    int
	errorCalls   int
	errs         []error
	drawnWith    []*model.ViewModel[appState]
	body         []app.UI
}

func (l *recordingLayout) Loading() app.UI {
	l.loadingCalls++
	return app.Div().Class("loading").DataSet("layout", l.name).Text("loading")
}

func (l *recordingLayout) Error(err error) app.UI {
	l.errorCalls++
	l.errs = append(l.errs, err)
	return app.Div().Class("error").DataSet("layout", l.name).Text(err.Error())
}

func (l *recordingLayout) Body(content ...app.UI) Layout[appState] {
	l.body = append(l.body, content...)
	return l
}

func (l *recordingLayout) Draw(m *model.ViewModel[appState]) app.HTMLDiv {
	l.drawCalls++
	l.drawnWith = append(l.drawnWith, m)
	body := []app.UI{app.Header().Class("chrome").Text("chrome")}
	body = append(body, l.body...)
	return app.Div().Class("shell").DataSet("layout", l.name).Body(
		body...,
	)
}

func (l *recordingLayout) lastError() error {
	if len(l.errs) == 0 {
		return nil
	}
	return l.errs[len(l.errs)-1]
}

type recordingScreen struct {
	route         string
	name          string
	routeCalls    int
	contentCalls  int
	mountCalls    int
	dismountCalls int
	mounted       bool
}

func newScreen(route, name string) *recordingScreen {
	return &recordingScreen{route: route, name: name}
}

func (s *recordingScreen) Mounted() bool { return s.mounted }

func (s *recordingScreen) OnMount(_ app.Context) Screen {
	s.mountCalls++
	s.mounted = true
	return s
}

func (s *recordingScreen) OnDismount() Screen {
	s.dismountCalls++
	s.mounted = false
	return s
}

func (s *recordingScreen) Route() string {
	s.routeCalls++
	return s.route
}

func (s *recordingScreen) Content() app.UI {
	s.contentCalls++
	return app.Div().Class("screen").Text(s.name)
}

func newLayout(name string) *recordingLayout {
	return &recordingLayout{name: name}
}

func newModel(user string) *model.ViewModel[appState] {
	viewModel, _ := model.NewViewModel("app-state", &appState{User: user})
	return viewModel
}

func newBase(t *testing.T, layout Layout[appState], m *model.ViewModel[appState], screens []Screen) *Base[appState] {
	t.Helper()

	base, err := NewBase(layout, m, screens)
	if err != nil {
		t.Fatalf("NewBase rejected the arguments: %v", err)
	}
	return base
}

func drawnBy(name string) string {
	return `data-layout="` + name + `"`
}

func html(ui app.UI) string {
	return app.HTMLString(ui)
}

func requireContains(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if !strings.Contains(markup, fragment) {
			t.Fatalf("expected %q in %s", fragment, markup)
		}
	}
}

func requireNotContains(t *testing.T, markup string, fragments ...string) {
	t.Helper()
	for _, fragment := range fragments {
		if strings.Contains(markup, fragment) {
			t.Fatalf("did not expect %q in %s", fragment, markup)
		}
	}
}

func didPanic(f func()) (panicked bool, value any) {
	defer func() {
		if r := recover(); r != nil {
			panicked = true
			value = r
		}
	}()
	f()
	return false, nil
}

func TestAPageIsAScreenOnlyAsAPointer(t *testing.T) {
	var pointer any = NewPage[appState]("/")
	if _, ok := pointer.(Screen); !ok {
		t.Fatal("expected *Page to satisfy Screen")
	}
	if _, ok := pointer.(app.Mounter); ok {
		t.Fatal("did not expect *Page to satisfy app.Mounter: a base mounts it, the engine does not")
	}
	if _, ok := pointer.(app.Navigator); ok {
		t.Fatal("did not expect *Page to satisfy app.Navigator")
	}

	var value any = Page[appState]{}
	if _, ok := value.(Screen); ok {
		t.Fatal("did not expect a Page value to satisfy Screen")
	}
}

func TestALayoutServesBothAPageAndABase(t *testing.T) {
	var layout any = newLayout("shared")
	if _, ok := layout.(Layout[appState]); !ok {
		t.Fatal("expected recordingLayout to satisfy Layout[appState]")
	}
}

func TestAPageIsBuiltByChainingALayoutAndAModel(t *testing.T) {
	layout := newLayout("page")
	vm := newModel("ada")

	page := NewPage[appState]("/").WithLayout(layout).WithModel(vm)

	if page.Err() != nil {
		t.Fatalf("expected a complete page to carry no error, got %v", page.Err())
	}

	ctx, _ := newContext(t)
	page.OnMount(ctx)

	requireContains(t, html(page.Content()), `class="shell"`, drawnBy("page"))
	if layout.drawnWith[0] != vm {
		t.Fatal("expected Draw to receive the model the page was built with")
	}
}

func TestChainingReturnsTheSamePageSoTheOrderDoesNotMatter(t *testing.T) {
	layout := newLayout("page")
	vm := newModel("ada")

	first := NewPage[appState]("/").WithLayout(layout).WithModel(vm)
	second := NewPage[appState]("/").WithModel(vm).WithLayout(layout)

	ctx, _ := newContext(t)
	first.OnMount(ctx)
	second.OnMount(ctx)

	requireContains(t, html(first.Content()), `class="shell"`)
	requireContains(t, html(second.Content()), `class="shell"`)
}

func TestAMissingLayoutIsReportedAsAnError(t *testing.T) {
	page := NewPage[appState]("/").WithLayout(nil).WithModel(newModel("ada"))

	if !errors.Is(page.Err(), ErrNoLayout) {
		t.Fatalf("expected a missing layout to be reported, got %v", page.Err())
	}
}

func TestAMissingModelIsReportedAsAnError(t *testing.T) {
	page := NewPage[appState]("/").WithLayout(newLayout("page")).WithModel(nil)

	if !errors.Is(page.Err(), ErrNoModel) {
		t.Fatalf("expected a missing model to be reported, got %v", page.Err())
	}
}

func TestTheFirstErrorInAChainWins(t *testing.T) {
	page := NewPage[appState]("/").WithLayout(nil).WithModel(nil)

	if !errors.Is(page.Err(), ErrNoLayout) {
		t.Fatalf("expected the first failure to be kept, got %v", page.Err())
	}
}

func TestAChainStopsApplyingStepsAfterAFailure(t *testing.T) {
	layout := newLayout("page")

	page := NewPage[appState]("/").WithModel(nil).WithLayout(layout)

	if !errors.Is(page.Err(), ErrNoModel) {
		t.Fatalf("expected the first failure to be kept, got %v", page.Err())
	}
	if layout.drawCalls != 0 || layout.loadingCalls != 0 {
		t.Fatal("expected a step after a failure to be skipped entirely")
	}
}

func TestAPageBuiltWithoutALayoutReportsItThroughTheError(t *testing.T) {
	page := NewPage[appState]("/").WithModel(newModel("ada"))

	if !errors.Is(page.Err(), ErrNoLayout) {
		t.Fatalf("expected a page with no layout to carry an error, got %v", page.Err())
	}
}

func TestAPageWithoutALayoutCannotRenderAnything(t *testing.T) {
	page := NewPage[appState]("/").WithModel(newModel("ada"))

	if panicked, _ := didPanic(func() { html(page.Content()) }); !panicked {
		t.Fatal("expected a page with no layout to fail loudly instead of shipping developer copy")
	}
}

func TestAPageWithoutAModelDrawsTheErrorInsteadOfTheContent(t *testing.T) {
	layout := newLayout("page")
	page := NewPage[appState]("/").WithLayout(layout)

	markup := html(page.Content())

	requireContains(t, markup, `class="error"`, drawnBy("page"))
	if !errors.Is(layout.lastError(), ErrNoModel) {
		t.Fatalf("expected the layout to be handed the missing model error, got %v", layout.lastError())
	}
	if layout.drawCalls != 0 {
		t.Fatalf("expected Draw not to run without a model, got %d calls", layout.drawCalls)
	}
}

func TestRouteAlwaysMatchesAUrlPath(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"/", "/"},
		{"", "/"},
		{"settings", "/settings"},
		{"/settings", "/settings"},
		{"a/b/c", "/a/b/c"},
		{"/a/b/c", "/a/b/c"},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			page := NewPage[appState](tc.path)
			if got := page.Route(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestAPageShowsItsLoadingLayoutUntilItIsMounted(t *testing.T) {
	layout := newLayout("page")
	page := NewPage[appState]("/").WithLayout(layout).WithModel(newModel("ada"))

	requireContains(t, html(page.Content()), `class="loading"`)
	if layout.drawCalls != 0 {
		t.Fatalf("expected Draw not to be called before mount, got %d", layout.drawCalls)
	}
}

func TestAMountedPageKeepsTheLayoutsOwnChildren(t *testing.T) {
	ctx, _ := newContext(t)
	page := NewPage[appState]("/").WithLayout(newLayout("page")).WithModel(newModel("ada"))
	page.OnMount(ctx)

	requireContains(t, html(page.Content()), `class="chrome"`)
}

func TestMountingReadsTheModelFromTheContext(t *testing.T) {
	ctx, _ := newContext(t)
	page := NewPage[appState]("/").WithLayout(newLayout("page")).WithModel(newModel("seed"))

	ctx.SetState("app-state", &appState{User: "stored", N: 9})
	page.OnMount(ctx)

	requireContains(t, html(page.Content()), `class="shell"`)
}

func TestMountingMarksThePageAsActive(t *testing.T) {
	ctx, _ := newContext(t)
	page := NewPage[appState]("/").WithLayout(newLayout("page")).WithModel(newModel("ada"))

	page.OnMount(ctx)

	if !page.Mounted() {
		t.Fatal("expected the page to be marked as active")
	}
}

func TestDismountingMarksThePageAsInactive(t *testing.T) {
	ctx, _ := newContext(t)
	page := NewPage[appState]("/").WithLayout(newLayout("page")).WithModel(newModel("ada"))

	page.OnMount(ctx)
	page.OnDismount()

	if page.Mounted() {
		t.Fatal("expected a dismounted page to stop being active")
	}
	requireContains(t, html(page.Content()), `class="loading"`)
}

func TestMountingAPageWithoutAModelIsReportedNotFatal(t *testing.T) {
	ctx, _ := newContext(t)
	layout := newLayout("page")
	page := NewPage[appState]("/").WithLayout(layout)

	if panicked, value := didPanic(func() { page.OnMount(ctx) }); panicked {
		t.Fatalf("expected a page without a model to be reported, not to panic: %v", value)
	}
	if !errors.Is(page.Err(), ErrNoModel) {
		t.Fatalf("expected the missing model to be recorded, got %v", page.Err())
	}
	requireContains(t, html(page.Content()), `class="error"`)
}

func TestAPageUsedAsAScreenRendersItsOwnContent(t *testing.T) {
	pageLayout := newLayout("page")
	shell := newLayout("shell")
	page := NewPage[appState]("/").WithLayout(pageLayout).WithModel(newModel("ada"))

	base := newBase(t, shell, newModel("ada"), []Screen{page})
	ctx, engine := mountInTree(t, base)
	base.OnMount(ctx)
	engine.ConsumeAll()

	markup := html(base.Render())

	if pageLayout.drawCalls == 0 {
		t.Fatalf("expected the selected page to render its own layout, got: %s", markup)
	}
	requireContains(t, markup, drawnBy("shell"), drawnBy("page"), `class="chrome"`)
}

func TestNewBaseKeepsWhatItIsGiven(t *testing.T) {
	shell := newLayout("shell")
	vm := newModel("ada")
	screens := []Screen{newScreen("/", "home")}

	base, err := NewBase(shell, vm, screens)

	if err != nil {
		t.Fatalf("expected a valid base to be accepted, got %v", err)
	}
	if base.BaseLayout != Layout[appState](shell) {
		t.Fatal("expected the shell layout to be kept")
	}
	if base.Model != vm {
		t.Fatal("expected the model to be kept")
	}
	if len(base.Screens) != 1 {
		t.Fatalf("expected one screen, got %d", len(base.Screens))
	}
	if base.selected != 0 {
		t.Fatalf("expected the first screen to be selected, got %d", base.selected)
	}
}

func TestNewBaseRejectsMissingDependencies(t *testing.T) {
	shell := newLayout("shell")
	vm := newModel("ada")
	screens := []Screen{newScreen("/", "home")}

	cases := []struct {
		name    string
		layout  Layout[appState]
		model   *model.ViewModel[appState]
		screens []Screen
		want    error
	}{
		{"no layout", nil, vm, screens, ErrNoLayout},
		{"no model", shell, nil, screens, ErrNoModel},
		{"no screens", shell, vm, nil, ErrNoScreens},
		{"nothing at all", nil, nil, nil, ErrNoLayout},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			base, err := NewBase(tc.layout, tc.model, tc.screens)

			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
			if base != nil {
				t.Fatalf("expected no base alongside the error, got %#v", base)
			}
		})
	}
}

func TestNewBaseAcceptsAnAllocatedButEmptyScreenList(t *testing.T) {
	base, err := NewBase(newLayout("shell"), newModel("ada"), []Screen{})

	if err != nil {
		t.Fatalf("expected an empty screen list to be accepted, got %v", err)
	}
	if base == nil {
		t.Fatal("expected a base to be returned")
	}
}

func TestABaseIsAMounterAndANavigatorButNotAScreen(t *testing.T) {
	var base any = &Base[appState]{}

	if _, ok := base.(app.Mounter); !ok {
		t.Fatal("expected *Base to satisfy app.Mounter")
	}
	if _, ok := base.(app.Navigator); !ok {
		t.Fatal("expected *Base to satisfy app.Navigator")
	}
	if _, ok := base.(Screen); ok {
		t.Fatal("did not expect *Base to satisfy Screen")
	}
}

func TestABaseShowsItsLoadingStateBeforeMount(t *testing.T) {
	shell := newLayout("shell")
	screen := newScreen("/", "home")
	base := newBase(t, shell, newModel("ada"), []Screen{screen})

	markup := html(base.Render())

	requireContains(t, markup, `class="loading"`, drawnBy("shell"))
	if shell.loadingCalls != 1 {
		t.Fatalf("expected one loading render, got %d", shell.loadingCalls)
	}
	if screen.contentCalls != 0 {
		t.Fatalf("expected the screen not to be asked for content before mount, got %d calls", screen.contentCalls)
	}
}

func TestAMountedBaseDrawsTheSelectedScreenInsideTheShell(t *testing.T) {
	ctx, _ := newContext(t)
	shell := newLayout("shell")
	base := newBase(t, shell, newModel("ada"), []Screen{
		newScreen("/", "home"),
		newScreen("/settings", "settings"),
	})
	base.OnMount(ctx)
	base.selected = 1

	markup := html(base.Render())

	requireContains(t, markup, `class="shell"`, drawnBy("shell"), `class="screen"`, "settings", `class="chrome"`)
	requireNotContains(t, markup, "home")
	if shell.drawCalls != 1 {
		t.Fatalf("expected one Draw call, got %d", shell.drawCalls)
	}
	if shell.drawnWith[0] != base.Model {
		t.Fatal("expected Draw to receive the base's own model")
	}
}

func TestAMountedBaseWithNoScreensStaysOnItsLoadingState(t *testing.T) {
	ctx, _ := newContext(t)
	shell := newLayout("shell")
	base := newBase(t, shell, newModel("ada"), []Screen{})
	base.OnMount(ctx)

	markup := html(base.Render())

	requireContains(t, markup, `class="loading"`, drawnBy("shell"))
	if shell.drawCalls != 0 {
		t.Fatalf("expected Draw not to be reached, got %d calls", shell.drawCalls)
	}
}

func TestAnUnmatchedRouteIsDrawnAsAnErrorByTheOneLayout(t *testing.T) {
	ctx, _ := newContext(t)
	shell := newLayout("shell")
	base := newBase(t, shell, newModel("ada"), []Screen{newScreen("/a", "alpha")})
	base.OnMount(ctx)
	base.selected = -1

	markup := html(base.Render())

	requireContains(t, markup, `class="error"`, drawnBy("shell"))
	requireNotContains(t, markup, "alpha")
	if !errors.Is(shell.lastError(), ErrNoRoute) {
		t.Fatalf("expected the layout to be handed the unmatched route error, got %v", shell.lastError())
	}
}

func TestABaseClampsAnOutOfRangeSelection(t *testing.T) {
	ctx, _ := newContext(t)
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{newScreen("/", "home")})
	base.OnMount(ctx)
	base.selected = 7

	markup := html(base.Render())

	requireContains(t, markup, "home")
	if base.selected != 0 {
		t.Fatalf("expected Render to reset the selection to the first screen, got %d", base.selected)
	}
}

func TestABaseAssembledWithoutALayoutCannotRender(t *testing.T) {
	base := &Base[appState]{Screens: []Screen{newScreen("/", "home")}}

	if panicked, _ := didPanic(func() { base.Render() }); !panicked {
		t.Fatal("expected a base assembled by hand without a layout to fail loudly")
	}
}

func TestABaseAssembledWithoutAModelDrawsTheError(t *testing.T) {
	ctx, _ := newContext(t)
	shell := newLayout("shell")
	base := &Base[appState]{BaseLayout: shell, Screens: []Screen{newScreen("/", "home")}}

	if panicked, value := didPanic(func() { base.OnMount(ctx) }); panicked {
		t.Fatalf("expected a base without a model to be reported, not to panic: %v", value)
	}

	requireContains(t, html(base.Render()), `class="error"`, drawnBy("shell"))
	if !errors.Is(shell.lastError(), ErrNoModel) {
		t.Fatalf("expected the layout to be handed the missing model error, got %v", shell.lastError())
	}
}

func TestMountingABaseLoadsTheModelAndMountsTheSelectedScreen(t *testing.T) {
	ctx, _ := newContext(t)
	ctx.SetState("app-state", &appState{User: "stored", N: 3})

	shell := newLayout("shell")
	home := newScreen("/", "home")
	base := newBase(t, shell, newModel("seed"), []Screen{home})

	base.OnMount(ctx)

	requireContains(t, html(base.Render()), `class="shell"`)
	if shell.loadingCalls != 0 {
		t.Fatalf("expected no loading render after mount, got %d", shell.loadingCalls)
	}
	if home.mountCalls != 1 {
		t.Fatalf("expected the selected screen to be mounted once, got %d", home.mountCalls)
	}
}

func TestMountingABaseTwiceIsHarmless(t *testing.T) {
	ctx, _ := newContext(t)
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{newScreen("/", "home")})

	base.OnMount(ctx)
	base.OnMount(ctx)

	requireContains(t, html(base.Render()), `class="shell"`)
}

func TestNavigatingSelectsTheScreenWhoseRouteMatchesThePath(t *testing.T) {
	settings := newScreen("/settings", "settings")
	home := newScreen("/", "home")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{settings, home})
	ctx, engine := mountInTree(t, base)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if base.selected != 1 {
		t.Fatalf("expected the matching screen to be selected, got %d", base.selected)
	}
	if home.mountCalls != 1 {
		t.Fatalf("expected the newly selected screen to be mounted, got %d", home.mountCalls)
	}
}

func TestNavigatingDismountsTheScreenItLeaves(t *testing.T) {
	settings := newScreen("/settings", "settings")
	home := newScreen("/", "home")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{settings, home})
	ctx, engine := mountInTree(t, base)
	base.OnMount(ctx)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if settings.dismountCalls != 1 {
		t.Fatalf("expected the screen being left to be dismounted once, got %d", settings.dismountCalls)
	}
	if home.dismountCalls != 0 {
		t.Fatalf("expected the screen being entered not to be dismounted, got %d", home.dismountCalls)
	}
}

func TestNavigatingKeepsTheUrlItWasGiven(t *testing.T) {
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{
		newScreen("/settings", "settings"),
		newScreen("/", "home"),
	})
	ctx, engine := mountInTree(t, base)
	base.OnMount(ctx)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if !base.mounted {
		t.Fatal("expected the shell to survive selecting a screen")
	}
	if got := ctx.Page().URL().Path; got != "/" {
		t.Fatalf("expected the URL to still be the one that was navigated to, got %q", got)
	}
}

func TestNavigatingNowhereClearsTheSelection(t *testing.T) {
	shell := newLayout("shell")
	base := newBase(t, shell, newModel("ada"), []Screen{
		newScreen("/a", "a"),
		newScreen("/b", "b"),
	})
	ctx, engine := mountInTree(t, base)
	base.OnMount(ctx)
	base.selected = 1

	base.OnNav(ctx)
	engine.ConsumeAll()

	if base.selected != -1 {
		t.Fatalf("expected an unmatched path to clear the selection, got %d", base.selected)
	}
	if !base.mounted {
		t.Fatal("expected the tree to survive a navigation that matched nothing")
	}
	requireContains(t, html(base.Render()), `class="error"`, drawnBy("shell"))
}

func TestNavigatingKeepsTheScreenItIsAlreadyShowing(t *testing.T) {
	home := newScreen("/", "home")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{home})
	ctx, engine := mountInTree(t, base)
	base.OnMount(ctx)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if base.selected != 0 {
		t.Fatalf("expected the screen already on display to stay selected, got %d", base.selected)
	}
	if home.dismountCalls != 0 {
		t.Fatalf("expected the screen on display not to be torn down, got %d dismounts", home.dismountCalls)
	}
	requireContains(t, html(base.Render()), "home")
}

func TestNavigatingReadsEachRouteOnce(t *testing.T) {
	miss := newScreen("/elsewhere", "elsewhere")
	match := newScreen("/", "home")
	base := newBase(t, newLayout("shell"), newModel("ada"), []Screen{miss, match})
	ctx, engine := mountInTree(t, base)

	base.OnNav(ctx)
	engine.ConsumeAll()

	if miss.routeCalls != 1 {
		t.Fatalf("expected a non-matching screen's Route to be read once, got %d", miss.routeCalls)
	}
	if match.routeCalls != 1 {
		t.Fatalf("expected the matching screen's Route to be read once, got %d", match.routeCalls)
	}
}

func TestNavigatingBeforeMountDoesNotMount(t *testing.T) {
	shell := newLayout("shell")
	base := newBase(t, shell, newModel("ada"), []Screen{newScreen("/nothing", "nothing")})
	ctx, engine := mountInTree(t, base)

	base.OnNav(ctx)
	engine.ConsumeAll()

	requireContains(t, html(base.Render()), `class="loading"`, drawnBy("shell"))
}

func TestMustParseUrlAcceptsRoutes(t *testing.T) {
	cases := []struct {
		name string
		path string
		want string
	}{
		{"root", "/", "/"},
		{"nested", "/settings/profile", "/settings/profile"},
		{"empty", "", ""},
		{"with query", "/search?q=go", "/search?q=go"},
		{"with fragment", "/docs#intro", "/docs#intro"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := mustParseUrl(tc.path).String(); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestMustParseUrlPanicsOnMalformedInput(t *testing.T) {
	for _, path := range []string{"://missing-scheme", "http://[::1", "%zz"} {
		if panicked, _ := didPanic(func() { mustParseUrl(path) }); !panicked {
			t.Fatalf("expected %q to panic", path)
		}
	}
}
