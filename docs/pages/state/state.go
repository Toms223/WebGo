package state

import (
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Pages, layouts and state"
	Route = "/state"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:        "state",
		Path:       Route,
		Title:      Title,
		URL:        "/data/pages-state.md",
		OnActivate: base.Activate,
	})
}
