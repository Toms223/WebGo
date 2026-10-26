package demos

import (
	"os"
	"path/filepath"
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

func TestExamplesRenderAndShowTheirCode(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		b, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		source.Write(b)
	}
	for _, e := range Examples() {
		if strings.Count(source.String(), e.Source) < 2 {
			t.Errorf("the source shown for %s is not the code of the package", e.Name)
		}
		if out := app.HTMLString(e.Live()); !strings.Contains(out, `"demo-`) {
			t.Errorf("%s renders no demo- class: %s", e.Name, out)
		}
		if !strings.Contains(e.CSS, "demo-") {
			t.Errorf("%s shows no rule CSS: %q", e.Name, e.CSS)
		}
	}
}

func TestListAddsAndRemovesByID(t *testing.T) {
	d := ListDemo().(*listDemo)
	d.add(app.Context{}, app.Event{})
	if len(d.items) != 2 {
		t.Fatalf("an empty draft should add nothing, got %d items", len(d.items))
	}
	d.draft = "Third"
	d.add(app.Context{}, app.Event{})
	d.remove(1)
	if len(d.items) != 2 || d.items[0].Text != "Remove me" || d.items[1] != (entry{3, "Third"}) || d.draft != "" {
		t.Fatalf("unexpected list after add and remove: %+v, draft %q", d.items, d.draft)
	}
	d.remove(2)
	d.remove(3)
	if out := app.HTMLString(d.Render()); !strings.Contains(out, "Nothing left.") || !strings.Contains(out, "0 items") {
		t.Errorf("an empty list should render the Else branch, got %s", out)
	}
}

func TestListMounts(t *testing.T) {
	engine := app.NewTestEngine()
	if err := engine.Load(ListDemo().(app.Composer)); err != nil {
		t.Fatalf("mounting the list failed: %v", err)
	}
	engine.ConsumeAll()
}
