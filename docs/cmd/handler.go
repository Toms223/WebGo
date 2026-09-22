package cmd

import "github.com/maxence-charriere/go-app/v11/pkg/app"

func Handler() *app.Handler {
	return &app.Handler{
		Name:        "WebGo",
		ShortName:   "WebGo",
		Title:       "WebGo Documentation",
		Description: "Documentation for the WebGo framework",
		Lang:        "en",
		Icon:        app.Icon{Default: "/web/icon.png", Large: "/web/icon.png", SVG: "/web/icon.svg"},
		Styles:      []string{"/web/app.css"},
	}
}
