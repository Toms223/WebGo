package state

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func shellSection() []app.UI {
	return []app.UI{
		doc.Go(`type Shell[U any] struct {
	app.Compo
	Layout  Layout[U]
	Model   *Model[U]
	Screens []Routed
	// unexported fields
}`),
		doc.P("`Shell` is the root component of an app. It holds every screen under one shared `Layout` and one shared model, and draws the screen whose route matches the URL inside that layout. There is one shell per app."),
		shellDiagram(),
		doc.H("NewShell"),
		doc.Go(`func NewShell[U any](layout Layout[U], model *Model[U], screens ...Routed) (*Shell[U], error)`),
		doc.List(
			doc.Item("If `layout` is `nil`, it returns `ErrNoLayout`."),
			doc.Item("Otherwise, if `model` is `nil`, it returns `ErrNoModel`."),
			doc.Item("Otherwise, if `screens` is `nil`, it returns `ErrNoScreens`. Calling it with no screens at all passes `nil`; an empty, non-nil slice spread with `...` is accepted."),
		),
		doc.P("On success it returns a shell with no screen selected."),
		doc.H("Serve"),
		doc.Go(`func Serve[U any](build func() *Shell[U])`),
		doc.P("`Serve` registers the route of every screen in the shell `build` returns, all leading to that shell. Call it before `app.RunWhenOnBrowser`, in the `main` shared by the server and the browser, as with `app.Route`."),
		doc.Go(`func main() {
	app.Serve(func() *app.Shell[AppState] {
		home := app.NewScreen[AppState]("/").WithLayout(&homeLayout{}).WithModel(model)
		settings := app.NewScreen[AppState]("/settings").WithLayout(&settingsLayout{}).WithModel(model)
		shell, err := app.NewShell(&appLayout{}, model, home, settings)
		if err != nil {
			panic(err)
		}
		return shell
	})
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{Name: "Shop"})
	http.ListenAndServe(":8080", nil)
}`),
		doc.List(
			doc.Item("`build` must return new screens, and may return a new layout, on every call. A screen built once outside it and reused fails on the server's second request with \"component is already mounted\"."),
			doc.Item("`Serve` calls `build` once itself, to read the routes."),
			doc.Item("In the browser the shell is built once and kept across navigations. Moving between its routes does not create a new root: the shell stays mounted and switches its screen in `OnNav`."),
			doc.Item("On the server it is built once per request."),
			doc.Item("Before the first render, the engine selects the screen for the current path, so the shell draws its layout from the first frame. A screen that is still loading shows its own `Loading()` inside it."),
		),
		doc.H("OnMount"),
		doc.Go(`func (s *Shell[U]) OnMount(ctx app.Context)`),
		doc.P("If `Err()` reports an error (no layout or no model), `OnMount` does nothing. Otherwise it reads the model's state from the store with `Get`. It does not observe the model."),
		doc.H("OnNav"),
		doc.Go(`func (s *Shell[U]) OnNav(ctx app.Context)`),
		doc.List(
			doc.Item("It takes the current path, removes the mount prefix, and compares it with each screen's `Route()` in the order of `Screens`. A trailing slash is ignored on both sides."),
			doc.Item("The first screen whose route matches is selected. If the address bar lacks the mount prefix, it is put back with `ReplaceURL`."),
			doc.Item("If no route matches, the selection is cleared and `Render` shows `Layout.Error(ErrNoRoute)`."),
		),
		doc.H("Render"),
		doc.Go(`func (s *Shell[U]) Render() app.UI`),
		doc.List(
			doc.Item("If `Layout` is `nil`, `Render` panics with `ErrNoLayout`."),
			doc.Item("Otherwise, if `Err()` reports an error, it returns `Layout.Error(err)`."),
			doc.Item("Otherwise, if no screen is selected, it returns `Layout.Error(ErrNoRoute)`."),
			doc.Item("Otherwise it returns `Layout.Body(selected).Draw(Model)`, with the selected screen itself as the body."),
		),
		doc.P("`Render` calls `Body` on every render. An implementation of `Body` must replace the content it stored before, not append to it, or the stored content grows without bound."),
		doc.H("The mount prefix"),
		doc.P("The shell reads the mount prefix from the environment variable `GOAPP_ROOT_PREFIX`, which `app.Handler` sets from its `Resources`: `app.GitHubPages(\"shop\")` makes it `/shop`, and the default makes it `/`. Links still need the prefix; see ", doc.A("/routing#serving-under-a-prefix", "Serving under a prefix"), "."),
		doc.H("Screens of the same type"),
		doc.P("Two screens of the same Go type switch cleanly, because `Screen.CompoID` is its route, and a screen on another route is mounted fresh. A custom `Routed` component used at several routes should implement `CompoID` too, returning its route; otherwise the engine keeps the mounted one and only copies its exported fields."),
	}
}

func interfaces() []app.UI {
	return []app.UI{
		doc.H("Layout[T]"),
		doc.Go(`type Layout[T any] interface {
	Loading() UI
	Error(error) UI
	Draw(m *Model[T]) HTMLDiv
	Body(...UI) Layout[T]
}`),
		doc.P("A `Screen` and a `Shell` each draw through a layout you implement."),
		doc.List(
			doc.Item("`Loading()` is what a `Screen` shows while its `OnLoad` runs. A `Shell` never calls it."),
			doc.Item("`Error(err)` is what a `Screen` shows for a builder error or a failed load, and what a `Shell` shows for a builder error or `ErrNoRoute`. See the ", doc.A("#errors", "four exported errors"), "."),
			doc.Item("`Draw(m)` returns the assembled markup for the current model."),
			doc.Item("`Body(content...)` receives what to show inside the layout and returns the layout, so it chains into `Draw`. A `Shell` calls it on every render with the selected screen; a `Screen` never calls it, so a screen's layout can return itself and ignore the content. Screens are built once and reused, so `Draw` must keep the content at the same position under the same elements on every render: vary what is around it, not where it is."),
		),
		doc.H("Routed"),
		doc.Go(`type Routed interface {
	Composer
	Route() string
}`),
		doc.P("`Routed` is a component served at a route; a `Shell`'s `Screens` are `Routed`. `*Screen[T]` implements it, and so can any component of your own. Implement `CompoID` as well when the same type is used at several routes, as described under ", doc.A("#screens-of-the-same-type", "Screens of the same type"), "."),
	}
}
