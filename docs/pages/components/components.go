package components

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Components"
	Route = "/components"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "components",
		Path:  Route,
		Title: Title,
		Lead:  "A component is a Go struct that embeds `app.Compo` and returns its markup from `Render`.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Architecture", Body: architecture},
			{Title: "Writing a component", Body: writing},
			{Title: "Elements and attributes", Body: elements},
			{Title: "Events", Body: events},
			{Title: "Conditions and lists", Body: conditions},
			{Title: "Lifecycle", Body: lifecycle},
			{Title: "Updating the UI", Body: updating},
			{Title: "Raw HTML and custom elements", Body: raw},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("This page covers the `app` package's component model: how a WebGo program runs on the server and in the browser, how a component describes its markup, how it reacts to events, and when it renders again. This is go-app's component model, now WebGo's `app` package."),
		doc.P("Choosing which component a URL shows is on the ", doc.A("/routing", "Routing"), " page. WebGo's `page.Page`, `page.Base` and `data.ViewModel`, built on these components, are on the ", doc.A("/state", "Pages, layouts and state"), " page."),
	}
}

func architecture() []app.UI {
	return []app.UI{
		doc.P("A WebGo app is one Go program built twice: once for the server, and once with `GOOS=js GOARCH=wasm` into `web/app.wasm` for the browser. The same `main` runs in both. `app.IsServer` and `app.IsClient` report which build is running."),
		architectureDiagram(),
		doc.H("The server"),
		doc.P("The server build serves an `app.Handler` over `net/http`, or writes the same files to disk with `app.GenerateStaticWebsite`. For every route, the handler answers with an HTML page that contains:"),
		doc.List(
			doc.Item("the markup of the routed component, prerendered on the server;"),
			doc.Item("links to `app.css` and to the handler's `Styles` (see ", doc.A("/applying#handler-and-build", "Handler and Build"), ");"),
			doc.Item("`wasm_exec.js`, Go's WebAssembly support file, and `app.js`, which downloads `app.wasm` and registers the service worker `app-worker.js`;"),
			doc.Item("`manifest.webmanifest`, which makes the app installable."),
		),
		doc.P("Files under `/web/` (`app.wasm`, images, stylesheets) are static resources. The handler's `Resources` field decides where they are served from: a local directory by default, a bucket with `app.RemoteBucket`, or a prefixed path with `app.GitHubPages`."),
		doc.H("The browser"),
		doc.P("The browser shows the prerendered markup at once. When `app.wasm` has loaded, `app.RunWhenOnBrowser` creates the component routed to the current URL, mounts it, and replaces the prerendered markup with the live tree. From then on the app is a single-page application: link clicks and back/forward are handled in Go without a page load."),
		doc.Go(`func RunWhenOnBrowser()`),
		doc.P("On the server, `RunWhenOnBrowser` returns immediately, so the code after it (starting the handler) runs. In the browser it never returns. Routes must be registered before it is called, and they must be registered on both sides, because the server needs them to prerender."),
	}
}

func writing() []app.UI {
	return []app.UI{
		doc.Go(`type Compo struct {
}

type Composer interface {
	UI
	Render() UI
}`),
		doc.P("A component is a struct that embeds `app.Compo` and defines `Render() app.UI`. `Composer` also has unexported methods, which only `*app.Compo` provides, so a component is always used through a pointer: `&greeting{}` is a `Composer`, `greeting{}` is not and does not compile where a `Composer` is expected."),
		doc.Go(`package main

import "github.com/Toms223/WebGo/app"

type greeting struct {
	app.Compo
	Name  string
	shown int
}

func (g *greeting) Render() app.UI {
	return app.P().Textf("Hello, %s", g.Name)
}

type home struct {
	app.Compo
}

func (h *home) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Home"),
		&greeting{Name: "WebGo"},
	)
}`),
		doc.P("`Render` must return one element, text or component. A `Render` that returns nothing (`nil`, or an `If` whose condition is false) fails to mount with an error. If it returns an `If` or `Range` that expands to several elements, only the first is used, so wrap lists in an element. A component without its own `Render` shows a placeholder box naming the type."),
		doc.H("Exported and unexported fields"),
		doc.P("Every time a parent renders, it creates new child values, such as `&greeting{Name: \"WebGo\"}` above. The engine does not mount the new value. It compares it with the child already mounted at that place:"),
		doc.List(
			doc.Item("If the types differ, the old child is dismounted and the new one is mounted."),
			doc.Item("If the types match, each exported field that differs is copied into the mounted child, and the child renders again. Unexported fields are never compared or copied, so `shown` above keeps its value across parent renders."),
			doc.Item("If no exported field differs, the mounted child is left as it is and does not render."),
		),
		doc.P("Exported fields are a component's inputs; unexported fields are its own state. Numbers, strings, booleans and `time.Time` are compared by value, everything else with `reflect.DeepEqual`. A non-nil `func` field never equals another, so a child with one is updated on every parent render."),
		doc.Note("One component value can only be mounted once. Putting the same pointer in two places of the tree fails with `component is already mounted`."),
		doc.H("Compo methods"),
		doc.Go(`func (c *Compo) Mounted() bool
func (c *Compo) JSValue() Value
func (c *Compo) ValueTo(v any) EventHandler`),
		doc.List(
			doc.Item("`Mounted` reports whether the component is in the tree. It turns false when the component is dismounted."),
			doc.Item("`JSValue` returns the DOM node of the component's root element, or `nil` before it has one."),
			doc.Item("`ValueTo` returns an event handler that stores the source element's `value` into `v`; see ", doc.A("#events", "Events"), "."),
		),
	}
}

func elements() []app.UI {
	return []app.UI{
		doc.P("Every HTML element has a constructor, such as `app.Div()`, `app.A()` or `app.Input()`, that returns a typed interface (`app.HTMLDiv`, `app.HTMLA`, `app.HTMLInput`). Attributes are set by chaining methods, and each method returns the same element."),
		doc.Go(`app.Div().ID("intro").Class("card", "wide").Body(
	app.H2().Text("Components"),
	app.P().Textf("%d sections", 8),
	app.Img().Src("/web/logo.svg").Alt("WebGo"),
	app.A().Href("/routing").Text("Next: Routing"),
)`),
		doc.List(
			doc.Item("`Body(children ...app.UI)` sets the children. Self-closing elements (`Img`, `Input`, `Br`, `Hr`, `Meta`, `Link`…) have no `Body`."),
			doc.Item("`Text(v any)` replaces the children with one text node made from any value, so `Text` and `Body` on one element do not combine; the last call wins. `Textf(format, v...)` formats it. `app.Text` and `app.Textf` make the same text node outside an element."),
			doc.Item("`Class(v ...string)` adds class names. Calling it again adds more; it never replaces."),
			doc.Item("Every element has `ID`, `Title`, `Hidden`, `TabIndex`, `Lang`, `Dir`, `Role`, `Draggable`, `ContentEditable` and `Spellcheck`, plus the methods for its own attributes (`Href`, `Src`, `Type`, `Value`, `Checked`…)."),
			doc.Item("`Aria(k, v)` sets `aria-k`, `DataSet(k, v)` sets `data-k`, `DataSets(map)` sets several, and `Attr(name, v)` sets any attribute by name."),
		),
		doc.H("Style and RawStyle"),
		doc.Go(`type ClassRule interface {
	ClassNames() []string
}

type Styler[E any] interface {
	UI
	Class(v ...string) E
	Style(rules ...ClassRule) E
	RawStyle(k, v string) E
	RawStylef(k, format string, v ...any) E
	RawStyles(s map[string]string) E
}`),
		doc.P("Every element implements `Styler` of its own type. `Style` takes typed rules from the `style` package and adds their class names to the element. A rule without class names (an element or id selector) is skipped and logged. Writing rules is on the ", doc.A("/styling", "Styling"), " page."),
		doc.P("`RawStyle(k, v)` appends an inline `k:v` declaration to the `style` attribute, unchecked. `RawStylef` formats the value, and `RawStyles` adds a map of them. These are go-app's `Style`, `Stylef` and `Styles`, renamed so that `Style` could take typed rules."),
		doc.Go(`var Card = style.Class("card").Set(
	css.Padding(css.Rem(1)),
	css.Border(css.Px(2), kw.Solid, css.Var("--ink")),
)

app.Div().Style(Card).RawStyle("width", "400px").Text("Typed class, inline width")`),
		doc.H("Rendering to a string"),
		doc.Go(`func HTMLString(ui UI) string
func PrintHTML(w io.Writer, ui UI)`),
		doc.P("Both write the markup of an element tree without mounting it. Lifecycle hooks do not run, so a component renders with the fields it was created with. They are useful in tests."),
	}
}

func events() []app.UI {
	return []app.UI{
		doc.Go(`type EventHandler func(ctx Context, e Event)

type Event struct {
	Value
}

func (e Event) PreventDefault()
func (e Event) StopImmediatePropagation()`),
		doc.P("Events are set with `On<Event>` methods such as `OnClick`, `OnInput`, `OnChange`, `OnSubmit` and `OnKeyDown`, or with `On(name, h)` for any event name. Each takes an `EventHandler` and optional `EventOption`s. Setting the same event twice keeps the last handler."),
		doc.Go(`type form struct {
	app.Compo
	name string
}

func (f *form) Render() app.UI {
	return app.Form().OnSubmit(f.submit).Body(
		app.Input().Type("text").Value(f.name).OnInput(f.ValueTo(&f.name)),
		app.Button().Type("submit").Text("Save"),
		app.P().Textf("Hello, %s", f.name),
	)
}

func (f *form) submit(ctx app.Context, e app.Event) {
	e.PreventDefault()
	ctx.LocalStorage().Set("name", f.name)
}`),
		doc.List(
			doc.Item("Handlers run on the UI goroutine. When one returns, the component that owns the element and every component above it render again, so changing a field in a handler is enough to update the screen."),
			doc.Item("`ctx` is bound to the element the event came from. `ctx.JSSrc()` is its DOM node, so `ctx.JSSrc().Get(\"checked\").Bool()` reads a checkbox. `ctx` also embeds a `context.Context`, the app's own, which is not cancelled when the element goes away; instead, `ctx.Dispatch`, `ctx.Defer` and `ctx.After` skip their function once the element is dismounted."),
			doc.Item("`e` wraps the browser's event object; `e.Get(\"key\").String()` reads a key name. `PreventDefault` stops a form from submitting or a link from navigating."),
		),
		doc.H("ValueTo"),
		doc.P("`c.ValueTo(&field)` returns a handler that reads the source element's `value` property and stores it in `field`. The pointer can be to a `string`, a signed or unsigned integer, a `float32` or `float64`, or a `bool`. Numbers and booleans are parsed with `strconv`, and a value that does not parse stores the zero value. Any other target stores nothing and logs an error. A checkbox's `value` is not its checked state; read `ctx.JSSrc().Get(\"checked\")` instead."),
		doc.H("Event options"),
		doc.Go(`func EventScope(v ...any) EventOption
func PassiveEvent() EventOption`),
		doc.P("When an element renders again, its handler is replaced only if it changed. Two handlers are the same when they have the same event, the same scope, the same passive flag and the same function code. Every closure made by one function literal has the same code, so in a list, a closure that captures its item is not replaced when the list changes, and keeps the item it captured first. `EventScope(key)` makes the scope part of the comparison: pass a stable key of the item, and the handler is replaced whenever the item at that place changes. The live example under ", doc.A("#conditions-and-lists", "Conditions and lists"), " does this."),
		doc.P("`PassiveEvent()` registers the listener as passive, which lets the browser scroll without waiting for it. Use it for `scroll`, `wheel` and `touchmove` handlers that never call `PreventDefault`."),
	}
}
