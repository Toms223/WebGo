//go:build !wasm

package app

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type testSource struct {
	name, url string
	check     error
}

func (s testSource) ImportMap() (string, string) { return s.name, s.url }

func (s testSource) Check() error { return s.check }

type testModule struct{ src ModuleSource }

func (m testModule) Source() ModuleSource { return m.src }

func moduleHandler(t *testing.T, imports ...Module) *Handler {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "m.js"), []byte("export {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	return &Handler{Imports: imports, WebDir: dir}
}

func localModule(name, url string) Module { return testModule{testSource{name: name, url: url}} }

func TestBuildAcceptsLocalAndRemoteModules(t *testing.T) {
	h := moduleHandler(t, localModule("m", "/web/m.js"), localModule("x", "https://esm.sh/x@1.0.0"))
	if err := h.Build(); err != nil {
		t.Fatal(err)
	}
	want := []importEntry{{"m", "/web/m.js"}, {"x", "https://esm.sh/x@1.0.0"}}
	if !reflect.DeepEqual(h.imports, want) {
		t.Fatalf("imports = %v, want %v", h.imports, want)
	}
}

func TestBuildRejectsMissingModuleFile(t *testing.T) {
	err := moduleHandler(t, localModule("m", "/web/missing.js")).Build()
	if !errors.Is(err, ErrMissingModule) || !strings.Contains(err.Error(), "missing.js") {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsModulePathOutsideWeb(t *testing.T) {
	for _, p := range []string{"/other/m.js", "/web/../m.js", "web/m.js"} {
		if err := moduleHandler(t, localModule("m", p)).Build(); !errors.Is(err, ErrPathOutsideWeb) {
			t.Errorf("%s: err = %v", p, err)
		}
	}
}

func TestBuildRejectsDuplicateModuleNames(t *testing.T) {
	err := moduleHandler(t, localModule("m", "/web/m.js"), localModule("m", "https://esm.sh/m@1.0.0")).Build()
	if !errors.Is(err, ErrDuplicateModule) || !strings.Contains(err.Error(), "m") {
		t.Fatalf("err = %v", err)
	}
}

func TestBuildRejectsInvalidModules(t *testing.T) {
	bad := errors.New("bad")
	cases := map[string]Module{
		"empty name": localModule("", "/web/m.js"),
		"nil source": testModule{},
		"check":      testModule{testSource{name: "named", url: "/web/m.js", check: bad}},
	}
	for name, m := range cases {
		err := moduleHandler(t, m).Build()
		if !errors.Is(err, ErrInvalidModule) {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	err := moduleHandler(t, cases["check"]).Build()
	if !errors.Is(err, bad) || !strings.Contains(err.Error(), "bad") || !strings.Contains(err.Error(), "named") {
		t.Fatalf("check err = %v", err)
	}
}

func TestBuildSkipsNilImports(t *testing.T) {
	h := moduleHandler(t, nil, localModule("m", "/web/m.js"))
	if err := h.Build(); err != nil {
		t.Fatal(err)
	}
	if len(h.imports) != 1 {
		t.Fatalf("imports = %v", h.imports)
	}
}

func TestPageWritesTheImportMapFirst(t *testing.T) {
	h := moduleHandler(t, localModule("m", "/web/m.js"), localModule(`x"y`, "https://esm.sh/x@1.0.0"))
	h.Resources = GitHubPages("site")
	routes.route("/importmap-test", NewZeroComponentFactory(&Compo{}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/importmap-test", nil))
	body := w.Body.String()
	want := `<script type="importmap">{"imports":{"m":"/site/web/m.js","x\"y":"https://esm.sh/x@1.0.0"}}</script>`
	if !strings.Contains(body, want) {
		t.Fatalf("body lacks %s", want)
	}
	im := strings.Index(body, `type="importmap"`)
	if wasmExec := strings.Index(body, `src="/site/wasm_exec.js"`); im > wasmExec {
		t.Errorf("importmap at %d after wasm_exec.js at %d", im, wasmExec)
	}
	if charset := strings.Index(body, `<meta charset="UTF-8">`); charset < 0 || charset > im {
		t.Errorf("importmap at %d before meta charset at %d", im, charset)
	}
}

func TestPageWithoutImportsHasNoImportMap(t *testing.T) {
	routes.route("/plain-test", NewZeroComponentFactory(&Compo{}))
	w := httptest.NewRecorder()
	(&Handler{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/plain-test", nil))
	if strings.Contains(w.Body.String(), "importmap") {
		t.Fatal("body has an importmap")
	}
}

func importsHandler(t *testing.T) *Handler {
	h := moduleHandler(t, localModule("m", "/web/m.js"), localModule("x", "https://esm.sh/x@1.0.0"))
	initialVersion(h)
	return h
}

func TestAppJSListsTheImportNames(t *testing.T) {
	if got := string(importsHandler(t).makeAppJS()); !strings.Contains(got, `const goappImports = ["m","x"];`) {
		t.Fatal("app.js lacks the import names")
	}
	h := &Handler{}
	initialVersion(h)
	if got := string(h.makeAppJS()); !strings.Contains(got, `const goappImports = [];`) {
		t.Fatal("app.js lacks the empty import list")
	}
}

func TestWorkerPrecachesModuleURLs(t *testing.T) {
	got := string(importsHandler(t).makeAppWorkerJS())
	for _, want := range []string{`"/web/m.js"`, `"https://esm.sh/x@1.0.0"`} {
		if !strings.Contains(got, want) {
			t.Errorf("worker lacks %s", want)
		}
	}
}

type configuredModule struct {
	token string
	testModule
}

func TestVersionIgnoresModuleConfig(t *testing.T) {
	version := func(token string) string {
		h := moduleHandler(t, configuredModule{token, testModule{testSource{name: "x", url: "https://esm.sh/x@1.0.0"}}})
		return initialVersion(h)
	}
	if a, b := version("a"), version("b"); a != b {
		t.Fatalf("versions differ: %q, %q", a, b)
	}
}
