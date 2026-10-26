//go:build !wasm

package app

func wasmExecJS() string {
	return wasmExecJSSource
}
