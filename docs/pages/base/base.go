package base

import (
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/page"
)

func New(entries []nav.Entry, screens []page.Screen) (*page.Base[State], error) {
	return page.NewBase(NewLayout(entries), shell, screens)
}
