package demos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
)

func packageSource(t *testing.T) string {
	t.Helper()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var all strings.Builder
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		all.Write(b)
	}
	return all.String()
}

func TestExamplesRenderAndShowTheirCode(t *testing.T) {
	source := packageSource(t)
	for _, e := range Examples() {
		if strings.Count(source, e.Source) < 2 {
			t.Errorf("the source shown for %s is not the code of the package", e.Name)
		}
		if out := app.HTMLString(e.Live()); !strings.Contains(out, `"demo-scripts-`) {
			t.Errorf("example %s renders no demo-scripts- class: %s", e.Name, out)
		}
	}
}

func TestConfettiSourceIsPinned(t *testing.T) {
	name, url := Confetti{}.Source().ImportMap()
	if name != "canvas-confetti" || url != "https://esm.sh/canvas-confetti@1.9.3" {
		t.Fatalf("got %q %q", name, url)
	}
}
