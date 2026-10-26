package demos

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/style"
)

func TestItemsBuild(t *testing.T) {
	handler := &app.Handler{
		Styles: []style.Style{style.Generated("/web/demos.css", Items()...)},
		WebDir: t.TempDir(),
	}
	if err := handler.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestEveryExampleRendersADemoClass(t *testing.T) {
	for _, e := range Examples() {
		if out := app.HTMLString(e.Live()); !strings.Contains(out, `"demo-`) {
			t.Errorf("example %s renders no demo- class: %s", e.Name, out)
		}
		if e.CSS == "" || !strings.Contains(e.CSS, "demo-") {
			t.Errorf("example %s shows no rule CSS: %q", e.Name, e.CSS)
		}
	}
}
