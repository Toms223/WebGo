package state

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func baseDiagram() app.UI {
	return d.Figure("One Base is the root for every route. It keeps one layout and one model, and draws the screen whose route matches the URL inside the layout.",
		d.Lane("root component", "page.Base[U]",
			d.Flow(d.Box(d.Sun, "URL", "OnNav"), d.Via("matches a Route()"), d.Box(d.Lilac, "selected Screen", "*page.Page[T]"), d.Box(d.Plain, "Content()", "Loading · Error · Draw(model)"), d.Via("BaseLayout.Body"), d.Box(d.Lime, "base layout", "Draw(Model): the shell"))),
	)
}

func modelDiagram() app.UI {
	return d.Figure("Read, change, write. Only Set and SetAndPersist reach the state store, so only they wake Observe and OnAction.",
		d.Lane("", "",
			d.Flow(d.Box(d.Plain, "Get · Load", "store → *T"), d.Box(d.Sun, "Apply(verb)", "changes *T in memory"), d.Box(d.Lilac, "Set · SetAndPersist", "*T → store, + local storage"))),
		d.Lane("then", "",
			d.Branch(d.Box(d.Lilac, "state store", "under the model's key"),
				d.When("Observe", d.Box(d.Lime, "observing components", "re-render")),
				d.When("OnAction", d.Box(d.Coral, "action named after the key", "style.When · style.Compute")),
			)),
	)
}
