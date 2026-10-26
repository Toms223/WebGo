package doc

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/theme"
)

type Sub struct {
	app.Compo
	Title string
	ID    string
}

func (s *Sub) Render() app.UI {
	heading := app.H3().Style(theme.DocSubheading).TabIndex(-1).Text(s.Title)
	if s.ID != "" {
		heading = heading.ID(s.ID)
	}
	return heading
}
