// Package isolated creates OpenFeature API instances that share nothing with
// the default instance behind the openfeature package functions.
package isolated

import "github.com/Toms223/WebGo/openfeature"

// New returns an independent API instance with its own providers, contexts,
// hooks and handlers.
func New() *openfeature.API { return new(openfeature.API) }
