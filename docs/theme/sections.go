package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Sections = style.Class("sections").
	Set(
		css.Position(kw.Sticky),
		css.Top(css.Rem(0.75)),
		css.ZIndex(css.Int(2)),
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.Gap(css.Rem(0.5)),
		css.Margin(css.Px(0), css.Px(0), css.Rem(1.75)),
		css.Padding(css.Rem(0.6)),
		line,
		css.BorderRadius(css.Var("r-l")),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow")),
	).
	Media(narrow, css.Position(kw.Static))

var SectionLink = style.Class("section-link").
	Set(pill...).
	Hover(css.Background(css.Var("panel-2"))).
	FocusVisible(ring...)

var SectionLinkActive = style.Class("section-link-active").
	Set(
		css.Background(css.Var("tab")),
		css.Color(css.Var("tab-on")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Hover(css.Background(css.Var("tab")))
