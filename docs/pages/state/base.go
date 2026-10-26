package state

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func baseSection() []app.UI {
	return []app.UI{
		doc.Go(`type Base[U any] struct {
	BaseLayout Layout[U]
	Model      *data.ViewModel[U]
	Screens    []Screen
	// unexported fields
}`),
		doc.P("`Base` is an `app` component that hosts a set of `Screen`s under one shared `Layout` and one shared model, and switches between them by URL route. It implements `app.Mounter` and `app.Navigator`, but it does not itself satisfy `Screen` — it is meant to be the root component, not a screen nested inside another `Base`."),
		baseDiagram(),
		doc.H("NewBase"),
		doc.Go(`func NewBase[U any](layout Layout[U], model *data.ViewModel[U], screens []Screen) (*Base[U], error)`),
		doc.P("`NewBase` requires all three arguments:"),
		doc.List(
			doc.Item("If `layout` is `nil`, it returns `ErrNoLayout`."),
			doc.Item("Otherwise, if `model` is `nil`, it returns `ErrNoModel`."),
			doc.Item("Otherwise, if `screens` is `nil`, it returns `ErrNoScreens`."),
		),
		doc.P("These checks run in that order, so if more than one argument is missing, the error names the first one checked. An empty, non-nil slice (`[]Screen{}`) is accepted — it is not the same as `nil`."),
		doc.P("On success, `NewBase` returns a `*Base[U]` with the first screen (index `0`) selected. Nothing is mounted yet."),
		doc.H("OnMount"),
		doc.Go(`func (b *Base[U]) OnMount(ctx app.Context)`),
		doc.P("If `b.Err()` reports an error (missing layout or model), `OnMount` does nothing. Otherwise it reads the model's current state from `ctx`. If there are no screens, it marks the base mounted and stops. Otherwise it calls `OnMount` on the currently selected screen (index `0` by default) and marks the base mounted. Calling `OnMount` a second time is harmless as long as a screen is currently selected. If `OnNav` previously found no matching route, the selection is negative and a second `OnMount` call panics indexing `Screens` with it."),
		doc.H("OnNav"),
		doc.Go(`func (b *Base[U]) OnNav(ctx app.Context)`),
		doc.P("`OnNav` matches the current browser path (`ctx.Page().URL().Path`) against each screen's `Route()`, in the order screens appear in `Screens`:"),
		doc.List(
			doc.Item("If a screen's route matches the path and that screen is already selected, `OnNav` returns without doing anything else."),
			doc.Item("If a screen's route matches the path and it is not already selected: the previously selected screen (if any is currently selected) is dismounted via `OnDismount`, the matching screen becomes selected, the matching screen is mounted via `OnMount`, and the browser URL is replaced with the matched route."),
			doc.Item("If no screen's route matches the path, the selection is cleared (set to an index that selects nothing)."),
		),
		doc.H("Render"),
		doc.Go(`func (b *Base[U]) Render() app.UI`),
		doc.List(
			doc.Item("If `b.BaseLayout` is `nil`, `Render` panics with `ErrNoLayout`."),
			doc.Item("Otherwise, if `b.Err()` reports an error, `Render` returns `BaseLayout.Error(err)`."),
			doc.Item("Otherwise, if the base is not yet mounted, or there are no screens, `Render` returns `BaseLayout.Loading()`."),
			doc.Item("Otherwise, if the selected index is greater than or equal to the number of `Screens`, it is reset to `0`."),
			doc.Item("Otherwise, if the selected index is negative (no route matched), `Render` returns `BaseLayout.Error(ErrNoRoute)`."),
			doc.Item("Otherwise, `Render` returns `BaseLayout.Body(selectedScreen.Content()).Draw(Model)`."),
		),
		doc.P("`Render` calls `Body(...)` on the layout on every render, passing the currently selected screen's content. An implementation of `Body` must replace whatever content it stored from a previous call rather than appending to it — appending makes the layout's stored content grow without bound as the app re-renders."),
	}
}

func interfaces() []app.UI {
	return []app.UI{
		doc.P("A consumer of this package implements both of these interfaces."),
		doc.H("Layout[T]"),
		doc.Go(`type Layout[T any] interface {
	Loading() app.UI
	Error(error) app.UI
	Draw(m *data.ViewModel[T]) app.HTMLDiv
	Body(...app.UI) Layout[T]
}`),
		doc.List(
			doc.Item("`Loading()` returns what to show while a `Page` or `Base` is not yet mounted."),
			doc.Item("`Error(err error)` returns what to show when a `Page` or `Base` has an error (see the ", doc.A("#errors", "four exported errors"), " below)."),
			doc.Item("`Draw(m *data.ViewModel[T])` returns the fully assembled markup, given the current model."),
			doc.Item("`Body(content ...app.UI)` receives the content to show inside the layout and returns the layout itself (so it can be chained into `Draw`). On a `Base`, `Body` is called on every render — store the given content by replacing what was stored before, not by appending to it."),
		),
		doc.H("Screen"),
		doc.Go(`type Screen interface {
	Content() app.UI
	Route() string
	OnMount(app.Context) Screen
	OnDismount() Screen
	Mounted() bool
}`),
		doc.List(
			doc.Item("`Content()` returns the screen's own markup."),
			doc.Item("`Route()` returns the screen's path (see `Page.Route` ", doc.A("#page", "above"), " for the leading-slash rule `*Page[T]` follows)."),
			doc.Item("`OnMount(ctx)` is called when the screen becomes selected; it returns the screen."),
			doc.Item("`OnDismount()` is called when the screen stops being selected; it returns the screen."),
			doc.Item("`Mounted()` reports whether the screen is currently mounted."),
		),
		doc.P("`*Page[T]` implements `Screen`, so a `Page` can be used directly as one of a `Base`'s `Screens`. A custom type can implement `Screen` directly instead."),
	}
}
