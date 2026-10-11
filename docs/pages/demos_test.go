package pages

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	styling "github.com/Toms223/WebGo/docs/pages/styling/demos"
	"github.com/Toms223/WebGo/markdown"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type demoPage struct {
	file  string
	demos map[string]func() app.UI
	pins  map[string]string
}

func demoPages() []demoPage {
	return []demoPage{
		{file: "styling.md", demos: styling.All(), pins: styling.Pins()},
		{file: "applying.md", demos: applying.All(), pins: applying.Pins()},
	}
}

func readData(t *testing.T, name string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join("..", "web", "data", name))
	if err != nil {
		t.Fatalf("reading %s failed: %v", name, err)
	}
	return string(source)
}

func TestEveryDemoFenceIsRegistered(t *testing.T) {
	registries := map[string]map[string]func() app.UI{}
	for _, page := range demoPages() {
		registries[page.file] = page.demos
	}
	files, err := filepath.Glob(filepath.Join("..", "web", "data", "*.md"))
	if err != nil || len(files) == 0 {
		t.Fatalf("listing the markdown files failed: %v (%d files)", err, len(files))
	}
	for _, file := range files {
		name := filepath.Base(file)
		registry := registries[name]
		for _, demo := range markdown.DemoNames(readData(t, name)) {
			if _, ok := registry[demo]; !ok {
				t.Errorf("%s has a fence for demo %q, which its page does not register", name, demo)
			}
		}
	}
}

func TestEveryDemoIsUsed(t *testing.T) {
	for _, page := range demoPages() {
		used := markdown.DemoNames(readData(t, page.file))
		for name := range page.demos {
			if !slices.Contains(used, name) {
				t.Errorf("demo %q is registered for %s but no fence names it", name, page.file)
			}
		}
	}
}

func TestShownCSSMatchesTheRule(t *testing.T) {
	for _, page := range demoPages() {
		source := readData(t, page.file)
		for name := range page.demos {
			if _, ok := page.pins[name]; !ok {
				t.Errorf("demo %q of %s has no pinned CSS", name, page.file)
			}
		}
		for name, want := range page.pins {
			got, err := cssAfter(source, name)
			if err != nil {
				t.Errorf("%s, demo %q: %s", page.file, name, err)
				continue
			}
			if normalised(got) != normalised(want) {
				t.Errorf("%s, demo %q shows\n%s\nbut the rule renders\n%s", page.file, name, got, want)
			}
		}
	}
}

func TestCSSAfterFindsTheFollowingFence(t *testing.T) {
	cases := []struct {
		source string
		css    string
		err    bool
	}{
		{source: "```go box\nx\n```\n\n```css\n.a {}\n```\n", css: ".a {}\n"},
		{source: "```go box\nx\n```\n```css\n.a {}\n.b {}\n```\n", css: ".a {}\n.b {}\n"},
		{source: "```go other\nx\n```\n\n```css\n.a {}\n```\n", err: true},
		{source: "```go box\nx\n```\n\nText.\n\n```css\n.a {}\n```\n", err: true},
		{source: "```go box\nx\n```\n\n```go\ny\n```\n", err: true},
	}
	for _, c := range cases {
		got, err := cssAfter(c.source, "box")
		if (err != nil) != c.err || got != c.css {
			t.Errorf("cssAfter(%q) = %q, %v; want %q, error %v", c.source, got, err, c.css, c.err)
		}
	}
}

func cssAfter(source, name string) (string, error) {
	lines := strings.Split(source, "\n")
	at := slices.IndexFunc(lines, func(line string) bool {
		return strings.Join(strings.Fields(line), " ") == "```go "+name
	})
	if at < 0 {
		return "", errors.New("no fence names it")
	}
	at++
	for at < len(lines) && strings.TrimSpace(lines[at]) != "```" {
		at++
	}
	at++
	for at < len(lines) && strings.TrimSpace(lines[at]) == "" {
		at++
	}
	if at >= len(lines) || strings.TrimSpace(lines[at]) != "```css" {
		return "", errors.New("the fence is not followed by a css fence")
	}
	var css strings.Builder
	for at++; at < len(lines) && strings.TrimSpace(lines[at]) != "```"; at++ {
		css.WriteString(lines[at] + "\n")
	}
	return css.String(), nil
}

func normalised(text string) string {
	return strings.TrimRight(text, "\n") + "\n"
}
