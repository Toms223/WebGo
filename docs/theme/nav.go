package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var SideNav = style.Class("side-nav").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		line,
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow")),
		css.Transform(css.Raw("rotate(-0.8deg)")),
	).
	Media(narrow,
		css.Display(kw.None),
		css.Order(css.Int(1)),
		css.FlexBasis(css.Pct(100)),
		css.Transform(kw.None),
	)

var NavLabel = style.Class("nav-label").
	Set(label...).
	Set(
		css.Display(kw.Flex),
		css.JustifyContent(kw.SpaceBetween),
		css.Padding(css.Rem(0.6), css.Rem(0.9)),
		css.BorderBottom(css.Px(2), kw.Solid, css.Var("line")),
		css.Background(css.Var("sun")),
		css.Color(css.Var("on-fill")),
	)

var SideNavList = SideNav.Child(style.Element("ul")).Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(0.4), css.Px(0)),
	css.ListStyle(kw.None),
	css.CounterReset(css.Ident("nav")),
)

var NavGroupGap = SideNavList.Next(NavLabel).Set(
	css.BorderTop(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("coral")),
)

var NavLink = style.Class("nav-link").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Baseline),
		css.Gap(css.Rem(0.6)),
		css.MinHeight(css.Px(40)),
		css.Padding(css.Rem(0.45), css.Rem(0.9)),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("sans")),
		css.FontSize(css.Rem(1)),
		css.FontWeight(css.Num(500)),
		css.LineHeight(css.Num(1.35)),
		css.TextDecoration(kw.None),
		css.CounterIncrement(css.Ident("nav")),
		css.Transition(css.Raw("background 120ms ease, padding 150ms ease")),
	).
	Before(
		css.Content(css.Raw("counter(nav, decimal-leading-zero)")),
		css.Flex(kw.None),
		css.Color(css.Var("ink-2")),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.8)),
	).
	Hover(css.Background(css.Var("panel-2")), css.PaddingLeft(css.Rem(1.25))).
	FocusVisible(ring...)

var NavLinkActive = style.Class("nav-link-active").
	Set(
		css.Background(css.Var("ink-blue")),
		css.Color(css.Var("on-blue")),
		css.FontWeight(css.Num(600)),
	).
	Before(css.Color(css.Var("ink-yellow"))).
	Hover(css.Background(css.Var("ink-blue")))

var MenuOpenNav = Shell.And(MenuOpen).Descendant(SideNav).Media(narrow, css.Display(kw.Flex))
