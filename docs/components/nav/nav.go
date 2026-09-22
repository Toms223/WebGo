package nav

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type Entry struct {
	Title string
	Route string
}

func Render(entries []Entry, active string) app.UI {
	items := make([]app.UI, 0, len(entries))
	for _, entry := range entries {
		class := "nav-link"
		if entry.Route == active {
			class = "nav-link nav-link-active"
		}
		items = append(items, app.Li().Body(
			app.A().Class(class).Href(entry.Route).Text(entry.Title),
		))
	}
	return app.Nav().Class("nav").Body(app.Ul().Body(items...))
}
