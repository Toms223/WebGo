package demos

import (
	"time"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var SiblingRow = style.Class("demo-siblings").Set(
	css.Display(kw.Flex),
	css.Gap(css.Rem(0.75)),
)

var SiblingBox = style.Class("demo-sibling").Set(
	css.Padding(css.Rem(0.5), css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.FontWeight(css.Num(700)),
)

var SiblingReady = style.Class("demo-sibling-ready").Set(
	css.Background(css.Var("--mint")),
)

type slowPart struct {
	app.Compo
	Delay time.Duration
	Run   int
}

func (p *slowPart) OnLoad(ctx app.Context) error {
	time.Sleep(p.Delay)
	return nil
}

func (p *slowPart) Loading() app.UI {
	return app.Div().Style(SiblingBox).Text("Loading for " + p.Delay.String() + "…")
}

func (p *slowPart) Render() app.UI {
	return app.Div().Style(SiblingBox, SiblingReady).Text("Ready after " + p.Delay.String())
}

type siblingsDemo struct {
	app.Compo
	runs int
}

func SiblingsDemo() app.UI { return &siblingsDemo{} }

func (d *siblingsDemo) Render() app.UI {
	return app.Div().Body(
		app.Div().Style(SiblingRow).Body(
			&slowPart{Delay: 300 * time.Millisecond, Run: d.runs},
			&slowPart{Delay: 900 * time.Millisecond, Run: d.runs},
		),
		app.Button().Type("button").Text("Load again").OnClick(d.again),
	)
}

func (d *siblingsDemo) again(ctx app.Context, e app.Event) {
	d.runs++
}

const siblingsSource = `type slowPart struct {
	app.Compo
	Delay time.Duration
	Run   int
}

func (p *slowPart) OnLoad(ctx app.Context) error {
	time.Sleep(p.Delay)
	return nil
}

func (p *slowPart) Loading() app.UI {
	return app.Div().Style(SiblingBox).Text("Loading for " + p.Delay.String() + "…")
}

func (p *slowPart) Render() app.UI {
	return app.Div().Style(SiblingBox, SiblingReady).Text("Ready after " + p.Delay.String())
}

type siblingsDemo struct {
	app.Compo
	runs int
}

func SiblingsDemo() app.UI { return &siblingsDemo{} }

func (d *siblingsDemo) Render() app.UI {
	return app.Div().Body(
		app.Div().Style(SiblingRow).Body(
			&slowPart{Delay: 300 * time.Millisecond, Run: d.runs},
			&slowPart{Delay: 900 * time.Millisecond, Run: d.runs},
		),
		app.Button().Type("button").Text("Load again").OnClick(d.again),
	)
}

func (d *siblingsDemo) again(ctx app.Context, e app.Event) {
	d.runs++
}`

var SiblingsExample = example.Example{Name: "siblings.go", Source: siblingsSource, Live: SiblingsDemo}
