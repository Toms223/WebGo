package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

const BannerKey = "demo/banner"

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
	Initially(true)

var banner = model(BannerKey, &Banner{})

type initiallyDemo struct{ app.Compo }

func InitiallyDemo() app.UI { return &initiallyDemo{} }

func (d *initiallyDemo) OnMount(ctx app.Context) {
	banner.OnAction(ctx)
}

func (d *initiallyDemo) Render() app.UI {
	label := "Hide the banner"
	if banner.State().Hidden {
		label = "Show the banner"
	}
	return app.Div().Body(
		app.Button().Type("button").Text(label).OnClick(d.toggle),
		app.Div().Style(DemoBanner).Text("Present from the start"),
	)
}

func (d *initiallyDemo) toggle(ctx app.Context, e app.Event) {
	banner.Load(ctx).Apply(func(ctx app.Context, b *Banner) *Banner {
		b.Hidden = !b.Hidden
		return b
	}).Set()
}
