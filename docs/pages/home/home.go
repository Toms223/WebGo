package home

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Home"
	Route = base.HomeRoute
)

type State struct {
	Count int
}

type Screen struct {
	*app.Screen[State]
}

func New() (*Screen, error) {
	viewModel, err := app.NewModel("home", &State{})
	if err != nil {
		return nil, err
	}
	built := app.NewScreen[State](Route).
		WithLayout(&Layout{}).
		WithModel(viewModel)
	if err := built.Err(); err != nil {
		return nil, err
	}
	return &Screen{Screen: built}, nil
}

func (s *Screen) OnMount(ctx app.Context) {
	s.Screen.OnMount(ctx)
	ctx.Page().SetTitle("WebGo · web apps in Go with typed CSS, pages and state")
	base.Activate(ctx, Route)
}
