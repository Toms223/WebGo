package parity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/memprovider"
)

type scenario struct {
	ID           string                 `json:"id"`
	Requirements []string               `json:"requirements"`
	Deviation    string                 `json:"deviation"`
	Partial      string                 `json:"partial"`
	Providers    map[string]providerDef `json:"providers"`
	Hooks        map[string]hookDef     `json:"hooks"`
	Steps        []step                 `json:"steps"`
}

type step struct {
	Op         string   `json:"op"`
	Provider   string   `json:"provider"`
	Domain     string   `json:"domain"`
	Context    val      `json:"context"`
	Target     string   `json:"target"`
	Error      errSpec  `json:"error"`
	Event      string   `json:"event"`
	Handler    string   `json:"handler"`
	ReadStatus *string  `json:"readStatus"`
	Throws     bool     `json:"throws"`
	Fresh      bool     `json:"fresh"`
	Level      string   `json:"level"`
	Hooks      []string `json:"hooks"`
	Type       string   `json:"type"`
	Flag       string   `json:"flag"`
	Default    val      `json:"default"`
	Hints      val      `json:"hints"`
	Name       string   `json:"name"`
	Details    val      `json:"details"`
	Version    *string  `json:"version"`
	Flags      val      `json:"flags"`
	Literal    bool     `json:"literalDomain"`
}

type providerFactory func(def providerDef, rec *recorder) openfeature.Provider

type harness struct {
	provider providerFactory
	emit     func(p openfeature.Provider, event string, details val)
	bubble   func(t *testing.T, f func(t *testing.T))
	settle   func(rec *recorder)
}

var goHarness = harness{
	provider: goFake,
	emit: func(p openfeature.Provider, event string, details val) {
		p.(openfeature.Eventer).Events().Emit(openfeature.Event(event), eventDetails(details.v))
	},
	bubble: synctest.Test,
	settle: func(rec *recorder) {
		synctest.Wait()
		drainMicrotasks(rec)
	},
}

func loadScenario(t *testing.T, id string) scenario {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("scenarios", id+".json"))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var sc scenario
	if err := dec.Decode(&sc); err != nil {
		t.Fatalf("%s: %v", id, err)
	}
	if sc.ID != id {
		t.Fatalf("%s: id %q does not match file name", id, sc.ID)
	}
	for pid, def := range sc.Providers {
		def.ID = pid
		sc.Providers[pid] = def
		for flag, types := range def.Flags {
			for typ, spec := range types {
				if spec.Value.set && !representable(typ, spec.Value.v) {
					t.Fatalf("%s: provider %s flag %s: %s value %v has no Go representation", id, pid, flag, typ, spec.Value.v)
				}
			}
		}
	}
	return sc
}

func representable(typ string, v any) bool {
	switch typ {
	case "boolean":
		_, ok := v.(bool)
		return ok
	case "string":
		_, ok := v.(string)
		return ok
	case "number":
		_, ok := v.(float64)
		return ok
	}
	return typ == "object"
}

func ofTarget(t string) string { return strings.TrimPrefix(t, "client:") }

func toContext(v any) openfeature.Context {
	m, _ := ctxValue(v).(map[string]any)
	c := openfeature.Context{}
	if tk, ok := m["targetingKey"].(string); ok {
		c.TargetingKey = tk
		delete(m, "targetingKey")
	}
	if len(m) > 0 {
		c.Attributes = m
	}
	return c
}

func ctxValue(v any) any {
	switch x := v.(type) {
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = ctxValue(e)
		}
		return out
	case object:
		if len(x) == 1 && x[0].Key == "$date" {
			if s, ok := x[0].Value.(string); ok {
				d, err := time.Parse(time.RFC3339Nano, s)
				if err != nil {
					panic(err)
				}
				return d.UTC()
			}
		}
		m := make(map[string]any, len(x))
		for _, e := range x {
			m[e.Key] = ctxValue(e.Value)
		}
		return m
	}
	return v
}

func trackingDetails(v any) openfeature.TrackingDetails {
	var d openfeature.TrackingDetails
	obj, _ := v.(object)
	for _, m := range obj {
		if f, ok := m.Value.(float64); ok && m.Key == "value" {
			d.Value = &f
			continue
		}
		if d.Attributes == nil {
			d.Attributes = map[string]any{}
		}
		d.Attributes[m.Key] = plain(m.Value)
	}
	return d
}

func eventDetails(v any) openfeature.EventDetails {
	var d openfeature.EventDetails
	m, _ := plain(v).(map[string]any)
	d.Message, _ = m["message"].(string)
	if code, ok := m["errorCode"].(string); ok {
		d.ErrorCode = openfeature.ErrorCode(code)
	}
	if fc, ok := m["flagsChanged"].([]any); ok {
		d.FlagsChanged = []string{}
		for _, f := range fc {
			s, _ := f.(string)
			d.FlagsChanged = append(d.FlagsChanged, s)
		}
	}
	d.Metadata, _ = m["metadata"].(map[string]any)
	return d
}

type run struct {
	t         *testing.T
	h         harness
	rec       *recorder
	providers map[string]openfeature.Provider
}

func (r *run) ret(op string, err error) {
	m := map[string]any{"kind": "return", "op": op}
	if err != nil {
		m["error"] = errorRecord(err)
	}
	r.rec.record(m)
}

func (r *run) runnerError(format string, a ...any) {
	r.rec.record(map[string]any{"kind": "runner", "error": fmt.Sprintf(format, a...)})
}

func (r *run) provider(id string) openfeature.Provider {
	p, ok := r.providers[id]
	if !ok {
		r.t.Fatalf("unknown provider %s", id)
	}
	return p
}

func (r *run) setOptions(s step) []openfeature.SetOption {
	var opts []openfeature.SetOption
	if d := dom(s.Domain, s.Literal); d.set {
		opts = append(opts, openfeature.ForDomain(d.name))
	}
	if s.Context.set {
		opts = append(opts, openfeature.WithContext(toContext(s.Context.v)))
	}
	return opts
}

func (r *run) hooks(ids []string) []openfeature.Hook {
	for _, id := range ids {
		if _, ok := r.rec.hooks[id]; !ok {
			r.t.Fatalf("unknown hook %s", id)
		}
	}
	return r.rec.hookList(ids)
}

func (r *run) handler(s step) openfeature.Handler {
	return func(d openfeature.EventDetails) {
		r.rec.record(map[string]any{"kind": "event", "target": s.Target, "handler": s.Handler, "event": s.Event, "details": d})
		if s.ReadStatus != nil {
			r.rec.record(map[string]any{"kind": "status", "domain": dom(*s.ReadStatus, false).record(), "status": string(r.rec.client(dom(*s.ReadStatus, false)).ProviderStatus())})
		}
		if s.Throws {
			panic(makeError(errSpec{Message: "handler " + s.Handler}))
		}
	}
}

func evaluate(c *openfeature.Client, s step, opts []openfeature.EvalOption) (map[string]any, error) {
	def := s.Default.v
	switch s.Type {
	case "boolean":
		if v, ok := def.(bool); ok {
			return details(c.BooleanDetails(s.Flag, v, opts...)), nil
		}
	case "string":
		if v, ok := def.(string); ok {
			return details(c.StringDetails(s.Flag, v, opts...)), nil
		}
	case "number":
		if v, ok := def.(float64); ok {
			return details(c.NumberDetails(s.Flag, v, opts...)), nil
		}
	case "object":
		return details(c.ObjectDetails(s.Flag, def, opts...)), nil
	}
	return nil, fmt.Errorf("evaluate %s: default %s has no Go representation", s.Type, canonical(def))
}

func (r *run) op(s step) func() {
	rec := r.rec
	switch s.Op {
	case "setProvider":
		p, opts := r.provider(s.Provider), r.setOptions(s)
		return func() { r.ret("setProvider", openfeature.SetProvider(p, opts...)) }
	case "setProviderAndWait":
		p, opts := r.provider(s.Provider), r.setOptions(s)
		return func() { r.ret("setProviderAndWait", openfeature.SetProviderAndWait(p, opts...)) }
	case "resolve", "reject":
		ch, ok := rec.take(s.Target)
		if !ok {
			return func() { r.runnerError("no pending call %s", s.Target) }
		}
		return func() {
			if s.Op == "reject" {
				ch <- makeError(s.Error)
			}
			close(ch)
		}
	case "setContext":
		ctx := toContext(s.Context.v)
		return func() {
			openfeature.SetContext(ctx, dom(s.Domain, false).args()...)
			r.ret("setContext", nil)
		}
	case "clearContext":
		return func() {
			openfeature.ClearContext(dom(s.Domain, false).args()...)
			r.ret("clearContext", nil)
		}
	case "clearContexts":
		return func() {
			openfeature.ClearContexts()
			r.ret("clearContexts", nil)
		}
	case "getContext":
		return func() {
			rec.record(map[string]any{"kind": "context", "domain": dom(s.Domain, false).record(), "context": openfeature.GetContext(dom(s.Domain, false).args()...)})
		}
	case "addHooks", "clearHooks":
		hooks := r.hooks(s.Hooks)
		return func() {
			switch {
			case s.Level == "api" && s.Op == "addHooks":
				openfeature.AddHooks(hooks...)
			case s.Level == "api":
				openfeature.ClearHooks()
			case s.Op == "addHooks":
				rec.client(dom(ofTarget(s.Level), false)).AddHooks(hooks...)
			default:
				rec.client(dom(ofTarget(s.Level), false)).ClearHooks()
			}
		}
	case "addHandler":
		h := r.handler(s)
		return func() {
			var remove func()
			switch {
			case s.Target == "api":
				remove = openfeature.AddHandler(openfeature.Event(s.Event), h)
			case s.Fresh:
				remove = getClient(dom(ofTarget(s.Target), s.Literal)).AddHandler(openfeature.Event(s.Event), h)
			default:
				remove = rec.client(dom(ofTarget(s.Target), s.Literal)).AddHandler(openfeature.Event(s.Event), h)
			}
			rec.mu.Lock()
			rec.handlers[s.Handler] = remove
			rec.mu.Unlock()
		}
	case "removeHandler":
		rec.mu.Lock()
		remove, ok := rec.handlers[s.Handler]
		rec.mu.Unlock()
		if !ok {
			r.t.Fatalf("unknown handler %s", s.Handler)
		}
		if remove == nil {
			return func() { r.runnerError("AddHandler returned a nil remover for %s", s.Handler) }
		}
		return remove
	case "emit":
		p := r.provider(s.Provider)
		if _, ok := p.(openfeature.Eventer); !ok {
			r.t.Fatalf("provider %s has no events", s.Provider)
		}
		return func() { r.h.emit(p, s.Event, s.Details) }
	case "evaluate":
		var opts []openfeature.EvalOption
		if s.Hooks != nil {
			opts = append(opts, openfeature.WithHooks(r.hooks(s.Hooks)...))
		}
		if s.Hints.v != nil {
			hints, _ := plain(s.Hints.v).(map[string]any)
			opts = append(opts, openfeature.WithHookHints(openfeature.NewHookHints(hints)))
		}
		return func() {
			d, err := evaluate(rec.client(dom(s.Domain, false)), s, opts)
			if err != nil {
				r.runnerError("%v", err)
				return
			}
			rec.record(map[string]any{"kind": "result", "details": d})
		}
	case "track":
		return func() {
			c := rec.client(dom(s.Domain, false))
			if s.Details.set {
				c.Track(s.Name, trackingDetails(s.Details.v))
			} else {
				c.Track(s.Name)
			}
		}
	case "status":
		return func() {
			rec.record(map[string]any{"kind": "status", "domain": dom(s.Domain, s.Literal).record(), "status": string(rec.client(dom(s.Domain, s.Literal)).ProviderStatus())})
		}
	case "clientMetadata":
		return func() {
			c := rec.client(dom(s.Domain, s.Literal))
			if s.Version != nil {
				c = openfeature.GetClientVersion(*s.Version, dom(s.Domain, s.Literal).args()...)
			}
			rec.record(map[string]any{"kind": "metadata", "which": "client", "value": c.Metadata()})
		}
	case "providerMetadata":
		return func() {
			rec.record(map[string]any{"kind": "metadata", "which": "provider", "value": openfeature.ProviderMetadata(dom(s.Domain, false).args()...)})
		}
	case "close":
		return func() {
			openfeature.Close()
			r.ret("close", nil)
		}
	case "clearProviders":
		return func() {
			openfeature.ClearProviders()
			r.ret("clearProviders", nil)
		}
	case "putConfiguration":
		return func() {
			if err := r.provider(s.Provider).(*memprovider.Provider).PutConfiguration(memFlags(s.Flags.v)...); err != nil {
				r.runnerError("putConfiguration: %v", err)
			}
		}
	}
	r.t.Fatalf("unknown op %s", s.Op)
	return nil
}

func runScenario(t *testing.T, id string, h harness) []byte {
	sc := loadScenario(t, id)
	var trace []byte
	h.bubble(t, func(t *testing.T) {
		openfeature.Reset()
		defer openfeature.Reset()
		rec := newRecorder()
		for hid, def := range sc.Hooks {
			rec.hooks[hid] = makeHook(hid, def, rec)
			rec.hookDefs[hid] = def
		}
		r := &run{t: t, h: h, rec: rec, providers: map[string]openfeature.Provider{}}
		for pid, def := range sc.Providers {
			if def.Memory.set {
				r.providers[pid] = memprovider.New(memFlags(def.Memory.v)...)
				continue
			}
			r.providers[pid] = h.provider(def, rec)
		}
		openfeature.SetLogger(logRecorder{rec})
		for i, s := range sc.Steps {
			rec.record(map[string]any{"kind": "step", "i": i, "op": s.Op})
			fn := r.op(s)
			go func() {
				defer func() {
					if p := recover(); p != nil {
						r.runnerError("step %d panicked: %v", i, p)
					}
				}()
				fn()
			}()
			h.settle(rec)
		}
		rec.mu.Lock()
		trace = append([]byte("[\n"), bytes.Join(rec.lines, []byte(",\n"))...)
		trace = append(trace, "\n]\n"...)
		rec.closed = true
		rec.mu.Unlock()
		release(rec, h.settle)
	})
	return trace
}

func drainMicrotasks(rec *recorder) {
	for {
		ch, ok := rec.nextMicrotask()
		if !ok {
			return
		}
		close(ch)
		synctest.Wait()
	}
}

func release(rec *recorder, settle func(*recorder)) {
	for {
		settle(rec)
		rec.mu.Lock()
		pending := rec.deferred
		rec.deferred = map[string]chan error{}
		rec.mu.Unlock()
		if len(pending) == 0 {
			return
		}
		for _, ch := range pending {
			close(ch)
		}
	}
}

func TestParity(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("scenarios", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no scenarios: %v", err)
	}
	slices.Sort(files)
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".json")
		if d := loadScenario(t, id).Deviation; d != "" {
			t.Logf("%s: deviation %s, compared by TestDeviation_%s instead", id, d, id)
			continue
		}
		t.Run(id, func(t *testing.T) { checkGolden(t, id, runScenario(t, id, goHarness)) })
	}
}

var sdkErrorCode = regexp.MustCompile(`"errorCode":"[A-Z_]+",`)

func partial(t *testing.T, id string, want, got []string) ([]string, []string) {
	switch d := loadScenario(t, id).Partial; d {
	case "":
	case "D10":
		for i := range min(len(want), len(got)) {
			if strings.Contains(got[i], `"event":"PROVIDER_ERROR"`) && !strings.Contains(want[i], `"errorCode"`) {
				got[i] = sdkErrorCode.ReplaceAllString(got[i], "")
			}
		}
	case "D11":
		const closed = `{"kind":"return","op":"close"},`
		w, g := slices.Index(want, closed), slices.Index(got, closed)
		if w < 0 || g < 0 {
			t.Fatalf("%s: no close return in golden (%d) or trace (%d)", id, w, g)
		}
		return want[:w+1], got[:g+1]
	default:
		t.Fatalf("%s: unknown partial deviation %q", id, d)
	}
	return want, got
}

func checkGolden(t *testing.T, id string, got []byte) {
	if t.Failed() {
		return
	}
	golden, err := os.ReadFile(filepath.Join("golden", id+".trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	want, have := partial(t, id, lines(golden), lines(got))
	if !slices.Equal(want, have) {
		t.Errorf("trace mismatch (-golden +go):\n%s", unifiedDiff(want, have))
	}
}

func lines(b []byte) []string {
	return strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
}

func unifiedDiff(a, b []string) string {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	type line struct {
		op   byte
		text string
		i, j int
	}
	var ops []line
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && a[i] == b[j]:
			ops = append(ops, line{' ', a[i], i, j})
			i, j = i+1, j+1
		case i < n && (j == m || lcs[i+1][j] >= lcs[i][j+1]):
			ops = append(ops, line{'-', a[i], i, j})
			i++
		default:
			ops = append(ops, line{'+', b[j], i, j})
			j++
		}
	}
	const ctx = 3
	var out strings.Builder
	for k := 0; k < len(ops); {
		if ops[k].op == ' ' {
			k++
			continue
		}
		start := max(k-ctx, 0)
		end := k
		for end < len(ops) {
			if ops[end].op != ' ' {
				end++
				continue
			}
			run := end
			for run < len(ops) && ops[run].op == ' ' {
				run++
			}
			if run == len(ops) || run-end > 2*ctx {
				end = min(end+ctx, len(ops))
				break
			}
			end = run
		}
		var na, nb int
		for _, l := range ops[start:end] {
			if l.op != '+' {
				na++
			}
			if l.op != '-' {
				nb++
			}
		}
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n", ops[start].i+1, na, ops[start].j+1, nb)
		for _, l := range ops[start:end] {
			fmt.Fprintf(&out, "%c%s\n", l.op, l.text)
		}
		k = end
	}
	return out.String()
}

func TestSyncSettleRunsAfterStep(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		rec := newRecorder()
		p := goFake(providerDef{ID: "p", Initialize: "sync", OnClose: "sync"}, rec)
		go func() {
			pend, _ := p.(openfeature.Initializer).Initialize(openfeature.Context{}, "")
			go func() {
				<-pend
				rec.record(map[string]any{"kind": "event", "event": "ready"})
				closing, _ := p.(openfeature.Closer).Close()
				<-closing
				rec.record(map[string]any{"kind": "event", "event": "closed"})
			}()
			rec.record(map[string]any{"kind": "return", "op": "setProvider"})
		}()
		synctest.Wait()
		drainMicrotasks(rec)
		want := `{"args":[{}],"call":"initialize","kind":"provider","p":"p"}` + "\n" +
			`{"kind":"return","op":"setProvider"}` + "\n" +
			`{"event":"ready","kind":"event"}` + "\n" +
			`{"args":[],"call":"onClose","kind":"provider","p":"p"}` + "\n" +
			`{"event":"closed","kind":"event"}`
		if got := string(bytes.Join(rec.lines, []byte("\n"))); got != want {
			t.Errorf("got\n%s\nwant\n%s", got, want)
		}
	})
}
