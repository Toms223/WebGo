package demos

import (
	"strconv"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var VotesBox = style.Class("demo-actions-votes").Set(
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.5), css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.Background(css.Var("--mint")),
	css.FontWeight(css.Num(700)),
)

const VotesKey = "demo/actions/votes"

type votesDemo struct{ app.Compo }

func VotesDemo() app.UI { return &votesDemo{} }

func (d *votesDemo) Render() app.UI {
	return app.Div().Style(VotesBox).Body(&voter{}, &tally{})
}

type voter struct{ app.Compo }

func (v *voter) Render() app.UI {
	return app.Span().Body(
		app.Button().Type("button").Text("Vote").OnClick(v.vote),
		app.Button().Type("button").Text("Reset").OnClick(v.reset),
	)
}

func (v *voter) vote(ctx app.Context, e app.Event) {
	var votes int
	ctx.GetState(VotesKey, &votes)
	ctx.SetState(VotesKey, votes+1)
}

func (v *voter) reset(ctx app.Context, e app.Event) {
	ctx.SetState(VotesKey, 0)
}

type tally struct {
	app.Compo
	votes int
}

func (t *tally) OnMount(ctx app.Context) {
	ctx.ObserveState(VotesKey, &t.votes)
}

func (t *tally) Render() app.UI {
	return app.Span().Body(
		app.Text("Votes "),
		app.Output().Text(strconv.Itoa(t.votes)),
	)
}

func Items() []style.Item {
	return []style.Item{VotesBox}
}

const votesSource = `const VotesKey = "demo/actions/votes"

type votesDemo struct{ app.Compo }

func VotesDemo() app.UI { return &votesDemo{} }

func (d *votesDemo) Render() app.UI {
	return app.Div().Style(VotesBox).Body(&voter{}, &tally{})
}

type voter struct{ app.Compo }

func (v *voter) Render() app.UI {
	return app.Span().Body(
		app.Button().Type("button").Text("Vote").OnClick(v.vote),
		app.Button().Type("button").Text("Reset").OnClick(v.reset),
	)
}

func (v *voter) vote(ctx app.Context, e app.Event) {
	var votes int
	ctx.GetState(VotesKey, &votes)
	ctx.SetState(VotesKey, votes+1)
}

func (v *voter) reset(ctx app.Context, e app.Event) {
	ctx.SetState(VotesKey, 0)
}

type tally struct {
	app.Compo
	votes int
}

func (t *tally) OnMount(ctx app.Context) {
	ctx.ObserveState(VotesKey, &t.votes)
}

func (t *tally) Render() app.UI {
	return app.Span().Body(
		app.Text("Votes "),
		app.Output().Text(strconv.Itoa(t.votes)),
	)
}`

var VotesExample = example.Example{Name: "votes.go", Source: votesSource, CSS: VotesBox.CSS(), Live: VotesDemo}

func Examples() []example.Example {
	return []example.Example{VotesExample}
}
