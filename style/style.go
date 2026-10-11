package style

import (
	"reflect"
	"slices"
)

type Style interface{ style() }

type Item interface {
	item()
	renderable
}

type renderable interface {
	blockTexts() []string
	Err() error
	subject() string
}

type static struct{ path string }

type generated struct {
	path  string
	items []Item
}

type dynamic struct{ items []Item }

func (static) style()    {}
func (generated) style() {}
func (dynamic) style()   {}

func Static(path string) Style { return static{path: path} }

func Generated(path string, items ...Item) Style {
	return generated{path: path, items: withoutNil(items)}
}

func Dynamic(items ...Item) Style { return dynamic{items: withoutNil(items)} }

func dynamicItems(styles []Style) []Item {
	var items []Item
	for _, s := range styles {
		if d, ok := s.(dynamic); ok {
			items = append(items, d.items...)
		}
	}
	return items
}

func withoutNil(items []Item) []Item {
	return slices.DeleteFunc(slices.Clone(items), func(i Item) bool {
		if i == nil {
			return true
		}
		v := reflect.ValueOf(i)
		return v.Kind() == reflect.Pointer && v.IsNil()
	})
}
