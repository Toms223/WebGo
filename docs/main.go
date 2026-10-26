package main

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/cmd"
	"github.com/Toms223/WebGo/docs/pages"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/style"
)

func main() {
	pages.Register()
	style.Run(theme.Styles()...)
	app.RunWhenOnBrowser()
	cmd.Execute()
}
