package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func updates() []app.UI {
	return []app.UI{
		doc.P("A WebGo app is a progressive web app. A service worker stores its files in the browser, so later visits load from that cache and work offline. The version decides when the cache is replaced."),
		doc.Go(`Version           string
DisableAutoUpdate bool

func TryUpdate()
func (ctx Context) AppUpdateAvailable() bool
func (ctx Context) Reload()

type AppUpdater interface {
	OnAppUpdate(Context)
}`),
		doc.H("Version"),
		doc.P("When `Version` is empty, the handler computes it as a hash of what the browser would download: the handler's settings, the generated scripts and stylesheet, the resolved style links, and the bytes of every local `/web/` file the page uses (`app.wasm`, the icons, `Image`, styles, scripts, fonts and `CacheableResources`). A file that is missing hashes as missing."),
		doc.List(
			doc.Item("Restarting the server, or running several replicas of the same build, gives the same version. Leaving `Version` empty is the recommended setting in production."),
			doc.Item("Rebuilding `app.wasm`, or changing any file the page uses, gives a new version."),
			doc.Item("A `/web/` file the page does not reference, such as an image only used in an `<img>`, is not part of the hash. Changing it alone does not update installed clients."),
		),
		doc.P("Setting `Version` yourself, for example to a commit hash, replaces the computed one. The version names the worker's cache (`app-<version>`) and is the `ETag` of the handler's responses."),
		doc.H("What happens on a deploy"),
		updatesDiagram(),
		doc.Ordered(
			doc.Item("Every page carries a small inline script that asks the browser to check for a new service worker: when the page loads, when the browser comes back online, when the tab becomes visible, and every 60 seconds. It skips the check while the tab is hidden or offline."),
			doc.Item("The new `/app-worker.js` contains the new version, so the browser installs it. The worker downloads every file it stores, bypassing the HTTP cache. If any of them fails, for example with a 404 or a redirect, it deletes the new cache and the install fails: the old version keeps running, and the next check tries again."),
			doc.Item("The new worker activates at once. It deletes the other `app-` caches, keeps caches with any other name, and takes control of the open pages."),
			doc.Item("Each open page that was controlled by the old worker reloads itself once, and runs the new version. A first visit, where no worker was in control yet, does not reload."),
		),
		doc.P("The worker stores `/`, `/app.css`, `/app.js`, `/manifest.webmanifest`, `/wasm_exec.js`, `/web/app.wasm`, the icons, the fonts, the stylesheet links, the scripts and `CacheableResources`, all resolved through `Resources`. It answers only `GET` requests for those same-origin URLs, from its cache and else from the network. Other requests, cross-origin files and other page routes go to the network as usual."),
		doc.P("The update checker is an inline `<script>`. A Content Security Policy without `'unsafe-inline'` or a matching hash blocks it: automatic updates stop, and the rest of the app works."),
		doc.H("DisableAutoUpdate"),
		doc.P("`DisableAutoUpdate: true` stops step 4, the reload. The checks still run and the new worker still takes control, so the open page keeps running the old `app.wasm` with the new worker until it reloads. Use it when a reload would lose work, such as a half-filled form, and let the user choose when:"),
		doc.Go(`type shell struct {
	app.Compo
	updated bool
}

func (s *shell) OnAppUpdate(ctx app.Context) {
	s.updated = ctx.AppUpdateAvailable()
}

func (s *shell) Render() app.UI {
	return app.Div().Body(
		app.If(s.updated, func() app.UI {
			return app.Button().Text("Reload to update").OnClick(s.reload)
		}),
	)
}

func (s *shell) reload(ctx app.Context, e app.Event) {
	ctx.Reload()
}

func main() {
	app.Route("/", func() app.Composer { return &shell{} })
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{
		Name:              "Notes",
		DisableAutoUpdate: true,
	})
	log.Fatal(http.ListenAndServe(":8080", nil))
}`),
		doc.List(
			doc.Item("`OnAppUpdate` is called on every mounted component that implements `AppUpdater` once the new worker is active, and `ctx.AppUpdateAvailable()` reports `true` from then on. If the update finished before `app.wasm` started, the call comes as soon as it does."),
			doc.Item("Without `DisableAutoUpdate`, the reload usually comes first, so an update button is rarely seen."),
			doc.Item("`ctx.Reload()` reloads the page. It does nothing on the server."),
			doc.Item("`app.TryUpdate()` runs a check now. The automatic checks make it rarely needed. It does nothing on the server or before the worker is registered."),
		),
	}
}

func installing() []app.UI {
	return []app.UI{
		doc.P("Every WebGo app can be installed, because the handler always serves a manifest and a service worker. The manifest takes `Name`, `ShortName`, `Description`, `Icon`, `BackgroundColor`, `ThemeColor` and `StartURL` from the handler, and opens the app in its own window (`display: standalone`)."),
		doc.P("Users install it from the browser: the install button in the address bar of Chrome and Edge, the browser menu on Android, or Share and then Add to Home Screen in Safari on iPhone and iPad. Chromium browsers also let the app offer the install itself:"),
		doc.Go(`func (ctx Context) IsAppInstallable() bool
func (ctx Context) IsAppleBrowser() bool
func (ctx Context) ShowAppInstallPrompt()

type AppInstaller interface {
	OnAppInstallChange(Context)
}`),
		doc.List(
			doc.Item("`IsAppInstallable` reports `true` when the app is not already running installed, and either the browser has offered an install prompt or the browser is Safari or an iOS or iPadOS browser. It reports `false` on the server."),
			doc.Item("`IsAppleBrowser` reports whether the browser is Safari on macOS, or any browser on iPhone or iPad."),
			doc.Item("`ShowAppInstallPrompt` shows the browser's install dialog, once per offer. It does nothing when `IsAppInstallable` is `false`, or when there is no dialog: Apple browsers report `IsAppInstallable` as `true` but install through the Share menu, so check `IsAppleBrowser` and show instructions there."),
			doc.Item("`OnAppInstallChange` is called on every mounted component that implements `AppInstaller` when the browser offers the prompt and after the app is installed. If that happened before `app.wasm` started, the call comes as soon as it does."),
		),
		doc.Go(`type installButton struct {
	app.Compo
	installable bool
	apple       bool
}

func (b *installButton) OnMount(ctx app.Context) {
	b.installable = ctx.IsAppInstallable()
	b.apple = ctx.IsAppleBrowser()
}

func (b *installButton) OnAppInstallChange(ctx app.Context) {
	b.installable = ctx.IsAppInstallable()
}

func (b *installButton) Render() app.UI {
	switch {
	case !b.installable:
		return app.Span()
	case b.apple:
		return app.P().Text("Tap Share, then Add to Home Screen.")
	default:
		return app.Button().Text("Install").OnClick(b.install)
	}
}

func (b *installButton) install(ctx app.Context, e app.Event) {
	ctx.ShowAppInstallPrompt()
}`),
		doc.P("`OnMount` reads the state once, because it runs only in the browser. `OnAppInstallChange` keeps it current, and hides the button once the app is installed."),
	}
}
