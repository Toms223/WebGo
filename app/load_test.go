package app

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func asClient(t *testing.T) {
	isClient = true
	t.Cleanup(func() { isClient = IsClient })
}

type loading struct {
	Compo
	Result  string
	Err     error
	loads   int
	mounted bool
	order   []string
}

func (l *loading) OnMount(Context) {
	l.mounted = true
	l.order = append(l.order, "mount")
}

func (l *loading) OnLoad(Context) error {
	l.loads++
	l.order = append(l.order, "load")
	return l.Err
}

func (l *loading) Loading() UI        { return Text("loading") }
func (l *loading) Error(err error) UI { return Text("failed: " + err.Error()) }
func (l *loading) Render() UI         { return Text("loaded " + l.Result) }

func TestLoaderShowsLoadingThenRenders(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &loading{Result: "ok"}
	require.NoError(t, e.Load(c))
	require.Equal(t, "loading", c.root().(*text).value)
	e.ConsumeAll()
	require.True(t, c.mounted)
	require.Equal(t, 1, c.loads)
	require.Equal(t, "loaded ok", c.root().(*text).value)
}

func TestOnMountRunsBeforeOnLoad(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &loading{Result: "ok"}
	require.NoError(t, e.Load(c))
	e.ConsumeAll()
	require.Equal(t, []string{"mount", "load"}, c.order)
}

func TestLoaderIsSkippedOnTheServer(t *testing.T) {
	isClient = false
	t.Cleanup(func() { isClient = IsClient })
	e := newTestEngine()
	c := &loading{Result: "ok"}
	require.NoError(t, e.Load(c))
	require.Equal(t, "loaded ok", c.root().(*text).value)
	e.ConsumeAll()
	require.Equal(t, 0, c.loads)
	require.Equal(t, "loaded ok", c.root().(*text).value)
}

func TestComponentWithoutLoaderIsUnchanged(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &hello{Greeting: "world"}
	require.NoError(t, e.Load(c))
	require.IsType(t, &htmlDiv{}, c.root())
	e.ConsumeAll()
	require.True(t, c.mounted)
}

type countingLoading struct {
	loading
	renders int
}

func (c *countingLoading) Render() UI {
	c.renders++
	return c.loading.Render()
}

type silentLoader struct {
	Compo
}

func (silentLoader) OnLoad(Context) error { return nil }
func (silentLoader) Render() UI           { return Text("done") }

type plainErrLoader struct {
	Compo
}

func (plainErrLoader) OnLoad(Context) error { return errors.New("boom") }
func (plainErrLoader) Render() UI           { return Text("rendered") }

func TestUpdateComponentRootHoldsWhileLoading(t *testing.T) {
	asClient(t)
	var m nodeManager
	ctx := makeTestContext()
	c := &countingLoading{loading: loading{Result: "ok"}}
	_, err := m.Mount(ctx, 1, c)
	require.NoError(t, err)
	root := c.root()
	require.Equal(t, "loaded ok", root.(*text).value)
	renders := c.renders

	c.load().loading = true
	_, err = m.UpdateComponentRoot(m.context(ctx, c), c)
	require.NoError(t, err)
	require.Same(t, root, c.root())
	require.Equal(t, "loaded ok", c.root().(*text).value)
	require.Equal(t, renders, c.renders)

	c.load().loading = false
	_, err = m.UpdateComponentRoot(m.context(ctx, c), c)
	require.NoError(t, err)
	require.Equal(t, renders+1, c.renders)
}

func TestLoaderWithoutLoadingRendererDrawsNothing(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &silentLoader{}
	require.NoError(t, e.Load(c))
	require.Equal(t, "", c.root().(*text).value)
	e.ConsumeAll()
	require.Equal(t, "done", c.root().(*text).value)
}

func TestLoadErrorShowsErrorView(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &loading{Err: errors.New("boom")}
	require.NoError(t, e.Load(c))
	e.ConsumeAll()
	require.Equal(t, "failed: boom", c.root().(*text).value)
}

func TestLoadErrorWithoutErrorRendererFallsBackToRender(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &plainErrLoader{}
	require.NoError(t, e.Load(c))
	e.ConsumeAll()
	require.Equal(t, "rendered", c.root().(*text).value)
}

func TestSiblingRendersWhileOneLoads(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	slow := &loading{Result: "a"}
	h := &hello{Greeting: "x"}
	parent := &compoWithCustomRoot{Root: Div().Body(slow, h)}
	require.NoError(t, e.Load(parent))
	require.Equal(t, "loading", slow.root().(*text).value)
	require.IsType(t, &htmlDiv{}, h.root())
	e.ConsumeAll()
	require.Equal(t, "loaded a", slow.root().(*text).value)
}

func capturingContext(t *testing.T) (nodeManager, Context, *loading, *[]func()) {
	asClient(t)
	var m nodeManager
	ctx := makeTestContext()
	captured := &[]func(){}
	ctx.async = func(f func()) { *captured = append(*captured, f) }
	c := &loading{Result: "ok"}
	_, err := m.Mount(ctx, 1, c)
	require.NoError(t, err)
	return m, ctx, c, captured
}

func TestResultFromAnOlderLoadIsDropped(t *testing.T) {
	m, ctx, c, captured := capturingContext(t)
	require.True(t, c.load().loading)
	c.load().loads++
	startLoad(m.context(ctx, c), c)
	require.Equal(t, 2, c.load().loads)
	require.Len(t, *captured, 2)
	(*captured)[0]()
	require.True(t, c.load().loading)
	(*captured)[1]()
	require.False(t, c.load().loading)
}

func TestLoadNeverStartsAfterDismount(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	c := &loading{Result: "ok"}
	require.NoError(t, e.Load(c))
	require.NoError(t, e.Load(&hello{Greeting: "bye"}))
	e.ConsumeAll()
	require.Equal(t, 0, c.loads)
}

func TestResultLandingAfterDismountIsDropped(t *testing.T) {
	m, _, c, captured := capturingContext(t)
	m.Dismount(c)
	require.NotPanics(t, func() { (*captured)[0]() })
	require.True(t, c.load().loading)
}

func TestFieldUpdateStartsANewLoad(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	require.NoError(t, e.Load(&loading{Result: "a"}))
	e.ConsumeAll()
	require.NoError(t, e.Load(&loading{Result: "b"}))
	root := e.body.body()[0].(*loading)
	require.Equal(t, "loading", root.root().(*text).value)
	e.ConsumeAll()
	require.Equal(t, 2, root.loads)
	require.Equal(t, "loaded b", root.root().(*text).value)
}

func TestFieldUpdateWithoutChangeDoesNotReload(t *testing.T) {
	asClient(t)
	e := newTestEngine()
	require.NoError(t, e.Load(&loading{Result: "a"}))
	e.ConsumeAll()
	require.NoError(t, e.Load(&loading{Result: "a"}))
	e.ConsumeAll()
	require.Equal(t, 1, e.body.body()[0].(*loading).loads)
}
