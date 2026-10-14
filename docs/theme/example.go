package theme

import (
	"github.com/Toms223/WebGo/container"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Example = style.Class("example").Set(
	css.ContainerType(kw.InlineSize),
	css.AlignSelf(kw.Stretch),
	css.Margin(css.Rem(0.35), css.Px(0)),
	css.MinWidth(css.Px(0)),
	line,
	css.BorderRadius(css.Var("r-l")),
	css.BoxShadow(css.Var("shadow")),
	css.Background(css.Var("code")),
	css.Overflow(kw.Hidden),
)

var ExampleBar = style.Class("example-bar").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(0.5)),
	css.Padding(css.Rem(0.45), css.Rem(0.6), css.Rem(0.45), css.Rem(1)),
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	css.Background(css.Var("lilac")),
	css.Color(css.Var("on-fill")),
)

var ExampleTools = style.Class("example-tools").Set(
	css.Display(kw.Flex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.5)),
)

var ExampleName = style.Class("example-name").
	Set(label...).
	Set(css.LetterSpacing(css.Em(0.1)), css.TextTransform(kw.None))

var ExampleTabs = style.Class("example-tabs").Set(
	css.Display(kw.Flex),
	css.Gap(css.Rem(0.3)),
	css.Padding(css.Px(3)),
	line,
	css.BorderRadius(css.Px(999)),
	css.Background(css.Var("panel")),
)

var ExampleTab = style.Class("example-tab").
	Set(label...).
	Set(
		css.Position(kw.Relative),
		css.Padding(css.Rem(0.25), css.Rem(0.7)),
		css.Border(css.Px(0)),
		css.BorderRadius(css.Px(999)),
		css.Background(kw.Transparent),
		css.Color(css.Var("ink")),
		css.Cursor(kw.Pointer),
		css.Transition(css.Raw("background 120ms ease, color 120ms ease")),
	).
	Hover(css.Background(css.Var("panel-2"))).
	FocusVisible(css.Outline(css.Px(3), kw.Solid, css.Var("on-fill")), css.OutlineOffset(css.Px(2))).
	After(hitArea...)

var ExampleTabOn = style.Class("example-tab-on").
	Set(css.Background(css.Var("tab")), css.Color(css.Var("tab-on"))).
	Hover(css.Background(css.Var("tab")))

var hitArea = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Absolute),
	css.Inset(css.Px(-10), css.Px(0)),
}

var ExampleCodeClamped = style.Class("example-code-clamped").Set(
	css.MaxHeight(css.Rem(24)),
	css.OverflowY(kw.Hidden),
)

var ExampleMore = style.Class("example-more").
	Set(label...).
	Set(
		css.Width(css.Pct(100)),
		css.MinHeight(css.Px(44)),
		css.Padding(css.Rem(0.4), css.Rem(1)),
		css.Border(css.Px(0)),
		css.BorderTop(css.Var("border"), kw.Solid, css.Var("line")),
		css.Background(css.Var("code")),
		css.Color(css.Var("code-on")),
		css.TextAlign(kw.Left),
		css.Cursor(kw.Pointer),
	).
	Hover(css.Color(css.Var("code-fn"))).
	FocusVisible(css.Outline(css.Px(3), kw.Solid, css.Var("code-on")), css.OutlineOffset(css.Px(-5)))

var CopyButton = style.Class("copy").
	Set(label...).
	Set(
		css.Position(kw.Relative),
		css.Padding(css.Rem(0.2), css.Rem(0.65)),
		line,
		css.BorderRadius(css.Px(999)),
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.FontSize(css.Rem(0.7)),
		css.Cursor(kw.Pointer),
		css.Transition(css.Raw("transform 120ms ease, box-shadow 120ms ease")),
	).
	Hover(lift...).
	Active(press...).
	FocusVisible(css.Outline(css.Px(3), kw.Solid, css.Var("on-fill")), css.OutlineOffset(css.Px(2))).
	After(hitArea...)

var CopyFloat = style.Class("copy-float").Set(
	css.Position(kw.Absolute),
	css.Top(css.Rem(0.6)),
	css.Right(css.Rem(0.75)),
	css.ZIndex(css.Int(1)),
	css.Background(css.Var("lilac")),
	css.Color(css.Var("on-fill")),
)

var ExampleCode = style.Class("example-code").Set(
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(1.1), css.Rem(1.25)),
	css.OverflowX(kw.Auto),
	css.Color(css.Var("code-on")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.875)),
	css.FontWeight(css.Num(500)),
	css.LineHeight(css.Num(1.7)),
	css.TabSize(css.Num(4)),
)

var ExamplePre = Example.Descendant(ExampleCode).Set(
	css.Margin(css.Px(0)),
	css.Border(css.Px(0)),
	css.BorderRadius(css.Px(0)),
	css.BoxShadow(kw.None),
	css.AlignSelf(kw.Auto),
)

var ExampleLabel = style.Class("example-label").
	Set(label...).
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.Padding(css.Rem(0.45), css.Rem(1)),
		css.BorderTop(css.Var("border"), kw.Solid, css.Var("line")),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
		css.Background(css.Var("sun")),
		css.Color(css.Var("on-fill")),
	).
	Before(
		css.Content(css.String("")),
		css.Width(css.Px(9)),
		css.Height(css.Px(9)),
		css.Border(css.Px(1.5), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Pct(50)),
		css.Background(css.Var("coral")),
	)

var ExampleLive = style.Class("example-live").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(1)),
		css.Padding(css.Rem(1.25)),
		css.Color(css.Var("ink")),
	).
	Set(paper...).
	Set(css.BackgroundSize(css.Px(16), css.Px(16))).
	Container(container.MaxWidth(css.Rem(20)),
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.Stretch),
	)

var ExampleLiveDoc = ExampleLive.Descendant(Doc).Set(css.Flex(css.Raw("1 1 100%")))

var LiveButton = ExampleLive.Descendant(style.Element("button")).
	Set(pill...).
	Set(css.MinHeight(css.Px(36)), css.Padding(css.Rem(0.3), css.Rem(0.85)), css.Transition(css.Raw("transform 120ms ease, box-shadow 120ms ease"))).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...)

var LiveLabel = ExampleLive.Descendant(style.Element("label")).Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.5)),
	css.FontWeight(css.Num(600)),
)

var LiveInput = ExampleLive.Descendant(style.Element("input")).
	Set(css.AccentColor(css.Var("coral"))).
	FocusVisible(ring...)
