package app

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

type shellLayout struct {
	name         string
	loadingCalls int
	drawCalls    int
	errs         []error
	drawnWith    []*Model[screenState]
	body         []UI
}

func (l *shellLayout) Loading() UI {
	l.loadingCalls++
	return Div().Class("loading").DataSet("layout", l.name).Text("loading")
}

func (l *shellLayout) Error(err error) UI {
	l.errs = append(l.errs, err)
	return Div().Class("error").DataSet("layout", l.name).Text(err.Error())
}

func (l *shellLayout) Body(content ...UI) Layout[screenState] {
	l.body = content
	return l
}

func (l *shellLayout) Draw(m *Model[screenState]) HTMLDiv {
	l.drawCalls++
	l.drawnWith = append(l.drawnWith, m)
	body := append([]UI{Header().Class("chrome").Text("chrome")}, l.body...)
	return Div().Class("shell").DataSet("layout", l.name).Body(body...)
}

func (l *shellLayout) lastError() error {
	if len(l.errs) == 0 {
		return nil
	}
	return l.errs[len(l.errs)-1]
}

type routedStub struct {
	Compo
	route         string
	name          string
	routeCalls    int
	mountCalls    int
	dismountCalls int
}

func newRoutedStub(route, name string) *routedStub {
	return &routedStub{route: route, name: name}
}

func (s *routedStub) Route() string {
	s.routeCalls++
	return s.route
}

func (s *routedStub) CompoID() string { return s.route }
func (s *routedStub) OnMount(Context) { s.mountCalls++ }
func (s *routedStub) OnDismount()     { s.dismountCalls++ }
func (s *routedStub) Render() UI      { return Div().Class("screen").Text(s.name) }

func newTestShell(t *testing.T, layout *shellLayout, screens ...Routed) *Shell[screenState] {
	t.Helper()
	t.Setenv("GOAPP_ROOT_PREFIX", "")
	if screens == nil {
		screens = []Routed{}
	}
	shell, err := NewShell[screenState](layout, newScreenModel("ada"), screens...)
	require.NoError(t, err)
	return shell
}

func loadShell(t *testing.T, shell *Shell[screenState]) *engineX {
	t.Helper()
	asClient(t)
	e := newTestEngine()
	require.NoError(t, e.Load(shell))
	e.ConsumeAll()
	return e
}

func navTo(t *testing.T, e *engineX, path string) {
	t.Helper()
	u, err := url.Parse(path)
	require.NoError(t, err)
	e.page().ReplaceURL(u)
	e.nodes.NotifyComponentEvent(e.baseContext(), e.body, nav{})
	e.ConsumeAll()
}

func TestNewShellKeepsWhatItIsGiven(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	model := newScreenModel("ada")
	home := newRoutedStub("/", "home")

	shell, err := NewShell[screenState](layout, model, home)

	require.NoError(t, err)
	require.Equal(t, Layout[screenState](layout), shell.Layout)
	require.Same(t, model, shell.Model)
	require.Equal(t, []Routed{home}, shell.Screens)
	require.Equal(t, -1, shell.selected)
}

func TestNewShellRejectsMissingDependencies(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	model := newScreenModel("ada")
	screens := []Routed{newRoutedStub("/", "home")}

	cases := []struct {
		name    string
		layout  Layout[screenState]
		model   *Model[screenState]
		screens []Routed
		want    error
	}{
		{"no layout", nil, model, screens, ErrNoLayout},
		{"no model", layout, nil, screens, ErrNoModel},
		{"no screens", layout, model, nil, ErrNoScreens},
		{"nothing at all", nil, nil, nil, ErrNoLayout},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shell, err := NewShell(tc.layout, tc.model, tc.screens...)

			require.ErrorIs(t, err, tc.want)
			require.Nil(t, shell)
		})
	}
}

func TestNewShellAcceptsAnAllocatedButEmptyScreenList(t *testing.T) {
	shell, err := NewShell[screenState](&shellLayout{}, newScreenModel("ada"), []Routed{}...)

	require.NoError(t, err)
	require.NotNil(t, shell)
}

func TestAShellIsAMounterAndANavigatorButNotRouted(t *testing.T) {
	var shell any = &Shell[screenState]{}

	require.Implements(t, (*Mounter)(nil), shell)
	require.Implements(t, (*Navigator)(nil), shell)
	_, routed := shell.(Routed)
	require.False(t, routed)
}

func TestRenderDrawsTheSelectedScreenInsideTheLayout(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	shell := newTestShell(t, layout, newRoutedStub("/", "home"), newRoutedStub("/settings", "settings"))
	shell.setPath("/settings")
	loadShell(t, shell)

	requireMarkup(t, shell, `class="shell"`, drawnBy("shell"), `class="screen"`, "settings", `class="chrome"`)
	require.NotContains(t, HTMLString(shell), "home")
	require.NotZero(t, layout.drawCalls)
	require.Same(t, shell.Model, layout.drawnWith[0])
}

func TestShellDrawsItsLayoutOnTheFirstFrame(t *testing.T) {
	asClient(t)
	layout := &shellLayout{name: "shell"}
	shell := newTestShell(t, layout, newRoutedStub("/", "home"))
	shell.setPath("/")

	require.NoError(t, newTestEngine().Load(shell))

	require.Equal(t, 1, layout.drawCalls)
	require.Zero(t, layout.loadingCalls)
}

func TestSetPathSelectsBeforeTheFirstRender(t *testing.T) {
	a := newRoutedStub("/a", "a")
	b := newRoutedStub("/b", "b")
	shell := newTestShell(t, &shellLayout{name: "shell"}, a, b)

	shell.setPath("/b")
	loadShell(t, shell)

	require.Equal(t, 1, b.mountCalls)
	require.Zero(t, a.mountCalls)
	require.Equal(t, 1, shell.selected)
}

func TestSetPathStripsTheRootPrefix(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/a", "a"), newRoutedStub("/b", "b"))
	t.Setenv("GOAPP_ROOT_PREFIX", "/WebGo")

	shell.setPath("/WebGo/b")

	require.Equal(t, 1, shell.selected)
}

func TestSwitchingScreensDismountsTheOldOne(t *testing.T) {
	a := newRoutedStub("/a", "a")
	b := newRoutedStub("/b", "b")
	shell := newTestShell(t, &shellLayout{name: "shell"}, a, b)
	e := loadShell(t, shell)

	navTo(t, e, "/a")
	navTo(t, e, "/b")

	require.Equal(t, 1, a.mountCalls)
	require.Equal(t, 1, a.dismountCalls)
	require.Equal(t, 1, b.mountCalls)
	require.Zero(t, b.dismountCalls)
	requireMarkup(t, shell, `class="screen"`, ">b<")
}

func TestAShellWithNoScreensDrawsNoRoute(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	shell := newTestShell(t, layout)
	loadShell(t, shell)

	requireMarkup(t, shell, `class="error"`, drawnBy("shell"))
	require.Zero(t, layout.drawCalls)
	require.ErrorIs(t, layout.lastError(), ErrNoRoute)
}

func TestAnUnmatchedRouteIsDrawnAsAnErrorByTheOneLayout(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	shell := newTestShell(t, layout, newRoutedStub("/a", "alpha"))
	e := loadShell(t, shell)

	navTo(t, e, "/nowhere")

	requireMarkup(t, shell, `class="error"`, drawnBy("shell"))
	require.NotContains(t, HTMLString(shell), "alpha")
	require.ErrorIs(t, layout.lastError(), ErrNoRoute)
}

func TestAnOutOfRangeSelectionIsDrawnAsNoRoute(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	shell := newTestShell(t, layout, newRoutedStub("/", "home"))
	shell.selected = 7

	requireMarkup(t, shell.Render(), `class="error"`)
	require.ErrorIs(t, layout.lastError(), ErrNoRoute)
}

func TestAShellAssembledWithoutALayoutCannotRender(t *testing.T) {
	shell := &Shell[screenState]{Screens: []Routed{newRoutedStub("/", "home")}}

	panicked, value := didPanic(func() { shell.Render() })

	require.True(t, panicked)
	require.Equal(t, ErrNoLayout, value)
}

func TestAShellAssembledWithoutAModelDrawsTheError(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	shell := &Shell[screenState]{Layout: layout, Screens: []Routed{newRoutedStub("/", "home")}}

	panicked, value := didPanic(func() { shell.OnMount(makeTestContext()) })

	require.False(t, panicked, "%v", value)
	requireMarkup(t, shell.Render(), `class="error"`, drawnBy("shell"))
	require.ErrorIs(t, layout.lastError(), ErrNoModel)
}

func TestMountingAShellLoadsTheModel(t *testing.T) {
	ctx, _ := newModelContext(t)
	ctx.SetState("app-state", &screenState{User: "stored", N: 3})
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"))

	shell.OnMount(ctx)

	require.Equal(t, screenState{User: "stored", N: 3}, *shell.Model.State())
}

func TestMountingAShellMountsTheSelectedScreenOnce(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	home := newRoutedStub("/", "home")
	shell := newTestShell(t, layout, home)
	shell.setPath("/")
	loadShell(t, shell)

	requireMarkup(t, shell, `class="shell"`)
	require.Zero(t, layout.loadingCalls)
	require.Equal(t, 1, home.mountCalls)
}

func TestMountingAShellTwiceIsHarmless(t *testing.T) {
	ctx, _ := newModelContext(t)
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"))
	shell.setPath("/")

	shell.OnMount(ctx)
	shell.OnMount(ctx)

	requireMarkup(t, shell.Render(), `class="shell"`)
}

func TestNavigatingSelectsTheScreenWhoseRouteMatchesThePath(t *testing.T) {
	settings := newRoutedStub("/settings", "settings")
	home := newRoutedStub("/", "home")
	shell := newTestShell(t, &shellLayout{name: "shell"}, settings, home)
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, 1, shell.selected)
	require.Equal(t, 1, home.mountCalls)
}

func TestNavigatingDismountsTheScreenItLeaves(t *testing.T) {
	settings := newRoutedStub("/settings", "settings")
	home := newRoutedStub("/", "home")
	shell := newTestShell(t, &shellLayout{name: "shell"}, settings, home)
	shell.setPath("/settings")
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, 1, settings.dismountCalls)
	require.Zero(t, home.dismountCalls)
}

func TestNavigatingKeepsTheUrlItWasGiven(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/settings", "settings"), newRoutedStub("/", "home"))
	shell.setPath("/settings")
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, "/", e.page().URL().Path)
}

func TestNavigatingNowhereClearsTheSelection(t *testing.T) {
	layout := &shellLayout{name: "shell"}
	b := newRoutedStub("/b", "b")
	shell := newTestShell(t, layout, newRoutedStub("/a", "a"), b)
	shell.setPath("/b")
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, -1, shell.selected)
	require.Equal(t, 1, b.dismountCalls)
	requireMarkup(t, shell, `class="error"`, drawnBy("shell"))
}

func TestNavigatingKeepsTheScreenItIsAlreadyShowing(t *testing.T) {
	home := newRoutedStub("/", "home")
	shell := newTestShell(t, &shellLayout{name: "shell"}, home)
	shell.setPath("/")
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, 0, shell.selected)
	require.Equal(t, 1, home.mountCalls)
	require.Zero(t, home.dismountCalls)
	requireMarkup(t, shell, "home")
}

func TestNavigatingReadsEachRouteOnce(t *testing.T) {
	miss := newRoutedStub("/elsewhere", "elsewhere")
	match := newRoutedStub("/", "home")
	shell := newTestShell(t, &shellLayout{name: "shell"}, miss, match)
	e := loadShell(t, shell)
	miss.routeCalls, match.routeCalls = 0, 0

	navTo(t, e, "/")

	require.Equal(t, 1, miss.routeCalls)
	require.Equal(t, 1, match.routeCalls)
}

func TestNavigatingNormalizesTheScreenRouteBeforeComparing(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("//", "home"))
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, 0, shell.selected)
}

func TestNavigatingRestoresTheMountInTheUrl(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"), newRoutedStub("/markdown", "markdown"))
	t.Setenv("GOAPP_ROOT_PREFIX", "/WebGo")
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, 0, shell.selected)
	require.Equal(t, "/WebGo", e.page().URL().Path)
}

func TestRestoringTheMountKeepsTheQueryAndFragment(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"), newRoutedStub("/styling", "styling"))
	t.Setenv("GOAPP_ROOT_PREFIX", "/WebGo")
	e := loadShell(t, shell)

	navTo(t, e, "/styling?tab=go#values")

	require.Equal(t, 1, shell.selected)
	got := e.page().URL()
	require.Equal(t, "/WebGo/styling", got.Path)
	require.Equal(t, "tab=go", got.RawQuery)
	require.Equal(t, "values", got.Fragment)
}

func TestNavigatingLeavesACorrectUrlAlone(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"))
	e := loadShell(t, shell)

	navTo(t, e, "/")

	require.Equal(t, "/", e.page().URL().Path)
}

func TestDeepLinkUnderAPrefixMatchesAndRestoresTheUrl(t *testing.T) {
	state := newRoutedStub("/state", "state")
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"), state)
	t.Setenv("GOAPP_ROOT_PREFIX", "/WebGo")
	e := loadShell(t, shell)

	navTo(t, e, "/WebGo/state")

	require.Equal(t, 1, shell.selected)
	require.Equal(t, "/WebGo/state", e.page().URL().Path)
	require.Equal(t, 1, state.mountCalls)

	navTo(t, e, "/state")

	require.Equal(t, 1, shell.selected)
	require.Equal(t, "/WebGo/state", e.page().URL().Path)
	require.Equal(t, 1, state.mountCalls)
}

func TestARootPrefixOfSlashIsNoPrefix(t *testing.T) {
	shell := newTestShell(t, &shellLayout{name: "shell"}, newRoutedStub("/", "home"), newRoutedStub("/state", "state"))
	t.Setenv("GOAPP_ROOT_PREFIX", "/")
	e := loadShell(t, shell)

	navTo(t, e, "/state")

	require.Equal(t, 1, shell.selected)
	require.Equal(t, "/state", e.page().URL().Path)
}

func TestTrimBasePathReturnsTheScreenRoute(t *testing.T) {
	cases := []struct {
		name string
		path string
		base string
		want string
	}{
		{"no mount", "/markdown", "", "/markdown"},
		{"no mount at root", "/", "", "/"},
		{"mounted page", "/WebGo/markdown", "/WebGo", "/markdown"},
		{"mounted root", "/WebGo", "/WebGo", "/"},
		{"path outside the mount", "/other/markdown", "/WebGo", "/other/markdown"},
		{"mount is a partial segment", "/webhooks", "/web", "/webhooks"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, trimBasePath(normalizePath(tc.path), tc.base))
		})
	}
}

func TestNormalizePathStripsTrailingSlashesButKeepsTheRoot(t *testing.T) {
	cases := map[string]string{
		"":             "/",
		"/":            "/",
		"//":           "/",
		"/markdown":    "/markdown",
		"/markdown/":   "/markdown",
		"markdown":     "/markdown",
		"/WebGo/state": "/WebGo/state",
	}

	for input, want := range cases {
		require.Equal(t, want, normalizePath(input), "normalizePath(%q)", input)
	}
}

func TestBrowserPathIsTheInverseOfTrimBasePath(t *testing.T) {
	for _, prefix := range []string{"", "/WebGo", "/a/b"} {
		t.Setenv("GOAPP_ROOT_PREFIX", prefix)
		for _, route := range []string{"/", "/markdown", "/state"} {
			browser := browserPath(route)
			require.Equal(t, route, trimBasePath(normalizePath(browser), prefix), "prefix %q route %q gave %q", prefix, route, browser)
		}
	}
}
