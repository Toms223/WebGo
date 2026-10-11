package theme

import (
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	"github.com/Toms223/WebGo/serve"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func TestStylesBuild(t *testing.T) {
	dir := t.TempDir()
	source, err := os.ReadFile(filepath.Join("..", "web", "app.css"))
	if err != nil {
		t.Fatalf("reading the static sheet failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.css"), source, 0o644); err != nil {
		t.Fatalf("copying the static sheet failed: %v", err)
	}

	handler := &serve.Handler{Styles: Styles(), WebDir: dir}
	built, err := handler.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if want := []string{"/web/app.css", "/web/theme.css"}; !slices.Equal(built.Styles, want) {
		t.Errorf("Styles = %q, want %q", built.Styles, want)
	}

	generated, err := os.ReadFile(filepath.Join(dir, "theme.css"))
	if err != nil {
		t.Fatalf("theme.css was not written: %v", err)
	}
	for _, want := range []string{
		"@import url(\"https://fonts.googleapis.com/css2?family=Archivo",
		"@layer base, components;",
		"@layer base {",
		"@keyframes pulse {",
		"animation: none !important;",
		"@container (max-width: 20rem) {",
		"@supports (color: color-mix(in srgb, red, blue)) {",
		"& ",
		".side {",
		".sections {",
		"@media (prefers-color-scheme: dark) {",
		"@media (max-width: 48rem) {",
		"\t&::before {\n\t\t\t@media (prefers-reduced-motion: reduce) {\n\t\t\t\tanimation: none !important;",
		"\t&::after {\n\t\t\t@media (prefers-reduced-motion: reduce) {\n\t\t\t\tanimation: none !important;",
		"\t.demo-card {\n",
		"\t@keyframes demo-pulse {\n",
		"\t@property --demo-angle {\n",
		"\t.demo-compact {\n",
	} {
		if !strings.Contains(string(generated), want) {
			t.Errorf("theme.css lacks %q", want)
		}
	}
}

func TestForceRulesFollowTheAction(t *testing.T) {
	cases := []struct {
		value any
		dark  bool
		light bool
	}{
		{value: true, dark: true},
		{value: false, light: true},
		{value: "yes"},
		{value: nil},
	}
	for _, c := range cases {
		action := app.Action{Name: DarkAction, Value: c.value}
		if got := darkRequested(action); got != c.dark {
			t.Errorf("darkRequested(%#v) = %v, want %v", c.value, got, c.dark)
		}
		if got := lightRequested(action); got != c.light {
			t.Errorf("lightRequested(%#v) = %v, want %v", c.value, got, c.light)
		}
	}
}

func TestNamesMatchTheMarkup(t *testing.T) {
	cases := []struct {
		rule *style.Rule
		want string
	}{
		{Shell, "shell"},
		{MenuOpen, "menu-open"},
		{Side, "side"},
		{Logo, "logo"},
		{LogoMark, "logo-mark"},
		{Landing, "landing"},
		{Topbar, "topbar"},
		{SideMenu, "side-menu"},
		{SideNav, "side-nav"},
		{Content, "content"},
		{ShellToggle, "shell-toggle"},
		{ShellLoading, "shell-loading"},
		{ShellError, "shell-error"},
		{NavLink, "nav-link"},
		{NavLinkActive, "nav-link-active"},
		{Sections, "sections"},
		{SectionLink, "section-link"},
		{SectionLinkActive, "section-link-active"},
		{Topic, "topic"},
		{TopicTitle, "topic-title"},
		{TopicLoading, "topic-loading"},
		{TopicError, "topic-error"},
		{Doc, "doc"},
		{DocHeading, "doc-heading"},
		{DocSubheading, "doc-subheading"},
		{DocEmpty, "doc-empty"},
		{Example, "example"},
		{ExampleCode, "example-code"},
		{ExampleLive, "example-live"},
		{ExampleLabel, "example-label"},
	}
	for _, c := range cases {
		if got := c.rule.Name(); got != c.want {
			t.Errorf("rule %q has Name() %q, want %q", c.rule.Selector(), got, c.want)
		}
	}
}

func TestStatusRulesPulse(t *testing.T) {
	for _, rule := range []*style.Rule{TopicLoading, ShellLoading, DocEmpty} {
		if want := "animation-name: " + Pulse.Name() + ";"; !strings.Contains(rule.CSS(), want) {
			t.Errorf("%s lacks %q:\n%s", rule.Selector(), want, rule.CSS())
		}
	}
}

func TestForcedPalettesAreTheDynamicStyle(t *testing.T) {
	if want := "--ground: #17161C;"; !strings.Contains(ForceDark.CSS(), want) {
		t.Errorf("ForceDark lacks %q:\n%s", want, ForceDark.CSS())
	}
	if want := "--ground: #F4F0E8;"; !strings.Contains(ForceLight.CSS(), want) {
		t.Errorf("ForceLight lacks %q:\n%s", want, ForceLight.CSS())
	}
	want := style.Dynamic(append([]style.Item{ForceDark, ForceLight}, applying.Dynamic()...)...)
	if got := Styles()[2]; !reflect.DeepEqual(got, want) {
		t.Errorf("Styles()[2] = %#v, want %#v", got, want)
	}
}

func TestNoClassLiteralsInLayouts(t *testing.T) {
	for _, root := range []string{
		filepath.Join("..", "components"),
		filepath.Join("..", "pages", "base"),
	} {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return err
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if count := strings.Count(string(source), `.Class("`); count > 0 {
				t.Errorf("%s sets %d class literal(s); use theme rule names", path, count)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walking %s failed: %v", root, err)
		}
	}
}
