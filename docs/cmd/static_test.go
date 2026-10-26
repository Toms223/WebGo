package cmd

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/pages"
)

func stubWebDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.css"), []byte("html {}\n"), 0o644); err != nil {
		t.Fatalf("writing the stub app.css failed: %v", err)
	}
	return dir
}

func TestGenerateWritesIndex(t *testing.T) {
	dir, web := t.TempDir(), stubWebDir(t)
	if err := Generate(dir, "", web); err != nil {
		t.Fatalf("Generate returned an error: %v", err)
	}
	for _, name := range []string{"index.html", "app.js", "manifest.webmanifest"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to exist: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(web, "theme.css")); err != nil {
		t.Errorf("expected theme.css to be written into the web directory: %v", err)
	}
}

func TestGenerateFailsWithoutTheStaticSheet(t *testing.T) {
	err := Generate(t.TempDir(), "", t.TempDir())
	if !errors.Is(err, app.ErrMissingStyle) {
		t.Errorf("Generate error = %v, want %v", err, app.ErrMissingStyle)
	}
}

func TestHandlerCarriesSiteMetadata(t *testing.T) {
	h, err := Handler("", stubWebDir(t))
	if err != nil {
		t.Fatalf("Handler returned an error: %v", err)
	}
	if h.Name == "" {
		t.Error("handler Name must be set")
	}
	if h.Title == "" {
		t.Error("handler Title must be set")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/app-worker.js", nil))
	for _, link := range []string{`"/web/app.css"`, `"/web/theme.css"`} {
		if !strings.Contains(w.Body.String(), link) {
			t.Errorf("app-worker.js does not cache %s", link)
		}
	}
}

func TestStaticCommandTakesTheWebDirectory(t *testing.T) {
	if got := newStaticCommand().Flags().Lookup("web"); got == nil || got.DefValue != "web" {
		t.Fatalf("expected a --web flag defaulting to web, got %v", got)
	}

	dir, web := t.TempDir(), stubWebDir(t)
	root := newRootCommand()
	root.SetArgs([]string{"static", "-o", dir, "--web", web})
	if err := root.Execute(); err != nil {
		t.Fatalf("the static command failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(web, "theme.css")); err != nil {
		t.Errorf("expected the static command to write theme.css into --web: %v", err)
	}
}

func TestGenerateWritesEveryRoute(t *testing.T) {
	pages.Register()

	dir := t.TempDir()
	if err := Generate(dir, "", stubWebDir(t)); err != nil {
		t.Fatalf("Generate returned an error: %v", err)
	}
	for _, name := range []string{"index.html", "start.html", "markdown.html", "state.html", "styling.html", "applying.html"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected %s to be generated: %v", name, err)
		}
	}
}
