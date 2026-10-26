package app

import (
	"context"
	"net/url"
	"reflect"

	"github.com/Toms223/WebGo/internal/errors"
)

// TestEngine runs components outside the browser for unit tests. Dispatched
// work stays queued until ConsumeNext or ConsumeAll runs it.
type TestEngine interface {
	Load(Composer) error

	ConsumeNext()

	ConsumeAll()
}

// NewTestEngine returns a TestEngine. Load a component, call ConsumeAll, then
// check the tree with Match.
func NewTestEngine() TestEngine {
	origin, _ := url.Parse("/")
	originPage := makeRequestPage(origin, nil)

	routes := makeRouter()
	return newEngine(context.Background(),
		&routes,
		nil,
		&originPage,
		map[string]ActionHandler{
			"/test": func(ctx Context, a Action) {},
		},
	)
}

// Match returns an error unless the element at path below root matches
// expected. path indexes children level by level; text compares by value, HTML
// elements by tag, attributes and handlers, components by exported fields.
func Match(expected UI, root UI, path ...int) error {
	return TestMatch(root, TestUIDescriptor{
		Path:     TestPath(path...),
		Expected: expected,
	})
}

type TestUIDescriptor struct {
	Path []int

	Expected UI
}

func TestPath(p ...int) []int {
	return p
}

// TestMatch is Match with a TestUIDescriptor.
func TestMatch(root UI, d TestUIDescriptor) error {
	if len(d.Path) != 0 {
		index := d.Path[0]

		switch root := root.(type) {
		case HTML:
			children := root.body()
			if index < 0 || index >= len(children) {
				return errors.New("element to match is out of range").
					WithTag("type", reflect.TypeOf(d.Expected)).
					WithTag("parent-type", reflect.TypeOf(root)).
					WithTag("parent-children-count", len(children)).
					WithTag("index", index)
			}
			d.Path = d.Path[1:]
			return TestMatch(children[index], d)

		case Composer:
			if index != 0 {
				return errors.New("element to match is out of range").
					WithTag("type", reflect.TypeOf(d.Expected)).
					WithTag("parent-type", reflect.TypeOf(root)).
					WithTag("parent-children-count", 1).
					WithTag("index", index)
			}
			d.Path = d.Path[1:]
			return TestMatch(root.root(), d)
		}
	}

	return match(root, d)
}

func match(n UI, d TestUIDescriptor) error {
	if a, b := reflect.TypeOf(d.Expected), reflect.TypeOf(n); a != b {
		return errors.New("types are not matching").
			WithTag("type", a).
			WithTag("expected-type", b)
	}

	switch d.Expected.(type) {
	case *text:
		return matchText(n.(*text), d)

	case HTML:
		return matchHTML(n.(HTML), d)

	case Composer:
		return matchComponent(n.(Composer), d)

	case *raw:
		return matchRaw(n.(*raw), d)

	default:
		return errors.New("unsupported element").
			WithTag("type", reflect.TypeOf(n))
	}
}

func matchText(n *text, d TestUIDescriptor) error {
	a := n
	b := d.Expected.(*text)

	if a.value != b.value {
		return errors.New("text does not match").
			WithTag("type", reflect.TypeOf(a)).
			WithTag("expected-value", b.value).
			WithTag("current-value", a.value)
	}
	return nil
}

func matchHTML(n HTML, d TestUIDescriptor) error {
	a := n
	b := d.Expected.(HTML)

	if typeA, typeB := reflect.TypeOf(a), reflect.TypeOf(b); typeA != typeB || a.Tag() != b.Tag() {
		return errors.New("types are not matching").
			WithTag("type", typeA).
			WithTag("expected-type", typeB)
	}

	if err := matchHTMLAttributes(a.attrs(), b.attrs()); err != nil {
		return errors.New("attributes does not match").
			WithTag("type", reflect.TypeOf(a)).
			Wrap(err)
	}

	if err := matchHTMLEventHandlers(a.events(), b.events()); err != nil {
		return errors.New("event handlers does not match").
			WithTag("type", reflect.TypeOf(a)).
			Wrap(err)
	}

	return nil
}

func matchHTMLAttributes(a, b attributes) error {
	for key, expectedValue := range b {
		value, exists := a[key]
		if !exists {
			return errors.New("expected attribute not found").
				WithTag("name", key)
		}

		if value != expectedValue {
			return errors.New("value does not match").
				WithTag("name", key).
				WithTag("value", value).
				WithTag("expected-value", expectedValue)
		}
	}

	for key, value := range a {
		if _, exists := b[key]; !exists {
			return errors.New("attribute is not expected").
				WithTag("name", key).
				WithTag("value", value)
		}
	}

	return nil
}

func matchHTMLEventHandlers(a, b eventHandlers) error {
	for key := range b {
		if _, exists := a[key]; !exists {
			return errors.New("expected event handler not found").
				WithTag("event", key)
		}
	}

	for key := range a {
		if _, exists := b[key]; !exists {
			return errors.New("event handler is not expected").
				WithTag("event", key)
		}
	}

	return nil
}

func matchComponent(n Composer, d TestUIDescriptor) error {
	a := reflect.Indirect(reflect.ValueOf(n))
	b := reflect.Indirect(reflect.ValueOf(d.Expected))

	for i := 0; i < b.NumField(); i++ {
		fieldA := a.Field(i)
		fieldB := b.Field(i)

		if !fieldA.CanSet() {
			continue
		}

		if _, ok := fieldA.Interface().(Compo); ok {
			continue
		}

		if !reflect.DeepEqual(fieldA.Interface(), fieldB.Interface()) {
			return errors.New("field are not matching").
				WithTag("type", reflect.TypeOf(n)).
				WithTag("field", a.Type().Field(i).Name).
				WithTag("value", fieldA.Interface()).
				WithTag("expected-value", fieldB.Interface())
		}
	}

	return nil
}

func matchRaw(n *raw, d TestUIDescriptor) error {
	a := n
	b := d.Expected.(*raw)

	if a.value != b.value {
		return errors.New("the raw html element is not matching with the descriptor").
			WithTag("type", reflect.TypeOf(n)).
			WithTag("reason", "unexpected value").
			WithTag("expected-value", b.value).
			WithTag("current-value", a.value)
	}

	return nil
}
