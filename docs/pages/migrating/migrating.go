package migrating

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Migrating"
	Route = "/migrating"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "migrating",
		Path:  Route,
		Title: Title,
		Lead:  "Move an app from go-app v11 and WebGo v0.0.5 to WebGo v0.1.0.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Import paths", Body: imports},
			{Title: "Element styles", Body: elementStyles},
			{Title: "One handler", Body: oneHandler},
			{Title: "Service worker and updates", Body: serviceWorker},
			{Title: "Behaviour changes without syntax changes", Body: behaviour},
			{Title: "Packages not included", Body: packages},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("WebGo v0.1.0 contains go-app v11.0.5's `pkg/app` as its own `app` package. Components, routing, the context, state and actions keep their names and signatures. What changes is the import path, the inline-style setters, the handler, the update flow and a few rendering details."),
		doc.P("Each section shows the code before and after. The compiler finds the syntax changes; the behaviour changes are listed under ", doc.A("#behaviour-changes-without-syntax-changes", "Behaviour changes without syntax changes"), " and ", doc.A("#service-worker-and-updates", "Service worker and updates"), "."),
	}
}

func imports() []app.UI {
	return []app.UI{
		doc.P("Before:"),
		doc.Go(`import (
	"github.com/Toms223/WebGo/serve"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)`),
		doc.Plain(`require (
	github.com/Toms223/WebGo v0.0.5
	github.com/maxence-charriere/go-app/v11 v11.0.5
)`),
		doc.P("After:"),
		doc.Go(`import "github.com/Toms223/WebGo/app"`),
		doc.Plain(`require github.com/Toms223/WebGo v0.1.0`),
		doc.Sh(`go get github.com/Toms223/WebGo@v0.1.0
go mod tidy`),
		doc.List(
			doc.Item("Replace every `github.com/maxence-charriere/go-app/v11/pkg/app` import with `github.com/Toms223/WebGo/app`. The package name is still `app`."),
			doc.Item("The two cannot be mixed. `page`, `data`, `markdown` and `style` take and return WebGo's `app` types, and go-app's `app.UI` is a different type: the compiler reports `cannot use … (variable of type \"github.com/maxence-charriere/go-app/v11/pkg/app\".UI) as \"github.com/Toms223/WebGo/app\".UI`."),
			doc.Item("The `serve` package is gone, under ", doc.A("#one-handler", "One handler"), "."),
			doc.Item("WebGo needs Go 1.27; go-app v11.0.5 needed 1.26. Build `app.wasm` with the same Go version as the server, because WebGo embeds a single `wasm_exec.js` instead of choosing one by Go version."),
		),
	}
}

func elementStyles() []app.UI {
	return []app.UI{
		doc.P("Every element's inline-style setters are renamed with a `Raw` prefix, because `Style` now takes typed class rules."),
		doc.P("Before:"),
		doc.Go(`app.Div().
	Style("display", "flex").
	Stylef("width", "%dpx", w).
	Styles(map[string]string{"gap": "8px", "color": "red"})`),
		doc.P("After:"),
		doc.Go(`app.Div().
	RawStyle("display", "flex").
	RawStylef("width", "%dpx", w).
	RawStyles(map[string]string{"gap": "8px", "color": "red"})`),
		doc.H("What the compiler says"),
		doc.List(
			doc.Item("`Stylef` and `Styles` no longer exist on any element: `app.Div().Stylef undefined (type app.HTMLDiv has no field or method Stylef)`."),
			doc.Item("`Style` still exists on 98 elements, with a new signature, so an old two-string call is a type error rather than \"undefined\": `cannot use \"display\" (untyped string constant) as app.ClassRule value in argument to app.Div().Style: string does not implement app.ClassRule (missing method ClassNames)`."),
			doc.Item("The other 12 elements have only the `Raw` setters, so `Style` there is \"undefined\": `Area`, `Base`, `DataList`, `Head`, `Link`, `Meta`, `Param`, `Script`, `Source`, `Style`, `Template` and `Title`."),
		),
		doc.P("Replace `.Style(` with `.RawStyle(` only where it has two string arguments, and `.Stylef(` and `.Styles(` everywhere. A plain search and replace of `.Style(` also rewrites typed calls."),
		doc.H("Typed styles"),
		doc.P("The new `Style` takes rules from the `style` package and adds their class names, like `Class` does. Code from WebGo v0.0.5 that passed class names still compiles; the typed form is shorter and checked by the compiler:"),
		doc.P("Before:"),
		doc.Go(`app.Div().Class(theme.Card.Name(), theme.Compact.Name())`),
		doc.P("After:"),
		doc.Go(`app.Div().Style(theme.Card, theme.Compact)`),
		doc.P("A rule that has no class name is logged and skipped. See ", doc.A("/styling", "Styling"), "."),
		doc.H("Generic code"),
		doc.P("`page.Layout[T].Draw` returns `app.HTMLDiv`, and a type that implements it must use WebGo's. Helpers that styled several element types can use the new `app.Styler[E]` constraint:"),
		doc.Go(`func card[E any](e app.Styler[E]) E {
	return e.Style(theme.Card)
}`),
	}
}

func oneHandler() []app.UI {
	return []app.UI{
		doc.P("`serve.Handler` wrapped go-app's handler to add WebGo stylesheets. Its fields are now part of `app.Handler`, which is flat, and `Build` changes the handler in place instead of returning a new one."),
		doc.P("Before:"),
		doc.Go(`h := &serve.Handler{
	Handler: app.Handler{
		Name:   "Hello",
		Styles: []string{"/web/extra.css"},
	},
	Styles: theme.Styles(),
	WebDir: "web",
}
built, err := h.Build()
if err != nil {
	log.Fatal(err)
}
http.Handle("/", built)
if err := app.GenerateStaticWebsite("dist", built); err != nil {
	log.Fatal(err)
}`),
		doc.P("After:"),
		doc.Go(`h := &app.Handler{
	Name:   "Hello",
	Styles: append(theme.Styles(), style.Static("/web/extra.css")),
}
if err := h.Build(); err != nil {
	log.Fatal(err)
}
http.Handle("/", h)
if err := app.GenerateStaticWebsite("dist", h); err != nil {
	log.Fatal(err)
}`),
		doc.List(
			doc.Item("`Styles` is `[]app.Stylesheet`, and `style.Style` is an alias of `app.Stylesheet`. A string becomes `style.Static(path)`; a URL too: `style.Static(\"https://cdn.example.com/x.css\")`."),
			doc.Item("A local Static path must be under `/web/` and exist in `WebDir`. go-app linked any string without checking, so `\"/styles.css\"` or `\"/web/x.css crossorigin\"` used to work and now fail `Build`."),
			doc.Item("`WebDir` defaults to `web`, or to `<dir>/web` when `Resources` is `app.LocalDir(dir)` with a directory."),
			doc.Item("`Build` is optional. The first request builds the handler, and `app.GenerateStaticWebsite` always builds it and returns the error. Call it yourself to fail at startup, because a server whose build fails answers every request with a 500."),
			doc.Item("The link order is `/app.css`, then library stylesheets sorted by path, then `Styles` in order. With `serve`, WebGo's links came before go-app's strings."),
			doc.Item("The handler no longer changes `h.Styles`. go-app used to prepend `/app.css` and the library paths to it."),
		),
		doc.H("Errors"),
		doc.P("Before:"),
		doc.Go(`errors.Is(err, serve.ErrMissingStyle)`),
		doc.P("After:"),
		doc.Go(`errors.Is(err, app.ErrMissingStyle)`),
		doc.P("The same goes for `ErrPathOutsideWeb`, `ErrNotCSSPath` and `ErrDuplicatePath`. Their messages start with `app: ` instead of `serve: `. See ", doc.A("/applying#handler-and-build", "Handler and Build"), " and ", doc.A("/applying#errors", "Errors"), "."),
	}
}

func serviceWorker() []app.UI {
	return []app.UI{
		doc.P("No code has to change, but installed apps behave differently after the first deploy. The full flow is under ", doc.A("/deploy#versions-and-updates", "Versions and updates"), "."),
		doc.List(
			doc.Item("`Version` used to default to a hash of the start time, so every restart looked like an update, and replicas disagreed. It now defaults to a hash of the content, so an empty `Version` is right in production. If you set it to work around the old default, you can remove it."),
			doc.Item("Open pages check for updates on load, when the browser comes back online, when the tab becomes visible, and every 60 seconds."),
			doc.Item("When a new version takes control, the page reloads itself once. Unsaved state in the page, such as a half-filled form, is lost."),
			doc.Item("A component that showed an update button from `OnAppUpdate` and `ctx.AppUpdateAvailable()` will rarely be seen, because the reload comes first. Set `DisableAutoUpdate: true` to keep it."),
			doc.Item("A file the worker stores that fails to download (a 404 or a redirect) now fails the install, and the old version keeps running. go-app logged the error and installed anyway."),
			doc.Item("The worker answers only `GET` requests for its own same-origin files. go-app answered every request from any cache, so the app's own Cache API entries and cross-origin files such as CDN stylesheets or `RemoteBucket` files are no longer served offline."),
			doc.Item("The worker deletes only old `app-` caches. go-app deleted every other cache, the app's own included."),
			doc.Item("The update checker is an inline script. A Content Security Policy without `'unsafe-inline'` or its hash blocks it, which turns automatic updates off."),
		),
		doc.H("DefaultAppWorkerJS"),
		doc.P("`app.DefaultAppWorkerJS` is a `var` now, not a `const`, and holds WebGo's worker. Code that used it in a constant expression fails with `is not constant`."),
		doc.P("Before:"),
		doc.Go(`const worker = app.DefaultAppWorkerJS + extra`),
		doc.P("After:"),
		doc.Go(`var worker = app.DefaultAppWorkerJS + extra`),
		doc.P("A custom `ServiceWorkerTemplate` written from go-app's worker loses the update behaviour above. Start again from `app.DefaultAppWorkerJS`."),
	}
}

func behaviour() []app.UI {
	return []app.UI{
		doc.P("These compile unchanged and render differently. Most of them only matter to tests that compare HTML."),
		doc.List(
			doc.Item("Attribute values in server-rendered HTML are escaped as HTML. go-app quoted them as Go strings, which broke on some values."),
			doc.Item("Attributes in server-rendered HTML are written in name order. go-app used map order, which changed between runs, so golden files are now stable."),
			doc.Item("`RawStyles` applies its keys in sorted order, on the server and in the browser."),
			doc.Item("Library stylesheets are linked sorted by path. go-app used map order."),
			doc.Item("The `ETag` check no longer covers `/web/` files and proxied files such as `robots.txt`. They are validated by their modification time instead, so a changed file is never answered with `304` because the version matched."),
			doc.Item("A handler whose build fails answers every request with `500 Internal Server Error` and logs the error. With `serve`, `Build` was required, so the same error stopped the program at startup."),
			doc.Item("A path with no route renders `app.NotFound` with status `404` on the server, and assigning `app.NotFound` takes effect. go-app answered a plain-text `404` and ignored the variable."),
			doc.Item("`LocalDir(dir)` links `/web/...` instead of the relative `dir/web/...`, and a handler using `PrefixedLocation` or `GitHubPages` serves `/web/` from `./web` instead of panicking."),
			doc.Item("With `Icon.Default` set and `Icon.SVG` empty, the favicon is `Default`, not go-app's logo."),
			doc.Item("`ShowAppInstallPrompt` does nothing on Apple browsers and on a second call, instead of throwing a JavaScript error."),
			doc.Item("`GenerateStaticWebsite` writes library stylesheets."),
			doc.Item("On the first request, a handler nobody built writes its Generated stylesheets to `WebDir`. On a read-only filesystem where they are missing or stale, that build fails with the 500 above."),
		),
		doc.P("Before:"),
		doc.Go(`app.Div().Title("a \"b\" & c")`),
		doc.HTML(`<div title="a \"b\" & c"></div>`),
		doc.P("After:"),
		doc.HTML(`<div title="a &#34;b&#34; &amp; c"></div>`),
	}
}

func packages() []app.UI {
	return []app.UI{
		doc.P("Only go-app's `pkg/app` became part of WebGo. Its other packages have no WebGo import path:"),
		doc.List(
			doc.Item("`ui`, go-app's layout components. Use plain elements styled with the ", doc.A("/styling", "style"), " package, and ", doc.A("/state", "page.Base"), " for the shell."),
			doc.Item("`analytics`. Add the provider's snippet to `Handler.RawHeaders` or `Handler.Scripts`, and call it through `app.Window()`."),
			doc.Item("`cache`, an in-memory cache interface. Use a map guarded by a `sync.Mutex`, or a cache library."),
			doc.Item("`cli`. Use `flag` from the standard library, or a library such as `cobra`."),
			doc.Item("`logs`. Use `app.Log` and `app.Logf`, which write to the browser console in the browser and to standard output on the server, or `log/slog`. Replace `app.DefaultLogger` to send them elsewhere."),
			doc.Item("`errors`, go-app's tagged errors. WebGo keeps a copy for its own use under `internal/`, which other modules cannot import. Use `fmt.Errorf` with `%w`, and `errors.Is` and `errors.As` from the standard library."),
		),
		doc.P("`cache`, `cli`, `errors` and `logs` do not import go-app's `app`, so they still compile next to WebGo if you keep go-app in `go.mod`. `ui` and `analytics` build on go-app's `app` and cannot be used with WebGo components."),
		doc.P("Errors that `app` returns are still go-app error values with tags. Their `Error()` text is JSON that includes the tags, but without the `errors` package the tags cannot be read one by one."),
	}
}
