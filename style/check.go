package style

import (
	"fmt"
	"slices"

	"github.com/Toms223/WebGo/app"
)

// Kind is app.StyleKind: whether a resolved style is static, generated or dynamic.
type Kind = app.StyleKind

const (
	StaticStyle    = app.StaticStyle
	GeneratedStyle = app.GeneratedStyle
	DynamicStyle   = app.DynamicStyle
)

// Source is app.StyleSource: a resolved style's kind, its path and, for a Generated style, the
// rendered bytes.
type Source = app.StyleSource

// Resolve checks a style's items and returns its Source, as app.Handler does for each style. A nil
// style resolves to a zero Source; on an error the Source is zero and the error wraps one of the Err
// sentinels and names the offending item.
func Resolve(s Style) (Source, error) {
	if s == nil {
		return Source{}, nil
	}
	return s.StyleSource()
}

func (s static) StyleSource() (Source, error) {
	return Source{Kind: StaticStyle, Path: s.path}, nil
}

func (s generated) StyleSource() (Source, error) {
	if err := checkGenerated(s.items); err != nil {
		return Source{}, err
	}
	return Source{Kind: GeneratedStyle, Path: s.path, Content: []byte(render(s.items))}, nil
}

func (s dynamic) StyleSource() (Source, error) {
	if err := checkDynamic(s.items); err != nil {
		return Source{}, err
	}
	return Source{Kind: DynamicStyle}, nil
}

type bindable interface{ dynamic() bool }

func checkGenerated(items []Item) error {
	var seen []string
	for _, i := range items {
		if l, ok := i.(*LayerRule); ok {
			if err := l.ownErr(); err != nil {
				return withSubject(err, l.subject())
			}
			if err := checkGenerated(l.items); err != nil {
				return err
			}
			continue
		}
		if err := placementErr(i); err != nil {
			return withSubject(err, i.subject())
		}
		if err := i.Err(); err != nil {
			return withSubject(err, i.subject())
		}
		if isBound(i) {
			return withSubject(ErrDynamicInGenerated, i.subject())
		}
		if key, ok := uniqueKey(i); ok {
			if slices.Contains(seen, key) {
				return withSubject(ErrDuplicateRule, key)
			}
			seen = append(seen, key)
		}
	}
	return nil
}

func uniqueKey(i Item) (string, bool) {
	switch i := i.(type) {
	case *Rule:
		return i.Selector(), true
	case *KeyframesRule:
		return i.subject(), true
	}
	return "", false
}

func checkDynamic(items []Item) error {
	for _, i := range items {
		if _, ok := i.(dynamicItem); !ok {
			return withSubject(ErrNotDynamic, i.subject())
		}
		if err := placementErr(i); err != nil {
			return withSubject(err, i.subject())
		}
		if err := i.Err(); err != nil {
			return withSubject(err, i.subject())
		}
		if !isBound(i) {
			return withSubject(ErrNotDynamic, i.subject())
		}
	}
	return nil
}

func placementErr(i Item) error {
	if r, ok := i.(*Rule); ok {
		return r.placementErr()
	}
	return nil
}

func isBound(i Item) bool {
	b, ok := i.(bindable)
	return ok && b.dynamic()
}

func withSubject(sentinel error, subject string) error {
	return fmt.Errorf("%w: %s", sentinel, subject)
}
