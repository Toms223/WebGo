//go:generate go run ../internal/htmlgen
//go:generate go fmt

// Package app builds web apps in Go. Components render HTML on the server and,
// compiled to WebAssembly, run in the browser as a progressive web app.
//
// One main serves both sides: register routes, call RunWhenOnBrowser, which
// takes over in the browser, then serve a Handler on the server.
//
//	type hello struct{ app.Compo }
//
//	func (h *hello) Render() app.UI { return app.H1().Text("Hello") }
//
//	func main() {
//		app.Route("/", func() app.Composer { return &hello{} })
//		app.RunWhenOnBrowser()
//		http.Handle("/", &app.Handler{Name: "Hello"})
//		log.Fatal(http.ListenAndServe(":8000", nil))
//	}
//
// Build the browser side with GOOS=js GOARCH=wasm into web/app.wasm.
package app

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"runtime"
)

const (
	// IsClient reports whether the code runs in the browser, compiled to
	// WebAssembly.
	IsClient = runtime.GOARCH == "wasm" && runtime.GOOS == "js"

	// IsServer reports whether the code runs outside the browser, including
	// during server-side rendering.
	IsServer = runtime.GOARCH != "wasm" || runtime.GOOS != "js"
)

var (
	routes = makeRouter()
	window = newBrowserWindow()
)

// Getenv returns the variable k, or "" when it is unset. In the browser it
// reads Handler.Env; on the server it reads the process environment, which the
// Handler fills from Env once it starts serving.
func Getenv(k string) string {
	if IsServer || !Window().Get("goappGetenv").Truthy() {
		return os.Getenv(k)
	}

	env := Window().Call("goappGetenv", k)
	if !env.Truthy() {
		return ""
	}
	return env.String()
}

// KeepBodyClean removes nodes that third-party scripts add to the body, until
// close is called. It is a no-op on the server.
func KeepBodyClean() (close func()) {
	if IsServer {
		return func() {}
	}

	release := Window().Call("goappKeepBodyClean")
	return func() {
		release.Invoke()
	}
}

// Window returns the JS window object. On the server it is an inert stub.
func Window() BrowserWindow {
	return window
}

// RunWhenOnBrowser starts the app in the browser and never returns. On the
// server it returns at once, so one main can register routes, call it, then
// serve a Handler. A panic during startup is shown in the loading screen.
func RunWhenOnBrowser() {
	if IsServer {
		return
	}

	defer func() {
		err := recover()
		displayLoadError(err)
		panic(err)
	}()

	resolveURL := clientResourceResolver(Getenv("GOAPP_STATIC_RESOURCES_URL"))
	originPage := makeRequestPage(Window().URL(), resolveURL)

	engine := newEngine(context.Background(),
		&routes,
		resolveURL,
		&originPage,
		actionHandlers,
	)

	engine.Navigate(window.URL(), false)
	engine.Start(120)
}

func displayLoadError(err any) {
	loadingLabel := Window().
		Get("document").
		Call("getElementById", "app-wasm-loader-label")
	if !loadingLabel.Truthy() {
		return
	}
	loadingLabel.setInnerText(fmt.Sprint(err))
}

// Route binds path to a component factory. The factory runs on each navigation
// to path; if the current page component has the same type, its exported fields
// are updated instead of remounting it.
func Route(path string, newComponent func() Composer) {
	routes.route(path, newComponent)
}

// RouteWithRegexp is like Route but matches paths against the regular
// expression pattern. Exact routes from Route win; patterns are tried in
// registration order, and an invalid pattern panics.
func RouteWithRegexp(pattern string, newComponent func() Composer) {
	routes.routeWithRegexp(pattern, newComponent)
}

// NewZeroComponentFactory returns a factory that creates zero values of c's
// type. c must be a pointer, such as &hello{}.
func NewZeroComponentFactory(c Composer) func() Composer {
	componentType := reflect.TypeOf(c)

	return func() Composer {
		return reflect.New(componentType.Elem()).Interface().(Composer)
	}
}

// TryUpdate asks the browser to check for a new app version now. The page
// already checks on load, when it comes back online or visible, and every
// minute. It is a no-op on the server.
func TryUpdate() {
	if tryUpdate := Window().Get("goappTryUpdate"); IsClient && tryUpdate.Truthy() {
		tryUpdate.Invoke()
	}
}
