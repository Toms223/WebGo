package demos

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/serve"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func TestItemsBuild(t *testing.T) {
	handler := &serve.Handler{
		Styles: []style.Style{style.Generated("/web/demos.css", Items()...)},
		WebDir: t.TempDir(),
	}
	if _, err := handler.Build(); err != nil {
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
