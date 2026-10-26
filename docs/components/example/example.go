package example

import "github.com/Toms223/WebGo/app"

type Example struct {
	Name   string
	Source string
	Lang   string
	CSS    string
	Live   func() app.UI
}
