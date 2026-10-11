package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var Button = style.Class("demo-button").
	Set(
		css.Padding(css.Rem(0.5), css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--surface")),
	).
	Before(css.Content(css.String("→ "))).
	Hover(css.Background(css.Var("--side"))).
	Active(css.Transform(css.Raw("translate(2px, 2px)"))).
	FocusVisible(css.Outline(css.Px(3), kw.Dashed, css.Var("--accent")))

func PseudoDemo() app.UI {
	return app.Button().Type("button").Class(Button.Name()).Text("Hover, press or tab to me")
}
