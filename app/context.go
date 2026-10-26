package app

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/Toms223/WebGo/internal/errors"
	"github.com/google/uuid"
)

// Context is handed to event handlers and lifecycle hooks, bound to the UI
// element they belong to. It drives updates, navigation, actions and state. Its
// embedded context.Context is never canceled.
type Context struct {
	context.Context

	page                  func() Page
	appUpdatable          bool
	resolveURL            func(string) string
	navigate              func(*url.URL, bool)
	localStorage          BrowserStorage
	sessionStorage        BrowserStorage
	dispatch              func(func())
	defere                func(func())
	async                 func(func())
	addComponentUpdate    func(Composer, int)
	removeComponentUpdate func(Composer)
	handleAction          func(string, UI, bool, ActionHandler)
	postAction            func(Context, Action)
	observeState          func(Context, string, any) Observer
	getState              func(Context, string, any)
	setState              func(Context, string, any) State
	delState              func(Context, string)

	sourceElement        UI
	notifyComponentEvent func(Context, UI, any)
}

func (ctx Context) Src() UI {
	return ctx.sourceElement
}

func (ctx Context) JSSrc() Value {
	return ctx.sourceElement.JSValue()
}

// AppUpdateAvailable reports whether a newer app version is installed and
// waiting for a reload.
func (ctx Context) AppUpdateAvailable() bool {
	return ctx.appUpdatable
}

// IsAppInstallable reports whether the browser offers to install the app. It is
// always false on the server.
func (ctx Context) IsAppInstallable() bool {
	if Window().Get("goappIsAppInstallable").Truthy() {
		return Window().Call("goappIsAppInstallable").Bool()
	}
	return false
}

func (ctx Context) IsAppleBrowser() bool {
	if Window().Get("goappIsAppleBrowser").Truthy() {
		return Window().Call("goappIsAppleBrowser").Bool()
	}
	return false
}

// ShowAppInstallPrompt opens the browser's install prompt, once per offer. It
// does nothing when the app is not installable or there is no prompt, as on
// Apple browsers, which install through the Share menu (see IsAppleBrowser).
func (ctx Context) ShowAppInstallPrompt() {
	if ctx.IsAppInstallable() {
		Window().Call("goappShowInstallPrompt")
	}
}

// DeviceID returns a random ID created on first use and kept in local storage.
// It panics if local storage fails.
func (ctx Context) DeviceID() string {
	var id string
	if err := ctx.localStorage.Get("/go-app/deviceID", &id); err != nil {
		panic(errors.New("retrieving device id failed").Wrap(err))
	}
	if id != "" {
		return id
	}

	id = uuid.NewString()
	if err := ctx.localStorage.Set("/go-app/deviceID", id); err != nil {
		panic(errors.New("creating device id failed").Wrap(err))
	}
	return id
}

func (ctx Context) Page() Page {
	return ctx.page()
}

// Reload reloads the page. It is a no-op on the server.
func (ctx Context) Reload() {
	if IsServer {
		return
	}
	Window().Get("location").Call("reload")
}

// Navigate goes to rawURL like a link click: paths on the same host are routed
// in the app, other hosts open in a new tab, and mailto:, tel: and
// Handler.InternalURLs load in the current tab. A parse error is logged.
func (ctx Context) Navigate(rawURL string) {
	u, err := url.Parse(rawURL)
	if err != nil {
		Log(errors.New("navigating to URL failed").
			WithTag("url", rawURL).
			Wrap(err))
		return
	}
	ctx.NavigateTo(u)
}

func (ctx Context) NavigateTo(u *url.URL) {
	ctx.navigate(u, true)
}

// ResolveStaticResource maps a /web/ path through Handler.Resources, for
// example to a bucket URL. Other paths are returned as is.
func (ctx Context) ResolveStaticResource(v string) string {
	return ctx.resolveURL(v)
}

func (ctx Context) ScrollTo(id string) {
	ctx.Defer(func(ctx Context) {
		Window().ScrollToID(id)
	})
}

func (ctx Context) LocalStorage() BrowserStorage {
	return ctx.localStorage
}

func (ctx Context) SessionStorage() BrowserStorage {
	return ctx.sessionStorage
}

// Encrypt encodes v as JSON and encrypts it with AES-GCM, keyed by DeviceID, so
// only this browser can decrypt it.
func (ctx Context) Encrypt(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, errors.New("encoding value failed").Wrap(err)
	}

	b, err = encrypt(ctx.cryptoKey(), b)
	if err != nil {
		return nil, errors.New("encrypting value failed").Wrap(err)
	}
	return b, nil
}

// Decrypt reverses Encrypt, decoding the JSON into v, a pointer.
func (ctx Context) Decrypt(crypted []byte, v any) error {
	b, err := decrypt(ctx.cryptoKey(), crypted)
	if err != nil {
		return errors.New("decrypting value failed").Wrap(err)
	}

	if err := json.Unmarshal(b, v); err != nil {
		return errors.New("decoding value failed").Wrap(err)
	}
	return nil
}

func (ctx Context) cryptoKey() string {
	return strings.ReplaceAll(ctx.DeviceID(), "-", "")
}

func (ctx Context) Notifications() NotificationService {
	return NotificationService{}
}

// Dispatch runs v on the UI goroutine, then re-renders the source component and
// its ancestors. It does nothing once the source is unmounted. v may be nil.
func (ctx Context) Dispatch(v func(Context)) {
	ctx.dispatch(func() {
		if !ctx.sourceElement.Mounted() {
			return
		}

		for c, ok := component(ctx.sourceElement); ok; c, ok = component(c.parent()) {
			ctx.addComponentUpdate(c, 1)
		}

		if v != nil {
			v(ctx)
		}
	})
}

// Defer runs v on the UI goroutine after the current render, without scheduling
// one. It does nothing once the source is unmounted.
func (ctx Context) Defer(v func(Context)) {
	ctx.defere(func() {
		if !ctx.sourceElement.Mounted() {
			return
		}

		if v != nil {
			v(ctx)
		}
	})
}

// Async runs v in a new goroutine. Server-side rendering waits for it before
// sending the page. Use Dispatch from v to change component fields.
func (ctx Context) Async(v func()) {
	ctx.async(v)
}

// After waits d in a goroutine, then calls Dispatch(f).
func (ctx Context) After(d time.Duration, f func(Context)) {
	ctx.async(func() {
		time.Sleep(d)
		ctx.Dispatch(f)
	})
}

// PreventUpdate cancels the re-render scheduled for the source component and
// its ancestors, for a handler that changed nothing visible.
func (ctx Context) PreventUpdate() {
	for c, ok := component(ctx.sourceElement); ok; c, ok = component(c.parent()) {
		ctx.addComponentUpdate(c, -1)
	}
}

// Update re-renders the source component and its ancestors. It is
// Dispatch(nil).
func (ctx Context) Update() {
	ctx.Dispatch(nil)
}

// Handle registers h for the named action on the source component. h runs on
// the UI goroutine and is dropped once the component is unmounted.
func (ctx Context) Handle(action string, h ActionHandler) {
	ctx.handleAction(action, ctx.sourceElement, false, h)
}

// NewAction posts the named action to the handlers registered with Handle and
// Context.Handle. Tags are merged into Action.Tags.
func (ctx Context) NewAction(action string, tags ...Tagger) {
	ctx.NewActionWithValue(action, nil, tags...)

}

// NewActionWithValue is NewAction with v as Action.Value.
func (ctx Context) NewActionWithValue(action string, v any, tags ...Tagger) {
	var tagMap Tags
	for _, tag := range tags {
		if tagMap == nil {
			tagMap = make(Tags)
		}

		for k, v := range tag.Tags() {
			tagMap[k] = v
		}
	}

	ctx.postAction(ctx, Action{
		Name:  action,
		Value: v,
		Tags:  tagMap,
	})
}

// ObserveState stores the value of state in recv, a pointer, and keeps it
// current: each SetState updates recv and re-renders the source component,
// until it is unmounted. Tune it with the returned Observer.
func (ctx Context) ObserveState(state string, recv any) Observer {
	return ctx.observeState(ctx, state, recv)
}

// GetState stores the value of state in recv, a pointer, falling back to the
// persisted copy. A missing or expired state leaves recv unchanged.
func (ctx Context) GetState(state string, recv any) {
	ctx.getState(ctx, state, recv)
}

// SetState sets state to v and updates its observers. Chain the returned State
// to persist, expire or broadcast it.
func (ctx Context) SetState(state string, v any) State {
	return ctx.setState(ctx, state, v)
}

// DelState deletes state and its persisted copy. Observers stay registered and
// receive the next SetState.
func (ctx Context) DelState(state string) {
	ctx.delState(ctx, state)
}

// ResizeContent calls OnResize on every Resizer at or below the source element,
// after the current render.
func (ctx Context) ResizeContent() {
	ctx.Defer(func(ctx Context) {
		ctx.Dispatch(func(ctx Context) {
			ctx.notifyComponentEvent(ctx, ctx.Src(), resize{})
		})
	})
}
