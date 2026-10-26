package app

import (
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (h *Handler) contentVersion() string {
	sum := sha1.New()
	write := func(v any) { fmt.Fprintf(sum, "%q\n", fmt.Sprint(v)) }

	for _, s := range []string{DefaultAppWorkerJS, h.ServiceWorkerTemplate, appJS, appCSS, updaterJS} {
		write(s)
	}
	for _, v := range []any{
		h.Name, h.ShortName, h.Title, h.Description, h.Lang, h.Author, h.Keywords,
		h.Image, h.LoadingLabel, h.BackgroundColor, h.ThemeColor, h.Icon, h.Domain,
		h.RawHeaders, h.Scripts, h.Fonts, h.CacheableResources, h.Preconnect, h.InternalURLs,
		h.WasmContentLength, h.WasmContentLengthHeader, h.styleLinks, h.DisableAutoUpdate,
		fmt.Sprintf("%T", h.Resources), wasmExecJS(),
	} {
		write(v)
	}
	if h.Resources != nil {
		write(h.Resources.Resolve("/"))
	}
	keys := make([]string, 0, len(h.Env))
	for k := range h.Env {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		write(k + "=" + h.Env[k])
	}
	for _, l := range h.Libraries {
		path, styles := l.Styles()
		write(path + "\n" + styles)
	}

	files := slices.Concat(
		[]string{"/web/app.wasm", h.Image, h.Icon.Default, h.Icon.Large, h.Icon.Maskable, h.Icon.SVG},
		h.styleLinks, h.Scripts, h.Fonts, h.CacheableResources,
	)
	webDir := h.webDir()
	for _, p := range files {
		u := parseHTTPResource(p).URL
		if u == "" || remoteLocation(u) {
			continue
		}
		rest, ok := strings.CutPrefix("/"+strings.TrimLeft(u, "/"), webPrefix)
		if !ok || !filepath.IsLocal(rest) {
			continue
		}
		write(u)
		if b, err := os.ReadFile(filepath.Join(webDir, rest)); err == nil {
			sum.Write(b)
		} else {
			write("missing")
		}
	}
	return fmt.Sprintf("%x", sum.Sum(nil))
}
