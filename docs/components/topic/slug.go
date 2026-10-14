package topic

import (
	"strconv"
	"strings"
)

type section struct {
	Section
	Slug string
}

func slugged(sections []Section) []section {
	seen := map[string]bool{}
	out := make([]section, 0, len(sections))
	for _, s := range sections {
		out = append(out, section{Section: s, Slug: unique(Slugify(s.Title), seen)})
	}
	return out
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
