//go:build !wasm

package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func serve(h *Handler, path string) string {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w.Body.String()
}

func TestOwnIconIsTheFavicon(t *testing.T) {
	routes.route("/icon-test", NewZeroComponentFactory(&Compo{}))
	h := &Handler{Icon: Icon{Default: "/web/icon.png"}}
	page := serve(h, "/icon-test")
	if !strings.Contains(page, `href="/web/icon.png" rel="icon"`) {
		t.Errorf("favicon is not the app's own icon:\n%s", page)
	}
	if strings.Contains(page, "go-app") {
		t.Errorf("page still references go-app's icons:\n%s", page)
	}
	manifest := serve(h, "/manifest.webmanifest")
	if strings.Contains(manifest, "go-app") || strings.Contains(manifest, `"src":"/"`) {
		t.Errorf("manifest has a go-app or empty icon: %s", manifest)
	}
}

func TestDefaultIconsWhenNoneIsSet(t *testing.T) {
	manifest := serve(&Handler{}, "/manifest.webmanifest")
	for _, want := range []string{"icon.png", "icon.svg"} {
		if !strings.Contains(manifest, want) {
			t.Errorf("default manifest lacks %s: %s", want, manifest)
		}
	}
}
