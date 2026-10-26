package tests

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Testing"
	Route = "/testing"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "testing",
		Path:  Route,
		Title: Title,
		Lead:  "Test components with go test: render them to HTML, mount them in a test engine, and match their trees.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Rendering to HTML", Body: rendering},
			{Title: "The test engine", Body: engine},
			{Title: "Matching UI trees", Body: matching},
			{Title: "Testing with state and actions", Body: stateAndActions},
			{Title: "What runs only in the browser", Body: browserOnly},
			{Title: "Example: how these docs are tested", Body: docsExample},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("This page covers testing components with plain `go test`, without a browser: rendering to HTML, the `TestEngine` that mounts components and runs their queued work, and `Match`, which compares a rendered tree with an expected one. Tests run on the server side of the `app` package, so the last sections cover what behaves differently there."),
		doc.P("The examples test this component:"),
		doc.Go(`type greeting struct {
	app.Compo
	Name string
}

func (g *greeting) Render() app.UI {
	return app.Div().Class("greeting").Body(
		app.H1().Text("Hello, " + g.Name),
	)
}`),
	}
}

func rendering() []app.UI {
	return []app.UI{
		doc.Go(`func HTMLString(ui UI) string
func PrintHTML(w io.Writer, ui UI)`),
		doc.List(
			doc.Item("`HTMLString` renders `ui` to HTML with a fresh test engine. Components are rendered by calling `Render`, but nothing is mounted and no lifecycle hook runs, so what you see is the tree built from the fields as they are."),
			doc.Item("The output is indented: each element on its own line, nested two spaces per level. Assert on substrings that stay on one line, or strip the indentation first."),
			doc.Item("Event handlers do not appear in the HTML. A rule passed to `Style(rule)` appears as its class name in the `class` attribute."),
			doc.Item("`PrintHTML` writes the same text to `w`."),
		),
		doc.Go(`func TestGreetingRenders(t *testing.T) {
	out := app.HTMLString(&greeting{Name: "Ada"})
	if !strings.Contains(out, "<h1>Hello, Ada</h1>") {
		t.Errorf("unexpected HTML: %s", out)
	}
}`),
		doc.P("This is the quickest test, and often enough: render a component with given fields and check what it shows."),
	}
}

func engine() []app.UI {
	return []app.UI{
		doc.Go(`type TestEngine interface {
	Load(Composer) error
	ConsumeNext()
	ConsumeAll()
}

func NewTestEngine() TestEngine`),
		doc.P("`NewTestEngine` creates an engine like the one the server uses to prerender a page: its URL is `/`, it has no routes, and its local and session storage are in memory. Engines are independent of each other."),
		doc.H("Load"),
		doc.List(
			doc.Item("The first `Load` mounts the component as the root of the engine's page. `OnInit` runs at once. `OnPreRender` is queued, and `OnMount` does not run, because tests run on the server side."),
			doc.Item("A later `Load` with a component of the same type updates the mounted one in place: its exported fields are set from the new value, and `OnUpdate` runs if they changed. A component of another type replaces it, and the old one's `OnDismount` runs."),
			doc.Item("A component can be mounted in one engine at a time. Loading a mounted component again returns an error."),
		),
		doc.H("ConsumeAll and ConsumeNext"),
		doc.List(
			doc.Item("Nothing queued runs until you consume it: not `OnPreRender`, not a `Dispatch`, not an observer's delivery."),
			doc.Item("`ConsumeAll` runs every queued dispatch, re-renders the flagged components, runs the deferred functions, waits for every `Async` goroutine, and repeats until the queue is empty. A goroutine that never returns makes it wait forever."),
			doc.Item("`ConsumeNext` waits for the `Async` goroutines, runs one queued dispatch, then re-renders and runs deferred functions. If nothing is queued, it blocks forever."),
		),
		doc.H("Getting a Context"),
		doc.P("Many methods to test take an `app.Context`, and `Load` does not return one. A component that implements `OnPreRender` receives one when the engine is consumed. Otherwise, mount a small host component twice:"),
		doc.Go(`type host struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (h *host) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *host) Render() app.UI { return app.Div() }

func newContext(t *testing.T) (app.Context, app.TestEngine) {
	t.Helper()
	engine := app.NewTestEngine()
	var ctx app.Context
	if err := engine.Load(&host{captured: &ctx}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Load(&host{Tick: 1}); err != nil {
		t.Fatal(err)
	}
	engine.ConsumeAll()
	return ctx, engine
}`),
		doc.P("The second `Load` changes the exported `Tick`, so the mounted host gets `OnUpdate`. Its unexported `captured` still points at `ctx`, which receives the host's context. The host stays mounted, so dispatches made through `ctx` are not skipped."),
	}
}

func matching() []app.UI {
	return []app.UI{
		doc.Go(`func Match(expected UI, root UI, path ...int) error

type TestUIDescriptor struct {
	Path     []int
	Expected UI
}

func TestPath(p ...int) []int
func TestMatch(root UI, d TestUIDescriptor) error`),
		doc.P("`Match` walks from `root` along `path` and compares the node it reaches with `expected`. It returns nil on a match, and otherwise an error describing the first difference. `TestMatch(root, TestUIDescriptor{Path: TestPath(path...), Expected: expected})` is the older spelling of the same call."),
		doc.H("Paths"),
		doc.List(
			doc.Item("Through an HTML element, index `i` is its `i`th child."),
			doc.Item("Through a component, the only index is `0`, its rendered root. A component has a root only once it is mounted, so `Load` it first. On a component that is not mounted, `Match` with a path fails."),
			doc.Item("An index out of range is an error."),
		),
		doc.H("Comparison"),
		doc.List(
			doc.Item("The types must be the same."),
			doc.Item("Text is compared by value."),
			doc.Item("An HTML element must have the same tag and exactly the same attributes, none missing and none extra. A class on the element that `expected` lacks is a mismatch. Event handlers are compared by event name only, not by function. Children are not compared: reach them with a longer path."),
			doc.Item("A component is compared by its exported fields, with `reflect.DeepEqual`. Unexported fields are ignored."),
			doc.Item("`Raw` HTML is compared by value."),
		),
		doc.Go(`func TestGreetingTree(t *testing.T) {
	g := &greeting{Name: "Ada"}
	engine := app.NewTestEngine()
	if err := engine.Load(g); err != nil {
		t.Fatal(err)
	}
	engine.ConsumeAll()

	if err := app.Match(&greeting{Name: "Ada"}, g); err != nil {
		t.Error(err)
	}
	if err := app.Match(app.Div().Class("greeting"), g, 0); err != nil {
		t.Error(err)
	}
	if err := app.Match(app.Text("Hello, Ada"), g, 0, 0, 0); err != nil {
		t.Error(err)
	}
}`),
		doc.P("The path `0, 0, 0` goes from `greeting` to its root `div`, to the `h1`, to the text inside it."),
	}
}

func stateAndActions() []app.UI {
	return []app.UI{
		doc.P("States and component action handlers work in a test engine, through a captured `ctx`. A few rules make them reach your component:"),
		doc.List(
			doc.Item("`OnMount` does not run, so observers and handlers registered there do not exist. Call the component's `OnMount(ctx)` yourself. They then belong to the host, which is mounted, and still write into your component's fields."),
			doc.Item("`SetState` and `NewAction` only queue work. Call `ConsumeAll` before checking the result."),
			doc.Item("Global handlers registered with `app.Handle` do not run in a test engine. Call the handler function directly, with `ctx` and an `app.Action`."),
			doc.Item("`Persist` writes to the engine's in-memory storage. `Broadcast` logs that it is not supported."),
			doc.Item("Each engine has its own state store, but a `data.ViewModel` keeps its own `*T`, often in a package variable. Tests that share one model share its value."),
		),
		doc.P("With `results` from the search example on ", doc.A("/actions#handling-actions", "State and actions"), ":"),
		doc.Go(`func TestResultsFollowTheSearch(t *testing.T) {
	ctx, engine := newContext(t)
	r := &results{}
	r.OnMount(ctx)

	ctx.NewActionWithValue("search", "go")
	ctx.SetState("search-hits", []string{"go.dev"})
	engine.ConsumeAll()

	if r.query != "go" || len(r.hits) != 1 {
		t.Fatalf("got query %q and hits %v", r.query, r.hits)
	}
}`),
	}
}

func browserOnly() []app.UI {
	return []app.UI{
		doc.P("`go test` builds for your machine, so `app.IsServer` is true and the browser is not there. In practice:"),
		doc.List(
			doc.Item("`OnMount` never runs on its own, and neither do `OnNav`, `OnAppUpdate`, `OnAppInstallChange` or `OnResize`. Call them with a captured `ctx` to test them."),
			doc.Item("`OnInit` runs on `Load`, and `OnPreRender` on the next consume."),
			doc.Item("`app.Window()` and every `Value` are stubs. Reads return zero values, calls do nothing, and `Value.Type` panics. `FuncOf` callbacks never run."),
			doc.Item("`ctx.Notifications()` reports `NotificationNotSupported`."),
			doc.Item("Dynamic styles are not inserted, because `style.Run` returns at once on the server. Test a bound rule with its `CSSFor` method, as on ", doc.A("/applying#actions", "Applying"), "."),
		),
		doc.P("Code that only makes sense in a browser, such as JavaScript calls and DOM measurements, is best kept in small functions that the rest of the component does not depend on, so the logic around it stays testable. To exercise it for real, build with `GOOS=js GOARCH=wasm` and run the app in a browser."),
	}
}

func docsExample() []app.UI {
	return []app.UI{
		doc.P("This site is a WebGo app, and its tests use the techniques above. Most render a piece of a page with `HTMLString`, strip the indentation, and look for substrings:"),
		doc.Go(`var indent = regexp.MustCompile(` + "`\\n *`" + `)

func html(ui ...app.UI) string {
	var out strings.Builder
	for _, u := range ui {
		out.WriteString(indent.ReplaceAllString(app.HTMLString(u), ""))
	}
	return out.String()
}

func TestInlineTurnsBackticksIntoCodeSpans(t *testing.T) {
	got := html(doc.P("A ` + "`Rule`" + ` is one selector."))
	if !strings.Contains(got, "<p>A <code>Rule</code>") {
		t.Errorf("unexpected HTML: %s", got)
	}
}`),
		doc.P("Pages with sections are tested by mounting them. The topic tests capture a `ctx` with a host as above, call the screen's `OnMount(ctx)`, change the URL with `ctx.Page().ReplaceURL`, call `OnNav(ctx)` by hand, consume, and then check which section `HTMLString` shows."),
		doc.P("Live examples have one more check. Each example's source is shown as text on the page, so a test reads the Go files of the example's package and fails unless the shown text appears there twice: once as the constant that is displayed, and once as the code that runs. The shown code cannot drift from the code that runs, and every live example is also rendered with `HTMLString` to make sure it builds a tree."),
	}
}
