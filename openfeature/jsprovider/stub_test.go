//go:build !wasm

package jsprovider

import (
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
)

func TestStubNotReady(t *testing.T) {
	p := New(app.Undefined())
	if r := p.ResolveString("f", "d", openfeature.Context{}, openfeature.DefaultLogger()); r.Value != "d" || r.ErrorCode != openfeature.CodeProviderNotReady {
		t.Errorf("stub = %#v", r)
	}
}
