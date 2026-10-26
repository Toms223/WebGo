package app

import "reflect"

// EventHandler handles a DOM event. It runs on the UI goroutine, and the source
// component re-renders afterwards unless it calls Context.PreventUpdate.
type EventHandler func(ctx Context, e Event)

type Event struct {
	Value
}

func (e Event) PreventDefault() {
	e.Call("preventDefault")
}

func (e Event) StopImmediatePropagation() {
	e.Call("stopImmediatePropagation")
}

type EventOption struct {
	name  string
	value string
}

// EventScope adds v to a handler's identity. Handlers made from the same
// function literal look equal, so in a list give each a scope, such as the item
// ID, or a reused element keeps the old closure.
func EventScope(v ...any) EventOption {
	return EventOption{
		name:  "scope",
		value: toPath(v...),
	}
}

// PassiveEvent registers the listener as passive: it cannot call
// PreventDefault, which lets the browser scroll smoothly. Use it for scroll,
// touch and wheel events.
func PassiveEvent() EventOption {
	return EventOption{
		name: "passive",
	}
}

type eventHandlers map[string]eventHandler

func (h eventHandlers) Set(event string, eh EventHandler, options ...EventOption) {
	if eh != nil {
		h[event] = makeEventHandler(event, eh, options...)
	}
}

type eventHandler struct {
	event     string
	scope     string
	passive   bool
	goHandler EventHandler
	jsHandler Func
	close     func()
}

func makeEventHandler(event string, h EventHandler, options ...EventOption) eventHandler {
	handler := eventHandler{
		event:     event,
		goHandler: h,
	}

	for _, option := range options {
		switch option.name {
		case "scope":
			handler.scope = option.value

		case "passive":
			handler.passive = true
		}
	}
	return handler
}

func (h eventHandler) Equal(v eventHandler) bool {
	return h.event == v.event &&
		h.scope == v.scope &&
		h.passive == v.passive &&
		reflect.ValueOf(h.goHandler).Pointer() == reflect.ValueOf(v.goHandler).Pointer()
}

func (h eventHandler) options() map[string]any {
	if h.passive {
		return map[string]any{
			"passive": true,
		}
	}
	return nil
}

func trackMousePosition(e Event) {
	x := e.Get("clientX")
	if !x.Truthy() {
		return
	}

	y := e.Get("clientY")
	if !y.Truthy() {
		return
	}

	Window().setCursorPosition(x.Int(), y.Int())
}
