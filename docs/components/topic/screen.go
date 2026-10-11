package topic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const fetchTimeout = 10 * time.Second

type Config struct {
	Key        string
	Path       string
	Title      string
	URL        string
	OnActivate func(app.Context, string)
	Demos      map[string]func() app.UI
}

type Screen struct {
	*page.Page[State]
	model      *data.ViewModel[State]
	url        string
	loader     func(string) (string, error)
	navigate   func(app.Context, *url.URL)
	onActivate func(app.Context, string)
}

func New(cfg Config) (*Screen, error) {
	viewModel, err := data.NewViewModel(cfg.Key, &State{})
	if err != nil {
		return nil, err
	}
	layout := NewLayout(cfg.Title, cfg.Demos)
	built := page.NewPage[State](cfg.Path).
		WithLayout(layout).
		WithModel(viewModel)
	if err := built.Err(); err != nil {
		return nil, err
	}
	screen := &Screen{
		Page:       built,
		model:      viewModel,
		url:        cfg.URL,
		loader:     fetch,
		navigate:   app.Context.NavigateTo,
		onActivate: cfg.OnActivate,
	}
	layout.activate = screen.Activate
	return screen, nil
}

func (s *Screen) Source() string {
	return s.url
}

func (s *Screen) OnMount(ctx app.Context) page.Screen {
	s.Page.OnMount(ctx)
	if s.onActivate != nil {
		s.onActivate(ctx, s.Route())
	}
	s.load(ctx)
	return s
}

func (s *Screen) load(ctx app.Context) {
	url := mountedURL(ctx.Page().URL().Path, s.Route(), s.url)
	loader := s.loader
	ctx.Async(func() {
		source, err := loader(url)
		ctx.Dispatch(func(ctx app.Context) {
			s.commit(ctx, source, err)
		})
	})
}

func (s *Screen) commit(ctx app.Context, source string, failure error) {
	fragment := ctx.Page().URL().Fragment
	s.model.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Loaded = true
		if failure != nil {
			state.Err = failure.Error()
			state.Source = ""
			state.Intro = ""
			state.Sections = nil
			state.Active = ""
			return state
		}
		state.Err = ""
		state.Source = source
		state.Intro, state.Sections = Split(source)
		state.Active = ""
		if len(state.Sections) > 0 {
			state.Active = state.Sections[0].Slug
		}
		if _, known := find(state.Sections, fragment); known {
			state.Active = fragment
		}
		return state
	}).Set()
}

func (s *Screen) Activate(ctx app.Context, slug string) {
	if _, known := find(s.model.State().Sections, slug); !known {
		return
	}
	next := *ctx.Page().URL()
	next.Fragment = slug
	s.navigate(ctx, &next)
	s.show(ctx, slug)
	ctx.ScrollTo(topicID)
}

func (s *Screen) OnNav(ctx app.Context) {
	slug := ctx.Page().URL().Fragment
	if _, known := find(s.model.State().Sections, slug); !known || slug == s.model.State().Active {
		return
	}
	s.show(ctx, slug)
}

func (s *Screen) show(ctx app.Context, slug string) {
	s.model.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Active = slug
		return state
	}).Set()
}

func find(sections []Section, slug string) (Section, bool) {
	for _, section := range sections {
		if section.Slug == slug {
			return section, true
		}
	}
	return Section{}, false
}

func mountedURL(browserPath, route, path string) string {
	base := strings.TrimSuffix(browserPath, route)
	base = strings.TrimSuffix(base, "/")
	if base == "" {
		return path
	}
	return base + path
}

func fetch(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("fetching %s: %s", url, response.Status)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
