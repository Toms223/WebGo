package openfeature

import (
	"fmt"
	"testing"
)

func TestDeviation_events_rebind_stale_forwarding(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p, q := newTestProvider("p"), newTestProvider("q")
	if err := SetProvider(p); err != nil {
		t.Fatal(err)
	}
	var got []string
	rec := func(id string) Handler {
		return func(d EventDetails) { got = append(got, id+":"+d.ProviderName+":"+d.Domain) }
	}
	GetClient("d").AddHandler(EventStale, rec("d"))
	GetClient("y").AddHandler(EventConfigurationChanged, rec("y"))
	if err := SetProvider(q, ForDomain("d")); err != nil {
		t.Fatal(err)
	}
	p.ev.Emit(EventStale, EventDetails{})
	q.ev.Emit(EventStale, EventDetails{})
	if fmt.Sprint(got) != "[d:q:d]" {
		t.Fatalf("rebound domain: got %v", got)
	}
	got = nil
	if err := SetProvider(q); err != nil {
		t.Fatal(err)
	}
	if err := SetProvider(p, ForDomain("y")); err != nil {
		t.Fatal(err)
	}
	p.ev.Emit(EventConfigurationChanged, EventDetails{})
	if fmt.Sprint(got) != "[y:p:y]" {
		t.Fatalf("rebound default: got %v", got)
	}
}

func TestReplayReadsStatusPerProvider(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p, q := newTestProvider("p"), newTestProvider("q")
	if err := SetProviderAndWait(p); err != nil {
		t.Fatal(err)
	}
	if err := SetProviderAndWait(q, ForDomain("d")); err != nil {
		t.Fatal(err)
	}
	var got []string
	AddHandler(EventReady, func(d EventDetails) {
		got = append(got, d.ProviderName)
		q.ev.Emit(EventError, EventDetails{})
	})
	if fmt.Sprint(got) != "[p]" {
		t.Fatalf("replayed for %v, want [p]", got)
	}
	if s := GetClient("d").ProviderStatus(); s != StatusError {
		t.Fatalf("d status = %s", s)
	}
}

type argLogger struct{ lines [][]any }

func (l *argLogger) Error(a ...any) { l.lines = append(l.lines, a) }
func (l *argLogger) Warn(...any)    {}
func (l *argLogger) Info(...any)    {}
func (l *argLogger) Debug(...any)   {}

func TestEmitterPublicMethods(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := newTestProvider("p")
	if err := SetProvider(p); err != nil {
		t.Fatal(err)
	}
	if r, c := len(p.ev.Handlers(EventReady)), len(p.ev.Handlers(EventConfigurationChanged)); r != 2 || c != 1 {
		t.Fatalf("SDK handlers: ready %d, configuration %d; want 2, 1", r, c)
	}
	p.ev.RemoveAllHandlers(EventStale)
	ran := 0
	AddHandler(EventConfigurationChanged, func(EventDetails) { ran++ })
	p.ev.Emit(EventStale, EventDetails{})
	if s := ProviderStatus(); s != StatusReady {
		t.Fatalf("status after removed stale tracking = %s", s)
	}
	p.ev.RemoveAllHandlers()
	p.ev.Emit(EventConfigurationChanged, EventDetails{})
	if ran != 0 {
		t.Fatal("forwarding survived RemoveAllHandlers")
	}

	e := NewEmitter()
	var got []string
	h := func(d EventDetails) { got = append(got, d.Message) }
	e.AddHandler(EventReady, h)
	remove := e.AddHandler(EventReady, h)
	e.AddHandler(EventReady, func(EventDetails) { panic("x") })
	if n := len(e.Handlers(EventReady)); n != 3 {
		t.Fatalf("Handlers = %d", n)
	}
	e.Emit(EventReady, EventDetails{Message: "1"})
	remove()
	l := &argLogger{}
	e.SetLogger(l)
	e.Emit(EventReady, EventDetails{Message: "2"})
	if fmt.Sprint(got) != "[1 1 2]" || fmt.Sprint(l.lines) != "[[Error running event handler: x]]" {
		t.Fatalf("got %v, logs %v", got, l.lines)
	}
}
