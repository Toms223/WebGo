package theme

import (
	"slices"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var Reset = style.Any().Set(css.BoxSizing(kw.BorderBox))

var paper = []css.Property{
	css.BackgroundColor(css.Var("ground")),
	css.BackgroundImage(
		css.Raw("linear-gradient(var(--grid) 1px, transparent 1px)"),
		css.Raw("linear-gradient(90deg, var(--grid) 1px, transparent 1px)"),
	),
	css.BackgroundSize(css.Px(32), css.Px(32)),
}

var Body = style.Element("body").
	Set(
		css.Margin(css.Px(0)),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("sans")),
		css.FontSize(css.Px(16)),
		css.LineHeight(css.Num(1.6)),
	).
	Set(paper...)

var reduce = media.PrefersReducedMotion(kw.Reduce)

var stillness = []css.Property{
	css.Animation(kw.None).Important(),
	css.Transition(kw.None).Important(),
}

var ReducedMotion = style.Any().
	Media(reduce, stillness...).
	Nest(
		style.Self().Before().Media(reduce, stillness...),
		style.Self().After().Media(reduce, stillness...),
	)

var line = css.Border(css.Var("border"), kw.Solid, css.Var("line"))

var box = []css.Property{
	line,
	css.BorderRadius(css.Var("r")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow")),
}

var lift = []css.Property{
	css.Transform(css.Raw("translate(-2px, -2px)")),
	css.BoxShadow(css.Var("shadow-l")),
}

var press = []css.Property{
	css.Transform(css.Raw("translate(2px, 2px)")),
	css.BoxShadow(kw.None),
}

var ring = []css.Property{
	css.Outline(css.Px(3), kw.Solid, css.Var("focus")),
	css.OutlineOffset(css.Px(3)),
}

var button = slices.Concat(box, []css.Property{
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.6)),
	css.MinHeight(css.Px(44)),
	css.Padding(css.Rem(0.55), css.Rem(1.1)),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(-0.01)),
	css.LineHeight(css.Num(1.1)),
	css.TextDecoration(kw.None),
	css.Cursor(kw.Pointer),
	css.Transition(css.Raw("transform 120ms ease, box-shadow 120ms ease")),
})

var pill = []css.Property{
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.MinHeight(css.Px(40)),
	css.Padding(css.Rem(0.35), css.Rem(0.95)),
	line,
	css.BorderRadius(css.Px(999)),
	css.Background(css.Var("panel")),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("sans")),
	css.FontSize(css.Rem(0.9)),
	css.FontWeight(css.Num(600)),
	css.LineHeight(css.Num(1.2)),
	css.TextDecoration(kw.None),
	css.WhiteSpace(kw.Nowrap),
	css.Cursor(kw.Pointer),
}

var label = []css.Property{
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.72)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.12)),
	css.TextTransform(kw.Uppercase),
}

var Button = style.Class("button").
	Set(button...).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...)

var ButtonPrimary = style.Class("button-primary").Set(
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
)

var Pill = style.Class("pill").
	Set(pill...).
	Hover(css.Background(css.Var("panel-2"))).
	FocusVisible(ring...)

var Shell = style.Class("shell").
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Px(272), css.Fr(1)),
		css.MinHeight(css.Vh(100)),
	).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)))

var MenuOpen = style.Class("menu-open")

var Side = style.Class("side").
	Set(
		css.Position(kw.Sticky),
		css.Top(css.Px(0)),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(1.5)),
		css.Height(css.Vh(100)),
		css.Padding(css.Rem(1.5), css.Rem(1.25)),
		css.OverflowY(kw.Auto),
		css.Background(css.Var("panel")),
		css.BorderRight(css.Var("border"), kw.Solid, css.Var("line")),
	).
	Media(narrow,
		css.ZIndex(css.Int(10)),
		css.Height(kw.Auto),
		css.FlexDirection(kw.Row),
		css.FlexWrap(kw.Wrap),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.75)),
		css.Padding(css.Rem(0.75), css.Rem(1)),
		css.OverflowY(kw.Visible),
		css.BorderRight(kw.None),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	)

var Logo = style.Class("logo").
	Set(
		css.Display(kw.InlineFlex),
		css.AlignItems(kw.Center),
		css.AlignSelf(kw.FlexStart),
		css.Gap(css.Rem(0.6)),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(1.4)),
		css.FontWeight(css.Num(800)),
		css.LetterSpacing(css.Em(-0.04)),
		css.LineHeight(css.Num(1)),
		css.TextDecoration(kw.None),
	).
	FocusVisible(ring...).
	Media(narrow, css.AlignSelf(kw.Center), css.MarginRight(kw.Auto), css.MinHeight(css.Px(44)))

var LogoMark = style.Class("logo-mark").Set(
	css.Display(kw.Grid),
	css.PlaceItems(kw.Center),
	css.Width(css.Px(36)),
	css.Height(css.Px(36)),
	line,
	css.BorderRadius(css.Px(10)),
	css.Background(css.Var("sun")),
	css.BoxShadow(css.Var("shadow-s")),
	css.Color(css.Var("on-fill")),
	css.FontSize(css.Rem(1.1)),
)

var SideMenu = style.Class("side-menu").
	Set(button...).
	Set(css.Display(kw.None), css.MinHeight(css.Px(40)), css.Padding(css.Rem(0.4), css.Rem(0.9)), css.FontSize(css.Rem(0.9))).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...).
	Media(narrow, css.Display(kw.InlineFlex), css.MinHeight(css.Px(44)))

var ShellToggle = style.Class("shell-toggle").
	Set(button...).
	Set(css.MarginTop(kw.Auto), css.AlignSelf(kw.FlexStart), css.MinHeight(css.Px(40)), css.Padding(css.Rem(0.4), css.Rem(0.9)), css.FontSize(css.Rem(0.9))).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...).
	Media(narrow, css.MarginTop(css.Px(0)), css.AlignSelf(kw.Center), css.MinHeight(css.Px(44)))

var Content = style.Class("content").
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1120px)")),
		css.AlignContent(kw.Start),
		css.MinWidth(css.Px(0)),
		css.Padding(css.Rem(2.5), css.Rem(3), css.Rem(4.5)),
	).
	Media(narrow, css.Padding(css.Rem(1.25), css.Rem(1), css.Rem(3)))

var ContentItem = Content.Child(style.Any()).Set(
	css.GridColumnStart(css.Int(1)),
	css.GridColumnEnd(css.Int(-1)),
	css.MinWidth(css.Px(0)),
)

var Skip = style.Class("skip").
	Set(button...).
	Set(
		css.Position(kw.Absolute),
		css.Top(css.Rem(0.75)),
		css.Left(css.Rem(0.75)),
		css.ZIndex(css.Int(100)),
		css.Transform(css.Raw("translateY(-200%)")),
		css.Background(css.Var("sun")),
		css.Color(css.Var("on-fill")),
	).
	FocusVisible(css.Transform(kw.None)).
	FocusVisible(ring...)
