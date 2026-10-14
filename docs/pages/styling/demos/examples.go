package demos

import (
	"strings"

	"github.com/Toms223/WebGo/docs/components/example"
)

const cardSource = `var Card = style.Class("demo-card").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Hover(css.Background(css.Var("--pink")))`

const selectorsSource = `var List = style.Class("demo-list")

var ListItem = List.Child(style.Element("li")).
	Set(css.Padding(css.Rem(0.25), css.Rem(0.75))).
	NthChild(css.Odd, css.Background(css.Var("--mint")))`

const pseudoSource = `var Button = style.Class("demo-button").
	Set(
		css.Padding(css.Rem(0.5), css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--surface")),
	).
	Before(css.Content(css.String("→ "))).
	Hover(css.Background(css.Var("--side"))).
	Active(css.Transform(css.Raw("translate(2px, 2px)"))).
	FocusVisible(css.Outline(css.Px(3), kw.Dashed, css.Var("--accent")))`

const nestingSource = `var PanelTitle = style.Class("demo-panel-title").
	Set(css.FontWeight(css.Num(900)))

var Panel = style.Class("demo-panel").
	Set(css.Padding(css.Rem(1)), css.Border(css.Px(3), kw.Solid, css.Var("--ink"))).
	Nest(
		PanelTitle,
		style.Element("a").Hover(css.Background(css.Var("--side"))),
		style.Self().Hover(css.Background(css.Var("--mint"))),
		style.Self().Media(media.MaxWidth(css.Rem(48)), css.Padding(css.Rem(0.5))),
	)`

const hovermediaSource = `var Tile = style.Class("demo-tile").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Nest(
		style.Self().
			Hover(css.Background(css.Var("--pink"))).
			Media(media.MaxWidth(css.Rem(48)), css.Background(css.Var("--side"))),
	)`

const valuesSource = `var Swatch = style.Class("demo-swatch").
	Set(
		css.Custom("demo-size", css.Rem(4)),
		css.Width(css.Var("--demo-size")),
		css.Height(css.Calc(css.Div(css.Var("--demo-size"), css.Num(2)))),
		css.Background(css.HSL(200, 80, 60)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	)`

const mediaSource = `var Wide = style.Class("demo-wide").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Media(media.MaxWidth(css.Rem(48)), css.Background(css.Var("--pink")))`

const containerSource = `var Frame = style.Class("demo-frame").
	Set(
		css.ContainerType(kw.InlineSize),
		css.Width(css.Rem(22)),
		css.MaxWidth(css.Pct(100)),
		css.Padding(css.Rem(0.5)),
		css.Border(css.Px(3), kw.Dashed, css.Var("--ink")),
	)

var Narrow = style.Class("demo-frame-narrow")

var NarrowFrame = Frame.And(Narrow).
	Set(css.Width(css.Rem(12)))

var Row = style.Class("demo-row").
	Set(css.Display(kw.Flex), css.Gap(css.Rem(0.5))).
	Container(container.MaxWidth(css.Rem(16)), css.FlexDirection(kw.Column))

var Cell = Row.Child(style.Element("span")).
	Set(css.Padding(css.Rem(0.25), css.Rem(0.5)), css.Background(css.Var("--mint")))`

const supportsSource = `var Mixed = style.Class("demo-mixed").
	Set(css.Padding(css.Rem(1)), css.BackgroundColor(css.Var("--mint"))).
	Supports(
		supports.Property(css.BackgroundColor(css.Raw("color-mix(in srgb, red, blue)"))),
		css.BackgroundColor(css.Raw("color-mix(in srgb, var(--mint), var(--pink))")),
	)`

const keyframesSource = `var Pulse = style.Keyframes("demo-pulse").
	From(css.Transform(css.Raw("scale(1)"))).
	To(css.Transform(css.Raw("scale(1.2)")))

var Pulsing = style.Class("demo-pulsing").
	Set(
		css.Width(css.Rem(3)),
		css.Height(css.Rem(3)),
		css.Background(css.Var("--accent")),
		css.AnimationName(css.Ident(Pulse.Name())),
		css.AnimationDuration(css.Ms(800)),
		css.AnimationIterationCount(kw.Infinite),
		css.AnimationDirection(kw.Alternate),
	)`

const propertySource = `var Angle = style.Property("demo-angle",
	css.Syntax(css.SyntaxAngle),
	css.Inherits(kw.False),
	css.InitialValue(css.Deg(0)),
)

var Turning = style.Class("demo-turning").
	Set(
		css.Width(css.Rem(10)),
		css.Height(css.Rem(4)),
		css.Background(css.Raw("linear-gradient(var(--demo-angle), var(--mint), var(--pink))")),
		css.Transition(css.Raw("--demo-angle 600ms")),
	).
	Hover(css.Custom(Angle.Name(), css.Deg(180)))`

const layersSource = `var LayeredId = style.Id("demo-layered").
	Set(css.Background(css.Var("--pink")))

var Layered = style.Class("demo-layered").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint")))

var Under = style.Layer("demo-under", LayeredId)

var Over = style.Layer("demo-over", Layered)`

const importantSource = `var Quiet = style.Id("demo-quiet").
	Set(css.Background(css.Var("--mint")))

var Loud = style.Class("demo-loud").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--pink")).Important())`

var CardExample = example.Example{Name: "card.go", Source: cardSource, CSS: joined(Card.CSS()), Live: CardDemo}

var SelectorsExample = example.Example{Name: "selectors.go", Source: selectorsSource, CSS: joined(ListItem.CSS()), Live: SelectorsDemo}

var PseudoExample = example.Example{Name: "pseudo.go", Source: pseudoSource, CSS: joined(Button.CSS()), Live: PseudoDemo}

var NestingExample = example.Example{Name: "nesting.go", Source: nestingSource, CSS: joined(Panel.CSS()), Live: NestingDemo}

var HoverMediaExample = example.Example{Name: "hovermedia.go", Source: hovermediaSource, CSS: joined(Tile.CSS()), Live: HoverMediaDemo}

var ValuesExample = example.Example{Name: "values.go", Source: valuesSource, CSS: joined(Swatch.CSS()), Live: ValuesDemo}

var MediaExample = example.Example{Name: "media.go", Source: mediaSource, CSS: joined(Wide.CSS()), Live: MediaDemo}

var ContainerExample = example.Example{Name: "container.go", Source: containerSource, CSS: joined(Frame.CSS(), NarrowFrame.CSS(), Row.CSS(), Cell.CSS()), Live: ContainerDemo}

var SupportsExample = example.Example{Name: "supports.go", Source: supportsSource, CSS: joined(Mixed.CSS()), Live: SupportsDemo}

var KeyframesExample = example.Example{Name: "keyframes.go", Source: keyframesSource, CSS: joined(Pulse.CSS(), Pulsing.CSS()), Live: KeyframesDemo}

var PropertyExample = example.Example{Name: "property.go", Source: propertySource, CSS: joined(Angle.CSS(), Turning.CSS()), Live: PropertyDemo}

var LayersExample = example.Example{Name: "layers.go", Source: layersSource, CSS: joined(Under.CSS(), Over.CSS()), Live: LayersDemo}

var ImportantExample = example.Example{Name: "important.go", Source: importantSource, CSS: joined(Quiet.CSS(), Loud.CSS()), Live: ImportantDemo}

func Examples() []example.Example {
	return []example.Example{CardExample, SelectorsExample, PseudoExample, NestingExample, HoverMediaExample, ValuesExample, MediaExample, ContainerExample, SupportsExample, KeyframesExample, PropertyExample, LayersExample, ImportantExample}
}

func joined(texts ...string) string {
	return strings.Join(texts, "\n")
}
