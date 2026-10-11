This page covers delivering rules to the browser: the three kinds of stylesheet, the `serve` handler that writes them, the engine that inserts dynamic rules, the actions that drive it, and the errors. The Styling page covers writing rules.

Every example with a Result panel is live. Its bound rule is in this site's `Dynamic` style, and the controls change it through the same actions your app would fire. The CSS under an example is the text of the rules shown, for the demo's starting state, pinned by a test.

## Stylesheets

```go
type Style interface{ style() }

type Item interface {
	item()
	renderable
}

func Static(path string) Style
func Generated(path string, items ...Item) Style
func Dynamic(items ...Item) Style
```

`Style` has an unexported method: these three constructors are the only way to make one. The `serve` package turns a list of them into go-app's handler, under Handler and Build. `Item` has unexported methods too. `*Rule`, `*KeyframesRule`, `*FontFaceRule`, `*PropertyRule`, `*ImportRule`, `*LayerRule` and `*LayerOrderRule` implement it, and each has `Err() error`. A nil `Style` in a handler's list, and a nil item passed to `Generated` or `Dynamic`, are ignored.

The three kinds are chosen per stylesheet, and one handler can mix them:

- `Static(path)` links a `.css` file that already exists.
  - A path starting with `/web/` is looked up in `WebDir`. If the file is missing, `Build` returns `serve.ErrMissingStyle`.
  - A URL starting with `http://`, `https://` or `//` is linked as given and is not checked.
  - Any other path is `serve.ErrPathOutsideWeb`.
  - The content of the file is never read or validated.
- `Generated(path, items...)` writes a `.css` file from the items. The path must start with `/web/` (otherwise `serve.ErrPathOutsideWeb`) and end in `.css` (otherwise `serve.ErrNotCSSPath`). `Build` renders the items and writes the text to `<WebDir>/…`, creating directories as needed, only if the bytes differ from what is already on disk. The file is then linked like a Static one. A Generated style with no items, or whose items have no properties, still writes an empty file.
- `Dynamic(items...)` adds no link and writes no file. It takes `*Rule` and `*KeyframesRule` values that carry a binding. `Build` only validates them. They are inserted into and removed from the page in the browser, as described under Run and Actions.

A Generated file writes every `@layer` order statement first, then every `@import`, then everything else, each group in slice order, with one blank line between items.

Paths are cleaned before use, so `/web/./app.css` and `/web/x/../app.css` both mean `/web/app.css`, and the link written into the page is the cleaned path. A path that leaves `/web/` once cleaned, such as `/web/../secret.css`, is `serve.ErrPathOutsideWeb`. URLs are not cleaned.

Use `Static` for CSS that has no typed form, such as a property the generator leaves out, and for a stylesheet hosted elsewhere.

### Resolve

```go
type Kind int

const (
	StaticStyle Kind = iota + 1
	GeneratedStyle
	DynamicStyle
)

type Source struct {
	Kind    Kind
	Path    string
	Content []byte
}

func Resolve(s Style) (Source, error)
```

`Resolve` is what the handler calls on each style. It checks the items and returns the style's kind, its path as given, and for a Generated style the rendered bytes, which are empty but not nil when the style has no items. A Static style is not checked beyond that, and a Dynamic one has no path. The errors it returns are the item errors under Errors, and on an error the `Source` is zero. A nil `Style` resolves to a zero `Source` with no error. You need `Resolve` only to deliver styles through something other than `serve.Handler`.

### A complete app

Two rules, delivered in a Generated file. The rules live in a package of their own, `theme`, which also holds the one function that lists the app's stylesheets:

```go
package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var Card = style.Class("card").
	Set(css.Padding(css.Rem(1)), css.BorderRadius(css.Px(8)))

var Link = style.Class("nav-link").
	Set(css.Color(css.Var("--ink"))).
	Hover(css.TextDecoration(kw.Underline)).
	Media(media.Screen().And(media.MaxWidth(css.Px(600))), css.FontSize(css.Rem(0.9)))

func Styles() []style.Style {
	return []style.Style{
		style.Generated("/web/theme.css", Card, Link),
	}
}
```

The program that uses it, in a module named `example.com/hello`:

```go
package main

import (
	"log"
	"net/http"

	"example.com/hello/theme"
	"github.com/Toms223/WebGo/serve"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type hello struct {
	app.Compo
}

func (h *hello) Render() app.UI {
	return app.Div().Class(theme.Card.Name()).Body(
		app.A().Class(theme.Link.Name()).Href("/").Text("Home"),
	)
}

func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	style.Run(theme.Styles()...)
	app.RunWhenOnBrowser()

	handler := &serve.Handler{
		Handler: app.Handler{Name: "Hello"},
		Styles:  theme.Styles(),
	}
	built, err := handler.Build()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", built)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

`Build` writes `web/theme.css` and puts `/web/theme.css` in the embedded `app.Handler`'s `Styles`. The file contains:

```css
.card {
	padding: 1rem;
	border-radius: 8px;
}

.nav-link {
	color: var(--ink);
}

.nav-link:hover {
	text-decoration: underline;
}

@media screen and (max-width: 600px) {
	.nav-link {
		font-size: 0.9rem;
	}
}
```

## Handler and Build

```go
package serve

type Handler struct {
	app.Handler
	Styles []style.Style
	WebDir string
}
```

The `serve` package, at `github.com/Toms223/WebGo/serve`, is the server side of styling: it holds the handler, `Build` and the path and file errors. `style` holds the rules and `Run`, and does not import `serve`.

`serve.Handler` is go-app's `app.Handler` with one field changed: `Styles` is `[]style.Style` instead of `[]string`. The outer `Styles` shadows the embedded one, which is still reachable as `h.Handler.Styles`. `Handler` also has unexported fields, so a composite literal must name its fields.

```go
handler := &serve.Handler{
	Handler: app.Handler{Name: "Docs", Title: "Docs"},
	Styles: []style.Style{
		style.Static("/web/app.css"),
		style.Generated("/web/theme.css", theme.Card, theme.Link),
		style.Dynamic(theme.Compact),
	},
	WebDir: "web",
}
```

`theme.Compact` stands for any bound rule, such as the ones under Actions.

### Build

```go
func (h *Handler) Build() (*app.Handler, error)
```

`Build` resolves every style in the order of `Styles`:

- It returns the first problem it finds as a `nil` handler and an error. It first checks every style, in the order of `Styles`: its items, through `style.Resolve`, and then its path. Only then does it look up the Static files, so a missing Static file is reported only when no style has an item or path error, wherever each sits in the list. Generated files are written last, so a failed check leaves the web directory as it was.
- After any failure, `h.Handler.Styles` is left as it was.
- On success it sets the embedded `Handler.Styles` to the paths of the Static and Generated styles, in order, followed by whatever `h.Handler.Styles` held when `Build` was first called. It returns `&h.Handler`. Dynamic styles contribute no path.
- Calling `Build` again gives the same paths. Nothing is added twice.
- `Build` never copies the handler. `app.Handler` carries a `sync.Once` and must not be copied, so the pointer it returns is the handler's own embedded value.
- go-app still passes every path through the handler's `Resources` resolver. With `Resources: app.GitHubPages("WebGo")`, the link to `/web/theme.css` is written as `/WebGo/web/theme.css`. Everything downstream, `http.Handle` and `app.GenerateStaticWebsite`, is unchanged.

`Build` has a pointer receiver. On a variable, `handler.Build()` works. A composite literal is not addressable, so `serve.Handler{…}.Build()` does not compile: write `(&serve.Handler{…}).Build()`, or keep the handler in a variable first:

```go
built, err := (&serve.Handler{Styles: theme.Styles()}).Build()
```

`Build` is where generation happens, wherever the handler is built: in a build step, such as a command that calls `app.GenerateStaticWebsite`, or at the start of a server that runs go-app server-side. `app.GenerateStaticWebsite` does not copy `WebDir` into its output, so a static site publishes a generated stylesheet the same way it publishes `app.wasm`: by copying `WebDir` to `/web/`.

### WebDir

`WebDir` is the local directory go-app serves at `/web/`. `Build` uses it to find Static files and to write Generated ones: `/web/x/y.css` maps to `<WebDir>/x/y.css`. It defaults to `"web"`. It does not change where go-app serves files from, so keep the two in agreement. A relative `WebDir` is resolved against the working directory of the process that calls `Build`.

### Determinism

The text a Generated style writes depends only on its items: the same items give the same bytes, with no timestamp and no banner. Items are written in slice order, blocks in rule order and properties in call order. `Build` leaves a file alone when its bytes already match, so building twice does not touch the file's modification time. The writer's format is fixed:

- Tabs for indentation, one blank line between blocks and a newline at the end.
- A numeric value is written in its shortest round-trip form: `0.9rem`, not `.9rem`. A zero keeps its unit as given: `css.Px(0)` is `0px`.
- A prefixed declaration comes immediately before the unprefixed one: `-webkit-backdrop-filter: blur(4px);` and then `backdrop-filter: blur(4px);`.

What is or is not committed to version control is up to you.

## Run

```go
func Run(styles ...Style)
```

`Run` is the browser entry point. It returns at once when the program runs on the server (`app.IsServer`), so it is safe to call everywhere. Call it exactly once per app, in `main`, before `app.RunWhenOnBrowser()`:

```go
func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	style.Run(theme.Styles()...)
	app.RunWhenOnBrowser()
}
```

`Run` only looks at the `Dynamic` styles, so give it the same list as the handler. Keep one function, here `theme.Styles() []style.Style`, that both the handler and `Run` consume, so the server and the browser see the same items.

A second call collides with the first. `app.Handle` keeps one global handler per action name, so a second `Run` with an item bound to an action name the first one used replaces the first engine's handler for that name, and the items the first engine bound to it stop reacting. Give `Run` every `Dynamic` style the app has, in one call.

On the client, `Run` does this:

- It registers one global go-app action handler per distinct action name bound in the `Dynamic` styles, through `app.Handle`. The engine fans out to every item bound to that name, in the order the items were passed to `Run`.
- go-app runs `app.Handle` handlers off the UI goroutine, so the handler only calls `ctx.Dispatch`. Evaluating a binding and changing the page happen there, one action at a time.
- The engine owns exactly one `<style data-webgo="dynamic">` element. It is created in `document.head` on the first insertion attempt, and the engine uses its `.sheet`.
- Inserting an item inserts each block of its text at the end of that sheet, with `insertRule`, as one CSSOM rule per block: a flat rule with a base block, a `Hover` block and a `Media` block is three. A rule with nested children is one block, and so is a `KeyframesRule`. Removing an item deletes its CSSOM rules with `deleteRule`, and every other item's position stays correct.
- The browser can refuse a block: `insertRule` throws a `SyntaxError` when the block does not parse as one rule, for example for a selector or pseudo-class the browser does not accept, or a `css.Raw` value that closes the block early. A merely invalid value is dropped by the browser without an exception. The engine logs the exception through `app.Logf` and leaves that block out. The item's other blocks stay. An item with no inserted block is absent and is tried again the next time an action asks for it. The app does not crash.
- Items with `Initially(true)` are inserted during `Run`.

Dynamic items need the page's Content-Security-Policy to allow inline styles, with a `style-src` that permits the injected `<style>`. When the browser gives the element no sheet, every insertion attempt is logged through `app.Logf` and the app keeps running with the item absent. `Static` and `Generated` stylesheets are linked files and are unaffected.

### Action names belong to WebGo

`app.Handle` is a plain map assignment in go-app v11.0.5. A later `app.Handle` call on the same name silently replaces an earlier one, in either direction. Action names bound through `On` and `Bind` belong to WebGo: do not register your own handler for them with `app.Handle`. Whichever call runs last wins and the other handler is lost.

Register your own handlers for those names with `ctx.Handle` from a component. Component-level handlers are keyed separately from the global map, so they run alongside the engine's. With `theme.DarkAction` and the `toggle` component from Two-way toggle under Actions:

```go
func (t *toggle) OnMount(ctx app.Context) {
	ctx.Handle(theme.DarkAction, t.onDark)
}

func (t *toggle) onDark(ctx app.Context, a app.Action) {
	if dark, ok := a.Value.(bool); ok {
		t.dark = dark
		ctx.Update()
	}
}
```

The global map is read once, when go-app's engine loads, and `app.RunWhenOnBrowser` does not return in the browser. Call `Run` before it.

## Actions

```go
func When[T any](action string, when func(*T) bool) Binding
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
func FrameTo(props ...css.Property) Frame
```

A bound item reacts to go-app actions. `Bind` takes a typed `Binding`: the action's value must be a `*T`, and the function receives it. `On` takes the v1 predicate over the raw `app.Action`. Each time an action with a bound name fires, the engine evaluates the item and inserts it, removes it, or replaces its text. An action whose name no item is bound to is ignored.

Each demo below keeps its state in a `data.ViewModel` whose key is the action name. `OnAction` turns every change of that state into an action carrying the `*T`. See ViewModel and OnAction.

### When

`When` decides presence. Its function's result says whether the rule is present. The rule's declared blocks are what is inserted.

```go when
const CompactKey = "demo/compact"

type Compact struct{ On bool }

var DemoCompactBase = style.Class("demo-compact").
	Set(
		css.Padding(css.Rem(2)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--mint")),
	)

var DemoCompact = style.Class("demo-compact").
	Set(css.Padding(css.Rem(0.25))).
	Bind(style.When(CompactKey, func(c *Compact) bool { return c.On }))
```

```css
.demo-compact {
	padding: 2rem;
	border: 3px solid var(--ink);
	background: var(--mint);
}

.demo-compact {
	padding: 0.25rem;
}
```

`DemoCompactBase` is in a Generated file, inside `@layer components`. `DemoCompact` is in `Dynamic` and shares the selector. It wins while present because the dynamic sheet is unlayered, and unlayered rules beat layered ones. Without the layer it would still win: the dynamic sheet comes after the generated one.

### Compute

`Compute` decides the properties. Its function returns the rule's base block. The rule's pseudo blocks, conditional blocks and nested children stay as declared. A `nil` or empty result makes the rule absent.

```go compute
const AccentKey = "demo/accent"

type Accent struct{ Color string }

var DemoAccent = style.Class("demo-accent").
	Bind(style.Compute(AccentKey, func(a *Accent) []css.Property {
		return []css.Property{
			css.Color(css.Hex(a.Color)),
			css.BorderBottom(css.Px(3), kw.Solid, css.Hex(a.Color)),
		}
	}))
```

```css
.demo-accent {
	color: #FF3B30;
	border-bottom: 3px solid #FF3B30;
}
```

The CSS is `DemoAccent.CSSFor(AccentKey, &Accent{Color: "#FF3B30"})`, the demo's starting colour.

### ComputeFrames

`ComputeFrames` is `Compute` for a `KeyframesRule`: its function returns the frames. `FrameAt`, `FrameFrom` and `FrameTo` build one frame each. A `FrameAt` stop outside 0% to 100% gives an empty frame, which is not written.

```go frames
const SpeedKey = "demo/speed"

type Speed struct{ Beats int }

var DemoPulse = style.Keyframes("demo-beat").
	Bind(style.ComputeFrames(SpeedKey, func(s *Speed) []style.Frame {
		var frames []style.Frame
		steps := 2 * s.Beats
		for step := 0; step <= steps; step++ {
			opacity := 1.0
			if step%2 == 1 {
				opacity = 0.2
			}
			at := css.Pct(float64(step) * 100 / float64(steps))
			frames = append(frames, style.FrameAt(at, css.Opacity(css.Num(opacity))))
		}
		return frames
	}))

var DemoPulseBox = style.Class("demo-beating").
	Set(
		css.Width(css.Rem(3)),
		css.Height(css.Rem(3)),
		css.Background(css.Var("--accent")),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.AnimationName(css.Ident(DemoPulse.Name())),
		css.AnimationDuration(css.S(2)),
		css.AnimationIterationCount(kw.Infinite),
		css.AnimationTimingFunction(kw.Linear),
	)
```

```css
@keyframes demo-beat {
	0% {
		opacity: 1;
	}
	50% {
		opacity: 0.2;
	}
	100% {
		opacity: 1;
	}
}

.demo-beating {
	width: 3rem;
	height: 3rem;
	background: var(--accent);
	border: 3px solid var(--ink);
	animation-name: demo-beat;
	animation-duration: 2s;
	animation-iteration-count: infinite;
	animation-timing-function: linear;
}
```

The slider sets `Beats`. `DemoPulseBox` is in a Generated file and names the keyframes. `DemoPulse` is in `Dynamic` and is replaced on every change.

### Initially

`Initially(true)` inserts the item when `Run` starts, before any action fires. The default is absent. There is no `Default` method: `Default` is the `:default` pseudo-class.

```go initially
const BannerKey = "demo/banner"

type Banner struct{ Hidden bool }

var DemoBannerHidden = style.Class("demo-banner").
	Set(css.Display(kw.None))

var DemoBanner = style.Class("demo-banner").
	Set(
		css.Display(kw.Block),
		css.Padding(css.Rem(0.75), css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--side")),
	).
	Bind(style.When(BannerKey, func(b *Banner) bool { return !b.Hidden })).
	Initially(true)
```

```css
.demo-banner {
	display: none;
}

.demo-banner {
	display: block;
	padding: 0.75rem 1rem;
	border: 3px solid var(--ink);
	background: var(--side);
}
```

`Initially(true)` on a `Compute` rule inserts its declared `Set` block before any action. With no `Set` block it inserts nothing until the first action.

### On

`On` is the untyped form. Its function receives the whole `app.Action` and reads `a.Value` itself.

```go on
const OutlineAction = "demo/outline"

var DemoOutlineBase = style.Class("demo-outlined").
	Set(
		css.Padding(css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	)

var DemoOutline = style.Class("demo-outlined").
	Set(
		css.Outline(css.Px(3), kw.Dashed, css.Var("--accent")),
		css.OutlineOffset(css.Px(4)),
	).
	On(OutlineAction, outlined)

func outlined(a app.Action) bool {
	on, ok := a.Value.(bool)
	return ok && on
}
```

```css
.demo-outlined {
	padding: 1rem;
	border: 3px solid var(--ink);
}

.demo-outlined {
	outline: 3px dashed var(--accent);
	outline-offset: 4px;
}
```

Any component can fire the action, with the value the predicate reads:

```go
func (d *onDemo) toggle(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(OutlineAction, ctx.JSSrc().Get("checked").Bool())
}
```

### Two-way toggle

This site's dark-mode switch uses `On` with two rules on one selector. In a `theme` package:

```go
package theme

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
}
```

The Generated file holds `Root`:

```css
html {
	--bg: #ffffff;
}

@media (prefers-color-scheme: dark) {
	html {
		--bg: #111111;
	}
}
```

The engine inserts `ForceDark` or `ForceLight`:

```css
html {
	--bg: #111111;
}

html {
	--bg: #ffffff;
}
```

A component fires the action:

```go
type toggle struct {
	app.Compo
	dark bool
}

func (t *toggle) Render() app.UI {
	return app.Button().Type("button").OnClick(t.flip).Text("Toggle")
}

func (t *toggle) flip(ctx app.Context, e app.Event) {
	t.dark = !t.dark
	ctx.NewActionWithValue(theme.DarkAction, t.dark)
}
```

Both force rules are absent at the start, so the `prefers-color-scheme` block of `Root` decides. An action carrying `true` inserts `ForceDark` and leaves `ForceLight` absent. An action carrying `false` removes `ForceDark` and inserts `ForceLight`.

### How bindings combine

- An item may carry several bindings, on one action name or on several, and may mix `On`, `When` and `Compute`. For one action, the last binding with that name decides. Across actions, the latest fired action wins.
- A `When` or `On` result equal to the item's current state changes nothing: the same value arriving twice never inserts an item twice.
- A value of the wrong type for that last binding is logged with `app.Logf`, and the item is left as it was. Earlier bindings of the same action are not consulted. A nil `*T` is passed to the function.
- A recompute that gives the same text keeps the item's position in the sheet. A different text moves it to the end.
- `Bind` with a `ComputeFrames` binding on a `Rule`, or with a `Compute` binding on a `KeyframesRule`, is `ErrBadBinding`. So are an empty action name and a nil function.
- `Dynamic` accepts only bound items (`ErrNotDynamic` otherwise), and `Generated` accepts only unbound ones (`ErrDynamicInGenerated`).
- Two items may share a selector, as `DemoCompactBase` and `DemoCompact` do. A Generated base with a Dynamic override on the same selector is the expected pattern, and two Dynamic rules on one selector is how a two-way toggle is built. The engine tells items apart by identity, never by selector text.

### CSSFor

`CSSFor(action, v)` returns what the engine inserts after an action named `action` carrying `v`: the computed text, the declared text for a `When` or `On` that says present, or `""` when the item is absent or `v` has the wrong type. Use it to test a binding without a browser.

## ViewModel and OnAction

```go
func (v *ViewModel[T]) OnAction(ctx app.Context) {
	if v.source != nil && v.source.Mounted() {
		return
	}
	v.source = ctx.Src()
	ctx.ObserveState(v.key, v.state).OnChange(func() {
		ctx.NewActionWithValue(v.key, v.state)
	})
}
```

`OnAction` observes the ViewModel's key in the go-app state store. Each `Set` or `SetAndPersist` through the ViewModel fires an action named after the key, whose value is the ViewModel's `*T`. A typed binding on that name receives the pointer.

- Call it in `OnMount`. Call it from lifecycle hooks only, never from `Render`.
- Use the ViewModel's key as the action name in `When` and `Compute`.
- It registers once per ViewModel, however many components call it. It registers again only when the component it registered on has dismounted.
- A direct write to the state, without `Set`, fires nothing.

The demos build each ViewModel with a small helper:

```go
func model[T any](key string, value *T) *data.ViewModel[T] {
	m, err := data.NewViewModel(key, value)
	if err != nil {
		panic(err)
	}
	return m
}
```

The `when` demo's ViewModel and component:

```go
var compact = model(CompactKey, &Compact{})

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
}
```

A `Compute` rule without `Initially(true)` starts absent. To apply the current value at once, call `Set` after `OnAction` in `OnMount`:

```go
func (d *computeDemo) OnMount(ctx app.Context) {
	accent.OnAction(ctx)
	accent.Load(ctx).Set()
}
```

The Pages, layouts and state page describes `ViewModel`, `Load`, `Apply` and `Set`.

## Errors

Builders never panic on bad input. Each item keeps the first error a call produced, and `Err` returns it. `serve.Handler.Build`, `style.Resolve` and the item's own `Err` report it. Match errors with `errors.Is`, for example `errors.Is(err, style.ErrDuplicateRule)` or `errors.Is(err, serve.ErrMissingStyle)`.

### Item errors

- `ErrNoName`: `Class("")`, `Id("")`, `Element("")`, `Keyframes("")`, `Property("")`, `Layer("")`, `LayerOrder()` or a `LayerOrder` holding `""`, `Import(nil)`, `InLayer("")`; a zero `Rule`; a combinator or `Nest` given a nil rule.
- `ErrBadSelector`: `And` could not join its operands; `Self()` has a `Set` block; `Self()` is used outside `Nest`.
- `ErrBadMediaQuery`, `ErrBadContainerQuery`, `ErrBadSupportsQuery`: `Media`, `Container`, `Supports`, `ForMedia` or `IfSupports` was given a query that carries an error. The block is left out, and a later valid call does not clear the error.
- `ErrBadBinding`: `On` or `Bind` was given an empty action name, a nil function, or a binding that does not fit the item.
- `ErrNested`: a rule nested in one parent was passed to another `Nest`, to a combinator, to `Generated` or to `Dynamic`; or a rule was nested inside itself.
- `ErrNestedBinding`: a nested child carries a binding.
- `ErrBadAtRule`: a `FontFace` without `font-family` or `src`; a `Property` without `syntax` or `inherits`; an `Import` or a `LayerOrder` inside a `Layer`; an `Import` target other than `css.Url` or `css.String`; a `Keyframes` stop outside 0% to 100%.

Once an item holds an error, later calls never replace it. A combinator's new rule takes the first error of its receiver, or else of its argument.

### Build errors

Each error `Build` returns wraps one sentinel and names the offending path or item, in the form `<sentinel message>: <path or item>`. The item is its selector, or `@keyframes name`, `@font-face`, `@property --name`, `@import url(…)` or `@layer name`. For example, `Static("/other/x.css")` gives `serve: path is not under /web/: /other/x.css`.

The path and file sentinels belong to `serve`:

- `serve.ErrMissingStyle`: a Static file under `/web/` does not exist in `WebDir`. The message continues with the operating system's error, which `errors.Is` also matches, for example `os.ErrNotExist`.
- `serve.ErrPathOutsideWeb`: a Static path that is not a URL, or a Generated path, does not start with `/web/`, or leaves it once cleaned.
- `serve.ErrNotCSSPath`: a Generated path does not end in `.css`.
- `serve.ErrDuplicatePath`: two Static or Generated styles in one handler use the same path. Paths are compared after cleaning, and URLs as given.

The sentinels about items belong to `style`, and `style.Resolve` returns them too:

- `style.ErrDuplicateRule`: the same selector appears twice among siblings of one Generated style: at its top level, or among the direct children of one layer. A selector may repeat across layers and across Generated styles. The same rule applies to `@keyframes` names.
- `style.ErrDynamicInGenerated`: a bound item was passed to `Generated`.
- `style.ErrNotDynamic`: an unbound item, or an item other than a `*Rule` or a `*KeyframesRule`, was passed to `Dynamic`.
- Every item error above.

A Generated base and a Dynamic override may share a selector, and so may two Dynamic rules.

For one item, its placement error (`ErrNested`, or `ErrBadSelector` for a bare `Self()`) and its own error are reported before `ErrDynamicInGenerated`, `ErrNotDynamic` and `ErrDuplicateRule`. In `Dynamic`, an item of a type it does not accept is `ErrNotDynamic` first.

A failure to write a generated file is not one of these sentinels. It is returned as an error that starts with `serve: ` and wraps the operating system's error.

### Query errors

`media`, `container` and `supports` each export one sentinel:

```go
var ErrBadQuery error
```

A query that cannot be expressed carries it, and its `Err` returns it. An operation on a query that already carries an error keeps that error.

- `media`: `And` of two different media types, such as `media.Screen().And(media.Print())`; `And` where either operand is a comma-separated list made by `Or`; `And` where one operand is a negated media type, such as `media.Not(media.Screen())`; `Not` of a comma-separated list; `And` or `Not` of the zero `Query`.
- `container`: `Named("")`; `Named` anywhere but first, such as `container.MinWidth(css.Px(1)).And(container.Named("x"))`; `Or` with a named operand; `And` where either operand is an `Or` list; `Not` of an `Or` list or of a query with no feature; the zero `Query`.
- `supports`: `Property(nil)`, `Selector("")`, `FontFormat(nil)`, `FontTech(nil)`; the zero `Query`.

`Rule.Media`, `Rule.Container` and `Rule.Supports` check `Err` and, for a query that carries one, latch `ErrBadMediaQuery`, `ErrBadContainerQuery` or `ErrBadSupportsQuery` on the rule and leave the block out.
