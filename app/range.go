package app

import (
	"reflect"
	"sort"

	"github.com/Toms223/WebGo/internal/errors"
)

type RangeLoop interface {
	UI

	Slice(f func(int) UI) RangeLoop

	Map(f func(string) UI) RangeLoop

	body() []UI
}

// Range builds UI from src: call Slice for a slice or array, or Map for a map
// with string keys, which visits keys in sorted order. Both panic on any other
// kind.
func Range(src any) RangeLoop {
	return rangeLoop{source: src}
}

type rangeLoop struct {
	children []UI
	source   any
}

func (r rangeLoop) Slice(f func(int) UI) RangeLoop {
	src := reflect.ValueOf(r.source)
	if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
		panic(errors.New("range loop source is not a slice or array").
			WithTag("src-type", src.Type),
		)
	}

	body := make([]UI, 0, src.Len())
	for i := 0; i < src.Len(); i++ {
		body = append(body, FilterUIElems(f(i))...)
	}

	r.children = body
	return r
}

func (r rangeLoop) Map(f func(string) UI) RangeLoop {
	src := reflect.ValueOf(r.source)
	if src.Kind() != reflect.Map {
		panic(errors.New("range loop source is not a map").
			WithTag("src-type", src.Type),
		)
	}

	if keyType := src.Type().Key(); keyType.Kind() != reflect.String {
		panic(errors.New("range loop source keys are not strings").
			WithTag("src-type", src.Type).
			WithTag("key-type", keyType),
		)
	}

	body := make([]UI, 0, src.Len())
	keys := make([]string, 0, src.Len())

	for _, k := range src.MapKeys() {
		keys = append(keys, k.String())
	}
	sort.Strings(keys)

	for _, k := range keys {
		body = append(body, FilterUIElems(f(k))...)
	}

	r.children = body
	return r
}

func (r rangeLoop) JSValue() Value {
	return nil
}

func (r rangeLoop) Mounted() bool {
	return false
}

func (r rangeLoop) setParent(UI) UI {
	return nil
}

func (r rangeLoop) parent() UI {
	return nil
}

func (r rangeLoop) body() []UI {
	return r.children
}
