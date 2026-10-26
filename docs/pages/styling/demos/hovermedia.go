package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var Tile = style.Class("demo-tile").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Nest(
		style.Self().
			Hover(css.Background(css.Var("--pink"))).
			Media(media.MaxWidth(css.Rem(48)), css.Background(css.Var("--side"))),
	)

func HoverMediaDemo() app.UI {
	return app.Div().Style(Tile).Text("Hover: pink on a wide screen, yellow at 48rem and below")
}
