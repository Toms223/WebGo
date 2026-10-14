package topic

import (
	"net/url"

	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
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
	sections   []section
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
		sections:   layout.sections,
		navigate:   app.Context.NavigateTo,
		onActivate: cfg.OnActivate,
	}
	layout.activate = screen.Activate
	return screen, nil
}

func (s *Screen) Sections() []string {
	slugs := make([]string, 0, len(s.sections))
	for _, section := range s.sections {
		slugs = append(slugs, section.Slug)
	}
	return slugs
}

func (s *Screen) OnMount(ctx app.Context) page.Screen {
	s.Page.OnMount(ctx)
	if s.onActivate != nil {
		s.onActivate(ctx, s.Route())
	}
	s.show(ctx, s.fragment(ctx, true))
	return s
}

func (s *Screen) OnNav(ctx app.Context) {
	slug := s.fragment(ctx, false)
	if slug == "" || slug == s.model.State().Active {
		return
	}
	s.show(ctx, slug)
}

func (s *Screen) Activate(ctx app.Context, slug string) {
	if !s.known(slug) {
		return
	}
	next := *ctx.Page().URL()
	next.Fragment = slug
	s.navigate(ctx, &next)
	s.show(ctx, slug)
	ctx.ScrollTo(sectionID)
}

func (s *Screen) fragment(ctx app.Context, fallback bool) string {
	slug := ctx.Page().URL().Fragment
	if s.known(slug) {
		return slug
	}
	if fallback && len(s.sections) > 0 {
		return s.sections[0].Slug
	}
	return ""
}

func (s *Screen) known(slug string) bool {
	for _, section := range s.sections {
		if section.Slug == slug {
			return true
		}
	}
	return false
}

func (s *Screen) show(ctx app.Context, slug string) {
	s.model.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		if state.Active != "" && state.Active != slug {
			state.Turn++
		}
		state.Active = slug
		return state
	}).Set()
}
