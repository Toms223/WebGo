package state

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
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
		Lead:  "Screens are a Layout and a ViewModel, hosted by a Base that switches them by route.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Page", Body: pageSection},
			{Title: "Base", Body: baseSection},
			{Title: "Layout and Screen interfaces", Body: interfaces},
			{Title: "Model", Body: model},
			{Title: "Errors", Body: errorsSection},
			{Title: "Example: fetching and displaying state", Body: fetchingExample},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("The `page` package builds screens out of a `Layout` and a `data.ViewModel`, and assembles them under a `Base` component that switches between screens by route. The `data` package holds the state each screen reads and writes."),
		doc.P("This page covers `github.com/Toms223/WebGo` from `v0.0.5` onward. `v0.0.5` renamed the `model` package to `data` (the API is otherwise identical)."),
	}
}

func pageSection() []app.UI {
	return []app.UI{
		doc.Go(`type Page[T any] struct {
	Path string
	// unexported fields
}

func NewPage[T any](path string) *Page[T]`),
		doc.P("`NewPage` creates a page for the given path. It has no layout and no model yet, so it is not usable until both are set."),
		doc.H("Building a page"),
		doc.P("`Page` is built by chaining calls. Each call returns the same `*Page[T]`, so the order of calls does not matter:"),
		doc.Go(`package main

import (
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type AppState struct {
	User string
}

type shellLayout struct {
	body []app.UI
}

func (s *shellLayout) Loading() app.UI {
	return app.Text("loading")
}

func (s *shellLayout) Error(err error) app.UI {
	return app.Text(err.Error())
}

func (s *shellLayout) Draw(m *data.ViewModel[AppState]) app.HTMLDiv {
	return app.Div().Body(s.body...)
}

func (s *shellLayout) Body(content ...app.UI) page.Layout[AppState] {
	s.body = content
	return s
}

func main() {
	vm, err := data.NewViewModel("app-state", &AppState{})
	if err != nil {
		panic(err)
	}
	layout := &shellLayout{}

	p := page.NewPage[AppState]("/settings").
		WithLayout(layout).
		WithModel(vm)

	if p.Err() != nil {
		panic(p.Err())
	}
}`),
		doc.List(
			doc.Item("`At(path string) *Page[T]` sets the page's path."),
			doc.Item("`WithLayout(layout Layout[T]) *Page[T]` sets the page's layout. Passing `nil` sets the page's error to `ErrNoLayout`."),
			doc.Item("`WithModel(m *data.ViewModel[T]) *Page[T]` sets the page's model. Passing `nil` sets the page's error to `ErrNoModel`."),
		),
		doc.H("Error latching"),
		doc.P("A `Page` keeps a single internal error. The first builder call that produces an error (`WithLayout(nil)` or `WithModel(nil)`) stores that error. Once an error is stored, every later builder call (`At`, `WithLayout`, `WithModel`) returns immediately without changing anything on the page — not the path, not the layout, not the model."),
		doc.P("This means only the first failure in a chain is kept. If both `WithLayout(nil)` and `WithModel(nil)` appear in the same chain, the page's error is whichever of the two ran first."),
		doc.H("Err"),
		doc.Go(`func (p *Page[T]) Err() error`),
		doc.P("`Err` returns the page's error, if one was latched by a builder call. If no error was latched, it checks the page's current layout and model:"),
		doc.List(
			doc.Item("If the layout is `nil`, it returns `ErrNoLayout`."),
			doc.Item("Otherwise, if the model is `nil`, it returns `ErrNoModel`."),
			doc.Item("Otherwise, it returns `nil`."),
		),
		doc.H("Route"),
		doc.Go(`func (p *Page[T]) Route() string`),
		doc.P("`Route` returns the page's path with exactly one leading slash, regardless of what was passed to `NewPage` or `At`. A path of `\"\"` becomes `/`, `\"settings\"` becomes `/settings`, and `\"/settings\"` stays `/settings`."),
		doc.H("OnMount, OnDismount, Mounted"),
		doc.Go(`func (p *Page[T]) OnMount(ctx app.Context) Screen
func (p *Page[T]) OnDismount() Screen
func (p *Page[T]) Mounted() bool`),
		doc.P("`OnMount` is called by a `Base` when the page becomes the selected screen. If `Err()` reports an error, `OnMount` stores that error and returns the page without mounting it — it does not panic, even if the layout or model is missing. Otherwise, it reads the model's current state from `ctx`, starts observing the model for further state changes for as long as the page reports itself as mounted, and marks the page mounted."),
		doc.P("`OnDismount` marks the page as not mounted and returns it."),
		doc.P("`Mounted` reports whether the page is currently mounted."),
		doc.H("Content"),
		doc.Go(`func (p *Page[T]) Content() app.UI`),
		doc.P("`Content` renders the page:"),
		doc.List(
			doc.Item("If the page has no layout, `Content` panics with `ErrNoLayout`."),
			doc.Item("Otherwise, if `Err()` reports an error, `Content` returns `layout.Error(err)`."),
			doc.Item("Otherwise, if the page is not mounted, `Content` returns `layout.Loading()`."),
			doc.Item("Otherwise, `Content` returns `layout.Draw(model)`."),
		),
		doc.P("Only `*Page[T]` (a pointer) satisfies the `Screen` interface described below. A `Page[T]` value does not."),
	}
}
