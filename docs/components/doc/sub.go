package doc

import (
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Sub struct {
	app.Compo
	Title string
	ID    string
}

func (s *Sub) Render() app.UI {
	heading := app.H3().Class(theme.DocSubheading.Name()).TabIndex(-1).Text(s.Title)
	if s.ID != "" {
		heading = heading.ID(s.ID)
	}
	return heading
}
