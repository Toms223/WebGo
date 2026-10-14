package demos

import (
	"strings"

	"github.com/Toms223/WebGo/docs/components/example"
)

const whenSource = `const CompactKey = "demo/compact"

type Compact struct{ On bool }

var DemoCompactBase = style.Class("demo-compact").
	Set(
		css.Padding(css.Rem(2)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--mint")),
	)

var DemoCompact = style.Class("demo-compact").
	Set(css.Padding(css.Rem(0.25))).
	Bind(style.When(CompactKey, func(c *Compact) bool { return c.On }))`

const computeSource = `const AccentKey = "demo/accent"

type Accent struct{ Color string }

var DemoAccent = style.Class("demo-accent").
	Bind(style.Compute(AccentKey, func(a *Accent) []css.Property {
		return []css.Property{
			css.Color(css.Hex(a.Color)),
			css.BorderBottom(css.Px(3), kw.Solid, css.Hex(a.Color)),
		}
	}))`

const framesSource = `const SpeedKey = "demo/speed"

type Speed struct{ Beats int }

var DemoPulse = style.Keyframes("demo-beat").
	Bind(style.ComputeFrames(SpeedKey, func(s *Speed) []style.Frame {
		var frames []style.Frame
		steps := 2 * s.Beats
		for step := 0; step <= steps; step++ {
			opacity := 1.0
			if step%2 == 1 {
				opacity = 0.2
			}
			at := css.Pct(float64(step) * 100 / float64(steps))
			frames = append(frames, style.FrameAt(at, css.Opacity(css.Num(opacity))))
		}
		return frames
	}))

var DemoPulseBox = style.Class("demo-beating").
	Set(
		css.Width(css.Rem(3)),
		css.Height(css.Rem(3)),
		css.Background(css.Var("--accent")),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.AnimationName(css.Ident(DemoPulse.Name())),
		css.AnimationDuration(css.S(2)),
		css.AnimationIterationCount(kw.Infinite),
		css.AnimationTimingFunction(kw.Linear),
	)`

const initiallySource = `const BannerKey = "demo/banner"

type Banner struct{ Hidden bool }

var DemoBannerHidden = style.Class("demo-banner").
	Set(css.Display(kw.None))

var DemoBanner = style.Class("demo-banner").
	Set(
		css.Display(kw.Block),
		css.Padding(css.Rem(0.75), css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--side")),
	).
	Bind(style.When(BannerKey, func(b *Banner) bool { return !b.Hidden })).
	Initially(true)`

const onSource = `const OutlineAction = "demo/outline"

var DemoOutlineBase = style.Class("demo-outlined").
	Set(
		css.Padding(css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	)

var DemoOutline = style.Class("demo-outlined").
	Set(
		css.Outline(css.Px(3), kw.Dashed, css.Var("--accent")),
		css.OutlineOffset(css.Px(4)),
	).
	On(OutlineAction, outlined)

func outlined(a app.Action) bool {
	on, ok := a.Value.(bool)
	return ok && on
}`

var WhenExample = example.Example{Name: "when.go", Source: whenSource, CSS: joined(DemoCompactBase.CSS(), DemoCompact.CSS()), Live: WhenDemo}

var ComputeExample = example.Example{Name: "compute.go", Source: computeSource, CSS: joined(DemoAccent.CSSFor(AccentKey, startAccent())), Live: ComputeDemo}

var FramesExample = example.Example{Name: "frames.go", Source: framesSource, CSS: joined(DemoPulse.CSSFor(SpeedKey, startSpeed()), DemoPulseBox.CSS()), Live: FramesDemo}

var InitiallyExample = example.Example{Name: "initially.go", Source: initiallySource, CSS: joined(DemoBannerHidden.CSS(), DemoBanner.CSS()), Live: InitiallyDemo}

var OnExample = example.Example{Name: "on.go", Source: onSource, CSS: joined(DemoOutlineBase.CSS(), DemoOutline.CSS()), Live: OnDemo}

func Examples() []example.Example {
	return []example.Example{WhenExample, ComputeExample, FramesExample, InitiallyExample, OnExample}
}

func joined(texts ...string) string {
	return strings.Join(texts, "\n")
}
