package applying

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func errs() []app.UI {
	return []app.UI{
		doc.P("Builders never panic on bad input. Each item keeps the first error a call produced, and `Err` returns it. `serve.Handler.Build`, `style.Resolve` and the item's own `Err` report it. Match errors with `errors.Is`, for example `errors.Is(err, style.ErrDuplicateRule)` or `errors.Is(err, serve.ErrMissingStyle)`."),
		doc.H("Item errors"),
		doc.List(
			doc.Item("`ErrNoName`: `Class(\"\")`, `Id(\"\")`, `Element(\"\")`, `Keyframes(\"\")`, `Property(\"\")`, `Layer(\"\")`, `LayerOrder()` or a `LayerOrder` holding `\"\"`, `Import(nil)`, `InLayer(\"\")`; a zero `Rule`; a combinator or `Nest` given a nil rule."),
			doc.Item("`ErrBadSelector`: `And` could not join its operands; `Self()` has a `Set` block; `Self()` is used outside `Nest`."),
			doc.Item("`ErrBadMediaQuery`, `ErrBadContainerQuery`, `ErrBadSupportsQuery`: `Media`, `Container`, `Supports`, `ForMedia` or `IfSupports` was given a query that carries an error. The block is left out, and a later valid call does not clear the error."),
			doc.Item("`ErrBadBinding`: `On` or `Bind` was given an empty action name, a nil function, or a binding that does not fit the item."),
			doc.Item("`ErrNested`: a rule nested in one parent was passed to another `Nest`, to a combinator, to `Generated` or to `Dynamic`; or a rule was nested inside itself."),
			doc.Item("`ErrNestedBinding`: a nested child carries a binding."),
			doc.Item("`ErrBadAtRule`: a `FontFace` without `font-family` or `src`; a `Property` without `syntax` or `inherits`; an `Import` or a `LayerOrder` inside a `Layer`; an `Import` target other than `css.Url` or `css.String`; a `Keyframes` stop outside 0% to 100%."),
		),
		doc.P("Once an item holds an error, later calls never replace it. A combinator's new rule takes the first error of its receiver, or else of its argument."),
		doc.H("Build errors"),
		doc.P("Each error `Build` returns wraps one sentinel and names the offending path or item, in the form `<sentinel message>: <path or item>`. The item is its selector, or `@keyframes name`, `@font-face`, `@property --name`, `@import url(…)` or `@layer name`. For example, `Static(\"/other/x.css\")` gives `serve: path is not under /web/: /other/x.css`."),
		doc.P("The path and file sentinels belong to `serve`:"),
		doc.List(
			doc.Item("`serve.ErrMissingStyle`: a Static file under `/web/` does not exist in `WebDir`. The message continues with the operating system's error, which `errors.Is` also matches, for example `os.ErrNotExist`."),
			doc.Item("`serve.ErrPathOutsideWeb`: a Static path that is not a URL, or a Generated path, does not start with `/web/`, or leaves it once cleaned."),
			doc.Item("`serve.ErrNotCSSPath`: a Generated path does not end in `.css`."),
			doc.Item("`serve.ErrDuplicatePath`: two Static or Generated styles in one handler use the same path. Paths are compared after cleaning, and URLs as given."),
		),
		doc.P("The sentinels about items belong to `style`, and `style.Resolve` returns them too:"),
		doc.List(
			doc.Item("`style.ErrDuplicateRule`: the same selector appears twice among siblings of one Generated style: at its top level, or among the direct children of one layer. A selector may repeat across layers and across Generated styles. The same rule applies to `@keyframes` names."),
			doc.Item("`style.ErrDynamicInGenerated`: a bound item was passed to `Generated`."),
			doc.Item("`style.ErrNotDynamic`: an unbound item, or an item other than a `*Rule` or a `*KeyframesRule`, was passed to `Dynamic`."),
			doc.Item("Every item error above."),
		),
		doc.P("A Generated base and a Dynamic override may share a selector, and so may two Dynamic rules."),
		doc.P("For one item, its placement error (`ErrNested`, or `ErrBadSelector` for a bare `Self()`) and its own error are reported before `ErrDynamicInGenerated`, `ErrNotDynamic` and `ErrDuplicateRule`. In `Dynamic`, an item of a type it does not accept is `ErrNotDynamic` first."),
		doc.P("A failure to write a generated file is not one of these sentinels. It is returned as an error that starts with `serve: ` and wraps the operating system's error."),
		doc.H("Query errors"),
		doc.P("`media`, `container` and `supports` each export one sentinel:"),
		doc.Go(`var ErrBadQuery error`),
		doc.P("A query that cannot be expressed carries it, and its `Err` returns it. An operation on a query that already carries an error keeps that error."),
		doc.List(
			doc.Item("`media`: `And` of two different media types, such as `media.Screen().And(media.Print())`; `And` where either operand is a comma-separated list made by `Or`; `And` where one operand is a negated media type, such as `media.Not(media.Screen())`; `Not` of a comma-separated list; `And` or `Not` of the zero `Query`."),
			doc.Item("`container`: `Named(\"\")`; `Named` anywhere but first, such as `container.MinWidth(css.Px(1)).And(container.Named(\"x\"))`; `Or` with a named operand; `And` where either operand is an `Or` list; `Not` of an `Or` list or of a query with no feature; the zero `Query`."),
			doc.Item("`supports`: `Property(nil)`, `Selector(\"\")`, `FontFormat(nil)`, `FontTech(nil)`; the zero `Query`."),
		),
		doc.P("`Rule.Media`, `Rule.Container` and `Rule.Supports` check `Err` and, for a query that carries one, latch `ErrBadMediaQuery`, `ErrBadContainerQuery` or `ErrBadSupportsQuery` on the rule and leave the block out."),
	}
}
