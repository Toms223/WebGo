package app

import (
	"fmt"
	"sync"
)

// Action is an event posted with Context.NewAction, with an optional Value and
// Tags.
type Action struct {
	Name string

	Value any

	Tags Tags
}

// ActionHandler handles an Action. Its Context is bound to the component that
// registered it.
type ActionHandler func(Context, Action)

// Handle registers h as the global handler for actionName, replacing any
// previous one. Global handlers run in their own goroutine, not on the UI
// goroutine, and must be registered before RunWhenOnBrowser.
func Handle(actionName string, h ActionHandler) {
	actionHandlers[actionName] = h
}

var actionHandlers = make(map[string]ActionHandler)

type actionHandler struct {
	Source   UI
	Function ActionHandler
	Async    bool
}

type actionManager struct {
	mutex    sync.Mutex
	handlers map[string]map[string]actionHandler
}

func (m *actionManager) Handle(action string, source UI, async bool, handler ActionHandler) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.handlers == nil {
		m.handlers = make(map[string]map[string]actionHandler)
	}

	handlers, ok := m.handlers[action]
	if !ok {
		handlers = make(map[string]actionHandler)
		m.handlers[action] = handlers
	}

	key := actionHandlerKey(source, handler)
	handlers[key] = actionHandler{
		Source:   source,
		Function: handler,
		Async:    async,
	}
}

func (m *actionManager) Post(ctx Context, a Action) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for key, handler := range m.handlers[a.Name] {
		source := handler.Source
		if !source.Mounted() {
			delete(m.handlers[a.Name], key)
			continue
		}
		ctx.sourceElement = source

		function := handler.Function
		if handler.Async {
			ctx.Async(func() {
				function(ctx, a)
			})
			continue
		}

		ctx.Dispatch(func(ctx Context) {
			function(ctx, a)
		})
	}
}

func (m *actionManager) Cleanup() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for action, handlers := range m.handlers {
		for key, handler := range handlers {
			if !handler.Source.Mounted() {
				delete(handlers, key)
			}
		}

		if len(handlers) == 0 {
			delete(m.handlers, action)
		}
	}
}

func actionHandlerKey(source UI, handler ActionHandler) string {
	return fmt.Sprintf("/%T/%p/%p", source, source, handler)
}
