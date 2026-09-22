package main

import (
	"github.com/Toms223/WebGo/docs/cmd"
	"github.com/Toms223/WebGo/docs/pages"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func main() {
	pages.Register()
	app.RunWhenOnBrowser()
	cmd.Execute()
}
