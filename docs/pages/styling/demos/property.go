package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Angle = style.Property("demo-angle",
	css.Syntax(css.SyntaxAngle),
	css.Inherits(kw.False),
	css.InitialValue(css.Deg(0)),
)

var Turning = style.Class("demo-turning").
	Set(
		css.Width(css.Rem(10)),
		css.Height(css.Rem(4)),
		css.Background(css.Raw("linear-gradient(var(--demo-angle), var(--mint), var(--pink))")),
		css.Transition(css.Raw("--demo-angle 600ms")),
	).
	Hover(css.Custom(Angle.Name(), css.Deg(180)))

func PropertyDemo() app.UI {
	return app.Div().Style(Turning)
}
