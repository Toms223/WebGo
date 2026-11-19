package app

import "strings"

// Shell is the root component: it draws the screen whose Route matches the URL, below the
// GOAPP_ROOT_PREFIX mount, through one layout and model.
type Shell[U any] struct {
	Compo
	Layout   Layout[U]
	Model    *Model[U]
	Screens  []Routed
	selected int
}

// NewShell returns a shell with no screen selected. A nil layout, model or screens slice gives
// ErrNoLayout, ErrNoModel or ErrNoScreens, checked in that order; an empty screens slice is accepted.
func NewShell[U any](layout Layout[U], model *Model[U], screens ...Routed) (*Shell[U], error) {
	if layout == nil {
		return nil, ErrNoLayout
	}
	if model == nil {
		return nil, ErrNoModel
	}
	if screens == nil {
		return nil, ErrNoScreens
	}
	return &Shell[U]{Layout: layout, Model: model, Screens: screens, selected: -1}, nil
}

// Err reports ErrNoLayout or ErrNoModel when Layout or Model is nil.
func (s *Shell[U]) Err() error {
	if s.Layout == nil {
		return ErrNoLayout
	}
	if s.Model == nil {
		return ErrNoModel
	}
	return nil
}

// OnMount reads the model's state from the store; it does nothing when Err reports an error.
func (s *Shell[U]) OnMount(ctx Context) {
	if s.Err() != nil {
		return
	}
	s.Model.Get(ctx)
}

// OnNav selects the screen whose route matches the URL path below the mount and puts the mount
// back in the URL when it is missing. With no match, Render shows ErrNoRoute.
func (s *Shell[U]) OnNav(ctx Context) {
	route, ok := s.selectPath(trimBasePath(normalizePath(ctx.Page().URL().Path), rootPrefix()))
	if ok {
		restoreURL(ctx, route)
	}
}

// Render draws the selected screen inside the layout, or the layout's error for a builder error or
// an unmatched route. It panics with ErrNoLayout when Layout is nil.
func (s *Shell[U]) Render() UI {
	if s.Layout == nil {
		panic(ErrNoLayout)
	}
	if err := s.Err(); err != nil {
		return s.Layout.Error(err)
	}
	if s.selected < 0 || s.selected >= len(s.Screens) {
		return s.Layout.Error(ErrNoRoute)
	}
	return s.Layout.Body(s.Screens[s.selected]).Draw(s.Model)
}

func (s *Shell[U]) setPath(path string) {
	s.selectPath(trimBasePath(normalizePath(path), rootPrefix()))
}

func (s *Shell[U]) selectPath(path string) (string, bool) {
	for index, screen := range s.Screens {
		route := normalizePath(screen.Route())
		if route == path {
			s.selected = index
			return route, true
		}
	}
	s.selected = -1
	return "", false
}

func restoreURL(ctx Context, route string) {
	wanted := browserPath(route)
	current := ctx.Page().URL()
	if normalizePath(current.Path) == wanted {
		return
	}
	next := *current
	next.Path = wanted
	next.RawPath = ""
	ctx.Page().ReplaceURL(&next)
}

func rootPrefix() string {
	prefix := strings.TrimRight(Getenv("GOAPP_ROOT_PREFIX"), "/")
	if prefix == "" {
		return ""
	}
	return normalizePath(prefix)
}

func browserPath(route string) string {
	prefix := rootPrefix()
	if prefix == "" {
		return route
	}
	if route == "/" {
		return prefix
	}
	return prefix + route
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
