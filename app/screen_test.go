package app

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type screenState struct {
	User string
	N    int
}

type recordingLayout struct {
	name         string
	loadingCalls int
	drawCalls    int
	errorCalls   int
	errs         []error
	drawnWith    []*Model[screenState]
	body         []UI
}

func (l *recordingLayout) Loading() UI {
	l.loadingCalls++
	return Div().Class("loading").DataSet("layout", l.name).Text("loading")
}

func (l *recordingLayout) Error(err error) UI {
	l.errorCalls++
	l.errs = append(l.errs, err)
	return Div().Class("error").DataSet("layout", l.name).Text(err.Error())
}

func (l *recordingLayout) Body(content ...UI) Layout[screenState] {
	l.body = append(l.body, content...)
	return l
}

func (l *recordingLayout) Draw(m *Model[screenState]) HTMLDiv {
	l.drawCalls++
	l.drawnWith = append(l.drawnWith, m)
	body := []UI{Header().Class("chrome").Text("chrome")}
	body = append(body, l.body...)
	return Div().Class("shell").DataSet("layout", l.name).Body(body...)
}

func (l *recordingLayout) lastError() error {
	if len(l.errs) == 0 {
		return nil
	}
	return l.errs[len(l.errs)-1]
}

func newRecordingLayout(name string) *recordingLayout {
	return &recordingLayout{name: name}
}

func newScreenModel(user string) *Model[screenState] {
	m, _ := NewModel("app-state", &screenState{User: user})
	return m
}

func drawnBy(name string) string {
	return `data-layout="` + name + `"`
}

func requireMarkup(t *testing.T, ui UI, fragments ...string) {
	t.Helper()
	markup := HTMLString(ui)
	for _, fragment := range fragments {
		if !strings.Contains(markup, fragment) {
			t.Fatalf("expected %q in %s", fragment, markup)
		}
	}
}

func TestALayoutServesAScreen(t *testing.T) {
	var layout any = newRecordingLayout("shared")
	_, ok := layout.(Layout[screenState])
	require.True(t, ok)
}

func TestAScreenIsBuiltByChainingALayoutAndAModel(t *testing.T) {
	layout := newRecordingLayout("page")
	m := newScreenModel("ada")

	s := NewScreen[screenState]("/").WithLayout(layout).WithModel(m)
	require.NoError(t, s.Err())

	ctx, _ := newModelContext(t)
	s.OnMount(ctx)

	requireMarkup(t, s.Render(), `class="shell"`, drawnBy("page"))
	require.Same(t, m, layout.drawnWith[0])
}

func TestChainingReturnsTheSameScreenSoTheOrderDoesNotMatter(t *testing.T) {
	layout := newRecordingLayout("page")
	m := newScreenModel("ada")

	first := NewScreen[screenState]("/").WithLayout(layout).WithModel(m)
	second := NewScreen[screenState]("/").WithModel(m).WithLayout(layout)

	ctx, _ := newModelContext(t)
	first.OnMount(ctx)
	second.OnMount(ctx)

	requireMarkup(t, first.Render(), `class="shell"`)
	requireMarkup(t, second.Render(), `class="shell"`)
}

func TestAMissingLayoutIsReportedAsAnError(t *testing.T) {
	s := NewScreen[screenState]("/").WithLayout(nil).WithModel(newScreenModel("ada"))
	require.ErrorIs(t, s.Err(), ErrNoLayout)
}

func TestAMissingModelIsReportedAsAnError(t *testing.T) {
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(nil)
	require.ErrorIs(t, s.Err(), ErrNoModel)
}

func TestTheFirstErrorInAChainWins(t *testing.T) {
	s := NewScreen[screenState]("/").WithLayout(nil).WithModel(nil)
	require.ErrorIs(t, s.Err(), ErrNoLayout)
}

func TestAChainStopsApplyingStepsAfterAFailure(t *testing.T) {
	layout := newRecordingLayout("page")

	s := NewScreen[screenState]("/").WithModel(nil).WithLayout(layout).At("/other")

	require.ErrorIs(t, s.Err(), ErrNoModel)
	require.Equal(t, "/", s.Path)
	require.Zero(t, layout.drawCalls)
	require.Zero(t, layout.loadingCalls)
}

func TestAScreenBuiltWithoutALayoutReportsItThroughTheError(t *testing.T) {
	s := NewScreen[screenState]("/").WithModel(newScreenModel("ada"))
	require.ErrorIs(t, s.Err(), ErrNoLayout)
}

func TestAScreenWithoutALayoutCannotRenderAnything(t *testing.T) {
	s := NewScreen[screenState]("/").WithModel(newScreenModel("ada"))

	for _, f := range []func(){
		func() { s.Render() },
		func() { s.Loading() },
		func() { s.Error(errors.New("boom")) },
	} {
		panicked, value := didPanic(f)
		require.True(t, panicked)
		require.Equal(t, ErrNoLayout, value)
	}
}

func TestAScreenWithoutAModelDrawsTheErrorInsteadOfTheContent(t *testing.T) {
	layout := newRecordingLayout("page")
	s := NewScreen[screenState]("/").WithLayout(layout)

	requireMarkup(t, s.Render(), `class="error"`, drawnBy("page"))
	require.ErrorIs(t, layout.lastError(), ErrNoModel)
	require.Zero(t, layout.drawCalls)
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
			s := NewScreen[screenState](tc.path)
			require.Equal(t, tc.want, s.Route())
			require.Equal(t, tc.want, s.CompoID())
		})
	}
}

func TestAMountedScreenKeepsTheLayoutsOwnChildren(t *testing.T) {
	ctx, _ := newModelContext(t)
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(newScreenModel("ada"))
	s.OnMount(ctx)

	requireMarkup(t, s.Render(), `class="chrome"`)
}

func TestMountingReadsTheModelFromTheContext(t *testing.T) {
	ctx, _ := newModelContext(t)
	m := newScreenModel("seed")
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(m)

	ctx.SetState("app-state", &screenState{User: "stored", N: 9})
	s.OnMount(ctx)

	requireMarkup(t, s.Render(), `class="shell"`)
	require.Equal(t, screenState{User: "stored", N: 9}, *m.State())
}

func TestMountingAScreenWithoutAModelIsReportedNotFatal(t *testing.T) {
	ctx, _ := newModelContext(t)
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page"))

	panicked, value := didPanic(func() { s.OnMount(ctx) })
	require.False(t, panicked, "%v", value)
	require.ErrorIs(t, s.Err(), ErrNoModel)
	requireMarkup(t, s.Render(), `class="error"`)
}

func TestAScreenWithALoaderShowsLoadingThenDraws(t *testing.T) {
	asClient(t)
	layout := newRecordingLayout("page")
	m := newScreenModel("seed")
	s := NewScreen[screenState]("/").WithLayout(layout).WithModel(m).
		WithLoader(func(_ Context, st *screenState) error {
			st.User = "ada"
			return nil
		})

	e := newTestEngine()
	require.NoError(t, e.Load(&compoWithCustomRoot{Root: Div().Body(s)}))
	require.Equal(t, 1, layout.loadingCalls)
	require.Zero(t, layout.drawCalls)

	e.ConsumeAll()
	require.Equal(t, 1, layout.drawCalls)
	require.Equal(t, "ada", m.State().User)
}

func TestAScreenLoaderErrorDrawsTheLayoutsError(t *testing.T) {
	asClient(t)
	boom := errors.New("boom")
	layout := newRecordingLayout("page")
	s := NewScreen[screenState]("/").WithLayout(layout).WithModel(newScreenModel("seed")).
		WithLoader(func(Context, *screenState) error { return boom })

	e := newTestEngine()
	require.NoError(t, e.Load(&compoWithCustomRoot{Root: Div().Body(s)}))
	e.ConsumeAll()

	require.Equal(t, boom, layout.lastError())
	require.Zero(t, layout.drawCalls)
}

func TestAScreenLoaderWorksOnACopy(t *testing.T) {
	ctx, engine := newModelContext(t)
	m := newScreenModel("seed")
	var during string
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(m).
		WithLoader(func(_ Context, st *screenState) error {
			st.User = "ada"
			during = m.State().User
			return nil
		})

	require.NoError(t, s.OnLoad(ctx))
	require.Equal(t, "seed", during)
	engine.ConsumeAll()
	require.Equal(t, "ada", m.State().User)
}

func TestAScreenLoaderErrorLeavesTheModelAlone(t *testing.T) {
	ctx, engine := newModelContext(t)
	m := newScreenModel("seed")
	boom := errors.New("boom")
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(m).
		WithLoader(func(_ Context, st *screenState) error {
			st.User = "ada"
			return boom
		})

	require.Equal(t, boom, s.OnLoad(ctx))
	engine.ConsumeAll()
	require.Equal(t, "seed", m.State().User)
}

func TestAScreenLoaderPublishesThroughSet(t *testing.T) {
	ctx, engine := newModelContext(t)
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(newScreenModel("seed")).
		WithLoader(func(_ Context, st *screenState) error {
			st.User = "ada"
			return nil
		})

	other := newScreenModel("other")
	other.Observe(ctx)

	require.NoError(t, s.OnLoad(ctx))
	engine.ConsumeAll()
	require.Equal(t, "ada", other.State().User)
}

func TestAScreenWithoutALoaderLoadsNothing(t *testing.T) {
	ctx, _ := newModelContext(t)
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(newScreenModel("seed"))
	require.NoError(t, s.OnLoad(ctx))
}

func TestScreensOnDifferentRoutesHaveDifferentCompoIDs(t *testing.T) {
	require.NotEqual(t, NewScreen[screenState]("/a").CompoID(), NewScreen[screenState]("/b").CompoID())
	require.Equal(t, NewScreen[screenState]("a").CompoID(), NewScreen[screenState]("/a").CompoID())
}

func TestAScreenIsRouted(t *testing.T) {
	var _ Routed = (*Screen[screenState])(nil)
	var _ Loader = (*Screen[screenState])(nil)
	var _ LoadingRenderer = (*Screen[screenState])(nil)
	var _ LoadErrorRenderer = (*Screen[screenState])(nil)
	var _ DismountEnforcer = (*Screen[screenState])(nil)
	var _ Mounter = (*Screen[screenState])(nil)
}

func TestAnOlderScreenLoadDoesNotOverwriteANewerOne(t *testing.T) {
	ctx, _ := newModelContext(t)
	var pending []func()
	ctx.dispatch = func(f func()) { pending = append(pending, f) }
	m := newScreenModel("seed")
	calls := 0
	s := NewScreen[screenState]("/").WithLayout(newRecordingLayout("page")).WithModel(m).
		WithLoader(func(_ Context, st *screenState) error {
			calls++
			st.N = calls
			return nil
		})

	s.load().loads = 1
	require.NoError(t, s.OnLoad(ctx))
	s.load().loads = 2
	require.NoError(t, s.OnLoad(ctx))
	require.Len(t, pending, 2)

	pending[1]()
	require.Equal(t, 2, m.State().N)
	pending[0]()
	require.Equal(t, 2, m.State().N)
}
