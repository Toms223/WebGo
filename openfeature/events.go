package openfeature

func (s *state) clientEmitter(domain string, named bool) *clientEmitter {
	for _, c := range s.clientEms {
		if c.domain == domain && c.named == named {
			return c
		}
	}
	return nil
}

func (s *state) emitterFor(domain string, named bool) *Emitter {
	if c := s.clientEmitter(domain, named); c != nil {
		return c.em
	}
	c := &clientEmitter{domain: domain, named: named, em: NewEmitter()}
	s.clientEms = append(s.clientEms, c)
	w := s.wrapperFor(domain)
	if w.ev != nil {
		p, em := w.p, c.em
		c.fwd = registration{src: w.ev, ids: w.ev.onEach(providerEvents, func(ev Event, d EventDetails) {
			d.at(domain, named)
			d.ProviderName = p.Metadata().Name
			em.Emit(ev, d)
		})}
	}
	return c.em
}

func (s *state) unboundEmitters() []*clientEmitter {
	var out []*clientEmitter
	var def *clientEmitter
	for _, c := range s.clientEms {
		switch {
		case !c.named:
			def = c
		case s.bound(c.domain) == nil:
			out = append(out, c)
		}
	}
	if def != nil {
		out = append(out, def)
	}
	return out
}

func (s *state) associatedEmitters(domain string) []*clientEmitter {
	if domain != "" {
		s.emitterFor(domain, true)
		return []*clientEmitter{s.clientEmitter(domain, true)}
	}
	return s.unboundEmitters()
}

func emittersOf(cs []*clientEmitter) []*Emitter {
	out := make([]*Emitter, len(cs))
	for i, c := range cs {
		out[i] = c.em
	}
	return out
}

func (s *state) transferListeners(p Provider, ev *Emitter, domain string, emitters []*clientEmitter) {
	if r, ok := s.transfers[domain]; ok {
		r.drop()
	}
	for _, c := range emitters {
		c.fwd.drop()
	}
	if ev == nil {
		delete(s.transfers, domain)
		return
	}
	ems, apiEm := emittersOf(emitters), s.apiEm
	ids := ev.onEach(providerEvents, func(e Event, d EventDetails) {
		d.at(domain, domain != "")
		d.ProviderName = p.Metadata().Name
		emitAll(ems, apiEm, e, d)
	})
	s.transfers[domain] = &registration{src: ev, ids: ids}
}

func (s *state) dropForwarding() {
	for _, r := range s.transfers {
		r.drop()
	}
	clear(s.transfers)
	for _, c := range s.clientEms {
		c.fwd.drop()
	}
}

func (s *state) emitFor(domain string, ev Event, d EventDetails) {
	mu.Lock()
	ems := emittersOf(s.associatedEmitters(domain))
	mu.Unlock()
	emitAll(ems, s.apiEm, ev, d)
}

func (s *state) runHandler(l Logger, h Handler, d EventDetails) {
	defer func() {
		if r := recover(); r != nil {
			l.Error("Error running event handler:", r)
		}
	}()
	h(d)
}

// AddHandler registers an API-level handler and returns its remover. The
// handler runs at once for every provider whose status matches e.
func (a *API) AddHandler(e Event, h Handler) func() {
	s := a.current()
	mu.Lock()
	ws := append([]binding{{"", s.def}}, s.bindings...)
	mu.Unlock()
	for _, b := range ws {
		mu.Lock()
		due, l := statusMatchesEvent(e, b.w.status), s.logger
		mu.Unlock()
		if due {
			s.runHandler(l, h, EventDetails{Domain: b.domain, HasDomain: b.domain != "", ProviderName: b.w.p.Metadata().Name})
		}
	}
	id := s.apiEm.on(e, h, nil)
	return func() { s.apiEm.off(id) }
}

// Handlers returns the API-level handlers for e.
func (a *API) Handlers(e Event) []Handler { return a.current().apiEm.Handlers(e) }

// ClearHandlers removes every API-level handler.
func (a *API) ClearHandlers() { a.current().apiEm.RemoveAllHandlers() }

// AddHandler registers an API-level handler and returns its remover. See [API.AddHandler].
func AddHandler(e Event, h Handler) func() { return global.AddHandler(e, h) }

// Handlers returns the API-level handlers for e.
func Handlers(e Event) []Handler { return global.Handlers(e) }

// ClearHandlers removes every API-level handler.
func ClearHandlers() { global.ClearHandlers() }
