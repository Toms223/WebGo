package pages

import (
	"slices"
	"testing"

	"github.com/Toms223/WebGo/docs/components/nav"
)

func TestEntriesAndScreensAgree(t *testing.T) {
	screens, err := Screens()
	if err != nil {
		t.Fatalf("building the screens failed: %v", err)
	}
	entries := Entries()
	if len(screens) != len(entries)+1 {
		t.Fatalf("expected one navigation entry per screen plus the landing page, got %d entries and %d screens", len(entries), len(screens))
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

func TestLandingPageServesRoot(t *testing.T) {
	screens, err := Screens()
	if err != nil {
		t.Fatalf("building the screens failed: %v", err)
	}
	if got := screens[0].Route(); got != "/" {
		t.Errorf("expected the first screen to serve /, got %s", got)
	}
}

func TestGuidesComeBeforeCore(t *testing.T) {
	entries := Entries()
	if len(entries) == 0 {
		t.Fatal("expected navigation entries, got none")
	}
	want := nav.Entry{Title: "Migrating", Route: "/migrating", Group: "Core"}
	if got := entries[len(entries)-1]; got != want {
		t.Errorf("expected the last navigation entry to be %+v, got %+v", want, got)
	}
	core := false
	for _, entry := range entries {
		if entry.Group == "Core" {
			core = true
		} else if core {
			t.Errorf("guide %q comes after a Core entry", entry.Title)
		}
	}
}

func TestScreensCoverEveryPage(t *testing.T) {
	screens, err := Screens()
	if err != nil {
		t.Fatalf("building the screens failed: %v", err)
	}
	if len(screens) != 13 {
		t.Fatalf("expected thirteen screens, got %d", len(screens))
	}
	if got := screens[len(screens)-1].Route(); got != "/migrating" {
		t.Errorf("expected the last screen to serve /migrating, got %s", got)
	}
}

func TestNavigationOrder(t *testing.T) {
	var titles []string
	for _, entry := range Entries() {
		titles = append(titles, entry.Title)
	}
	want := []string{"Getting Started", "Markdown rendering", "Pages, layouts and state", "Styling", "Applying", "Components", "Routing", "State and actions", "Concurrency and the browser", "Handler and deployment", "Testing", "Migrating"}
	if !slices.Equal(titles, want) {
		t.Errorf("navigation titles = %q, want %q", titles, want)
	}
}
