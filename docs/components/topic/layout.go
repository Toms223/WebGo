package topic

import (
	"strings"

	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/components/docview"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const topicID = "topic"

type Layout struct {
	title    string
	demos    map[string]func() app.UI
	body     []app.UI
	activate func(app.Context, string)
}

func NewLayout(title string, demos map[string]func() app.UI) *Layout {
	return &Layout{title: title, demos: demos}
}

func (l *Layout) Loading() app.UI {
	return app.Div().Class(theme.Topic.Name(), theme.TopicLoading.Name()).Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Class(theme.Topic.Name(), theme.TopicError.Name()).Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	l.body = content
	return l
}

func (l *Layout) Draw(m *data.ViewModel[State]) app.HTMLDiv {
	state := m.State()
	if state.Err != "" {
		return app.Div().Class(theme.Topic.Name(), theme.TopicError.Name()).Text(state.Err)
	}
	if !state.Loaded {
		return app.Div().Class(theme.Topic.Name(), theme.TopicLoading.Name()).Text("Loading")
	}
	children := []app.UI{app.H2().Class(theme.TopicTitle.Name()).Text(l.title)}
	if len(state.Sections) == 0 {
		children = append(children, docview.Render(state.Source, l.demos))
		return app.Div().Class(theme.Topic.Name()).ID(topicID).Body(children...)
	}
	if strings.TrimSpace(state.Intro) != "" {
		children = append(children, docview.Render(state.Intro, l.demos))
	}
	children = append(children,
		l.sections(state),
		docview.Render(activeSection(state).Source, l.demos),
	)
	return app.Div().Class(theme.Topic.Name()).ID(topicID).Body(children...)
}

func (l *Layout) sections(state *State) app.UI {
	links := make([]app.UI, 0, len(state.Sections))
	for _, section := range state.Sections {
		classes := []string{theme.SectionLink.Name()}
		if section.Slug == state.Active {
			classes = append(classes, theme.SectionLinkActive.Name())
		}
		links = append(links, app.A().
			Class(classes...).
			Href("#"+section.Slug).
			Text(section.Title).
			OnClick(l.choose(section.Slug)))
	}
	return app.Nav().Class(theme.Sections.Name()).Body(links...)
}

func (l *Layout) choose(slug string) app.EventHandler {
	return func(ctx app.Context, e app.Event) {
		e.PreventDefault()
		if l.activate != nil {
			l.activate(ctx, slug)
		}
	}
}

func activeSection(state *State) Section {
	if section, known := find(state.Sections, state.Active); known {
		return section
	}
	return state.Sections[0]
}
