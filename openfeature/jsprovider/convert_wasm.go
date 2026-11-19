package jsprovider

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strings"
	"syscall/js"
	"time"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
)

func global(name string) app.Value { return app.Window().Get(name) }

func object() app.Value { return global("Object").New() }

func nullish(v app.Value) bool { return v.IsUndefined() || v.IsNull() }

func isObject(v app.Value) bool {
	t := v.Type()
	return t == app.TypeObject || t == app.TypeFunction
}

func jsString(v app.Value) string {
	if v.Type() == app.TypeString {
		return v.String()
	}
	return global("String").Invoke(v).String()
}

func optString(v app.Value) string {
	if nullish(v) {
		return ""
	}
	return jsString(v)
}

func try(f func()) (thrown app.Value) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(js.Error)
			if !ok {
				panic(r)
			}
			thrown = app.ValueOf(e.Value)
		}
	}()
	f()
	return nil
}

func invoke(this app.Value, method string, args ...any) (res, thrown app.Value) {
	thrown = try(func() {
		res = global("Reflect").Call("apply", this.Get(method), this, args)
	})
	return res, thrown
}

func thenable(v app.Value) bool {
	return isObject(v) && v.Get("then").Type() == app.TypeFunction
}

type thrownJS struct {
	v   app.Value
	msg string
}

func (e thrownJS) Error() string { return e.msg }

func (e thrownJS) JSString() string {
	var out string
	if try(func() { out = jsString(e.v) }) != nil {
		return ""
	}
	return out
}

func (e thrownJS) JSField(name string) (string, bool) {
	if !isObject(e.v) {
		return "", false
	}
	var out string
	var ok bool
	try(func() {
		if f := e.v.Get(name); !f.IsUndefined() {
			out, ok = jsString(f), true
		}
	})
	return out, ok
}

func (e thrownJS) JSIsError() bool { return isObject(e.v) && e.v.InstanceOf(global("Error")) }

func typeError(msg string) error { return goError(global("TypeError").New(msg)) }

func callback(f func(args []app.Value)) app.Func {
	return app.FuncOf(func(_ app.Value, args []app.Value) any {
		defer func() { _ = recover() }()
		f(args)
		return nil
	})
}

func goError(v app.Value) error {
	if !isObject(v) {
		return thrownJS{v: v}
	}
	t := thrownJS{v: v, msg: optString(v.Get("message"))}
	if c := v.Get("code"); c.Truthy() {
		return &openfeature.Error{Code: openfeature.ErrorCode(jsString(c)), Message: t.msg, Cause: t}
	}
	return t
}

func errorJS(err error) app.Value {
	var t thrownJS
	if errors.As(err, &t) {
		return t.v
	}
	e := global("Error").New(err.Error())
	var oe *openfeature.Error
	if errors.As(err, &oe) {
		e.Set("code", string(oe.Code))
	}
	return e
}

func date(t time.Time) app.Value { return global("Date").New(float64(t.UnixMilli())) }

type visit struct {
	p uintptr
	n int
	t reflect.Type
}

func toJS(v any) app.Value { return toJSIn(v, map[visit]app.Value{}) }

func toJSIn(v any, seen map[visit]app.Value) app.Value {
	switch x := v.(type) {
	case nil:
		return app.Null()
	case app.Value:
		return x
	case time.Time:
		return date(x)
	case *time.Time:
		if x == nil {
			return app.Null()
		}
		return date(*x)
	case openfeature.Object:
		if x == nil {
			return app.Null()
		}
		k := visit{p: reflect.ValueOf(x).Pointer(), n: len(x), t: reflect.TypeFor[openfeature.Object]()}
		if o, ok := seen[k]; ok {
			return o
		}
		o := object()
		seen[k] = o
		for _, f := range x {
			o.Set(f.Key, toJSIn(f.Value, seen))
		}
		return o
	case json.Marshaler:
		if rv := reflect.ValueOf(v); rv.Kind() == reflect.Pointer && rv.IsNil() {
			return app.Null()
		}
		return viaJSON(v)
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Bool:
		return app.ValueOf(rv.Bool())
	case reflect.String:
		return app.ValueOf(rv.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return app.ValueOf(float64(rv.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return app.ValueOf(float64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		return app.ValueOf(rv.Float())
	case reflect.Interface:
		if rv.IsNil() {
			return app.Null()
		}
		return toJSIn(rv.Elem().Interface(), seen)
	case reflect.Pointer:
		if rv.IsNil() {
			return app.Null()
		}
		k := visit{p: rv.Pointer(), t: rv.Type()}
		if o, ok := seen[k]; ok {
			return o
		}
		seen[k] = app.Null()
		o := toJSIn(rv.Elem().Interface(), seen)
		seen[k] = o
		return o
	case reflect.Map:
		if rv.IsNil() {
			return app.Null()
		}
		if rv.Type().Key().Kind() != reflect.String {
			break
		}
		k := visit{p: rv.Pointer(), t: rv.Type()}
		if o, ok := seen[k]; ok {
			return o
		}
		o := object()
		seen[k] = o
		keys := rv.MapKeys()
		slices.SortFunc(keys, func(a, b reflect.Value) int { return strings.Compare(a.String(), b.String()) })
		for _, key := range keys {
			o.Set(key.String(), toJSIn(rv.MapIndex(key).Interface(), seen))
		}
		return o
	case reflect.Slice:
		if rv.IsNil() {
			return app.Null()
		}
		k := visit{p: rv.Pointer(), n: rv.Len(), t: rv.Type()}
		if a, ok := seen[k]; ok {
			return a
		}
		a := global("Array").New()
		seen[k] = a
		for i := range rv.Len() {
			a.Call("push", toJSIn(rv.Index(i).Interface(), seen))
		}
		return a
	case reflect.Array:
		a := global("Array").New()
		for i := range rv.Len() {
			a.Call("push", toJSIn(rv.Index(i).Interface(), seen))
		}
		return a
	}
	return viaJSON(v)
}

func viaJSON(v any) app.Value {
	raw, err := json.Marshal(v)
	if err != nil {
		return app.Undefined()
	}
	return global("JSON").Call("parse", string(raw))
}

func contextJS(c openfeature.Context) app.Value {
	o := object()
	if c.TargetingKey != "" {
		o.Set("targetingKey", c.TargetingKey)
	}
	keys := make([]string, 0, len(c.Attributes))
	for k := range c.Attributes {
		if k != "targetingKey" || c.TargetingKey == "" {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)
	for _, k := range keys {
		o.Set(k, toJS(c.Attributes[k]))
	}
	return o
}

func skipped(v app.Value) bool {
	t := v.Type()
	return t == app.TypeUndefined || t == app.TypeFunction || t == app.TypeSymbol
}

type ancestor struct {
	v   app.Value
	out any
}

func fromJS(v app.Value) any { return fromJSIn(v, nil) }

func fromJSIn(v app.Value, anc []ancestor) any {
	switch v.Type() {
	case app.TypeBoolean:
		return v.Bool()
	case app.TypeNumber:
		return v.Float()
	case app.TypeString:
		return v.String()
	case app.TypeObject:
		if v.IsNull() {
			return nil
		}
		for _, a := range anc {
			if a.v.Equal(v) {
				return a.out
			}
		}
		if global("Array").Call("isArray", v).Bool() {
			out := make([]any, v.Length())
			anc = append(anc, ancestor{v, out})
			for i := range out {
				if e := v.Index(i); !skipped(e) {
					out[i] = fromJSIn(e, anc)
				}
			}
			return out
		}
		if v.InstanceOf(global("Date")) {
			ms := v.Call("getTime")
			if ms.IsNaN() {
				return nil
			}
			return time.UnixMilli(int64(ms.Float())).UTC()
		}
		keys := global("Object").Call("keys", v)
		var vals []app.Value
		var out openfeature.Object
		for i := range keys.Length() {
			k := keys.Index(i).String()
			if e := v.Get(k); !skipped(e) {
				out = append(out, openfeature.Field{Key: k})
				vals = append(vals, e)
			}
		}
		if out == nil {
			out = openfeature.Object{}
		}
		anc = append(anc, ancestor{v, out})
		for i, e := range vals {
			out[i].Value = fromJSIn(e, anc)
		}
		return out
	}
	return nil
}

func fromJSMap(v app.Value) (map[string]any, bool) {
	o, ok := fromJS(v).(openfeature.Object)
	if !ok {
		return nil, false
	}
	m := make(map[string]any, len(o))
	for _, f := range o {
		m[f.Key] = f.Value
	}
	return m, true
}

func loggerArg(v app.Value) any {
	if isObject(v) && v.InstanceOf(global("Error")) {
		return goError(v)
	}
	return jsString(v)
}
