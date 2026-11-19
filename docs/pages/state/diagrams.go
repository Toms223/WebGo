package state

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func shellDiagram() app.UI {
	return d.Figure("One Shell is the root for every route. It keeps one layout and one model, and draws the screen whose route matches the URL inside the layout. The screen loads on its own, after OnMount, with OnLoad.",
		d.Lane("root component", "app.Shell[U]",
			d.Flow(d.Box(d.Sun, "URL", "OnNav"), d.Via("matches a Route()"), d.Box(d.Lilac, "selected Screen", "*app.Screen[T]"), d.Box(d.Plain, "OnMount · OnLoad", "Loading · Error · Draw(model)"), d.Via("Layout.Body"), d.Box(d.Lime, "shell layout", "Draw(Model)"))),
	)
}

func modelDiagram() app.UI {
	return d.Figure("Read, change, write. Only Set and SetAndPersist reach the state store, so only they wake observers and fire the model's action. A Screen's OnLoad writes through the same Set.",
		d.Lane("", "",
			d.Flow(d.Box(d.Plain, "Get · Load", "store → *T"), d.Box(d.Sun, "Apply(verb)", "changes *T in memory"), d.Box(d.Lilac, "Set · SetAndPersist", "*T → store, + local storage"))),
		d.Lane("then", "",
			d.Branch(d.Box(d.Lilac, "state store", "under the model's key"),
				d.When("Observe", d.Box(d.Lime, "observing components", "re-render")),
				d.When("action", d.Box(d.Coral, "action named after the key", "style.When · style.Compute")),
			)),
	)
}
