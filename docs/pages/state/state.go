package state

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Pages, layouts and state"
	Route = "/state"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "state",
		Path:  Route,
		Title: Title,
		Lead:  "Screens are a Layout and a Model, hosted by a Shell that switches them by route.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Screen", Body: screenSection},
			{Title: "Shell", Body: shellSection},
			{Title: "Layout and Routed", Body: interfaces},
			{Title: "Model", Body: model},
			{Title: "Errors", Body: errorsSection},
			{Title: "Example: fetching and displaying state", Body: fetchingExample},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("A `Screen` draws a `Model` through a `Layout` and can load its data before it draws. A `Shell` is the root of the app: it holds every screen under one shared layout and model, and draws the screen whose route matches the URL. A `Model` holds the typed state screens read and write."),
		doc.P("Since `v1.0.0`, all of them live in the `app` package, next to `Compo`, and the engine routes to them. Up to `v0.1.0` they lived in the `page` and `data` packages, which are deleted."),
	}
}

func screenSection() []app.UI {
	return []app.UI{
		doc.Go(`type Screen[T any] struct {
	app.Compo
	Path string
	// unexported fields
}

func NewScreen[T any](path string) *Screen[T]`),
		doc.P("A `Screen` is a component served at a route. `NewScreen` creates one for the given path. It has no layout and no model yet, so it is not usable until both are set."),
		doc.H("Building a screen"),
		doc.P("A screen is built by chaining calls. Each call returns the same `*Screen[T]`:"),
		doc.Go(`package main

import "github.com/Toms223/WebGo/app"

type Settings struct {
	User string
}

type settingsLayout struct{}

func (settingsLayout) Loading() app.UI {
	return app.Text("loading")
}

func (settingsLayout) Error(err error) app.UI {
	return app.Text(err.Error())
}

func (settingsLayout) Draw(m *app.Model[Settings]) app.HTMLDiv {
	return app.Div().Text(m.State().User)
}

func (l settingsLayout) Body(...app.UI) app.Layout[Settings] {
	return l
}

func main() {
	model, err := app.NewModel("settings", &Settings{})
	if err != nil {
		panic(err)
	}

	s := app.NewScreen[Settings]("/settings").
		WithLayout(settingsLayout{}).
		WithModel(model)

	if s.Err() != nil {
		panic(s.Err())
	}
}`),
		doc.List(
			doc.Item("`At(path string) *Screen[T]` sets the screen's path."),
			doc.Item("`WithLayout(layout Layout[T]) *Screen[T]` sets the layout. Passing `nil` latches `ErrNoLayout`."),
			doc.Item("`WithModel(m *Model[T]) *Screen[T]` sets the model. Passing `nil` latches `ErrNoModel`."),
			doc.Item("`WithLoader(loader func(ctx Context, state *T) error) *Screen[T]` sets the function `OnLoad` runs. See ", doc.A("#onload", "OnLoad"), " below."),
		),
		doc.H("Error latching"),
		doc.P("A `Screen` keeps a single builder error. The first call that produces one (`WithLayout(nil)` or `WithModel(nil)`) stores it. From then on every builder call (`At`, `WithLayout`, `WithModel`, `WithLoader`) returns at once without changing anything. Only the first failure in a chain is kept."),
		doc.H("Err"),
		doc.Go(`func (s *Screen[T]) Err() error`),
		doc.P("`Err` returns the latched error, if there is one. Otherwise it returns `ErrNoLayout` when the layout is `nil`, then `ErrNoModel` when the model is `nil`, and `nil` when both are set."),
		doc.H("Route and CompoID"),
		doc.Go(`func (s *Screen[T]) Route() string
func (s *Screen[T]) CompoID() string`),
		doc.P("`Route` returns the path with exactly one leading slash: `\"\"` becomes `/`, `\"settings\"` becomes `/settings`, and `\"/settings\"` stays `/settings`."),
		doc.P("`CompoID` returns the route. When a parent renders a screen of the same Go type on another route, the mounted one is dismounted and the new one mounted, instead of copying fields across. That is how a `Shell` switches between two screens of the same type cleanly. See ", doc.A("/components#lifecycle", "Lifecycle"), "."),
		doc.H("OnMount"),
		doc.Go(`func (s *Screen[T]) OnMount(ctx app.Context)`),
		doc.P("If `Err()` reports an error, `OnMount` stores it and returns, so `Render` shows it. Otherwise it reads the model's state from the store with `Get` and observes the model while the screen is mounted, so every `Set` on the model renders the screen again."),
		doc.H("OnLoad"),
		doc.Go(`func (s *Screen[T]) OnLoad(ctx app.Context) error`),
		doc.P("A `Screen` is an `app.Loader`: the engine runs `OnLoad` in the browser, on its own goroutine, after `OnMount` and again after each `OnUpdate`, and holds the screen's render until it returns. The load phase is described under ", doc.A("/components#lifecycle", "Lifecycle"), "."),
		doc.List(
			doc.Item("Without a loader, or with a builder error, `OnLoad` returns `nil` at once."),
			doc.Item("Otherwise it makes a shallow copy of the model's state and calls the loader with a pointer to the copy."),
			doc.Item("When the loader returns `nil`, the copy is written back on the UI goroutine through `Load(ctx).Apply(...).Set()`, so observers re-render and the model's action fires. The screen then draws with the new state."),
			doc.Item("When the loader returns an error, the copy is dropped, the model is unchanged, and the screen draws `Error(err)`. The next load, after a remount or an `OnUpdate`, tries again."),
			doc.Item("`OnLoad` never runs on the server."),
		),
		doc.Note("The copy is shallow. Slices and maps inside the state are shared with the model. A loader that assigns a new slice or map is safe; one that appends to or writes into the existing one races with the UI goroutine."),
		doc.P("A screen always implements `Loader`, so even without a loader it shows `Loading()` for the moment its empty load takes to come back through the UI goroutine."),
		doc.H("Loading, Error and Render"),
		doc.Go(`func (s *Screen[T]) Loading() app.UI
func (s *Screen[T]) Error(err error) app.UI
func (s *Screen[T]) Render() app.UI`),
		doc.List(
			doc.Item("While `OnLoad` runs, the engine draws `Loading()`, which is the layout's `Loading()`."),
			doc.Item("After a failed load, it draws `Error(err)`, which is the layout's `Error(err)`."),
			doc.Item("Otherwise it draws `Render()`. `Render` returns the layout's `Error(err)` when `Err()` reports an error, and `layout.Draw(model)` otherwise."),
			doc.Item("All three panic with `ErrNoLayout` when the screen has no layout."),
		),
		doc.P("Only `*Screen[T]` (a pointer) is a component. A screen does not need a `Shell`: it can be rendered as a child of any component, as the ", doc.A("#example-fetching-and-displaying-state", "example"), " below does."),
	}
}
