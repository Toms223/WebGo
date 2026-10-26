package routing

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func navigationDiagram() app.UI {
	return d.Figure("NavigateTo takes the first case that applies, from top to bottom. Only the last two touch the app and run OnNav.",
		d.Lane("browser", "NavigateTo(u)",
			d.Branch(d.Box(d.Sun, "u", "after Navigate parses it"),
				d.When("InternalURLs · mailto: · tel:", d.Box(d.Plain, "browser loads it", "leaves the app")),
				d.When("another host", d.Box(d.Plain, "new window", "")),
				d.When("same URL", d.Box(d.Plain, "nothing", "no history, no OnNav")),
				d.When("same path, new fragment", d.Box(d.Lime, "OnNav", "root untouched, then scroll")),
				d.When("anything else", d.Box(d.Lilac, "route or NotFound", "loaded as the root"), d.Box(d.Lime, "OnNav", "then scroll to the fragment")),
			)),
	)
}
