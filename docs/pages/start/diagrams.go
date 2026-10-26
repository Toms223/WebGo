package start

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func staticDiagram() app.UI {
	return d.Figure("Two builds of the same main: one for the browser, one that writes the site next to web/. Any static host can serve the result.",
		d.Lane("", "",
			d.Flow(d.Box(d.Sun, "go build", "GOOS=js GOARCH=wasm → web/app.wasm"), d.Box(d.Lilac, "go run .", "GenerateStaticWebsite(\".\", h)"), d.Box(d.Lime, "static files", "index.html · app.js · app-worker.js · web/"), d.Box(d.Plain, "static host", "no Go server"))),
	)
}
