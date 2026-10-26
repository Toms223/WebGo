package pages

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/pages/actions"
	"github.com/Toms223/WebGo/docs/pages/applying"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/browser"
	"github.com/Toms223/WebGo/docs/pages/components"
	"github.com/Toms223/WebGo/docs/pages/deploy"
	"github.com/Toms223/WebGo/docs/pages/home"
	"github.com/Toms223/WebGo/docs/pages/markdown"
	"github.com/Toms223/WebGo/docs/pages/migrating"
	"github.com/Toms223/WebGo/docs/pages/routing"
	"github.com/Toms223/WebGo/docs/pages/start"
	"github.com/Toms223/WebGo/docs/pages/state"
	"github.com/Toms223/WebGo/docs/pages/styling"
	"github.com/Toms223/WebGo/docs/pages/tests"
	"github.com/Toms223/WebGo/page"
)

func Entries() []nav.Entry {
	return []nav.Entry{
		{Title: start.Title, Route: start.Route},
		{Title: markdown.Title, Route: markdown.Route},
		{Title: state.Title, Route: state.Route},
		{Title: styling.Title, Route: styling.Route},
		{Title: applying.Title, Route: applying.Route},
		{Title: components.Title, Route: components.Route, Group: "Core"},
		{Title: routing.Title, Route: routing.Route, Group: "Core"},
		{Title: actions.Title, Route: actions.Route, Group: "Core"},
		{Title: browser.Title, Route: browser.Route, Group: "Core"},
		{Title: deploy.Title, Route: deploy.Route, Group: "Core"},
		{Title: tests.Title, Route: tests.Route, Group: "Core"},
		{Title: migrating.Title, Route: migrating.Route, Group: "Core"},
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
	componentsScreen, err := components.New()
	if err != nil {
		return nil, err
	}
	routingScreen, err := routing.New()
	if err != nil {
		return nil, err
	}
	actionsScreen, err := actions.New()
	if err != nil {
		return nil, err
	}
	browserScreen, err := browser.New()
	if err != nil {
		return nil, err
	}
	deployScreen, err := deploy.New()
	if err != nil {
		return nil, err
	}
	testsScreen, err := tests.New()
	if err != nil {
		return nil, err
	}
	migratingScreen, err := migrating.New()
	if err != nil {
		return nil, err
	}
	return []page.Screen{homeScreen, startScreen, markdownScreen, stateScreen, stylingScreen, applyingScreen, componentsScreen, routingScreen, actionsScreen, browserScreen, deployScreen, testsScreen, migratingScreen}, nil
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
