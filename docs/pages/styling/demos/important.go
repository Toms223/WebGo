package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var Quiet = style.Id("demo-quiet").
	Set(css.Background(css.Var("--mint")))

var Loud = style.Class("demo-loud").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--pink")).Important())

func ImportantDemo() app.UI {
	return app.Div().ID(Quiet.Name()).Class(Loud.Name()).Text("The id says mint, the important class says pink")
}
