package home

import (
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	Title = "Home"
	Route = base.HomeRoute
)

type State struct{}

type Screen struct {
	*page.Page[State]
}

func New() (*Screen, error) {
	viewModel, err := data.NewViewModel("home", &State{})
	if err != nil {
		return nil, err
	}
	built := page.NewPage[State](Route).
		WithLayout(&Layout{}).
		WithModel(viewModel)
	if err := built.Err(); err != nil {
		return nil, err
	}
	return &Screen{Page: built}, nil
}

func (s *Screen) OnMount(ctx app.Context) page.Screen {
	s.Page.OnMount(ctx)
	base.Activate(ctx, Route)
	return s
}
