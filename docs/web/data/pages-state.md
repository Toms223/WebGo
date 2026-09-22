The `page` package builds screens out of a `Layout` and a `model.ViewModel`, and assembles them under a `Base` component that switches between screens by route. The `model` package holds the state each screen reads and writes.

This page covers `github.com/Toms223/WebGo` from `v0.0.3` onward. `ViewModel.State()` only exists from `v0.0.3` onward.

## Page

```go
type Page[T any] struct {
	Path string
	// unexported fields
}

func NewPage[T any](path string) *Page[T]
```

`NewPage` creates a page for the given path. It has no layout and no model yet, so it is not usable until both are set.

### Building a page

`Page` is built by chaining calls. Each call returns the same `*Page[T]`, so the order of calls does not matter:

```go
package main

import (
	"github.com/Toms223/WebGo/model"
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

func (s *shellLayout) Draw(m *model.ViewModel[AppState]) app.HTMLDiv {
	return app.Div().Body(s.body...)
}

func (s *shellLayout) Body(content ...app.UI) page.Layout[AppState] {
	s.body = content
	return s
}

func main() {
	vm, err := model.NewViewModel("app-state", &AppState{})
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
}
```

- `At(path string) *Page[T]` sets the page's path.
- `WithLayout(layout Layout[T]) *Page[T]` sets the page's layout. Passing `nil` sets the page's error to `ErrNoLayout`.
- `WithModel(m *model.ViewModel[T]) *Page[T]` sets the page's model. Passing `nil` sets the page's error to `ErrNoModel`.

### Error latching

A `Page` keeps a single internal error. The first builder call that produces an error (`WithLayout(nil)` or `WithModel(nil)`) stores that error. Once an error is stored, every later builder call (`At`, `WithLayout`, `WithModel`) returns immediately without changing anything on the page — not the path, not the layout, not the model.

This means only the first failure in a chain is kept. If both `WithLayout(nil)` and `WithModel(nil)` appear in the same chain, the page's error is whichever of the two ran first.

### Err

```go
func (p *Page[T]) Err() error
```

`Err` returns the page's error, if one was latched by a builder call. If no error was latched, it checks the page's current layout and model:

- If the layout is `nil`, it returns `ErrNoLayout`.
- Otherwise, if the model is `nil`, it returns `ErrNoModel`.
- Otherwise, it returns `nil`.

### Route

```go
func (p *Page[T]) Route() string
```

`Route` returns the page's path with exactly one leading slash, regardless of what was passed to `NewPage` or `At`. A path of `""` becomes `/`, `"settings"` becomes `/settings`, and `"/settings"` stays `/settings`.

### OnMount, OnDismount, Mounted

```go
func (p *Page[T]) OnMount(ctx app.Context) Screen
func (p *Page[T]) OnDismount() Screen
func (p *Page[T]) Mounted() bool
```

`OnMount` is called by a `Base` when the page becomes the selected screen. If `Err()` reports an error, `OnMount` stores that error and returns the page without mounting it — it does not panic, even if the layout or model is missing. Otherwise, it reads the model's current state from `ctx`, starts observing the model for further state changes for as long as the page reports itself as mounted, and marks the page mounted.

`OnDismount` marks the page as not mounted and returns it.

`Mounted` reports whether the page is currently mounted.

### Content

```go
func (p *Page[T]) Content() app.UI
```

`Content` renders the page:

- If the page has no layout, `Content` panics with `ErrNoLayout`.
- Otherwise, if `Err()` reports an error, `Content` returns `layout.Error(err)`.
- Otherwise, if the page is not mounted, `Content` returns `layout.Loading()`.
- Otherwise, `Content` returns `layout.Draw(model)`.

Only `*Page[T]` (a pointer) satisfies the `Screen` interface described below. A `Page[T]` value does not.

## Base

```go
type Base[U any] struct {
	BaseLayout Layout[U]
	Model      *model.ViewModel[U]
	Screens    []Screen
	// unexported fields
}
```

`Base` is a `go-app` component that hosts a set of `Screen`s under one shared `Layout` and one shared model, and switches between them by URL route. It implements `go-app`'s `app.Mounter` and `app.Navigator`, but it does not itself satisfy `Screen` — it is meant to be the root component, not a screen nested inside another `Base`.

### NewBase

```go
func NewBase[U any](layout Layout[U], model *model.ViewModel[U], screens []Screen) (*Base[U], error)
```

`NewBase` requires all three arguments:

- If `layout` is `nil`, it returns `ErrNoLayout`.
- Otherwise, if `model` is `nil`, it returns `ErrNoModel`.
- Otherwise, if `screens` is `nil`, it returns `ErrNoScreens`.

These checks run in that order, so if more than one argument is missing, the error names the first one checked. An empty, non-nil slice (`[]Screen{}`) is accepted — it is not the same as `nil`.

On success, `NewBase` returns a `*Base[U]` with the first screen (index `0`) selected. Nothing is mounted yet.

### OnMount

```go
func (b *Base[U]) OnMount(ctx app.Context)
```

If `b.Err()` reports an error (missing layout or model), `OnMount` does nothing. Otherwise it reads the model's current state from `ctx`. If there are no screens, it marks the base mounted and stops. Otherwise it calls `OnMount` on the currently selected screen (index `0` by default) and marks the base mounted. Calling `OnMount` a second time is harmless as long as a screen is currently selected. If `OnNav` previously found no matching route, the selection is negative and a second `OnMount` call panics indexing `Screens` with it.

### OnNav

```go
func (b *Base[U]) OnNav(ctx app.Context)
```

`OnNav` matches the current browser path (`ctx.Page().URL().Path`) against each screen's `Route()`, in the order screens appear in `Screens`:

- If a screen's route matches the path and that screen is already selected, `OnNav` returns without doing anything else.
- If a screen's route matches the path and it is not already selected: the previously selected screen (if any is currently selected) is dismounted via `OnDismount`, the matching screen becomes selected, the matching screen is mounted via `OnMount`, and the browser URL is replaced with the matched route.
- If no screen's route matches the path, the selection is cleared (set to an index that selects nothing).

### Render

```go
func (b *Base[U]) Render() app.UI
```

- If `b.BaseLayout` is `nil`, `Render` panics with `ErrNoLayout`.
- Otherwise, if `b.Err()` reports an error, `Render` returns `BaseLayout.Error(err)`.
- Otherwise, if the base is not yet mounted, or there are no screens, `Render` returns `BaseLayout.Loading()`.
- Otherwise, if the selected index is greater than or equal to the number of `Screens`, it is reset to `0`.
- Otherwise, if the selected index is negative (no route matched), `Render` returns `BaseLayout.Error(ErrNoRoute)`.
- Otherwise, `Render` returns `BaseLayout.Body(selectedScreen.Content()).Draw(Model)`.

`Render` calls `Body(...)` on the layout on every render, passing the currently selected screen's content. An implementation of `Body` must replace whatever content it stored from a previous call rather than appending to it — appending makes the layout's stored content grow without bound as the app re-renders.

## Layout and Screen interfaces

A consumer of this package implements both of these interfaces.

### Layout[T]

```go
type Layout[T any] interface {
	Loading() app.UI
	Error(error) app.UI
	Draw(m *model.ViewModel[T]) app.HTMLDiv
	Body(...app.UI) Layout[T]
}
```

- `Loading()` returns what to show while a `Page` or `Base` is not yet mounted.
- `Error(err error)` returns what to show when a `Page` or `Base` has an error (see the four exported errors below).
- `Draw(m *model.ViewModel[T])` returns the fully assembled markup, given the current model.
- `Body(content ...app.UI)` receives the content to show inside the layout and returns the layout itself (so it can be chained into `Draw`). On a `Base`, `Body` is called on every render — store the given content by replacing what was stored before, not by appending to it.

### Screen

```go
type Screen interface {
	Content() app.UI
	Route() string
	OnMount(app.Context) Screen
	OnDismount() Screen
	Mounted() bool
}
```

- `Content()` returns the screen's own markup.
- `Route()` returns the screen's path (see `Page.Route` above for the leading-slash rule `*Page[T]` follows).
- `OnMount(ctx)` is called when the screen becomes selected; it returns the screen.
- `OnDismount()` is called when the screen stops being selected; it returns the screen.
- `Mounted()` reports whether the screen is currently mounted.

`*Page[T]` implements `Screen`, so a `Page` can be used directly as one of a `Base`'s `Screens`. A custom type can implement `Screen` directly instead.

## Model

```go
type ViewModel[T any] struct {
	// unexported fields
}
```

A `ViewModel[T]` binds a state value of type `T` to a key in `go-app`'s state store.

### NewViewModel

```go
func NewViewModel[T any](key string, value *T) (*ViewModel[T], error)
```

- If `value` is `nil`, it returns an error (`"value is nil"`).
- If `key` is `""`, it returns an error (`"key is empty"`).
- Otherwise, it returns a `*ViewModel[T]` that holds `value` directly — no copy is made, so later changes made through the returned `ViewModel` are visible through the original pointer, and vice versa.

`NewViewModel` does not read or write the state store. The value passed in is used as-is until `Get` or `Load` is called.

### Get

```go
func (v *ViewModel[T]) Get(ctx app.Context) *ViewModel[T]
```

`Get` reads the current value for the model's key from the state store into the model's existing state value, and returns the model itself. If the key has never been written, or the stored value is a different type, `Get` leaves the model's state unchanged.

### State

```go
func (v *ViewModel[T]) State() *T
```

`State` returns the model's current in-memory state pointer. It does not touch the state store — call `Get` or `Load` first if the state needs to be refreshed from the store.

### Load

```go
func (v *ViewModel[T]) Load(ctx app.Context) *Action[T]
```

`Load` refreshes the model's state from the store (the same read `Get` performs), then returns a new `*Action[T]` bound to `ctx` and the model, ready for `Apply`, `Set`, or `SetAndPersist`.

### Observe

```go
func (v *ViewModel[T]) Observe(ctx app.Context) app.Observer
```

`Observe` returns an `app.Observer` that watches the model's key in the state store. It fires when the store's value for that key changes through `Set` or `SetAndPersist` — not on a direct write to the model's in-memory state.

### Action.Apply

```go
type Action[T any] struct {
	// unexported fields
}

func (a *Action[T]) Apply(verb func(ctx app.Context, state *T) *T) *Action[T]
```

`Apply` calls `verb` with the action's context and the model's current in-memory state, then returns the action itself. If `verb` returns a non-nil pointer, the model's in-memory state is replaced with that pointer. If `verb` returns `nil`, the model's state is left unchanged.

Return the same pointer you were given rather than a fresh `&T{...}`: the observer registered at `Page.OnMount` still holds the previous pointer, so returning a new one detaches the model from that observer and the page silently stops updating.

Calls to `Apply` chain: each call's `verb` sees the state left by the previous call.

`Apply` only changes the model's in-memory state. It does not publish anything to the state store — call `Set` or `SetAndPersist` for that.

### Set

```go
func (a *Action[T]) Set() *ViewModel[T]
```

`Set` writes the model's current in-memory state (as left by any `Apply` calls, or as it was when `Load` ran if `Apply` was never called) to the state store, and returns the model. `Set` does not write to local storage.

### SetAndPersist

```go
func (a *Action[T]) SetAndPersist(value *T) *ViewModel[T]
```

`SetAndPersist` sets the model's in-memory state directly to `value` (ignoring whatever `Apply` may have built up), writes `value` to the state store, and persists a snapshot of `value` to local storage. It returns the model.

The local storage snapshot is fixed at the moment `SetAndPersist` runs. Later direct changes to the model's in-memory state do not change what was persisted.

## Errors

The `page` package exports four sentinel errors:

- `ErrNoLayout` — a `Page` or `Base` has no layout.
- `ErrNoModel` — a `Page` or `Base` has no model.
- `ErrNoScreens` — `NewBase` was called with a `nil` screens slice.
- `ErrNoRoute` — a `Base`'s current selection does not match any screen's route.

Check for a specific one with `errors.Is`, for example `errors.Is(p.Err(), page.ErrNoModel)`.

## Example: fetching and displaying state

Storing a fetched value and reading it back in a layout's `Draw`:

```go
package main

import (
	"github.com/Toms223/WebGo/model"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type AppState struct {
	User string
}

func fetchUser() string {
	return "ada"
}

func loadUser(ctx app.Context, vm *model.ViewModel[AppState]) {
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

func (l *AppLayout) Draw(m *model.ViewModel[AppState]) app.HTMLDiv {
	current := m.State()
	return app.Div().Text(current.User)
}

func main() {
	vm, err := model.NewViewModel("app-state", &AppState{})
	if err != nil {
		panic(err)
	}
	layout := &AppLayout{}
	layout.Draw(vm)

	_ = loadUser
}
```

`loadUser` fetches the user, then applies and publishes the new state through the model. `Draw` reads the same model's current state back out with `State()` on every render.
