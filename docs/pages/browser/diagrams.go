package browser

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func goroutinesDiagram() app.UI {
	return d.Figure("The UI goroutine owns every component. Blocking work runs on an Async goroutine and hands its result back with Dispatch.",
		d.Lane("UI goroutine", "One thing at a time",
			d.Flow(d.Box(d.Sun, "event handler", "calls ctx.Async(work)"), d.Box(d.Plain, "keeps going", "renders, other events"), d.Box(d.Lime, "Dispatch's fn", "sets fields"), d.Via("next frame"), d.Box(d.Lilac, "re-render", "component + ancestors"))),
		d.Lane("Async goroutine", "Blocking work",
			d.Flow(d.Box(d.Plain, "work", "network, disk, CPU"), d.Via("ctx.Dispatch(fn)"), d.Box(d.Lime, "fn queued", "on the UI goroutine"))),
	)
}
