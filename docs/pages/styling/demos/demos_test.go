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

func TestEveryDemoRendersADemoClass(t *testing.T) {
	for name, demo := range All() {
		if out := app.HTMLString(demo()); !strings.Contains(out, `"demo-`) {
			t.Errorf("demo %q renders no demo- class: %s", name, out)
		}
	}
}

func TestEveryDemoIsPinned(t *testing.T) {
	pins := Pins()
	if len(pins) != len(All()) {
		t.Errorf("%d pins for %d demos", len(pins), len(All()))
	}
	for name := range All() {
		if pins[name] == "" {
			t.Errorf("demo %q has no pinned CSS", name)
		}
	}
}
