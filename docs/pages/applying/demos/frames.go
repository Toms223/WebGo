package demos

import (
	"strconv"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const SpeedKey = "demo/speed"

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
	)

var speed = model(SpeedKey, startSpeed())

func startSpeed() *Speed { return &Speed{Beats: 1} }

type framesDemo struct{ app.Compo }

func FramesDemo() app.UI { return &framesDemo{} }

func (d *framesDemo) OnMount(ctx app.Context) {
	speed.OnAction(ctx)
	speed.Load(ctx).Set()
}

func (d *framesDemo) Render() app.UI {
	return app.Div().Body(
		app.Label().Body(
			app.Text("Beats per cycle "),
			app.Input().Type("range").Min(1).Max(5).Value(speed.State().Beats).OnInput(d.change),
		),
		app.Div().Class(DemoPulseBox.Name()),
	)
}

func (d *framesDemo) change(ctx app.Context, e app.Event) {
	beats, err := strconv.Atoi(ctx.JSSrc().Get("value").String())
	if err != nil {
		return
	}
	speed.Load(ctx).Apply(func(ctx app.Context, s *Speed) *Speed {
		s.Beats = beats
		return s
	}).Set()
}
