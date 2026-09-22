package base

import (
	"github.com/Toms223/WebGo/model"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const shellKey = "shell"

type State struct {
	Title  string
	Active string
}

var shell = mustShellModel(shellKey)

func newShellModel(key string) (*model.ViewModel[State], error) {
	return model.NewViewModel(key, &State{Title: "WebGo"})
}

func mustShellModel(key string) *model.ViewModel[State] {
	viewModel, err := newShellModel(key)
	if err != nil {
		panic(err)
	}
	return viewModel
}

func Activate(ctx app.Context, route string) {
	shell.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Active = route
		return state
	}).Set()
}
