package app

import (
	"net/url"
)

type Type int

const (
	TypeUndefined Type = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeString
	TypeSymbol
	TypeObject
	TypeFunction
)

func (t Type) String() string {
	switch t {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeString:
		return "string"
	case TypeSymbol:
		return "symbol"
	case TypeObject:
		return "object"
	case TypeFunction:
		return "function"
	default:
		panic("bad type")
	}
}

type Wrapper interface {
	JSValue() Value
}

// Value mirrors syscall/js.Value. In the browser it wraps a JS value; on the
// server it is an inert stub whose methods return zero values.
type Value interface {
	Bool() bool

	Call(m string, args ...any) Value

	Delete(p string)

	Equal(w Value) bool

	Float() float64

	Get(p string) Value

	Index(i int) Value

	InstanceOf(t Value) bool

	Int() int

	Invoke(args ...any) Value

	IsNaN() bool

	IsNull() bool

	IsUndefined() bool

	JSValue() Value

	Length() int

	New(args ...any) Value

	Set(p string, x any)

	SetIndex(i int, x any)

	String() string

	Truthy() bool

	Type() Type

	// Then calls f with the result once the promise resolves.
	Then(f func(Value))

	getAttr(k string) string
	setAttr(k, v string)
	delAttr(k string)
	firstChild() Value
	appendChild(c Wrapper)
	replaceChild(new, old Wrapper)
	removeChild(c Wrapper)
	firstElementChild() Value
	addEventListener(event string, fn Func, options map[string]any)
	removeEventListener(event string, fn Func)
	setNodeValue(v string)
	setInnerHTML(v string)
	setInnerText(v string)
}

func Null() Value {
	return null()
}

func Undefined() Value {
	return undefined()
}

// ValueOf converts x to a JS value like js.ValueOf: nil, bool, numbers, string,
// []any, map[string]any, Value and Func. In the browser it panics on other
// types.
func ValueOf(x any) Value {
	return valueOf(x)
}

// Func is a Go function callable from JS. Call Release once JS no longer needs
// it.
type Func interface {
	Value

	Release()
}

// FuncOf wraps fn for JS like js.FuncOf. fn runs on the JS event loop, so
// blocking in it freezes the page; start a goroutine for slow work.
func FuncOf(fn func(this Value, args []Value) any) Func {
	return funcOf(fn)
}

type BrowserWindow interface {
	Value

	URL() *url.URL

	Size() (w, h int)

	CursorPosition() (x, y int)

	setCursorPosition(x, y int)

	GetElementByID(id string) Value

	ScrollToID(id string)

	setBody(body UI)
	createElement(tag, xmlns string) (Value, error)
	createTextNode(v string) Value
	addHistory(u *url.URL)
	replaceHistory(u *url.URL)
}

func CopyBytesToGo(dst []byte, src Value) int {
	return copyBytesToGo(dst, src)
}

func CopyBytesToJS(dst Value, src []byte) int {
	return copyBytesToJS(dst, src)
}
