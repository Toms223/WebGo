package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/container"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Frame = style.Class("demo-frame").
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
	Set(css.Padding(css.Rem(0.25), css.Rem(0.5)), css.Background(css.Var("--mint")))

func ContainerDemo() app.UI {
	return app.Div().Body(
		app.Div().Style(Frame).Body(cells()),
		app.Div().Style(Frame, Narrow).Body(cells()),
	)
}

func cells() app.UI {
	return app.Div().Style(Row).Body(
		app.Span().Text("One"),
		app.Span().Text("Two"),
		app.Span().Text("Three"),
	)
}
