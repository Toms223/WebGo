package actions

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func handlersDiagram() app.UI {
	return d.Figure("An action reaches every handler of its name: the global one on its own goroutine, each component handler on the UI goroutine.",
		d.Lane("", "",
			d.Branch(d.Box(d.Sun, "ctx.NewAction(name)", "Value · Tags"),
				d.When("app.Handle", d.Box(d.Plain, "global handler", "new goroutine · re-renders nothing")),
				d.When("ctx.Handle", d.Box(d.Lilac, "component handler", "UI goroutine"), d.Box(d.Lime, "re-render", "the component and its ancestors")),
			)),
	)
}

func persistingDiagram() app.UI {
	return d.Figure("SetState stores the value in memory and tells this tab's observers. Persist and Broadcast each send a copy further, once, when they are called.",
		d.Lane("this tab", "",
			d.Branch(d.Box(d.Sun, "ctx.SetState(key, v)", "value in memory"),
				d.When("always", d.Box(d.Lime, "observers", "OnChange")),
				d.When("Persist()", d.Box(d.Plain, "local storage", "JSON, read back after a reload")),
				d.When("Broadcast()", d.Box(d.Lilac, "BroadcastChannel", "to the other tabs")),
			)),
		d.Lane("other tabs", "",
			d.Flow(d.Box(d.Lilac, "message", "JSON"), d.Box(d.Lime, "WithBroadcast observers", "OnChange; stored value unchanged"))),
	)
}
