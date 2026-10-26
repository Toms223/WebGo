package demos

import (
	"strings"
	"testing"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/style"
)

func TestItemsBuild(t *testing.T) {
	handler := &app.Handler{
		Styles: []style.Style{
			style.Generated("/web/demos.css", Items()...),
			style.Dynamic(Dynamic()...),
		},
		WebDir: t.TempDir(),
	}
	if err := handler.Build(); err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestEveryExampleMounts(t *testing.T) {
	for _, e := range Examples() {
		compo, ok := e.Live().(app.Composer)
		if !ok {
			t.Errorf("example %s is not a component", e.Name)
			continue
		}
		engine := app.NewTestEngine()
		if err := engine.Load(compo); err != nil {
			t.Errorf("mounting example %s failed: %v", e.Name, err)
		}
		engine.ConsumeAll()
		if e.CSS == "" || !strings.Contains(e.CSS, "demo-") {
			t.Errorf("example %s shows no rule CSS: %q", e.Name, e.CSS)
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
