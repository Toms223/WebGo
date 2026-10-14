package topic

import (
	"strconv"
	"strings"

	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type sub struct {
	Title string
	ID    string
}

type section struct {
	Section
	Slug string
	Subs []sub
}

func slugged(sections []Section) []section {
	seen := map[string]bool{topicID: true, sectionID: true, headingID: true}
	out := make([]section, 0, len(sections))
	for _, s := range sections {
		out = append(out, section{Section: s, Slug: unique(Slugify(s.Title), seen)})
	}
	for i := range out {
		if out[i].Body == nil {
			continue
		}
		for _, heading := range headings(out[i].Body()) {
			out[i].Subs = append(out[i].Subs, sub{Title: heading.Title, ID: unique(Slugify(heading.Title), seen)})
		}
	}
	return out
}

func headings(body []app.UI) []*doc.Sub {
	var found []*doc.Sub
	for _, ui := range body {
		if heading, ok := ui.(*doc.Sub); ok {
			found = append(found, heading)
		}
	}
	return found
}

func anchor(body []app.UI, subs []sub) []app.UI {
	for i, heading := range headings(body) {
		if i < len(subs) {
			heading.ID = subs[i].ID
		}
	}
	return body
}

func Slugify(title string) string {
	var slug strings.Builder
	dash := false
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			slug.WriteRune(r)
			dash = false
			continue
		}
		if !dash && slug.Len() > 0 {
			slug.WriteByte('-')
			dash = true
		}
	}
	trimmed := strings.TrimSuffix(slug.String(), "-")
	if trimmed == "" {
		return "section"
	}
	return trimmed
}

func unique(slug string, seen map[string]bool) string {
	candidate := slug
	for n := 2; seen[candidate]; n++ {
		candidate = slug + "-" + strconv.Itoa(n)
	}
	seen[candidate] = true
	return candidate
}
