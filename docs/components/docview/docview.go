package docview

import (
	"github.com/Toms223/WebGo/markdown"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var Style = markdown.Style{
	Heading:    "doc-heading",
	Subheading: "doc-subheading",
}

func Render(source string) app.UI {
	body := markdown.RenderMarkdown(source, Style)
	if len(body) == 0 {
		return app.Div().Class("doc doc-empty")
	}
	return app.Div().Class("doc").Body(body...)
}
