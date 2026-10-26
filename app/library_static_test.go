//go:build !wasm

package app

import (
	"os"
	"path/filepath"
	"testing"
)

type testLibrary struct{ path, styles string }

func (l testLibrary) Styles() (string, string) { return l.path, l.styles }

func TestStaticWebsiteWritesLibraryStyles(t *testing.T) {
	dir := t.TempDir()
	h := &Handler{
		Resources: LocalDir(t.TempDir()),
		Libraries: []Library{testLibrary{"/lib/ui.css", ".ui{color:red}"}, testLibrary{"/empty.css", ""}},
	}
	if err := GenerateStaticWebsite(dir, h); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "lib", "ui.css"))
	if err != nil || string(b) != ".ui{color:red}" {
		t.Fatalf("lib/ui.css = %q, %v; want the library's styles", b, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "empty.css")); err == nil {
		t.Error("a library without styles must not be written")
	}
}
