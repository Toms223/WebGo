package pages

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/docs/components/topic"
)

var (
	linkPattern    = regexp.MustCompile(`doc\.A\("([^"]*)"`)
	routePattern   = regexp.MustCompile(`Route\s*=\s*"([^"]+)"`)
	sectionPattern = regexp.MustCompile(`\{Title:\s*"([^"]+)",\s*Body:`)
	headingPattern = regexp.MustCompile(`doc\.H\("([^"]+)"\)`)
)

func pageSources(t *testing.T) map[string]string {
	t.Helper()
	dirs, err := filepath.Glob("*")
	if err != nil {
		t.Fatal(err)
	}
	sources := map[string]string{}
	for _, dir := range dirs {
		if info, err := os.Stat(dir); err != nil || !info.IsDir() {
			continue
		}
		files, _ := filepath.Glob(filepath.Join(dir, "*.go"))
		var all strings.Builder
		for _, file := range files {
			if strings.HasSuffix(file, "_test.go") {
				continue
			}
			b, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			all.Write(b)
		}
		sources[dir] = all.String()
	}
	return sources
}

func anchors(source string) []string {
	var out []string
	for _, p := range []*regexp.Regexp{sectionPattern, headingPattern} {
		for _, m := range p.FindAllStringSubmatch(source, -1) {
			out = append(out, topic.Slugify(m[1]))
		}
	}
	return out
}

func TestInternalLinksResolve(t *testing.T) {
	sources := pageSources(t)
	byRoute := map[string]string{}
	for dir, source := range sources {
		if m := routePattern.FindStringSubmatch(source); m != nil {
			byRoute[m[1]] = dir
		}
	}
	for dir, source := range sources {
		for _, m := range linkPattern.FindAllStringSubmatch(source, -1) {
			href := m[1]
			if strings.Contains(href, "://") {
				continue
			}
			route, fragment, _ := strings.Cut(href, "#")
			target := dir
			if route != "" {
				var ok bool
				if target, ok = byRoute[route]; !ok {
					t.Errorf("%s links to %q, which no page serves", dir, href)
					continue
				}
			}
			if fragment != "" && !slices.Contains(anchors(sources[target]), fragment) {
				t.Errorf("%s links to %q, but %s has no section or heading with that anchor", dir, href, target)
			}
		}
	}
}
