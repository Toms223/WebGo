package demos

import (
	"strconv"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var CounterBox = style.Class("demo-counter").Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.5), css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.Background(css.Var("--mint")),
	css.FontWeight(css.Num(700)),
)

const CounterKey = "demo/counter"

type Counter struct{ Count int }

var counter = mustModel(CounterKey, &Counter{})

type counterDemo struct{ app.Compo }

func CounterDemo() app.UI { return &counterDemo{} }

func (d *counterDemo) OnMount(ctx app.Context) {
	counter.Get(ctx).Observe(ctx)
}

func (d *counterDemo) Render() app.UI {
	return app.Div().Body(
		app.Div().Class(CounterBox.Name()).Body(
			app.Text("Count "),
			app.Output().Text(strconv.Itoa(counter.State().Count)),
		),
		app.Button().Type("button").Text("+1").OnClick(d.bump),
		app.Button().Type("button").Text("Reset").OnClick(d.reset),
		app.Button().Type("button").Text("Save").OnClick(d.save),
	)
}

func (d *counterDemo) bump(ctx app.Context, e app.Event) {
	counter.Load(ctx).Apply(func(ctx app.Context, c *Counter) *Counter {
		c.Count++
		return c
	}).Set()
}

func (d *counterDemo) reset(ctx app.Context, e app.Event) {
	counter.Load(ctx).Apply(func(ctx app.Context, c *Counter) *Counter {
		c.Count = 0
		return c
	}).Set()
}

func (d *counterDemo) save(ctx app.Context, e app.Event) {
	counter.Load(ctx).SetAndPersist(counter.State())
}

func mustModel[T any](key string, value *T) *data.ViewModel[T] {
	m, err := data.NewViewModel(key, value)
	if err != nil {
		panic(err)
	}
	return m
}

func Items() []style.Item {
	return []style.Item{CounterBox}
}

const counterSource = `const CounterKey = "demo/counter"

type Counter struct{ Count int }

var counter = mustModel(CounterKey, &Counter{})

type counterDemo struct{ app.Compo }

func CounterDemo() app.UI { return &counterDemo{} }

func (d *counterDemo) OnMount(ctx app.Context) {
	counter.Get(ctx).Observe(ctx)
}

func (d *counterDemo) Render() app.UI {
	return app.Div().Body(
		app.Div().Class(CounterBox.Name()).Body(
			app.Text("Count "),
			app.Output().Text(strconv.Itoa(counter.State().Count)),
		),
		app.Button().Type("button").Text("+1").OnClick(d.bump),
		app.Button().Type("button").Text("Reset").OnClick(d.reset),
		app.Button().Type("button").Text("Save").OnClick(d.save),
	)
}

func (d *counterDemo) bump(ctx app.Context, e app.Event) {
	counter.Load(ctx).Apply(func(ctx app.Context, c *Counter) *Counter {
		c.Count++
		return c
	}).Set()
}

func (d *counterDemo) reset(ctx app.Context, e app.Event) {
	counter.Load(ctx).Apply(func(ctx app.Context, c *Counter) *Counter {
		c.Count = 0
		return c
	}).Set()
}

func (d *counterDemo) save(ctx app.Context, e app.Event) {
	counter.Load(ctx).SetAndPersist(counter.State())
}`

var CounterExample = example.Example{Name: "counter.go", Source: counterSource, Live: CounterDemo}

func Examples() []example.Example {
	return []example.Example{CounterExample}
}
