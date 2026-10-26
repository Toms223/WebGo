package browser

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func storage() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) LocalStorage() BrowserStorage
func (ctx Context) SessionStorage() BrowserStorage

type BrowserStorage interface {
	Set(k string, v any) error
	Get(k string, v any) error
	Del(k string)
	Len() int
	Clear()
	ForEach(f func(k string))
	Contains(k string) bool
}`),
		doc.P("`LocalStorage` is the browser's `localStorage`, kept across visits for the origin. `SessionStorage` is `sessionStorage`, kept for one tab until it is closed. Both store JSON."),
		doc.List(
			doc.Item("`Set` encodes `v` as JSON and stores it. It returns an error if encoding fails or the browser refuses the write, for example when storage is full or disabled."),
			doc.Item("`Get` decodes the stored JSON into `v`, which must be a pointer. A missing key leaves `v` unchanged and returns nil. A value that is not JSON, such as one written by another script, returns a decoding error."),
			doc.Item("`Del` removes a key, and does nothing if it is missing. `Contains` reports whether a key exists, and `Len` counts the keys."),
			doc.Item("`Clear` removes every key of the origin, including persisted states and the device ID that `PersistWithEncryption` depends on."),
			doc.Item("`ForEach` calls `f` with each key, in no particular order. In the browser the keys are collected first, so `f` can delete them."),
			doc.Item("On the server and in tests, both are in-memory stores that belong to one engine, so nothing outlives a request or a test."),
		),
		doc.Go(`type settings struct {
	Theme string
	Size  int
}

type preferences struct {
	app.Compo
	settings settings
}

func (p *preferences) OnMount(ctx app.Context) {
	p.settings = settings{Theme: "light", Size: 16}
	if err := ctx.LocalStorage().Get("preferences", &p.settings); err != nil {
		app.Log(err)
	}
}

func (p *preferences) save(ctx app.Context, e app.Event) {
	if err := ctx.LocalStorage().Set("preferences", p.settings); err != nil {
		app.Log(err)
	}
}`),
		doc.P("When other components need the value too, use a state with `Persist` instead, under ", doc.A("/actions#persisting-and-broadcasting", "Persisting and broadcasting"), ". It writes to the same local storage, and observers follow the changes."),
	}
}

func notifications() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) Notifications() NotificationService

func (s NotificationService) Permission() NotificationPermission
func (s NotificationService) RequestPermission() NotificationPermission
func (s NotificationService) New(n Notification)
func (s NotificationService) Subscribe(vapIDPublicKey string) (NotificationSubscription, error)

type NotificationPermission string

const (
	NotificationDefault      NotificationPermission = "default"
	NotificationGranted      NotificationPermission = "granted"
	NotificationDenied       NotificationPermission = "denied"
	NotificationNotSupported NotificationPermission = "unsupported"
)`),
		doc.List(
			doc.Item("`Permission` reads the current permission without asking. `NotificationDefault` means the user has not answered yet, and the browser treats it as denied. `NotificationNotSupported` means there is no Notification API, which is always the case on the server."),
			doc.Item("`RequestPermission` shows the browser's prompt and blocks the calling goroutine until the user answers. Browsers only show it in response to a user action, so call it from an event handler. The page does not re-render while it waits."),
			doc.Item("`New` shows a notification. When the app's service worker is active, which is normally the case once the app has loaded, the worker shows it. Otherwise the page creates it directly."),
		),
		doc.H("Notification"),
		doc.Go(`type Notification struct {
	Title              string
	Path               string
	Lang               string
	Badge              string
	Body               string
	Tag                string
	Icon               string
	Image              string
	Data               map[string]any
	Renotify           bool
	RequireInteraction bool
	Silent             bool
	Vibrate            []int
	Actions            []NotificationAction
}

type NotificationAction struct {
	Action string
	Title  string
	Icon   string
	Path   string
}`),
		doc.List(
			doc.Item("Clicking the notification focuses the app and navigates to `Path`. Clicking one of its `Actions` navigates to that action's `Path` instead."),
			doc.Item("The service worker uses `Title`, `Body`, `Icon`, `Badge`, `Path` and `Actions`, and ignores the other fields. They apply only when the page creates the notification itself, and that path cannot show `Actions`."),
			doc.Item("The `goapp` key of `Data` is reserved."),
		),
		doc.Go(`type alerts struct {
	app.Compo
	permission app.NotificationPermission
}

func (a *alerts) OnMount(ctx app.Context) {
	a.permission = ctx.Notifications().Permission()
}

func (a *alerts) Render() app.UI {
	return app.Div().Body(
		app.If(a.permission == app.NotificationDefault, func() app.UI {
			return app.Button().Text("Enable alerts").OnClick(a.enable)
		}).ElseIf(a.permission == app.NotificationGranted, func() app.UI {
			return app.Button().Text("Send a test").OnClick(a.test)
		}).Else(func() app.UI {
			return app.Text("Alerts are off")
		}),
	)
}

func (a *alerts) enable(ctx app.Context, e app.Event) {
	a.permission = ctx.Notifications().RequestPermission()
}

func (a *alerts) test(ctx app.Context, e app.Event) {
	ctx.Notifications().New(app.Notification{
		Title: "Build finished",
		Body:  "All checks passed.",
		Path:  "/builds",
	})
}`),
		doc.H("Push"),
		doc.Go(`type NotificationSubscription struct {
	Endpoint string
	Keys     struct {
		Auth   string
		P256dh string
	}
}`),
		doc.List(
			doc.Item("`Subscribe` subscribes the app's service worker to push messages with your VAPID public key, and returns the subscription to send to your server. It needs the permission granted and an active service worker, and blocks until the browser answers."),
			doc.Item("It returns an error if the key is empty, or if the browser cannot subscribe. The browser logs the reason to its console."),
			doc.Item("Your server sends a push message whose payload is a JSON-encoded `app.Notification`, with any Web Push library and your VAPID private key. The service worker that `app.Handler` generates shows it, with the same fields and click behaviour as `New`. A payload that is not JSON is ignored."),
			doc.Item("A custom worker set with `Handler.ServiceWorkerTemplate` must keep the `push` and `notificationclick` handlers for this to work."),
		),
		doc.Go(`func (a *alerts) subscribe(ctx app.Context) {
	ctx.Async(func() {
		sub, err := ctx.Notifications().Subscribe(app.Getenv("VAPID_PUBLIC_KEY"))
		if err != nil {
			app.Log(err)
			return
		}
		body, err := json.Marshal(sub)
		if err != nil {
			app.Log(err)
			return
		}
		res, err := http.Post("/push/register", "application/json", bytes.NewReader(body))
		if err != nil {
			app.Log(err)
			return
		}
		res.Body.Close()
	})
}`),
		doc.P("The public key reaches the browser through `Handler.Env`, under ", doc.A("#environment-and-logging", "Environment and logging"), ". `/push/register` is a handler of your own that stores subscriptions."),
	}
}

func environment() []app.UI {
	return []app.UI{
		doc.Go(`func Getenv(k string) string`),
		doc.P("`Handler.Env` is a `map[string]string` of values for both sides of the app."),
		doc.List(
			doc.Item("When the handler initializes, on its first request, each pair of `Env` is set in the server's environment with `os.Setenv`, and the whole map is written into `/app.js`."),
			doc.Item("`Getenv` reads `os.Getenv` on the server. In the browser it reads the map from `/app.js`, and returns `\"\"` for a missing key."),
			doc.Item("`/app.js` is public. Never put a secret in `Env`."),
			doc.Item("The handler sets `GOAPP_VERSION`, `GOAPP_STATIC_RESOURCES_URL`, `GOAPP_ROOT_PREFIX` and `GOAPP_INTERNAL_URLS` itself, replacing any value of yours with those names."),
		),
		doc.Go(`handler := &app.Handler{
	Name: "Alerts",
	Env:  map[string]string{"VAPID_PUBLIC_KEY": publicKey},
}`),
		doc.H("Logging"),
		doc.Go(`var DefaultLogger func(format string, v ...any)

func Log(v ...any)
func Logf(format string, v ...any)`),
		doc.List(
			doc.Item("`Logf` formats like `fmt.Printf` and passes the result to `DefaultLogger`. `Log` joins its operands with spaces, like `fmt.Println` without the newline."),
			doc.Item("In the browser, `DefaultLogger` writes to the console, with `console.error` when one of the operands is an `error` and `console.log` otherwise."),
			doc.Item("On the server it prints to standard output, prefixed `ERROR ‣` or `INFO ‣` by the same rule, in color."),
			doc.Item("`DefaultLogger` is a variable. Replace it, early in `main`, to send the app's logs elsewhere. Errors that `app` handles internally, such as a failed `GetState`, go through it too."),
		),
		doc.Go(`func main() {
	app.DefaultLogger = log.Printf
}`),
	}
}
