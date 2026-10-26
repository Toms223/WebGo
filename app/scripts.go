//go:build !wasm

package app

import _ "embed"

// DefaultAppWorkerJS is the default service worker template. It precaches the
// app per Version, answers only precached GET requests from the cache, and
// lets the page reload once a new version activates.
//
//go:embed gen/app-worker.js
var DefaultAppWorkerJS string

//go:embed gen/app.js
var appJS string

//go:embed gen/app.css
var appCSS string

//go:embed gen/app-updates.js
var updaterJS string

//go:embed gen/wasm_exec.js
var wasmExecJSSource string
