package openfeature

import (
	"errors"
	"math"
	"testing"
)

type numProvider struct {
	*testProvider
	num       float64
	gotCtx    Context
	gotLog    Logger
	panicWith any
	md        FlagMetadata
}

func (p *numProvider) ResolveNumber(_ string, _ float64, ctx Context, l Logger) ResolutionDetails[float64] {
	p.gotCtx, p.gotLog = ctx, l
	md := p.md
	if md == nil {
		md = FlagMetadata{"m": "x"}
	}
	return ResolutionDetails[float64]{Value: p.num, Variant: "v", Reason: ReasonStatic, FlagMetadata: md}
}

func (p *numProvider) ResolveBoolean(_ string, def bool, ctx Context, _ Logger) ResolutionDetails[bool] {
	if p.panicWith != nil {
		panic(p.panicWith)
	}
	p.gotCtx = ctx
	return ResolutionDetails[bool]{Value: !def}
}

func setNum(t *testing.T, num float64) *numProvider {
	t.Helper()
	Reset()
	t.Cleanup(Reset)
	p := &numProvider{testProvider: newTestProvider("num"), num: num}
	if err := SetProviderAndWait(p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestIntDetails(t *testing.T) {
	setNum(t, 3)
	d := GetClient().IntDetails("f", 7)
	if d.Value != 3 || d.ErrorCode != "" || d.Variant != "v" || d.Reason != ReasonStatic {
		t.Fatalf("integral: %+v", d)
	}
	if GetClient().Int("f", 7) != 3 {
		t.Fatal("Int")
	}
	for _, f := range []float64{3.5, 1e19, -1e19, math.NaN(), math.Inf(1)} {
		p := setNum(t, f)
		p.num = f
		var afterRan bool
		var hookErr error
		var finally EvaluationDetails[any]
		h := Hook{
			After:   func(HookContext, EvaluationDetails[any], HookHints) error { afterRan = true; return nil },
			Error:   func(_ HookContext, err error, _ HookHints) { hookErr = err },
			Finally: func(_ HookContext, d EvaluationDetails[any], _ HookHints) { finally = d },
		}
		d := GetClient().IntDetails("f", 7, WithHooks(h))
		if d.Value != 7 || d.ErrorCode != CodeTypeMismatch || d.ErrorMessage != "value is not an integer" || d.Reason != ReasonError || d.Variant != "" || d.FlagMetadata["m"] != "x" {
			t.Fatalf("%v: %+v", f, d)
		}
		if afterRan || codeOf(hookErr) != CodeTypeMismatch || hookErr.Error() != "value is not an integer" {
			t.Fatalf("%v: after=%v err=%v", f, afterRan, hookErr)
		}
		if finally.Value != int64(7) || finally.ErrorCode != CodeTypeMismatch {
			t.Fatalf("%v: finally %+v", f, finally)
		}
	}
}

func TestResolverPanicBecomesGeneral(t *testing.T) {
	for _, c := range []struct {
		panic any
		msg   string
	}{{errors.New("boom"), "boom"}, {NewError(CodeParseError, "bad"), "bad"}, {"boom", ""}, {42, ""}} {
		p := setNum(t, 0)
		p.panicWith = c.panic
		var hookErr error
		h := Hook{Error: func(_ HookContext, err error, _ HookHints) { hookErr = err }}
		d := GetClient().BooleanDetails("f", true, WithHooks(h))
		want := CodeGeneral
		if e, ok := c.panic.(*Error); ok {
			want = e.Code
		}
		if d.Value != true || d.ErrorCode != want || d.ErrorMessage != c.msg || d.Reason != ReasonError || len(d.FlagMetadata) != 0 {
			t.Fatalf("%v: %+v", c.panic, d)
		}
		if hookErr == nil {
			t.Fatalf("%v: error hook not run", c.panic)
		}
	}
}

func TestErrorFinallyLoggerReadAfterPanic(t *testing.T) {
	setNum(t, 1)
	c := GetClient()
	el, fl := &memLogger{}, &memLogger{}
	h := Hook{
		Error:   func(HookContext, error, HookHints) { c.SetLogger(el); panic("e") },
		Finally: func(HookContext, EvaluationDetails[any], HookHints) { c.SetLogger(fl); panic("f") },
	}
	c.IntDetails("f", 0, WithHooks(Hook{Before: func(HookContext, HookHints) error { return errors.New("x") }}, h))
	if len(el.lines) == 0 || el.lines[0] != "Unhandled error during 'error' hook: e" {
		t.Fatalf("error hook logged %v", el.lines)
	}
	if len(fl.lines) == 0 || fl.lines[0] != "Unhandled error during 'finally' hook: f" {
		t.Fatalf("finally hook logged %v", fl.lines)
	}
}

type countingProvider struct {
	*numProvider
	calls int
}

func (p *countingProvider) Metadata() Metadata { p.calls++; return Metadata{Name: "count"} }

func TestMetadataReadPerHookContext(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	p := &countingProvider{numProvider: &numProvider{testProvider: newTestProvider("count")}}
	if err := SetProviderAndWait(p); err != nil {
		t.Fatal(err)
	}
	for _, n := range []int{0, 1, 3} {
		hooks := make([]Hook, n)
		p.calls = 0
		GetClient("d").Number("f", 0, WithHooks(hooks...))
		if p.calls != 2*n {
			t.Fatalf("%d hooks: %d Metadata calls, want %d", n, p.calls, 2*n)
		}
	}
}

func TestHookContextIsSnapshot(t *testing.T) {
	p := setNum(t, 0)
	SetContext(Context{TargetingKey: "u", Attributes: map[string]any{"a": 1.0}})
	var seen []map[string]any
	mut := Hook{Before: func(hc HookContext, _ HookHints) error {
		seen = append(seen, hc.Context.Attributes)
		hc.Context.Attributes["x"] = true
		return nil
	}}
	GetClient().Boolean("f", false, WithHooks(mut, mut))
	if _, ok := p.gotCtx.Attributes["x"]; ok || p.gotCtx.Attributes["a"] != 1.0 {
		t.Fatalf("provider saw %v", p.gotCtx.Attributes)
	}
	if len(seen) != 2 || len(seen[1]) != 2 {
		t.Fatalf("hooks share context: %v", seen)
	}
	if _, ok := GetContext().Attributes["x"]; ok {
		t.Fatal("global context mutated")
	}
}

func TestBeforeHookRebindResolvesWithCurrentProvider(t *testing.T) {
	setNum(t, 1)
	next := &numProvider{testProvider: newTestProvider("next"), num: 2}
	var meta []string
	h := Hook{
		Before: func(hc HookContext, _ HookHints) error {
			meta = append(meta, hc.ProviderMetadata.Name)
			return SetProvider(next)
		},
		Finally: func(hc HookContext, _ EvaluationDetails[any], _ HookHints) {
			meta = append(meta, hc.ProviderMetadata.Name)
		},
	}
	if v := GetClient().Number("f", 0, WithHooks(h)); v != 2 {
		t.Fatalf("resolved with old provider: %v", v)
	}
	if len(meta) != 2 || meta[0] != "num" || meta[1] != "num" {
		t.Fatalf("hook metadata %v", meta)
	}
}

func TestLoggerReadWhenUsed(t *testing.T) {
	p := setNum(t, 1)
	c := GetClient()
	l := &memLogger{}
	h := Hook{
		Before:  func(HookContext, HookHints) error { c.SetLogger(l); return nil },
		Finally: func(HookContext, EvaluationDetails[any], HookHints) { panic("late") },
	}
	c.Number("f", 0, WithHooks(h))
	if sl, ok := p.gotLog.(safeLogger); !ok || sl.l != l {
		t.Fatalf("resolver logger %#v", p.gotLog)
	}
	if len(l.lines) == 0 || l.lines[0] != "Unhandled error during 'finally' hook: late" {
		t.Fatalf("logged %v", l.lines)
	}
}

type debugLogger struct{ memLogger }

func (l *debugLogger) Debug(a ...any) { l.Error(a...) }

func TestTrackLoggerReadWhenUsed(t *testing.T) {
	Reset()
	t.Cleanup(Reset)
	c := GetClient()
	l := &debugLogger{}
	tp := &loggerSwapTracker{testProvider: newTestProvider("swap"), swap: func() { c.SetLogger(l) }}
	if err := SetProviderAndWait(tp); err != nil {
		t.Fatal(err)
	}
	c.Track("e")
	if len(l.lines) != 1 || l.lines[0] != "Error recording tracking event." {
		t.Fatalf("logged %v", l.lines)
	}
}

type loggerSwapTracker struct {
	*testProvider
	swap func()
}

func (p *loggerSwapTracker) Track(string, Context, TrackingDetails) {
	p.swap()
	panic("x")
}

func TestHookHintsAreReadOnly(t *testing.T) {
	src := map[string]any{"h": 1}
	hints := NewHookHints(src)
	src["h"] = 2
	hints.Map()["h"] = 3
	if hints.Value("h") != 1 || hints.Value("missing") != nil {
		t.Fatalf("hints changed: %v", hints.Map())
	}
	if (HookHints{}).Map() != nil || NewHookHints(nil).Map() != nil {
		t.Fatal("unset hints must stay nil")
	}
}

func TestProviderFlagMetadataIsCloned(t *testing.T) {
	p := setNum(t, 1)
	p.md = FlagMetadata{"m": "x"}
	h := Hook{After: func(_ HookContext, d EvaluationDetails[any], _ HookHints) error {
		d.FlagMetadata["m"] = "changed"
		return nil
	}}
	GetClient().NumberDetails("f", 0, WithHooks(h))
	if p.md["m"] != "x" {
		t.Fatalf("provider metadata mutated: %v", p.md)
	}
}
