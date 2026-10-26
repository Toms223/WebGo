// Package page builds app screens from a Layout and a data.ViewModel and hosts them under a Base
// component that switches screens by route.
package page

import (
	"errors"
	"strings"

	"github.com/Toms223/WebGo/data"

	"github.com/Toms223/WebGo/app"
)

// Errors reported by Page and Base; match them with errors.Is.
var (
	ErrNoLayout  = errors.New("page: no layout")
	ErrNoModel   = errors.New("page: no model")
	ErrNoScreens = errors.New("page: no screens")
	ErrNoRoute   = errors.New("page: no screen matches the current route")
)

// Layout draws a Page or Base: Loading before it mounts, Error when it has an error, and Draw
// otherwise. Base calls Body on every render, so Body must replace the content it stored, not append
// to it.
type Layout[T any] interface {
	Loading() app.UI
	Error(error) app.UI
	Draw(m *data.ViewModel[T]) app.HTMLDiv
	Body(...app.UI) Layout[T]
}

// Screen is one routed view of a Base. Base calls OnMount when the screen becomes selected and
// OnDismount when it stops being selected. *Page implements it, and a custom type can too.
type Screen interface {
	Content() app.UI
	Route() string
	OnMount(app.Context) Screen
	OnDismount() Screen
	Mounted() bool
}

// Page is a Screen that draws a ViewModel through a Layout. The first builder error is latched and
// every later builder call is ignored. Only *Page implements Screen.
type Page[T any] struct {
	Path    string
	model   *data.ViewModel[T]
	layout  Layout[T]
	err     error
	mounted bool
}

// NewPage returns a page at path with no layout or model; set both with WithLayout and WithModel
// before use.
func NewPage[T any](path string) *Page[T] {
	return &Page[T]{Path: path}
}

func (p *Page[T]) At(path string) *Page[T] {
	if p.err != nil {
		return p
	}
	p.Path = path
	return p
}

func (p *Page[T]) WithLayout(layout Layout[T]) *Page[T] {
	if p.err != nil {
		return p
	}
	if layout == nil {
		p.err = ErrNoLayout
		return p
	}
	p.layout = layout
	return p
}

func (p *Page[T]) WithModel(m *data.ViewModel[T]) *Page[T] {
	if p.err != nil {
		return p
	}
	if m == nil {
		p.err = ErrNoModel
		return p
	}
	p.model = m
	return p
}

// Err returns the error a builder latched, or else ErrNoLayout or ErrNoModel when the layout or model
// is missing.
func (p *Page[T]) Err() error {
	if p.err != nil {
		return p.err
	}
	if p.layout == nil {
		return ErrNoLayout
	}
	if p.model == nil {
		return ErrNoModel
	}
	return nil
}

func (p *Page[T]) OnMount(ctx app.Context) Screen {
	if err := p.Err(); err != nil {
		p.err = err
		return p
	}
	p.model.Get(ctx).Observe(ctx).While(p.Mounted)
	p.mounted = true
	return p
}

func (p *Page[T]) OnDismount() Screen {
	p.mounted = false
	return p
}

func (p *Page[T]) Mounted() bool {
	return p.mounted
}

// Content draws the page through its layout: Error on an error, Loading until mounted, then Draw. It
// panics with ErrNoLayout when there is no layout.
func (p *Page[T]) Content() app.UI {
	if p.layout == nil {
		panic(ErrNoLayout)
	}
	if err := p.Err(); err != nil {
		return p.layout.Error(err)
	}
	if !p.mounted {
		return p.layout.Loading()
	}
	return p.layout.Draw(p.model)
}

func (p *Page[T]) Route() string {
	return "/" + strings.TrimPrefix(p.Path, "/")
}
