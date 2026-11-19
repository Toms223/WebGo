package state

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/state/demos"
)

func model() []app.UI {
	return []app.UI{
		doc.P("A `Model[T]` binds a state value of type `T` to a key in the app's state store, described under ", doc.A("/actions#state", "State and actions"), "."),
		modelDiagram(),
		doc.P("The counter below is one `Model` driving a component that observes it with `Get(ctx).Observe(ctx)` in `OnMount`. `+1` and `Reset` go through `Load(ctx).Apply(...).Set()`, so the count changes on screen but is not stored; `Save` goes through `SetAndPersist`, so the count survives a reload. Each piece is described below."),
		doc.Example(demos.CounterExample),
		doc.Go(`type Model[T any] struct {
	// unexported fields
}`),
		doc.H("NewModel"),
		doc.Go(`func NewModel[T any](key string, value *T) (*Model[T], error)`),
		doc.List(
			doc.Item("If `value` is `nil`, it returns an error (`\"value is nil\"`)."),
			doc.Item("If `key` is `\"\"`, it returns an error (`\"key is empty\"`)."),
			doc.Item("Otherwise, it returns a `*Model[T]` that holds `value` itself, without a copy, so changes through the model are visible through the original pointer, and the other way round, until `Apply` replaces the pointer."),
		),
		doc.P("`NewModel` does not read or write the state store. The value passed in is used as it is until `Get` or `Load` is called."),
		doc.H("Get"),
		doc.Go(`func (m *Model[T]) Get(ctx app.Context) *Model[T]`),
		doc.P("`Get` reads the stored value for the model's key into the model's state, and returns the model. If the key has never been written, or the stored value is of another type, the state is left unchanged."),
		doc.H("State"),
		doc.Go(`func (m *Model[T]) State() *T`),
		doc.P("`State` returns the model's in-memory state pointer. It does not touch the store; call `Get` or `Load` first to refresh it."),
		doc.H("Load"),
		doc.Go(`func (m *Model[T]) Load(ctx app.Context) *Change[T]`),
		doc.P("`Load` refreshes the state from the store, as `Get` does, then returns a `*Change[T]` bound to `ctx` and the model, ready for `Apply`, `Set` or `SetAndPersist`."),
		doc.H("Observe"),
		doc.Go(`func (m *Model[T]) Observe(ctx app.Context) app.Observer`),
		doc.P("`Observe` returns an `app.Observer` of the model's key. It fires on `Set` and `SetAndPersist`, not on a direct write to the state. The observer follows the model, not the pointer it held when it started, so it keeps working after `Apply` returns a new pointer."),
		doc.H("Actions"),
		doc.P("Every `Set` or `SetAndPersist` through the model fires an app action named after the model's key, whose value is the model's `*T`. The state store posts it, so there is nothing to register. A direct write to the state fires nothing."),
		doc.P("Use the key as the action name in `style.When` and `style.Compute` and the page's CSS follows the state; the ", doc.A("/applying#models-and-actions", "Applying"), " page shows it. Use it with `ctx.Handle` to react in a component."),
		doc.H("Change.Apply"),
		doc.Go(`type Change[T any] struct {
	// unexported fields
}

func (c *Change[T]) Apply(verb func(ctx app.Context, state *T) *T) *Change[T]`),
		doc.P("`Apply` calls `verb` with the change's context and the model's current state, and returns the change. If `verb` returns a non-nil pointer, the model's state is replaced with it; if it returns `nil`, the state is left as it is."),
		doc.P("The result may be a fresh pointer, such as `&T{...}`: observers and actions follow the model, not the pointer."),
		doc.P("Calls to `Apply` chain: each `verb` sees the state the previous one left. `Apply` changes memory only. Nothing reaches the store until `Set` or `SetAndPersist`."),
		doc.H("Set"),
		doc.Go(`func (c *Change[T]) Set() *Model[T]`),
		doc.P("`Set` writes the model's current state, as `Apply` left it, to the store, which notifies observers and fires the model's action. It returns the model. It does not write to local storage."),
		doc.H("SetAndPersist"),
		doc.Go(`func (c *Change[T]) SetAndPersist(value *T) *Model[T]`),
		doc.P("`SetAndPersist` replaces the model's state with `value`, discarding earlier `Apply` calls, writes it to the store, and persists a snapshot to local storage. It returns the model."),
		doc.P("The snapshot is fixed when `SetAndPersist` runs. Later direct changes to the state do not change what was persisted."),
	}
}

func errorsSection() []app.UI {
	return []app.UI{
		doc.P("The `app` package exports four sentinel errors for screens and shells:"),
		doc.List(
			doc.Item("`ErrNoLayout`: a `Screen` or `Shell` has no layout."),
			doc.Item("`ErrNoModel`: a `Screen` or `Shell` has no model."),
			doc.Item("`ErrNoScreens`: `NewShell` was called with a `nil` screens slice."),
			doc.Item("`ErrNoRoute`: no screen of a `Shell` matches the current route."),
		),
		doc.P("Check for one with `errors.Is`, for example `errors.Is(s.Err(), app.ErrNoModel)`."),
	}
}

func fetchingExample() []app.UI {
	return []app.UI{
		doc.P("A screen that fetches its data before it draws, with `WithLoader`:"),
		doc.Go(`type Profile struct {
	Name string
}

func fetchProfile(ctx app.Context, p *Profile) error {
	res, err := http.Get("/api/profile")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(p)
}

var profile, _ = app.NewModel("profile", &Profile{})

func profileScreen() *app.Screen[Profile] {
	return app.NewScreen[Profile]("/profile").
		WithLayout(&profileLayout{}).
		WithModel(profile).
		WithLoader(fetchProfile)
}`),
		doc.P("`profileScreen` builds a new screen on each call, so it can be called inside the `build` function passed to ", doc.A("#serve", "Serve"), ", which must return new screens every time."),
		doc.List(
			doc.Item("`fetchProfile` runs on its own goroutine, so a blocking call does not freeze the page. It fills a copy of the state."),
			doc.Item("While it runs, the screen shows its layout's `Loading()`. The rest of the shell keeps rendering."),
			doc.Item("When it returns `nil`, the copy becomes the model's state through `Set`, and the layout's `Draw` reads it with `m.State()`. When it returns an error, the layout's `Error(err)` shows."),
		),
		doc.P("In the demo below, a 600 ms `time.Sleep` stands in for the fetch. The screen sits inside a plain component rather than a `Shell`. Load again gives it a new route, so its `CompoID` changes, it is mounted fresh, and the loader runs again."),
		doc.Example(demos.LoaderExample),
		doc.H("Loading side by side"),
		doc.P("The load phase is not limited to screens. Any component that implements `app.Loader` is held until its own `OnLoad` returns, and the hold is per component. Below, two parts load for 300 ms and 900 ms next to each other: each shows its own `Loading()` and draws when its own load is done, without waiting for the other. Load again changes an exported field of both, so each loads again after its update."),
		doc.Example(demos.SiblingsExample),
	}
}
