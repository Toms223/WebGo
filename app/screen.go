package app

import (
	"errors"
	"strings"
)

// Errors reported by Screen and its host; match them with errors.Is.
var (
	ErrNoLayout  = errors.New("app: no layout")
	ErrNoModel   = errors.New("app: no model")
	ErrNoScreens = errors.New("app: no screens")
	ErrNoRoute   = errors.New("app: no screen matches the current route")
)

// Layout draws a screen: Loading while it loads, Error on an error, and Draw otherwise.
// Draw must keep Body's content at the same position under the same elements on every render; vary what is around it, not where it is.
type Layout[T any] interface {
	Loading() UI
	Error(error) UI
	Draw(m *Model[T]) HTMLDiv
	Body(...UI) Layout[T]
}

// Routed is a component served at a route.
type Routed interface {
	Composer
	Route() string
}

// Screen is a routed component that draws a Model through a Layout; the first builder error is latched.
type Screen[T any] struct {
	Compo
	Path   string
	model  *Model[T]
	layout Layout[T]
	loader func(ctx Context, state *T) error
	err    error
}

// NewScreen returns a screen at path; set a layout and a model before use.
func NewScreen[T any](path string) *Screen[T] {
	return &Screen[T]{Path: path}
}

// At sets the screen's path.
func (s *Screen[T]) At(path string) *Screen[T] {
	if s.err != nil {
		return s
	}
	s.Path = path
	return s
}

// WithLayout sets the layout, latching ErrNoLayout when it is nil.
func (s *Screen[T]) WithLayout(layout Layout[T]) *Screen[T] {
	if s.err != nil {
		return s
	}
	if layout == nil {
		s.err = ErrNoLayout
		return s
	}
	s.layout = layout
	return s
}

// WithModel sets the model, latching ErrNoModel when it is nil.
func (s *Screen[T]) WithModel(m *Model[T]) *Screen[T] {
	if s.err != nil {
		return s
	}
	if m == nil {
		s.err = ErrNoModel
		return s
	}
	s.model = m
	return s
}

// WithLoader sets a loader that fills a copy of the state in OnLoad; the copy is published when it returns nil.
func (s *Screen[T]) WithLoader(loader func(ctx Context, state *T) error) *Screen[T] {
	if s.err != nil {
		return s
	}
	s.loader = loader
	return s
}

// Err returns the latched builder error, or else ErrNoLayout or ErrNoModel when one is missing.
func (s *Screen[T]) Err() error {
	if s.err != nil {
		return s.err
	}
	if s.layout == nil {
		return ErrNoLayout
	}
	if s.model == nil {
		return ErrNoModel
	}
	return nil
}

// Route returns the screen's path as a URL path.
func (s *Screen[T]) Route() string {
	return "/" + strings.TrimPrefix(s.Path, "/")
}

// CompoID returns the route, so a screen on another route remounts.
func (s *Screen[T]) CompoID() string {
	return s.Route()
}

// OnMount reads the model from the store and observes it while the screen is mounted.
func (s *Screen[T]) OnMount(ctx Context) {
	if err := s.Err(); err != nil {
		s.err = err
		return
	}
	s.model.Get(ctx).Observe(ctx).While(s.Mounted)
}

// OnLoad runs the loader on a copy of the state and publishes the copy when the loader succeeds and no newer load has started.
func (s *Screen[T]) OnLoad(ctx Context) error {
	if s.loader == nil || s.Err() != nil {
		return nil
	}
	n := s.load().loads
	loaded := *s.model.State()
	if err := s.loader(ctx, &loaded); err != nil {
		return err
	}
	ctx.Dispatch(func(ctx Context) {
		if s.load().loads != n {
			return
		}
		s.model.Load(ctx).Apply(func(_ Context, st *T) *T {
			*st = loaded
			return st
		}).Set()
	})
	return nil
}

// Loading returns the layout's loading view; it panics with ErrNoLayout when there is no layout.
func (s *Screen[T]) Loading() UI {
	return s.requireLayout().Loading()
}

// Error returns the layout's error view for err; it panics with ErrNoLayout when there is no layout.
func (s *Screen[T]) Error(err error) UI {
	return s.requireLayout().Error(err)
}

// Render draws the error when there is one and the model otherwise; it panics with ErrNoLayout when there is no layout.
func (s *Screen[T]) Render() UI {
	layout := s.requireLayout()
	if err := s.Err(); err != nil {
		return layout.Error(err)
	}
	return layout.Draw(s.model)
}

func (s *Screen[T]) requireLayout() Layout[T] {
	if s.layout == nil {
		panic(ErrNoLayout)
	}
	return s.layout
}
