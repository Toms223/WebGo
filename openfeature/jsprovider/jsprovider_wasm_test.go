package jsprovider

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
)

func evalJS(src string) app.Value {
	return app.Window().Call("eval", "("+src+")")
}

const resolvers = `
	resolveBooleanEvaluation(k, d) { return { value: d } },
	resolveStringEvaluation(k, d) { return { value: d } },
	resolveNumberEvaluation(k, d) { return { value: d } },
	resolveObjectEvaluation(k, d) { return { value: d } },`

func wait(t *testing.T, p openfeature.Pending) error {
	t.Helper()
	if p == nil {
		t.Fatal("nil Pending")
	}
	select {
	case err := <-p:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("Pending never settled")
		return nil
	}
}

func useDomain(t *testing.T, domain string, p openfeature.Provider) *openfeature.Client {
	t.Helper()
	t.Cleanup(openfeature.Reset)
	if err := openfeature.SetProviderAndWait(p, openfeature.ForDomain(domain)); err != nil {
		t.Fatal(err)
	}
	return openfeature.GetClient(domain)
}

type logs struct{ lines [][]any }

func (l *logs) Error(a ...any) { l.lines = append(l.lines, append([]any{"error"}, a...)) }
func (l *logs) Warn(a ...any)  { l.lines = append(l.lines, append([]any{"warn"}, a...)) }
func (l *logs) Info(a ...any)  { l.lines = append(l.lines, append([]any{"info"}, a...)) }
func (l *logs) Debug(a ...any) { l.lines = append(l.lines, append([]any{"debug"}, a...)) }

func TestResolveConvertsDetails(t *testing.T) {
	p := New(evalJS(`{
		metadata: { name: "t" },
		resolveBooleanEvaluation() { return { value: true, variant: "on", reason: "STATIC", flagMetadata: { a: 1, b: "x", c: false } } },
		resolveStringEvaluation() { return { value: "v" } },
		resolveNumberEvaluation() { return { value: 1.5, errorCode: "PARSE_ERROR", errorMessage: "bad" } },
		resolveObjectEvaluation() { return { value: { x: [1, "s", null, undefined], d: new Date(0), u: undefined, n: null } } },
	}`))
	b := p.ResolveBoolean("f", false, openfeature.Context{}, openfeature.DefaultLogger())
	want := openfeature.ResolutionDetails[bool]{Value: true, Variant: "on", Reason: "STATIC", FlagMetadata: openfeature.FlagMetadata{"a": 1.0, "b": "x", "c": false}}
	if !reflect.DeepEqual(b, want) {
		t.Errorf("boolean = %#v, want %#v", b, want)
	}
	if s := p.ResolveString("f", "d", openfeature.Context{}, openfeature.DefaultLogger()); !reflect.DeepEqual(s, openfeature.ResolutionDetails[string]{Value: "v"}) {
		t.Errorf("string = %#v, want only the value", s)
	}
	n := p.ResolveNumber("f", 0, openfeature.Context{}, openfeature.DefaultLogger())
	if n.Value != 1.5 || n.ErrorCode != openfeature.CodeParseError || n.ErrorMessage != "bad" {
		t.Errorf("number = %#v", n)
	}
	o := p.ResolveObject("f", nil, openfeature.Context{}, openfeature.DefaultLogger())
	wantObj := openfeature.Object{{Key: "x", Value: []any{1.0, "s", nil, nil}}, {Key: "d", Value: time.UnixMilli(0).UTC()}, {Key: "n", Value: nil}}
	if !reflect.DeepEqual(o.Value, wantObj) {
		t.Errorf("object = %#v, want %#v", o.Value, wantObj)
	}
}

func TestResolveWrongTypeIsTypeMismatch(t *testing.T) {
	p := New(evalJS(`{
		resolveBooleanEvaluation() { return { value: "yes", errorMessage: "m" } },
		resolveStringEvaluation() { return {} },
		resolveObjectEvaluation(k) { return k === "null" ? { value: null } : {} },
	}`))
	l := openfeature.DefaultLogger()
	if r := p.ResolveBoolean("f", true, openfeature.Context{}, l); !r.Value || r.ErrorCode != openfeature.CodeTypeMismatch || r.ErrorMessage != "" {
		t.Errorf("wrong type = %#v, want default and TYPE_MISMATCH", r)
	}
	if r := p.ResolveString("f", "d", openfeature.Context{}, l); r.Value != "d" || r.ErrorCode != openfeature.CodeTypeMismatch {
		t.Errorf("missing value = %#v, want default and TYPE_MISMATCH", r)
	}
	if r := p.ResolveObject("f", "d", openfeature.Context{}, l); r.Value != "d" || r.ErrorCode != openfeature.CodeTypeMismatch {
		t.Errorf("missing object = %#v, want default and TYPE_MISMATCH", r)
	}
	if r := p.ResolveObject("null", "d", openfeature.Context{}, l); r.Value != nil || r.ErrorCode != "" {
		t.Errorf("null object = %#v, want nil value", r)
	}
}

func TestResolveThrowMapsCode(t *testing.T) {
	c := useDomain(t, "throw", New(evalJS(`{
		resolveBooleanEvaluation() { throw Object.assign(new Error("m"), { code: "FLAG_NOT_FOUND" }) },
		resolveStringEvaluation() { throw "boom" },
		resolveNumberEvaluation() { throw new TypeError("plain") },
		resolveObjectEvaluation() { return undefined },
	}`)))
	b := c.BooleanDetails("f", true)
	if b.ErrorCode != openfeature.CodeFlagNotFound || b.ErrorMessage != "m" || b.Reason != openfeature.ReasonError || !b.Value {
		t.Errorf("coded throw = %#v", b)
	}
	s := c.StringDetails("f", "d")
	if s.ErrorCode != openfeature.CodeGeneral || s.ErrorMessage != "" || s.Value != "d" {
		t.Errorf(`throw "boom" = %#v, want GENERAL without message`, s)
	}
	n := c.NumberDetails("f", 2)
	if n.ErrorCode != openfeature.CodeGeneral || n.ErrorMessage != "plain" {
		t.Errorf("plain throw = %#v", n)
	}
	if o := c.ObjectDetails("f", nil); o.ErrorCode != openfeature.CodeGeneral || o.ErrorMessage == "" {
		t.Errorf("undefined resolution = %#v, want a GENERAL TypeError", o)
	}
}

func TestResolveLoggerForwards(t *testing.T) {
	p := New(evalJS(`{ resolveBooleanEvaluation(k, d, c, l) {
		l.error("e", 1, null, { a: 1 }); l.warn(Object.assign(new Error("w"), { code: "GENERAL" })); l.info(); l.debug(undefined);
		return { value: true };
	} }`))
	var l logs
	p.ResolveBoolean("f", false, openfeature.Context{}, &l)
	if len(l.lines) != 4 {
		t.Fatalf("lines = %#v", l.lines)
	}
	if want := []any{"error", "e", "1", "null", "[object Object]"}; !reflect.DeepEqual(l.lines[0], want) {
		t.Errorf("error line = %#v, want %#v", l.lines[0], want)
	}
	var oe *openfeature.Error
	if err, ok := l.lines[1][1].(error); !ok || !errors.As(err, &oe) || oe.Message != "w" {
		t.Errorf("warn line = %#v, want the Error as a Go error", l.lines[1])
	}
	if want := []any{"debug", "undefined"}; !reflect.DeepEqual(l.lines[3], want) {
		t.Errorf("debug line = %#v, want %#v", l.lines[3], want)
	}
}

func TestContextConversion(t *testing.T) {
	p := New(evalJS(`{ resolveBooleanEvaluation(k, d, c) { globalThis.seen = c; return { value: true } } }`))
	at := time.Date(2001, 2, 3, 4, 5, 6, 7e6, time.UTC)
	p.ResolveBoolean("f", false, openfeature.Context{TargetingKey: "k", Attributes: map[string]any{
		"t": at, "n": map[string]any{"a": []any{1, "x", int64(2)}}, "z": nil, "u": uint8(3), "b": true,
	}}, openfeature.DefaultLogger())
	seen := app.Window().Get("seen")
	if seen.Get("targetingKey").String() != "k" {
		t.Errorf("targetingKey = %v", seen.Get("targetingKey"))
	}
	if d := seen.Get("t"); !d.InstanceOf(app.Window().Get("Date")) || int64(d.Call("getTime").Float()) != at.UnixMilli() {
		t.Errorf("time = %v, want a Date at %d", d, at.UnixMilli())
	}
	if got := app.Window().Get("JSON").Call("stringify", seen).String(); got != `{"targetingKey":"k","b":true,"n":{"a":[1,"x",2]},"t":"2001-02-03T04:05:06.007Z","u":3,"z":null}` {
		t.Errorf("context = %s", got)
	}
	p.ResolveBoolean("f", false, openfeature.Context{Attributes: map[string]any{"a": 1}}, openfeature.DefaultLogger())
	if got := app.Window().Get("JSON").Call("stringify", app.Window().Get("seen")).String(); got != `{"a":1}` {
		t.Errorf("empty targeting key: context = %s, want no targetingKey", got)
	}
}

func TestInitializeRejectFatal(t *testing.T) {
	p := New(evalJS(`{` + resolvers + `
		initialize(ctx, domain) { globalThis.init = [ctx.targetingKey, domain]; return Promise.reject(Object.assign(new Error("dead"), { code: "PROVIDER_FATAL" })) },
	}`))
	pend, err := p.(openfeature.Initializer).Initialize(openfeature.Context{TargetingKey: "k"}, "")
	if err != nil {
		t.Fatal(err)
	}
	var oe *openfeature.Error
	if err := wait(t, pend); !errors.As(err, &oe) || oe.Code != openfeature.CodeProviderFatal || oe.Message != "dead" {
		t.Errorf("err = %#v, want PROVIDER_FATAL dead", err)
	}
	if got := app.Window().Get("JSON").Call("stringify", app.Window().Get("init")).String(); got != `["k",null]` {
		t.Errorf("initialize args = %s, want domain undefined", got)
	}
	sync := New(evalJS(`{` + resolvers + ` initialize() { throw new Error("sync") } }`))
	func() {
		defer func() {
			if err, ok := recover().(error); !ok || err.Error() != "sync" {
				t.Errorf("sync throw panicked with %v, want sync", err)
			}
		}()
		_, _ = sync.(openfeature.Initializer).Initialize(openfeature.Context{}, "d")
		t.Error("sync throw did not panic")
	}()
	plain := New(evalJS(`{` + resolvers + ` initialize() {} }`))
	if pend, err := plain.(openfeature.Initializer).Initialize(openfeature.Context{}, "d"); pend != nil || err != nil {
		t.Errorf("void initialize = %v, %v, want nil, nil", pend, err)
	}
}

func TestOnContextChangeSyncVsPromise(t *testing.T) {
	cc := func(body string) openfeature.ContextChanger {
		return New(evalJS(`{` + resolvers + ` onContextChange(o, n) { globalThis.change = [o.targetingKey, n.targetingKey]; ` + body + ` } }`)).(openfeature.ContextChanger)
	}
	old, next := openfeature.Context{TargetingKey: "a"}, openfeature.Context{TargetingKey: "b"}
	if pend, err := cc(`return 1`).OnContextChange(old, next); pend != nil || err != nil {
		t.Errorf("sync = %v, %v, want nil, nil", pend, err)
	}
	if got := app.Window().Get("JSON").Call("stringify", app.Window().Get("change")).String(); got != `["a","b"]` {
		t.Errorf("args = %s", got)
	}
	pend, err := cc(`return new Promise(r => setTimeout(r, 1))`).OnContextChange(old, next)
	if err != nil || pend == nil {
		t.Fatalf("promise = %v, %v, want a Pending", pend, err)
	}
	if err := wait(t, pend); err != nil {
		t.Errorf("resolved promise = %v", err)
	}
	pend, _ = cc(`return Promise.reject("boom")`).OnContextChange(old, next)
	if err := wait(t, pend); err == nil || err.Error() != "" {
		t.Errorf("rejected with a string = %#v, want an error without message", err)
	}
	if _, err := cc(`throw new Error("t")`).OnContextChange(old, next); err == nil || err.Error() != "t" {
		t.Errorf("sync throw = %v", err)
	}
}

const emitter = `{
	h: {},
	addHandler(t, f) { (this.h[t] ??= []).push(f) },
	removeHandler(t, f) { this.h[t] = (this.h[t] || []).filter(x => x !== f) },
	emit(t, d) { (this.h[t] || []).forEach(f => f(d)) },
}`

func TestEventsForwarded(t *testing.T) {
	v := evalJS(`{` + resolvers + ` events: ` + emitter + ` }`)
	p := New(v)
	em := p.(openfeature.Eventer).Events()
	if em == nil || em != p.(openfeature.Eventer).Events() {
		t.Fatal("Events must return the same non-nil emitter")
	}
	c := useDomain(t, "events", p)
	var got []openfeature.EventDetails
	for _, ev := range []openfeature.Event{openfeature.EventStale, openfeature.EventConfigurationChanged, openfeature.EventError} {
		c.AddHandler(ev, func(d openfeature.EventDetails) { got = append(got, d) })
	}
	v.Get("events").Call("emit", "PROVIDER_STALE", evalJS(`{ message: "m" }`))
	if openfeature.ProviderStatus("events") != openfeature.StatusStale {
		t.Errorf("status = %s, want STALE", openfeature.ProviderStatus("events"))
	}
	v.Get("events").Call("emit", "PROVIDER_CONFIGURATION_CHANGED", evalJS(`{ flagsChanged: ["a", "b"], metadata: { k: 1 } }`))
	v.Get("events").Call("emit", "PROVIDER_ERROR", evalJS(`{ errorCode: "PROVIDER_FATAL", message: "x" }`))
	v.Get("events").Call("emit", "PROVIDER_STALE")
	if len(got) != 4 {
		t.Fatalf("events = %#v", got)
	}
	if got[0].Message != "m" || got[0].Domain != "events" {
		t.Errorf("stale = %#v", got[0])
	}
	if !reflect.DeepEqual(got[1].FlagsChanged, []string{"a", "b"}) || !reflect.DeepEqual(got[1].Metadata, map[string]any{"k": 1.0}) {
		t.Errorf("configuration changed = %#v", got[1])
	}
	if got[2].ErrorCode != openfeature.CodeProviderFatal || got[2].Message != "x" {
		t.Errorf("error = %#v", got[2])
	}
	if openfeature.ProviderStatus("events") != openfeature.StatusStale {
		t.Errorf("status = %s, want STALE after the last event", openfeature.ProviderStatus("events"))
	}
}

func TestJSHookWrapped(t *testing.T) {
	v := evalJS(`{` + resolvers + `
		metadata: { name: "hooked" },
		hooks: [{
			before(hc, hints) { globalThis.hook = { key: hc.flagKey, type: hc.flagValueType, def: hc.defaultValue, ctx: hc.context.targetingKey, provider: hc.providerMetadata.name, domain: hc.clientMetadata.domain, hint: hints && hints.h, frozen: Object.isFrozen(hc) }; hc.hookData.set("n", 1); globalThis.first = hc },
			after(hc, details) { globalThis.hook.after = [hc.hookData.get("n"), details.value, details.flagKey, hc === globalThis.first] },
			finally(hc, details) { globalThis.hook.finally = details.reason ?? "none" },
		}, {
			error(hc, err) { globalThis.hookErr = [err.message, err.code] },
		}],
	}`)
	c := useDomain(t, "hooks", New(v))
	openfeature.SetContext(openfeature.Context{TargetingKey: "tk"}, "hooks")
	if !c.Boolean("f", true, openfeature.WithHookHints(openfeature.NewHookHints(map[string]any{"h": "x"}))) {
		t.Fatal("value")
	}
	got := app.Window().Get("JSON").Call("stringify", app.Window().Get("hook")).String()
	if want := `{"key":"f","type":"boolean","def":true,"ctx":"tk","provider":"hooked","domain":"hooks","hint":"x","frozen":true,"after":[1,true,"f",true],"finally":"none"}`; got != want {
		t.Errorf("hook saw %s, want %s", got, want)
	}
	v.Set("resolveBooleanEvaluation", evalJS(`function() { throw Object.assign(new Error("e"), { code: "PARSE_ERROR" }) }`))
	c.Boolean("f", true)
	if got := app.Window().Get("JSON").Call("stringify", app.Window().Get("hookErr")).String(); got != `["e","PARSE_ERROR"]` {
		t.Errorf("error hook saw %s", got)
	}
	v.Set("resolveBooleanEvaluation", evalJS(`function() { return { value: true, errorCode: "FLAG_NOT_FOUND", errorMessage: "nf" } }`))
	c.Boolean("f", true)
	if got := app.Window().Get("JSON").Call("stringify", app.Window().Get("hookErr")).String(); got != `["nf","FLAG_NOT_FOUND"]` {
		t.Errorf("error hook saw %s for a returned error code", got)
	}
}

func TestJSHookThrows(t *testing.T) {
	var l logs
	v := evalJS(`{` + resolvers + ` hooks: [{ before() { throw "boom" }, finally() { throw "fin" } }] }`)
	c := useDomain(t, "hookthrow", New(v))
	c.SetLogger(&l)
	d := c.BooleanDetails("f", true)
	if d.ErrorCode != openfeature.CodeGeneral || d.ErrorMessage != "" {
		t.Errorf("before throw = %#v, want GENERAL without message", d)
	}
	if len(l.lines) == 0 || l.lines[0][1] != "Unhandled error during 'finally' hook: fin" {
		t.Errorf("logs = %#v", l.lines)
	}
}

func TestOptionalInterfaces(t *testing.T) {
	for bits := range 4 {
		v := evalJS(`{` + resolvers + `}`)
		for i, m := range []string{"initialize", "track"} {
			if bits>>i&1 == 1 {
				v.Set(m, evalJS("function() {}"))
			}
		}
		p := New(v)
		got := 0
		for i, ok := range []bool{implements[openfeature.Initializer](p), implements[openfeature.Tracker](p)} {
			if ok {
				got |= 1 << i
			}
		}
		if got != bits {
			t.Errorf("members %02b: implements %02b", bits, got)
		}
		if !implements[openfeature.ContextChanger](p) || !implements[openfeature.Closer](p) || !implements[openfeature.Hooked](p) ||
			!implements[openfeature.Eventer](p) || !implements[openfeature.RunsOn](p) || !implements[openfeature.DomainScoped](p) {
			t.Errorf("members %02b: an unobservable optional interface is missing", bits)
		}
	}
	v := evalJS(`{` + resolvers + ` initialize: 1, track: {} }`)
	if p := New(v); implements[openfeature.Initializer](p) || implements[openfeature.Tracker](p) {
		t.Error("non-function initialize or track must not count as present")
	}
	p := New(evalJS(`{` + resolvers + ` runsOn: "server", domainScoped: 1, hooks: [{}, {}] }`))
	if p.(openfeature.RunsOn).RunsOn() != "server" || !p.(openfeature.DomainScoped).DomainScoped() || len(p.(openfeature.Hooked).Hooks()) != 2 {
		t.Error("runsOn, domainScoped or hooks not passed through")
	}
}

func implements[I any](p openfeature.Provider) bool {
	_, ok := p.(I)
	return ok
}

func TestMetadataReadLive(t *testing.T) {
	v := evalJS(`{` + resolvers + ` metadata: { name: "a", v: "1", n: 2 } }`)
	p := New(v)
	if md := p.Metadata(); md.Name != "a" || !reflect.DeepEqual(md.Attributes, map[string]string{"v": "1", "n": "2"}) {
		t.Errorf("metadata = %#v", md)
	}
	v.Set("metadata", evalJS(`Object.create({ get name() { return "b" } })`))
	if md := p.Metadata(); md.Name != "b" || md.Attributes != nil {
		t.Errorf("metadata after change = %#v", md)
	}
}

func TestTrack(t *testing.T) {
	v := evalJS(`{` + resolvers + ` track(name, ctx, details) { globalThis.tracked = [name, ctx, details]; if (name === "x") throw new Error("t") } }`)
	p := New(v).(openfeature.Tracker)
	p.Track("n", openfeature.Context{TargetingKey: "k"}, openfeature.TrackingDetails{Value: new(2.5), Attributes: map[string]any{"a": "b"}})
	if got := app.Window().Get("JSON").Call("stringify", app.Window().Get("tracked")).String(); got != `["n",{"targetingKey":"k"},{"a":"b","value":2.5}]` {
		t.Errorf("track args = %s", got)
	}
	defer func() {
		if r := recover(); r == nil {
			t.Error("a throwing track must panic for the SDK to recover")
		}
	}()
	p.Track("x", openfeature.Context{}, openfeature.TrackingDetails{})
}
