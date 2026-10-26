//go:build wasm

package app

import (
	"runtime"

	"github.com/Toms223/WebGo/internal/errors"
)

// GenerateStaticWebsite panics in the browser build; call it on the server.
func GenerateStaticWebsite(dir string, h *Handler, pages ...string) error {
	panic(errors.New("unsupported instruction").
		WithTag("architecture", runtime.GOARCH))
}

func wasmExecJS() string {
	return ""
}
