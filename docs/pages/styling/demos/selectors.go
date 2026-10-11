package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var List = style.Class("demo-list")

var ListItem = List.Child(style.Element("li")).
	Set(css.Padding(css.Rem(0.25), css.Rem(0.75))).
	NthChild(css.Odd, css.Background(css.Var("--mint")))

func SelectorsDemo() app.UI {
	return app.Ul().Class(List.Name()).Body(
		app.Li().Text("First"),
		app.Li().Text("Second"),
		app.Li().Text("Third"),
	)
}
