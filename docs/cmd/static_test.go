package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Toms223/WebGo/docs/pages"
)

func TestGenerateWritesIndex(t *testing.T) {
	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate returned an error: %v", err)
	}
	for _, name := range []string{"index.html", "app.js", "manifest.webmanifest"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
}

func TestHandlerCarriesSiteMetadata(t *testing.T) {
	h := Handler()
	if h.Name == "" {
		t.Error("handler Name must be set")
	}
	if h.Title == "" {
		t.Error("handler Title must be set")
	}
}

func TestGenerateWritesEveryRoute(t *testing.T) {
	pages.Register()

	dir := t.TempDir()
	if err := Generate(dir); err != nil {
		t.Fatalf("Generate returned an error: %v", err)
	}
	for _, name := range []string{"index.html", "markdown.html", "state.html"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to be generated: %v", name, err)
		}
	}
}
