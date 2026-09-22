package topic

import (
	"github.com/Toms223/WebGo/docs/components/docview"
	"github.com/Toms223/WebGo/model"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Layout struct {
	title string
	body  []app.UI
}

func NewLayout(title string) *Layout {
	return &Layout{title: title}
}

func (l *Layout) Loading() app.UI {
	return app.Div().Class("topic topic-loading").Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Class("topic topic-error").Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	l.body = content
	return l
}

func (l *Layout) Draw(m *model.ViewModel[State]) app.HTMLDiv {
	state := m.State()
	if state.Err != "" {
		return app.Div().Class("topic topic-error").Text(state.Err)
	}
	if !state.Loaded {
		return app.Div().Class("topic topic-loading").Text("Loading")
	}
	return app.Div().Class("topic").Body(
		app.H2().Class("topic-title").Text(l.title),
		docview.Render(state.Source),
	)
}
