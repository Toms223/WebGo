package theme

import (
	"github.com/Toms223/WebGo/container"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var compactDiagram = container.MaxWidth(css.Rem(34))

var Diagram = style.Class("diagram").Set(
	css.ContainerType(kw.InlineSize),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.9)),
	css.Width(css.Pct(100)),
	css.Margin(css.Rem(0.4), css.Px(0)),
	css.Padding(css.Rem(1)),
	line,
	css.BoxShadow(css.Var("shadow")),
	css.Custom("dot", css.Var("lilac")),
	css.BackgroundColor(css.Var("ground")),
	css.BackgroundImage(css.Raw("linear-gradient(150deg, var(--ground) 55%, transparent), radial-gradient(circle, var(--dot) 34%, transparent 37%)")),
	css.BackgroundSize(css.Raw("auto, 12px 12px")),
)

var DiagramLane = style.Class("dg-lane").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.65)),
	css.Padding(css.Rem(0.85), css.Rem(1), css.Rem(1)),
	line,
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow-s")),
	css.MinWidth(css.Px(0)),
)

var DiagramTag = style.Class("dg-tag").
	Set(label...).
	Set(
		css.AlignSelf(kw.FlexStart),
		css.Padding(css.Rem(0.2), css.Rem(0.55)),
		css.Background(css.Var("on-fill")),
		css.Color(css.Var("sun")),
		css.FontSize(css.Rem(0.8125)),
		css.Rotate(css.Deg(-2)),
	)

var DiagramTitle = style.Class("dg-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.3)),
	css.FontWeight(css.Num(400)),
	css.LineHeight(css.Num(1.1)),
	css.TextTransform(kw.Uppercase),
	css.Color(css.Var("mis-ink")),
	css.TextShadow(css.Var("mis-shadow")),
)

var DiagramFlow = style.Class("dg-flow").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.FlexWrap(kw.Wrap),
		css.Gap(css.Rem(0.5)),
		css.MinWidth(css.Px(0)),
	)

var DiagramStep = style.Class("dg-step").Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.FlexWrap(kw.Wrap),
	css.Gap(css.Rem(0.5)),
	css.MinWidth(css.Px(0)),
)

var DiagramStack = style.Class("dg-stack").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.35)),
	css.MinWidth(css.Px(0)),
)

var DiagramStackStep = DiagramStack.Child(DiagramStep).Set(
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.35)),
)

var DiagramBox = style.Class("dg-box").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(0.15)),
		css.MinWidth(kw.MinContent),
		css.MaxWidth(css.Rem(17)),
		css.Padding(css.Rem(0.55), css.Rem(0.8)),
		line,
		css.Custom("fill", css.Var("panel")),
		css.Custom("dot", css.Var("ink-pink")),
		css.BackgroundColor(css.Var("fill")),
		css.BackgroundImage(css.Raw("linear-gradient(100deg, var(--fill) 62%, transparent), radial-gradient(circle, var(--dot) 34%, transparent 37%)")),
		css.BackgroundSize(css.Raw("auto, 7px 7px")),
		css.Color(css.Var("ink")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Container(compactDiagram, css.MaxWidth(kw.None))

var DiagramSun = DiagramBox.And(style.Class("dg-sun")).Set(css.Custom("fill", css.Var("sun")), css.Custom("dot", css.Var("ink-pink")), css.Color(css.Var("on-fill")))

var DiagramLime = DiagramBox.And(style.Class("dg-lime")).Set(css.Custom("fill", css.Var("lime")), css.Custom("dot", css.Var("ink-pink-deep")), css.Color(css.Var("on-fill")))

var DiagramLilac = DiagramBox.And(style.Class("dg-lilac")).Set(css.Custom("fill", css.Var("lilac")), css.Custom("dot", css.Var("ink-blue")), css.Color(css.Var("on-fill")))

var DiagramCoral = DiagramBox.And(style.Class("dg-coral")).Set(css.Custom("fill", css.Var("coral")), css.Custom("dot", css.Var("ink-blue")), css.Color(css.Var("on-fill")))

var DiagramName = style.Class("dg-name").Set(
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.05)),
	css.FontWeight(css.Num(400)),
	css.LetterSpacing(css.Em(0.02)),
	css.LineHeight(css.Num(1.15)),
	css.TextTransform(kw.Uppercase),
)

var DiagramSub = style.Class("dg-sub").Set(
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.8125)),
	css.LineHeight(css.Num(1.35)),
	css.OverflowWrap(kw.Anywhere),
)

var DiagramArrow = style.Class("dg-arrow").Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.35)),
	css.Flex(kw.None),
	css.Color(css.Var("ink")),
)

var DiagramGlyph = style.Class("dg-glyph").
	Set(
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(1.25)),
		css.FontWeight(css.Num(700)),
		css.LineHeight(css.Num(1)),
	)

var DiagramDown = DiagramGlyph.And(style.Class("dg-down")).
	Set(css.Transform(css.Raw("rotate(90deg)")))

var DiagramVia = style.Class("dg-via").Set(
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.8125)),
	css.FontWeight(css.Num(700)),
	css.Color(css.Var("ink-2")),
)

var DiagramBranch = style.Class("dg-branch").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.MinWidth(css.Px(0)),
	).
	Container(compactDiagram, css.FlexDirection(kw.Column), css.AlignItems(kw.Stretch))

var DiagramWays = style.Class("dg-ways").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.6)),
	css.MinWidth(css.Px(0)),
)

var DiagramWay = style.Class("dg-way").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.FlexWrap(kw.Wrap),
		css.Gap(css.Rem(0.5)),
		css.MinWidth(css.Px(0)),
	)

var DiagramCaption = style.Class("dg-caption").
	Set(
		css.Margin(css.Px(0)),
		css.Padding(css.Rem(0.45), css.Rem(0.6)),
		css.Border(css.Px(2), kw.Dashed, css.Var("line")),
		css.Background(css.Var("ground")),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.8125)),
		css.Color(css.Var("ink-2")),
	).
	Before(
		css.Content(css.String("fig. ")),
		css.Color(css.Var("code-str")),
		css.FontWeight(css.Num(700)),
	)
