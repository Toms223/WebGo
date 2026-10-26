package nav

import (
	"strings"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/theme"
)

const ID = "side-nav"

type Entry struct {
	Title string
	Route string
	Group string
}

func Href(route string) string {
	return strings.TrimSuffix(app.Getenv("GOAPP_ROOT_PREFIX"), "/") + route
}

func Render(entries []Entry, active string) app.UI {
	var groups []app.UI
	var items []app.UI
	group := ""
	flush := func() {
		if items != nil {
			groups = append(groups, app.Span().Style(theme.NavLabel).Text(group), app.Ul().Body(items...))
		}
		items = nil
	}
	for _, entry := range entries {
		name := entry.Group
		if name == "" {
			name = "Guides"
		}
		if name != group {
			flush()
			group = name
		}
		classes := []app.ClassRule{theme.NavLink}
		if entry.Route == active {
			classes = append(classes, theme.NavLinkActive)
		}
		items = append(items, app.Li().Body(
			app.A().Style(classes...).Href(Href(entry.Route)).Text(entry.Title),
		))
	}
	flush()
	return app.Nav().ID(ID).Style(theme.SideNav).Body(groups...)
}
