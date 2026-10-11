package nav

import (
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const ID = "side-nav"

type Entry struct {
	Title string
	Route string
}

func Render(entries []Entry, active string) app.UI {
	items := make([]app.UI, 0, len(entries))
	for _, entry := range entries {
		classes := []string{theme.NavLink.Name()}
		if entry.Route == active {
			classes = append(classes, theme.NavLinkActive.Name())
		}
		items = append(items, app.Li().Body(
			app.A().Class(classes...).Href(entry.Route).Text(entry.Title),
		))
	}
	return app.Nav().ID(ID).Class(theme.SideNav.Name()).Body(
		app.Span().Class(theme.NavLabel.Name()).Text("Guides"),
		app.Ul().Body(items...),
	)
}
