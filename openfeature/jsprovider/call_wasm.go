package jsprovider

import (
	"sync"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
)

var providerEvents = []openfeature.Event{
	openfeature.EventReady, openfeature.EventError, openfeature.EventConfigurationChanged,
	openfeature.EventContextChanged, openfeature.EventReconciling, openfeature.EventStale,
}

const hookKey = "\x00jsprovider.hookContext"

const mapHookData = `return class MapHookData {
	constructor() { this.data = new Map() }
	set(key, value) { this.data.set(key, value) }
	get(key) { return this.data.get(key) }
	has(key) { return this.data.has(key) }
	delete(key) { return this.data.delete(key) }
	clear() { this.data.clear() }
}`

var (
	regMu     sync.Mutex
	ids       app.Value
	providers = map[int]openfeature.Provider{}
	loggers   = map[any]app.Value{}
	shared    openfeature.Logger
	hookData  app.Value
)

type base struct{ *adapter }

type ini struct{ a *adapter }

type trk struct{ a *adapter }

var shapes = [4]func(*adapter) openfeature.Provider{
	func(a *adapter) openfeature.Provider { return base{a} },
	func(a *adapter) openfeature.Provider {
		return struct {
			base
			ini
		}{base{a}, ini{a}}
	},
	func(a *adapter) openfeature.Provider {
		return struct {
			base
			trk
		}{base{a}, trk{a}}
	},
	func(a *adapter) openfeature.Provider {
		return struct {
			base
			ini
			trk
		}{base{a}, ini{a}, trk{a}}
	},
}

// New adapts a JavaScript OpenFeature web provider to openfeature.Provider.
// The same JS object always yields the same Go provider. A value that is not
// a JS object yields nil, which SetProvider ignores as the Web SDK does. The result is an
// openfeature.Initializer or openfeature.Tracker only when provider has an
// initialize or track function, the two members whose absence the SDK can
// observe; every other optional member behaves as absent when JS lacks it.
//
// A JS exception thrown by initialize, a resolver, hook or track call is re-raised as a Go
// panic, which the SDK recovers the way the Web SDK catches the throw. Go event
// handlers run inside the JS provider's emit call, on the JS event loop, so
// they must not block on JavaScript.
func New(provider app.Value) openfeature.Provider {
	if provider.Type() != app.TypeObject {
		return nil
	}
	regMu.Lock()
	defer regMu.Unlock()
	if ids == nil {
		ids = global("WeakMap").New()
	}
	if id := ids.Call("get", provider); !id.IsUndefined() {
		return providers[id.Int()]
	}
	a := &adapter{p: provider}
	if !nullish(provider.Get("events")) {
		a.em = openfeature.NewEmitter()
	}
	p := shapes[shapeOf(provider)](a)
	id := len(providers) + 1
	ids.Call("set", provider, id)
	providers[id] = p
	return p
}

func shapeOf(p app.Value) int {
	bits := 0
	for i, m := range []string{"initialize", "track"} {
		if p.Get(m).Type() == app.TypeFunction {
			bits |= 1 << i
		}
	}
	return bits
}

type adapter struct {
	p    app.Value
	em   *openfeature.Emitter
	mu   sync.Mutex
	subs [6]bool
}

func comparable(v any) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()
	_ = map[any]bool{v: true}
	return true
}

func loggerJS(l openfeature.Logger) app.Value {
	regMu.Lock()
	defer regMu.Unlock()
	var key any = l
	if !comparable(l) {
		key, shared = nil, l
	}
	if o, ok := loggers[key]; ok {
		return o
	}
	o := object()
	for _, level := range []string{"error", "warn", "info", "debug"} {
		o.Set(level, callback(func(args []app.Value) {
			target := l
			if key == nil {
				regMu.Lock()
				target = shared
				regMu.Unlock()
			}
			if target == nil {
				return
			}
			out := make([]any, len(args))
			for i, v := range args {
				out[i] = loggerArg(v)
			}
			switch level {
			case "error":
				target.Error(out...)
			case "warn":
				target.Warn(out...)
			case "info":
				target.Info(out...)
			default:
				target.Debug(out...)
			}
		}))
	}
	loggers[key] = o
	return o
}

func await(v app.Value) openfeature.Pending {
	ch := make(chan error, 1)
	var once sync.Once
	var ok, fail app.Func
	settle := func(err error) {
		once.Do(func() {
			ok.Release()
			fail.Release()
			ch <- err
		})
	}
	ok = app.FuncOf(func(app.Value, []app.Value) any {
		settle(nil)
		return nil
	})
	fail = app.FuncOf(func(_ app.Value, args []app.Value) any {
		reason := app.Undefined()
		if len(args) > 0 {
			reason = args[0]
		}
		settle(goError(reason))
		return nil
	})
	if _, thrown := invoke(v, "then", ok, fail); thrown != nil {
		settle(goError(thrown))
	}
	return ch
}

func settled(res, thrown app.Value) (openfeature.Pending, error) {
	if thrown != nil {
		return nil, goError(thrown)
	}
	if !thenable(res) {
		return nil, nil
	}
	return await(res), nil
}

func (a *adapter) Metadata() openfeature.Metadata {
	var md openfeature.Metadata
	m := a.p.Get("metadata")
	if !isObject(m) {
		return md
	}
	md.Name = optString(m.Get("name"))
	keys := global("Object").Call("keys", m)
	for i := range keys.Length() {
		k := keys.Index(i).String()
		if v := m.Get(k); k != "name" && !nullish(v) {
			if md.Attributes == nil {
				md.Attributes = map[string]string{}
			}
			md.Attributes[k] = jsString(v)
		}
	}
	return md
}

func resolve[T any](a *adapter, method, flag string, def T, ctx openfeature.Context, l openfeature.Logger) openfeature.ResolutionDetails[T] {
	switch fn := a.p.Get(method); {
	case nullish(fn):
		panic(typeError("Cannot read properties of " + jsString(fn) + " (reading 'call')"))
	case fn.Type() != app.TypeFunction:
		panic(typeError("resolver.call is not a function"))
	}
	r, thrown := invoke(a.p, method, flag, toJS(def), contextJS(ctx), loggerJS(l))
	if thrown != nil {
		panic(goError(thrown))
	}
	if nullish(r) {
		panic(typeError("Cannot read properties of " + jsString(r) + " (reading 'flagMetadata')"))
	}
	var out openfeature.ResolutionDetails[T]
	jv := r.Get("value")
	v, typed := fromJS(jv).(T)
	if !typed && jv.IsNull() {
		_, typed = any(&v).(*any)
	}
	out.Value = v
	out.Variant = optString(r.Get("variant"))
	out.Reason = openfeature.Reason(optString(r.Get("reason")))
	if c := r.Get("errorCode"); c.Truthy() {
		out.ErrorCode = openfeature.ErrorCode(jsString(c))
	}
	out.ErrorMessage = optString(r.Get("errorMessage"))
	if !typed && out.ErrorCode == "" {
		out.Value, out.ErrorCode, out.ErrorMessage = def, openfeature.CodeTypeMismatch, ""
	}
	if md, ok := fromJSMap(r.Get("flagMetadata")); ok {
		out.FlagMetadata = md
	}
	return out
}

func (a *adapter) ResolveBoolean(flag string, def bool, ctx openfeature.Context, l openfeature.Logger) openfeature.ResolutionDetails[bool] {
	return resolve(a, "resolveBooleanEvaluation", flag, def, ctx, l)
}

func (a *adapter) ResolveString(flag string, def string, ctx openfeature.Context, l openfeature.Logger) openfeature.ResolutionDetails[string] {
	return resolve(a, "resolveStringEvaluation", flag, def, ctx, l)
}

func (a *adapter) ResolveNumber(flag string, def float64, ctx openfeature.Context, l openfeature.Logger) openfeature.ResolutionDetails[float64] {
	return resolve(a, "resolveNumberEvaluation", flag, def, ctx, l)
}

func (a *adapter) ResolveObject(flag string, def any, ctx openfeature.Context, l openfeature.Logger) openfeature.ResolutionDetails[any] {
	return resolve(a, "resolveObjectEvaluation", flag, def, ctx, l)
}

func (x ini) Initialize(ctx openfeature.Context, domain string) (openfeature.Pending, error) {
	var d any = app.Undefined()
	if domain != "" {
		d = domain
	}
	pend, err := settled(invoke(x.a.p, "initialize", contextJS(ctx), d))
	if err != nil {
		panic(err)
	}
	return pend, nil
}

func (a *adapter) OnContextChange(old, new openfeature.Context) (openfeature.Pending, error) {
	if a.p.Get("onContextChange").Type() != app.TypeFunction {
		return nil, nil
	}
	return settled(invoke(a.p, "onContextChange", contextJS(old), contextJS(new)))
}

func (a *adapter) Close() (openfeature.Pending, error) {
	if nullish(a.p.Get("onClose")) {
		return nil, nil
	}
	return settled(invoke(a.p, "onClose"))
}

func (x trk) Track(name string, ctx openfeature.Context, d openfeature.TrackingDetails) {
	details := toJS(d.Attributes)
	if !isObject(details) {
		details = object()
	}
	if d.Value != nil {
		details.Set("value", *d.Value)
	}
	c := global("Object").Call("freeze", contextJS(ctx))
	if _, thrown := invoke(x.a.p, "track", name, c, details); thrown != nil {
		panic(goError(thrown))
	}
}

func (a *adapter) RunsOn() string {
	if v := a.p.Get("runsOn"); v.Truthy() {
		return jsString(v)
	}
	return ""
}

func (a *adapter) DomainScoped() bool { return a.p.Get("domainScoped").Truthy() }

func (a *adapter) Events() *openfeature.Emitter {
	if a.em == nil {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	events := a.p.Get("events")
	for i, ev := range providerEvents {
		if a.subs[i] {
			continue
		}
		f := callback(func(args []app.Value) {
			d := app.Undefined()
			if len(args) > 0 {
				d = args[0]
			}
			a.em.Emit(ev, eventDetails(d))
		})
		if _, thrown := invoke(events, "addHandler", string(ev), f); thrown != nil {
			f.Release()
			panic(goError(thrown))
		}
		a.subs[i] = true
	}
	return a.em
}

func eventDetails(v app.Value) openfeature.EventDetails {
	var d openfeature.EventDetails
	if !isObject(v) {
		return d
	}
	d.Message = optString(v.Get("message"))
	d.ErrorCode = openfeature.ErrorCode(optString(v.Get("errorCode")))
	if fc, ok := fromJS(v.Get("flagsChanged")).([]any); ok {
		d.FlagsChanged = make([]string, 0, len(fc))
		for i := range fc {
			d.FlagsChanged = append(d.FlagsChanged, jsString(v.Get("flagsChanged").Index(i)))
		}
	}
	if md, ok := fromJSMap(v.Get("metadata")); ok {
		d.Metadata = md
	}
	return d
}

func (a *adapter) Hooks() []openfeature.Hook {
	hs := a.p.Get("hooks")
	if !global("Array").Call("isArray", hs).Bool() {
		return nil
	}
	out := make([]openfeature.Hook, hs.Length())
	for i := range out {
		out[i] = a.hook(hs.Index(i))
	}
	return out
}

func (a *adapter) hook(h app.Value) openfeature.Hook {
	var out openfeature.Hook
	if !isObject(h) {
		return out
	}
	has := func(stage string) bool { return !nullish(h.Get(stage)) }
	if has("before") {
		out.Before = func(hc openfeature.HookContext, hints openfeature.HookHints) error {
			a.stage(h, "before", hc, nil, hints)
			return nil
		}
	}
	if has("after") {
		out.After = func(hc openfeature.HookContext, d openfeature.EvaluationDetails[any], hints openfeature.HookHints) error {
			a.stage(h, "after", hc, detailsJS(d), hints)
			return nil
		}
	}
	if has("error") {
		out.Error = func(hc openfeature.HookContext, err error, hints openfeature.HookHints) {
			a.stage(h, "error", hc, errorJS(err), hints)
		}
	}
	if has("finally") {
		out.Finally = func(hc openfeature.HookContext, d openfeature.EvaluationDetails[any], hints openfeature.HookHints) {
			a.stage(h, "finally", hc, detailsJS(d), hints)
		}
	}
	return out
}

func (a *adapter) stage(h app.Value, name string, hc openfeature.HookContext, arg app.Value, hints openfeature.HookHints) {
	args := []any{a.hookContext(hc)}
	if arg != nil {
		args = append(args, arg)
	}
	hv := app.Undefined()
	if m := hints.Map(); m != nil {
		hv = global("Object").Call("freeze", toJS(m))
	}
	args = append(args, hv)
	if _, thrown := invoke(h, name, args...); thrown != nil {
		panic(goError(thrown))
	}
}

func newHookData() app.Value {
	regMu.Lock()
	defer regMu.Unlock()
	if hookData == nil {
		hookData = global("Function").New(mapHookData).Invoke()
	}
	return hookData.New()
}

func (a *adapter) hookContext(hc openfeature.HookContext) app.Value {
	if hc.HookData != nil {
		if v, ok := hc.HookData.Get(hookKey).(app.Value); ok {
			return v
		}
	}
	ctx := global("Object").Call("freeze", contextJS(hc.Context))
	pm := a.p.Get("metadata")
	cm := object()
	if hc.ClientMetadata.HasDomain {
		cm.Set("name", hc.ClientMetadata.Name)
		cm.Set("domain", hc.ClientMetadata.Domain)
	}
	if hc.ClientMetadata.Version != "" {
		cm.Set("version", hc.ClientMetadata.Version)
	}
	cm.Set("sdk", hc.ClientMetadata.SDK)
	cm.Set("paradigm", hc.ClientMetadata.Paradigm)
	cm.Set("providerMetadata", pm)
	o := object()
	o.Set("flagKey", hc.FlagKey)
	o.Set("defaultValue", toJS(hc.DefaultValue))
	o.Set("flagValueType", string(hc.FlagType))
	o.Set("clientMetadata", cm)
	o.Set("providerMetadata", pm)
	o.Set("context", ctx)
	o.Set("logger", loggerJS(hc.Logger))
	o.Set("hookData", newHookData())
	global("Object").Call("freeze", o)
	if hc.HookData != nil {
		hc.HookData.Set(hookKey, o)
	}
	return o
}

func detailsJS(d openfeature.EvaluationDetails[any]) app.Value {
	o := object()
	o.Set("value", toJS(d.Value))
	for _, f := range [][2]string{{"variant", d.Variant}, {"reason", string(d.Reason)}, {"errorCode", string(d.ErrorCode)}} {
		if f[1] != "" {
			o.Set(f[0], f[1])
		}
	}
	if d.ErrorCode != "" || d.ErrorMessage != "" {
		o.Set("errorMessage", d.ErrorMessage)
	}
	md := toJS(map[string]any(d.FlagMetadata))
	if !isObject(md) {
		md = object()
	}
	o.Set("flagMetadata", global("Object").Call("freeze", md))
	o.Set("flagKey", d.FlagKey)
	return o
}
