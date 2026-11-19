package base

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/nav"
)

func New(entries []nav.Entry, screens []app.Routed) (*app.Shell[State], error) {
	return app.NewShell(NewLayout(entries), shell, screens...)
}
