package topic

import (
	"net/url"
	"time"

	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	revealTries = 10
	revealEvery = 50 * time.Millisecond
	siteName    = "WebGo"
)

type Section struct {
	Title string
	Body  func() []app.UI
}

type Config struct {
	Key        string
	Path       string
	Title      string
	Lead       string
	Intro      func() []app.UI
	Sections   []Section
	OnActivate func(app.Context, string)
}

type Screen struct {
	*page.Page[State]
	model      *data.ViewModel[State]
	layout     *Layout
	title      string
	navigate   func(app.Context, *url.URL)
	onActivate func(app.Context, string)
}

func New(cfg Config) (*Screen, error) {
	viewModel, err := data.NewViewModel(cfg.Key, &State{})
	if err != nil {
		return nil, err
	}
	layout := NewLayout(cfg)
	built := page.NewPage[State](cfg.Path).
		WithLayout(layout).
		WithModel(viewModel)
	if err := built.Err(); err != nil {
		return nil, err
	}
	screen := &Screen{
		Page:       built,
		model:      viewModel,
		layout:     layout,
		title:      cfg.Title,
		navigate:   app.Context.NavigateTo,
		onActivate: cfg.OnActivate,
	}
	layout.activate = screen.Activate
	return screen, nil
}

func (s *Screen) Sections() []string {
	slugs := make([]string, 0, len(s.layout.sections))
	for _, section := range s.layout.sections {
		slugs = append(slugs, section.Slug)
	}
	return slugs
}

func (s *Screen) OnMount(ctx app.Context) page.Screen {
	s.Page.OnMount(ctx)
	if s.onActivate != nil {
		s.onActivate(ctx, s.Route())
	}
	s.heading(ctx)
	if len(s.layout.sections) == 0 {
		return s
	}
	fragment := ctx.Page().URL().Fragment
	at, known := s.layout.find(fragment)
	s.show(ctx, s.layout.sections[at].Slug)
	if known {
		reveal(ctx, target(s.layout.sections[at], fragment), false)
	}
	return s
}

func (s *Screen) OnNav(ctx app.Context) {
	fragment := ctx.Page().URL().Fragment
	at, known := s.layout.find(fragment)
	if !known {
		return
	}
	section := s.layout.sections[at]
	if section.Slug != s.model.State().Active {
		s.show(ctx, section.Slug)
	}
	reveal(ctx, target(section, fragment), false)
}

func (s *Screen) Activate(ctx app.Context, fragment string) {
	at, known := s.layout.find(fragment)
	if !known {
		return
	}
	section := s.layout.sections[at]
	next := *ctx.Page().URL()
	next.Fragment = fragment
	s.navigate(ctx, &next)
	s.show(ctx, section.Slug)
	reveal(ctx, target(section, fragment), true)
}

func (s *Screen) show(ctx app.Context, slug string) {
	s.model.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		if state.Active != "" && state.Active != slug {
			state.Turn++
		}
		state.Active = slug
		return state
	}).Set()
	ctx.Page().SetTitle(s.layout.sections[s.layout.index(slug)].Title + " · " + s.title + " · " + siteName)
}

func (s *Screen) heading(ctx app.Context) {
	ctx.Page().SetTitle(s.title + " · " + siteName)
}

func target(s section, fragment string) string {
	if fragment == s.Slug {
		return headingID
	}
	return fragment
}

func reveal(ctx app.Context, id string, focus bool) {
	if !app.IsClient {
		return
	}
	ctx.Defer(func(ctx app.Context) { settle(ctx, id, focus, revealTries) })
}

func settle(ctx app.Context, id string, focus bool, tries int) {
	element := app.Window().GetElementByID(id)
	if !element.Truthy() {
		if tries > 0 {
			ctx.After(revealEvery, func(ctx app.Context) { settle(ctx, id, focus, tries-1) })
		}
		return
	}
	element.Call("scrollIntoView")
	if focus {
		element.Call("focus", map[string]any{"preventScroll": true})
	}
}
