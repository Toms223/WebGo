package demos

import (
	"os"
	"path/filepath"
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
		if out := app.HTMLString(e.Live()); !strings.Contains(out, `"demo-actions-`) {
			t.Errorf("example %s renders no demo-actions- class: %s", e.Name, out)
		}
	}
}

func TestVoterSharesTheCountThroughState(t *testing.T) {
	var ctx app.Context
	engine := app.NewTestEngine()
	if err := engine.Load(&host{captured: &ctx}); err != nil {
		t.Fatal(err)
	}
	if err := engine.Load(&host{Tick: 1}); err != nil {
		t.Fatal(err)
	}
	engine.ConsumeAll()
	v := &voter{}
	v.vote(ctx, app.Event{})
	v.vote(ctx, app.Event{})
	var votes int
	ctx.GetState(VotesKey, &votes)
	if votes != 2 {
		t.Fatalf("expected 2 votes, got %d", votes)
	}
	v.reset(ctx, app.Event{})
	ctx.GetState(VotesKey, &votes)
	if votes != 0 {
		t.Fatalf("expected reset to 0, got %d", votes)
	}
}
