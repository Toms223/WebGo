package diagram

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
)

func TestFlowPutsOneArrowBetweenBoxes(t *testing.T) {
	out := app.HTMLString(Flow(Box(Sun, "a", "first"), Via("then"), Box(Plain, "b", ""), Box(Lime, "c", "")))
	if n := strings.Count(out, `class="dg-arrow"`); n != 2 {
		t.Fatalf("got %d arrows, want 2:\n%s", n, out)
	}
	if strings.Count(out, ">then<") != 1 || !strings.Contains(out, ">first<") {
		t.Fatalf("missing the arrow label or a box subtitle:\n%s", out)
	}
	if n := strings.Count(out, `class="dg-step"`); n != 2 {
		t.Fatalf("each arrow must be glued to the box it points to, got %d steps:\n%s", n, out)
	}
	if !strings.Contains(out, `class="dg-box dg-sun"`) || !strings.Contains(out, `class="dg-box dg-lime"`) {
		t.Fatalf("filled boxes must keep the box class:\n%s", out)
	}
}

func TestStackArrowsPointDown(t *testing.T) {
	out := app.HTMLString(Stack(Box(Plain, "a", ""), Box(Plain, "b", "")))
	if !strings.Contains(out, `class="dg-glyph dg-down"`) {
		t.Fatalf("stack arrow is not turned down:\n%s", out)
	}
}

func TestBranchLabelsEachWay(t *testing.T) {
	out := app.HTMLString(Branch(Box(Plain, "q?", ""), When("yes", Box(Lime, "a", "")), When("no", Box(Coral, "b", ""), Box(Sun, "c", ""))))
	for _, want := range []string{">q?<", ">yes<", ">no<", ">a<", ">b<", ">c<"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in:\n%s", want, out)
		}
	}
	if n := strings.Count(out, `class="dg-way"`); n != 2 {
		t.Errorf("got %d ways, want 2", n)
	}
}

func TestFigureHasCaptionAndLanes(t *testing.T) {
	out := app.HTMLString(Figure("How it works", Lane("wasm", "In the browser", Flow(Box(Plain, "x", "")))))
	for _, want := range []string{"<figure", `class="diagram"`, "<figcaption", ">How it works<", ">wasm<", ">In the browser<"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in:\n%s", want, out)
		}
	}
}
