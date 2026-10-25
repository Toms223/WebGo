package doc

import (
	"time"

	"github.com/Toms223/WebGo/docs/theme"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const copiedFor = 1500 * time.Millisecond

type Copy struct {
	app.Compo
	Text  string
	Float bool
	done  bool
}

func (c *Copy) Render() app.UI {
	label := "Copy"
	classes := []string{theme.CopyButton.Name()}
	if c.Float {
		classes = append(classes, theme.CopyFloat.Name())
	}
	if c.done {
		label = "Copied"
	}
	return app.Button().Type("button").Class(classes...).Aria("live", "polite").Text(label).OnClick(c.copy)
}

func (c *Copy) copy(ctx app.Context, e app.Event) {
	clipboard := app.Window().Get("navigator").Get("clipboard")
	if !clipboard.Truthy() {
		return
	}
	clipboard.Call("writeText", c.Text)
	c.done = true
	ctx.After(copiedFor, func(ctx app.Context) {
		c.done = false
		ctx.Update()
	})
}
