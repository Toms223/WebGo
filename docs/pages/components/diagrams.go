package components

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func architectureDiagram() app.UI {
	return d.Figure("One main, built twice: the server answers with prerendered HTML and the runtime files, then app.wasm takes over in the browser.",
		d.Lane("native", "On the server",
			d.Flow(d.Box(d.Sun, "main", "app.Route · app.Handler"), d.Box(d.Lilac, "app.Handler", "prerenders the route"), d.Box(d.Plain, "HTML page", "markup · app.js · manifest"))),
		d.Lane("wasm", "In the browser",
			d.Flow(d.Box(d.Plain, "prerendered markup", "shown at once"), d.Via("app.wasm loads"), d.Box(d.Lime, "RunWhenOnBrowser", "mounts the routed component"), d.Box(d.Coral, "live tree", "a single-page app from here"))),
	)
}

func lifecycleDiagram() app.UI {
	return d.Figure("Hooks in the order they run. OnPreRender runs only on the server; OnMount and the hooks after it only in the browser.",
		d.Lane("server", "Prerendering a page",
			d.Flow(d.Box(d.Plain, "OnInit", "defaults, no context"), d.Box(d.Sun, "OnPreRender", "title and description"), d.Box(d.Plain, "OnNav", "reads the URL"), d.Box(d.Lilac, "HTML", "sent to the browser"))),
		d.Lane("browser", "A mounted component",
			d.Flow(d.Box(d.Plain, "OnInit", ""), d.Box(d.Lime, "OnMount", "in the DOM"), d.Box(d.Lime, "OnNav", "every navigation"), d.Box(d.Lilac, "OnUpdate · OnResize", "OnAppUpdate · OnAppInstallChange"), d.Box(d.Coral, "OnDismount", "stop what OnMount started"))),
	)
}

func updatingDiagram() app.UI {
	return d.Figure("Anything that touches a component flags it and its ancestors; the next frame renders each flagged component once.",
		d.Lane("", "",
			d.Flow(d.Box(d.Sun, "trigger", "event · hook · Dispatch · After · Handle · ObserveState · Update"), d.Via("flags"), d.Box(d.Plain, "component + ancestors", "marked for update"), d.Via("next frame"), d.Box(d.Lilac, "Render", "parents before children"), d.Box(d.Lime, "DOM", "only what changed"))),
	)
}
