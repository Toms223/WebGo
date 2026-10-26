package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var PanelTitle = style.Class("demo-panel-title").
	Set(css.FontWeight(css.Num(900)))

var Panel = style.Class("demo-panel").
	Set(css.Padding(css.Rem(1)), css.Border(css.Px(3), kw.Solid, css.Var("--ink"))).
	Nest(
		PanelTitle,
		style.Element("a").Hover(css.Background(css.Var("--side"))),
		style.Self().Hover(css.Background(css.Var("--mint"))),
		style.Self().Media(media.MaxWidth(css.Rem(48)), css.Padding(css.Rem(0.5))),
	)

func NestingDemo() app.UI {
	return app.Div().Style(Panel).Body(
		app.P().Style(PanelTitle).Text("A nested title"),
		app.A().Href("#nesting").Text("A nested link"),
	)
}
