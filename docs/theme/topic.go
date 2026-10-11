package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Topic = style.Class("topic").Set(
	css.Display(kw.Grid),
	css.GridTemplateColumns(kw.Subgrid),
	css.AlignItems(kw.Start),
)

var TopicItem = Topic.Child(style.Any()).Set(
	css.GridColumnStart(css.Int(1)),
	css.GridColumnEnd(css.Int(-1)),
	css.MinWidth(css.Px(0)),
)

var TopicTitle = style.Class("topic-title").Set(
	css.JustifySelf(kw.Start),
	css.Margin(css.Rem(0.5), css.Px(0), css.Rem(2)),
	css.Padding(css.Em(0.1), css.Em(0.45)),
	css.Border(css.Var("border-l"), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Var("r-l")),
	css.BoxShadow(css.Var("shadow-l")),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2), css.Vw(4.5), css.Rem(3.25))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.035)),
	css.LineHeight(css.Num(1.05)),
	css.Transform(css.Raw("rotate(-1deg)")),
)
