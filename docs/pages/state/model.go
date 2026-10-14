package state

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/state/demos"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func model() []app.UI {
	return []app.UI{
		doc.P("`ViewModel[T]` lives in the `data` package. A `ViewModel[T]` binds a state value of type `T` to a key in `go-app`'s state store."),
		doc.P("The counter below is one `ViewModel` driving a component that observes it with `Get(ctx).Observe(ctx)` in `OnMount`. `+1` and `Reset` go through `Load(ctx).Apply(...).Set()`, so the count changes on screen but is not stored; `Save` goes through `SetAndPersist`, so the count survives a reload. Each piece is described below."),
		doc.Example(demos.CounterExample),
		doc.Go(`type ViewModel[T any] struct {
	// unexported fields
}`),
		doc.H("NewViewModel"),
		doc.Go(`func NewViewModel[T any](key string, value *T) (*ViewModel[T], error)`),
		doc.List(
			doc.Item("If `value` is `nil`, it returns an error (`\"value is nil\"`)."),
			doc.Item("If `key` is `\"\"`, it returns an error (`\"key is empty\"`)."),
			doc.Item("Otherwise, it returns a `*ViewModel[T]` that holds `value` directly — no copy is made, so later changes made through the returned `ViewModel` are visible through the original pointer, and vice versa."),
		),
		doc.P("`NewViewModel` does not read or write the state store. The value passed in is used as-is until `Get` or `Load` is called."),
		doc.H("Get"),
		doc.Go(`func (v *ViewModel[T]) Get(ctx app.Context) *ViewModel[T]`),
		doc.P("`Get` reads the current value for the model's key from the state store into the model's existing state value, and returns the model itself. If the key has never been written, or the stored value is a different type, `Get` leaves the model's state unchanged."),
		doc.H("State"),
		doc.Go(`func (v *ViewModel[T]) State() *T`),
		doc.P("`State` returns the model's current in-memory state pointer. It does not touch the state store — call `Get` or `Load` first if the state needs to be refreshed from the store."),
		doc.H("Load"),
		doc.Go(`func (v *ViewModel[T]) Load(ctx app.Context) *Action[T]`),
		doc.P("`Load` refreshes the model's state from the store (the same read `Get` performs), then returns a new `*Action[T]` bound to `ctx` and the model, ready for `Apply`, `Set`, or `SetAndPersist`."),
		doc.H("Observe"),
		doc.Go(`func (v *ViewModel[T]) Observe(ctx app.Context) app.Observer`),
		doc.P("`Observe` returns an `app.Observer` that watches the model's key in the state store. It fires when the store's value for that key changes through `Set` or `SetAndPersist` — not on a direct write to the model's in-memory state."),
		doc.H("OnAction"),
		doc.Go(`func (v *ViewModel[T]) OnAction(ctx app.Context) {
	if v.source != nil && v.source.Mounted() {
		return
	}
	v.source = ctx.Src()
	ctx.ObserveState(v.key, v.state).OnChange(func() {
		ctx.NewActionWithValue(v.key, v.state)
	})
}`),
		doc.P("`OnAction` turns state changes into go-app actions. Each `Set` or `SetAndPersist` through the model fires an action named after the model's key, whose value is the model's `*T`. A direct write to the in-memory state fires nothing."),
		doc.List(
			doc.Item("Call it in `OnMount`."),
			doc.Item("Use the same key in `style.When` and `style.Compute`: the key is the action name, and `*T` is the type the binding receives."),
			doc.Item("Call it from lifecycle hooks only, never from `Render`."),
		),
		doc.P("It registers once per model, however many components call it. It registers again only when the component it registered on has dismounted, so a component that mounts again after a navigation calls it again in its own `OnMount`. The ", doc.A("/applying", "Applying"), " page shows it driving dynamic rules."),
		doc.H("Action.Apply"),
		doc.Go(`type Action[T any] struct {
	// unexported fields
}

func (a *Action[T]) Apply(verb func(ctx app.Context, state *T) *T) *Action[T]`),
		doc.P("`Apply` calls `verb` with the action's context and the model's current in-memory state, then returns the action itself. If `verb` returns a non-nil pointer, the model's in-memory state is replaced with that pointer. If `verb` returns `nil`, the model's state is left unchanged."),
		doc.P("Return the same pointer you were given rather than a fresh `&T{...}`: the observer registered at `Page.OnMount` still holds the previous pointer, so returning a new one detaches the model from that observer and the page silently stops updating."),
		doc.P("Calls to `Apply` chain: each call's `verb` sees the state left by the previous call."),
		doc.P("`Apply` only changes the model's in-memory state. It does not publish anything to the state store — call `Set` or `SetAndPersist` for that."),
		doc.H("Set"),
		doc.Go(`func (a *Action[T]) Set() *ViewModel[T]`),
		doc.P("`Set` writes the model's current in-memory state (as left by any `Apply` calls, or as it was when `Load` ran if `Apply` was never called) to the state store, and returns the model. `Set` does not write to local storage."),
		doc.H("SetAndPersist"),
		doc.Go(`func (a *Action[T]) SetAndPersist(value *T) *ViewModel[T]`),
		doc.P("`SetAndPersist` sets the model's in-memory state directly to `value` (ignoring whatever `Apply` may have built up), writes `value` to the state store, and persists a snapshot of `value` to local storage. It returns the model."),
		doc.P("The local storage snapshot is fixed at the moment `SetAndPersist` runs. Later direct changes to the model's in-memory state do not change what was persisted."),
	}
}

func errorsSection() []app.UI {
	return []app.UI{
		doc.P("The `page` package exports four sentinel errors:"),
		doc.List(
			doc.Item("`ErrNoLayout` — a `Page` or `Base` has no layout."),
			doc.Item("`ErrNoModel` — a `Page` or `Base` has no model."),
			doc.Item("`ErrNoScreens` — `NewBase` was called with a `nil` screens slice."),
			doc.Item("`ErrNoRoute` — a `Base`'s current selection does not match any screen's route."),
		),
		doc.P("Check for a specific one with `errors.Is`, for example `errors.Is(p.Err(), page.ErrNoModel)`."),
	}
}

func fetchingExample() []app.UI {
	return []app.UI{
		doc.P("Storing a fetched value and reading it back in a layout's `Draw`:"),
		doc.Go(`package main

import (
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type AppState struct {
	User string
}

func fetchUser() string {
	return "ada"
}

func loadUser(ctx app.Context, vm *data.ViewModel[AppState]) {
	ctx.Async(func() {
		user := fetchUser()
		ctx.Dispatch(func(ctx app.Context) {
			vm.Load(ctx).
				Apply(func(ctx app.Context, state *AppState) *AppState {
					state.User = user
					return state
				}).
				Set()
		})
	})
}

type AppLayout struct {
	body []app.UI
}

func (l *AppLayout) Loading() app.UI {
	return app.Text("loading")
}

func (l *AppLayout) Error(err error) app.UI {
	return app.Text(err.Error())
}

func (l *AppLayout) Body(content ...app.UI) page.Layout[AppState] {
	l.body = content
	return l
}

func (l *AppLayout) Draw(m *data.ViewModel[AppState]) app.HTMLDiv {
	current := m.State()
	return app.Div().Text(current.User)
}

func main() {
	vm, err := data.NewViewModel("app-state", &AppState{})
	if err != nil {
		panic(err)
	}
	layout := &AppLayout{}
	layout.Draw(vm)

	_ = loadUser
}`),
		doc.P("`loadUser` fetches the user, then applies and publishes the new state through the model. `Draw` reads the same model's current state back out with `State()` on every render."),
	}
}
