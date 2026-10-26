package styling

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/styling/demos"
)

func mediaQueries() []app.UI {
	return []app.UI{
		doc.P("`Query` is a struct with unexported fields. Build one from a media type or a media feature:"),
		doc.Go(`func Screen() Query
func Print() Query
func All() Query
func (q Query) And(o Query) Query
func (q Query) Or(o Query) Query
func Not(q Query) Query
func (q Query) CSS() string
func (q Query) Err() error`),
		doc.P("`Rule.Media` takes a `media.Query`. `CSS` writes the query text, as the writer puts it after `@media`."),
		doc.Example(demos.MediaExample),
		doc.H("Combining queries"),
		doc.List(
			doc.Item("`media.Screen()` writes `screen`."),
			doc.Item("`media.Screen().And(media.MaxWidth(css.Px(600)))` writes `screen and (max-width: 600px)`."),
			doc.Item("`media.MinWidth(css.Px(400)).And(media.MaxWidth(css.Px(800)))` writes `(min-width: 400px) and (max-width: 800px)`."),
			doc.Item("`media.Screen().Or(media.Print())` writes `screen, print`. `Or` joins queries into a comma-separated list, which CSS reads as any of them."),
			doc.Item("`media.Not(media.Screen())` writes `not screen`."),
			doc.Item("`media.Not(media.PrefersColorScheme(kw.Dark))` writes `not (prefers-color-scheme: dark)`."),
			doc.Item("`media.Screen().And(media.Not(media.PrefersColorScheme(kw.Dark)))` writes `screen and (not (prefers-color-scheme: dark))`."),
		),
		doc.P("`Not` negates the whole operand. A media type may not appear inside parentheses and may not be joined with `or`: `Or` writes a comma list and `Not` is written in front of the query."),
		doc.H("Features"),
		doc.P("Media features are generated, one set of functions per feature, from the same data as the properties."),
		doc.P("A range feature takes its value kind directly, not a marker interface, and comes in four forms. For a feature whose Go name is `X` they are `X`, `MinX`, `MaxX` and `XBetween`. At the current pins the range features are `width`, `height`, `device-width` and `device-height`, which take a `css.Length`; `aspect-ratio` and `device-aspect-ratio`, which take a `css.RatioValue`; `resolution`, which takes a `css.Resolution`; and `color` and `monochrome`, which take a `css.Integer`. For `width`:"),
		doc.Go(`func Width(v css.Length) Query
func MinWidth(v css.Length) Query
func MaxWidth(v css.Length) Query
func WidthBetween(lo, hi css.Length) Query`),
		doc.List(
			doc.Item("`media.Width(css.Px(600))` writes `(width: 600px)`."),
			doc.Item("`media.MinWidth(css.Px(600))` writes `(min-width: 600px)`, and `MaxWidth` writes `(max-width: 600px)`."),
			doc.Item("`media.WidthBetween(css.Px(400), css.Px(600))` writes `(400px <= width <= 600px)`, the range syntax. The generator emits it only when the latest release of every considered browser can render it without a vendor prefix; otherwise it emits the `min-` and `max-` pair instead: `(min-width: 400px) and (max-width: 600px)`. The choice is made once, when the package is generated, for every `Between` function. It is the range syntax at the current pins."),
		),
		doc.P("A discrete feature takes the keywords it allows, typed in the same way as a property:"),
		doc.Go(`media.PrefersColorScheme(kw.Dark)
media.Orientation(kw.Portrait)
media.Pointer(kw.Coarse)`),
		doc.P("These write `(prefers-color-scheme: dark)`, `(orientation: portrait)` and `(pointer: coarse)`. A keyword the feature does not allow is a compile error. A discrete feature also accepts `css.Raw`, but not `css.Var`: a browser does not substitute `var()` inside a media query."),
		doc.P("A combination that cannot be written as one media query carries `media.ErrBadQuery`, a `var ErrBadQuery error`. `container` and `supports` export the same sentinel. The Applying page lists the cases under ", doc.A("/applying#errors", "Errors"), "."),
	}
}

func containerQueries() []app.UI {
	return []app.UI{
		doc.P("A container query matches the size of the nearest ancestor that is a container. Make an element a container with `css.ContainerType(kw.InlineSize)`, and name it with `css.ContainerName` when a query should target that one by name."),
		doc.Go(`func Named(name string) Query
func (q Query) And(o Query) Query
func (q Query) Or(o Query) Query
func Not(q Query) Query
func (q Query) CSS() string
func (q Query) Err() error`),
		doc.P("`Rule.Container` takes a `container.Query`. The features are generated from the `@container` descriptors:"),
		doc.List(
			doc.Item("Range features, in the four forms `X`, `MinX`, `MaxX` and `XBetween`: `Width`, `Height`, `InlineSize` and `BlockSize`, which take a `css.Length`, and `AspectRatio`, which takes a `css.RatioValue`."),
			doc.Item("One discrete feature: `Orientation`, which takes `kw.Portrait` or `kw.Landscape`."),
		),
		doc.Example(demos.ContainerExample),
		doc.P("Both frames hold the same row. The wide one is over `16rem`, so its row stays a row. The narrow one is under it, so its row becomes a column."),
		doc.List(
			doc.Item("`container.Named(\"card\").And(container.MinWidth(css.Px(400)))` writes `card (min-width: 400px)`. `Named` may appear once, first."),
			doc.Item("`container.MinWidth(css.Rem(30)).Or(container.Orientation(kw.Portrait))` writes `(min-width: 30rem) or (orientation: portrait)`."),
			doc.Item("`container.Not(container.MaxWidth(css.Rem(20)))` writes `not (max-width: 20rem)`."),
			doc.Item("`container.WidthBetween(css.Rem(10), css.Rem(20))` writes `(10rem <= width <= 20rem)`. Container queries always use the range syntax."),
			doc.Item("A `Container` block on a rule that is not nested is written flat, as `@container … { .demo-row { … } }`. On a nested rule or a `Self()` child it is written inside the rule's braces, with bare declarations."),
		),
	}
}

func supportsQueries() []app.UI {
	return []app.UI{
		doc.Go(`func Property(p css.Property) Query
func Selector(text string) Query
func FontFormat(f css.FormatValue) Query
func FontTech(t css.Value) Query
func (q Query) And(o Query) Query
func (q Query) Or(o Query) Query
func Not(q Query) Query
func (q Query) CSS() string
func (q Query) Err() error`),
		doc.P("`Rule.Supports` takes a `supports.Query`. Write the fallback in `Set` and the feature in `Supports`:"),
		doc.Example(demos.SupportsExample),
		doc.List(
			doc.Item("`supports.Property(css.Display(kw.Grid))` writes `(display: grid)`. A prefixed property writes only its unprefixed declaration: `supports.Property(css.BackdropFilter(css.Raw(\"blur(4px)\")))` writes `(backdrop-filter: blur(4px))`."),
			doc.Item("`supports.Selector(\":has(a)\")` writes `selector(:has(a))`. The text is raw selector text."),
			doc.Item("`supports.FontFormat(kw.Woff2)` writes `font-format(woff2)`. `FontTech` writes `font-tech(…)`."),
			doc.Item("`And` and `Or` join two conditions, and a joined operand is wrapped in parentheses: `supports.Property(css.Display(kw.Grid)).Or(supports.Property(css.Display(kw.Flex))).And(supports.Selector(\":has(a)\"))` writes `((display: grid) or (display: flex)) and selector(:has(a))`."),
			doc.Item("`supports.Not(supports.Property(css.Display(kw.Grid)))` writes `not (display: grid)`."),
		),
	}
}
