package nav

import (
	"strings"
	"testing"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

var entries = []Entry{
	{Title: "Getting Started", Route: "/"},
	{Title: "Markdown", Route: "/markdown"},
	{Title: "Pages and State", Route: "/state"},
}

func TestRenderListsEveryEntry(t *testing.T) {
	out := app.HTMLString(Render(entries, "/"))
	for _, entry := range entries {
		if !strings.Contains(out, entry.Title) {
			t.Errorf("expected %q in the navigation, got %s", entry.Title, out)
		}
		if !strings.Contains(out, `href="`+entry.Route+`"`) {
			t.Errorf("expected a link to %q, got %s", entry.Route, out)
		}
	}
}

func TestRenderMarksOnlyTheActiveEntry(t *testing.T) {
	out := app.HTMLString(Render(entries, "/markdown"))
	if count := strings.Count(out, "nav-link-active"); count != 1 {
		t.Errorf("expected exactly one active link, found %d in %s", count, out)
	}
}

func TestRenderMarksNothingForAnUnknownRoute(t *testing.T) {
	out := app.HTMLString(Render(entries, "/missing"))
	if strings.Contains(out, "nav-link-active") {
		t.Errorf("an unknown route must not mark any entry active, got %s", out)
	}
}
