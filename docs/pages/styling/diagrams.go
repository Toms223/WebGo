package styling

import (
	"github.com/Toms223/WebGo/app"
	d "github.com/Toms223/WebGo/docs/components/diagram"
)

func blocksDiagram() app.UI {
	return d.Figure("The blocks of one rule, in the order CSS() writes them, whatever order the calls were made in. Empty blocks are left out.",
		d.Lane("*style.Rule", "style.Class(\"card\")",
			d.Stack(
				d.Box(d.Sun, "base block", ".card { … } · Set"),
				d.Box(d.Lilac, "pseudo blocks", ".card:hover · in first-call order"),
				d.Box(d.Plain, "@container blocks", "in first-call order"),
				d.Box(d.Plain, "@supports blocks", "in first-call order"),
				d.Box(d.Lime, "@media blocks", "in first-call order"),
			)),
	)
}
