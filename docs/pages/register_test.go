package pages

import (
	"testing"
)

func TestEntriesAndScreensAgree(t *testing.T) {
	screens, err := Screens()
	if err != nil {
		t.Fatalf("building the screens failed: %v", err)
	}
	entries := Entries()
	if len(screens) != len(entries) {
		t.Fatalf("expected one navigation entry per screen, got %d entries and %d screens", len(entries), len(screens))
	}
	routes := map[string]bool{}
	for _, screen := range screens {
		routes[screen.Route()] = true
	}
	for _, entry := range entries {
		if !routes[entry.Route] {
			t.Errorf("navigation entry %q points at %q, which no screen serves", entry.Title, entry.Route)
		}
	}
}

func TestStartPageServesRoot(t *testing.T) {
	screens, err := Screens()
	if err != nil {
		t.Fatalf("building the screens failed: %v", err)
	}
	if got := screens[0].Route(); got != "/" {
		t.Errorf("expected the first screen to serve /, got %s", got)
	}
}
