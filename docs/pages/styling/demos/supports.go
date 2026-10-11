package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/style"
	"github.com/Toms223/WebGo/supports"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var Mixed = style.Class("demo-mixed").
	Set(css.Padding(css.Rem(1)), css.BackgroundColor(css.Var("--mint"))).
	Supports(
		supports.Property(css.BackgroundColor(css.Raw("color-mix(in srgb, red, blue)"))),
		css.BackgroundColor(css.Raw("color-mix(in srgb, var(--mint), var(--pink))")),
	)

func SupportsDemo() app.UI {
	return app.Div().Class(Mixed.Name()).Text("Mint, or mint mixed with pink where color-mix is supported")
}
