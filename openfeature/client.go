package openfeature

import "math"

// Client evaluates flags for one domain.
type Client struct {
	api     *API
	domain  string
	named   bool
	version string
	hooks   []Hook
	logger  Logger
}

// GetClient returns a client for a domain, or the default client.
func (a *API) GetClient(domain ...string) *Client {
	return &Client{api: a, domain: optDomain(domain), named: len(domain) > 0}
}

// GetClientVersion returns a client with a version for a domain, or the default client.
func (a *API) GetClientVersion(version string, domain ...string) *Client {
	c := a.GetClient(domain...)
	c.version = version
	return c
}

// GetClient returns a client of the default instance for a domain, or its default client.
func GetClient(domain ...string) *Client { return global.GetClient(domain...) }

// GetClientVersion returns a client of the default instance with a version.
func GetClientVersion(version string, domain ...string) *Client {
	return global.GetClientVersion(version, domain...)
}

func (c *Client) log(s *state) Logger {
	if c.logger != nil {
		return c.logger
	}
	return s.logger
}

func (c *Client) logNow(s *state) Logger {
	mu.Lock()
	defer mu.Unlock()
	return c.log(s)
}

// Metadata returns the client metadata.
func (c *Client) Metadata() ClientMetadata {
	return ClientMetadata{
		Domain:           c.domain,
		Name:             c.domain,
		Version:          c.version,
		SDK:              "webgo",
		Paradigm:         "client",
		ProviderMetadata: c.api.GetProvider(c.domain).Metadata(),
		HasDomain:        c.named,
	}
}

// ProviderStatus returns the status of the client's provider.
func (c *Client) ProviderStatus() Status { return c.api.ProviderStatus(c.domain) }

// AddHandler registers a handler for the client's domain and returns its
// remover. The handler runs at once if the provider's status matches e.
func (c *Client) AddHandler(e Event, h Handler) func() {
	s := c.api.current()
	mu.Lock()
	em := s.emitterFor(c.domain, c.named)
	mu.Unlock()
	id := em.on(e, h, nil)
	mu.Lock()
	w := s.wrapperFor(c.domain)
	due := statusMatchesEvent(e, w.status)
	l := c.log(s)
	mu.Unlock()
	if due {
		d := EventDetails{ProviderName: w.p.Metadata().Name}
		d.at(c.domain, c.named)
		s.runHandler(l, h, d)
	}
	return func() { em.off(id) }
}

// Handlers returns the client's handlers for e.
func (c *Client) Handlers(e Event) []Handler {
	s := c.api.current()
	mu.Lock()
	em := s.emitterFor(c.domain, c.named)
	mu.Unlock()
	return em.Handlers(e)
}

// SetLogger sets the client logger. A nil logger prints a TypeError and
// falls back to the default logger, as JavaScript does.
func (c *Client) SetLogger(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	c.logger = newSafeLogger(l)
}

// AddHooks adds client-level hooks.
func (c *Client) AddHooks(h ...Hook) {
	mu.Lock()
	defer mu.Unlock()
	c.hooks = append(append([]Hook(nil), c.hooks...), h...)
}

// Hooks returns the client-level hooks.
func (c *Client) Hooks() []Hook {
	mu.Lock()
	defer mu.Unlock()
	return append([]Hook(nil), c.hooks...)
}

// ClearHooks removes every client-level hook.
func (c *Client) ClearHooks() {
	mu.Lock()
	defer mu.Unlock()
	c.hooks = nil
}

// Boolean evaluates a boolean flag.
func (c *Client) Boolean(flag string, def bool, opts ...EvalOption) bool {
	return c.BooleanDetails(flag, def, opts...).Value
}

// BooleanDetails evaluates a boolean flag with details.
func (c *Client) BooleanDetails(flag string, def bool, opts ...EvalOption) EvaluationDetails[bool] {
	return evaluate(c, flag, def, TypeBoolean, opts, func(p Provider, ctx Context, l Logger) ResolutionDetails[bool] {
		return p.ResolveBoolean(flag, def, ctx, l)
	})
}

// String evaluates a string flag.
func (c *Client) String(flag string, def string, opts ...EvalOption) string {
	return c.StringDetails(flag, def, opts...).Value
}

// StringDetails evaluates a string flag with details.
func (c *Client) StringDetails(flag string, def string, opts ...EvalOption) EvaluationDetails[string] {
	return evaluate(c, flag, def, TypeString, opts, func(p Provider, ctx Context, l Logger) ResolutionDetails[string] {
		return p.ResolveString(flag, def, ctx, l)
	})
}

// Number evaluates a number flag.
func (c *Client) Number(flag string, def float64, opts ...EvalOption) float64 {
	return c.NumberDetails(flag, def, opts...).Value
}

// NumberDetails evaluates a number flag with details.
func (c *Client) NumberDetails(flag string, def float64, opts ...EvalOption) EvaluationDetails[float64] {
	return evaluate(c, flag, def, TypeNumber, opts, func(p Provider, ctx Context, l Logger) ResolutionDetails[float64] {
		return p.ResolveNumber(flag, def, ctx, l)
	})
}

// Int evaluates a number flag that must hold an integer.
func (c *Client) Int(flag string, def int64, opts ...EvalOption) int64 {
	return c.IntDetails(flag, def, opts...).Value
}

// IntDetails evaluates a number flag that must hold an integer, with details.
// A non-integral or out-of-range number is a TYPE_MISMATCH error.
func (c *Client) IntDetails(flag string, def int64, opts ...EvalOption) EvaluationDetails[int64] {
	return evaluate(c, flag, def, TypeNumber, opts, func(p Provider, ctx Context, l Logger) ResolutionDetails[int64] {
		r := p.ResolveNumber(flag, float64(def), ctx, l)
		out := ResolutionDetails[int64]{Variant: r.Variant, Reason: r.Reason, ErrorCode: r.ErrorCode, ErrorMessage: r.ErrorMessage, FlagMetadata: r.FlagMetadata}
		switch {
		case r.ErrorCode != "":
		case r.Value != math.Trunc(r.Value) || r.Value < -(1<<63) || r.Value >= 1<<63:
			out.ErrorCode, out.ErrorMessage = CodeTypeMismatch, "value is not an integer"
		default:
			out.Value = int64(r.Value)
		}
		return out
	})
}

// Object evaluates an object flag.
func (c *Client) Object(flag string, def any, opts ...EvalOption) any {
	return c.ObjectDetails(flag, def, opts...).Value
}

// ObjectDetails evaluates an object flag with details.
func (c *Client) ObjectDetails(flag string, def any, opts ...EvalOption) EvaluationDetails[any] {
	return evaluate(c, flag, def, TypeObject, opts, func(p Provider, ctx Context, l Logger) ResolutionDetails[any] {
		return p.ResolveObject(flag, def, ctx, l)
	})
}

// Track records a tracking event with the provider.
func (c *Client) Track(name string, d ...TrackingDetails) {
	var td TrackingDetails
	if len(d) > 0 {
		td = d[0]
	}
	s := c.api.current()
	err := catch(func() error {
		switch c.api.ProviderStatus(c.domain) {
		case StatusNotReady:
			return NewError(CodeProviderNotReady, "provider has not yet initialized")
		case StatusFatal:
			return NewError(CodeProviderFatal, "provider is in an irrecoverable error state")
		}
		mu.Lock()
		p := s.wrapperFor(c.domain).p
		mu.Unlock()
		t, ok := p.(Tracker)
		if !ok {
			c.logNow(s).Debug("Provider does not support the track function; will no-op.")
			return nil
		}
		ctx := c.api.GetContext(c.domain)
		t.Track(name, ctx, td)
		return nil
	})
	if err != nil {
		c.logNow(s).Debug("Error recording tracking event.", err)
	}
}
