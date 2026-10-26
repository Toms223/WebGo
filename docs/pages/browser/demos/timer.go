package demos

import (
	"fmt"
	"time"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var TimerBox = style.Class("demo-browser-timer").Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.5), css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.Background(css.Var("--mint")),
	css.FontWeight(css.Num(700)),
)

type timerDemo struct {
	app.Compo
	tenths int
	stop   chan struct{}
}

func TimerDemo() app.UI { return &timerDemo{} }

func (d *timerDemo) Render() app.UI {
	label := "Start"
	if d.stop != nil {
		label = "Stop"
	}
	return app.Div().Style(TimerBox).Body(
		app.Output().Text(fmt.Sprintf("%d.%d s", d.tenths/10, d.tenths%10)),
		app.Button().Type("button").Text(label).OnClick(d.toggle),
	)
}

func (d *timerDemo) toggle(ctx app.Context, e app.Event) {
	if d.stop != nil {
		d.halt()
		return
	}
	stop := make(chan struct{})
	d.stop = stop
	ctx.Async(func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				ctx.Dispatch(func(ctx app.Context) {
					d.tenths++
				})
			}
		}
	})
}

func (d *timerDemo) OnDismount() {
	if d.stop != nil {
		d.halt()
	}
}

func (d *timerDemo) halt() {
	close(d.stop)
	d.stop = nil
}

func Items() []style.Item {
	return []style.Item{TimerBox}
}

const timerSource = `type timerDemo struct {
	app.Compo
	tenths int
	stop   chan struct{}
}

func TimerDemo() app.UI { return &timerDemo{} }

func (d *timerDemo) Render() app.UI {
	label := "Start"
	if d.stop != nil {
		label = "Stop"
	}
	return app.Div().Style(TimerBox).Body(
		app.Output().Text(fmt.Sprintf("%d.%d s", d.tenths/10, d.tenths%10)),
		app.Button().Type("button").Text(label).OnClick(d.toggle),
	)
}

func (d *timerDemo) toggle(ctx app.Context, e app.Event) {
	if d.stop != nil {
		d.halt()
		return
	}
	stop := make(chan struct{})
	d.stop = stop
	ctx.Async(func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				ctx.Dispatch(func(ctx app.Context) {
					d.tenths++
				})
			}
		}
	})
}

func (d *timerDemo) OnDismount() {
	if d.stop != nil {
		d.halt()
	}
}

func (d *timerDemo) halt() {
	close(d.stop)
	d.stop = nil
}`

var TimerExample = example.Example{Name: "timer.go", Source: timerSource, CSS: TimerBox.CSS(), Live: TimerDemo}

func Examples() []example.Example {
	return []example.Example{TimerExample}
}
