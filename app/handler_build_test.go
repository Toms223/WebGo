//go:build !wasm

package app_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/style"
)

func get(h http.Handler, path string) string {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w.Body.String()
}

func TestServeHTTPBuildsOnFirstRequest(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/t.css", plainRule("a")))
	body := get(h, "/")
	if !strings.Contains(body, `href="/web/t.css"`) {
		t.Errorf("page does not link the generated stylesheet:\n%s", body)
	}
	if _, err := os.Stat(filepath.Join(h.WebDir, "t.css")); err != nil {
		t.Errorf("generated stylesheet not written: %v", err)
	}
}

func TestServeHTTPFailsEveryRequestOnBuildError(t *testing.T) {
	h := handlerFor(t, style.Static("/web/missing.css"))
	for _, path := range []string{"/", "/", "/app.js", "/web/x.css"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)
		r.Header.Set("If-None-Match", `""`)
		h.ServeHTTP(w, r)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("GET %s = %d, want 500", path, w.Code)
		}
	}
}

func TestGenerateStaticWebsiteReturnsBuildErrors(t *testing.T) {
	h := handlerFor(t, style.Static("/web/missing.css"))
	if err := app.GenerateStaticWebsite(t.TempDir(), h); !errors.Is(err, app.ErrMissingStyle) {
		t.Fatalf("GenerateStaticWebsite() error = %v, want %v", err, app.ErrMissingStyle)
	}
}

func TestWebDirFollowsLocalDir(t *testing.T) {
	dir := t.TempDir()
	h := &app.Handler{
		Resources: app.LocalDir(dir),
		Styles:    []style.Style{style.Generated("/web/t.css", plainRule("a"))},
	}
	if err := h.Build(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "web", "t.css")); err != nil {
		t.Errorf("generated stylesheet not under the LocalDir's web folder: %v", err)
	}
}

type library string

func (l library) Styles() (string, string) { return string(l), "a{}" }

func TestLibraryLinksAreSortedAfterAppCSS(t *testing.T) {
	h := handlerFor(t, style.Generated("/web/t.css", plainRule("a")))
	for _, l := range []string{"/z.css", "/m.css", "/a.css"} {
		h.Libraries = append(h.Libraries, library(l))
	}
	body := get(h, "/")
	order := []string{`href="/app.css"`, `href="/a.css"`, `href="/m.css"`, `href="/z.css"`, `href="/web/t.css"`}
	last := -1
	for _, want := range order {
		i := strings.Index(body, want)
		if i <= last {
			t.Fatalf("%s is missing or out of order in:\n%s", want, body)
		}
		last = i
	}
}

func TestWebFilesAreNotValidatedByVersion(t *testing.T) {
	dir := t.TempDir()
	putFile(t, filepath.Join(dir, "web", "logo.svg"), "<svg/>")
	h := &app.Handler{Resources: app.LocalDir(dir), Version: "v1"}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/app.js", nil))
	etag := w.Header().Get("ETag")

	r := httptest.NewRequest(http.MethodGet, "/web/logo.svg", nil)
	r.Header.Set("If-None-Match", etag)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK || w.Body.String() != "<svg/>" {
		t.Fatalf("GET /web/logo.svg with the version ETag = %d %q, want 200 with the file", w.Code, w.Body.String())
	}
	if got := w.Header().Get("ETag"); got == etag {
		t.Errorf("/web/ file carries the version ETag %s", got)
	}
}
