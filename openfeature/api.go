package openfeature

import (
	"fmt"
	"reflect"
	"slices"
	"sync"
)

type wrapper struct {
	p       Provider
	ev      *Emitter
	status  Status
	pending int
	emits   bool
}

type binding struct {
	domain string
	w      *wrapper
}

type registration struct {
	src *Emitter
	ids []uint64
}

func (r *registration) drop() {
	if r.src != nil {
		r.src.off(r.ids...)
	}
	*r = registration{}
}

type clientEmitter struct {
	domain string
	named  bool
	em     *Emitter
	fwd    registration
}

type state struct {
	def       *wrapper
	bindings  []binding
	ctx       Context
	domainCtx map[string]Context
	hooks     []Hook
	logger    Logger
	apiEm     *Emitter
	clientEms []*clientEmitter
	transfers map[string]*registration
}

var states []*state

func dropState(s *state) {
	states = slices.DeleteFunc(states, func(o *state) bool { return o == s })
}

// API is an OpenFeature API instance with its own providers, contexts, hooks
// and handlers. The package-level functions act on a shared default instance;
// package isolated creates independent ones.
type API struct {
	s *state
}

var (
	mu     sync.Mutex
	global = &API{}
)

func newState() *state {
	return &state{
		def:       &wrapper{p: noop, status: StatusNotReady},
		domainCtx: map[string]Context{},
		logger:    defaultLogger{},
		apiEm:     NewEmitter(),
		transfers: map[string]*registration{},
	}
}

func (a *API) st() *state {
	if a == nil {
		a = global
	}
	if a.s == nil {
		a.s = newState()
	}
	return a.s
}

func (a *API) current() *state {
	mu.Lock()
	defer mu.Unlock()
	return a.st()
}

func emitsLifecycle(p Provider, ev *Emitter) bool {
	l, ok := p.(LifecycleEmitter)
	return ok && ev != nil && l.EmitsLifecycleEvents()
}

func (s *state) wrapperOf(p Provider) *wrapper {
	if same(s.def.p, p) {
		return s.def
	}
	for _, b := range s.bindings {
		if same(b.w.p, p) {
			return b.w
		}
	}
	return nil
}

func (s *state) registeredElsewhere(p Provider) bool {
	for _, o := range states {
		if o != s && o.inAll(p) {
			return true
		}
	}
	return false
}

func (s *state) log() Logger {
	mu.Lock()
	defer mu.Unlock()
	return s.logger
}

func optDomain(domain []string) string {
	if len(domain) == 0 {
		return ""
	}
	return domain[0]
}

func same(a, b Provider) (eq bool) {
	defer func() {
		if recover() != nil {
			eq = false
		}
	}()
	ta, tb := reflect.TypeOf(a), reflect.TypeOf(b)
	return ta == tb && ta != nil && ta.Comparable() && a == b
}

func (s *state) wrapperFor(domain string) *wrapper {
	if domain != "" {
		for _, b := range s.bindings {
			if b.domain == domain {
				return b.w
			}
		}
	}
	return s.def
}

func (s *state) bound(domain string) *wrapper {
	for _, b := range s.bindings {
		if b.domain == domain {
			return b.w
		}
	}
	return nil
}

func (s *state) bind(domain string, w *wrapper) {
	if domain == "" {
		s.def = w
		return
	}
	for i, b := range s.bindings {
		if b.domain == domain {
			s.bindings[i].w = w
			return
		}
	}
	s.bindings = append(s.bindings, binding{domain, w})
}

func (s *state) inAll(p Provider) bool {
	for _, b := range s.bindings {
		if same(b.w.p, p) {
			return true
		}
	}
	return same(s.def.p, p)
}

func (s *state) instanceBinding(p Provider) (string, bool) {
	if same(s.def.p, p) {
		return "", true
	}
	for _, b := range s.bindings {
		if same(b.w.p, p) {
			return b.domain, true
		}
	}
	return "", false
}

func eventsOf(p Provider) *Emitter {
	if e, ok := p.(Eventer); ok {
		return e.Events()
	}
	return nil
}

func newWrapper(p Provider, ev *Emitter) *wrapper {
	w := &wrapper{p: p, ev: ev, status: StatusNotReady, emits: emitsLifecycle(p, ev)}
	if ev != nil {
		evs := []Event{EventReady, EventStale, EventError}
		if w.emits {
			evs = append(evs, EventReconciling, EventContextChanged)
		}
		ev.onEach(evs, func(e Event, d EventDetails) {
			mu.Lock()
			defer mu.Unlock()
			switch {
			case e == EventReady || e == EventContextChanged:
				w.status = StatusReady
			case e == EventReconciling:
				w.status = StatusReconciling
			case e == EventStale:
				w.status = StatusStale
			case e == EventError && d.ErrorCode == CodeProviderFatal:
				w.status = StatusFatal
			case e == EventError:
				w.status = StatusError
			}
		})
	}
	return w
}

func (s *state) setAwaitableProvider(domain string, p Provider) (func() error, error) {
	if p == nil {
		s.log().Debug("No provider defined, ignoring setProvider call")
		return nil, nil
	}
	name := p.Metadata().Name
	scoped := false
	if d, ok := p.(DomainScoped); ok {
		scoped = d.DomainScoped()
	}
	runsOn := ""
	if r, ok := p.(RunsOn); ok {
		runsOn = r.RunsOn()
	}
	named := domain != ""

	var (
		old      Provider
		scopeErr bool
		ignored  bool
		taken    bool
		l        Logger
	)
	func() {
		mu.Lock()
		defer mu.Unlock()
		l = s.logger
		old = s.wrapperFor(domain).p
		if scoped {
			existing, ok := s.instanceBinding(p)
			scopeErr = ok && existing != domain
		}
		ignored = same(old, p)
		taken = s.registeredElsewhere(p)
	}()
	if taken {
		return nil, NewError(CodeGeneral, fmt.Sprintf("Provider '%s' is already registered with another API instance.", name))
	}
	if scopeErr {
		return nil, NewError(CodeGeneral, fmt.Sprintf("Cannot bind domain-scoped provider '%s' to more than one domain.", name))
	}
	if ignored {
		l.Debug("Provider is already set, ignoring setProvider call")
		return nil, nil
	}
	if runsOn == "" {
		l.Debug(fmt.Sprintf("Provider '%s' has not defined its intended use.", name))
	} else if runsOn != "client" {
		return nil, NewError(CodeGeneral, fmt.Sprintf("Provider '%s' is intended for use on the %s.", name, runsOn))
	}
	var ev *Emitter
	if err := catch(func() error { ev = eventsOf(p); return nil }); err != nil {
		return nil, err
	}

	var (
		emitters []*clientEmitter
		w        *wrapper
		initable bool
		initCtx  Context
		closeOld bool
	)
	init, isInit := p.(Initializer)
	attach := func() {
		mu.Lock()
		defer mu.Unlock()
		s.bind(domain, w)
		s.transferListeners(p, ev, domain, emitters)
		closeOld = !s.inAll(old)
		if !slices.Contains(states, s) {
			states = append(states, s)
		}
	}
	func() {
		mu.Lock()
		defer mu.Unlock()
		emitters = s.associatedEmitters(domain)
		w = newWrapper(p, ev)
		initable = isInit && !s.inAll(p)
		if prev := s.wrapperOf(p); prev != nil && w.emits {
			w.status = prev.status
		}
		initCtx = s.ctx
		if c, ok := s.domainCtx[domain]; ok && named {
			initCtx = c
		}
	}()

	var wait func() error
	if initable && w.emits {
		attach()
	}
	if initable {
		var (
			pend Pending
			ierr error
		)
		if err := catch(func() error { pend, ierr = init.Initialize(initCtx.copy(), domain); return nil }); err != nil {
			return nil, err
		}
		wait = func() error {
			err := ierr
			if err == nil && pend != nil {
				err = <-pend
			}
			if w.emits {
				return err
			}
			d := EventDetails{ProviderName: name}
			d.at(domain, named)
			mu.Lock()
			if err == nil {
				w.status = StatusReady
			} else if codeOf(err) == CodeProviderFatal {
				w.status = StatusFatal
			} else {
				w.status = StatusError
			}
			ems := emittersOf(s.associatedEmitters(domain))
			mu.Unlock()
			if err == nil {
				emitAll(ems, s.apiEm, EventReady, d)
				return nil
			}
			d.Message = err.Error()
			d.ErrorCode = errorCode(err)
			emitAll(ems, s.apiEm, EventError, d)
			return err
		}
	} else {
		mu.Lock()
		ready := !w.emits || !isInit || w.status == StatusReady
		if ready {
			w.status = StatusReady
		}
		mu.Unlock()
		if ready {
			d := EventDetails{ProviderName: name}
			d.at(domain, named)
			emitAll(emittersOf(emitters), s.apiEm, EventReady, d)
		}
	}

	if !initable || !w.emits {
		attach()
	}

	if c, ok := old.(Closer); ok && closeOld {
		pend, err := c.Close()
		if err != nil {
			if wait != nil {
				go func() { _ = wait() }()
			}
			return nil, err
		}
		if pend != nil {
			go func() {
				if err := <-pend; err != nil {
					s.log().Error(fmt.Sprintf("error closing provider: %s, %s", jsProp(err, "message", err.Error()), jsProp(err, "stack", jsErrorString(err))))
				}
			}()
		}
	}
	return wait, nil
}

func (s *state) setContextOption(o setOptions) {
	if !o.hasCtx {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	if o.domain != "" {
		s.domainCtx[o.domain] = o.ctx.copy()
	} else {
		s.ctx = o.ctx.copy()
	}
}

func setOpts(opts []SetOption) setOptions {
	var o setOptions
	for _, f := range opts {
		f(&o)
	}
	return o
}

func (a *API) begin(p Provider, opts []SetOption) (*state, string, Provider) {
	o := setOpts(opts)
	s := a.current()
	if o.hasDomain && o.domain == "" {
		return s, "", nil
	}
	s.setContextOption(o)
	return s, o.domain, p
}

// SetProvider binds p to the default provider or to a domain. It returns
// only synchronous validation errors; initialization runs in the background.
// Concurrent calls for the same domain from different goroutines are not supported.
func (a *API) SetProvider(p Provider, opts ...SetOption) error {
	s, domain, p := a.begin(p, opts)
	wait, err := s.setAwaitableProvider(domain, p)
	if err != nil || wait == nil {
		return err
	}
	go func() {
		if err := wait(); err != nil {
			s.log().Error("Error during provider initialization:", err)
		}
	}()
	return nil
}

// SetProviderAndWait binds p and waits for its initialization.
// Concurrent calls for the same domain from different goroutines are not supported.
func (a *API) SetProviderAndWait(p Provider, opts ...SetOption) error {
	s, domain, p := a.begin(p, opts)
	wait, err := s.setAwaitableProvider(domain, p)
	if err != nil || wait == nil {
		return err
	}
	return wait()
}

// GetProvider returns the provider of a domain, or the default.
func (a *API) GetProvider(domain ...string) Provider {
	mu.Lock()
	defer mu.Unlock()
	return a.st().wrapperFor(optDomain(domain)).p
}

// ProviderMetadata returns the metadata of a domain's provider, or the default's.
func (a *API) ProviderMetadata(domain ...string) Metadata { return a.GetProvider(domain...).Metadata() }

// ProviderStatus returns the status of a domain's provider, or the default's.
func (a *API) ProviderStatus(domain ...string) Status {
	mu.Lock()
	defer mu.Unlock()
	return a.st().wrapperFor(optDomain(domain)).status
}

// AddHooks adds API-level hooks.
func (a *API) AddHooks(h ...Hook) {
	mu.Lock()
	defer mu.Unlock()
	s := a.st()
	s.hooks = append(append([]Hook(nil), s.hooks...), h...)
}

// Hooks returns the API-level hooks.
func (a *API) Hooks() []Hook {
	mu.Lock()
	defer mu.Unlock()
	return append([]Hook(nil), a.st().hooks...)
}

// ClearHooks removes every API-level hook.
func (a *API) ClearHooks() {
	mu.Lock()
	defer mu.Unlock()
	a.st().hooks = nil
}

// SetLogger sets the API logger. A nil logger prints a TypeError and
// falls back to the default logger, as JavaScript does.
func (a *API) SetLogger(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	a.st().logger = newSafeLogger(l)
}

func (s *state) shutdownError(p Provider, err error) {
	l := s.log()
	if j, ok := asJS(err); ok {
		l.Error(fmt.Sprintf("Error during shutdown of provider %s: %s", p.Metadata().Name, j.JSString()))
		logField(l, j, "stack")
		return
	}
	l.Error(fmt.Sprintf("Error during shutdown of provider %s: %s", p.Metadata().Name, jsErrorString(err)))
	l.Error(jsErrorString(err))
}

func (s *state) close() {
	mu.Lock()
	def := s.def.p
	mu.Unlock()
	if c, ok := def.(Closer); ok {
		pend, err := c.Close()
		if err == nil && pend != nil {
			err = <-pend
		}
		if err != nil {
			s.shutdownError(def, err)
		}
	}
	mu.Lock()
	ws := append([]binding(nil), s.bindings...)
	mu.Unlock()
	var wg sync.WaitGroup
	for _, b := range ws {
		c, ok := b.w.p.(Closer)
		if !ok {
			continue
		}
		pend, err := c.Close()
		if err != nil {
			s.shutdownError(b.w.p, err)
			continue
		}
		if pend != nil {
			wg.Go(func() {
				if err := <-pend; err != nil {
					s.shutdownError(b.w.p, err)
				}
			})
		}
	}
	wg.Wait()
}

func (a *API) reset() {
	mu.Lock()
	defer mu.Unlock()
	s := a.st()
	s.dropForwarding()
	dropState(s)
	a.s = nil
}

// Close shuts every bound provider down and then resets the instance:
// providers, contexts, hooks, handlers and the logger return to their
// initial state.
func (a *API) Close() {
	a.current().close()
	a.reset()
}

// ClearProviders closes and unbinds every provider and clears domain contexts.
func (a *API) ClearProviders() {
	s := a.current()
	s.close()
	mu.Lock()
	defer mu.Unlock()
	s.bindings = nil
	s.def = &wrapper{p: noop, status: StatusNotReady}
	s.dropForwarding()
	dropState(s)
	clear(s.domainCtx)
}

// SetProvider binds p on the default instance. See [API.SetProvider].
func SetProvider(p Provider, opts ...SetOption) error { return global.SetProvider(p, opts...) }

// SetProviderAndWait binds p on the default instance and waits for its initialization.
func SetProviderAndWait(p Provider, opts ...SetOption) error {
	return global.SetProviderAndWait(p, opts...)
}

// GetProvider returns the provider of a domain, or the default.
func GetProvider(domain ...string) Provider { return global.GetProvider(domain...) }

// ProviderMetadata returns the metadata of a domain's provider, or the default's.
func ProviderMetadata(domain ...string) Metadata { return global.ProviderMetadata(domain...) }

// ProviderStatus returns the status of a domain's provider, or the default's.
func ProviderStatus(domain ...string) Status { return global.ProviderStatus(domain...) }

// AddHooks adds API-level hooks.
func AddHooks(h ...Hook) { global.AddHooks(h...) }

// Hooks returns the API-level hooks.
func Hooks() []Hook { return global.Hooks() }

// ClearHooks removes every API-level hook.
func ClearHooks() { global.ClearHooks() }

// SetLogger sets the API logger. See [API.SetLogger].
func SetLogger(l Logger) { global.SetLogger(l) }

// Close shuts every provider down and resets the default instance. See [API.Close].
func Close() { global.Close() }

// ClearProviders closes and unbinds every provider and clears domain contexts.
func ClearProviders() { global.ClearProviders() }

// Reset restores the initial global state without shutting providers down.
// Intended for tests.
func Reset() { global.reset() }
