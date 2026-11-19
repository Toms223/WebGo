package scripts

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/scripts/demos"
)

const (
	Title = "Scripts"
	Route = "/scripts"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "scripts",
		Path:  Route,
		Title: Title,
		Lead:  "Load ES modules through one importmap and reach them from Go through a typed wrapper.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Declaring a module", Body: declaring},
			{Title: "Sources", Body: sources},
			{Title: "Loading order", Body: loading},
			{Title: "Using a module", Body: using},
			{Title: "Writing a source", Body: writing},
			{Title: "Example: confetti", Body: confetti},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("An ES module is a JavaScript file that other code loads by name. `Handler.Imports` lists the modules a page needs. The handler writes one importmap for them, the browser loads them while `app.wasm` downloads, and Go code calls them through a typed wrapper from the `script` package."),
		doc.P("Classic scripts, the kind that set globals, stay on `Handler.Scripts`, covered under ", doc.A("/deploy#head-resources", "Head resources"), "."),
	}
}

func declaring() []app.UI {
	return []app.UI{
		doc.Go(`Imports []Module

type Module interface {
	Source() ModuleSource
}

type ModuleSource interface {
	ImportMap() (name, url string)
}`),
		doc.P("A `Module` returns a `ModuleSource`, which gives the name the module is imported by and the URL it is served from. The `script` package builds the sources, and a wrapper type ties one to the Go code that calls it. This one loads Datadog's browser SDK. `Init` runs once and configures the module from it:"),
		doc.Go(`type RUM struct{ ClientToken, AppID string }

func (RUM) Source() app.ModuleSource { return script.CDN("npm", "@datadog/browser-rum", "5.20.0") }

func (r RUM) Init(ctx app.Context, mod app.Value) app.Value {
	rum := mod.Get("datadogRum")
	rum.Call("init", map[string]any{"clientToken": r.ClientToken, "applicationId": r.AppID})
	return rum
}

var rum = RUM{ClientToken: "...", AppID: "..."}

// Handler
Imports: []app.Module{rum}

// a component
if r, ok := script.Use(ctx, rum); ok {
	r.Call("addAction", "clicked")
}`),
		doc.P("Declare the value once, as the package-level `rum`, and use it in both places. `Init` runs with whichever value reaches `Use` first, so a component that wrote `RUM{}` would configure Datadog with empty credentials."),
		doc.P("`Build` checks the imports and returns one of these errors:"),
		doc.List(
			doc.Item("`ErrMissingModule`: a local module file is not under `WebDir`."),
			doc.Item("`ErrDuplicateModule`: two modules have the same name."),
			doc.Item("`ErrInvalidModule`: a module has no source or no name, or its source failed its own check, such as a CDN version that is not exact."),
		),
	}
}

func sources() []app.UI {
	return []app.UI{
		doc.Go(`func Local(name, path string) app.ModuleSource
func CDN(registry, name, version string) app.ModuleSource`),
		doc.List(
			doc.Item("`Local` serves the module from a file under `/web/`. `Build` checks that the file exists, and the URL goes through `Resources`, so it follows a GitHub Pages prefix like every other local file."),
			doc.Item("`CDN` serves it from esm.sh. The registry is `npm`, `gh` or `jsr`."),
		),
		doc.Go(`script.CDN("npm", "canvas-confetti", "1.9.3")
// https://esm.sh/canvas-confetti@1.9.3

script.CDN("gh", "user/repo", "1.2.3")
// https://esm.sh/gh/user/repo@1.2.3

script.CDN("jsr", "@std/path", "1.0.0")
// https://esm.sh/jsr/@std/path@1.0.0`),
		doc.P("The version must be exact, such as `1.9.3` or `1.0.0-beta.1`, with no `v` prefix and no range. `latest` and `^1.0.0` fail `Build`. A pinned URL always returns the same file, and `Build` never goes to the network."),
	}
}

func loading() []app.UI {
	return []app.UI{
		doc.P("The importmap is the first thing in `<head>` after the charset, because the browser reads an importmap only before the first module import. The order is under ", doc.A("/deploy#the-generated-document", "The generated document"), "."),
		doc.List(
			doc.Item("`app.js` imports every module in parallel with the `app.wasm` download and waits for all of them before it starts the wasm."),
			doc.Item("A module that fails to import, for example because the CDN is blocked, is logged to the console. The wasm still starts and the page works."),
			doc.Item("The loaded modules are on `window.goappModules`, by name. `script.Use` reads them from there."),
		),
	}
}

func using() []app.UI {
	return []app.UI{
		doc.Go(`type Script[T any] interface {
	Module
	Init(ctx app.Context, mod app.Value) T
}

func Use[T any](ctx app.Context, s Script[T]) (T, bool)`),
		doc.List(
			doc.Item("`Use` returns the wrapper and `true`, or the zero value and `false`. It is `false` on the server, where nothing is loaded and `Init` never runs, and in the browser when the import failed."),
			doc.Item("`Init` runs once per module name. Later calls reuse its result."),
			doc.Item("`Use` is allowed in `Render`, but a module with side effects belongs in `OnMount` or an event handler, where it runs once in the browser."),
			doc.Item("`Init` may `Use` another module. It must not `Use` its own, directly or through a cycle, which would wait for itself."),
		),
		doc.Go(`func (c *page) OnMount(ctx app.Context) {
	if r, ok := script.Use(ctx, rum); ok {
		r.Call("addAction", "page-view")
	}
}`),
		doc.P("Check the `bool` every time. A page that renders on the server and a visitor whose network blocks the CDN both get `false`."),
	}
}

func writing() []app.UI {
	return []app.UI{
		doc.P("`Local` and `CDN` are two implementations of `ModuleSource`. Write your own for a registry or a URL scheme they do not cover:"),
		doc.Go(`type jsdelivr struct{ name, version string }

func (u jsdelivr) ImportMap() (name, url string) {
	return u.name, "https://cdn.jsdelivr.net/npm/" + u.name + "@" + u.version + "/+esm"
}`),
		doc.P("A source may also have a `Check() error` method. `Build` calls it, and a non-nil error comes back wrapped in `ErrInvalidModule`. `CDN` uses it to reject a version that is not exact."),
	}
}

func confetti() []app.UI {
	return []app.UI{
		doc.Example(demos.ConfettiExample),
		doc.P("The click calls `script.Use`. When canvas-confetti loaded, `Fire` shoots 120 particles at a 70 degree spread. When it did not, the component shows `module not loaded` and nothing else breaks."),
	}
}
