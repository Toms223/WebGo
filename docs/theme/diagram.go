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
	css.BorderRadius(css.Var("r")),
	css.Background(css.Var("ground")),
	css.BackgroundImage(css.Raw("linear-gradient(var(--grid) 2px, transparent 2px), linear-gradient(90deg, var(--grid) 2px, transparent 2px)")),
	css.BackgroundSize(css.Raw("32px 32px")),
)

var DiagramLane = style.Class("dg-lane").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.65)),
	css.Padding(css.Rem(0.85), css.Rem(1), css.Rem(1)),
	line,
	css.BorderRadius(css.Var("r")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow")),
	css.MinWidth(css.Px(0)),
)

var DiagramTag = style.Class("dg-tag").
	Set(label...).
	Set(
		css.AlignSelf(kw.FlexStart),
		css.Padding(css.Rem(0.2), css.Rem(0.55)),
		css.BorderRadius(css.Px(6)),
		css.Background(css.Var("ink")),
		css.Color(css.Var("ground")),
	)

var DiagramTitle = style.Class("dg-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.1)),
	css.FontWeight(css.Num(800)),
	css.LineHeight(css.Num(1.2)),
	css.Color(css.Var("ink")),
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
		css.BorderRadius(css.Px(10)),
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Container(compactDiagram, css.MaxWidth(kw.None))

var DiagramSun = DiagramBox.And(style.Class("dg-sun")).Set(css.Background(css.Var("sun")), css.Color(css.Var("on-fill")))

var DiagramLime = DiagramBox.And(style.Class("dg-lime")).Set(css.Background(css.Var("lime")), css.Color(css.Var("on-fill")))

var DiagramLilac = DiagramBox.And(style.Class("dg-lilac")).Set(css.Background(css.Var("lilac")), css.Color(css.Var("on-fill")))

var DiagramCoral = DiagramBox.And(style.Class("dg-coral")).Set(css.Background(css.Var("coral")), css.Color(css.Var("on-fill")))

var DiagramName = style.Class("dg-name").Set(
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(0.95)),
	css.FontWeight(css.Num(800)),
	css.LineHeight(css.Num(1.2)),
)

var DiagramSub = style.Class("dg-sub").Set(
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.74)),
	css.LineHeight(css.Num(1.35)),
	css.Opacity(css.Num(0.8)),
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
	css.FontSize(css.Rem(0.72)),
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

var DiagramCaption = style.Class("dg-caption").Set(
	css.Margin(css.Px(0)),
	css.FontSize(css.Rem(0.85)),
	css.Color(css.Var("ink-2")),
)
