package topic

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Toms223/WebGo/model"
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
}

type Screen struct {
	*page.Page[State]
	model      *model.ViewModel[State]
	url        string
	loader     func(string) (string, error)
	onActivate func(app.Context, string)
}

func New(cfg Config) (*Screen, error) {
	viewModel, err := model.NewViewModel(cfg.Key, &State{})
	if err != nil {
		return nil, err
	}
	built := page.NewPage[State](cfg.Path).
		WithLayout(NewLayout(cfg.Title)).
		WithModel(viewModel)
	if err := built.Err(); err != nil {
		return nil, err
	}
	return &Screen{
		Page:       built,
		model:      viewModel,
		url:        cfg.URL,
		loader:     fetch,
		onActivate: cfg.OnActivate,
	}, nil
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
	url := s.url
	loader := s.loader
	ctx.Async(func() {
		source, err := loader(url)
		ctx.Dispatch(func(ctx app.Context) {
			s.commit(ctx, source, err)
		})
	})
}

func (s *Screen) commit(ctx app.Context, source string, failure error) {
	s.model.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Loaded = true
		if failure != nil {
			state.Err = failure.Error()
			state.Source = ""
			return state
		}
		state.Err = ""
		state.Source = source
		return state
	}).Set()
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
