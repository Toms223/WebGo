package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
)

var List = style.Class("demo-list")

var ListItem = List.Child(style.Element("li")).
	Set(css.Padding(css.Rem(0.25), css.Rem(0.75))).
	NthChild(css.Odd, css.Background(css.Var("--mint")))

func SelectorsDemo() app.UI {
	return app.Ul().Style(List).Body(
		app.Li().Text("First"),
		app.Li().Text("Second"),
		app.Li().Text("Third"),
	)
}
