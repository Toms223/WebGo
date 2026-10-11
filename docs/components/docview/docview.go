package docview

import (
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/markdown"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var Style = markdown.Style{
	Heading:      theme.DocHeading.Name(),
	Subheading:   theme.DocSubheading.Name(),
	Example:      theme.Example.Name(),
	ExampleCode:  theme.ExampleCode.Name(),
	ExampleLive:  theme.ExampleLive.Name(),
	ExampleLabel: theme.ExampleLabel.Name(),
}

func Render(source string, demos map[string]func() app.UI) app.UI {
	style := Style
	style.Demos = demos
	body := markdown.RenderMarkdown(source, style)
	if len(body) == 0 {
		return app.Div().Class(theme.Doc.Name(), theme.DocEmpty.Name())
	}
	return app.Div().Class(theme.Doc.Name()).Body(body...)
}
