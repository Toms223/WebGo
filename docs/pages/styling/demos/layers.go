package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
)

var LayeredId = style.Id("demo-layered").
	Set(css.Background(css.Var("--pink")))

var Layered = style.Class("demo-layered").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint")))

var Under = style.Layer("demo-under", LayeredId)

var Over = style.Layer("demo-over", Layered)

func LayersDemo() app.UI {
	return app.Div().ID(LayeredId.Name()).Style(Layered).Text("The id says pink, the later layer says mint")
}
