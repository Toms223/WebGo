package parity

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/jsprovider"
)

var jsExcluded = map[string]string{
	"ctx-on-change-sync-throw":       "deviation D2",
	"events-rebind-stale-forwarding": "deviation D3",
	"mem-evaluation-feature":         "memory provider",
	"mem-prototype-variants":         "memory provider",
	"mem-put-configuration":          "memory provider",
	"mem-resolution-branches":        "memory provider",
}

func TestParityThroughJSProvider(t *testing.T) {
	fake := app.Window().Get("__parityFake")
	if fake.IsUndefined() {
		t.Fatal("fake.mjs not loaded: run with -exec js/wasm_exec_with_fake.sh")
	}
	files, err := filepath.Glob(filepath.Join("scenarios", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no scenarios: %v", err)
	}
	slices.Sort(files)
	for _, f := range files {
		id := strings.TrimSuffix(filepath.Base(f), ".json")
		sc := loadScenario(t, id)
		_, excluded := jsExcluded[id]
		if want := sc.Deviation != "" || strings.HasPrefix(id, "mem-"); excluded != want {
			t.Errorf("%s: excluded=%v, but deviation=%q", id, excluded, sc.Deviation)
		}
		if excluded {
			continue
		}
		t.Run(id, func(t *testing.T) { checkGolden(t, id, runScenario(t, id, jsHarness(fake))) })
	}
}

func jsHarness(fake app.Value) harness {
	objects := map[openfeature.Provider]app.Value{}
	return harness{
		provider: func(def providerDef, rec *recorder) openfeature.Provider {
			env := app.Window().Get("Object").New()
			env.Set("record", app.FuncOf(func(_ app.Value, args []app.Value) any {
				v, err := decodeJSON([]byte(app.Window().Get("JSON").Call("stringify", args[0]).String()))
				if err != nil {
					panic(err)
				}
				rec.record(v)
				return nil
			}))
			env.Set("defer", app.FuncOf(func(_ app.Value, args []app.Value) any {
				ch := rec.wait(args[0].String())
				var settle []app.Value
				exec := app.FuncOf(func(_ app.Value, fns []app.Value) any {
					settle = fns
					return nil
				})
				promise := app.Window().Get("Promise").New(exec)
				exec.Release()
				go func() {
					if err := <-ch; err != nil {
						settle[1].Invoke(fake.Call("makeError", app.ValueOf(errorRecord(err))))
						return
					}
					settle[0].Invoke()
				}()
				return promise
			}))
			env.Set("hook", app.FuncOf(func(_ app.Value, args []app.Value) any {
				id := args[0].String()
				return fake.Call("makeHook", id, parseJSON(rec.hookDefs[id]), env)
			}))
			obj := fake.Call("makeProvider", def.ID, app.Window().Get("JSON").Call("parse", string(def.raw)), env)
			p := jsprovider.New(obj)
			objects[p] = obj
			return p
		},
		emit: func(p openfeature.Provider, event string, details val) {
			d := app.Undefined()
			if details.set {
				d = parseJSON(details.v)
			}
			objects[p].Get("events").Call("emit", event, d)
		},
		bubble: func(t *testing.T, f func(t *testing.T)) { f(t) },
		settle: func(*recorder) {
			tick()
			tick()
		},
	}
}

func parseJSON(v any) app.Value {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return app.Window().Get("JSON").Call("parse", string(b))
}

func tick() {
	done := make(chan struct{})
	f := app.FuncOf(func(app.Value, []app.Value) any {
		close(done)
		return nil
	})
	app.Window().Call("setTimeout", f, 0)
	<-done
	f.Release()
}
