// Package flipt runs Flipt's JavaScript OpenFeature web provider under the Go
// openfeature SDK, loading it through the Handler's importmap.
//
// Behaviour is inherited from @openfeature/flipt-web-provider 0.1.6 and
// @flipt-io/flipt-client-js 0.2.x. Flags refresh only on SetContext, so
// SetContext(GetContext()) is the refresh. An empty TargetingKey makes every
// evaluation fail with GENERAL. Context attributes are stringified, so objects
// become "[object Object]". A missing flag is GENERAL, never FLAG_NOT_FOUND.
// Initialization failure is PROVIDER_FATAL. The Flipt client is never closed.
//
// The Flipt server must allow CORS, and on Flipt v1 must list
// x-flipt-environment and If-None-Match in cors.allowed_headers.
package flipt

import (
	"path"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/jsprovider"
	"github.com/Toms223/WebGo/script"
)

const providerName = "@openfeature/flipt-web-provider"

// Options configures the Flipt provider. Empty fields are omitted, and
// Namespace defaults to "default". Set at most one of ClientToken and JWT.
type Options struct {
	Namespace   string
	Environment string
	URL         string
	ClientToken string
	JWT         string
}

type source struct{ name, url string }

func (s source) ImportMap() (string, string) { return s.name, s.url }

type dep struct{ source }

func (d dep) Source() app.ModuleSource { return d.source }

type provider struct{ dep }

func (provider) Init(_ app.Context, mod app.Value) app.Value { return mod }

type file struct{ name, cdn, local string }

var files = []file{
	{"@openfeature/core", "https://cdn.jsdelivr.net/npm/@openfeature/core@1.12.0/dist/esm/index.js", "openfeature-core.js"},
	{"@openfeature/web-sdk", "https://cdn.jsdelivr.net/npm/@openfeature/web-sdk@1.10.0/dist/esm/index.js", "openfeature-web-sdk.js"},
	{"@flipt-io/flipt-client-js/browser", "https://cdn.jsdelivr.net/npm/@flipt-io/flipt-client-js@0.2.0/dist/browser/index.mjs", "flipt-client-js.mjs"},
	{providerName, "https://cdn.jsdelivr.net/npm/@openfeature/flipt-web-provider@0.1.6/index.esm.js", "flipt-web-provider.js"},
}

func modules(url func(file) string) []app.Module {
	out := make([]app.Module, len(files))
	for i, f := range files {
		d := dep{source{f.name, url(f)}}
		if f.name == providerName {
			out[i] = provider{d}
		} else {
			out[i] = d
		}
	}
	return out
}

// Modules returns the four importmap entries for Handler.Imports, served from
// jsDelivr at pinned versions.
func Modules() []app.Module { return modules(func(f file) string { return f.cdn }) }

// LocalModules returns the same four entries served from /web/<dir>/, for
// self-hosting. Handler.Build fails with app.ErrMissingModule when a file is
// absent.
func LocalModules(dir string) []app.Module {
	return modules(func(f file) string { return path.Join("/web", dir, f.local) })
}

func jsOptions(o Options) (map[string]any, error) {
	if o.ClientToken != "" && o.JWT != "" {
		return nil, errBothAuth
	}
	opts := map[string]any{}
	if o.URL != "" {
		opts["url"] = o.URL
	}
	if o.Environment != "" {
		opts["environment"] = o.Environment
	}
	if o.ClientToken != "" {
		opts["authentication"] = map[string]any{"clientToken": o.ClientToken}
	}
	if o.JWT != "" {
		opts["authentication"] = map[string]any{"jwtToken": o.JWT}
	}
	return opts, nil
}

type bothAuthError struct{}

func (bothAuthError) Error() string { return "flipt: set ClientToken or JWT, not both" }

var errBothAuth error = bothAuthError{}

// New creates the provider. ok is false on the server, when the JavaScript
// module is not loaded, or when both ClientToken and JWT are set.
func New(ctx app.Context, o Options) (openfeature.Provider, bool) {
	opts, err := jsOptions(o)
	if err != nil {
		app.Log(err.Error())
		return nil, false
	}
	mod, ok := script.Use[app.Value](ctx, provider{dep{source{providerName, ""}}})
	if !ok {
		return nil, false
	}
	ns := o.Namespace
	if ns == "" {
		ns = "default"
	}
	return jsprovider.New(mod.Get("FliptWebProvider").New(ns, app.ValueOf(opts))), true
}
