package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Swatch = style.Class("demo-swatch").
	Set(
		css.Custom("demo-size", css.Rem(4)),
		css.Width(css.Var("--demo-size")),
		css.Height(css.Calc(css.Div(css.Var("--demo-size"), css.Num(2)))),
		css.Background(css.HSL(200, 80, 60)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	)

func ValuesDemo() app.UI {
	return app.Div().Style(Swatch)
}
