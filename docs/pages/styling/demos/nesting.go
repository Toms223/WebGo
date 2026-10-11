package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
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
	return app.Div().Class(Panel.Name()).Body(
		app.P().Class(PanelTitle.Name()).Text("A nested title"),
		app.A().Href("#nesting").Text("A nested link"),
	)
}
