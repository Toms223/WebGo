package example

import "github.com/maxence-charriere/go-app/v11/pkg/app"

type Example struct {
	Name   string
	Source string
	Lang   string
	CSS    string
	Live   func() app.UI
}
