package parity

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"sync"
	"testing"

	"github.com/Toms223/WebGo/openfeature"
)

type errSpec struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type flagSpec struct {
	Value        val      `json:"value"`
	Variant      string   `json:"variant"`
	Reason       string   `json:"reason"`
	ErrorCode    string   `json:"errorCode"`
	ErrorMessage string   `json:"errorMessage"`
	FlagMetadata val      `json:"flagMetadata"`
	Throw        *errSpec `json:"throw"`
}

type providerDef struct {
	ID              string                         `json:"-"`
	Name            *string                        `json:"name"`
	Metadata        map[string]string              `json:"metadata"`
	RunsOn          *string                        `json:"runsOn"`
	DomainScoped    *bool                          `json:"domainScoped"`
	Events          bool                           `json:"events"`
	Hooks           []string                       `json:"hooks"`
	Initialize      string                         `json:"initialize"`
	OnContextChange string                         `json:"onContextChange"`
	OnClose         string                         `json:"onClose"`
	Track           val                            `json:"track"`
	Flags           map[string]map[string]flagSpec `json:"flags"`
	Memory          val                            `json:"memory"`
	raw             []byte
}

func (d *providerDef) UnmarshalJSON(b []byte) error {
	type plainDef providerDef
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if err := dec.Decode((*plainDef)(d)); err != nil {
		return err
	}
	d.raw = slices.Clone(b)
	return nil
}

type hookDef struct {
	ThrowIn   []string `json:"throwIn"`
	Data      bool     `json:"data"`
	Telemetry bool     `json:"telemetry"`
}

type recorder struct {
	mu       sync.Mutex
	closed   bool
	lines    [][]byte
	deferred map[string]chan error
	micro    []chan error
	hooks    map[string]openfeature.Hook
	hookDefs map[string]hookDef
	clients  map[domain]*openfeature.Client
	handlers map[string]func()
}

func newRecorder() *recorder {
	return &recorder{
		deferred: map[string]chan error{},
		hooks:    map[string]openfeature.Hook{},
		hookDefs: map[string]hookDef{},
		clients:  map[domain]*openfeature.Client{},
		handlers: map[string]func(){},
	}
}

func (r *recorder) record(m any) {
	line := canonical(m)
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.closed {
		r.lines = append(r.lines, line)
	}
}

func (r *recorder) wait(key string) chan error {
	ch := make(chan error, 1)
	r.mu.Lock()
	r.deferred[key] = ch
	r.mu.Unlock()
	return ch
}

func (r *recorder) microtask() chan error {
	ch := make(chan error, 1)
	r.mu.Lock()
	r.micro = append(r.micro, ch)
	r.mu.Unlock()
	return ch
}

func (r *recorder) nextMicrotask() (chan error, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.micro) == 0 {
		return nil, false
	}
	ch := r.micro[0]
	r.micro = r.micro[1:]
	return ch, true
}

func (r *recorder) take(key string) (chan error, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ch, ok := r.deferred[key]
	delete(r.deferred, key)
	return ch, ok
}

func (r *recorder) client(dm domain) *openfeature.Client {
	r.mu.Lock()
	c, ok := r.clients[dm]
	r.mu.Unlock()
	if ok {
		return c
	}
	c = getClient(dm)
	r.mu.Lock()
	defer r.mu.Unlock()
	if prev, ok := r.clients[dm]; ok {
		return prev
	}
	r.clients[dm] = c
	return c
}

func (r *recorder) hookList(ids []string) []openfeature.Hook {
	out := make([]openfeature.Hook, len(ids))
	for i, id := range ids {
		out[i] = r.hooks[id]
	}
	return out
}

type domain struct {
	name string
	set  bool
}

func dom(d string, literal bool) domain { return domain{d, d != "" || literal} }

func (d domain) args() []string {
	if !d.set {
		return nil
	}
	return []string{d.name}
}

func (d domain) record() any {
	if !d.set {
		return skip{}
	}
	return d.name
}

func getClient(d domain) *openfeature.Client { return openfeature.GetClient(d.args()...) }

type logRecorder struct{ r *recorder }

func (l logRecorder) Error(a ...any) { l.r.log("error", a) }
func (l logRecorder) Warn(a ...any)  { l.r.log("warn", a) }
func (l logRecorder) Info(a ...any)  { l.r.log("info", a) }
func (l logRecorder) Debug(a ...any) { l.r.log("debug", a) }

func (r *recorder) log(level string, a []any) {
	msg := "undefined"
	if len(a) > 0 {
		msg = jsString(a[0])
	}
	rec := map[string]any{"kind": "log", "level": level, "msg": msg}
	if len(a) > 1 {
		args := make([]any, len(a)-1)
		for i, x := range a[1:] {
			args[i] = jsString(x)
		}
		rec["args"] = args
	}
	r.record(rec)
}

var knownCodes = map[string]bool{}

func init() {
	for _, c := range []openfeature.ErrorCode{
		openfeature.CodeProviderNotReady, openfeature.CodeProviderFatal, openfeature.CodeFlagNotFound,
		openfeature.CodeParseError, openfeature.CodeTypeMismatch, openfeature.CodeTargetingKeyMissing,
		openfeature.CodeInvalidContext, openfeature.CodeGeneral,
	} {
		knownCodes[string(c)] = true
	}
}

func makeError(e errSpec) error {
	if e.Code == "" {
		return errors.New(e.Message)
	}
	code := openfeature.CodeGeneral
	if knownCodes[e.Code] {
		code = openfeature.ErrorCode(e.Code)
	}
	return openfeature.NewError(code, e.Message)
}

func jsString(v any) string {
	switch x := v.(type) {
	case nil:
		return "null"
	case error:
		return x.Error()
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	}
	switch rv := reflect.ValueOf(v); rv.Kind() {
	case reflect.String:
		return rv.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return jsNumber(float64(rv.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return jsNumber(float64(rv.Uint()))
	case reflect.Float32, reflect.Float64:
		if f := rv.Float(); math.IsNaN(f) {
			return "NaN"
		} else if math.IsInf(f, 0) {
			return map[bool]string{true: "Infinity", false: "-Infinity"}[f > 0]
		}
		return jsNumber(rv.Float())
	}
	return fmt.Sprint(v)
}

func errorRecord(err error) map[string]any {
	m := map[string]any{}
	if err == nil {
		return m
	}
	m["message"] = err.Error()
	if oe, ok := err.(*openfeature.Error); ok {
		put(m, "code", string(oe.Code))
	}
	return m
}

func optional(s string) any {
	if s == "" {
		return skip{}
	}
	return s
}

func trimArgs(args []any) []any {
	args = append([]any{}, args...)
	for len(args) > 0 {
		if _, ok := args[len(args)-1].(skip); !ok {
			break
		}
		args = args[:len(args)-1]
	}
	return args
}

type fake struct {
	def    providerDef
	rec    *recorder
	counts map[string]int
	events *openfeature.Emitter
}

func (f *fake) call(method string, args ...any) string {
	f.rec.mu.Lock()
	f.counts[method]++
	n := f.counts[method]
	f.rec.mu.Unlock()
	f.rec.record(map[string]any{"kind": "provider", "p": f.def.ID, "call": method, "args": trimArgs(args)})
	return fmt.Sprintf("%s.%s#%d", f.def.ID, method, n)
}

func resolve[T any](f *fake, typ, method, flag string, def T, ctx openfeature.Context) openfeature.ResolutionDetails[T] {
	f.call(method, flag, def, ctx)
	spec, ok := f.def.Flags[flag][typ]
	if !ok {
		return openfeature.ResolutionDetails[T]{Value: def, Reason: openfeature.ReasonDefault}
	}
	if spec.Throw != nil {
		panic(makeError(*spec.Throw))
	}
	r := openfeature.ResolutionDetails[T]{
		Value:        def,
		Variant:      spec.Variant,
		Reason:       openfeature.Reason(spec.Reason),
		ErrorCode:    openfeature.ErrorCode(spec.ErrorCode),
		ErrorMessage: spec.ErrorMessage,
	}
	if spec.FlagMetadata.set {
		md, _ := plain(spec.FlagMetadata.v).(map[string]any)
		r.FlagMetadata = md
	}
	if spec.Value.set {
		if v, ok := spec.Value.v.(T); ok {
			r.Value = v
		} else {
			var zero T
			r.Value = zero
		}
	}
	return r
}

func (f *fake) settle(mode, key string) openfeature.Pending {
	if mode == "deferred" {
		return f.rec.wait(key)
	}
	return f.rec.microtask()
}

type Base struct{ F *fake }

func (b Base) Metadata() openfeature.Metadata {
	name := b.F.def.ID
	if b.F.def.Name != nil {
		name = *b.F.def.Name
	}
	return openfeature.Metadata{Name: name, Attributes: b.F.def.Metadata}
}

func (b Base) ResolveBoolean(flag string, def bool, ctx openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[bool] {
	return resolve(b.F, "boolean", "resolveBooleanEvaluation", flag, def, ctx)
}

func (b Base) ResolveString(flag string, def string, ctx openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[string] {
	return resolve(b.F, "string", "resolveStringEvaluation", flag, def, ctx)
}

func (b Base) ResolveNumber(flag string, def float64, ctx openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[float64] {
	return resolve(b.F, "number", "resolveNumberEvaluation", flag, def, ctx)
}

func (b Base) ResolveObject(flag string, def any, ctx openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[any] {
	return resolve(b.F, "object", "resolveObjectEvaluation", flag, def, ctx)
}

type Running struct{ F *fake }

func (x Running) RunsOn() string { return *x.F.def.RunsOn }

func (b Base) DomainScoped() bool { return b.F.def.DomainScoped != nil && *b.F.def.DomainScoped }

func (b Base) Hooks() []openfeature.Hook { return b.F.rec.hookList(b.F.def.Hooks) }

type Init struct{ F *fake }

func (x Init) Initialize(ctx openfeature.Context, domain string) (openfeature.Pending, error) {
	return x.F.settle(x.F.def.Initialize, x.F.call("initialize", ctx, optional(domain))), nil
}

type Ctx struct{ F *fake }

func (x Ctx) OnContextChange(old, new openfeature.Context) (openfeature.Pending, error) {
	key := x.F.call("onContextChange", old, new)
	switch x.F.def.OnContextChange {
	case "throw":
		return nil, makeError(errSpec{Message: x.F.def.ID + " onContextChange throw"})
	case "deferred":
		return x.F.rec.wait(key), nil
	}
	return nil, nil
}

type Closing struct{ F *fake }

func (x Closing) Close() (openfeature.Pending, error) {
	return x.F.settle(x.F.def.OnClose, x.F.call("onClose")), nil
}

type Tracking struct{ F *fake }

func (x Tracking) Track(name string, ctx openfeature.Context, d openfeature.TrackingDetails) {
	x.F.call("track", name, ctx, d)
	if x.F.def.Track.v == "throw" {
		panic(makeError(errSpec{Message: x.F.def.ID + " track throw"}))
	}
}

type Emitting struct{ F *fake }

func (x Emitting) Events() *openfeature.Emitter { return x.F.events }

var shapes = [64]reflect.Type{
	reflect.TypeFor[struct{ Base }](),
	reflect.TypeFor[struct {
		Base
		Init
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
	}](),
	reflect.TypeFor[struct {
		Base
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Tracking
	}](),
	reflect.TypeFor[struct {
		Base
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Tracking
		Emitting
	}](),
	reflect.TypeFor[struct {
		Base
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Tracking
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Closing
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Closing
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Ctx
		Closing
		Tracking
		Emitting
		Running
	}](),
	reflect.TypeFor[struct {
		Base
		Init
		Ctx
		Closing
		Tracking
		Emitting
		Running
	}](),
}

func has(mode string) bool { return mode != "" && mode != "none" }

func goFake(def providerDef, rec *recorder) openfeature.Provider {
	f := &fake{def: def, rec: rec, counts: map[string]int{}}
	bits := 0
	for i, on := range []bool{
		has(def.Initialize), has(def.OnContextChange), has(def.OnClose),
		def.Track.set && def.Track.v != false && def.Track.v != nil, def.Events, def.RunsOn != nil,
	} {
		if on {
			bits |= 1 << i
		}
	}
	if def.Events {
		f.events = openfeature.NewEmitter()
	}
	v := reflect.New(shapes[bits])
	for i := range v.Elem().NumField() {
		v.Elem().Field(i).Field(0).Set(reflect.ValueOf(f))
	}
	p := v.Interface().(openfeature.Provider)
	if got := shapeBits(p); got != bits {
		panic(fmt.Sprintf("fake %s: optional interfaces %06b, want %06b", def.ID, got, bits))
	}
	return p
}

func shapeBits(p openfeature.Provider) int {
	bits := 0
	for i, ok := range []bool{
		implements[openfeature.Initializer](p), implements[openfeature.ContextChanger](p),
		implements[openfeature.Closer](p), implements[openfeature.Tracker](p), implements[openfeature.Eventer](p),
		implements[openfeature.RunsOn](p),
	} {
		if ok {
			bits |= 1 << i
		}
	}
	return bits
}

func implements[I any](p openfeature.Provider) bool {
	_, ok := p.(I)
	return ok
}

func makeHook(id string, def hookDef, rec *recorder) openfeature.Hook {
	throws := func(stage string) error {
		for _, s := range def.ThrowIn {
			if s == stage {
				return makeError(errSpec{Message: fmt.Sprintf("hook %s %s", id, stage)})
			}
		}
		return nil
	}
	stage := func(name string, hc openfeature.HookContext, hints openfeature.HookHints, extra map[string]any) error {
		r := map[string]any{
			"kind": "hook", "h": id, "stage": name,
			"flagKey": hc.FlagKey, "defaultValue": hc.DefaultValue, "flagValueType": string(hc.FlagType),
			"context": hc.Context, "domain": optional(hc.ClientMetadata.Domain), "providerName": hc.ProviderMetadata.Name,
			"hints": hints.Map(),
		}
		for k, v := range extra {
			r[k] = v
		}
		if def.Data {
			n, _ := hc.HookData.Get("n").(float64)
			r["data"] = n
			hc.HookData.Set("n", n+1)
		}
		rec.record(r)
		if def.Telemetry && name == "finally" {
			rec.record(map[string]any{"kind": "telemetry", "event": openfeature.CreateEvaluationEvent(hc, extra["details"].(openfeature.EvaluationDetails[any]))})
		}
		return throws(name)
	}
	return openfeature.Hook{
		Before: func(hc openfeature.HookContext, hints openfeature.HookHints) error {
			return stage("before", hc, hints, nil)
		},
		After: func(hc openfeature.HookContext, d openfeature.EvaluationDetails[any], hints openfeature.HookHints) error {
			return stage("after", hc, hints, map[string]any{"details": d})
		},
		Error: func(hc openfeature.HookContext, err error, hints openfeature.HookHints) {
			if e := stage("error", hc, hints, map[string]any{"error": errorRecord(err)}); e != nil {
				panic(e)
			}
		},
		Finally: func(hc openfeature.HookContext, d openfeature.EvaluationDetails[any], hints openfeature.HookHints) {
			if e := stage("finally", hc, hints, map[string]any{"details": d}); e != nil {
				panic(e)
			}
		},
	}
}

func TestGoFakeShapes(t *testing.T) {
	modes := []string{"", "sync"}
	for bits := range 64 {
		def := providerDef{
			ID:              "p",
			Initialize:      modes[bits&1],
			OnContextChange: modes[bits>>1&1],
			OnClose:         modes[bits>>2&1],
			Events:          bits&16 != 0,
		}
		if bits&8 != 0 {
			def.Track = val{v: true, set: true}
		}
		if bits&32 != 0 {
			def.RunsOn = new("client")
		}
		p := goFake(def, newRecorder())
		if got := shapeBits(p); got != bits {
			t.Errorf("shape %06b: provider implements %06b", bits, got)
		}
		if r, ok := p.(openfeature.RunsOn); ok && r.RunsOn() != "client" {
			t.Errorf("shape %06b: RunsOn() = %q", bits, r.RunsOn())
		}
	}
}
