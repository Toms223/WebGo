package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var SideNav = style.Class("side-nav").
	Set(css.Display(kw.Flex), css.FlexDirection(kw.Column), css.Gap(css.Rem(0.5))).
	Media(narrow,
		css.Display(kw.None),
		css.Order(css.Int(1)),
		css.FlexBasis(css.Pct(100)),
	)

var NavLabel = style.Class("nav-label").
	Set(label...).
	Set(css.Padding(css.Px(0), css.Rem(0.9)), css.Color(css.Var("ink-2")))

var SideNavList = SideNav.Child(style.Element("ul")).Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.25)),
	css.Margin(css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
)

var NavLink = style.Class("nav-link").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.JustifyContent(kw.SpaceBetween),
		css.Gap(css.Rem(0.5)),
		css.MinHeight(css.Px(44)),
		css.Padding(css.Rem(0.5), css.Rem(0.9)),
		css.Border(css.Var("border"), kw.Solid, kw.Transparent),
		css.BorderRadius(css.Var("r")),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(1)),
		css.FontWeight(css.Num(700)),
		css.LineHeight(css.Num(1.2)),
		css.TextDecoration(kw.None),
	).
	Hover(css.Background(css.Var("ground")), css.BorderColor(css.Var("line"))).
	FocusVisible(ring...)

var NavLinkActive = style.Class("nav-link-active").
	Set(
		css.Background(css.Var("sun")),
		css.Color(css.Var("on-fill")),
		css.BorderColor(css.Var("line")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Hover(css.Background(css.Var("sun")))

var MenuOpenNav = Shell.And(MenuOpen).Descendant(SideNav).Media(narrow, css.Display(kw.Flex))
