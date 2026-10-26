package app

import (
	"reflect"
	"strings"

	"github.com/Toms223/WebGo/internal/errors"
)

// Composer is a component: a pointer to a struct that embeds Compo and
// implements Render.
type Composer interface {
	UI

	// Render returns the component's UI. It runs on every update, so keep it
	// free of side effects.
	Render() UI

	setRef(Composer) Composer
	depth() uint
	setDepth(uint) Composer
	parent() UI
	root() UI
	setRoot(UI) Composer
}

// Initializer is implemented by components that set up fields before their
// first render. OnInit runs on the server and in the browser, before any other
// hook.
type Initializer interface {
	OnInit()
}

// PreRenderer is implemented by components that prepare server-side rendering,
// for example by setting the page title. OnPreRender runs only on the server.
type PreRenderer interface {
	OnPreRender(Context)
}

// Mounter is implemented by components that act once they are in the DOM.
// OnMount runs only in the browser, on the UI goroutine.
type Mounter interface {
	OnMount(Context)
}

// Dismounter is implemented by components that clean up when they leave the
// tree. OnDismount runs on the UI goroutine.
type Dismounter interface {
	OnDismount()
}

// Navigator is implemented by components notified of navigation. OnNav runs on
// every mounted component after each navigation, including the first load,
// server-side rendering and fragment-only changes.
type Navigator interface {
	OnNav(Context)
}

// DismountEnforcer forces a remount: when CompoID differs between the current
// and the new component, the current one is dismounted instead of updated in
// place.
type DismountEnforcer interface {
	CompoID() string
}

// Updater is implemented by components that react when their parent changes
// their exported fields. OnUpdate runs on the UI goroutine.
type Updater interface {
	OnUpdate(Context)
}

// AppUpdater is implemented by components notified that a new app version is
// installed. Unless Handler.DisableAutoUpdate is set, the page reloads right
// after, so the hook has little time to show anything.
type AppUpdater interface {
	OnAppUpdate(Context)
}

// AppInstaller is implemented by components notified when the app becomes
// installable or gets installed; see Context.IsAppInstallable.
type AppInstaller interface {
	OnAppInstallChange(Context)
}

// Resizer is implemented by components notified when the window resizes or
// Context.ResizeContent is called on an ancestor. OnResize runs on the UI
// goroutine.
type Resizer interface {
	OnResize(Context)
}

// Compo is embedded in every component struct. It provides the Composer
// plumbing and a placeholder Render.
type Compo struct {
	treeDepth     uint
	ref           Composer
	parentElement UI
	rootElement   UI
}

// JSValue returns the DOM node of the component's root element, or a null value
// when it has none.
func (c *Compo) JSValue() Value {
	if c.rootElement == nil {
		return ValueOf(nil)
	}
	return c.rootElement.JSValue()
}

// Mounted reports whether the component is in the tree. Dispatch and Defer skip
// unmounted components.
func (c *Compo) Mounted() bool {
	return c.ref != nil
}

// Render is a placeholder that tells you to implement Render. Components
// override it.
func (c *Compo) Render() UI {
	componentName := reflect.TypeOf(c.ref).Name()

	return Div().
		DataSet("compo-type", componentName).
		RawStyle("border", "1px solid currentColor").
		RawStyle("padding", "12px 0").
		Body(
			H1().Text("Component "+strings.TrimPrefix(componentName, "*")),
			P().Body(
				Text("Change appearance by implementing: "),
				Code().
					RawStyle("color", "deepskyblue").
					RawStyle("margin", "0 6px").
					Text("func (c "+componentName+") Render() app.UI"),
			),
		)
}

// ValueTo returns an EventHandler that stores the source element's value in v,
// a pointer to a string, integer, unsigned integer or float64. A conversion
// error is logged and leaves v unchanged.
func (c *Compo) ValueTo(v any) EventHandler {
	return func(ctx Context, e Event) {
		value := ctx.JSSrc().Get("value")
		if err := stringTo(value.String(), v); err != nil {
			Log(errors.New("storing dom element value failed").Wrap(err))
			return
		}
	}
}

func (c *Compo) setRef(v Composer) Composer {
	c.ref = v
	return v
}

func (c *Compo) depth() uint {
	return c.treeDepth
}

func (c *Compo) setDepth(v uint) Composer {
	c.treeDepth = v
	return c.ref
}

func (c *Compo) parent() UI {
	return c.parentElement
}

func (c *Compo) setParent(p UI) UI {
	c.parentElement = p
	return c.ref
}

func (c *Compo) root() UI {
	return c.rootElement
}

func (c *Compo) setRoot(v UI) Composer {
	c.rootElement = v
	return c.ref
}
