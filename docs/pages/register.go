package pages

import (
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/pages/applying"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/home"
	"github.com/Toms223/WebGo/docs/pages/markdown"
	"github.com/Toms223/WebGo/docs/pages/start"
	"github.com/Toms223/WebGo/docs/pages/state"
	"github.com/Toms223/WebGo/docs/pages/styling"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Entries() []nav.Entry {
	return []nav.Entry{
		{Title: start.Title, Route: start.Route},
		{Title: markdown.Title, Route: markdown.Route},
		{Title: state.Title, Route: state.Route},
		{Title: styling.Title, Route: styling.Route},
		{Title: applying.Title, Route: applying.Route},
	}
}

func Routes() []string {
	routes := []string{home.Route}
	for _, entry := range Entries() {
		routes = append(routes, entry.Route)
	}
	return routes
}

func Screens() ([]page.Screen, error) {
	homeScreen, err := home.New()
	if err != nil {
		return nil, err
	}
	startScreen, err := start.New()
	if err != nil {
		return nil, err
	}
	markdownScreen, err := markdown.New()
	if err != nil {
		return nil, err
	}
	stateScreen, err := state.New()
	if err != nil {
		return nil, err
	}
	stylingScreen, err := styling.New()
	if err != nil {
		return nil, err
	}
	applyingScreen, err := applying.New()
	if err != nil {
		return nil, err
	}
	return []page.Screen{homeScreen, startScreen, markdownScreen, stateScreen, stylingScreen, applyingScreen}, nil
}

func Register() {
	for _, route := range Routes() {
		app.Route(route, func() app.Composer {
			screens, err := Screens()
			if err != nil {
				panic(err)
			}
			shell, err := base.New(Entries(), screens)
			if err != nil {
				panic(err)
			}
			return shell
		})
	}
}
