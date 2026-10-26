package nav

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
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

func TestRenderIsASidebarList(t *testing.T) {
	out := app.HTMLString(Render(entries, "/"))
	open, _, _ := strings.Cut(out, ">")
	if !strings.HasPrefix(open, "<nav ") || !strings.Contains(open, `class="side-nav"`) || !strings.Contains(open, `id="`+ID+`"`) {
		t.Errorf("expected the navigation to be the side-nav list with id %q, got %s", ID, out)
	}
	if count := strings.Count(out, `class="nav-link`); count != len(entries) {
		t.Errorf("expected %d nav links, found %d in %s", len(entries), count, out)
	}
}

func TestRenderLabelsEachGroupOnce(t *testing.T) {
	grouped := []Entry{
		{Title: "A", Route: "/a"},
		{Title: "B", Route: "/b", Group: "Core"},
		{Title: "C", Route: "/c", Group: "Core"},
	}
	out := app.HTMLString(Render(grouped, "/a"))
	if strings.Count(out, ">Guides<") != 1 || strings.Count(out, ">Core<") != 1 || strings.Count(out, "<ul") != 2 {
		t.Errorf("expected one Guides and one Core group, got %s", out)
	}
}
