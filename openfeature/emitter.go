package openfeature

import (
	"slices"
	"sync"
)

// Handler receives an event.
type Handler func(EventDetails)

// EventDetails describes an event.
type EventDetails struct {
	ProviderName  string
	Domain        string
	ClientName    string
	Message       string
	FlagsChanged  []string
	ErrorCode     ErrorCode
	Metadata      map[string]any
	HasDomain     bool
	HasClientName bool
}

func (d *EventDetails) at(domain string, named bool) {
	d.ClientName, d.Domain, d.HasClientName, d.HasDomain = domain, domain, named, named
}

// Emitter lets a provider emit events. The zero value is ready to use.
type Emitter struct {
	mu     sync.Mutex
	next   uint64
	ls     []listener
	logger Logger
}

type listener struct {
	id uint64
	ev Event
	h  Handler
	fn func(Event, EventDetails)
}

var providerEvents = []Event{EventReady, EventError, EventConfigurationChanged, EventContextChanged, EventReconciling, EventStale}

var eventStatus = map[Status]Event{
	StatusReady:       EventReady,
	StatusError:       EventError,
	StatusFatal:       EventError,
	StatusStale:       EventStale,
	StatusReconciling: EventReconciling,
}

func statusMatchesEvent(ev Event, s Status) bool { return eventStatus[s] == ev }

// NewEmitter returns an Emitter for a provider.
func NewEmitter() *Emitter { return &Emitter{} }

// Emit runs the handlers for ev synchronously, in registration order.
func (e *Emitter) Emit(ev Event, d EventDetails) {
	e.mu.Lock()
	var run []listener
	for _, l := range e.ls {
		if l.ev == ev {
			run = append(run, l)
		}
	}
	e.mu.Unlock()
	for _, l := range run {
		quietly(func() { l.fn(ev, d) })
	}
}

// AddHandler registers h for ev and returns its remover. A panic in h is
// logged to the logger set with SetLogger, if any.
func (e *Emitter) AddHandler(ev Event, h Handler) func() {
	id := e.on(ev, h, func(_ Event, d EventDetails) {
		defer func() {
			if r := recover(); r != nil {
				e.mu.Lock()
				l := e.logger
				e.mu.Unlock()
				if l != nil {
					l.Error("Error running event handler:", r)
				}
			}
		}()
		h(d)
	})
	return func() { e.off(id) }
}

// Handlers returns every handler registered for ev, the SDK's included, in
// registration order.
func (e *Emitter) Handlers(ev Event) []Handler {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []Handler
	for _, l := range e.ls {
		if l.ev != ev {
			continue
		}
		if l.h != nil {
			out = append(out, l.h)
		} else {
			out = append(out, func(d EventDetails) { l.fn(ev, d) })
		}
	}
	return out
}

// RemoveAllHandlers removes every handler for the given events, the SDK's
// included, or every handler when no event is given.
func (e *Emitter) RemoveAllHandlers(evs ...Event) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if len(evs) == 0 {
		e.ls = nil
		return
	}
	var keep []listener
	for _, l := range e.ls {
		if !slices.Contains(evs, l.ev) {
			keep = append(keep, l)
		}
	}
	e.ls = keep
}

// SetLogger sets the logger for panics in handlers added with AddHandler. A
// nil logger prints a TypeError and falls back to the default logger.
func (e *Emitter) SetLogger(l Logger) {
	sl := newSafeLogger(l)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.logger = sl
}

func (e *Emitter) onEach(evs []Event, fn func(Event, EventDetails)) []uint64 {
	ids := make([]uint64, len(evs))
	for i, ev := range evs {
		ids[i] = e.on(ev, nil, fn)
	}
	return ids
}

func quietly(f func()) {
	defer func() { _ = recover() }()
	f()
}

func (e *Emitter) on(ev Event, h Handler, fn func(Event, EventDetails)) uint64 {
	if fn == nil {
		fn = func(_ Event, d EventDetails) { h(d) }
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.next++
	e.ls = append(e.ls, listener{id: e.next, ev: ev, h: h, fn: fn})
	return e.next
}

func (e *Emitter) off(ids ...uint64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	var keep []listener
	for _, l := range e.ls {
		removed := false
		for _, id := range ids {
			removed = removed || l.id == id
		}
		if !removed {
			keep = append(keep, l)
		}
	}
	e.ls = keep
}

func emitAll(ems []*Emitter, apiEm *Emitter, ev Event, d EventDetails) {
	for _, em := range ems {
		em.Emit(ev, d)
	}
	apiEm.Emit(ev, d)
}
