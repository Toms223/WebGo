package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var Card = style.Class("demo-card").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Hover(css.Background(css.Var("--pink")))

func CardDemo() app.UI {
	return app.Div().Class(Card.Name()).Text("Hover this card")
}
