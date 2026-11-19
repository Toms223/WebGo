//go:build wasm

package flipt

import (
	"testing"

	"github.com/Toms223/WebGo/app"
)

func TestNewConstructsFliptWebProvider(t *testing.T) {
	win := app.Window()
	win.Call("eval", `globalThis.goappModules = {"@openfeature/flipt-web-provider": {FliptWebProvider: class {
		constructor(ns, opts) { globalThis.fliptArgs = JSON.stringify([ns, opts.url, opts.authentication, Object.keys(opts).length]); this.metadata = {name: "fake"} }
		resolveBooleanEvaluation(k, d) { return {value: d} }
		resolveStringEvaluation(k, d) { return {value: d} }
		resolveNumberEvaluation(k, d) { return {value: d} }
		resolveObjectEvaluation(k, d) { return {value: d} }
	}}}`)
	p, ok := New(app.Context{}, Options{URL: "https://f", ClientToken: "t"})
	if !ok || p == nil {
		t.Fatal("New failed")
	}
	want := `["default","https://f",{"clientToken":"t"},2]`
	if got := win.Get("fliptArgs").String(); got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}
