package topic

import (
	"strconv"
	"strings"
)

type Section struct {
	Slug   string
	Title  string
	Source string
}

func Split(source string) (string, []Section) {
	var intro strings.Builder
	var sections []Section
	seen := map[string]bool{}
	var fence fenceMarker
	for _, line := range strings.SplitAfter(source, "\n") {
		text := strings.TrimRight(line, "\r\n")
		if fence.open() {
			if fence.closedBy(text) {
				fence = fenceMarker{}
			}
		} else if opening := openingFence(text); opening.open() {
			fence = opening
		} else if title, heading := strings.CutPrefix(text, "## "); heading {
			title = strings.TrimSpace(title)
			sections = append(sections, Section{Slug: unique(slugify(title), seen), Title: title})
		}
		if len(sections) == 0 {
			intro.WriteString(line)
			continue
		}
		sections[len(sections)-1].Source += line
	}
	return intro.String(), sections
}

type fenceMarker struct {
	char   byte
	length int
}

func (f fenceMarker) open() bool {
	return f.length > 0
}

func openingFence(line string) fenceMarker {
	text := strings.TrimLeft(line, " ")
	if len(line)-len(text) > 3 || text == "" || (text[0] != '`' && text[0] != '~') {
		return fenceMarker{}
	}
	run := len(text) - len(strings.TrimLeft(text, text[:1]))
	if run < 3 {
		return fenceMarker{}
	}
	return fenceMarker{char: text[0], length: run}
}

func (f fenceMarker) closedBy(line string) bool {
	text := strings.TrimLeft(line, " ")
	if len(line)-len(text) > 3 {
		return false
	}
	rest := strings.TrimLeft(text, string(f.char))
	return len(text)-len(rest) >= f.length && strings.TrimSpace(rest) == ""
}

func slugify(title string) string {
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
