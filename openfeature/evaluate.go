package openfeature

import (
	"fmt"
	"maps"
)

func anyDetails[T any](d EvaluationDetails[T]) EvaluationDetails[any] {
	r := d.ResolutionDetails
	return EvaluationDetails[any]{FlagKey: d.FlagKey, ResolutionDetails: ResolutionDetails[any]{
		Value: r.Value, Variant: r.Variant, Reason: r.Reason,
		ErrorCode: r.ErrorCode, ErrorMessage: r.ErrorMessage, FlagMetadata: r.FlagMetadata,
	}}
}

func errorDetails[T any](flag string, def T, err error, md FlagMetadata) EvaluationDetails[T] {
	return EvaluationDetails[T]{FlagKey: flag, ResolutionDetails: ResolutionDetails[T]{
		Value: def, Reason: ReasonError, ErrorCode: errorCode(err), ErrorMessage: errorMessage(err), FlagMetadata: md,
	}}
}

func errorCode(err error) ErrorCode {
	if c := codeOf(err); c != "" {
		return c
	}
	return CodeGeneral
}

func knownCode(c ErrorCode) ErrorCode {
	if _, ok := errorClasses[c]; ok {
		return c
	}
	return CodeGeneral
}

func catch(f func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = panicError(r)
		}
	}()
	return f()
}

func unhandled(logger func() Logger, stage string, f func()) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		l := logger()
		if j, ok := asJS(r); ok {
			l.Error(fmt.Sprintf("Unhandled error during '%s' hook: %s", stage, j.JSString()))
			if j.JSIsError() {
				logField(l, j, "stack")
			}
			logField(l, j, "stack")
			return
		}
		err, isErr := r.(error)
		if isErr {
			l.Error(fmt.Sprintf("Unhandled error during '%s' hook: %s", stage, jsErrorString(err)))
			l.Error(jsErrorString(err))
			l.Error(jsErrorString(err))
			return
		}
		l.Error(fmt.Sprintf("Unhandled error during '%s' hook: %v", stage, r))
		l.Error()
	}()
	f()
}

func evaluate[T any](c *Client, flag string, def T, typ FlagType, opts []EvalOption, resolve func(Provider, Context, Logger) ResolutionDetails[T]) EvaluationDetails[T] {
	var o evalOptions
	for _, f := range opts {
		f(&o)
	}
	s := c.api.current()
	mu.Lock()
	p := s.wrapperFor(c.domain).p
	hooks := append(append([]Hook(nil), s.hooks...), c.hooks...)
	ctx, missing := s.lookupContext(c.domain)
	ctx = ctx.copy()
	l, apiLog := c.log(s), s.logger
	mu.Unlock()
	hooks = append(hooks, o.hooks...)
	if h, ok := p.(Hooked); ok {
		hooks = append(hooks, h.Hooks()...)
	}
	if missing {
		missingContext(apiLog, c.domain)
	}
	hcs := make([]HookContext, len(hooks))
	for i := range hooks {
		hcs[i] = HookContext{
			FlagKey: flag, DefaultValue: def, FlagType: typ, Context: ctx.copy(),
			ClientMetadata: c.Metadata(), ProviderMetadata: p.Metadata(), Logger: l, HookData: NewHookData(),
		}
	}
	errorHooks := func(err error) {
		for i := len(hooks) - 1; i >= 0; i-- {
			if h := hooks[i].Error; h != nil {
				unhandled(func() Logger { return c.logNow(s) }, "error", func() { h(hcs[i], err, o.hints) })
			}
		}
	}

	var details EvaluationDetails[T]
	err := catch(func() error {
		for i, h := range hooks {
			if h.Before != nil {
				if err := h.Before(hcs[i], o.hints); err != nil {
					return err
				}
			}
		}
		switch c.api.ProviderStatus(c.domain) {
		case StatusNotReady:
			return NewError(CodeProviderNotReady, "provider has not yet initialized")
		case StatusFatal:
			return NewError(CodeProviderFatal, "provider is in an irrecoverable error state")
		}
		mu.Lock()
		rp := s.wrapperFor(c.domain).p
		mu.Unlock()
		res := resolve(rp, ctx, c.logNow(s))
		res.FlagMetadata = maps.Clone(res.FlagMetadata)
		if res.FlagMetadata == nil {
			res.FlagMetadata = FlagMetadata{}
		}
		rd := EvaluationDetails[T]{FlagKey: flag, ResolutionDetails: res}
		if res.ErrorCode != "" {
			err := NewError(knownCode(res.ErrorCode), res.ErrorMessage)
			errorHooks(err)
			details = errorDetails(flag, def, err, res.FlagMetadata)
			return nil
		}
		ad := anyDetails(rd)
		for i := len(hooks) - 1; i >= 0; i-- {
			if h := hooks[i].After; h != nil {
				if err := h(hcs[i], ad, o.hints); err != nil {
					return err
				}
			}
		}
		details = rd
		return nil
	})
	if err != nil {
		errorHooks(err)
		details = errorDetails(flag, def, err, FlagMetadata{})
	}
	ad := anyDetails(details)
	for i := len(hooks) - 1; i >= 0; i-- {
		if h := hooks[i].Finally; h != nil {
			unhandled(func() Logger { return c.logNow(s) }, "finally", func() { h(hcs[i], ad, o.hints) })
		}
	}
	return details
}
