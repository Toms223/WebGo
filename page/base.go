package page

import (
	"net/url"
	"strings"

	"github.com/Toms223/WebGo/model"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Base[U any] struct {
	app.Compo
	BaseLayout Layout[U]
	Model      *model.ViewModel[U]
	Screens    []Screen
	selected   int
	mounted    bool
	basePath   string
	baseKnown  bool
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

func (b *Base[U]) BasePath() string {
	return b.basePath
}

func (b *Base[U]) OnNav(ctx app.Context) {
	path := b.screenPath(ctx)
	for index, screen := range b.Screens {
		route := normalizePath(screen.Route())
		if route != path {
			continue
		}
		b.restoreURL(ctx, route)
		if index == b.selected {
			return
		}
		if b.selected >= 0 && b.selected < len(b.Screens) {
			b.Screens[b.selected] = b.Screens[b.selected].OnDismount()
		}
		b.selected = index
		b.Screens[index] = screen.OnMount(ctx)
		return
	}
	b.selected = -1
}

func (b *Base[U]) restoreURL(ctx app.Context, route string) {
	wanted := b.browserPath(route)
	if normalizePath(ctx.Page().URL().Path) == wanted {
		return
	}
	ctx.Page().ReplaceURL(mustParseUrl(wanted))
}

func (b *Base[U]) browserPath(route string) string {
	if b.basePath == "" {
		return route
	}
	if route == "/" {
		return b.basePath
	}
	return b.basePath + route
}

func (b *Base[U]) OnMount(ctx app.Context) {
	if b.Err() != nil {
		return
	}
	b.Model.Get(ctx)
	if !b.baseKnown {
		b.basePath = deduceBasePath(currentPath(ctx), b.Screens)
		b.baseKnown = true
	}
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

func (b *Base[U]) screenPath(ctx app.Context) string {
	return trimBasePath(currentPath(ctx), b.basePath)
}

func currentPath(ctx app.Context) string {
	return normalizePath(ctx.Page().URL().Path)
}

func normalizePath(path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	if path == "" {
		return "/"
	}
	return path
}

func deduceBasePath(path string, screens []Screen) string {
	base := ""
	longest := -1
	for _, screen := range screens {
		route := normalizePath(screen.Route())
		if route == "/" {
			continue
		}
		if !strings.HasSuffix(path, route) {
			continue
		}
		if len(route) > longest {
			base = strings.TrimSuffix(path, route)
			longest = len(route)
		}
	}
	if longest < 0 {
		return normalizeBasePath(path)
	}
	return normalizeBasePath(base)
}

func normalizeBasePath(base string) string {
	base = strings.TrimRight(base, "/")
	if base == "/" {
		return ""
	}
	return base
}

func trimBasePath(path, base string) string {
	if base == "" {
		return path
	}
	if path == base {
		return "/"
	}
	if strings.HasPrefix(path, base+"/") {
		return normalizePath(strings.TrimPrefix(path, base))
	}
	return path
}

func mustParseUrl(path string) *url.URL {
	newUrl, err := url.Parse(path)
	if err != nil {
		panic(err)
	}
	return newUrl
}
