package markdown

import (
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Markdown rendering"
	Route = "/markdown"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:        "markdown",
		Path:       Route,
		Title:      Title,
		URL:        "/data/markdown.md",
		OnActivate: base.Activate,
	})
}
