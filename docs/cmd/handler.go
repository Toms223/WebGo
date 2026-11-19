package cmd

import (
	"github.com/Toms223/WebGo/app"
	demos "github.com/Toms223/WebGo/docs/pages/scripts/demos"
	"github.com/Toms223/WebGo/docs/theme"
)

func Handler(repository, webDir string) (*app.Handler, error) {
	handler := &app.Handler{
		Name:        "WebGo",
		ShortName:   "WebGo",
		Title:       "WebGo Documentation",
		Description: "Documentation for the WebGo framework",
		Lang:        "en",
		Icon:        app.Icon{Default: "/web/icon.png", Large: "/web/icon.png", SVG: "/web/icon.svg"},
		Imports:     []app.Module{demos.Confetti{}},
		Styles:      theme.Styles(),
		WebDir:      webDir,
	}
	if repository != "" {
		handler.Resources = app.GitHubPages(repository)
	}
	if err := handler.Build(); err != nil {
		return nil, err
	}
	return handler, nil
}
