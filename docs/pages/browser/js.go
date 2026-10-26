package browser

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func values() []app.UI {
	return []app.UI{
		doc.Go(`type Value interface {
	Bool() bool
	Int() int
	Float() float64
	String() string
	Truthy() bool
	Type() Type
	IsNull() bool
	IsUndefined() bool
	IsNaN() bool
	Equal(w Value) bool
	InstanceOf(t Value) bool

	Get(p string) Value
	Set(p string, x any)
	Delete(p string)
	Index(i int) Value
	SetIndex(i int, x any)
	Length() int

	Call(m string, args ...any) Value
	Invoke(args ...any) Value
	New(args ...any) Value
	Then(f func(Value))
	JSValue() Value
}

type Wrapper interface {
	JSValue() Value
}

func ValueOf(x any) Value
func Null() Value
func Undefined() Value`),
		doc.P("`Value` mirrors `syscall/js.Value`, but it compiles on every platform. `Value` also has unexported methods, so only the `app` package implements it."),
		doc.List(
			doc.Item("In the browser it wraps a JavaScript value. `Bool`, `Int` and `Float` panic if the value is not of that type. `String` never panics: for a value that is not a string it returns a description such as `<number: 42>`. Check with `Type`, `Truthy`, `IsNull` or `IsUndefined` first."),
			doc.Item("On the server it is an empty stub. `Get`, `Call`, `Index`, `New` and `Invoke` return another stub, `Bool`, `Int` and `Float` return zero, `String` returns `\"\"`, `Truthy` returns false, and `IsNull` and `IsUndefined` return true. `Type` panics. Code that reads JavaScript runs on the server without crashing, but sees nothing."),
			doc.Item("Arguments to `Set`, `SetIndex`, `Call`, `Invoke` and `New`, and the argument of `ValueOf`, are converted to JavaScript: a `Value` or `Wrapper` as its JavaScript value, `map[string]any` and `[]any` recursively, `nil` as `null`, and booleans, numbers and strings as themselves. Any other type panics in the browser."),
			doc.Item("A `Wrapper` can be passed wherever a JavaScript value is expected. HTML elements and components implement it: a component's `JSValue` is the DOM node of its root once it is mounted."),
		),
		doc.H("Window"),
		doc.Go(`func Window() BrowserWindow

type BrowserWindow interface {
	Value
	URL() *url.URL
	Size() (w, h int)
	CursorPosition() (x, y int)
	GetElementByID(id string) Value
	ScrollToID(id string)
}`),
		doc.P("`Window` is the global `window` object. `URL` parses `location.href`, `Size` is the inner width and height, `CursorPosition` is the pointer position of the last event an app event handler received, and `GetElementByID` is `document.getElementById`. On the server, `URL` is empty and the sizes are zero."),
		doc.H("The source element"),
		doc.Go(`func (ctx Context) Src() UI
func (ctx Context) JSSrc() Value`),
		doc.P("`Src` is the element or component that `ctx` belongs to, and `JSSrc` is its DOM node. In an event handler that is the element the handler is set on, which is how a handler reads a form control:"),
		doc.Go(`func (f *signup) onEmail(ctx app.Context, e app.Event) {
	f.email = ctx.JSSrc().Get("value").String()
}`),
		doc.H("IsClient and IsServer"),
		doc.Go(`const (
	IsClient = runtime.GOARCH == "wasm" && runtime.GOOS == "js"
	IsServer = runtime.GOARCH != "wasm" || runtime.GOOS != "js"
)`),
		doc.P("Branch on these constants for code that only makes sense on one side. Because they are constants, the compiler drops the branch that cannot run. The lifecycle already splits some of this: `OnMount` runs only in the browser, `OnPreRender` only on the server, and `OnNav` on both."),
	}
}

func calling() []app.UI {
	return []app.UI{
		doc.List(
			doc.Item("`v.Call(m, args...)` calls the method `m` of `v` and returns its result. `v.Invoke(args...)` calls `v` itself as a function, and `v.New(args...)` calls it as a constructor."),
			doc.Item("`v.Then(f)` treats `v` as a promise and calls `f` once with the value it resolves to. A rejection never calls `f`. `f` runs from the JavaScript event loop, not on the UI goroutine, so it hands results over with `ctx.Dispatch`."),
		),
		doc.Go(`lang := app.Window().Get("navigator").Get("language").String()

app.Window().Get("navigator").Get("clipboard").
	Call("writeText", "copied from Go").
	Then(func(app.Value) {
		app.Log("copied")
	})`),
		doc.H("Loading scripts"),
		doc.P("A JavaScript library is loaded like on any page. `Handler.Scripts` adds a `<script>` to the `<head>` of every page, and a path may be followed by a space and `async`, `defer` or `crossorigin`. `Handler.RawHeaders` inserts raw HTML into the `<head>`. Inside a component, `app.Script()` renders a `<script>` element."),
		doc.Go(`handler := &app.Handler{
	Name:    "Maps",
	Scripts: []string{"https://cdn.example.com/maps.js defer"},
}`),
		doc.H("FuncOf"),
		doc.Go(`type Func interface {
	Value
	Release()
}

func FuncOf(fn func(this Value, args []Value) any) Func`),
		doc.List(
			doc.Item("`FuncOf` wraps `fn` so JavaScript can call it, as a callback or an event listener. Its return value is converted like an argument to `ValueOf`."),
			doc.Item("`fn` does not run on the UI goroutine. Change fields inside `ctx.Dispatch`. The JavaScript event loop waits while `fn` runs, so `fn` must not block. Start blocking work with `ctx.Async`."),
			doc.Item("Call `Release` once JavaScript will no longer call the function, to free it. Calling a released function from JavaScript throws."),
			doc.Item("On the server, `FuncOf` returns a stub and `fn` never runs."),
		),
		doc.P("This component listens to the browser's `online` and `offline` events, and removes its listeners when it leaves the page:"),
		doc.Go(`type connection struct {
	app.Compo
	online  bool
	release func()
}

func (c *connection) OnMount(ctx app.Context) {
	c.online = app.Window().Get("navigator").Get("onLine").Bool()
	changed := app.FuncOf(func(this app.Value, args []app.Value) any {
		online := args[0].Get("type").String() == "online"
		ctx.Dispatch(func(ctx app.Context) {
			c.online = online
		})
		return nil
	})
	window := app.Window()
	window.Call("addEventListener", "online", changed)
	window.Call("addEventListener", "offline", changed)
	c.release = func() {
		window.Call("removeEventListener", "online", changed)
		window.Call("removeEventListener", "offline", changed)
		changed.Release()
	}
}

func (c *connection) OnDismount() {
	if c.release != nil {
		c.release()
	}
}

func (c *connection) Render() app.UI {
	if c.online {
		return app.Span().Text("Online")
	}
	return app.Span().Text("Offline")
}`),
		doc.H("KeepBodyClean"),
		doc.Go(`func KeepBodyClean() (close func())`),
		doc.P("Some third-party scripts insert elements into `<body>`, where they sit next to the app's own nodes. `KeepBodyClean` removes what they add until the returned `close` is called. On the server it does nothing."),
	}
}
