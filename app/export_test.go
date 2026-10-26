//go:build !wasm

package app

func (h *Handler) StyleLinks() []string { return h.styleLinks }

var (
	WebPath        = webPath
	IsURL          = isURL
	WriteIfChanged = writeIfChanged
)
