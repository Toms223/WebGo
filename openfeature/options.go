package openfeature

type setOptions struct {
	domain    string
	hasDomain bool
	ctx       Context
	hasCtx    bool
}

// SetOption configures SetProvider and SetProviderAndWait.
type SetOption func(*setOptions)

// ForDomain scopes the call to a domain.
func ForDomain(domain string) SetOption {
	return func(o *setOptions) { o.domain, o.hasDomain = domain, true }
}

// WithContext supplies a context for the call.
func WithContext(c Context) SetOption {
	return func(o *setOptions) { o.ctx, o.hasCtx = c, true }
}

type evalOptions struct {
	hooks []Hook
	hints HookHints
}

// EvalOption configures one evaluation.
type EvalOption func(*evalOptions)

// WithHooks adds hooks for one evaluation.
func WithHooks(h ...Hook) EvalOption {
	return func(o *evalOptions) { o.hooks = append(o.hooks, h...) }
}

// WithHookHints sets hints for one evaluation.
func WithHookHints(h HookHints) EvalOption {
	return func(o *evalOptions) { o.hints = h }
}
