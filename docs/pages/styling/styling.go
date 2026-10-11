package styling

import (
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/styling/demos"
)

const (
	Title = "Styling"
	Route = "/styling"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:        "styling",
		Path:       Route,
		Title:      Title,
		URL:        "/data/styling.md",
		OnActivate: base.Activate,
		Demos:      demos.All(),
	})
}
