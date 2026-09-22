package base

import (
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/model"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Layout struct {
	entries []nav.Entry
	body    []app.UI
}

func NewLayout(entries []nav.Entry) *Layout {
	return &Layout{entries: entries}
}

func (l *Layout) Loading() app.UI {
	return app.Div().Class("shell shell-loading").Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Class("shell shell-error").Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	l.body = content
	return l
}

func (l *Layout) Draw(m *model.ViewModel[State]) app.HTMLDiv {
	state := m.State()
	return app.Div().Class("shell").Body(
		app.Header().Class("shell-header").Body(
			app.H1().Class("shell-title").Text(state.Title),
			nav.Render(l.entries, state.Active),
		),
		app.Main().Class("shell-main").Body(l.body...),
	)
}
