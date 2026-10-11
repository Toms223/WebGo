package base

import (
	"github.com/Toms223/WebGo/data"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const shellKey = "shell"

type State struct {
	Title  string
	Active string
	Dark   bool
	Chosen bool
	Menu   bool
}

func (s *State) IsDark() bool {
	if s.Chosen {
		return s.Dark
	}
	return app.Window().Call("matchMedia", "(prefers-color-scheme: dark)").Get("matches").Bool()
}

var shell = mustShellModel(shellKey)

func newShellModel(key string) (*data.ViewModel[State], error) {
	return data.NewViewModel(key, &State{Title: "WebGo"})
}

func mustShellModel(key string) *data.ViewModel[State] {
	viewModel, err := newShellModel(key)
	if err != nil {
		panic(err)
	}
	return viewModel
}

func Activate(ctx app.Context, route string) {
	shell.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Active = route
		state.Menu = false
		return state
	}).Set()
}
