package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/flipt"
)

var contexts = []openfeature.Context{
	{TargetingKey: "u1", Attributes: map[string]any{"fizz": "buzz"}},
	{TargetingKey: "u1"},
	{TargetingKey: ""},
	{TargetingKey: "u1", Attributes: map[string]any{"fizz": 1.0}},
}

var flags = []string{"flag1", "flag_boolean", "missing"}

type page struct{ app.Compo }

func (p *page) Render() app.UI { return app.Div().Text("flipt parity: go") }

func (p *page) OnMount(ctx app.Context) { ctx.Async(func() { run(ctx) }) }

func run(ctx app.Context) {
	win := app.Window()
	defer func() {
		if r := recover(); r != nil {
			win.Set("parityError", fmt.Sprint(r))
		}
		win.Set("parityDone", true)
	}()
	c := openfeature.GetClient()
	var out []any
	for i, cx := range contexts {
		p, ok := flipt.New(ctx, flipt.Options{URL: app.Getenv("FLIPT_URL")})
		if !ok {
			panic("flipt module not loaded")
		}
		_ = openfeature.SetProviderAndWait(p, openfeature.WithContext(cx))
		out = append(out, record(c, i, "setProviderAndWait"))
		openfeature.SetContext(contexts[(i+1)%len(contexts)])
		out = append(out, record(c, i, "setContext"))
	}
	win.Set("parityResult", out)
}

func record(c *openfeature.Client, run int, step string) map[string]any {
	var ds []any
	for _, f := range flags {
		ds = append(ds,
			details(c.BooleanDetails(f, false)),
			details(c.StringDetails(f, "def")),
			details(c.NumberDetails(f, 1.5)),
			details(c.ObjectDetails(f, map[string]any{"d": 1.0})),
		)
	}
	return map[string]any{"run": run, "step": step, "status": string(c.ProviderStatus()), "details": ds}
}

func details[T any](d openfeature.EvaluationDetails[T]) map[string]any {
	m := map[string]any{"flagKey": d.FlagKey, "value": plain(d.Value), "flagMetadata": plain(map[string]any(d.FlagMetadata))}
	for k, v := range map[string]string{"variant": d.Variant, "reason": string(d.Reason), "errorCode": string(d.ErrorCode)} {
		if v != "" {
			m[k] = v
		}
	}
	if d.ErrorCode != "" || d.ErrorMessage != "" {
		m["errorMessage"] = d.ErrorMessage
	}
	return m
}

func plain(v any) any {
	switch x := v.(type) {
	case openfeature.Object:
		if x == nil {
			return nil
		}
		m := make(map[string]any, len(x))
		for _, f := range x {
			m[f.Key] = plain(f.Value)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = plain(e)
		}
		return m
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = plain(e)
		}
		return out
	}
	return v
}

func main() {
	app.Route("/", func() app.Composer { return &page{} })
	app.RunWhenOnBrowser()
	h := &app.Handler{
		Title:   "flipt parity: go",
		Imports: flipt.Modules(),
		Env:     map[string]string{"FLIPT_URL": os.Getenv("FLIPT_URL")},
	}
	log.Fatal(http.ListenAndServe(os.Getenv("ADDR"), h))
}
