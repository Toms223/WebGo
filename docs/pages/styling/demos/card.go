package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
)

var Card = style.Class("demo-card").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Hover(css.Background(css.Var("--pink")))

func CardDemo() app.UI {
	return app.Div().Style(Card).Text("Hover this card")
}
