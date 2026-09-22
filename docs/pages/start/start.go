package start

import (
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Getting Started"
	Route = "/"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:        "start",
		Path:       "",
		Title:      Title,
		URL:        "/data/start.md",
		OnActivate: base.Activate,
	})
}
