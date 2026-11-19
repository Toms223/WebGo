package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/script"
	"github.com/Toms223/WebGo/style"
)

var ConfettiBox = style.Class("demo-scripts-confetti").Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.5), css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.Background(css.Var("--mint")),
	css.FontWeight(css.Num(700)),
)

type Confetti struct{}

func (Confetti) Source() app.ModuleSource { return script.CDN("npm", "canvas-confetti", "1.9.3") }

func (Confetti) Init(ctx app.Context, mod app.Value) Cannon { return Cannon{fire: mod.Get("default")} }

type Cannon struct{ fire app.Value }

func (c Cannon) Fire() { c.fire.Invoke(map[string]any{"particleCount": 120, "spread": 70}) }

type confettiDemo struct {
	app.Compo
	missing bool
}

func ConfettiDemo() app.UI { return &confettiDemo{} }

func (d *confettiDemo) Render() app.UI {
	return app.Div().Style(ConfettiBox).Body(
		app.Button().Type("button").Text("Fire").OnClick(d.fire),
		app.If(d.missing, func() app.UI { return app.Span().Text("module not loaded") }),
	)
}

func (d *confettiDemo) fire(ctx app.Context, e app.Event) {
	c, ok := script.Use(ctx, Confetti{})
	d.missing = !ok
	if ok {
		c.Fire()
	}
}

func Items() []style.Item {
	return []style.Item{ConfettiBox}
}

const confettiSource = `type Confetti struct{}

func (Confetti) Source() app.ModuleSource { return script.CDN("npm", "canvas-confetti", "1.9.3") }

func (Confetti) Init(ctx app.Context, mod app.Value) Cannon { return Cannon{fire: mod.Get("default")} }

type Cannon struct{ fire app.Value }

func (c Cannon) Fire() { c.fire.Invoke(map[string]any{"particleCount": 120, "spread": 70}) }

type confettiDemo struct {
	app.Compo
	missing bool
}

func ConfettiDemo() app.UI { return &confettiDemo{} }

func (d *confettiDemo) Render() app.UI {
	return app.Div().Style(ConfettiBox).Body(
		app.Button().Type("button").Text("Fire").OnClick(d.fire),
		app.If(d.missing, func() app.UI { return app.Span().Text("module not loaded") }),
	)
}

func (d *confettiDemo) fire(ctx app.Context, e app.Event) {
	c, ok := script.Use(ctx, Confetti{})
	d.missing = !ok
	if ok {
		c.Fire()
	}
}`

var ConfettiExample = example.Example{Name: "confetti.go", Source: confettiSource, CSS: ConfettiBox.CSS(), Live: ConfettiDemo}

func Examples() []example.Example {
	return []example.Example{ConfettiExample}
}
