// Package style describes CSS rules and at-rules as Go values and delivers them as static, generated
// or dynamic stylesheets. Rules are built from the typed values in the css, kw, media, container and
// supports packages, so a misspelled property or keyword fails to compile.
package style

import (
	"reflect"
	"slices"

	"github.com/Toms223/WebGo/app"
)

// Style is one stylesheet for app.Handler.Styles. It is an alias of app.Stylesheet, so any type with a
// StyleSource method is a Style.
type Style = app.Stylesheet

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

// Static links the CSS file at path as is: a URL, or a file under /web/ that must exist.
func Static(path string) Style { return static{path: path} }

// Generated renders items into the .css file at path, which must start with /web/, when the handler
// builds, and links it like a Static file. Nil items are ignored, and a bound item is
// ErrDynamicInGenerated.
func Generated(path string, items ...Item) Style {
	return generated{path: path, items: withoutNil(items)}
}

// Dynamic holds bound *Rule and *KeyframesRule items that Run inserts into and removes from the page
// as actions fire. It writes no file and adds no link; an unbound item is ErrNotDynamic.
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
