package deploy

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func updatesDiagram() app.UI {
	return d.Figure("The service worker across visits. After a deploy, an open page reloads itself once onto the new version, unless DisableAutoUpdate is set.",
		d.Lane("first visit", "No worker yet",
			d.Flow(d.Box(d.Plain, "page", "from the network"), d.Box(d.Sun, "app-worker.js installs", "stores every listed file"), d.Box(d.Lime, "takes control", "no reload"))),
		d.Lane("next visits", "Same version",
			d.Flow(d.Box(d.Plain, "page", ""), d.Box(d.Lime, "worker answers", "from this version's cache"), d.Box(d.Plain, "works offline", ""))),
		d.Lane("after a deploy", "New version",
			d.Flow(d.Box(d.Plain, "update check", "load · online · visible · every 60\u00a0s"), d.Box(d.Sun, "new worker installs", "bypasses the HTTP cache"), d.Via("all files stored"), d.Box(d.Lilac, "activates", "deletes old app- caches"), d.Box(d.Coral, "page reloads once", "runs the new app.wasm"))),
	)
}

func staticDiagram() app.UI {
	return d.Figure("From source to a static host: build app.wasm, generate the pages, copy web/, publish.",
		d.Lane("build", "",
			d.Flow(d.Box(d.Sun, "go build", "GOOS=js GOARCH=wasm → web/app.wasm"), d.Box(d.Lilac, "GenerateStaticWebsite", "pages · app.js · worker · manifest"), d.Box(d.Plain, "copy web/", "into dir/web"), d.Box(d.Lime, "static host", "GitHub Pages under /repo/"))),
	)
}
