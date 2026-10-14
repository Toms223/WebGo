package applying

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/applying/demos"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func run() []app.UI {
	return []app.UI{
		doc.Go(`func Run(styles ...Style)`),
		doc.P("`Run` is the browser entry point. It returns at once when the program runs on the server (`app.IsServer`), so it is safe to call everywhere. Call it exactly once per app, in `main`, before `app.RunWhenOnBrowser()`:"),
		doc.Go(`func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	style.Run(theme.Styles()...)
	app.RunWhenOnBrowser()
}`),
		doc.P("`Run` only looks at the `Dynamic` styles, so give it the same list as the handler. Keep one function, here `theme.Styles() []style.Style`, that both the handler and `Run` consume, so the server and the browser see the same items."),
		doc.P("A second call collides with the first. `app.Handle` keeps one global handler per action name, so a second `Run` with an item bound to an action name the first one used replaces the first engine's handler for that name, and the items the first engine bound to it stop reacting. Give `Run` every `Dynamic` style the app has, in one call."),
		doc.P("On the client, `Run` does this:"),
		doc.List(
			doc.Item("It registers one global go-app action handler per distinct action name bound in the `Dynamic` styles, through `app.Handle`. The engine fans out to every item bound to that name, in the order the items were passed to `Run`."),
			doc.Item("go-app runs `app.Handle` handlers off the UI goroutine, so the handler only calls `ctx.Dispatch`. Evaluating a binding and changing the page happen there, one action at a time."),
			doc.Item("The engine owns exactly one `<style data-webgo=\"dynamic\">` element. It is created in `document.head` on the first insertion attempt, and the engine uses its `.sheet`."),
			doc.Item("Inserting an item inserts each block of its text at the end of that sheet, with `insertRule`, as one CSSOM rule per block: a flat rule with a base block, a `Hover` block and a `Media` block is three. A rule with nested children is one block, and so is a `KeyframesRule`. Removing an item deletes its CSSOM rules with `deleteRule`, and every other item's position stays correct."),
			doc.Item("The browser can refuse a block: `insertRule` throws a `SyntaxError` when the block does not parse as one rule, for example for a selector or pseudo-class the browser does not accept, or a `css.Raw` value that closes the block early. A merely invalid value is dropped by the browser without an exception. The engine logs the exception through `app.Logf` and leaves that block out. The item's other blocks stay. An item with no inserted block is absent and is tried again the next time an action asks for it. The app does not crash."),
			doc.Item("Items with `Initially(true)` are inserted during `Run`."),
		),
		doc.P("Dynamic items need the page's Content-Security-Policy to allow inline styles, with a `style-src` that permits the injected `<style>`. When the browser gives the element no sheet, every insertion attempt is logged through `app.Logf` and the app keeps running with the item absent. `Static` and `Generated` stylesheets are linked files and are unaffected."),
		doc.H("Action names belong to WebGo"),
		doc.P("`app.Handle` is a plain map assignment in go-app v11.0.5. A later `app.Handle` call on the same name silently replaces an earlier one, in either direction. Action names bound through `On` and `Bind` belong to WebGo: do not register your own handler for them with `app.Handle`. Whichever call runs last wins and the other handler is lost."),
		doc.P("Register your own handlers for those names with `ctx.Handle` from a component. Component-level handlers are keyed separately from the global map, so they run alongside the engine's. With `theme.DarkAction` and the `toggle` component from Two-way toggle under ", doc.A("#actions", "Actions"), ":"),
		doc.Go(`func (t *toggle) OnMount(ctx app.Context) {
	ctx.Handle(theme.DarkAction, t.onDark)
}

func (t *toggle) onDark(ctx app.Context, a app.Action) {
	if dark, ok := a.Value.(bool); ok {
		t.dark = dark
		ctx.Update()
	}
}`),
		doc.P("The global map is read once, when go-app's engine loads, and `app.RunWhenOnBrowser` does not return in the browser. Call `Run` before it."),
	}
}

func actions() []app.UI {
	return []app.UI{
		doc.P("A bound item reacts to go-app actions. `Bind` takes a typed `Binding`: the action's value must be a `*T`, and the function receives it. `On` takes the v1 predicate over the raw `app.Action`. Each time an action with a bound name fires, the engine evaluates the item and inserts it, removes it, or replaces its text. An action whose name no item is bound to is ignored."),
		doc.Go(`func When[T any](action string, when func(*T) bool) Binding
func Compute[T any](action string, props func(*T) []css.Property) Binding
func ComputeFrames[T any](action string, frames func(*T) []Frame) Binding

func (r *Rule) On(action string, when func(app.Action) bool) *Rule
func (r *Rule) Bind(b Binding) *Rule
func (r *Rule) Initially(present bool) *Rule
func (r *Rule) CSSFor(action string, v any) string

func (k *KeyframesRule) On(action string, when func(app.Action) bool) *KeyframesRule
func (k *KeyframesRule) Bind(b Binding) *KeyframesRule
func (k *KeyframesRule) Initially(present bool) *KeyframesRule
func (k *KeyframesRule) CSSFor(action string, v any) string

func FrameAt(stop css.Percentage, props ...css.Property) Frame
func FrameFrom(props ...css.Property) Frame
func FrameTo(props ...css.Property) Frame`),
		doc.P("Each demo below keeps its state in a `data.ViewModel` whose key is the action name. `OnAction` turns every change of that state into an action carrying the `*T`. See ", doc.A("#viewmodel-and-onaction", "ViewModel and OnAction"), "."),
		doc.H("When"),
		doc.P("`When` decides presence. Its function's result says whether the rule is present. The rule's declared blocks are what is inserted."),
		doc.Example(demos.WhenExample),
		doc.P("`DemoCompactBase` is in a Generated file, inside `@layer components`. `DemoCompact` is in `Dynamic` and shares the selector. It wins while present because the dynamic sheet is unlayered, and unlayered rules beat layered ones. Without the layer it would still win: the dynamic sheet comes after the generated one."),
		doc.H("Compute"),
		doc.P("`Compute` decides the properties. Its function returns the rule's base block. The rule's pseudo blocks, conditional blocks and nested children stay as declared. A `nil` or empty result makes the rule absent."),
		doc.Example(demos.ComputeExample),
		doc.P("The CSS is `DemoAccent.CSSFor(AccentKey, &Accent{Color: \"#FF3B30\"})`, the demo's starting colour."),
		doc.H("ComputeFrames"),
		doc.P("`ComputeFrames` is `Compute` for a `KeyframesRule`: its function returns the frames. `FrameAt`, `FrameFrom` and `FrameTo` build one frame each. A `FrameAt` stop outside 0% to 100% gives an empty frame, which is not written."),
		doc.Example(demos.FramesExample),
		doc.P("The slider sets `Beats`. `DemoPulseBox` is in a Generated file and names the keyframes. `DemoPulse` is in `Dynamic` and is replaced on every change."),
		doc.H("Initially"),
		doc.P("`Initially(true)` inserts the item when `Run` starts, before any action fires. The default is absent. There is no `Default` method: `Default` is the `:default` pseudo-class."),
		doc.Example(demos.InitiallyExample),
		doc.P("`Initially(true)` on a `Compute` rule inserts its declared `Set` block before any action. With no `Set` block it inserts nothing until the first action."),
		doc.H("On"),
		doc.P("`On` is the untyped form. Its function receives the whole `app.Action` and reads `a.Value` itself."),
		doc.Example(demos.OnExample),
		doc.P("Any component can fire the action, with the value the predicate reads:"),
		doc.Go(`func (d *onDemo) toggle(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(OutlineAction, ctx.JSSrc().Get("checked").Bool())
}`),
		doc.H("Two-way toggle"),
		doc.P("This site's dark-mode switch uses `On` with two rules on one selector. In a `theme` package:"),
		doc.Go(`package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const DarkAction = "theme/dark"

var Root = style.Element("html").
	Set(css.Custom("--bg", css.Hex("#ffffff"))).
	Media(media.PrefersColorScheme(kw.Dark), css.Custom("--bg", css.Hex("#111111")))

var ForceDark = style.Element("html").
	Set(css.Custom("--bg", css.Hex("#111111"))).
	On(DarkAction, darkRequested)

var ForceLight = style.Element("html").
	Set(css.Custom("--bg", css.Hex("#ffffff"))).
	On(DarkAction, lightRequested)

func darkRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && v
}

func lightRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && !v
}

func Styles() []style.Style {
	return []style.Style{
		style.Generated("/web/theme.css", Root),
		style.Dynamic(ForceDark, ForceLight),
	}
}`),
		doc.P("The Generated file holds `Root`:"),
		doc.CSS(`html {
	--bg: #ffffff;
}

@media (prefers-color-scheme: dark) {
	html {
		--bg: #111111;
	}
}`),
		doc.P("The engine inserts `ForceDark` or `ForceLight`:"),
		doc.CSS(`html {
	--bg: #111111;
}

html {
	--bg: #ffffff;
}`),
		doc.P("A component fires the action:"),
		doc.Go(`type toggle struct {
	app.Compo
	dark bool
}

func (t *toggle) Render() app.UI {
	return app.Button().Type("button").OnClick(t.flip).Text("Toggle")
}

func (t *toggle) flip(ctx app.Context, e app.Event) {
	t.dark = !t.dark
	ctx.NewActionWithValue(theme.DarkAction, t.dark)
}`),
		doc.P("Both force rules are absent at the start, so the `prefers-color-scheme` block of `Root` decides. An action carrying `true` inserts `ForceDark` and leaves `ForceLight` absent. An action carrying `false` removes `ForceDark` and inserts `ForceLight`."),
		doc.H("How bindings combine"),
		doc.List(
			doc.Item("An item may carry several bindings, on one action name or on several, and may mix `On`, `When` and `Compute`. For one action, the last binding with that name decides. Across actions, the latest fired action wins."),
			doc.Item("A `When` or `On` result equal to the item's current state changes nothing: the same value arriving twice never inserts an item twice."),
			doc.Item("A value of the wrong type for that last binding is logged with `app.Logf`, and the item is left as it was. Earlier bindings of the same action are not consulted. A nil `*T` is passed to the function."),
			doc.Item("A recompute that gives the same text keeps the item's position in the sheet. A different text moves it to the end."),
			doc.Item("`Bind` with a `ComputeFrames` binding on a `Rule`, or with a `Compute` binding on a `KeyframesRule`, is `ErrBadBinding`. So are an empty action name and a nil function."),
			doc.Item("`Dynamic` accepts only bound items (`ErrNotDynamic` otherwise), and `Generated` accepts only unbound ones (`ErrDynamicInGenerated`)."),
			doc.Item("Two items may share a selector, as `DemoCompactBase` and `DemoCompact` do. A Generated base with a Dynamic override on the same selector is the expected pattern, and two Dynamic rules on one selector is how a two-way toggle is built. The engine tells items apart by identity, never by selector text."),
		),
		doc.H("CSSFor"),
		doc.P("`CSSFor(action, v)` returns what the engine inserts after an action named `action` carrying `v`: the computed text, the declared text for a `When` or `On` that says present, or `\"\"` when the item is absent or `v` has the wrong type. Use it to test a binding without a browser."),
	}
}

func viewModel() []app.UI {
	return []app.UI{
		doc.Go(`func (v *ViewModel[T]) OnAction(ctx app.Context) {
	if v.source != nil && v.source.Mounted() {
		return
	}
	v.source = ctx.Src()
	ctx.ObserveState(v.key, v.state).OnChange(func() {
		ctx.NewActionWithValue(v.key, v.state)
	})
}`),
		doc.P("`OnAction` observes the ViewModel's key in the go-app state store. Each `Set` or `SetAndPersist` through the ViewModel fires an action named after the key, whose value is the ViewModel's `*T`. A typed binding on that name receives the pointer."),
		doc.List(
			doc.Item("Call it in `OnMount`. Call it from lifecycle hooks only, never from `Render`."),
			doc.Item("Use the ViewModel's key as the action name in `When` and `Compute`."),
			doc.Item("It registers once per ViewModel, however many components call it. It registers again only when the component it registered on has dismounted."),
			doc.Item("A direct write to the state, without `Set`, fires nothing."),
		),
		doc.P("The demos build each ViewModel with a small helper:"),
		doc.Go(`func model[T any](key string, value *T) *data.ViewModel[T] {
	m, err := data.NewViewModel(key, value)
	if err != nil {
		panic(err)
	}
	return m
}`),
		doc.P("The `when` demo's ViewModel and component:"),
		doc.Go(`var compact = model(CompactKey, &Compact{})

type whenDemo struct{ app.Compo }

func WhenDemo() app.UI { return &whenDemo{} }

func (d *whenDemo) OnMount(ctx app.Context) {
	compact.OnAction(ctx)
}

func (d *whenDemo) Render() app.UI {
	return app.Div().Body(
		app.Label().Body(
			app.Input().Type("checkbox").Checked(compact.State().On).OnChange(d.toggle),
			app.Text(" Compact"),
		),
		app.Div().Class(DemoCompact.Name()).Text("The padding follows the checkbox"),
	)
}

func (d *whenDemo) toggle(ctx app.Context, e app.Event) {
	on := ctx.JSSrc().Get("checked").Bool()
	compact.Load(ctx).Apply(func(ctx app.Context, c *Compact) *Compact {
		c.On = on
		return c
	}).Set()
}`),
		doc.P("A `Compute` rule without `Initially(true)` starts absent. To apply the current value at once, call `Set` after `OnAction` in `OnMount`:"),
		doc.Go(`func (d *computeDemo) OnMount(ctx app.Context) {
	accent.OnAction(ctx)
	accent.Load(ctx).Set()
}`),
		doc.P("The ", doc.A("/state", "Pages, layouts and state"), " page describes `ViewModel`, `Load`, `Apply` and `Set`."),
	}
}
