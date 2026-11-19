package openfeature

// Provider resolves flag values. Provider identity is Go ==, so use
// pointer types: two equal struct values count as one instance.
type Provider interface {
	Metadata() Metadata
	ResolveBoolean(flag string, def bool, ctx Context, log Logger) ResolutionDetails[bool]
	ResolveString(flag string, def string, ctx Context, log Logger) ResolutionDetails[string]
	ResolveNumber(flag string, def float64, ctx Context, log Logger) ResolutionDetails[float64]
	ResolveObject(flag string, def any, ctx Context, log Logger) ResolutionDetails[any]
}

// Pending completes an Initialize, OnContextChange or Close that is still in
// progress: it yields nil on success or the failure.
type Pending <-chan error

// Initializer is implemented by providers that need setup before use. The
// method must have exactly this signature. A nil Pending means initialization
// finished at once; a non-nil Pending means it is still in progress.
type Initializer interface {
	Initialize(ctx Context, domain string) (Pending, error)
}

// ContextChanger is implemented by providers that react to context changes.
// A nil Pending means the change was applied synchronously.
type ContextChanger interface {
	OnContextChange(old, new Context) (Pending, error)
}

// Closer is implemented by providers that release resources on shutdown.
// The method must have exactly this signature; a Close() error method is not
// detected. A non-nil Pending means shutdown is still in progress.
type Closer interface {
	Close() (Pending, error)
}

// Tracker is implemented by providers that record tracking events.
type Tracker interface {
	Track(name string, ctx Context, d TrackingDetails)
}

// Hooked is implemented by providers that contribute hooks.
type Hooked interface {
	Hooks() []Hook
}

// Eventer is implemented by providers that emit events. Events must return
// the same *Emitter on every call.
type Eventer interface {
	Events() *Emitter
}

// LifecycleEmitter is implemented by providers that emit their own lifecycle
// events: ready or error before Initialize completes, and reconciling,
// context-changed or error around OnContextChange. The SDK then raises none
// on their behalf and takes the provider status from those events alone.
// Such a provider must also implement Eventer. Providers without this marker
// use the deprecated path where the SDK raises these events itself.
type LifecycleEmitter interface {
	EmitsLifecycleEvents() bool
}

// RunsOn is implemented by providers that declare their paradigm,
// "client" or "server".
type RunsOn interface {
	RunsOn() string
}

// DomainScoped is implemented by providers that may be bound to one domain only.
type DomainScoped interface {
	DomainScoped() bool
}

type noopProvider struct{}

var noop Provider = noopProvider{}

// NoopProvider returns the provider used before any is set.
func NoopProvider() Provider { return noop }

func (noopProvider) Metadata() Metadata { return Metadata{Name: "No-op Provider"} }

func (noopProvider) ResolveBoolean(_ string, def bool, _ Context, _ Logger) ResolutionDetails[bool] {
	return ResolutionDetails[bool]{Value: def, Reason: "No-op"}
}

func (noopProvider) ResolveString(_ string, def string, _ Context, _ Logger) ResolutionDetails[string] {
	return ResolutionDetails[string]{Value: def, Reason: "No-op"}
}

func (noopProvider) ResolveNumber(_ string, def float64, _ Context, _ Logger) ResolutionDetails[float64] {
	return ResolutionDetails[float64]{Value: def, Reason: "No-op"}
}

func (noopProvider) ResolveObject(_ string, def any, _ Context, _ Logger) ResolutionDetails[any] {
	return ResolutionDetails[any]{Value: def, Reason: "No-op"}
}
