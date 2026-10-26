package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var Wide = style.Class("demo-wide").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Media(media.MaxWidth(css.Rem(48)), css.Background(css.Var("--pink")))

func MediaDemo() app.UI {
	return app.Div().Style(Wide).Text("Mint on a wide screen, pink at 48rem and below")
}
