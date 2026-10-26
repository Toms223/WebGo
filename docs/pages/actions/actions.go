package actions

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/actions/demos"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "State and actions"
	Route = "/actions"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "actions",
		Path:  Route,
		Title: Title,
		Lead:  "Share values between components with named states, and send named events with actions.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "State", Body: state},
			{Title: "Observing state", Body: observing},
			{Title: "Persisting and broadcasting", Body: persisting},
			{Title: "Actions", Body: actions},
			{Title: "Handling actions", Body: handling},
			{Title: "Choosing between state, actions and ViewModel", Body: choosing},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("This page covers the two ways components in the `app` package talk to each other without holding references: states, which are named values any component can read and observe, and actions, which are named events carrying a value. This is go-app's state and action system, now part of WebGo's `app` package."),
		doc.P("WebGo builds on both. `data.ViewModel`, on ", doc.A("/state", "Pages, layouts and state"), ", is a typed wrapper over one state, and dynamic styles, on ", doc.A("/applying#actions", "Applying"), ", react to actions. The UI goroutine, `Dispatch` and `Async` are on ", doc.A("/browser", "Concurrency and the browser"), "."),
	}
}

func state() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) SetState(state string, v any) State
func (ctx Context) GetState(state string, recv any)
func (ctx Context) DelState(state string)`),
		doc.P("A state is a value stored under a string key. There is one store per running app, and any component can read or write any key. The store is safe for concurrent use, so these methods can be called from any goroutine."),
		doc.H("SetState"),
		doc.List(
			doc.Item("`SetState` stores `v` under the key, replacing the previous value and its expiry."),
			doc.Item("For every component observing the key, it queues a delivery on the UI goroutine: the value is copied into the observer's receiver, its `OnChange` runs, and the component re-renders. Observers do not see the value during the `SetState` call, only when the UI goroutine gets to it."),
			doc.Item("It returns a `State`, whose methods persist, expire or broadcast the value just set. They are under ", doc.A("#persisting-and-broadcasting", "Persisting and broadcasting"), "."),
			doc.Item("A pointer is stored as the pointer. Readers get a copy of what it points to at the time they read, so changing the pointed-to value after `SetState` changes what later readers get without notifying observers. Call `SetState` again after a change."),
		),
		doc.H("GetState"),
		doc.List(
			doc.Item("`GetState` copies the current value into `recv`, which must be a pointer. If the stored value is a pointer, the value it points to is copied."),
			doc.Item("The type `recv` points to must be exactly the stored type. A state set with `SetState(\"n\", 1)` can be read into an `*int` but not an `*int64`. On a mismatch, or when `recv` is not a pointer, `GetState` logs the error through `app.Log` and leaves `recv` unchanged."),
			doc.Item("If the key was not set since the app started, `GetState` looks for a persisted copy in local storage and decodes it into `recv`. If there is none, `recv` is left as it was, so give it its default before the call."),
			doc.Item("If the state has expired, it is deleted from memory and local storage, and `recv` is left as it was."),
			doc.Item("A state stored as `nil` sets `*recv` to its zero value."),
		),
		doc.H("DelState"),
		doc.List(
			doc.Item("`DelState` removes the value from memory and from local storage."),
			doc.Item("It does not notify observers and does not remove them. They keep their last value until the next `SetState` on the key. To reset what observers show, set a zero value instead."),
		),
		doc.Go(`type greeting struct {
	app.Compo
	name string
}

func (g *greeting) OnMount(ctx app.Context) {
	g.name = "guest"
	ctx.GetState("user-name", &g.name)
}

func (g *greeting) Render() app.UI {
	return app.Div().Body(
		app.P().Text("Hello, " + g.name),
		app.Input().Value(g.name).OnChange(g.rename),
	)
}

func (g *greeting) rename(ctx app.Context, e app.Event) {
	ctx.SetState("user-name", ctx.JSSrc().Get("value").String())
}`),
		doc.P("This component reads the name once, when it mounts, so it does not see later changes made by other components. ", doc.A("#observing-state", "Observing state"), " covers following a key."),
	}
}

func observing() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) ObserveState(state string, recv any) Observer

func (o Observer) While(condition func() bool) Observer
func (o Observer) OnChange(h func()) Observer
func (o Observer) WithBroadcast() Observer`),
		doc.P("`ObserveState` reads the current value into `recv` at once, with the rules of `GetState`, and then keeps `recv` up to date: each later `SetState` on the key copies the new value into `recv` on the UI goroutine and updates the component. Call it in `OnMount`, with a pointer to a field of the component."),
		doc.List(
			doc.Item("The observer belongs to the source of `ctx`, which in `OnMount` is the component. There is one observer per key and component. Calling `ObserveState` again for the same key from the same component replaces the earlier observer, with its `While` and `OnChange`."),
			doc.Item("Observation ends when the component is dismounted, or when the `While` condition returns false. Both are checked when a new value is delivered and after each frame of updates, not on a timer. The observer is then removed, and `recv` stops changing."),
			doc.Item("`OnChange(h)` runs `h` on the UI goroutine each time a new value has been copied into `recv`. It does not run for the first read."),
			doc.Item("`While(condition)` runs `condition` on the UI goroutine before each delivery."),
			doc.Item("`WithBroadcast` also accepts values sent from other tabs. It is under ", doc.A("#persisting-and-broadcasting", "Persisting and broadcasting"), "."),
			doc.Item("`OnMount` runs only in the browser. During server prerendering and in tests, a component that observes in `OnMount` renders the values its fields start with."),
		),
		doc.Example(demos.VotesExample),
		doc.P("`voter` and `tally` are separate components, and neither holds a reference to the other. `voter` reads the count with `GetState` and writes it with `SetState`. `tally` observes the key, so its `votes` field changes and it re-renders on every vote. Reset sets `0` rather than calling `DelState`, because `DelState` does not reach observers. The count lives in memory only, so reloading the page starts again from zero."),
		doc.H("While and OnChange"),
		doc.Go(`type recent struct {
	app.Compo
	query string
	seen  []string
}

func (r *recent) OnMount(ctx app.Context) {
	ctx.ObserveState("query", &r.query).
		While(func() bool { return len(r.seen) < 10 }).
		OnChange(func() { r.seen = append(r.seen, r.query) })
}`),
		doc.P("`recent` records each new query until it has ten, then stops observing. Both callbacks run on the UI goroutine, so they can read and change fields."),
	}
}

func persisting() []app.UI {
	return []app.UI{
		doc.Go(`func (s State) ExpiresIn(d time.Duration) State
func (s State) ExpiresAt(t time.Time) State
func (s State) Persist() State
func (s State) PersistWithEncryption() State
func (s State) Broadcast() State`),
		doc.P("These methods are chained on the `State` that `SetState` returns. Each one acts once, when it is called, on the value that `SetState` just stored."),
		persistingDiagram(),
		doc.Go(`ctx.SetState("cart", cart).
	ExpiresIn(24 * time.Hour).
	Persist().
	Broadcast()`),
		doc.H("Persist"),
		doc.List(
			doc.Item("`Persist` encodes the value as JSON and writes it to local storage under the state key, wrapped with its expiry. Only what `encoding/json` encodes is kept, so unexported fields are lost."),
			doc.Item("If encoding or writing fails, `Persist` logs the error. The value in memory is still set."),
			doc.Item("After a reload, the first `GetState` or `ObserveState` on a key that is not in memory decodes the persisted copy into its receiver."),
			doc.Item("Persisting happens per call. A later `SetState` without `Persist` changes memory only, and the next page load sees the older persisted value."),
			doc.Item("State keys and keys written through `ctx.LocalStorage()` share one namespace. Do not use the same key for both."),
			doc.Item("On the server and in tests, local storage is an in-memory store that belongs to one engine, so persisted values do not outlive a request or a test."),
		),
		doc.H("Encryption"),
		doc.P("`PersistWithEncryption` encrypts the JSON with AES-GCM before writing it. The key is derived from `ctx.DeviceID()`, a random identifier stored in plain text in the same local storage, under `/go-app/deviceID`. This keeps a value from being readable at a glance, but anyone who can read local storage can decrypt it, so it does not protect secrets. `DeviceID` panics if local storage cannot be read or written."),
		doc.H("Expiry"),
		doc.List(
			doc.Item("`ExpiresIn(d)` is `ExpiresAt(time.Now().Add(d))`."),
			doc.Item("An expired state is deleted on the next read, which leaves the receiver unchanged. Values already queued for observers are not delivered once the state has expired."),
			doc.Item("`Persist` stores the expiry the state has at the time of the call. Call `ExpiresIn` or `ExpiresAt` before `Persist`, as in the example above. In the other order, the persisted copy never expires."),
			doc.Item("When the app starts in the browser, expired persisted states are removed from local storage."),
		),
		doc.H("Broadcast"),
		doc.List(
			doc.Item("`Broadcast` sends the value, encoded as JSON, to the other tabs and windows of the same origin through a `BroadcastChannel`. The sending tab does not receive its own message."),
			doc.Item("A receiving tab delivers it only to observers chained with `WithBroadcast()`. Others ignore it. The value is decoded into the receiver and `OnChange` runs, but the receiving tab's stored value does not change, so `GetState` there still returns the old one."),
			doc.Item("Where `BroadcastChannel` does not exist, including on the server, `Broadcast` logs that it is not supported and sends nothing."),
			doc.Item("An open `BroadcastChannel` makes the page ineligible for the browser's back/forward cache, in the tab that sends and in the tabs that listen."),
		),
		doc.Go(`func (c *cartBadge) OnMount(ctx app.Context) {
	ctx.ObserveState("cart", &c.cart).WithBroadcast()
}`),
	}
}
