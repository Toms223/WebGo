package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const CompactKey = "demo/compact"

type Compact struct{ On bool }

var compact = model(CompactKey, &Compact{})

var DemoCompactBase = style.Class("demo-compact").
	Set(
		css.Padding(css.Rem(2)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--mint")),
	)

var DemoCompact = style.Class("demo-compact").
	Set(css.Padding(css.Rem(0.25))).
	Bind(style.When(CompactKey, func(c *Compact) bool { return c.On }))

type whenDemo struct{ app.Compo }

func WhenDemo() app.UI { return &whenDemo{} }

func (d *whenDemo) OnMount(ctx app.Context) {
	compact.OnAction(ctx)
}

func (d *whenDemo) Render() app.UI {
	return app.Div().Body(
		app.Label().Body(
			app.Input().Type("checkbox").Checked(compact.State().On).OnChange(d.toggle),
			app.Text(" Compact"),
		),
		app.Div().Class(DemoCompact.Name()).Text("The padding follows the checkbox"),
	)
}

func (d *whenDemo) toggle(ctx app.Context, e app.Event) {
	on := ctx.JSSrc().Get("checked").Bool()
	compact.Load(ctx).Apply(func(ctx app.Context, c *Compact) *Compact {
		c.On = on
		return c
	}).Set()
}
