package components

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/components/demos"
)

func conditions() []app.UI {
	return []app.UI{
		doc.P("The list below is one component. Typing goes through `ValueTo`, `Add` and `Remove` are click handlers that change the component's fields, `If`/`Else` switches to a message when the list is empty, and `Range` renders one `Li` per item. Each `Remove` handler is a closure scoped by the item's ID, so removing the first item does not leave the second button removing a stale item."),
		doc.Example(demos.ListExample),
		doc.H("If, ElseIf, Else"),
		doc.Go(`func If(expr bool, elem func() UI) Condition
func IfSlice(expr bool, elems func() []UI) Condition

type Condition interface {
	UI
	ElseIf(expr bool, elem func() UI) Condition
	ElseIfSlice(expr bool, elems func() []UI) Condition
	Else(elem func() UI) Condition
	ElseSlice(elems func() []UI) Condition
}`),
		doc.P("A `Condition` is placed in a `Body` like any child. Only the function of the first true branch is called, so a branch can safely read a value that is only valid when its condition holds. If no branch matches and there is no `Else`, the condition renders nothing. The `Slice` forms return several elements from one branch."),
		doc.Go(`app.If(score > 7, func() app.UI {
	return app.Strong().Text("Good")
}).ElseIf(score < 4, func() app.UI {
	return app.Em().Text("Bad")
}).Else(func() app.UI {
	return app.Span().Text("So so")
})`),
		doc.H("Range"),
		doc.Go(`func Range(src any) RangeLoop

type RangeLoop interface {
	UI
	Slice(f func(int) UI) RangeLoop
	Map(f func(string) UI) RangeLoop
}`),
		doc.List(
			doc.Item("`Slice(f)` calls `f` with each index of a slice or array. Any other source panics."),
			doc.Item("`Map(f)` calls `f` with each key of a map, in sorted order. The keys must be strings; any other source, or a map with other keys, panics."),
			doc.Item("`f` may return `nil`, or an `If` or another `Range`; they are flattened into the parent's children."),
		),
		doc.Go(`langs := map[string]int{"Go": 10, "C": 8, "Python": 6}

app.Ul().Body(
	app.Range(langs).Map(func(k string) app.UI {
		return app.Li().Textf("%s: %d/10", k, langs[k])
	}),
)`),
		doc.P("Children are matched by position, not by key. Removing the first item of a list updates every following `Li` with its new content and removes the last one. Element content and attributes follow the data, but handlers made by one closure and child components with unexported state do not, which is what `EventScope` is for, see ", doc.A("#events", "Events"), "."),
	}
}

func lifecycle() []app.UI {
	return []app.UI{
		doc.P("A component opts into lifecycle events by implementing one-method interfaces. Each method is called on the UI goroutine. Every hook that takes a `Context`, except `OnUpdate`, is run through `ctx.Dispatch`, so the component renders again after it returns."),
		lifecycleDiagram(),
		doc.Go(`type Initializer interface{ OnInit() }
type PreRenderer interface{ OnPreRender(Context) }
type Mounter interface{ OnMount(Context) }
type Navigator interface{ OnNav(Context) }
type Updater interface{ OnUpdate(Context) }
type Resizer interface{ OnResize(Context) }
type AppUpdater interface{ OnAppUpdate(Context) }
type AppInstaller interface{ OnAppInstallChange(Context) }
type Dismounter interface{ OnDismount() }
type DismountEnforcer interface{ CompoID() string }`),
		doc.List(
			doc.Item("`OnInit` runs once, before the first `Render`, on the server and in the browser. It has no context; use it to set defaults."),
			doc.Item("`OnPreRender` runs on the server only, while the page is prerendered. Use it to set the title and description of the served HTML through `ctx.Page()`."),
			doc.Item("`OnMount` runs in the browser only, once the component is in the DOM. It is the place to read state, start observers and start timers."),
			doc.Item("`OnNav` runs on every navigation, including the first load, back and forward, and a change of fragment only, in every mounted component that implements it. It also runs on the server while a page is prerendered. See ", doc.A("/routing#fragments-and-onnav", "Fragments and OnNav"), "."),
			doc.Item("`OnUpdate` runs when a parent's render changed one of the component's exported fields, after the fields are copied and before the component renders."),
			doc.Item("`OnResize` runs after the window is resized (debounced by 250 ms) and when an ancestor calls `ctx.ResizeContent()`."),
			doc.Item("`OnAppUpdate` runs when a new version of the app has been downloaded, and `ctx.AppUpdateAvailable()` is then true. WebGo reloads the page by itself when the new version takes over, so this is only useful with `Handler.DisableAutoUpdate` set, to show an update button that calls `ctx.Reload()`."),
			doc.Item("`OnAppInstallChange` runs when the app becomes installable or stops being so; `ctx.IsAppInstallable()` and `ctx.ShowAppInstallPrompt()` go with it."),
			doc.Item("`OnDismount` runs once the component has been removed from the tree, when `Mounted()` is already false. It takes no context. Stop what `OnMount` started here."),
			doc.Item("`CompoID` is not a hook. When a parent renders a component of the same type whose `CompoID` differs from the mounted one, the mounted one is dismounted and the new one mounted, instead of copying exported fields."),
		),
		doc.Go(`type clock struct {
	app.Compo
	now time.Time
}

func (c *clock) OnMount(ctx app.Context) {
	c.now = time.Now()
	c.tick(ctx)
}

func (c *clock) tick(ctx app.Context) {
	ctx.After(time.Second, func(ctx app.Context) {
		c.now = time.Now()
		c.tick(ctx)
	})
}

func (c *clock) Render() app.UI {
	return app.If(c.now.IsZero(), func() app.UI {
		return app.Span().Text("--:--:--")
	}).Else(func() app.UI {
		return app.Time().Text(c.now.Format("15:04:05"))
	})
}`),
		doc.P("The server never calls `OnMount`, so it prerenders `--:--:--`. In the browser the clock ticks once a second. It stops by itself when the component is dismounted, because `After` skips its function once the component is gone, and the loop is not rescheduled."),
		doc.P("WebGo's `page.Page` and `page.Base` implement `OnMount`, `OnNav` and `OnDismount` for you and hand them to screens; see ", doc.A("/state#base", "Base"), "."),
	}
}

func updating() []app.UI {
	return []app.UI{
		doc.P("A component renders again when it is flagged for update. The engine collects the flags and, on the next frame, renders every flagged component that is still mounted, parents before children. Several flags in one frame give one render."),
		updatingDiagram(),
		doc.P("A component is flagged, together with every component above it, when:"),
		doc.List(
			doc.Item("an event handler on one of its elements returns;"),
			doc.Item("a lifecycle hook that takes a `Context` returns;"),
			doc.Item("a function passed to `ctx.Dispatch` or `ctx.After` returns;"),
			doc.Item("an action handler registered with `ctx.Handle` runs, or a state the component observes with `ctx.ObserveState` is set;"),
			doc.Item("`ctx.Update()` is called."),
		),
		doc.P("A parent's render also updates a child in place when it changes the child's exported fields, as described in ", doc.A("#writing-a-component", "Writing a component"), "."),
		doc.Go(`func (ctx Context) Dispatch(fn func(Context))
func (ctx Context) Async(fn func())
func (ctx Context) After(d time.Duration, fn func(Context))
func (ctx Context) Defer(fn func(Context))
func (ctx Context) Update()
func (ctx Context) PreventUpdate()`),
		doc.List(
			doc.Item("`Dispatch` runs `fn` on the UI goroutine and flags the component. It is safe to call from any goroutine. If the component is no longer mounted, `fn` is not called."),
			doc.Item("`Async` runs `fn` on a new goroutine that the engine tracks; the server waits for it before sending a prerendered page. `fn` must not touch the component's fields; it calls `ctx.Dispatch` to do so."),
			doc.Item("`After` waits `d` on a goroutine, then dispatches `fn`."),
			doc.Item("`Defer` runs `fn` on the UI goroutine after the next frame's renders, when the DOM is up to date, for example to focus an element that was just rendered. It does not flag an update."),
			doc.Item("`Update` flags the component, as `Dispatch` with no function does."),
			doc.Item("`PreventUpdate` cancels one pending flag on the component and its ancestors, for a handler that changed nothing visible."),
		),
		doc.Go(`type quote struct {
	app.Compo
	text string
}

func (q *quote) OnMount(ctx app.Context) {
	ctx.Async(func() {
		text := fetchQuote()
		ctx.Dispatch(func(ctx app.Context) {
			q.text = text
		})
	})
}`),
		doc.Note("There is no `Update` method on `app.Compo`. Writing a field from another goroutine without `Dispatch` is a data race, and nothing renders it until something else flags the component."),
		doc.P("For state shared by several components, WebGo's `data.ViewModel` wraps the state store with typed reads, writes and observers; see ", doc.A("/state#model", "Model"), ". The store itself is on ", doc.A("/actions#state", "State and actions"), "."),
	}
}

func raw() []app.UI {
	return []app.UI{
		doc.H("Raw HTML"),
		doc.Go(`func Raw(v string) UI`),
		doc.P("`Raw` inserts HTML that is not built from elements. The string is trimmed and used as is: it is not escaped or checked, so never pass it text that came from a user without sanitizing it."),
		doc.List(
			doc.Item("It must have exactly one root element. If the trimmed string does not start with an opening tag, it is replaced by an empty `<div></div>`."),
			doc.Item("In the browser, only that first root is kept; anything after it is dropped."),
			doc.Item("A `<script>` inside does not run."),
			doc.Item("When the string changes, the whole node is replaced; when it is the same, nothing happens."),
		),
		doc.Go(`app.Raw("<svg width='24' height='24' viewBox='0 0 24 24'><circle cx='12' cy='12' r='10' fill='currentColor'/></svg>")`),
		doc.H("Custom elements"),
		doc.Go(`func Elem(tag string) HTMLElem
func ElemSelfClosing(tag string) HTMLElemSelfClosing`),
		doc.P("`Elem` builds an element with any tag name, for web components or tags with no constructor. It has the global attributes, `Attr`, `Body`, the `Styler` methods and every global event, plus `On(name, h)` for custom events. `ElemSelfClosing` is the same without `Body`."),
		doc.Go(`app.Elem("sl-tooltip").Attr("content", "Saved").Body(
	app.Button().Text("Save"),
).On("sl-show", c.onShow)`),
		doc.H("SVG"),
		doc.P("SVG elements must be created in the SVG namespace. `HTMLElem.XMLNS(ns)` sets it, and it is not inherited, so every element of the drawing needs it. Use this when the drawing changes with the component's state; for a fixed icon, `Raw` is simpler."),
		doc.Go(`const svgNS = "http://www.w3.org/2000/svg"

func gauge(percent int) app.UI {
	return app.Elem("svg").XMLNS(svgNS).Attr("viewBox", "0 0 100 10").Body(
		app.Elem("rect").XMLNS(svgNS).Attr("width", percent).Attr("height", 10),
	)
}`),
	}
}
