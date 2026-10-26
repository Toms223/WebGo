package browser

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/browser/demos"
)

const (
	Title = "Concurrency and the browser"
	Route = "/browser"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "browser",
		Path:  Route,
		Title: Title,
		Lead:  "Keep component state on the UI goroutine, and reach JavaScript, storage and notifications from Go.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "The UI goroutine", Body: uiGoroutine},
			{Title: "Async work", Body: async},
			{Title: "JavaScript values", Body: values},
			{Title: "Calling JavaScript", Body: calling},
			{Title: "Browser storage", Body: storage},
			{Title: "Notifications", Body: notifications},
			{Title: "Environment and logging", Body: environment},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("This page covers how an app runs in the browser: the one goroutine that owns the UI and how to work around it, the `Value` type that reaches JavaScript, local and session storage, notifications, environment variables and logging. Every API here compiles on the server too, where most of it does nothing."),
		doc.P("States and actions, which also cross goroutines, are on ", doc.A("/actions", "State and actions"), ". Testing code that uses these APIs is on ", doc.A("/testing", "Testing"), "."),
	}
}

func uiGoroutine() []app.UI {
	return []app.UI{
		doc.P("An app has one UI goroutine, an event loop. Rendering, lifecycle hooks, event handlers, component action handlers, state deliveries to observers, and the functions given to `Dispatch` and `Defer` all run on it, one at a time. Read and write component fields only there. Code on any other goroutine hands its results over with `Dispatch`."),
		goroutinesDiagram(),
		doc.Go(`func (ctx Context) Dispatch(fn func(Context))
func (ctx Context) Defer(fn func(Context))
func (ctx Context) Update()
func (ctx Context) PreventUpdate()`),
		doc.List(
			doc.Item("`Dispatch` queues `fn` on the UI goroutine and flags the component that `ctx` belongs to, and each of its ancestors, for a re-render. It can be called from any goroutine. If the component is no longer mounted when `fn`'s turn comes, `fn` is skipped."),
			doc.Item("Re-renders are batched. Flagged components are updated on the next frame, and in the browser frames run at most 120 times a second, only while there is work."),
			doc.Item("`Defer` queues `fn` to run after the next frame's updates, when the page shows them. It flags nothing for update, and it is skipped if the component is no longer mounted. `ctx.ScrollTo` uses it."),
			doc.Item("`Update` is `Dispatch(nil)`: it only flags the component and its ancestors."),
			doc.Item("`PreventUpdate` withdraws one pending flag from the component and each of its ancestors. An event handler flags its component once, so calling `PreventUpdate` in the handler skips that re-render."),
		),
		doc.P("Event handlers run through `Dispatch`, which is why a click handler can set a field and the component re-renders without any other call."),
	}
}

func async() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) Async(fn func())
func (ctx Context) After(d time.Duration, fn func(Context))`),
		doc.List(
			doc.Item("`Async` runs `fn` on a new goroutine that the engine tracks. Use it for anything that blocks: network calls, timers, waiting on a channel."),
			doc.Item("`fn` must not touch component fields. It hands results to the UI goroutine with `ctx.Dispatch`."),
			doc.Item("Tracking matters on the server and in tests. Prerendering a page waits for every tracked goroutine before it writes the HTML, and `TestEngine.ConsumeAll` waits for them too. A goroutine started with a plain `go` statement is not waited for. A tracked goroutine that never returns blocks the prerender or the test forever."),
			doc.Item("`After(d, fn)` is `Async` that sleeps for `d` and then calls `Dispatch(fn)`. `fn` is skipped if the component was dismounted in the meantime."),
			doc.Item("The `context.Context` inside `Context` is not canceled when a component is dismounted. In the browser it is never canceled, and on the server it is the request's context. A long-running goroutine has to be stopped by your own code, usually from `OnDismount`."),
			doc.Item("`Async` keeps the UI goroutine free, but it does not run Go in parallel with the page. In the browser, Go runs on the page's single thread, so a long computation that never blocks still freezes the page."),
		),
		doc.Go(`type quote struct {
	app.Compo
	text string
}

func (q *quote) OnMount(ctx app.Context) {
	ctx.Async(func() {
		res, err := http.Get("/api/quote")
		if err != nil {
			app.Log(err)
			return
		}
		defer res.Body.Close()
		b, err := io.ReadAll(res.Body)
		if err != nil {
			app.Log(err)
			return
		}
		ctx.Dispatch(func(ctx app.Context) {
			q.text = string(b)
		})
	})
}

func (q *quote) Render() app.UI {
	return app.Blockquote().Text(q.text)
}`),
		doc.P("The request runs off the UI goroutine. The field is only set inside `Dispatch`, which also re-renders `quote`."),
		doc.H("Example: a timer"),
		doc.Example(demos.TimerExample),
		doc.P("`toggle` runs on the UI goroutine. Start makes a channel and one `Async` goroutine that ticks every 100 ms and dispatches an increment, so `tenths` is only changed on the UI goroutine. Stop closes the channel and the goroutine returns. `OnDismount` closes it too, because nothing else would stop the goroutine when the timer leaves the page: its dispatches would be skipped, but it would tick forever. A tick already queued when Stop is clicked still counts."),
	}
}
