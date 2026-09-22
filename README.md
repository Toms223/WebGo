# WebGo

Small building blocks for writing web front ends in Go, compiled to WebAssembly.

WebGo sits on top of [go-app](https://github.com/maxence-charriere/go-app) and adds three things: a `Page` that pairs a layout with a view model, a `Base` that swaps pages by route, and a `ViewModel` that holds the state a page reads and writes. A markdown renderer is included for pages whose content is written as text.

**Documentation: https://toms223.github.io/WebGo/**

The docs site is itself a WebGo app — it is built with the packages it documents, and its markdown is rendered by `RenderMarkdown`. The source is in [`docs/`](docs/).

## Install

```sh
go get github.com/Toms223/WebGo@v0.0.4
```

`v0.0.4` is the current release. `v0.0.3` added `ViewModel.State()`; `v0.0.4` made `Base` work when the app is mounted under a subpath rather than at the domain root. `v0.0.1` and `v0.0.2` do not work and should not be used.

You will import `github.com/maxence-charriere/go-app/v11/pkg/app` directly too, so run `go mod tidy` afterwards to record it as a direct dependency.

## Quick start

A WebGo app is one Go program that runs in two places: compiled to WebAssembly in the browser, and natively on your machine to serve or generate the site.

```go
package main

import (
	"net/http"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type hello struct {
	app.Compo
}

func (h *hello) Render() app.UI {
	return app.Text("Hello, WebGo")
}

func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{Name: "Hello"})
	http.ListenAndServe(":8080", nil)
}
```

`app.RunWhenOnBrowser` takes over when the binary is running as WebAssembly and does nothing otherwise, so execution falls through to the server code on your machine. Register every route before calling it, and return a fresh component instance from each factory.

Build the browser binary, then run the server:

```sh
GOARCH=wasm GOOS=js go build -o web/app.wasm .
go run .
```

## The parts

| Type | What it does |
| --- | --- |
| `page.Page[T]` | One screen: a path, a `Layout[T]`, and a `ViewModel[T]` |
| `page.Base[U]` | The app shell: holds the screens and switches between them on navigation |
| `model.ViewModel[T]` | Typed state under a key, readable from a layout and writable through an action |
| `page.Layout[T]` | Interface you implement: loading, error and drawn states |
| `page.Screen` | Interface a screen satisfies; `*Page[T]` already does |
| `markdown` | Renders a markdown string into `[]app.UI` |

### Building a page

`Page` is built by chaining. The first error is latched, so later calls become no-ops and `Err()` reports what went wrong.

```go
state := &AppState{}
vm, err := model.NewViewModel("app", state)
if err != nil {
	log.Fatal(err)
}

p := page.NewPage[AppState]("/settings").
	WithLayout(&settingsLayout{}).
	WithModel(vm)

if err := p.Err(); err != nil {
	log.Fatal(err)
}
```

`NewPage("")` serves `/` — `Route()` always returns a leading slash.

### Implementing a layout

```go
type Layout[T any] interface {
	Loading() app.UI
	Error(error) app.UI
	Draw(m *model.ViewModel[T]) app.HTMLDiv
	Body(...app.UI) Layout[T]
}
```

`Loading` shows before mount, `Error` shows when `Err()` is non-nil, and `Draw` renders the real thing. `Page` picks between the three for you, so a layout never needs its own conditionals.

> `Body` is called on **every** render of a `Base`. Store what it hands you by replacing the previous content, not by appending to it, or the slice grows without bound.

### Reading and writing state

State flows one way: something writes through an action, the layout reads it back in `Draw`.

```go
func (s *screen) load(ctx app.Context) {
	ctx.Async(func() {
		user, err := fetchUser()
		ctx.Dispatch(func(ctx app.Context) {
			s.model.Load(ctx).Apply(func(ctx app.Context, st *AppState) *AppState {
				st.User, st.Err = user, err
				return st
			}).Set()
		})
	})
}

func (l *layout) Draw(m *model.ViewModel[AppState]) app.HTMLDiv {
	return app.Div().Text(m.State().User)
}
```

> Return the **same pointer** your `Apply` verb was given. Returning a new one detaches the model from the observer registered at mount, and the page silently stops updating.

`Set()` writes to the state store and notifies observers. `SetAndPersist(v)` also persists it. A verb that returns `nil` leaves the state untouched.

### Assembling the shell

```go
shell, err := page.NewBase(&appLayout{}, vm, []page.Screen{home, settings})
```

`Base.OnNav` matches the current path against each screen's `Route()`, dismounts the previous screen and mounts the new one. If nothing matches, it renders `Error(ErrNoRoute)`.

### Mounting under a subpath

Screens always use paths rooted at `/`. They never need to know where the app is served from.

On its first mount, `Base` works out its own mount point by comparing the landing URL against the routes it holds, and stores it. From then on it subtracts that prefix before matching, and puts it back in the address bar after each navigation. So an app whose screens are `/`, `/markdown` and `/state` runs unchanged at `example.com/` and at `example.com/docs/`, and the links it renders stay reloadable and shareable in both.

You only need the mount yourself when building a URL to something that is not a route — a data file, say. If you hold the `Base`, ask it:

```go
source := shell.BasePath() + "/data/page.md"
```

From inside a screen, which has no reference to the `Base`, subtract your own route from the current path instead. `Base` fixes the address bar before mounting a screen, so the path is correct by then:

```go
func (s *screen) sourceURL(ctx app.Context) string {
	base := strings.TrimSuffix(ctx.Page().URL().Path, s.Route())
	return strings.TrimSuffix(base, "/") + "/data/page.md"
}
```

Nothing reads an environment variable, and no build flag configures it.

### Rendering markdown

```go
ui := markdown.RenderMarkdown(source, markdown.Style{
	Heading:    "doc-heading",
	Subheading: "doc-subheading",
})
```

Returns one `app.UI` per top-level block, or `nil` for empty input. It is a deliberate subset: headings, lists, blockquotes, code blocks, thematic breaks, emphasis, code spans and links. Heading levels 1–2 become `h3` with your `Heading` class, deeper levels become `h4` with `Subheading`.

Links are filtered. Relative paths starting with `/`, anchors, `http://`, `https://` and `mailto:` become anchors carrying `rel="noopener noreferrer"`; everything else — including bare relative links like `guide.md` and protocol-relative `//host` — renders as a plain `span`. Tables and other GitHub extensions are not supported and pass through as text.

### Errors

`ErrNoLayout`, `ErrNoModel`, `ErrNoScreens` and `ErrNoRoute` are exported sentinels. Match them with `errors.Is`.

## Shipping a static site

`GenerateStaticWebsite` writes `index.html`, the go-app runtime files, and one HTML file per registered route:

```go
app.Route("/", func() app.Composer { return &hello{} })
app.RunWhenOnBrowser()

if err := app.GenerateStaticWebsite("dist", &app.Handler{Name: "Hello"}); err != nil {
	log.Fatal(err)
}
```

Build `web/app.wasm` first — the generator does not produce it, and expects it to already be in place.

Because routing happens in the browser, the host must serve `index.html` for unknown extensionless paths. Most static hosts do this by default; GitHub Pages does it via `404.html`.

## Running the docs locally

```sh
cd docs
podman build -t webgo-docs .
podman run --rm --sysctl net.ipv4.ip_unprivileged_port_start=0 -p 8080:80 webgo-docs
```

Then open http://localhost:8080. The gateway listens on port 80 inside the container and runs as a non-root user, which is what the `--sysctl` flag is for.

## Deploying the docs

[`.github/workflows/pages.yml`](.github/workflows/pages.yml) builds the site and publishes it to GitHub Pages on every push to `main`.

A project Pages site is served from `/<repo>/`, not the domain root, so the build passes the repository name to the generator:

```sh
go run . static -o dist --github-pages WebGo
```

That sets go-app's resource resolver, so the generated HTML points at `/WebGo/app.js`, `/WebGo/web/app.wasm` and so on instead of the domain root.

The flag covers go-app's own assets only. The app's routing needs nothing: `Base` deduces its mount at runtime, so the same source serves correctly from `/` in the container and from `/WebGo/` on Pages, with no build-time switch and no environment variable.

## Requirements

- Go matching the version in your `go.mod`
- `github.com/Toms223/WebGo` v0.0.4 or later
- `github.com/maxence-charriere/go-app/v11`
