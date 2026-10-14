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
		".rail {",
		"\t.example-live button {",
		"@keyframes reveal {",
		".demo-counter {",
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
		{Rail, "rail"},
		{RailLink, "rail-link"},
		{RailLinkActive, "rail-link-active"},
		{Pager, "pager"},
		{PagerLink, "pager-link"},
		{Topic, "topic"},
		{TopicHeader, "topic-header"},
		{TopicTitle, "topic-title"},
		{TopicSection, "topic-section"},
		{SectionHeading, "section-heading"},
		{TopicLoading, "topic-loading"},
		{TopicError, "topic-error"},
		{Doc, "doc"},
		{DocHeading, "doc-heading"},
		{DocSubheading, "doc-subheading"},
		{Note, "note"},
		{Example, "example"},
		{ExampleBar, "example-bar"},
		{ExampleTab, "example-tab"},
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
	for _, rule := range []*style.Rule{TopicLoading, ShellLoading} {
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

func TestDarkPaletteKeepsStructureVisible(t *testing.T) {
	css := Root.CSS()
	for _, want := range []string{
		"--line: #9791AE;",
		"--shadow-ink: #4A4760;",
		"--focus: #F3F0EA;",
		"--mint: #1F6B52;",
		"--line: #15131A;",
		"--focus: #15131A;",
	} {
		if !strings.Contains(css, want) {
			t.Errorf("Root lacks %q:\n%s", want, css)
		}
	}
	if strings.Contains(css, "#09090B") {
		t.Errorf("the near-black dark line is back:\n%s", css)
	}
}

func TestFocusRingUsesTheContrastToken(t *testing.T) {
	for _, rule := range []*style.Rule{Button, NavLink, RailLink, PagerLink, DocLink} {
		if want := "outline: 3px solid var(--focus);"; !strings.Contains(rule.CSS(), want) {
			t.Errorf("%s lacks %q:\n%s", rule.Selector(), want, rule.CSS())
		}
	}
	for _, rule := range []*style.Rule{ExampleTab, CopyButton} {
		if want := "outline: 3px solid var(--on-fill);"; !strings.Contains(rule.CSS(), want) {
			t.Errorf("%s lacks %q on the filled bar:\n%s", rule.Selector(), want, rule.CSS())
		}
	}
}

func TestHitAreaIsAnchoredToItsOwnControl(t *testing.T) {
	for _, rule := range []*style.Rule{ExampleTab, CopyButton} {
		if want := "position: relative;"; !strings.Contains(rule.CSS(), want) {
			t.Errorf("%s lacks %q, so its hit area would cover the whole panel:\n%s", rule.Selector(), want, rule.CSS())
		}
	}
}

func TestSmallControlsKeepATouchSizedHitArea(t *testing.T) {
	for _, rule := range []*style.Rule{ExampleTab, CopyButton} {
		if want := "inset: -10px 0px;"; !strings.Contains(rule.CSS(), want) {
			t.Errorf("%s lacks the hit area %q:\n%s", rule.Selector(), want, rule.CSS())
		}
	}
	for _, rule := range []*style.Rule{RailLink, SideMenu, ShellToggle, Logo, RailSummary} {
		if want := "min-height: 44px;"; !strings.Contains(rule.CSS(), want) {
			t.Errorf("%s lacks %q:\n%s", rule.Selector(), want, rule.CSS())
		}
	}
}

func TestLongInlineCodeWrapsButCodeBlocksDoNot(t *testing.T) {
	if want := "overflow-wrap: anywhere;"; !strings.Contains(DocCode.CSS(), want) {
		t.Errorf("DocCode lacks %q:\n%s", want, DocCode.CSS())
	}
	if want := "overflow-wrap: normal;"; !strings.Contains(DocPreCode.CSS(), want) {
		t.Errorf("DocPreCode lacks %q:\n%s", want, DocPreCode.CSS())
	}
}

func TestSkipLinkIsHiddenUntilFocused(t *testing.T) {
	css := Skip.CSS()
	for _, want := range []string{"translateY(-200%)", ".skip:focus-visible {"} {
		if !strings.Contains(css, want) {
			t.Errorf("Skip lacks %q:\n%s", want, css)
		}
	}
}

func TestExampleCodeDropsTheGenericCodeBlockFrame(t *testing.T) {
	for _, want := range []string{"border: 0px;", "box-shadow: none;", "border-radius: 0px;"} {
		if !strings.Contains(ExamplePre.CSS(), want) {
			t.Errorf("ExamplePre lacks %q:\n%s", want, ExamplePre.CSS())
		}
	}
	if want := ".example .example-code {"; !strings.Contains(ExamplePre.CSS(), want) {
		t.Errorf("ExamplePre must outrank the doc pre rule, want %q:\n%s", want, ExamplePre.CSS())
	}
}
