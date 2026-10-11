package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const OutlineAction = "demo/outline"

var DemoOutlineBase = style.Class("demo-outlined").
	Set(
		css.Padding(css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	)

var DemoOutline = style.Class("demo-outlined").
	Set(
		css.Outline(css.Px(3), kw.Dashed, css.Var("--accent")),
		css.OutlineOffset(css.Px(4)),
	).
	On(OutlineAction, outlined)

func outlined(a app.Action) bool {
	on, ok := a.Value.(bool)
	return ok && on
}

type onDemo struct{ app.Compo }

func OnDemo() app.UI { return &onDemo{} }

func (d *onDemo) OnMount(ctx app.Context) {
	ctx.NewActionWithValue(OutlineAction, false)
}

func (d *onDemo) Render() app.UI {
	return app.Div().Body(
		app.Label().Body(
			app.Input().Type("checkbox").OnChange(d.toggle),
			app.Text(" Outline"),
		),
		app.Div().Class(DemoOutline.Name()).Text("Outlined while the box is checked"),
	)
}

func (d *onDemo) toggle(ctx app.Context, e app.Event) {
	ctx.NewActionWithValue(OutlineAction, ctx.JSSrc().Get("checked").Bool())
}
