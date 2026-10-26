package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Pulse = style.Keyframes("demo-pulse").
	From(css.Transform(css.Raw("scale(1)"))).
	To(css.Transform(css.Raw("scale(1.2)")))

var Pulsing = style.Class("demo-pulsing").
	Set(
		css.Width(css.Rem(3)),
		css.Height(css.Rem(3)),
		css.Background(css.Var("--accent")),
		css.AnimationName(css.Ident(Pulse.Name())),
		css.AnimationDuration(css.Ms(800)),
		css.AnimationIterationCount(kw.Infinite),
		css.AnimationDirection(kw.Alternate),
	)

func KeyframesDemo() app.UI {
	return app.Div().Style(Pulsing)
}
