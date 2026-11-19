package openfeature

import (
	"fmt"
	"sync"
)

func (s *state) lookupContext(domain string) (Context, bool) {
	if domain == "" {
		return s.ctx, false
	}
	if c, ok := s.domainCtx[domain]; ok {
		return c, false
	}
	return s.ctx, true
}

func missingContext(l Logger, domain string) {
	l.Debug(fmt.Sprintf("Unable to find context for '%s'.", domain))
}

func (s *state) reconcile(domain string, w *wrapper, old, new Context) func() {
	name := w.p.Metadata().Name
	if name == "" {
		name = "unnamed-provider"
	}
	settle := func(err error) {
		mu.Lock()
		if err == nil {
			w.status = StatusReady
		} else {
			w.status = StatusError
		}
		settled := w.pending == 0
		mu.Unlock()
		if !settled {
			return
		}
		d := EventDetails{ProviderName: name}
		d.at(domain, domain != "")
		if err == nil {
			s.emitFor(domain, EventContextChanged, d)
			return
		}
		d.Message = fmt.Sprintf("Error running %s's context change handler: %s", name, jsProp(err, "message", err.Error()))
		d.ErrorCode = errorCode(err)
		s.log().Error(d.Message, err)
		s.emitFor(domain, EventError, d)
	}
	cc, ok := w.p.(ContextChanger)
	if w.emits {
		if !ok {
			return nil
		}
		pend, err := cc.OnContextChange(old.copy(), new.copy())
		if err != nil || pend == nil {
			return nil
		}
		return func() { <-pend }
	}
	if !ok {
		settle(nil)
		return nil
	}
	pend, err := cc.OnContextChange(old.copy(), new.copy())
	if err != nil || pend == nil {
		settle(err)
		return nil
	}
	mu.Lock()
	w.pending++
	w.status = StatusReconciling
	mu.Unlock()
	s.emitFor(domain, EventReconciling, EventDetails{Domain: domain, HasDomain: domain != "", ProviderName: name})
	return func() {
		err := <-pend
		mu.Lock()
		w.pending--
		mu.Unlock()
		settle(err)
	}
}

func waitAll(waits []func()) {
	var wg sync.WaitGroup
	for _, w := range waits {
		if w != nil {
			wg.Go(w)
		}
	}
	wg.Wait()
}

// SetContext sets the global or a domain context and waits for providers to reconcile.
func (a *API) SetContext(ctx Context, domain ...string) {
	s := a.current()
	c := ctx.copy()
	d := optDomain(domain)
	mu.Lock()
	if d != "" {
		w := s.bound(d)
		old, missing := s.lookupContext(d)
		s.domainCtx[d] = c
		l := s.logger
		mu.Unlock()
		if w != nil {
			if missing {
				missingContext(l, d)
			}
			waitAll([]func(){s.reconcile(d, w, old, c)})
		}
		return
	}
	old := s.ctx
	s.ctx = c
	recs := []binding{{"", s.def}}
	for _, b := range s.bindings {
		if _, ok := s.domainCtx[b.domain]; !ok {
			recs = append(recs, b)
		}
	}
	mu.Unlock()
	var waits []func()
	for _, b := range recs {
		waits = append(waits, s.reconcile(b.domain, b.w, old, c))
	}
	waitAll(waits)
}

// GetContext returns the global or a domain context.
func (a *API) GetContext(domain ...string) Context {
	d := optDomain(domain)
	mu.Lock()
	c, missing := a.st().lookupContext(d)
	l := a.st().logger
	mu.Unlock()
	if missing {
		missingContext(l, d)
	}
	return c.copy()
}

func (s *state) clearDomain(d string) func() {
	mu.Lock()
	w := s.bound(d)
	old, missing := s.lookupContext(d)
	delete(s.domainCtx, d)
	c := s.ctx
	l := s.logger
	mu.Unlock()
	if w == nil {
		return nil
	}
	if missing {
		missingContext(l, d)
	}
	return s.reconcile(d, w, old, c)
}

// ClearContext removes the global or a domain context.
func (a *API) ClearContext(domain ...string) {
	if d := optDomain(domain); d != "" {
		waitAll([]func(){a.current().clearDomain(d)})
		return
	}
	a.SetContext(Context{})
}

// ClearContexts removes the global and every domain context.
func (a *API) ClearContexts() {
	a.ClearContext()
	s := a.current()
	mu.Lock()
	ds := make([]string, len(s.bindings))
	for i, b := range s.bindings {
		ds[i] = b.domain
	}
	mu.Unlock()
	var waits []func()
	for _, d := range ds {
		waits = append(waits, s.clearDomain(d))
	}
	waitAll(waits)
}

// SetContext sets the global or a domain context and waits for providers to reconcile.
func SetContext(ctx Context, domain ...string) { global.SetContext(ctx, domain...) }

// GetContext returns the global or a domain context.
func GetContext(domain ...string) Context { return global.GetContext(domain...) }

// ClearContext removes the global or a domain context.
func ClearContext(domain ...string) { global.ClearContext(domain...) }

// ClearContexts removes the global and every domain context.
func ClearContexts() { global.ClearContexts() }
