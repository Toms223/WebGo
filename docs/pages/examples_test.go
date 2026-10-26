package pages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/example"
	actions "github.com/Toms223/WebGo/docs/pages/actions/demos"
	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	browser "github.com/Toms223/WebGo/docs/pages/browser/demos"
	components "github.com/Toms223/WebGo/docs/pages/components/demos"
	state "github.com/Toms223/WebGo/docs/pages/state/demos"
	styling "github.com/Toms223/WebGo/docs/pages/styling/demos"
)

func examplePackages() map[string][]example.Example {
	return map[string][]example.Example{
		"styling/demos":    styling.Examples(),
		"applying/demos":   applying.Examples(),
		"state/demos":      state.Examples(),
		"components/demos": components.Examples(),
		"actions/demos":    actions.Examples(),
		"browser/demos":    browser.Examples(),
	}
}

func packageSource(t *testing.T, dir string) string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil || len(files) == 0 {
		t.Fatalf("listing %s failed: %v (%d files)", dir, err, len(files))
	}
	var all strings.Builder
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("reading %s failed: %v", file, err)
		}
		all.Write(source)
	}
	return all.String()
}

func TestShownSourceIsTheCode(t *testing.T) {
	for dir, examples := range examplePackages() {
		source := packageSource(t, dir)
		for _, e := range examples {
			if strings.Count(source, e.Source) < 2 {
				t.Errorf("%s: the source shown for %s is not the code of the package", dir, e.Name)
			}
		}
	}
}

func TestEveryExampleRendersAndHasCSS(t *testing.T) {
	for dir, examples := range examplePackages() {
		for _, e := range examples {
			if e.Name == "" || e.Source == "" || e.Live == nil {
				t.Errorf("%s: example %q is incomplete", dir, e.Name)
				continue
			}
			if dir != "state/demos" && e.CSS == "" {
				t.Errorf("%s: example %s shows no CSS", dir, e.Name)
			}
			if out := app.HTMLString(e.Live()); !strings.Contains(out, `"demo-`) {
				t.Errorf("%s: example %s renders no demo- class: %s", dir, e.Name, out)
			}
		}
	}
}
