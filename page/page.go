package page

import (
	"errors"
	"github.com/Toms223/WebGo/model"
	"strings"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var (
	ErrNoLayout  = errors.New("page: no layout")
	ErrNoModel   = errors.New("page: no model")
	ErrNoScreens = errors.New("page: no screens")
	ErrNoRoute   = errors.New("page: no screen matches the current route")
)

type Layout[T any] interface {
	Loading() app.UI
	Error(error) app.UI
	Draw(m *model.ViewModel[T]) app.HTMLDiv
	Body(...app.UI) Layout[T]
}

type Screen interface {
	Content() app.UI
	Route() string
	OnMount(app.Context) Screen
	OnDismount() Screen
	Mounted() bool
}

type Page[T any] struct {
	Path    string
	model   *model.ViewModel[T]
	layout  Layout[T]
	err     error
	mounted bool
}

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

func (p *Page[T]) WithModel(m *model.ViewModel[T]) *Page[T] {
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
