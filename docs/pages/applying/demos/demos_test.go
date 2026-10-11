package demos

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/serve"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func TestItemsBuild(t *testing.T) {
	handler := &serve.Handler{
		Styles: []style.Style{
			style.Generated("/web/demos.css", Items()...),
			style.Dynamic(Dynamic()...),
		},
		WebDir: t.TempDir(),
	}
	if _, err := handler.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestEveryDemoMounts(t *testing.T) {
	for name, demo := range All() {
		compo, ok := demo().(app.Composer)
		if !ok {
			t.Errorf("demo %q is not a component", name)
			continue
		}
		engine := app.NewTestEngine()
		if err := engine.Load(compo); err != nil {
			t.Errorf("mounting demo %q failed: %v", name, err)
		}
		engine.ConsumeAll()
	}
}

func TestEveryDemoIsPinned(t *testing.T) {
	pins := Pins()
	if len(pins) != len(All()) {
		t.Errorf("%d pins for %d demos", len(pins), len(All()))
	}
	for name := range All() {
		if pins[name] == "" {
			t.Errorf("demo %q has no pinned CSS", name)
		}
	}
}

func TestBindingsFollowTheState(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"compact on", DemoCompact.CSSFor(CompactKey, &Compact{On: true}), DemoCompact.CSS()},
		{"compact off", DemoCompact.CSSFor(CompactKey, &Compact{}), ""},
		{"compact wrong type", DemoCompact.CSSFor(CompactKey, true), ""},
		{"banner shown", DemoBanner.CSSFor(BannerKey, &Banner{}), DemoBanner.CSS()},
		{"banner hidden", DemoBanner.CSSFor(BannerKey, &Banner{Hidden: true}), ""},
		{"outline on", DemoOutline.CSSFor(OutlineAction, true), DemoOutline.CSS()},
		{"outline off", DemoOutline.CSSFor(OutlineAction, false), ""},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestComputedRulesFollowTheState(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"accent", DemoAccent.CSSFor(AccentKey, &Accent{Color: Blue}), "color: " + Blue + ";"},
		{"two beats", DemoPulse.CSSFor(SpeedKey, &Speed{Beats: 2}), "25% {"},
		{"four beats", DemoPulse.CSSFor(SpeedKey, &Speed{Beats: 4}), "12.5% {"},
	}
	for _, c := range cases {
		if !strings.Contains(c.got, c.want) {
			t.Errorf("%s: %q lacks %q", c.name, c.got, c.want)
		}
	}
}
