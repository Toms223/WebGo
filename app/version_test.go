//go:build !wasm

package app

import (
	"os"
	"path/filepath"
	"testing"
)

func versionHandler(t *testing.T, dir string) *Handler {
	t.Helper()
	for name, content := range map[string]string{"app.wasm": "wasm", "icon.png": "png", "site.css": "body{}"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for name, content := range map[string]string{"a.js": "1", "f.woff2": "1"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return &Handler{
		Scripts: []string{"web/a.js"},
		Fonts:   []string{"/web/f.woff2 crossorigin"},
		Title:   "Site",
		Icon:    Icon{Default: "/web/icon.png"},
		Styles:  []Stylesheet{testSheet("/web/site.css")},
		WebDir:  dir,
	}
}

func initialVersion(h *Handler) string {
	h.once.Do(h.init)
	return h.Version
}

func TestDefaultVersionIsStable(t *testing.T) {
	dir := t.TempDir()
	if a, b := initialVersion(versionHandler(t, dir)), initialVersion(versionHandler(t, dir)); a != b {
		t.Fatalf("same inputs gave versions %q and %q", a, b)
	}
}

func TestDefaultVersionTracksContent(t *testing.T) {
	write := func(dir, name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	changes := map[string]func(dir string, h *Handler){
		"icon bytes":       func(dir string, _ *Handler) { write(dir, "icon.png", "other") },
		"static css bytes": func(dir string, _ *Handler) { write(dir, "site.css", "p{}") },
		"wasm bytes":       func(dir string, _ *Handler) { write(dir, "app.wasm", "other") },
		"title":            func(_ string, h *Handler) { h.Title = "Other" },
		"raw header":       func(_ string, h *Handler) { h.RawHeaders = []string{"<meta name=x>"} },
		"auto update":      func(_ string, h *Handler) { h.DisableAutoUpdate = true },
		"relative script":  func(dir string, _ *Handler) { write(dir, "a.js", "2") },
		"attributed font":  func(dir string, _ *Handler) { write(dir, "f.woff2", "2") },
		"internal urls":    func(_ string, h *Handler) { h.InternalURLs = []string{"https://x"} },
	}
	for name, change := range changes {
		dir := t.TempDir()
		base := initialVersion(versionHandler(t, dir))
		h := versionHandler(t, dir)
		change(dir, h)
		if got := initialVersion(h); got == base {
			t.Errorf("%s: version did not change", name)
		}
	}
}

func TestExplicitVersionWins(t *testing.T) {
	h := versionHandler(t, t.TempDir())
	h.Version = "v7"
	if got := initialVersion(h); got != "v7" {
		t.Fatalf("version = %q, want v7", got)
	}
}

func TestDefaultVersionIgnoresMissingFiles(t *testing.T) {
	a := initialVersion(&Handler{Icon: Icon{Default: "/web/nope.png"}, WebDir: t.TempDir()})
	b := initialVersion(&Handler{Icon: Icon{Default: "/web/nope.png"}, WebDir: t.TempDir()})
	if a != b {
		t.Fatalf("missing files gave different versions %q and %q", a, b)
	}
}
