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
}

var grain = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Fixed),
	css.Inset(css.Px(0)),
	css.ZIndex(css.Int(200)),
	css.PointerEvents(kw.None),
	css.BackgroundImage(css.Var("noise")),
	css.Opacity(css.Var("grain-opacity")),
	css.MixBlendMode(css.Var("blend")),
}

var Body = style.Element("body").
	Set(
		css.Margin(css.Px(0)),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("sans")),
		css.FontSize(css.Px(17)),
		css.LineHeight(css.Num(1.6)),
		css.OverflowX(kw.Clip),
	).
	Set(paper...).
	Before(grain...)

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
	css.Transform(css.Raw("translate(-2px, -2px) rotate(-1deg)")),
	css.BoxShadow(css.Raw("7px 7px 0 var(--shadow-ink)")),
}

var press = []css.Property{
	css.Transform(css.Raw("translate(4px, 4px)")),
	css.BoxShadow(css.Raw("1px 1px 0 var(--shadow-ink)")),
}

var ring = []css.Property{
	css.Outline(css.Px(3), kw.Solid, css.Var("focus")),
	css.OutlineOffset(css.Px(3)),
}

var misprint = []css.Property{
	css.Color(css.Var("mis-ink")),
	css.TextShadow(css.Var("mis-shadow")),
}

var button = slices.Concat(box, []css.Property{
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.6)),
	css.MinHeight(css.Px(44)),
	css.Padding(css.Rem(0.6), css.Rem(1.15)),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("sans")),
	css.FontSize(css.Rem(1)),
	css.FontWeight(css.Num(700)),
	css.LineHeight(css.Num(1.1)),
	css.TextDecoration(kw.None),
	css.Cursor(kw.Pointer),
	css.Transition(css.Raw("transform 120ms ease, box-shadow 120ms ease")),
})

var pill = []css.Property{
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.MinHeight(css.Px(40)),
	css.Padding(css.Rem(0.35), css.Rem(0.8)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Var("r")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow-s")),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("sans")),
	css.FontSize(css.Rem(0.95)),
	css.FontWeight(css.Num(600)),
	css.LineHeight(css.Num(1.2)),
	css.TextDecoration(kw.None),
	css.WhiteSpace(kw.Nowrap),
	css.Cursor(kw.Pointer),
	css.Transition(css.Raw("transform 150ms ease, background 150ms ease")),
}

var pillOn = []css.Property{
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.Transform(css.Raw("rotate(-2deg)")),
}

var label = []css.Property{
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.8)),
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
	Hover(pillOn...).
	FocusVisible(ring...)

var Shell = style.Class("shell").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Px(288), css.Fr(1)),
		css.GridTemplateRows(kw.Auto, css.Fr(1)),
		css.MinHeight(css.Vh(100)),
	).
	Before(
		css.Content(css.String("WebGo zine · Guides · printed in three inks on app.wasm")),
		css.GridColumn(css.Raw("1 / -1")),
		css.Padding(css.Rem(0.5), css.Rem(1.5)),
		css.Background(css.Var("tab")),
		css.Color(css.Var("tab-on")),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.8)),
		css.FontWeight(css.Num(700)),
		css.LetterSpacing(css.Em(0.06)),
		css.LineHeight(css.Num(1.4)),
		css.TextTransform(kw.Uppercase),
		css.WhiteSpace(kw.Nowrap),
		css.Overflow(kw.Hidden),
		css.TextOverflow(kw.Ellipsis),
	).
	After(
		css.Content(css.String("Free. Take one.")),
		css.Position(kw.Absolute),
		css.Top(css.Px(0)),
		css.Right(css.Rem(1.5)),
		css.Padding(css.Rem(0.5), css.Px(0), css.Rem(0.5), css.Rem(1)),
		css.Background(css.Var("tab")),
		css.Color(css.Var("tab-on")),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.8)),
		css.FontWeight(css.Num(700)),
		css.LetterSpacing(css.Em(0.06)),
		css.LineHeight(css.Num(1.4)),
		css.TextTransform(kw.Uppercase),
	).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.GridTemplateRows(kw.Auto, kw.Auto, css.Fr(1))).
	Nest(
		style.Self().Before().Media(narrow, css.Content(css.String("WebGo zine · Guides")), css.Padding(css.Rem(0.5), css.Rem(1))),
		style.Self().After().Media(narrow, css.Right(css.Rem(1))),
	)

var MenuOpen = style.Class("menu-open")

var Side = style.Class("side").
	Set(
		css.Position(kw.Sticky),
		css.Top(css.Px(0)),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(1.75)),
		css.Height(css.Vh(100)),
		css.Padding(css.Rem(1.75), css.Rem(1.25), css.Rem(1.5), css.Rem(1.5)),
		css.OverflowY(kw.Auto),
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
		css.Background(css.Var("panel")),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	)

var Logo = style.Class("logo").
	Set(
		css.Display(kw.InlineFlex),
		css.AlignItems(kw.Center),
		css.AlignSelf(kw.FlexStart),
		css.Gap(css.Rem(0.55)),
		css.MinHeight(css.Px(44)),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(2.4)),
		css.FontWeight(css.Num(400)),
		css.LetterSpacing(css.Em(0.01)),
		css.LineHeight(css.Num(1)),
		css.TextTransform(kw.Uppercase),
		css.TextShadow(css.Raw("0.05em 0.035em 0 var(--mis)")),
		css.TextDecoration(kw.None),
	).
	FocusVisible(ring...).
	Media(narrow, css.AlignSelf(kw.Center), css.MarginRight(kw.Auto), css.FontSize(css.Rem(1.9)))

var LogoMark = style.Class("logo-mark").Set(
	css.Display(kw.Grid),
	css.PlaceItems(kw.Center),
	css.Width(css.Px(34)),
	css.Height(css.Px(34)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("sun")),
	css.BoxShadow(css.Var("shadow-s")),
	css.Color(css.Var("on-fill")),
	css.FontSize(css.Rem(1.3)),
	css.TextShadow(kw.None),
	css.Transform(css.Raw("rotate(-4deg)")),
)

var SideMenu = style.Class("side-menu").
	Set(button...).
	Set(css.Display(kw.None), css.MinHeight(css.Px(40)), css.Padding(css.Rem(0.4), css.Rem(0.9)), css.FontSize(css.Rem(0.95)), css.Background(css.Var("sun")), css.Color(css.Var("on-fill"))).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...).
	Media(narrow, css.Display(kw.InlineFlex), css.MinHeight(css.Px(44)))

var ShellToggle = style.Class("shell-toggle").
	Set(button...).
	Set(css.MarginTop(kw.Auto), css.AlignSelf(kw.FlexStart), css.MinHeight(css.Px(40)), css.Padding(css.Rem(0.4), css.Rem(0.9)), css.FontSize(css.Rem(0.95))).
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
		css.Padding(css.Rem(2.75), css.Rem(3), css.Rem(4.5)),
	).
	Media(narrow, css.Padding(css.Rem(1.5), css.Rem(1), css.Rem(3)))

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
