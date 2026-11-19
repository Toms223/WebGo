package jsprovider

import (
	"reflect"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
)

func stringify(v app.Value) string { return app.Window().Get("JSON").Call("stringify", v).String() }

func record(t *testing.T) *[]openfeature.Event {
	var got []openfeature.Event
	for _, ev := range []openfeature.Event{
		openfeature.EventReady, openfeature.EventError, openfeature.EventContextChanged,
		openfeature.EventReconciling, openfeature.EventStale,
	} {
		openfeature.AddHandler(ev, func(openfeature.EventDetails) { got = append(got, ev) })
	}
	return &got
}

func TestNoEventsAfterClose(t *testing.T) {
	v := evalJS(`{` + resolvers + ` events: ` + emitter + `, onClose() {} }`)
	c := useDomain(t, "closing", New(v))
	var n int
	c.AddHandler(openfeature.EventStale, func(openfeature.EventDetails) { n++ })
	openfeature.Close()
	v.Get("events").Call("emit", "PROVIDER_STALE")
	if n != 0 || openfeature.ProviderStatus("closing") != openfeature.StatusNotReady {
		t.Errorf("after Close: %d events, status %s", n, openfeature.ProviderStatus("closing"))
	}
}

func TestCloseThenRebind(t *testing.T) {
	v := evalJS(`{` + resolvers + ` events: ` + emitter + `, onClose() { return new Promise(r => setTimeout(r, 5)) } }`)
	p := New(v)
	useDomain(t, "rebind", p)
	openfeature.ClearProviders()
	if err := openfeature.SetProviderAndWait(p, openfeature.ForDomain("rebind")); err != nil {
		t.Fatal(err)
	}
	var n int
	openfeature.GetClient("rebind").AddHandler(openfeature.EventStale, func(openfeature.EventDetails) { n++ })
	done := make(chan struct{})
	app.Window().Call("setTimeout", app.FuncOf(func(app.Value, []app.Value) any { close(done); return nil }), 20)
	<-done
	v.Get("events").Call("emit", "PROVIDER_STALE")
	if n != 1 {
		t.Errorf("after close and rebind: %d events, want 1", n)
	}
	if h := v.Get("events").Get("h").Get("PROVIDER_STALE").Length(); h != 1 {
		t.Errorf("JS handlers for STALE = %d, want 1", h)
	}
}

func TestSameJSObjectSameProvider(t *testing.T) {
	v := evalJS(`{` + resolvers + ` initialize() { globalThis.inits = (globalThis.inits || 0) + 1 } }`)
	if New(v) != New(v) {
		t.Fatal("New(v) twice returned different providers")
	}
	useDomain(t, "same", New(v))
	got := record(t)
	if err := openfeature.SetProviderAndWait(New(v), openfeature.ForDomain("same")); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Errorf("rebinding the same JS object: events %v, want only the replayed Ready", *got)
	}
	if err := openfeature.SetProviderAndWait(New(v), openfeature.ForDomain("other")); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 2 || app.Window().Get("inits").Int() != 1 {
		t.Errorf("events %v, initialize calls %d; want a second Ready and one initialize", *got, app.Window().Get("inits").Int())
	}
}

func TestAbsentMembersBehaveAsAbsent(t *testing.T) {
	t.Run("onContextChange", func(t *testing.T) {
		useDomain(t, "noctx", New(evalJS(`{`+resolvers+`}`)))
		got := record(t)
		openfeature.SetContext(openfeature.Context{TargetingKey: "k"}, "noctx")
		if !reflect.DeepEqual(*got, []openfeature.Event{openfeature.EventReady, openfeature.EventContextChanged}) {
			t.Errorf("events = %v, want the replayed Ready then ContextChanged, no Reconciling", *got)
		}
	})
	t.Run("onClose", func(t *testing.T) {
		var l logs
		useDomain(t, "noclose", New(evalJS(`{`+resolvers+`}`)))
		openfeature.SetLogger(&l)
		openfeature.Close()
		openfeature.SetProviderAndWait(New(evalJS(`{`+resolvers+`}`)), openfeature.ForDomain("noclose"))
		for _, line := range l.lines {
			if line[0] == "error" {
				t.Errorf("logged %v", line)
			}
		}
	})
	t.Run("hooks", func(t *testing.T) {
		p := New(evalJS(`{` + resolvers + ` hooks: undefined }`))
		if p.(openfeature.Hooked).Hooks() != nil {
			t.Error("Hooks() must be nil without hooks")
		}
		if c := useDomain(t, "nohooks", p); !c.Boolean("f", true) {
			t.Error("evaluation failed")
		}
	})
	t.Run("events", func(t *testing.T) {
		p := New(evalJS(`{` + resolvers + `}`))
		if p.(openfeature.Eventer).Events() != nil {
			t.Error("Events() must be nil without events")
		}
		got := record(t)
		useDomain(t, "noevents", p)
		if !reflect.DeepEqual(*got, []openfeature.Event{openfeature.EventReady}) || openfeature.ProviderStatus("noevents") != openfeature.StatusReady {
			t.Errorf("events = %v, status %s", *got, openfeature.ProviderStatus("noevents"))
		}
	})
	t.Run("runsOn", func(t *testing.T) {
		var l logs
		t.Cleanup(openfeature.Reset)
		openfeature.SetLogger(&l)
		openfeature.SetProviderAndWait(New(evalJS(`{`+resolvers+` metadata: { name: "n" } }`)), openfeature.ForDomain("norun"))
		if len(l.lines) == 0 || !reflect.DeepEqual(l.lines[0], []any{"debug", "Provider 'n' has not defined its intended use."}) {
			t.Errorf("logs = %v", l.lines)
		}
	})
	t.Run("domainScoped", func(t *testing.T) {
		p := New(evalJS(`{` + resolvers + `}`))
		useDomain(t, "scope1", p)
		if err := openfeature.SetProviderAndWait(p, openfeature.ForDomain("scope2")); err != nil {
			t.Errorf("binding twice = %v, want nil", err)
		}
	})
}

func TestCycles(t *testing.T) {
	p := New(evalJS(`{ resolveObjectEvaluation(k, d, c) {
		globalThis.ctxCycle = c.m.self === c.m && c.s[0] === c.s;
		const o = { a: [1] }; o.self = o; o.a.push(o.a);
		return { value: o };
	} }`))
	m := map[string]any{}
	m["self"] = m
	s := []any{nil}
	s[0] = s
	r := p.ResolveObject("f", nil, openfeature.Context{Attributes: map[string]any{"m": m, "s": s}}, openfeature.DefaultLogger())
	if !app.Window().Get("ctxCycle").Bool() {
		t.Error("Go cycles not preserved in JS")
	}
	o := r.Value.(openfeature.Object)
	av, _ := o.Get("a")
	a := av.([]any)
	self, _ := o.Get("self")
	if reflect.ValueOf(self).Pointer() != reflect.ValueOf(o).Pointer() || reflect.ValueOf(a[1]).Pointer() != reflect.ValueOf(a).Pointer() {
		t.Error("JS cycles not preserved in Go")
	}
	ord := openfeature.Object{{Key: "z", Value: 1}, {Key: "a", Value: openfeature.Object{{Key: "y", Value: 2}, {Key: "b", Value: nil}}}}
	ord = append(ord, openfeature.Field{Key: "m", Value: map[string]any{"k": 1}})
	p = New(evalJS(`{ resolveObjectEvaluation(k, d) {
		globalThis.orderSeen = JSON.stringify(d);
		return { value: { zeta: 1, alpha: { b: 1, a: 2 }, 2: 0 } };
	} }`))
	r = p.ResolveObject("f", ord, openfeature.Context{}, openfeature.DefaultLogger())
	if got := app.Window().Get("orderSeen").String(); got != `{"z":1,"a":{"y":2,"b":null},"m":{"k":1}}` {
		t.Errorf("Object to JS = %s", got)
	}
	want := openfeature.Object{{Key: "2", Value: 0.0}, {Key: "zeta", Value: 1.0}, {Key: "alpha", Value: openfeature.Object{{Key: "b", Value: 1.0}, {Key: "a", Value: 2.0}}}}
	if !reflect.DeepEqual(r.Value, want) {
		t.Errorf("JS to Object = %#v", r.Value)
	}
}

func TestJSErrorLogLines(t *testing.T) {
	cases := map[string][][]any{
		`Object.assign(new TypeError("x"), { stack: "STACK" })`: {{"error", "Unhandled error during 'finally' hook: TypeError: x"}, {"error", "STACK"}, {"error", "STACK"}},
		`{ a: 1 }`:    {{"error", "Unhandled error during 'finally' hook: [object Object]"}, {"error"}},
		`null`:        {{"error", "Unhandled error during 'finally' hook: null"}, {"error"}},
		`Symbol("s")`: {{"error", "Unhandled error during 'finally' hook: Symbol(s)"}, {"error"}},
	}
	for thrown, want := range cases {
		var l logs
		c := useDomain(t, "logs", New(evalJS(`{`+resolvers+` hooks: [{ finally() { throw `+thrown+` } }] }`)))
		c.SetLogger(&l)
		c.Boolean("f", true)
		if !reflect.DeepEqual(l.lines, want) {
			t.Errorf("throw %s logged %v, want %v", thrown, l.lines, want)
		}
		openfeature.Reset()
	}
	var l logs
	useDomain(t, "shutdown", New(evalJS(`{`+resolvers+` metadata: { name: "p" }, onClose() { return Promise.reject(Object.assign(new TypeError("x"), { stack: "STACK" })) } }`)))
	openfeature.SetLogger(&l)
	openfeature.Close()
	if want := [][]any{{"error", "Error during shutdown of provider p: TypeError: x"}, {"error", "STACK"}}; !reflect.DeepEqual(l.lines, want) {
		t.Errorf("shutdown logged %v, want %v", l.lines, want)
	}
}

func TestLoggerPerGoLogger(t *testing.T) {
	p := New(evalJS(`{ resolveBooleanEvaluation(k, d, c, l) { (globalThis.ls ??= []).push(l); return { value: true } } }`))
	var l1, l2 logs
	for _, l := range []openfeature.Logger{&l1, &l1, &l2} {
		p.ResolveBoolean("f", false, openfeature.Context{}, l)
	}
	ls := app.Window().Get("ls")
	if !ls.Index(0).Equal(ls.Index(1)) || ls.Index(0).Equal(ls.Index(2)) {
		t.Error("want one JS logger per distinct Go logger")
	}
}

func TestMissingResolver(t *testing.T) {
	c := useDomain(t, "missing", New(evalJS(`{ resolveBooleanEvaluation() { return { value: true } }, resolveStringEvaluation: 1 }`)))
	if d := c.ObjectDetails("f", nil); d.ErrorCode != openfeature.CodeGeneral || d.ErrorMessage != "Cannot read properties of undefined (reading 'call')" {
		t.Errorf("missing resolver = %#v", d)
	}
	if d := c.StringDetails("f", ""); d.ErrorMessage != "resolver.call is not a function" {
		t.Errorf("non-function resolver = %#v", d)
	}
}

func TestHookDataAndFrozenHints(t *testing.T) {
	c := useDomain(t, "hookdata", New(evalJS(`{`+resolvers+` hooks: [{
		before(hc) { const d = hc.hookData; globalThis.hd = [d.set("a", 1) === undefined, d.get("a"), d.has("a"), d.delete("a"), d.has("a"), d instanceof Map, d.data instanceof Map] },
		after(hc, details, hints) { globalThis.frozenHints = Object.isFrozen(hints) },
	}], track(n, ctx) { globalThis.frozenCtx = Object.isFrozen(ctx) } }`)))
	c.Boolean("f", true, openfeature.WithHookHints(openfeature.NewHookHints(map[string]any{"h": 1})))
	if got := stringify(app.Window().Get("hd")); got != `[true,1,true,true,false,false,true]` {
		t.Errorf("hookData = %s", got)
	}
	c.Track("t")
	if !app.Window().Get("frozenHints").Bool() || !app.Window().Get("frozenCtx").Bool() {
		t.Error("hints in after or the track context not frozen")
	}
}

func TestCallbackPanicContained(t *testing.T) {
	p := New(evalJS(`{ resolveBooleanEvaluation(k, d, c, l) { l.error({ toString() { throw new Error("no") } }); return { value: true } } }`))
	if !p.ResolveBoolean("f", false, openfeature.Context{}, &logs{}).Value {
		t.Error("resolution lost")
	}
}

func TestAddHandlerThrows(t *testing.T) {
	v := evalJS(`{` + resolvers + ` events: { n: 0, addHandler(t) { if (++this.n === 3) throw new Error("full"); (this.t ??= []).push(t) } } }`)
	p := New(v).(openfeature.Eventer)
	func() {
		defer func() {
			if err, ok := recover().(error); !ok || err.Error() != "full" {
				t.Errorf("recovered %v, want the JS error", err)
			}
		}()
		p.Events()
	}()
	if p.Events() == nil {
		t.Fatal("retry failed")
	}
	if got := stringify(v.Get("events").Get("t")); got != `["PROVIDER_READY","PROVIDER_ERROR","PROVIDER_CONFIGURATION_CHANGED","PROVIDER_CONTEXT_CHANGED","PROVIDER_RECONCILING","PROVIDER_STALE"]` {
		t.Errorf("registered %s, want each event once", got)
	}
}

func TestPointerCycle(t *testing.T) {
	p := New(evalJS(`{ resolveBooleanEvaluation(k, d, c) { globalThis.ptr = c.x; return { value: true } } }`))
	var x any
	x = &x
	p.ResolveBoolean("f", false, openfeature.Context{Attributes: map[string]any{"x": x}}, openfeature.DefaultLogger())
	if !app.Window().Get("ptr").IsNull() {
		t.Errorf("pointer cycle = %v, want null", app.Window().Get("ptr"))
	}
}

func TestHookDataClassName(t *testing.T) {
	c := useDomain(t, "hdname", New(evalJS(`{`+resolvers+` hooks: [{ before(hc) { globalThis.hdName = hc.hookData.constructor.name } }] }`)))
	c.Boolean("f", true)
	if got := app.Window().Get("hdName").String(); got != "MapHookData" {
		t.Errorf("constructor.name = %q", got)
	}
}

func TestNewNonObject(t *testing.T) {
	values := []app.Value{
		app.Undefined(), app.Null(), app.ValueOf("x"), app.ValueOf(""), app.ValueOf(1), app.ValueOf(0),
		app.ValueOf(true), app.ValueOf(false), evalJS(`function() {}`),
	}
	for _, v := range values {
		if p := New(v); p != nil {
			t.Errorf("New(%v) = %v, want nil", v, p)
		}
		var l logs
		openfeature.Reset()
		openfeature.SetLogger(&l)
		if err := openfeature.SetProviderAndWait(New(v), openfeature.ForDomain("nonobject")); err != nil || openfeature.ProviderMetadata("nonobject").Name != "No-op Provider" {
			t.Errorf("New(%v): err %v, provider %q", v, err, openfeature.ProviderMetadata("nonobject").Name)
		}
		if len(l.lines) != 1 || !reflect.DeepEqual(l.lines[0], []any{"debug", "No provider defined, ignoring setProvider call"}) {
			t.Errorf("New(%v): logs %v", v, l.lines)
		}
	}
	openfeature.Reset()
}

func TestAddHandlerThrowIsSetProviderError(t *testing.T) {
	t.Cleanup(openfeature.Reset)
	v := evalJS(`{` + resolvers + ` events: { addHandler() { throw new Error("nope") } } }`)
	if err := openfeature.SetProviderAndWait(New(v), openfeature.ForDomain("throwing")); err == nil || err.Error() != "nope" {
		t.Errorf("err = %v, want nope", err)
	}
	if openfeature.ProviderMetadata("throwing").Name != "No-op Provider" {
		t.Error("provider was bound")
	}
}
