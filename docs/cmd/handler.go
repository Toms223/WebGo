package cmd

import (
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/serve"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Handler(repository, webDir string) (*app.Handler, error) {
	handler := &serve.Handler{
		Handler: app.Handler{
			Name:        "WebGo",
			ShortName:   "WebGo",
			Title:       "WebGo Documentation",
			Description: "Documentation for the WebGo framework",
			Lang:        "en",
			Icon:        app.Icon{Default: "/web/icon.png", Large: "/web/icon.png", SVG: "/web/icon.svg"},
		},
		Styles: theme.Styles(),
		WebDir: webDir,
	}
	if repository != "" {
		handler.Resources = app.GitHubPages(repository)
	}
	return handler.Build()
}
