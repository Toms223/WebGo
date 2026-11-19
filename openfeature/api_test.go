package openfeature

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"testing/synctest"
)

type testProvider struct {
	name   string
	runsOn string
	scoped bool
	ev     *Emitter
}

func (p *testProvider) Metadata() Metadata { return Metadata{Name: p.name} }
func (p *testProvider) ResolveBoolean(_ string, def bool, _ Context, _ Logger) ResolutionDetails[bool] {
	return ResolutionDetails[bool]{Value: def}
}
func (p *testProvider) ResolveString(_ string, def string, _ Context, _ Logger) ResolutionDetails[string] {
	return ResolutionDetails[string]{Value: def}
}
func (p *testProvider) ResolveNumber(_ string, def float64, _ Context, _ Logger) ResolutionDetails[float64] {
	return ResolutionDetails[float64]{Value: def}
}
func (p *testProvider) ResolveObject(_ string, def any, _ Context, _ Logger) ResolutionDetails[any] {
	return ResolutionDetails[any]{Value: def}
}
func (p *testProvider) RunsOn() string          { return p.runsOn }
func (p *testProvider) DomainScoped() bool      { return p.scoped }
func (p *testProvider) Events() *Emitter        { return p.ev }
func newTestProvider(name string) *testProvider { return &testProvider{name: name, ev: NewEmitter()} }

type memLogger struct {
	mu    sync.Mutex
	lines []string
}

func (l *memLogger) Error(a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, fmt.Sprint(a[0]))
}
func (l *memLogger) Warn(...any)  {}
func (l *memLogger) Info(...any)  {}
func (l *memLogger) Debug(...any) {}

func TestSetProviderRejectsServerProvider(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := newTestProvider("x")
	p.runsOn = "server"
	err := SetProvider(p)
	if err == nil || err.Error() != "Provider 'x' is intended for use on the server." || codeOf(err) != CodeGeneral {
		t.Fatalf("got %v", err)
	}
	if ProviderStatus() != StatusNotReady || ProviderMetadata().Name != "No-op Provider" {
		t.Fatal("server provider was bound")
	}
}

func TestDomainScopedDoubleBind(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := newTestProvider("x")
	p.scoped = true
	if err := SetProvider(p, ForDomain("d1")); err != nil {
		t.Fatal(err)
	}
	err := SetProvider(p, ForDomain("d2"))
	if err == nil || err.Error() != "Cannot bind domain-scoped provider 'x' to more than one domain." {
		t.Fatalf("got %v", err)
	}
	if ProviderStatus("d2") != StatusNotReady {
		t.Fatal("d2 was bound")
	}
}

func TestRemoveHandlerRemovesOnlyThatRegistration(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := newTestProvider("p")
	if err := SetProvider(p); err != nil {
		t.Fatal(err)
	}
	var got []string
	h := func(id string) Handler { return func(EventDetails) { got = append(got, id) } }
	removeA := AddHandler(EventConfigurationChanged, h("a"))
	AddHandler(EventConfigurationChanged, h("b"))
	removeC := GetClient().AddHandler(EventConfigurationChanged, h("c"))
	GetClient().AddHandler(EventConfigurationChanged, h("d"))
	removeA()
	removeC()
	removeA()
	p.ev.Emit(EventConfigurationChanged, EventDetails{})
	if fmt.Sprint(got) != "[b d]" {
		t.Fatalf("got %v", got)
	}
	if len(Handlers(EventConfigurationChanged)) != 1 || len(GetClient().Handlers(EventConfigurationChanged)) != 1 {
		t.Fatal("handler lists")
	}
}

func TestHandlerPanicIsLoggedAndOthersRun(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	l := &memLogger{}
	SetLogger(l)
	p := newTestProvider("p")
	var got []string
	AddHandler(EventStale, func(EventDetails) { got = append(got, "a"); panic("a") })
	AddHandler(EventStale, func(EventDetails) { got = append(got, "b") })
	if err := SetProvider(p); err != nil {
		t.Fatal(err)
	}
	p.ev.Emit(EventStale, EventDetails{})
	if fmt.Sprint(got) != "[a b]" || len(l.lines) != 0 {
		t.Fatalf("dispatch: got %v, logs %v", got, l.lines)
	}
	AddHandler(EventStale, func(EventDetails) { got = append(got, "c"); panic("c") })
	GetClient().AddHandler(EventStale, func(EventDetails) { got = append(got, "d"); panic("d") })
	if fmt.Sprint(got) != "[a b c d]" || fmt.Sprint(l.lines) != "[Error running event handler: Error running event handler:]" {
		t.Fatalf("replay: got %v, logs %v", got, l.lines)
	}
}

type sliceProvider struct {
	noopProvider
	tags []string
}

func TestSameNonComparableProvider(t *testing.T) {
	var a, b Provider = sliceProvider{}, sliceProvider{}
	if same(a, b) {
		t.Fatal("non-comparable providers reported same")
	}
	var x, y any = []int{1}, []int{1}
	var i, j Provider = struct {
		noopProvider
		v any
	}{v: x}, struct {
		noopProvider
		v any
	}{v: y}
	if same(i, j) {
		t.Fatal("panic not recovered as not same")
	}
}

func TestErrorDetailsEmptyCodeIsGeneral(t *testing.T) {
	d := errorDetails("f", true, NewError("", "m"), FlagMetadata{})
	if d.ErrorCode != CodeGeneral || d.ErrorMessage != "m" {
		t.Fatalf("got %+v", d)
	}
}

func TestEmptyDomainIgnored(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	if err := SetProvider(newTestProvider("p"), ForDomain("")); err != nil {
		t.Fatal(err)
	}
	if ProviderMetadata().Name != "No-op Provider" {
		t.Fatal("empty domain bound a provider")
	}
	if GetClient("").Metadata().HasDomain != true || GetClient().Metadata().HasDomain {
		t.Fatal("HasDomain")
	}
}

func TestSetProviderReentrantFromReadyHandler(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		Reset()
		defer Reset()
		q := newTestProvider("q")
		nested := false
		AddHandler(EventReady, func(EventDetails) {
			if nested {
				return
			}
			nested = true
			if err := SetProvider(q, ForDomain("d")); err != nil {
				t.Error(err)
			}
		})
		done := make(chan struct{})
		go func() {
			defer close(done)
			if err := SetProvider(newTestProvider("p")); err != nil {
				t.Error(err)
			}
		}()
		synctest.Wait()
		select {
		case <-done:
		default:
			t.Fatal("SetProvider deadlocked")
		}
		if ProviderMetadata("d").Name != "q" || ProviderStatus("d") != StatusReady {
			t.Fatal("re-entrant SetProvider did not bind d")
		}
	})
}

type panickyEvents struct {
	*testProvider
	calls int
}

func (p *panickyEvents) Events() *Emitter { p.calls++; panic(errors.New("full")) }

func TestEventsPanicIsSetProviderError(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := &panickyEvents{testProvider: newTestProvider("x")}
	p.runsOn = "server"
	if err := SetProvider(p); err == nil || p.calls != 0 {
		t.Fatalf("runsOn check must run before Events: err %v, calls %d", err, p.calls)
	}
	p.runsOn = "client"
	if err := SetProviderAndWait(p, ForDomain("d")); err == nil || err.Error() != "full" || p.calls != 1 {
		t.Fatalf("got %v after %d calls, want the Events panic as an error", err, p.calls)
	}
	if ProviderStatus("d") != StatusNotReady {
		t.Fatal("provider was bound")
	}
}
