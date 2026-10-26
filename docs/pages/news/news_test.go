package news

import (
	"html"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
)

func TestArticlesAreNewestFirstAndUnique(t *testing.T) {
	seen := map[string]bool{}
	articles := Articles()
	for i, a := range articles {
		if seen[a.Slug] {
			t.Errorf("duplicate slug %s", a.Slug)
		}
		seen[a.Slug] = true
		if a.Title == "" || a.Teaser == "" || len(a.Short) == 0 || a.Fill == nil || a.Body == nil {
			t.Errorf("%s is missing content", a.Slug)
		}
		if i > 0 && a.Version >= articles[i-1].Version {
			t.Errorf("%s (%s) is not older than %s", a.Slug, a.Version, articles[i-1].Version)
		}
	}
	if Latest().Version != articles[0].Version {
		t.Error("Latest is not the first article")
	}
}

func TestEveryPageRenders(t *testing.T) {
	out := app.HTMLString((&feedLayout{}).Draw(nil))
	for _, a := range Articles() {
		if !strings.Contains(out, a.Title) {
			t.Errorf("feed is missing %q", a.Title)
		}
		page := app.HTMLString((&articleLayout{article: a}).Draw(nil))
		for _, want := range []string{a.Title, a.Codename, "The short version", a.SignOff} {
			if !strings.Contains(page, html.EscapeString(want)) {
				t.Errorf("%s page is missing %q", a.Slug, want)
			}
		}
	}
}

func TestCatchUpListsWhatIsNewer(t *testing.T) {
	articles := Articles()
	c := &catchUp{}
	if got := app.HTMLString(c.answer(articles)); !strings.Contains(got, "You missed 2 updates.") || !strings.Contains(got, "@v"+articles[0].Version) {
		t.Fatalf("default answer should list both newer updates and the latest go get:\n%s", got)
	}
	c.from = 2
	if got := app.HTMLString(c.answer(articles)); !strings.Contains(got, "You missed one update.") || !strings.Contains(got, articles[0].Title) {
		t.Fatalf("from the middle version, one update is missed:\n%s", got)
	}
	c.from = 1
	if got := app.HTMLString(c.answer(articles)); !strings.Contains(got, "up to date.") {
		t.Fatalf("on the latest version there is nothing to catch up on:\n%s", got)
	}
}
