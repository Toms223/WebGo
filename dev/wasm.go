//go:build wasm

package dev

import (
	"context"
	"errors"
	"os/exec"
)

// RunDevServer returns an error in the browser build; call it on the host.
func RunDevServer(ctx context.Context, tempDir, addr string, generate func(dir string) *exec.Cmd) error {
	return errors.New("dev: RunDevServer does not run in the browser")
}
