package applying

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func stylesheetsDiagram() app.UI {
	return d.Figure("Three kinds of stylesheet in one list. On the server, Build turns them into files and links; in the browser, Run inserts the dynamic rules.",
		d.Lane("native", "On the server",
			d.Branch(d.Box(d.Sun, "Handler.Styles", "[]style.Style"),
				d.When("Static", d.Box(d.Plain, "check the file", "a /web/ path, or a URL"), d.Box(d.Lime, "<link>", "")),
				d.When("Generated", d.Box(d.Lilac, "render the items", "write <WebDir>/….css"), d.Box(d.Lime, "<link>", "")),
				d.When("Dynamic", d.Box(d.Plain, "validate only", "no file, no link")),
			)),
		d.Lane("wasm", "In the browser",
			d.Flow(d.Box(d.Sun, "style.Run(styles...)", "before RunWhenOnBrowser"), d.Box(d.Lilac, "dynamic <style>", "rules follow the actions"))),
	)
}

func bindingDiagram() app.UI {
	return d.Figure("A bound rule follows its action: the binding reads the action's *T, and the engine inserts, removes or rewrites the rule.",
		d.Lane("wasm", "",
			d.Flow(d.Box(d.Coral, "action", "name · *T"), d.Via("a bound name"), d.Box(d.Lilac, "binding", "When · Compute · ComputeFrames"), d.Box(d.Sun, "engine", "on the UI goroutine"), d.Box(d.Lime, "dynamic <style>", "inserted · removed · rewritten"))),
	)
}
