package script

import (
	"sync"

	"github.com/Toms223/WebGo/app"
)

var isServer = app.IsServer

var modules = func() app.Value { return app.Window().Get("goappModules") }

type entry struct {
	once  sync.Once
	value any
	ok    bool
}

var (
	entries = map[string]*entry{}
	mu      sync.Mutex
)

// Use returns the T for s, building it with Init on the first call for that
// module name. ok is false on the server and when the module failed to load;
// T is then its zero value and Init is not called. Init may call Use for
// another module, never for its own.
// Wrappers sharing a module name must share T. Call it from the UI goroutine
// or from Async.
func Use[T any](ctx app.Context, s Script[T]) (T, bool) {
	var zero T
	if isServer {
		return zero, false
	}
	name, _ := s.Source().ImportMap()
	mu.Lock()
	e, found := entries[name]
	if !found {
		e = &entry{}
		entries[name] = e
	}
	mu.Unlock()
	e.once.Do(func() {
		mods := modules()
		if !mods.Truthy() {
			return
		}
		mod := mods.Get(name)
		if !mod.Truthy() {
			app.Log("script: module not loaded: " + name)
			return
		}
		e.value = s.Init(ctx, mod)
		e.ok = true
	})
	if !e.ok {
		return zero, false
	}
	v, ok := e.value.(T)
	return v, ok
}
