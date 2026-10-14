package topic

import (
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	topicID   = "topic"
	sectionID = "section"
)

type Layout struct {
	title    string
	lead     string
	intro    func() []app.UI
	sections []section
	body     []app.UI
	activate func(app.Context, string)
}

func NewLayout(cfg Config) *Layout {
	return &Layout{title: cfg.Title, lead: cfg.Lead, intro: cfg.Intro, sections: slugged(cfg.Sections)}
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
	header := []app.UI{app.H1().Class(theme.TopicTitle.Name()).Text(l.title)}
	if l.lead != "" {
		header = append(header, app.P().Class(theme.TopicLead.Name()).Text(l.lead))
	}
	children := []app.UI{app.Header().Class(theme.TopicHeader.Name()).Body(header...)}
	if l.intro != nil {
		children = append(children, app.Div().Class(theme.TopicIntro.Name(), theme.Doc.Name()).Body(l.intro()...))
	}
	if len(l.sections) > 0 {
		at := l.index(state.Active)
		children = append(children, l.rail(at), l.section(at, state.Turn), l.pager(at))
	}
	return app.Div().Class(theme.Topic.Name()).ID(topicID).Body(children...)
}

func (l *Layout) index(slug string) int {
	for i, s := range l.sections {
		if s.Slug == slug {
			return i
		}
	}
	return 0
}

func (l *Layout) rail(at int) app.UI {
	links := make([]app.UI, 0, len(l.sections))
	for i, s := range l.sections {
		link := app.A().Class(theme.RailLink.Name()).Href("#" + s.Slug).Text(s.Title).OnClick(l.choose(s.Slug))
		if i == at {
			link = link.Class(theme.RailLinkActive.Name()).Aria("current", "true")
		}
		links = append(links, app.Li().Body(link))
	}
	return app.Nav().Class(theme.Rail.Name()).Aria("label", "On this page").Body(
		app.Span().Class(theme.RailLabel.Name()).Text("On this page"),
		app.Ul().Body(links...),
	)
}

func (l *Layout) section(at, turn int) app.UI {
	s := l.sections[at]
	classes := []string{theme.TopicSection.Name(), theme.Doc.Name()}
	if turn%2 == 1 {
		classes = append(classes, theme.TurnB.Name())
	}
	body := append([]app.UI{app.H2().Class(theme.SectionHeading.Name()).Text(s.Title)}, s.Body()...)
	return app.Section().Class(classes...).ID(sectionID).Body(body...)
}

func (l *Layout) pager(at int) app.UI {
	var links []app.UI
	if at > 0 {
		links = append(links, l.pagerLink(l.sections[at-1], "Previous", false))
	}
	if at < len(l.sections)-1 {
		links = append(links, l.pagerLink(l.sections[at+1], "Next", true))
	}
	return app.Nav().Class(theme.Pager.Name()).Aria("label", "Sections").Body(links...)
}

func (l *Layout) pagerLink(s section, label string, next bool) app.UI {
	classes := []string{theme.PagerLink.Name()}
	if next {
		classes = append(classes, theme.PagerNext.Name())
	}
	return app.A().Class(classes...).Href("#"+s.Slug).OnClick(l.choose(s.Slug)).Body(
		app.Span().Class(theme.PagerLabel.Name()).Text(label),
		app.Span().Text(s.Title),
	)
}

func (l *Layout) choose(slug string) app.EventHandler {
	return func(ctx app.Context, e app.Event) {
		e.PreventDefault()
		if l.activate != nil {
			l.activate(ctx, slug)
		}
	}
}
