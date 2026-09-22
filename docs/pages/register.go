package pages

import (
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/markdown"
	"github.com/Toms223/WebGo/docs/pages/start"
	"github.com/Toms223/WebGo/docs/pages/state"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Entries() []nav.Entry {
	return []nav.Entry{
		{Title: start.Title, Route: start.Route},
		{Title: markdown.Title, Route: markdown.Route},
		{Title: state.Title, Route: state.Route},
	}
}

func Screens() ([]page.Screen, error) {
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
	return []page.Screen{startScreen, markdownScreen, stateScreen}, nil
}

func Register() {
	for _, entry := range Entries() {
		app.Route(entry.Route, func() app.Composer {
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
