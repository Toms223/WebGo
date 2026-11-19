//go:build !wasm

package jsprovider

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
)

// New adapts a JavaScript OpenFeature web provider. Outside the browser there
// is no JavaScript, so every resolver returns the default with
// PROVIDER_NOT_READY.
func New(provider app.Value) openfeature.Provider { return &stub{} }

type stub struct{ _ byte }

func (*stub) Metadata() openfeature.Metadata { return openfeature.Metadata{Name: "jsprovider"} }

func notReady[T any](def T) openfeature.ResolutionDetails[T] {
	return openfeature.ResolutionDetails[T]{
		Value: def, Reason: openfeature.ReasonError, ErrorCode: openfeature.CodeProviderNotReady,
		ErrorMessage: "JavaScript providers run only in the browser",
	}
}

func (*stub) ResolveBoolean(_ string, def bool, _ openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[bool] {
	return notReady(def)
}

func (*stub) ResolveString(_ string, def string, _ openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[string] {
	return notReady(def)
}

func (*stub) ResolveNumber(_ string, def float64, _ openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[float64] {
	return notReady(def)
}

func (*stub) ResolveObject(_ string, def any, _ openfeature.Context, _ openfeature.Logger) openfeature.ResolutionDetails[any] {
	return notReady(def)
}
