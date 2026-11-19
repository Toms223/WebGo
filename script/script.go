// Package script loads ES modules through the Handler's importmap and hands page code a typed wrapper
// for each.
package script

import "github.com/Toms223/WebGo/app"

// Module is an entry of app.Handler.Imports. It is an alias of app.Module.
type Module = app.Module

// Script is a Module that wraps the loaded JavaScript module mod in a T. Init
// runs once per module name, on the first Use.
type Script[T any] interface {
	Module
	Init(ctx app.Context, mod app.Value) T
}
