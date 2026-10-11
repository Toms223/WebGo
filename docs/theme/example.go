package theme

import (
	"github.com/Toms223/WebGo/container"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var ExampleLabel = style.Class("example-label").
	Set(label...).
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.Padding(css.Rem(0.45), css.Rem(1)),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
		css.Background(css.Var("lime")),
		css.Color(css.Var("on-fill")),
	)

var ExampleCode = style.Class("example-code").Set(
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
)

var ExampleLive = style.Class("example-live").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(1)),
		css.Padding(css.Rem(1.25)),
	).
	Set(paper...).
	Set(css.BackgroundSize(css.Px(16), css.Px(16))).
	Container(container.MaxWidth(css.Rem(20)),
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.Stretch),
	)

var exampleResult = ExampleCode.Next(ExampleLabel).Set(css.Background(css.Var("sun")))

var exampleSource = ExampleCode.Descendant(style.Element("pre")).
	Set(
		css.Margin(css.Px(0)),
		css.Padding(css.Rem(1.1), css.Rem(1.25)),
		css.Border(css.Px(0)),
		css.BorderRadius(css.Px(0)),
		css.BoxShadow(kw.None),
	).
	Before(css.Display(kw.None))

var Example = style.Class("example").
	Set(
		css.ContainerType(kw.InlineSize),
		css.Margin(css.Px(0), css.Px(0), css.Rem(2)),
		line,
		css.BorderRadius(css.Var("r-l")),
		css.BoxShadow(css.Var("shadow")),
		css.Background(css.Var("panel")),
		css.Overflow(kw.Hidden),
	).
	Nest(exampleResult, exampleSource)
