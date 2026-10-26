//go:build !wasm

package app

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWasmExecMatchesToolchain(t *testing.T) {
	want, err := os.ReadFile(filepath.Join(runtime.GOROOT(), "lib", "wasm", "wasm_exec.js"))
	if err != nil {
		t.Fatal(err)
	}
	if wasmExecJS() != withoutLineComments(string(want)) {
		t.Fatal("app/gen/wasm_exec.js differs from the toolchain's; copy it from $(go env GOROOT)/lib/wasm")
	}
}

func withoutLineComments(src string) string {
	var lines []string
	for _, line := range strings.Split(src, "\n") {
		code := line
		if i := strings.Index(line, "//"); i >= 0 {
			code = line[:i]
		}
		code = strings.TrimRight(code, " \t")
		if code == "" && strings.TrimSpace(line) != "" {
			continue
		}
		lines = append(lines, code)
	}
	text := strings.Join(lines, "\n")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	return strings.TrimLeft(text, "\n")
}
