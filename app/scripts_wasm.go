package app

var (
	// DefaultAppWorkerJS is empty in the browser build; only the server uses
	// it.
	DefaultAppWorkerJS string
	appJS              string
	appCSS             string
	updaterJS          string
)
