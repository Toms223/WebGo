package flipt

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Toms223/WebGo/app"
)

func entries(t *testing.T, ms []app.Module) [][2]string {
	t.Helper()
	var out [][2]string
	for _, m := range ms {
		n, u := m.Source().ImportMap()
		out = append(out, [2]string{n, u})
	}
	return out
}

func TestModulesImportMap(t *testing.T) {
	want := [][2]string{
		{"@openfeature/core", "https://cdn.jsdelivr.net/npm/@openfeature/core@1.12.0/dist/esm/index.js"},
		{"@openfeature/web-sdk", "https://cdn.jsdelivr.net/npm/@openfeature/web-sdk@1.10.0/dist/esm/index.js"},
		{"@flipt-io/flipt-client-js/browser", "https://cdn.jsdelivr.net/npm/@flipt-io/flipt-client-js@0.2.0/dist/browser/index.mjs"},
		{"@openfeature/flipt-web-provider", "https://cdn.jsdelivr.net/npm/@openfeature/flipt-web-provider@0.1.6/index.esm.js"},
	}
	if got := entries(t, Modules()); !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
}

func TestLocalModulesPaths(t *testing.T) {
	want := []string{"/web/flags/openfeature-core.js", "/web/flags/openfeature-web-sdk.js", "/web/flags/flipt-client-js.mjs", "/web/flags/flipt-web-provider.js"}
	for i, e := range entries(t, LocalModules("flags")) {
		if e[1] != want[i] {
			t.Errorf("%d: got %s want %s", i, e[1], want[i])
		}
	}
}

func TestBuildFailsWhenLocalFileMissing(t *testing.T) {
	h := app.Handler{Imports: LocalModules("flags"), WebDir: t.TempDir()}
	if err := h.Build(); !errors.Is(err, app.ErrMissingModule) {
		t.Fatalf("got %v", err)
	}
}

func TestOptionsObject(t *testing.T) {
	got, err := jsOptions(Options{URL: "u", ClientToken: "t"})
	want := map[string]any{"url": "u", "authentication": map[string]any{"clientToken": "t"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, %v", got, err)
	}
	got, _ = jsOptions(Options{Environment: "e", JWT: "j"})
	want = map[string]any{"environment": "e", "authentication": map[string]any{"jwtToken": "j"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v", got)
	}
	if got, _ = jsOptions(Options{}); len(got) != 0 {
		t.Fatalf("got %v", got)
	}
	if _, err := jsOptions(Options{ClientToken: "t", JWT: "j"}); err == nil || err.Error() != "flipt: set ClientToken or JWT, not both" {
		t.Fatalf("got %v", err)
	}
}

func TestNewReturnsFalse(t *testing.T) {
	if p, ok := New(app.Context{}, Options{ClientToken: "t", JWT: "j"}); ok || p != nil {
		t.Fatal("both auth set must fail")
	}
	if app.IsServer {
		if _, ok := New(app.Context{}, Options{}); ok {
			t.Fatal("server must fail")
		}
	}
}
