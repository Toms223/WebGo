package demos

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Toms223/WebGo/app"
)

type host struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (h *host) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *host) Render() app.UI { return app.Div() }

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
		if out := app.HTMLString(e.Live()); !strings.Contains(out, `"demo-browser-`) {
			t.Errorf("example %s renders no demo-browser- class: %s", e.Name, out)
		}
	}
}

func TestTimerTicksUntilStopped(t *testing.T) {
	var ctx app.Context
	engine := app.NewTestEngine()
	if err := engine.Load(&host{captured: &ctx}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Load(&host{Tick: 1}); err != nil {
		t.Fatal(err)
	}
	engine.ConsumeAll()
	d := TimerDemo().(*timerDemo)
	d.toggle(ctx, app.Event{})
	time.Sleep(350 * time.Millisecond)
	d.toggle(ctx, app.Event{})
	engine.ConsumeAll()
	if d.stop != nil {
		t.Fatal("expected the second toggle to stop the timer")
	}
	if d.tenths < 1 {
		t.Fatalf("expected at least one tick, got %d", d.tenths)
	}
	if out := app.HTMLString(d.Render()); !strings.Contains(out, "Start") {
		t.Errorf("expected the button to offer Start again, got %s", out)
	}
}
