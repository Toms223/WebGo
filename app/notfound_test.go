//go:build !wasm

package app

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type customNotFound struct{ Compo }

func (c *customNotFound) Render() UI { return Div().ID("custom-404").Text("Nothing here") }

func TestUnknownRouteServesNotFoundWith404(t *testing.T) {
	w := httptest.NewRecorder()
	(&Handler{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/no-such-route", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if body := w.Body.String(); !strings.Contains(body, "<html") || !strings.Contains(body, "Not Found") {
		t.Fatalf("want the app page with the default not-found component, got:\n%s", body)
	}
}

func TestReplacingNotFoundTakesEffect(t *testing.T) {
	defer func(ui UI) { NotFound = ui }(NotFound)
	NotFound = &customNotFound{}

	w := httptest.NewRecorder()
	(&Handler{}).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/no-such-route", nil))
	if w.Code != http.StatusNotFound || !strings.Contains(w.Body.String(), `id="custom-404"`) {
		t.Fatalf("status %d, want 404 with the custom component, got:\n%s", w.Code, w.Body.String())
	}

	e := NewTestEngine().(*engineX)
	u, _ := url.Parse("/another-missing-route")
	e.Navigate(u, false)
	e.ConsumeAll()
	if html := HTMLString(e.body); !strings.Contains(html, "Nothing here") {
		t.Fatalf("engine did not mount the custom NotFound:\n%s", html)
	}
}
