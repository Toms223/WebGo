package script

import (
	"testing"
	"time"

	"github.com/Toms223/WebGo/app"
)

type fakeValue struct {
	app.Value
	truthy bool
	props  map[string]app.Value
}

func (f fakeValue) Get(p string) app.Value {
	if v, ok := f.props[p]; ok {
		return v
	}
	return fakeValue{}
}

func (f fakeValue) Truthy() bool { return f.truthy }

type counter struct{ inits *int }

func (c counter) Source() app.ModuleSource { return Local("c", "/web/c.js") }

func (c counter) Init(ctx app.Context, mod app.Value) string {
	*c.inits++
	return "built"
}

func reset(t *testing.T, mods app.Value) {
	oldServer, oldModules, oldEntries := isServer, modules, entries
	isServer = false
	modules = func() app.Value { return mods }
	entries = map[string]*entry{}
	t.Cleanup(func() { isServer, modules, entries = oldServer, oldModules, oldEntries })
}

func loaded(names ...string) fakeValue {
	props := map[string]app.Value{}
	for _, n := range names {
		props[n] = fakeValue{truthy: true}
	}
	return fakeValue{truthy: true, props: props}
}

func TestUseOnTheServerReturnsZeroAndFalse(t *testing.T) {
	reset(t, loaded("c"))
	isServer = true
	n := 0
	got, ok := Use(app.Context{}, counter{&n})
	if got != "" || ok || n != 0 {
		t.Fatalf("got (%q, %v), inits %d", got, ok, n)
	}
}

func TestUseInitsOnceAndCaches(t *testing.T) {
	reset(t, loaded("c"))
	n := 0
	for i := 0; i < 2; i++ {
		got, ok := Use(app.Context{}, counter{&n})
		if got != "built" || !ok {
			t.Fatalf("call %d: got (%q, %v)", i, got, ok)
		}
	}
	if n != 1 {
		t.Fatalf("inits = %d, want 1", n)
	}
}

func TestUseReportsAMissingModule(t *testing.T) {
	reset(t, loaded())
	logs := 0
	old := app.DefaultLogger
	app.DefaultLogger = func(string, ...any) { logs++ }
	t.Cleanup(func() { app.DefaultLogger = old })
	n := 0
	for i := 0; i < 2; i++ {
		got, ok := Use(app.Context{}, counter{&n})
		if got != "" || ok {
			t.Fatalf("call %d: got (%q, %v)", i, got, ok)
		}
	}
	if n != 0 || logs != 1 {
		t.Fatalf("inits = %d, logs = %d", n, logs)
	}
}

func TestUseWithoutTheModuleMapReturnsFalse(t *testing.T) {
	reset(t, fakeValue{})
	n := 0
	got, ok := Use(app.Context{}, counter{&n})
	if got != "" || ok || n != 0 {
		t.Fatalf("got (%q, %v), inits %d", got, ok, n)
	}
}

type outer struct{ innerOK *bool }

func (o outer) Source() app.ModuleSource { return Local("a", "/web/a.js") }

func (o outer) Init(ctx app.Context, mod app.Value) string {
	n := 0
	_, *o.innerOK = Use(ctx, counter{&n})
	return "outer"
}

func TestUseFromInitOfAnotherModuleWorks(t *testing.T) {
	reset(t, loaded("a", "c"))
	innerOK := false
	done := make(chan bool)
	go func() {
		_, ok := Use(app.Context{}, outer{&innerOK})
		done <- ok
	}()
	select {
	case ok := <-done:
		if !ok || !innerOK {
			t.Fatalf("outer ok = %v, inner ok = %v", ok, innerOK)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock")
	}
}

type other struct{ inits *int }

func (o other) Source() app.ModuleSource { return Local("c", "/web/c.js") }

func (o other) Init(ctx app.Context, mod app.Value) int {
	*o.inits++
	return 7
}

func TestUseSharedNameWithDifferentTypeReturnsZeroAndFalse(t *testing.T) {
	reset(t, loaded("c"))
	n := 0
	if _, ok := Use(app.Context{}, counter{&n}); !ok {
		t.Fatal("first Use failed")
	}
	m := 0
	got, ok := Use(app.Context{}, other{&m})
	if got != 0 || ok || m != 0 || n != 1 {
		t.Fatalf("got (%d, %v), inits %d %d", got, ok, n, m)
	}
}
