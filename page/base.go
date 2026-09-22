package page

import (
	"WebGo/model"
	"net/url"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Base[U any] struct {
	app.Compo
	BaseLayout Layout[U]
	Model      *model.ViewModel[U]
	Screens    []Screen
	selected   int
	mounted    bool
}

func NewBase[U any](layout Layout[U], model *model.ViewModel[U], screens []Screen) (*Base[U], error) {
	if layout == nil {
		return nil, ErrNoLayout
	}
	if model == nil {
		return nil, ErrNoModel
	}
	if screens == nil {
		return nil, ErrNoScreens
	}
	return &Base[U]{
		BaseLayout: layout,
		Model:      model,
		Screens:    screens,
	}, nil
}

func (b *Base[U]) Err() error {
	if b.BaseLayout == nil {
		return ErrNoLayout
	}
	if b.Model == nil {
		return ErrNoModel
	}
	return nil
}

func (b *Base[U]) OnNav(ctx app.Context) {
	path := ctx.Page().URL().Path
	for index, screen := range b.Screens {
		route := screen.Route()
		if route != path {
			continue
		}
		if index == b.selected {
			return
		}
		if b.selected >= 0 && b.selected < len(b.Screens) {
			b.Screens[b.selected] = b.Screens[b.selected].OnDismount()
		}
		b.selected = index
		b.Screens[index] = screen.OnMount(ctx)
		ctx.Page().ReplaceURL(mustParseUrl(route))
		return
	}
	b.selected = -1
}

func (b *Base[U]) OnMount(ctx app.Context) {
	if b.Err() != nil {
		return
	}
	b.Model.Get(ctx)
	if len(b.Screens) == 0 {
		b.mounted = true
		return
	}
	b.Screens[b.selected] = b.Screens[b.selected].OnMount(ctx)
	b.mounted = true
}

func (b *Base[U]) Render() app.UI {
	if b.BaseLayout == nil {
		panic(ErrNoLayout)
	}
	if err := b.Err(); err != nil {
		return b.BaseLayout.Error(err)
	}
	if !b.mounted || len(b.Screens) == 0 {
		return b.BaseLayout.Loading()
	}
	if b.selected >= len(b.Screens) {
		b.selected = 0
	}
	if b.selected < 0 {
		return b.BaseLayout.Error(ErrNoRoute)
	}
	return b.BaseLayout.Body(
		b.Screens[b.selected].Content(),
	).Draw(b.Model)
}

func mustParseUrl(path string) *url.URL {
	newUrl, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return newUrl
}
