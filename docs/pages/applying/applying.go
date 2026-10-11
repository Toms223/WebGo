package applying

import (
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/applying/demos"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Applying"
	Route = "/applying"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:        "applying",
		Path:       Route,
		Title:      Title,
		URL:        "/data/applying.md",
		OnActivate: base.Activate,
		Demos:      demos.All(),
	})
}
