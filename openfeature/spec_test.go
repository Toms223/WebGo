package openfeature_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	of "github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/isolated"
	"github.com/Toms223/WebGo/openfeature/memprovider"
)

type fake struct {
	name     string
	ev       *of.Emitter
	out      *of.ResolutionDetails[any]
	panicErr any
	hooks    []of.Hook
	log      []string
	ctx      of.Context

	initCtx        of.Context
	domain         string
	oldCtx, newCtx of.Context
	trackName      string
	trackCtx       of.Context
	trackDetails   of.TrackingDetails
	init           func() (of.Pending, error)
	change         func() (of.Pending, error)
}

func newFake(name string) *fake { return &fake{name: name, ev: of.NewEmitter()} }

func (f *fake) Metadata() of.Metadata { return of.Metadata{Name: f.name} }
func (f *fake) Events() *of.Emitter   { return f.ev }
func (f *fake) Hooks() []of.Hook      { return f.hooks }

func resolve[T any](f *fake, flag string, def T, ctx of.Context) of.ResolutionDetails[T] {
	f.log = append(f.log, "resolve:"+flag)
	f.ctx = ctx
	if f.panicErr != nil {
		panic(f.panicErr)
	}
	d := of.ResolutionDetails[T]{Value: def}
	if o := f.out; o != nil {
		d.Variant, d.Reason, d.ErrorCode, d.ErrorMessage, d.FlagMetadata = o.Variant, o.Reason, o.ErrorCode, o.ErrorMessage, o.FlagMetadata
		if v, ok := o.Value.(T); ok {
			d.Value = v
		}
	}
	return d
}

func (f *fake) ResolveBoolean(k string, d bool, c of.Context, _ of.Logger) of.ResolutionDetails[bool] {
	return resolve(f, k, d, c)
}
func (f *fake) ResolveString(k string, d string, c of.Context, _ of.Logger) of.ResolutionDetails[string] {
	return resolve(f, k, d, c)
}
func (f *fake) ResolveNumber(k string, d float64, c of.Context, _ of.Logger) of.ResolutionDetails[float64] {
	return resolve(f, k, d, c)
}
func (f *fake) ResolveObject(k string, d any, c of.Context, _ of.Logger) of.ResolutionDetails[any] {
	return resolve(f, k, d, c)
}

type lifecycle struct{ *fake }

func full(name string) *lifecycle { return &lifecycle{newFake(name)} }

func (l *lifecycle) Initialize(ctx of.Context, domain string) (of.Pending, error) {
	l.log = append(l.log, "initialize")
	l.initCtx, l.domain = ctx, domain
	if l.init != nil {
		return l.init()
	}
	return nil, nil
}

func (l *lifecycle) OnContextChange(old, new of.Context) (of.Pending, error) {
	l.log = append(l.log, "change")
	l.oldCtx, l.newCtx = old, new
	if l.change != nil {
		return l.change()
	}
	return nil, nil
}

func (l *lifecycle) Close() (of.Pending, error) {
	l.log = append(l.log, "close")
	return nil, nil
}

func (l *lifecycle) Track(name string, ctx of.Context, d of.TrackingDetails) {
	l.trackName, l.trackCtx, l.trackDetails = name, ctx, d
}

type emitting struct{ *lifecycle }

func (emitting) EmitsLifecycleEvents() bool { return true }

type scoped struct{ *lifecycle }

func (scoped) DomainScoped() bool { return true }

type recLogger struct{ lines []string }

func (l *recLogger) add(a ...any)   { l.lines = append(l.lines, fmt.Sprint(a...)) }
func (l *recLogger) Error(a ...any) { l.add(a...) }
func (l *recLogger) Warn(a ...any)  { l.add(a...) }
func (l *recLogger) Info(a ...any)  { l.add(a...) }
func (l *recLogger) Debug(a ...any) { l.add(a...) }

func eq[T any](t *testing.T, what string, got, want T) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %#v, want %#v", what, got, want)
	}
}

func set(t *testing.T, p of.Provider, opts ...of.SetOption) {
	t.Helper()
	if err := of.SetProviderAndWait(p, opts...); err != nil {
		t.Fatalf("SetProviderAndWait: %v", err)
	}
	synctest.Wait()
}

func failed(err error) func() (of.Pending, error) {
	return func() (of.Pending, error) {
		ch := make(chan error, 1)
		ch <- err
		return ch, nil
	}
}

func count(log []string, s string) (n int) {
	for _, l := range log {
		if l == s {
			n++
		}
	}
	return n
}

func withOut(t *testing.T, out of.ResolutionDetails[any]) (*fake, *of.Client) {
	t.Helper()
	p := newFake("p")
	p.out = &out
	set(t, p)
	return p, of.GetClient()
}

func stageHook(log *[]string, name string) of.Hook {
	add := func(stage string) { *log = append(*log, name+"."+stage) }
	return of.Hook{
		Before:  func(of.HookContext, of.HookHints) error { add("before"); return nil },
		After:   func(of.HookContext, of.EvaluationDetails[any], of.HookHints) error { add("after"); return nil },
		Error:   func(of.HookContext, error, of.HookHints) { add("error") },
		Finally: func(of.HookContext, of.EvaluationDetails[any], of.HookHints) { add("finally") },
	}
}

func exportedAPI(t *testing.T) []string {
	t.Helper()
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, f := range pkgs["openfeature"].Files {
		for _, d := range f.Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.IsExported() {
				names = append(names, fn.Name.Name)
			}
		}
	}
	if len(names) < 20 {
		t.Fatalf("parsed only %d exported functions", len(names))
	}
	return names
}

const (
	dynamic    = "applies to the dynamic-context paradigm only; this SDK is static-context"
	staticOnly = "static-context: transaction, client, invocation and before-hook contexts do not exist (3.2.2.2, 3.3.2.1, 4.3.3.1), so there is nothing to merge or pass on"
	author     = "obligation on provider or hook authors, not on the SDK"
)

type cover struct {
	test func(*testing.T)
	na   string
	gap  string
}

func na(reason string) cover  { return cover{na: reason} }
func gap(reason string) cover { return cover{gap: reason} }

func TestSpec(t *testing.T) {
	data, err := os.ReadFile("testdata/spec-v0.9.0.json")
	if err != nil {
		t.Fatal(err)
	}
	type rule struct {
		ID       string `json:"id"`
		Keyword  string `json:"RFC 2119 keyword"`
		Content  string `json:"content"`
		Children []rule `json:"children"`
	}
	var doc struct {
		Rules []rule `json:"rules"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	inSpec := map[string]bool{}
	var gaps []string
	var walk func([]rule)
	walk = func(rs []rule) {
		for _, r := range rs {
			walk(r.Children)
			if r.Keyword == "" {
				continue
			}
			id := r.ID[strings.LastIndex(r.ID, " ")+1:]
			inSpec[id] = true
			c, ok := spec[id]
			switch {
			case !ok:
				t.Errorf("%s (%s) has no test, n/a reason or recorded gap: %s", id, r.Keyword, r.Content)
			case c.test != nil:
				t.Run(id, func(t *testing.T) {
					synctest.Test(t, func(t *testing.T) {
						of.Reset()
						defer of.Reset()
						c.test(t)
					})
				})
			case c.gap != "":
				gaps = append(gaps, id)
				t.Logf("gap: %s (%s) %s", id, r.Keyword, c.gap)
			case c.na == "":
				t.Errorf("%s: empty coverage entry", id)
			}
		}
	}
	walk(doc.Rules)
	for id := range spec {
		if !inSpec[id] {
			t.Errorf("%s is not a requirement of spec v0.9.0", id)
		}
	}
	if len(inSpec) != 145 {
		t.Errorf("spec v0.9.0 has 145 requirements, parsed %d", len(inSpec))
	}
	slices.Sort(gaps)
	if len(gaps) != 0 {
		t.Errorf("recorded gaps: %v", gaps)
	}
}

var spec = map[string]cover{
	"1.1.1": {test: func(t *testing.T) {
		set(t, newFake("p"))
		eq(t, "api", of.ProviderMetadata().Name, "p")
		eq(t, "client", of.GetClient().Metadata().ProviderMetadata.Name, "p")
		eq(t, "other client", of.GetClient("x").Metadata().ProviderMetadata.Name, "p")
	}},
	"1.1.2.1": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		if of.GetProvider() != of.Provider(p) {
			t.Error("default provider not set")
		}
	}},
	"1.1.2.2": {test: func(t *testing.T) {
		p := full("p")
		set(t, p, of.ForDomain("d"))
		of.GetClient("d").Boolean("f", true)
		eq(t, "domain", p.domain, "d")
		eq(t, "order", p.log, []string{"initialize", "resolve:f"})
	}},
	"1.1.2.3": {test: func(t *testing.T) {
		a, b := full("a"), full("b")
		set(t, a, of.ForDomain("d"))
		set(t, a)
		set(t, b)
		eq(t, "still used by d", count(a.log, "close"), 0)
		set(t, b, of.ForDomain("d"))
		eq(t, "no longer used", count(a.log, "close"), 1)
	}},
	"1.1.2.4": {test: func(t *testing.T) {
		p := full("p")
		ch := make(chan error)
		p.init = func() (of.Pending, error) { return ch, nil }
		var ready, done bool
		of.AddHandler(of.EventReady, func(of.EventDetails) { ready = true })
		go func() { _ = of.SetProviderAndWait(p); done = true }()
		synctest.Wait()
		if done || ready {
			t.Error("returned before initialize terminated")
		}
		close(ch)
		synctest.Wait()
		if !done || !ready {
			t.Errorf("done %v, ready handler %v", done, ready)
		}
	}},
	"1.1.3": {test: func(t *testing.T) {
		set(t, newFake("a"), of.ForDomain("d"))
		eq(t, "bound", of.GetClient("d").Metadata().ProviderMetadata.Name, "a")
		set(t, newFake("b"), of.ForDomain("d"))
		eq(t, "overwritten", of.GetClient("d").Metadata().ProviderMetadata.Name, "b")
		eq(t, "default untouched", of.ProviderMetadata().Name, "No-op Provider")
	}},
	"1.1.4": {test: func(t *testing.T) {
		var log []string
		of.AddHooks(stageHook(&log, "a"))
		of.AddHooks(stageHook(&log, "b"), stageHook(&log, "c"))
		of.GetClient().Boolean("f", true)
		eq(t, "hooks", len(of.Hooks()), 3)
		eq(t, "before order", log[:3], []string{"a.before", "b.before", "c.before"})
	}},
	"1.1.5": {test: func(t *testing.T) {
		set(t, newFake("p"))
		eq(t, "name", of.ProviderMetadata().Name, "p")
	}},
	"1.1.6": {test: func(t *testing.T) {
		eq(t, "domain", of.GetClient("d").Metadata().Domain, "d")
		if of.GetClient() == nil {
			t.Error("no default client")
		}
	}},
	"1.1.7": {test: func(t *testing.T) {
		for _, d := range []string{"", "d", "\x00", strings.Repeat("x", 1<<16)} {
			if of.GetClient(d) == nil {
				t.Errorf("nil client for %q", d)
			}
		}
	}},
	"1.1.8.1": {test: func(t *testing.T) {
		p := scoped{full("p")}
		set(t, p, of.ForDomain("a"))
		if err := of.SetProvider(p, of.ForDomain("b")); err == nil {
			t.Error("second domain accepted")
		}
		eq(t, "b not bound", of.GetClient("b").Metadata().ProviderMetadata.Name, "No-op Provider")
	}},
	"1.2.1": {test: func(t *testing.T) {
		var log []string
		c := of.GetClient()
		c.AddHooks(stageHook(&log, "a"))
		c.AddHooks(stageHook(&log, "b"))
		c.Boolean("f", true)
		eq(t, "hooks", len(c.Hooks()), 2)
		eq(t, "before order", log[:2], []string{"a.before", "b.before"})
	}},
	"1.2.2": {test: func(t *testing.T) {
		eq(t, "domain", of.GetClient("d").Metadata().Domain, "d")
	}},
	"1.3.1.1": na(dynamic),
	"1.3.2.1": {test: func(t *testing.T) {
		p, c := withOut(t, of.ResolutionDetails[any]{Value: true})
		eq(t, "boolean", c.Boolean("f", false), true)
		p.out.Value = "s"
		eq(t, "string", c.String("f", "d"), "s")
		p.out.Value = 1.5
		eq(t, "number", c.Number("f", 0), 1.5)
		p.out.Value = map[string]any{"k": "v"}
		eq(t, "structure", c.Object("f", nil), any(map[string]any{"k": "v"}))
		eq(t, "options", c.Boolean("f", false, of.WithHooks(), of.WithHookHints(of.HookHints{})), false)
	}},
	"1.3.3.1": {test: func(t *testing.T) {
		_, c := withOut(t, of.ResolutionDetails[any]{Value: 3.0})
		eq(t, "int", c.Int("f", 0), int64(3))
		eq(t, "float", c.Number("f", 0), 3.0)
	}},
	"1.3.4": {test: func(t *testing.T) {
		_, c := withOut(t, of.ResolutionDetails[any]{Value: 2.5})
		d := c.IntDetails("f", 7)
		eq(t, "value", d.Value, int64(7))
		eq(t, "code", d.ErrorCode, of.CodeTypeMismatch)
	}},
	"1.4.1.1": na(dynamic),
	"1.4.2.1": {test: func(t *testing.T) {
		p, c := withOut(t, of.ResolutionDetails[any]{Value: true})
		eq(t, "boolean", c.BooleanDetails("f", false).Value, true)
		p.out.Value = "s"
		eq(t, "string", c.StringDetails("f", "d").Value, "s")
		p.out.Value = 1.5
		eq(t, "number", c.NumberDetails("f", 0).Value, 1.5)
		eq(t, "int", c.IntDetails("f", 0, of.WithHooks()).ErrorCode, of.CodeTypeMismatch)
		p.out.Value = []any{1.0}
		eq(t, "structure", c.ObjectDetails("f", nil).Value, any([]any{1.0}))
	}},
	"1.4.3": {test: func(t *testing.T) {
		_, c := withOut(t, of.ResolutionDetails[any]{Value: true})
		eq(t, "value", c.BooleanDetails("f", false).Value, true)
	}},
	"1.4.4.1": {test: func(t *testing.T) {
		c := of.GetClient()
		var _ of.EvaluationDetails[bool] = c.BooleanDetails("f", false)
		var _ of.EvaluationDetails[string] = c.StringDetails("f", "")
		var _ of.EvaluationDetails[float64] = c.NumberDetails("f", 0)
		var _ of.EvaluationDetails[int64] = c.IntDetails("f", 0)
		var _ of.EvaluationDetails[any] = c.ObjectDetails("f", nil)
	}},
	"1.4.5": {test: func(t *testing.T) {
		eq(t, "flag key", of.GetClient().BooleanDetails("my-flag", false).FlagKey, "my-flag")
	}},
	"1.4.6": {test: func(t *testing.T) {
		_, c := withOut(t, of.ResolutionDetails[any]{Variant: "on"})
		eq(t, "variant", c.BooleanDetails("f", false).Variant, "on")
	}},
	"1.4.7": {test: func(t *testing.T) {
		_, c := withOut(t, of.ResolutionDetails[any]{Reason: of.ReasonTargetingMatch})
		eq(t, "reason", c.BooleanDetails("f", false).Reason, of.ReasonTargetingMatch)
	}},
	"1.4.8":  {test: abnormal},
	"1.4.9":  {test: abnormal},
	"1.4.10": {test: abnormal},
	"1.4.11": {test: func(t *testing.T) {
		p := full("p")
		p.out = &of.ResolutionDetails[any]{Value: true, Variant: "on", Reason: of.ReasonStatic}
		set(t, p)
		l := &recLogger{}
		of.SetLogger(l)
		c := of.GetClient()
		c.BooleanDetails("f", false, of.WithHooks(of.Hook{}))
		c.String("f", "")
		c.Number("f", 0)
		c.Object("f", nil)
		c.Track("event")
		c.AddHooks(of.Hook{})
		c.AddHandler(of.EventReady, func(of.EventDetails) {})()
		c.Metadata()
		c.ProviderStatus()
		eq(t, "log lines", l.lines, nil)
	}},
	"1.4.12": na("static-context evaluation reads provider state already in memory and returns synchronously, as in the Web SDK; nothing blocks"),
	"1.4.13": {test: abnormal},
	"1.4.14": {test: flagMetadata},
	"1.4.15.1": {test: func(t *testing.T) {
		md := of.FlagMetadata{"k": "v"}
		_, c := withOut(t, of.ResolutionDetails[any]{FlagMetadata: md})
		c.BooleanDetails("f", false).FlagMetadata["k"] = "changed"
		eq(t, "provider metadata", md, of.FlagMetadata{"k": "v"})
		eq(t, "next evaluation", c.BooleanDetails("f", false).FlagMetadata, of.FlagMetadata{"k": "v"})
	}},
	"1.5.1": {test: func(t *testing.T) {
		var log []string
		of.AddHooks(stageHook(&log, "api"))
		c := of.GetClient()
		c.AddHooks(stageHook(&log, "client"))
		c.Boolean("f", true, of.WithHooks(stageHook(&log, "o1"), stageHook(&log, "o2")))
		eq(t, "before order", log[:4], []string{"api.before", "client.before", "o1.before", "o2.before"})
	}},
	"1.6.1": {test: func(t *testing.T) {
		a, b, c := full("a"), full("b"), full("c")
		c.init = failed(errors.New("down"))
		set(t, a)
		set(t, b, of.ForDomain("d"))
		_ = of.SetProviderAndWait(c, of.ForDomain("e"))
		eq(t, "status of c", of.ProviderStatus("e"), of.StatusError)
		of.Close()
		for _, p := range []*lifecycle{a, b, c} {
			eq(t, p.name+" shutdown calls", count(p.log, "close"), 1)
		}
	}},
	"1.6.2": {test: shutdownResets},
	"1.7.1": {test: func(t *testing.T) {
		c := of.GetClient()
		eq(t, "initial", c.ProviderStatus(), of.StatusNotReady)
		p := newFake("p")
		set(t, p)
		eq(t, "set", c.ProviderStatus(), of.StatusReady)
		p.ev.Emit(of.EventStale, of.EventDetails{})
		eq(t, "stale", c.ProviderStatus(), of.StatusStale)
		p.ev.Emit(of.EventError, of.EventDetails{})
		eq(t, "error", c.ProviderStatus(), of.StatusError)
		p.ev.Emit(of.EventError, of.EventDetails{ErrorCode: of.CodeProviderFatal})
		eq(t, "fatal", c.ProviderStatus(), of.StatusFatal)
	}},
	"1.7.2.1": {test: reconciling},
	"1.7.3": {test: func(t *testing.T) {
		p := full("p")
		set(t, p)
		eq(t, "after initialize", of.GetClient().ProviderStatus(), of.StatusReady)
		p.ev.Emit(of.EventError, of.EventDetails{})
		p.ev.Emit(of.EventReady, of.EventDetails{})
		eq(t, "after emitted READY", of.GetClient().ProviderStatus(), of.StatusReady)
	}},
	"1.7.4": {test: initError},
	"1.7.5": {test: func(t *testing.T) {
		p := full("p")
		p.init = failed(of.NewError(of.CodeProviderFatal, "dead"))
		if err := of.SetProviderAndWait(p); err == nil {
			t.Error("no error from SetProviderAndWait")
		}
		eq(t, "after initialize", of.GetClient().ProviderStatus(), of.StatusFatal)
		q := full("q")
		set(t, q)
		q.ev.Emit(of.EventError, of.EventDetails{ErrorCode: of.CodeProviderFatal})
		eq(t, "after emitted fatal", of.GetClient().ProviderStatus(), of.StatusFatal)
		eq(t, "evaluation", of.GetClient().BooleanDetails("f", true).ErrorCode, of.CodeProviderFatal)
	}},
	"1.7.6": {test: shutdownResets},
	"1.8.1": {test: isolatedInstances},
	"1.8.2": {test: sameContract},
	"1.8.3": {test: factoryElsewhere},
	"1.8.4": {test: oneInstancePerProvider},

	"2.1.1": {test: func(t *testing.T) {
		eq(t, "noop", of.NoopProvider().Metadata().Name, "No-op Provider")
		eq(t, "in-memory", memprovider.New().Metadata().Name, "in-memory")
	}},
	"2.2.1": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		of.SetContext(of.Context{TargetingKey: "u"})
		eq(t, "default", of.GetClient().String("flag", "def"), "def")
		eq(t, "flag key", p.log, []string{"resolve:flag"})
		eq(t, "context", p.ctx.TargetingKey, "u")
	}},
	"2.2.2.1": {test: func(t *testing.T) {
		var p of.Provider = memprovider.New()
		var _ func(string, bool, of.Context, of.Logger) of.ResolutionDetails[bool] = p.ResolveBoolean
		var _ func(string, string, of.Context, of.Logger) of.ResolutionDetails[string] = p.ResolveString
		var _ func(string, float64, of.Context, of.Logger) of.ResolutionDetails[float64] = p.ResolveNumber
		var _ func(string, any, of.Context, of.Logger) of.ResolutionDetails[any] = p.ResolveObject
	}},
	"2.2.3": {test: memNormal},
	"2.2.4": {test: memNormal},
	"2.2.5": {test: memNormal},
	"2.2.6": {test: memNormal},
	"2.2.7": {test: memAbnormal},
	"2.2.8.1": {test: func(t *testing.T) {
		var _ of.ResolutionDetails[bool] = of.NoopProvider().ResolveBoolean("f", true, of.Context{}, of.DefaultLogger())
		var _ of.ResolutionDetails[any] = of.NoopProvider().ResolveObject("f", nil, of.Context{}, of.DefaultLogger())
	}},
	"2.2.9":  {test: flagMetadata},
	"2.2.10": {test: flagMetadata},
	"2.3.1": {test: func(t *testing.T) {
		var log []string
		p := newFake("p")
		p.hooks = []of.Hook{stageHook(&log, "provider")}
		set(t, p)
		of.GetClient().Boolean("f", true)
		eq(t, "stages", log, []string{"provider.before", "provider.after", "provider.finally"})
	}},
	"2.3.2": {test: memNormal},
	"2.3.3": {test: memAbnormal},
	"2.4.1": {test: func(t *testing.T) {
		p := full("p")
		of.SetContext(of.Context{TargetingKey: "u"})
		set(t, p)
		eq(t, "context", p.initCtx.TargetingKey, "u")
		eq(t, "domain", p.domain, "")
		set(t, newFake("no-initialize"), of.ForDomain("d"))
		eq(t, "optional", of.ProviderStatus("d"), of.StatusReady)
	}},
	"2.4.2.1": {test: initError},
	"2.4.3": {test: func(t *testing.T) {
		set(t, scoped{full("scoped")}, of.ForDomain("a"))
		p := full("shared")
		set(t, p, of.ForDomain("b"))
		set(t, p, of.ForDomain("c"))
		eq(t, "shared provider", of.ProviderMetadata("c").Name, "shared")
	}},
	"2.4.4": {test: func(t *testing.T) {
		p := scoped{full("p")}
		set(t, p, of.ForDomain("a"))
		eq(t, "domain", p.domain, "a")
	}},
	"2.5.1": {test: func(t *testing.T) {
		p := full("p")
		set(t, p)
		set(t, newFake("without shutdown"))
		eq(t, "shutdown calls", count(p.log, "close"), 1)
		of.Close()
	}},
	"2.5.2": na(author + "; the shipped Go providers define no shutdown"),
	"2.5.3": na(author + "; the shipped Go providers define no shutdown"),
	"2.6.1": {test: func(t *testing.T) {
		p := full("p")
		of.SetContext(of.Context{TargetingKey: "old"})
		set(t, p)
		of.SetContext(of.Context{TargetingKey: "new"})
		synctest.Wait()
		eq(t, "old", p.oldCtx.TargetingKey, "old")
		eq(t, "new", p.newCtx.TargetingKey, "new")
		set(t, newFake("without on-context-changed"))
		of.SetContext(of.Context{TargetingKey: "x"})
		synctest.Wait()
		eq(t, "optional", of.ProviderStatus(), of.StatusReady)
	}},
	"2.7.1": {test: track},
	"2.8.1": {test: selfEmitting},
	"2.8.2": {test: selfEmitting},
	"2.8.3": {test: selfEmitting},
	"2.8.4": {test: selfEmitting},
	"2.8.5.1": {test: func(t *testing.T) {
		var ready []string
		of.AddHandler(of.EventReady, func(d of.EventDetails) { ready = append(ready, d.ProviderName) })
		if err := of.SetProvider(newFake("p")); err != nil {
			t.Fatal(err)
		}
		eq(t, "status at registration", of.ProviderStatus(), of.StatusReady)
		synctest.Wait()
		eq(t, "ready handlers", ready, []string{"p"})
	}},

	"3.1.1": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		of.GetClient().Boolean("f", true)
		eq(t, "absent", p.ctx.TargetingKey, "")
		of.SetContext(of.Context{TargetingKey: "user-1"})
		of.GetClient().Boolean("f", true)
		eq(t, "set", p.ctx.TargetingKey, "user-1")
	}},
	"3.1.2": {test: contextFields},
	"3.1.3": {test: contextFields},
	"3.1.4": {test: func(t *testing.T) {
		c := of.Context{Attributes: map[string]any{"k": 1}}
		c.Attributes["k"] = 2
		of.SetContext(c)
		eq(t, "one value per key", of.GetContext().Attributes, map[string]any{"k": 2})
	}},
	"3.2.1.1": na(dynamic),
	"3.2.2.1": {test: func(t *testing.T) {
		of.SetContext(of.Context{TargetingKey: "u"})
		eq(t, "global", of.GetContext().TargetingKey, "u")
	}},
	"3.2.2.2": {test: func(t *testing.T) {
		ctx := reflect.TypeFor[of.Context]()
		c := reflect.TypeFor[*of.Client]()
		for i := range c.NumMethod() {
			m := c.Method(i)
			for j := range m.Type.NumIn() {
				if in := m.Type.In(j); in == ctx || in.Kind() == reflect.Slice && in.Elem() == ctx {
					t.Errorf("Client.%s accepts an evaluation context", m.Name)
				}
			}
		}
		for _, name := range exportedAPI(t) {
			if strings.HasPrefix(name, "With") && strings.Contains(name, "Context") && name != "WithContext" {
				t.Errorf("%s looks like an invocation context option", name)
			}
		}
		var _ of.SetOption = of.WithContext(of.Context{})
	}},
	"3.2.2.3": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p, of.ForDomain("d"))
		of.SetContext(of.Context{TargetingKey: "global"})
		of.SetContext(of.Context{TargetingKey: "domain"}, "d")
		eq(t, "domain", of.GetContext("d").TargetingKey, "domain")
		eq(t, "global", of.GetContext().TargetingKey, "global")
		of.GetClient("d").Boolean("f", true)
		eq(t, "evaluation", p.ctx.TargetingKey, "domain")
	}},
	"3.2.2.4": {test: func(t *testing.T) {
		of.SetContext(of.Context{TargetingKey: "global"})
		of.SetContext(of.Context{TargetingKey: "domain"}, "d")
		eq(t, "get", of.GetContext("d").TargetingKey, "domain")
		of.ClearContext("d")
		eq(t, "cleared", of.GetContext("d").TargetingKey, "global")
		set(t, newFake("p"), of.ForDomain("d"))
		of.SetContext(of.Context{TargetingKey: "again"}, "d")
		of.ClearContexts()
		eq(t, "all cleared", of.GetContext("d"), of.Context{})
	}},
	"3.2.3": {test: apiContextOnly},
	"3.2.4.1": {test: func(t *testing.T) {
		p, q := full("p"), full("q")
		set(t, p)
		set(t, q, of.ForDomain("d"))
		of.SetContext(of.Context{TargetingKey: "u"})
		synctest.Wait()
		eq(t, "default provider", count(p.log, "change"), 1)
		eq(t, "domain provider without own context", count(q.log, "change"), 1)
		eq(t, "new context", p.newCtx.TargetingKey, "u")
	}},
	"3.2.4.2": {test: func(t *testing.T) {
		p, q, r := full("p"), full("q"), full("r")
		set(t, p)
		set(t, q, of.ForDomain("d"))
		set(t, r, of.ForDomain("e"))
		of.SetContext(of.Context{TargetingKey: "u"}, "d")
		synctest.Wait()
		eq(t, "associated", count(q.log, "change"), 1)
		eq(t, "default", count(p.log, "change"), 0)
		eq(t, "other domain", count(r.log, "change"), 0)
	}},
	"3.3.1.1":   na(dynamic),
	"3.3.1.2.1": na(dynamic),
	"3.3.1.2.2": na(dynamic),
	"3.3.1.2.3": na(dynamic),
	"3.3.2.1": {test: func(t *testing.T) {
		for _, name := range exportedAPI(t) {
			if l := strings.ToLower(name); strings.Contains(l, "transaction") || strings.Contains(l, "propagator") {
				t.Errorf("%s looks like a transaction context propagator", name)
			}
		}
	}},

	"4.1.1":   {test: hookContext},
	"4.1.2":   {test: hookContext},
	"4.1.3":   {test: hookContextImmutable},
	"4.1.4.1": na(dynamic),
	"4.1.5":   {test: hookData},
	"4.2.1":   {test: hints},
	"4.2.2.1": {test: hints},
	"4.2.2.2": {test: hookContextImmutable},
	"4.2.2.3": {test: hookContextImmutable},
	"4.3.1":   na(author + "; a Hook with no stage is accepted and does nothing"),
	"4.3.2":   {test: hookData},
	"4.3.2.1": na(dynamic),
	"4.3.3.1": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		var hints of.HookHints
		var key string
		h := of.Hook{Before: func(hc of.HookContext, hh of.HookHints) error {
			p.log = append(p.log, "before")
			key, hints = hc.FlagKey, hh
			return nil
		}}
		of.GetClient().Boolean("f", true, of.WithHooks(h), of.WithHookHints(of.NewHookHints(map[string]any{"k": 1})))
		eq(t, "order", p.log, []string{"before", "resolve:f"})
		eq(t, "hook context", key, "f")
		eq(t, "hints", hints.Value("k"), any(1))
	}},
	"4.3.4": na(staticOnly),
	"4.3.5": na(staticOnly),
	"4.3.6": {test: func(t *testing.T) {
		p, c := withOut(t, of.ResolutionDetails[any]{Value: true, Variant: "on"})
		var got of.EvaluationDetails[any]
		var hints of.HookHints
		h := of.Hook{After: func(hc of.HookContext, d of.EvaluationDetails[any], hh of.HookHints) error {
			p.log = append(p.log, "after:"+hc.FlagKey)
			got, hints = d, hh
			return nil
		}}
		c.Boolean("f", false, of.WithHooks(h), of.WithHookHints(of.NewHookHints(map[string]any{"k": 1})))
		eq(t, "order", p.log, []string{"resolve:f", "after:f"})
		eq(t, "details value", got.Value, any(true))
		eq(t, "details variant", got.Variant, "on")
		eq(t, "details flag key", got.FlagKey, "f")
		eq(t, "hints", hints.Value("k"), any(1))
	}},
	"4.3.7": {test: func(t *testing.T) {
		boom := errors.New("boom")
		run := func(name string, p *fake, hooks ...of.Hook) {
			of.Reset()
			set(t, p)
			var got error
			var key string
			var hints of.HookHints
			eh := of.Hook{Error: func(hc of.HookContext, err error, hh of.HookHints) { got, key, hints = err, hc.FlagKey, hh }}
			of.GetClient().Boolean("f", true, of.WithHooks(append(hooks, eh)...), of.WithHookHints(of.NewHookHints(map[string]any{"k": 1})))
			if got == nil || got.Error() != "boom" {
				t.Errorf("%s: error hook got %v", name, got)
			}
			eq(t, name+" hook context", key, "f")
			eq(t, name+" hints", hints.Value("k"), any(1))
		}
		run("before", newFake("p"), of.Hook{Before: func(of.HookContext, of.HookHints) error { return boom }})
		run("after", newFake("p"), of.Hook{After: func(of.HookContext, of.EvaluationDetails[any], of.HookHints) error { return boom }})
		p := newFake("p")
		p.panicErr = boom
		run("resolution", p)
	}},
	"4.3.8": {test: func(t *testing.T) {
		var log []string
		var got of.EvaluationDetails[any]
		var hints of.HookHints
		h := stageHook(&log, "h")
		h.Finally = func(hc of.HookContext, d of.EvaluationDetails[any], hh of.HookHints) {
			log = append(log, "h.finally:"+hc.FlagKey)
			got, hints = d, hh
		}
		_, c := withOut(t, of.ResolutionDetails[any]{Value: true})
		c.Boolean("f", false, of.WithHooks(h), of.WithHookHints(of.NewHookHints(map[string]any{"k": 1})))
		eq(t, "success", log, []string{"h.before", "h.after", "h.finally:f"})
		eq(t, "details", got.Value, any(true))
		eq(t, "hints", hints.Value("k"), any(1))
		log = nil
		fail := of.Hook{After: func(of.HookContext, of.EvaluationDetails[any], of.HookHints) error { return errors.New("boom") }}
		c.Boolean("f", false, of.WithHooks(h, fail))
		eq(t, "failure", log, []string{"h.before", "h.error", "h.finally:f"})
		eq(t, "failure details", got.ErrorCode, of.CodeGeneral)
	}},
	"4.3.9.1": na("the condition does not hold: the stage is the exported field Hook.Finally, which is not a reserved word in Go"),
	"4.4.1":   {test: hookOrder},
	"4.4.2":   {test: hookOrder},
	"4.4.3": {test: func(t *testing.T) {
		var log []string
		bad := stageHook(&log, "bad")
		bad.Finally = func(of.HookContext, of.EvaluationDetails[any], of.HookHints) {
			log = append(log, "bad.finally")
			panic("boom")
		}
		_, c := withOut(t, of.ResolutionDetails[any]{Value: true})
		v := c.Boolean("f", false, of.WithHooks(stageHook(&log, "a"), bad, stageHook(&log, "c")))
		eq(t, "value", v, true)
		eq(t, "finally stages", log[len(log)-3:], []string{"c.finally", "bad.finally", "a.finally"})
	}},
	"4.4.4": {test: func(t *testing.T) {
		var log []string
		bad := stageHook(&log, "bad")
		bad.Error = func(of.HookContext, error, of.HookHints) { log = append(log, "bad.error"); panic("boom") }
		fail := of.Hook{Before: func(of.HookContext, of.HookHints) error { return errors.New("x") }}
		v := of.GetClient().Boolean("f", true, of.WithHooks(stageHook(&log, "a"), bad, stageHook(&log, "c"), fail))
		eq(t, "value", v, true)
		eq(t, "stages", log, []string{"a.before", "bad.before", "c.before", "c.error", "bad.error", "a.error", "c.finally", "bad.finally", "a.finally"})
	}},
	"4.4.5": {test: hookErrors},
	"4.4.6": {test: hookErrors},
	"4.4.7": {test: hookErrors},
	"4.5.1": {test: hints},
	"4.5.2": {test: hints},
	"4.5.3": {test: hints},
	"4.6.1": {test: hookData},

	"5.1.1": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		var got []of.Event
		for _, e := range []of.Event{of.EventReady, of.EventError, of.EventConfigurationChanged, of.EventStale, of.EventReconciling, of.EventContextChanged} {
			of.GetClient().AddHandler(e, func(of.EventDetails) { got = append(got, e) })
		}
		got = nil
		want := []of.Event{of.EventError, of.EventConfigurationChanged, of.EventStale, of.EventReconciling, of.EventContextChanged, of.EventReady}
		for _, e := range want {
			p.ev.Emit(e, of.EventDetails{})
		}
		eq(t, "events", got, want)
	}},
	"5.1.2": {test: routing},
	"5.1.3": {test: routing},
	"5.1.4": {test: initError},
	"5.1.5": {test: errorCodes},
	"5.2.1": {test: routing},
	"5.2.2": {test: routing},
	"5.2.3": {test: routing},
	"5.2.4": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		var got of.EventDetails
		var h of.Handler = func(d of.EventDetails) { got = d }
		of.AddHandler(of.EventConfigurationChanged, h)
		p.ev.Emit(of.EventConfigurationChanged, of.EventDetails{FlagsChanged: []string{"f"}, Message: "m", Metadata: map[string]any{"k": "v"}})
		eq(t, "flags changed", got.FlagsChanged, []string{"f"})
		eq(t, "message", got.Message, "m")
		eq(t, "metadata", got.Metadata, map[string]any{"k": "v"})
	}},
	"5.2.5": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		var ran []string
		of.AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "a"); panic("boom") })
		of.AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "b") })
		of.GetClient().AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "c"); panic("boom") })
		of.GetClient().AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "d") })
		p.ev.Emit(of.EventStale, of.EventDetails{})
		eq(t, "handlers run", len(ran), 4)
	}},
	"5.2.6": {test: func(t *testing.T) {
		var api, client []string
		of.AddHandler(of.EventConfigurationChanged, func(d of.EventDetails) { api = append(api, d.ProviderName) })
		of.GetClient().AddHandler(of.EventConfigurationChanged, func(d of.EventDetails) { client = append(client, d.ProviderName) })
		a, b := newFake("a"), newFake("b")
		set(t, a)
		a.ev.Emit(of.EventConfigurationChanged, of.EventDetails{})
		set(t, b)
		b.ev.Emit(of.EventConfigurationChanged, of.EventDetails{})
		eq(t, "api handler", api, []string{"a", "b"})
		eq(t, "client handler", client, []string{"a", "b"})
	}},
	"5.2.7": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		var ran []string
		removeAPI := of.AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "api") })
		removeClient := of.GetClient().AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "client") })
		of.AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "kept") })
		removeAPI()
		removeClient()
		p.ev.Emit(of.EventStale, of.EventDetails{})
		eq(t, "handlers run", ran, []string{"kept"})
	}},
	"5.3.1": {test: func(t *testing.T) {
		var ran []string
		of.AddHandler(of.EventReady, func(of.EventDetails) { ran = append(ran, "api") })
		of.GetClient().AddHandler(of.EventReady, func(of.EventDetails) { ran = append(ran, "client") })
		p := full("p")
		set(t, p)
		eq(t, "after initialize", len(ran), 2)
		p.ev.Emit(of.EventReady, of.EventDetails{})
		eq(t, "after emitted READY", len(ran), 4)
	}},
	"5.3.2": {test: initError},
	"5.3.3": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		var ran []string
		of.AddHandler(of.EventReady, func(of.EventDetails) { ran = append(ran, "api ready") })
		of.GetClient().AddHandler(of.EventReady, func(of.EventDetails) { ran = append(ran, "client ready") })
		of.AddHandler(of.EventError, func(of.EventDetails) { ran = append(ran, "api error") })
		synctest.Wait()
		eq(t, "ready", ran, []string{"api ready", "client ready"})
		p.ev.Emit(of.EventStale, of.EventDetails{})
		of.GetClient().AddHandler(of.EventStale, func(of.EventDetails) { ran = append(ran, "client stale") })
		synctest.Wait()
		eq(t, "stale", ran[2:], []string{"client stale"})
	}},
	"5.3.4.1": {test: reconciling},
	"5.3.4.2": {test: reconciling},
	"5.3.4.3": {test: func(t *testing.T) {
		p := full("p")
		set(t, p)
		var errs []string
		var changed int
		of.AddHandler(of.EventError, func(d of.EventDetails) { errs = append(errs, d.Message) })
		of.AddHandler(of.EventContextChanged, func(of.EventDetails) { changed++ })
		first, second := make(chan error), make(chan error)
		next := []chan error{first, second}
		p.change = func() (of.Pending, error) { ch := next[0]; next = next[1:]; return ch, nil }
		go of.SetContext(of.Context{TargetingKey: "a"})
		synctest.Wait()
		go of.SetContext(of.Context{TargetingKey: "b"})
		synctest.Wait()
		first <- errors.New("first failed")
		synctest.Wait()
		eq(t, "handlers while another invocation is running", len(errs)+changed, 0)
		second <- errors.New("second failed")
		synctest.Wait()
		eq(t, "error handlers", len(errs), 1)
		eq(t, "context changed handlers", changed, 0)
		eq(t, "status", of.ProviderStatus(), of.StatusError)
		if len(errs) == 1 && !strings.Contains(errs[0], "second failed") {
			t.Errorf("error message %q", errs[0])
		}
	}},
	"5.3.5": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		seen := map[of.Event]of.Status{}
		for _, e := range []of.Event{of.EventReady, of.EventError, of.EventStale} {
			of.GetClient().AddHandler(e, func(of.EventDetails) { seen[e] = of.GetClient().ProviderStatus() })
		}
		clear(seen)
		p.ev.Emit(of.EventStale, of.EventDetails{})
		p.ev.Emit(of.EventError, of.EventDetails{})
		p.ev.Emit(of.EventReady, of.EventDetails{})
		eq(t, "status seen by handlers", seen, map[of.Event]of.Status{
			of.EventStale: of.StatusStale, of.EventError: of.StatusError, of.EventReady: of.StatusReady,
		})
	}},

	"6.1.1.1": na(dynamic),
	"6.1.2.1": {test: track},
	"6.1.3":   {test: apiContextOnly},
	"6.1.4": {test: func(t *testing.T) {
		p := newFake("p")
		set(t, p)
		of.GetClient().Track("event", of.TrackingDetails{})
		eq(t, "provider untouched", len(p.log), 0)
	}},
	"6.2.1": {test: track},
	"6.2.2": {test: track},
}

func abnormal(t *testing.T) {
	_, c := withOut(t, of.ResolutionDetails[any]{Value: true, ErrorCode: of.CodeFlagNotFound, ErrorMessage: "no such flag"})
	d := c.BooleanDetails("f", false)
	eq(t, "provider error: value", d.Value, false)
	eq(t, "provider error: code", d.ErrorCode, of.CodeFlagNotFound)
	eq(t, "provider error: reason", d.Reason, of.ReasonError)
	eq(t, "provider error: message", d.ErrorMessage, "no such flag")

	p := newFake("p")
	p.panicErr = errors.New("boom")
	set(t, p)
	eq(t, "boolean", c.Boolean("f", true), true)
	eq(t, "string", c.String("f", "def"), "def")
	eq(t, "number", c.Number("f", 1.5), 1.5)
	eq(t, "int", c.Int("f", 2), int64(2))
	eq(t, "object", c.Object("f", "def"), any("def"))
	d = c.BooleanDetails("f", true)
	eq(t, "panic: value", d.Value, true)
	eq(t, "panic: code", d.ErrorCode, of.CodeGeneral)
	eq(t, "panic: reason", d.Reason, of.ReasonError)
	eq(t, "panic: message", d.ErrorMessage, "boom")

	p.panicErr = nil
	of.SetLogger(&recLogger{})
	bad := of.Hook{
		Before:  func(of.HookContext, of.HookHints) error { panic("before") },
		Error:   func(of.HookContext, error, of.HookHints) { panic("error") },
		Finally: func(of.HookContext, of.EvaluationDetails[any], of.HookHints) { panic("finally") },
	}
	eq(t, "panicking hooks", c.String("f", "def", of.WithHooks(bad)), "def")
	c.Track("event")
	c.AddHandler(of.EventReady, func(of.EventDetails) { panic("handler") })
}

func flagMetadata(t *testing.T) {
	md := of.FlagMetadata{"s": "v", "n": 1.5, "b": true}
	p, c := withOut(t, of.ResolutionDetails[any]{FlagMetadata: md})
	eq(t, "set", c.BooleanDetails("f", false).FlagMetadata, md)
	p.out.FlagMetadata = nil
	got := c.BooleanDetails("f", false).FlagMetadata
	if got == nil || len(got) != 0 {
		t.Errorf("unset: got %#v, want an empty record", got)
	}
	p.out.ErrorCode = of.CodeGeneral
	got = c.BooleanDetails("f", false).FlagMetadata
	if got == nil || len(got) != 0 {
		t.Errorf("error: got %#v, want an empty record", got)
	}
}

func reconciling(t *testing.T) {
	p := full("p")
	set(t, p)
	ch := make(chan error)
	p.change = func() (of.Pending, error) { return ch, nil }
	var events []of.Event
	for _, e := range []of.Event{of.EventReconciling, of.EventContextChanged, of.EventError} {
		of.AddHandler(e, func(of.EventDetails) { events = append(events, e) })
		of.GetClient().AddHandler(e, func(of.EventDetails) { events = append(events, e) })
	}
	go of.SetContext(of.Context{TargetingKey: "u"})
	synctest.Wait()
	eq(t, "status while running", of.GetClient().ProviderStatus(), of.StatusReconciling)
	eq(t, "handlers while running", events, []of.Event{of.EventReconciling, of.EventReconciling})
	eq(t, "evaluation while running", of.GetClient().BooleanDetails("f", true).ErrorCode, of.ErrorCode(""))
	close(ch)
	synctest.Wait()
	eq(t, "status after", of.GetClient().ProviderStatus(), of.StatusReady)
	eq(t, "handlers after", events[2:], []of.Event{of.EventContextChanged, of.EventContextChanged})
}

func initError(t *testing.T) {
	var got []of.EventDetails
	of.AddHandler(of.EventError, func(d of.EventDetails) { got = append(got, d) })
	of.GetClient().AddHandler(of.EventError, func(d of.EventDetails) { got = append(got, d) })
	p := full("p")
	p.init = failed(of.NewError(of.CodeParseError, "bad config"))
	if err := of.SetProviderAndWait(p); err == nil || err.Error() != "bad config" {
		t.Errorf("SetProviderAndWait: %v", err)
	}
	synctest.Wait()
	eq(t, "status after initialize", of.GetClient().ProviderStatus(), of.StatusError)
	if len(got) != 2 {
		t.Fatalf("error handlers: got %d runs, want 2", len(got))
	}
	for _, d := range got {
		eq(t, "provider name", d.ProviderName, "p")
		eq(t, "error message", d.Message, "bad config")
	}
	p.ev.Emit(of.EventReady, of.EventDetails{})
	p.ev.Emit(of.EventError, of.EventDetails{Message: "lost connection", ErrorCode: of.CodeGeneral})
	eq(t, "status after emitted ERROR", of.GetClient().ProviderStatus(), of.StatusError)
	eq(t, "emitted message", got[len(got)-1].Message, "lost connection")
	eq(t, "emitted code", got[len(got)-1].ErrorCode, of.CodeGeneral)
}

func memProvider(t *testing.T) *of.Client {
	t.Helper()
	set(t, memprovider.New(memprovider.Flag{
		Key: "flag", Variants: map[string]any{"on": true, "off": false}, DefaultVariant: "on",
	}))
	return of.GetClient()
}

func memNormal(t *testing.T) {
	d := memProvider(t).BooleanDetails("flag", false)
	eq(t, "value", d.Value, true)
	eq(t, "variant", d.Variant, "on")
	eq(t, "reason", d.Reason, of.ReasonStatic)
	eq(t, "error code", d.ErrorCode, of.ErrorCode(""))
	eq(t, "error message", d.ErrorMessage, "")
}

func memAbnormal(t *testing.T) {
	d := memProvider(t).BooleanDetails("missing", true)
	eq(t, "value", d.Value, true)
	eq(t, "error code", d.ErrorCode, of.CodeFlagNotFound)
	if d.ErrorMessage == "" {
		t.Error("no error message")
	}
	p := newFake("p")
	p.panicErr = of.NewError(of.CodeInvalidContext, "bad context")
	set(t, p)
	d = of.GetClient().BooleanDetails("f", true)
	eq(t, "idiomatic error: code", d.ErrorCode, of.CodeInvalidContext)
	eq(t, "idiomatic error: message", d.ErrorMessage, "bad context")
}

func track(t *testing.T) {
	p := full("p")
	set(t, p)
	of.SetContext(of.Context{TargetingKey: "u"})
	synctest.Wait()
	c := of.GetClient()
	c.Track("plain")
	eq(t, "name", p.trackName, "plain")
	eq(t, "no value", p.trackDetails.Value, (*float64)(nil))
	v := 9.99
	attrs := map[string]any{"b": true, "s": "x", "n": 2.0, "o": map[string]any{"k": "v"}}
	c.Track("purchase", of.TrackingDetails{Value: &v, Attributes: attrs})
	eq(t, "name", p.trackName, "purchase")
	eq(t, "context", p.trackCtx.TargetingKey, "u")
	if p.trackDetails.Value == nil || *p.trackDetails.Value != 9.99 {
		t.Errorf("value: %v", p.trackDetails.Value)
	}
	eq(t, "custom fields", p.trackDetails.Attributes, attrs)
}

func apiContextOnly(t *testing.T) {
	p := full("p")
	set(t, p)
	of.SetContext(of.Context{TargetingKey: "u", Attributes: map[string]any{"k": "api"}})
	synctest.Wait()
	want := of.Context{TargetingKey: "u", Attributes: map[string]any{"k": "api"}}
	of.GetClient().Boolean("f", true)
	eq(t, "evaluation context", p.ctx, want)
	of.GetClient().Track("event")
	eq(t, "tracking context", p.trackCtx, want)
}

func contextFields(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	attrs := map[string]any{
		"bool": true, "string": "s", "number": 1.5, "datetime": when,
		"structure": map[string]any{"list": []any{1.0, "x"}},
	}
	p := newFake("p")
	set(t, p)
	of.SetContext(of.Context{TargetingKey: "u", Attributes: attrs})
	synctest.Wait()
	of.GetClient().Boolean("f", true)
	eq(t, "provider sees all pairs", p.ctx.Attributes, attrs)
	got := of.GetContext()
	eq(t, "by key", got.Attributes["datetime"], any(when))
	keys := 0
	for range got.Attributes {
		keys++
	}
	eq(t, "all pairs", keys, len(attrs))
}

func hookContext(t *testing.T) {
	p := newFake("p")
	set(t, p, of.ForDomain("d"))
	of.SetContext(of.Context{TargetingKey: "u"})
	synctest.Wait()
	var hc of.HookContext
	h := of.Hook{Before: func(c of.HookContext, _ of.HookHints) error { hc = c; return nil }}
	of.GetClient("d").Number("flag", 1.5, of.WithHooks(h))
	eq(t, "flag key", hc.FlagKey, "flag")
	eq(t, "flag value type", hc.FlagType, of.TypeNumber)
	eq(t, "default value", hc.DefaultValue, any(1.5))
	eq(t, "evaluation context", hc.Context.TargetingKey, "u")
	eq(t, "client metadata", hc.ClientMetadata.Domain, "d")
	eq(t, "provider metadata", hc.ProviderMetadata.Name, "p")
	if hc.HookData == nil {
		t.Error("no hook data")
	}
}

func hookContextImmutable(t *testing.T) {
	set(t, newFake("p"), of.ForDomain("d"))
	var seen []of.HookContext
	mutate := func(hc of.HookContext) {
		seen = append(seen, hc)
		hc.FlagKey, hc.FlagType, hc.DefaultValue = "changed", of.TypeString, "changed"
		hc.ClientMetadata.Domain, hc.ProviderMetadata.Name = "changed", "changed"
	}
	h := of.Hook{
		Before:  func(hc of.HookContext, _ of.HookHints) error { mutate(hc); return nil },
		After:   func(hc of.HookContext, _ of.EvaluationDetails[any], _ of.HookHints) error { mutate(hc); return nil },
		Finally: func(hc of.HookContext, _ of.EvaluationDetails[any], _ of.HookHints) { mutate(hc) },
	}
	d := of.GetClient("d").BooleanDetails("flag", true, of.WithHooks(h, h))
	eq(t, "stages", len(seen), 6)
	for _, hc := range seen {
		eq(t, "flag key", hc.FlagKey, "flag")
		eq(t, "flag type", hc.FlagType, of.TypeBoolean)
		eq(t, "default value", hc.DefaultValue, any(true))
		eq(t, "client metadata", hc.ClientMetadata.Domain, "d")
		eq(t, "provider metadata", hc.ProviderMetadata.Name, "p")
	}
	eq(t, "result flag key", d.FlagKey, "flag")
	eq(t, "client metadata afterwards", of.GetClient("d").Metadata().Domain, "d")
}

func hookData(t *testing.T) {
	type custom struct{ n int }
	var got []string
	mk := func(name string, value any) of.Hook {
		return of.Hook{
			Before: func(hc of.HookContext, _ of.HookHints) error {
				if hc.HookData.Has("v") {
					t.Errorf("%s: hook data not fresh", name)
				}
				hc.HookData.Set("v", value)
				return nil
			},
			After: func(hc of.HookContext, _ of.EvaluationDetails[any], _ of.HookHints) error {
				got = append(got, fmt.Sprintf("%s.after=%v", name, hc.HookData.Get("v")))
				hc.HookData.Set("v", "changed in after")
				return nil
			},
			Finally: func(hc of.HookContext, _ of.EvaluationDetails[any], _ of.HookHints) {
				got = append(got, fmt.Sprintf("%s.finally=%v", name, hc.HookData.Get("v")))
			},
		}
	}
	a, b := mk("a", custom{1}), mk("b", []string{"x"})
	set(t, newFake("p"))
	c := of.GetClient()
	for range 2 {
		got = nil
		c.Boolean("f", true, of.WithHooks(a, b))
		eq(t, "stages", got, []string{"b.after=[x]", "a.after={1}", "b.finally=changed in after", "a.finally=changed in after"})
	}
}

func hints(t *testing.T) {
	when := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	src := map[string]any{"bool": true, "string": "s", "number": 1.5, "datetime": when, "structure": map[string]any{"k": "v"}}
	hh := of.NewHookHints(src)
	src["string"] = "changed by caller"
	var seen []of.HookHints
	see := func(h of.HookHints) {
		h.Map()["string"] = "changed by hook"
		seen = append(seen, h)
	}
	h := of.Hook{
		Before: func(_ of.HookContext, h of.HookHints) error { see(h); return nil },
		After: func(_ of.HookContext, _ of.EvaluationDetails[any], h of.HookHints) error {
			see(h)
			return nil
		},
		Error:   func(_ of.HookContext, _ error, h of.HookHints) { see(h) },
		Finally: func(_ of.HookContext, _ of.EvaluationDetails[any], h of.HookHints) { see(h) },
	}
	p := newFake("p")
	set(t, p)
	of.AddHooks(h)
	of.GetClient().Boolean("f", true, of.WithHooks(h), of.WithHookHints(hh))
	p.panicErr = errors.New("boom")
	of.GetClient().Boolean("f", true, of.WithHooks(h), of.WithHookHints(hh))
	eq(t, "stages given hints", len(seen), 12)
	for _, s := range seen {
		eq(t, "string", s.Value("string"), any("s"))
		eq(t, "datetime", s.Value("datetime"), any(when))
		eq(t, "structure", s.Value("structure"), any(map[string]any{"k": "v"}))
		eq(t, "all", len(s.Map()), 5)
	}
}

func hookOrder(t *testing.T) {
	var log []string
	p := newFake("p")
	p.hooks = []of.Hook{stageHook(&log, "provider1"), stageHook(&log, "provider2")}
	set(t, p)
	of.AddHooks(stageHook(&log, "api1"), stageHook(&log, "api2"))
	c := of.GetClient()
	c.AddHooks(stageHook(&log, "client"))
	c.Boolean("f", true, of.WithHooks(stageHook(&log, "invocation")))
	names := []string{"api1", "api2", "client", "invocation", "provider1", "provider2"}
	var want []string
	for _, n := range names {
		want = append(want, n+".before")
	}
	for _, stage := range []string{".after", ".finally"} {
		for i := len(names) - 1; i >= 0; i-- {
			want = append(want, names[i]+stage)
		}
	}
	eq(t, "success", log, want)

	log = nil
	p.panicErr = errors.New("boom")
	c.Boolean("f", true, of.WithHooks(stageHook(&log, "invocation")))
	want = nil
	for _, n := range names {
		want = append(want, n+".before")
	}
	for _, stage := range []string{".error", ".finally"} {
		for i := len(names) - 1; i >= 0; i-- {
			want = append(want, names[i]+stage)
		}
	}
	eq(t, "failure", log, want)
}

func hookErrors(t *testing.T) {
	var log []string
	p, c := withOut(t, of.ResolutionDetails[any]{Value: "resolved"})
	failBefore := stageHook(&log, "bad")
	failBefore.Before = func(of.HookContext, of.HookHints) error { log = append(log, "bad.before"); return errors.New("x") }
	d := c.StringDetails("f", "def", of.WithHooks(stageHook(&log, "a"), failBefore, stageHook(&log, "c")))
	eq(t, "before: stages", log, []string{"a.before", "bad.before", "c.error", "bad.error", "a.error", "c.finally", "bad.finally", "a.finally"})
	eq(t, "before: value", d.Value, "def")
	eq(t, "before: resolution skipped", len(p.log), 0)

	log = nil
	failAfter := stageHook(&log, "bad")
	failAfter.After = func(of.HookContext, of.EvaluationDetails[any], of.HookHints) error {
		log = append(log, "bad.after")
		return errors.New("x")
	}
	d = c.StringDetails("f", "def", of.WithHooks(stageHook(&log, "a"), failAfter, stageHook(&log, "c")))
	eq(t, "after: stages", log, []string{
		"a.before", "bad.before", "c.before", "c.after", "bad.after",
		"c.error", "bad.error", "a.error", "c.finally", "bad.finally", "a.finally",
	})
	eq(t, "after: value", d.Value, "def")
}

func routing(t *testing.T) {
	a, b := newFake("a"), newFake("b")
	set(t, a)
	set(t, b, of.ForDomain("d"))
	var got []string
	on := func(who string) of.Handler {
		return func(d of.EventDetails) { got = append(got, who+"<-"+d.ProviderName) }
	}
	of.AddHandler(of.EventConfigurationChanged, on("api"))
	of.GetClient().AddHandler(of.EventConfigurationChanged, on("default"))
	of.GetClient("d").AddHandler(of.EventConfigurationChanged, on("d"))
	of.GetClient("unbound").AddHandler(of.EventConfigurationChanged, on("unbound"))
	a.ev.Emit(of.EventConfigurationChanged, of.EventDetails{})
	slices.Sort(got)
	eq(t, "default provider", got, []string{"api<-a", "default<-a", "unbound<-a"})
	got = nil
	b.ev.Emit(of.EventConfigurationChanged, of.EventDetails{})
	slices.Sort(got)
	eq(t, "domain provider", got, []string{"api<-b", "d<-b"})
}

func watch(a *of.API, evs ...of.Event) *[]string {
	var got []string
	for _, e := range evs {
		a.AddHandler(e, func(d of.EventDetails) { got = append(got, "api:"+string(e)+":"+string(d.ErrorCode)) })
		a.GetClient().AddHandler(e, func(d of.EventDetails) { got = append(got, "client:"+string(e)+":"+string(d.ErrorCode)) })
	}
	return &got
}

func take(got *[]string) []string {
	out := *got
	*got = nil
	slices.Sort(out)
	return out
}

func selfEmitting(t *testing.T) {
	a := isolated.New()
	defer a.Close()
	got := watch(a, of.EventReady, of.EventError, of.EventReconciling, of.EventContextChanged)

	silent := emitting{full("silent")}
	if err := a.SetProviderAndWait(silent); err != nil {
		t.Fatal(err)
	}
	synctest.Wait()
	eq(t, "status without an emitted event", a.ProviderStatus(), of.StatusNotReady)
	eq(t, "handlers without an emitted event", take(got), nil)

	p := emitting{full("p")}
	p.init = func() (of.Pending, error) {
		p.ev.Emit(of.EventReady, of.EventDetails{})
		return nil, nil
	}
	if err := a.SetProviderAndWait(p); err != nil {
		t.Fatal(err)
	}
	synctest.Wait()
	eq(t, "status after emitted READY", a.ProviderStatus(), of.StatusReady)
	eq(t, "ready handlers", take(got), []string{"api:PROVIDER_READY:", "client:PROVIDER_READY:"})

	ch := make(chan error)
	p.change = func() (of.Pending, error) {
		p.ev.Emit(of.EventReconciling, of.EventDetails{})
		return ch, nil
	}
	go a.SetContext(of.Context{TargetingKey: "u"})
	synctest.Wait()
	eq(t, "status while reconciling", a.ProviderStatus(), of.StatusReconciling)
	eq(t, "reconciling handlers", take(got), []string{"api:PROVIDER_RECONCILING:", "client:PROVIDER_RECONCILING:"})
	p.ev.Emit(of.EventContextChanged, of.EventDetails{})
	close(ch)
	synctest.Wait()
	eq(t, "status after CONTEXT_CHANGED", a.ProviderStatus(), of.StatusReady)
	eq(t, "context-changed handlers", take(got), []string{"api:PROVIDER_CONTEXT_CHANGED:", "client:PROVIDER_CONTEXT_CHANGED:"})

	p.change = func() (of.Pending, error) {
		p.ev.Emit(of.EventError, of.EventDetails{ErrorCode: of.CodeInvalidContext})
		return nil, errors.New("rejected")
	}
	a.SetContext(of.Context{TargetingKey: "v"})
	synctest.Wait()
	eq(t, "status after failed change", a.ProviderStatus(), of.StatusError)
	eq(t, "error handlers", take(got), []string{"api:PROVIDER_ERROR:INVALID_CONTEXT", "client:PROVIDER_ERROR:INVALID_CONTEXT"})

	q := emitting{full("q")}
	q.init = func() (of.Pending, error) {
		q.ev.Emit(of.EventError, of.EventDetails{ErrorCode: of.CodeProviderFatal})
		return nil, of.NewError(of.CodeProviderFatal, "dead")
	}
	if err := a.SetProviderAndWait(q); err == nil {
		t.Error("no error from SetProviderAndWait")
	}
	synctest.Wait()
	eq(t, "status after emitted fatal", a.ProviderStatus(), of.StatusFatal)
	eq(t, "fatal handlers", take(got), []string{"api:PROVIDER_ERROR:PROVIDER_FATAL", "client:PROVIDER_ERROR:PROVIDER_FATAL"})
}

func errorCodes(t *testing.T) {
	got := watch(nil, of.EventError)
	p := full("p")
	p.init = failed(of.NewError(of.CodeParseError, "bad config"))
	_ = of.SetProviderAndWait(p)
	synctest.Wait()
	eq(t, "pending failure", take(got), []string{"api:PROVIDER_ERROR:PARSE_ERROR", "client:PROVIDER_ERROR:PARSE_ERROR"})

	q := full("q")
	q.init = func() (of.Pending, error) { return nil, errors.New("plain") }
	if err := of.SetProviderAndWait(q); err == nil || err.Error() != "plain" {
		t.Errorf("SetProviderAndWait: %v", err)
	}
	synctest.Wait()
	eq(t, "returned failure is bound", of.GetProvider(), of.Provider(q))
	eq(t, "returned failure status", of.ProviderStatus(), of.StatusError)
	eq(t, "returned failure", take(got), []string{"api:PROVIDER_ERROR:GENERAL", "client:PROVIDER_ERROR:GENERAL"})

	r := full("r")
	set(t, r)
	r.change = func() (of.Pending, error) { return nil, of.NewError(of.CodeInvalidContext, "no") }
	of.SetContext(of.Context{TargetingKey: "u"})
	synctest.Wait()
	eq(t, "context change failure", take(got), []string{"api:PROVIDER_ERROR:INVALID_CONTEXT", "client:PROVIDER_ERROR:INVALID_CONTEXT"})

	r.ev.Emit(of.EventError, of.EventDetails{ErrorCode: of.CodeProviderFatal})
	eq(t, "emitted code passes through", take(got), []string{"api:PROVIDER_ERROR:PROVIDER_FATAL", "client:PROVIDER_ERROR:PROVIDER_FATAL"})
}

func shutdownResets(t *testing.T) {
	p, d := full("p"), full("d")
	set(t, p)
	set(t, d, of.ForDomain("d"))
	of.AddHooks(of.Hook{})
	of.SetContext(of.Context{TargetingKey: "global"})
	of.SetContext(of.Context{TargetingKey: "domain"}, "d")
	ran := 0
	of.AddHandler(of.EventStale, func(of.EventDetails) { ran++ })
	c, cd := of.GetClient(), of.GetClient("d")
	c.AddHandler(of.EventStale, func(of.EventDetails) { ran++ })
	cd.AddHandler(of.EventStale, func(of.EventDetails) { ran++ })
	eq(t, "status before", []of.Status{c.ProviderStatus(), cd.ProviderStatus()}, []of.Status{of.StatusReady, of.StatusReady})

	of.Close()

	eq(t, "shutdown calls", []int{count(p.log, "close"), count(d.log, "close")}, []int{1, 1})
	eq(t, "status after", []of.Status{c.ProviderStatus(), cd.ProviderStatus()}, []of.Status{of.StatusNotReady, of.StatusNotReady})
	eq(t, "providers", []of.Provider{of.GetProvider(), of.GetProvider("d")}, []of.Provider{of.NoopProvider(), of.NoopProvider()})
	eq(t, "hooks", len(of.Hooks()), 0)
	eq(t, "api handlers", len(of.Handlers(of.EventStale)), 0)
	eq(t, "client handlers", len(c.Handlers(of.EventStale))+len(cd.Handlers(of.EventStale)), 0)
	eq(t, "global context", of.GetContext().TargetingKey, "")
	eq(t, "domain context", of.GetContext("d").TargetingKey, "")
	p.ev.Emit(of.EventStale, of.EventDetails{})
	d.ev.Emit(of.EventStale, of.EventDetails{})
	eq(t, "handler runs after shutdown", ran, 0)
	eq(t, "evaluation after shutdown", c.BooleanDetails("f", true).ErrorCode, of.CodeProviderNotReady)
}

func isolatedInstances(t *testing.T) {
	a, b := isolated.New(), isolated.New()
	defer a.Close()
	defer b.Close()
	if a == b {
		t.Fatal("factory returned the same instance twice")
	}
	pa, pb, pg := newFake("a"), newFake("b"), newFake("global")
	pa.out = &of.ResolutionDetails[any]{Value: "from a"}
	pb.out = &of.ResolutionDetails[any]{Value: "from b"}
	set(t, pg)
	if err := a.SetProviderAndWait(pa); err != nil {
		t.Fatal(err)
	}
	if err := b.SetProviderAndWait(pb, of.ForDomain("d")); err != nil {
		t.Fatal(err)
	}
	synctest.Wait()
	eq(t, "providers", []string{a.ProviderMetadata().Name, b.ProviderMetadata("d").Name, b.ProviderMetadata().Name, of.ProviderMetadata().Name},
		[]string{"a", "b", "No-op Provider", "global"})
	eq(t, "evaluation", []string{a.GetClient().String("f", "x"), b.GetClient("d").String("f", "x"), of.GetClient().String("f", "x")},
		[]string{"from a", "from b", "x"})

	var log []string
	a.AddHooks(stageHook(&log, "a"))
	a.SetContext(of.Context{TargetingKey: "ctx-a"})
	ran := 0
	a.AddHandler(of.EventStale, func(of.EventDetails) { ran++ })
	a.GetClient().AddHandler(of.EventStale, func(of.EventDetails) { ran++ })
	eq(t, "hooks", []int{len(a.Hooks()), len(b.Hooks()), len(of.Hooks())}, []int{1, 0, 0})
	eq(t, "contexts", []string{a.GetContext().TargetingKey, b.GetContext().TargetingKey, of.GetContext().TargetingKey}, []string{"ctx-a", "", ""})
	eq(t, "handlers", []int{len(a.Handlers(of.EventStale)), len(b.Handlers(of.EventStale)), len(of.Handlers(of.EventStale))}, []int{1, 0, 0})
	pb.ev.Emit(of.EventStale, of.EventDetails{})
	pg.ev.Emit(of.EventStale, of.EventDetails{})
	eq(t, "other instances' events", ran, 0)
	pa.ev.Emit(of.EventStale, of.EventDetails{})
	eq(t, "own events", ran, 2)
	eq(t, "statuses", []of.Status{a.ProviderStatus(), b.ProviderStatus("d"), of.ProviderStatus()}, []of.Status{of.StatusStale, of.StatusStale, of.StatusStale})
	b.GetClient("d").Boolean("f", true)
	of.GetClient().Boolean("f", true)
	eq(t, "hooks of a on other instances", len(log), 0)

	a.Close()
	eq(t, "after shutting a down", []string{a.ProviderMetadata().Name, b.ProviderMetadata("d").Name, of.ProviderMetadata().Name},
		[]string{"No-op Provider", "b", "global"})
}

func sameContract(t *testing.T) {
	global := map[string]any{
		"SetProvider": of.SetProvider, "SetProviderAndWait": of.SetProviderAndWait, "GetProvider": of.GetProvider,
		"ProviderMetadata": of.ProviderMetadata, "ProviderStatus": of.ProviderStatus,
		"AddHooks": of.AddHooks, "Hooks": of.Hooks, "ClearHooks": of.ClearHooks, "SetLogger": of.SetLogger,
		"SetContext": of.SetContext, "GetContext": of.GetContext, "ClearContext": of.ClearContext, "ClearContexts": of.ClearContexts,
		"AddHandler": of.AddHandler, "Handlers": of.Handlers, "ClearHandlers": of.ClearHandlers,
		"GetClient": of.GetClient, "GetClientVersion": of.GetClientVersion,
		"Close": of.Close, "ClearProviders": of.ClearProviders,
	}
	inst := reflect.TypeOf(isolated.New())
	eq(t, "method count", inst.NumMethod(), len(global))
	for name, fn := range global {
		m, ok := inst.MethodByName(name)
		if !ok {
			t.Errorf("instance has no %s", name)
			continue
		}
		ft := reflect.TypeOf(fn)
		ins := make([]reflect.Type, ft.NumIn()+1)
		ins[0] = inst
		for i := range ft.NumIn() {
			ins[i+1] = ft.In(i)
		}
		outs := make([]reflect.Type, ft.NumOut())
		for i := range outs {
			outs[i] = ft.Out(i)
		}
		eq(t, name+" signature", m.Type, reflect.FuncOf(ins, outs, ft.IsVariadic()))
	}
	api := map[string]bool{}
	for _, n := range exportedAPI(t) {
		api[n] = true
	}
	for name := range global {
		if !api[name] {
			t.Errorf("%s is not exported by the package", name)
		}
	}
	eq(t, "client type", reflect.TypeOf(isolated.New().GetClient()), reflect.TypeOf(of.GetClient()))
}

func factoryElsewhere(t *testing.T) {
	eq(t, "factory package", reflect.TypeOf(isolated.New()).Elem().PkgPath()+"/isolated", "github.com/Toms223/WebGo/openfeature/isolated")
	pkgs, err := parser.ParseDir(token.NewFileSet(), ".", func(fi os.FileInfo) bool { return !strings.HasSuffix(fi.Name(), "_test.go") }, 0)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range pkgs["openfeature"].Files {
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || !fn.Name.IsExported() || fn.Recv != nil || fn.Type.Results == nil {
				continue
			}
			for _, r := range fn.Type.Results.List {
				if s, ok := r.Type.(*ast.StarExpr); ok {
					if id, ok := s.X.(*ast.Ident); ok && id.Name == "API" {
						t.Errorf("%s returns an API instance from the singleton's package", fn.Name.Name)
					}
				}
			}
		}
	}
	if _, err := os.Stat("isolated/isolated.go"); err != nil {
		t.Error(err)
	}
}

func oneInstancePerProvider(t *testing.T) {
	a := isolated.New()
	defer a.Close()
	p := full("p")
	set(t, p)
	set(t, p, of.ForDomain("d"))
	err := a.SetProviderAndWait(p)
	if err == nil {
		t.Fatal("second instance accepted a provider registered with the default instance")
	}
	eq(t, "error", err.Error(), "Provider 'p' is already registered with another API instance.")
	eq(t, "second instance provider", a.GetProvider(), of.NoopProvider())
	eq(t, "initialize calls", count(p.log, "initialize"), 1)
	of.ClearProviders()
	if err := a.SetProviderAndWait(p); err != nil {
		t.Fatalf("after release: %v", err)
	}
	if err := of.SetProviderAndWait(p); err == nil {
		t.Error("default instance accepted a provider registered with another instance")
	}
	a.Close()
	set(t, p)
}
