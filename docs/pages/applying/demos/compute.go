package demos

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const AccentKey = "demo/accent"

type Accent struct{ Color string }

var DemoAccent = style.Class("demo-accent").
	Bind(style.Compute(AccentKey, func(a *Accent) []css.Property {
		return []css.Property{
			css.Color(css.Hex(a.Color)),
			css.BorderBottom(css.Px(3), kw.Solid, css.Hex(a.Color)),
		}
	}))

const (
	Red   = "#FF3B30"
	Blue  = "#2F6BFF"
	Green = "#12A150"
)

var accent = model(AccentKey, startAccent())

func startAccent() *Accent { return &Accent{Color: Red} }

type computeDemo struct{ app.Compo }

func ComputeDemo() app.UI { return &computeDemo{} }

func (d *computeDemo) OnMount(ctx app.Context) {
	accent.OnAction(ctx)
	accent.Load(ctx).Set()
}

func (d *computeDemo) Render() app.UI {
	return app.Div().Body(
		app.Button().Type("button").Text("Red").OnClick(d.pick(Red)),
		app.Button().Type("button").Text("Blue").OnClick(d.pick(Blue)),
		app.Button().Type("button").Text("Green").OnClick(d.pick(Green)),
		app.P().Class(DemoAccent.Name()).Text("This text takes the chosen colour"),
	)
}

func (d *computeDemo) pick(color string) app.EventHandler {
	return func(ctx app.Context, e app.Event) {
		accent.Load(ctx).Apply(func(ctx app.Context, a *Accent) *Accent {
			a.Color = color
			return a
		}).Set()
	}
}
