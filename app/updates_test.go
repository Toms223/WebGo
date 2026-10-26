//go:build !wasm

package app

import (
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

func TestUpdateScriptsRunInNode(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is required for the service-worker tests")
	}
	cmd := exec.Command(node, "--test", "updates_test.cjs", "browser_updates_test.cjs", "app_test.cjs")
	cmd.Dir = "gen"
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("service-worker tests: %v\n%s", err, out)
	}
}

func TestPageInlinesTheUpdaterBeforeAppJS(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		h := &Handler{DisableAutoUpdate: disabled}
		w := httptest.NewRecorder()
		routes.route("/updater-test", NewZeroComponentFactory(&Compo{}))
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/updater-test", nil))
		body := w.Body.String()
		want := "const autoReload = true;"
		if disabled {
			want = "const autoReload = false;"
		}
		updater, wasmExec := strings.Index(body, want), strings.Index(body, `src="/wasm_exec.js"`)
		if updater < 0 || wasmExec < 0 || updater > wasmExec {
			t.Errorf("DisableAutoUpdate=%v: updater with %q must come before wasm_exec.js", disabled, want)
		}
	}
}
