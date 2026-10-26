package style

import (
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/media"

	"github.com/Toms223/WebGo/app"
)

type settings struct {
	Accent  string
	Compact bool
}

func compact(s *settings) bool { return s.Compact }

func accent(s *settings) []css.Property {
	if s == nil || s.Accent == "" {
		return nil
	}
	return []css.Property{css.Custom("--c", css.Hex(s.Accent))}
}

func fading(s *settings) []Frame {
	if s == nil || s.Compact {
		return nil
	}
	return []Frame{FrameFrom(css.Opacity(css.Num(1))), FrameTo(css.Opacity(css.Num(0.4)))}
}

type cssForCase struct {
	name   string
	action string
	value  any
	want   string
}

func checkCSSFor(t *testing.T, cssFor func(string, any) string, cases []cssForCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := cssFor(c.action, c.value); got != c.want {
				t.Errorf("CSSFor(%q, %v) = %q, want %q", c.action, c.value, got, c.want)
			}
		})
	}
}

func TestWhenEvaluatesTypedValue(t *testing.T) {
	r := Class("shell").Set(css.Custom("--p", css.Px(1))).Bind(When("settings", compact))
	if err := r.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
	checkCSSFor(t, r.CSSFor, []cssForCase{
		{"true", "settings", &settings{Compact: true}, ".shell {\n\t--p: 1px;\n}\n"},
		{"false", "settings", &settings{}, ""},
		{"wrong type", "settings", "wrong", ""},
		{"value instead of pointer", "settings", settings{Compact: true}, ""},
		{"unknown action", "other", &settings{Compact: true}, ""},
	})
}

func TestComputeReplacesTheBaseBlock(t *testing.T) {
	r := Class("accent").
		Set(css.Custom("--x", css.Px(1))).
		Hover(css.Custom("--h", css.Px(2))).
		Media(media.Screen(), css.Custom("--m", css.Px(3))).
		Bind(Compute("settings", accent))
	nested := Class("card").
		Set(css.Custom("--x", css.Px(1))).
		Nest(Class("title").Set(css.Custom("--t", css.Px(3)))).
		Bind(Compute("settings", accent))

	checkCSSFor(t, r.CSSFor, []cssForCase{
		{"computed", "settings", &settings{Accent: "#0A84FF"},
			".accent {\n\t--c: #0A84FF;\n}\n\n.accent:hover {\n\t--h: 2px;\n}\n\n@media screen {\n\t.accent {\n\t\t--m: 3px;\n\t}\n}\n"},
		{"recomputed", "settings", &settings{Accent: "#FF453A"},
			".accent {\n\t--c: #FF453A;\n}\n\n.accent:hover {\n\t--h: 2px;\n}\n\n@media screen {\n\t.accent {\n\t\t--m: 3px;\n\t}\n}\n"},
		{"nil pointer", "settings", (*settings)(nil), ""},
		{"wrong type", "settings", 42, ""},
	})
	checkCSSFor(t, nested.CSSFor, []cssForCase{
		{"nested", "settings", &settings{Accent: "#0A84FF"},
			".card {\n\t--c: #0A84FF;\n\t& .title {\n\t\t--t: 3px;\n\t}\n}\n"},
	})
	if got, want := r.CSS(), ".accent {\n\t--x: 1px;\n}\n\n.accent:hover {\n\t--h: 2px;\n}\n\n@media screen {\n\t.accent {\n\t\t--m: 3px;\n\t}\n}\n"; got != want {
		t.Errorf("CSS() = %q, want %q: computing must not change the declared rule", got, want)
	}
}

func TestComputeNilMeansAbsent(t *testing.T) {
	cases := []struct {
		name  string
		props func(*settings) []css.Property
	}{
		{"nil", func(*settings) []css.Property { return nil }},
		{"empty", func(*settings) []css.Property { return []css.Property{} }},
		{"only nil properties", func(*settings) []css.Property { return []css.Property{nil, nil} }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := Class("a").Set(css.Custom("--x", css.Px(1))).Bind(Compute("settings", c.props))
			if got := r.CSSFor("settings", &settings{}); got != "" {
				t.Errorf("CSSFor() = %q, want empty", got)
			}
		})
	}

	r := Class("a").Bind(Compute("settings", func(*settings) []css.Property {
		return []css.Property{nil, css.Custom("--y", css.Px(2))}
	}))
	if got, want := r.CSSFor("settings", &settings{}), ".a {\n\t--y: 2px;\n}\n"; got != want {
		t.Errorf("CSSFor() = %q, want %q", got, want)
	}
}

func TestCSSForUsesTheLastBindingOfTheAction(t *testing.T) {
	yes := func(app.Action) bool { return true }
	no := func(app.Action) bool { return false }
	base := func(name string) *Rule { return Class(name).Set(css.Custom("--n", css.Px(1))) }

	cases := []struct {
		name string
		rule *Rule
		want string
	}{
		{"On true", base("a").On("x", yes), baseBlock("a") + "\n"},
		{"On false", base("a").On("x", no), ""},
		{"later On wins", base("a").On("x", no).On("x", yes), baseBlock("a") + "\n"},
		{"later When wins", base("a").On("x", yes).Bind(When("x", compact)), ""},
		{"later Compute wins", base("a").Bind(When("x", func(*settings) bool { return true })).Bind(Compute("x", accent)), ".a {\n\t--c: #000;\n}\n"},
		{"other action ignored", base("a").On("x", yes).On("y", no), baseBlock("a") + "\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.rule.CSSFor("x", &settings{Accent: "#000"}); got != c.want {
				t.Errorf("CSSFor() = %q, want %q", got, c.want)
			}
		})
	}

	mixed := base("a").Bind(When("x", compact)).On("x", yes)
	if got, want := mixed.CSSFor("x", "wrong"), baseBlock("a")+"\n"; got != want {
		t.Errorf("CSSFor() with a value only an earlier binding cannot read = %q, want %q", got, want)
	}
	typedLast := base("a").On("x", yes).Bind(When("x", compact))
	if got := typedLast.CSSFor("x", "wrong"); got != "" {
		t.Errorf("CSSFor() with a value the last binding cannot read = %q, want empty", got)
	}
}

func TestBindRejectsEmptyActionAndNilFunction(t *testing.T) {
	r := func() *Rule { return Class("a").Set(css.Custom("--x", css.Px(1))) }
	k := func() *KeyframesRule { return Keyframes("k").From(css.Opacity(css.Num(1))) }
	checkErrors(t, []errorCase{
		{"When", r().Bind(When("x", compact)).Err(), nil},
		{"Compute", r().Bind(Compute("x", accent)).Err(), nil},
		{"ComputeFrames", k().Bind(ComputeFrames("x", fading)).Err(), nil},
		{"keyframes When", k().Bind(When("x", compact)).Err(), nil},
		{"When without action", r().Bind(When("", compact)).Err(), ErrBadBinding},
		{"When without function", r().Bind(When[settings]("x", nil)).Err(), ErrBadBinding},
		{"Compute without action", r().Bind(Compute("", accent)).Err(), ErrBadBinding},
		{"Compute without function", r().Bind(Compute[settings]("x", nil)).Err(), ErrBadBinding},
		{"ComputeFrames without action", k().Bind(ComputeFrames("", fading)).Err(), ErrBadBinding},
		{"ComputeFrames without function", k().Bind(ComputeFrames[settings]("x", nil)).Err(), ErrBadBinding},
		{"zero Binding on rule", r().Bind(Binding{}).Err(), ErrBadBinding},
		{"zero Binding on keyframes", k().Bind(Binding{}).Err(), ErrBadBinding},
		{"keyframes When without function", k().Bind(When[settings]("x", nil)).Err(), ErrBadBinding},
	})

	bad := r().Bind(When("", compact))
	if bad.dynamic() {
		t.Error("a rejected binding was recorded")
	}
}

func TestComputeFramesOnKeyframes(t *testing.T) {
	k := Keyframes("pulse").Bind(ComputeFrames("speed", fading))
	if err := k.Err(); err != nil {
		t.Fatalf("Err() = %v", err)
	}
	checkCSSFor(t, k.CSSFor, []cssForCase{
		{"computed", "speed", &settings{}, "@keyframes pulse {\n\tfrom {\n\t\topacity: 1;\n\t}\n\tto {\n\t\topacity: 0.4;\n\t}\n}\n"},
		{"nil", "speed", &settings{Compact: true}, ""},
		{"wrong type", "speed", true, ""},
		{"unknown action", "other", &settings{}, ""},
	})

	sparse := Keyframes("sparse").Bind(ComputeFrames("speed", func(*settings) []Frame {
		return []Frame{{}, FrameAt(css.Pct(50), nil), FrameAt(css.Pct(25), nil, css.Opacity(css.Num(0.5)))}
	}))
	if got, want := sparse.CSSFor("speed", &settings{}), "@keyframes sparse {\n\t25% {\n\t\topacity: 0.5;\n\t}\n}\n"; got != want {
		t.Errorf("CSSFor() = %q, want %q", got, want)
	}

	outside := Keyframes("outside").Bind(ComputeFrames("speed", func(*settings) []Frame {
		return []Frame{FrameAt(css.Pct(-5), css.Opacity(css.Num(0))), FrameAt(css.Pct(50), css.Opacity(css.Num(0.5))), FrameAt(css.Pct(120), css.Opacity(css.Num(1)))}
	}))
	if err := outside.Err(); err != nil {
		t.Errorf("Err() with computed stops outside 0%%-100%% = %v, want nil", err)
	}
	if got, want := outside.CSSFor("speed", &settings{}), "@keyframes outside {\n\t50% {\n\t\topacity: 0.5;\n\t}\n}\n"; got != want {
		t.Errorf("CSSFor() = %q, want %q", got, want)
	}
	if got := FrameAt(css.Pct(101), css.Opacity(css.Num(1))); len(got.props) != 0 || got.stop != "" {
		t.Errorf("FrameAt(101%%) = %+v, want an empty frame", got)
	}

	declared := Keyframes("pulse").From(css.Opacity(css.Num(0))).Bind(When("speed", compact))
	checkCSSFor(t, declared.CSSFor, []cssForCase{
		{"When true", "speed", &settings{Compact: true}, "@keyframes pulse {\n\tfrom {\n\t\topacity: 0;\n\t}\n}\n"},
		{"When false", "speed", &settings{}, ""},
	})
}

func TestMismatchedBindingKind(t *testing.T) {
	checkErrors(t, []errorCase{
		{"ComputeFrames on a rule", Class("a").Bind(ComputeFrames("x", fading)).Err(), ErrBadBinding},
		{"Compute on keyframes", Keyframes("a").Bind(Compute("x", accent)).Err(), ErrBadBinding},
	})
}

func TestFramesMatchTheBuilder(t *testing.T) {
	built := Keyframes("k").At(css.Pct(10), css.Opacity(css.Num(0.1))).From(css.Opacity(css.Num(0))).To(css.Opacity(css.Num(1)))
	computed := Keyframes("k").Bind(ComputeFrames("x", func(*settings) []Frame {
		return []Frame{FrameAt(css.Pct(10), css.Opacity(css.Num(0.1))), FrameFrom(css.Opacity(css.Num(0))), FrameTo(css.Opacity(css.Num(1)))}
	}))
	if got, want := computed.CSSFor("x", &settings{}), built.CSS(); got != want {
		t.Errorf("CSSFor() = %q, want the builder's %q", got, want)
	}
}
