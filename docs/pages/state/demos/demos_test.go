package demos

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
)

type host struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (h *host) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *host) Render() app.UI { return app.Div() }

func TestCounterMountsAndCounts(t *testing.T) {
	var ctx app.Context
	engine := app.NewTestEngine()
	if err := engine.Load(&host{captured: &ctx}); err != nil {
		t.Fatalf("mounting the host failed: %v", err)
	}
	if err := engine.Load(&host{Tick: 1}); err != nil {
		t.Fatalf("updating the host failed: %v", err)
	}
	engine.ConsumeAll()
	demo := CounterDemo().(*counterDemo)
	demo.bump(ctx, app.Event{})
	demo.bump(ctx, app.Event{})
	engine.ConsumeAll()
	if got := counter.State().Count; got != 2 {
		t.Fatalf("expected two bumps to count 2, got %d", got)
	}
	if out := app.HTMLString(demo.Render()); !strings.Contains(out, "<output>2</output>") {
		t.Errorf("expected the count in the output, got %s", out)
	}
	demo.reset(ctx, app.Event{})
	engine.ConsumeAll()
	if got := counter.State().Count; got != 0 {
		t.Fatalf("expected reset to count 0, got %d", got)
	}
}
