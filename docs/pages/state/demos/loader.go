package demos

import (
	"strconv"
	"time"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var LoaderBox = style.Class("demo-loader").Set(
	css.Display(kw.InlineFlex),
	css.Padding(css.Rem(0.5), css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.Background(css.Var("--mint")),
	css.FontWeight(css.Num(700)),
)

const ProfileKey = "demo/profile"

type Profile struct {
	Name  string
	Loads int
}

var profile = mustModel(ProfileKey, &Profile{})

type profileLayout struct{}

func (profileLayout) Loading() app.UI {
	return app.Div().Style(LoaderBox).Text("Loading the profile…")
}

func (profileLayout) Error(err error) app.UI {
	return app.Div().Style(LoaderBox).Text(err.Error())
}

func (profileLayout) Draw(m *app.Model[Profile]) app.HTMLDiv {
	p := m.State()
	return app.Div().Style(LoaderBox).Text(p.Name + ", loaded " + strconv.Itoa(p.Loads) + " times")
}

func (l profileLayout) Body(...app.UI) app.Layout[Profile] { return l }

func fetchProfile(ctx app.Context, p *Profile) error {
	time.Sleep(600 * time.Millisecond)
	p.Name = "Ada"
	p.Loads++
	return nil
}

type loaderDemo struct {
	app.Compo
	runs int
}

func LoaderDemo() app.UI { return &loaderDemo{} }

func (d *loaderDemo) Render() app.UI {
	return app.Div().Body(
		app.NewScreen[Profile]("/demo/profile/"+strconv.Itoa(d.runs)).
			WithLayout(profileLayout{}).
			WithModel(profile).
			WithLoader(fetchProfile),
		app.Button().Type("button").Text("Load again").OnClick(d.again),
	)
}

func (d *loaderDemo) again(ctx app.Context, e app.Event) {
	d.runs++
}

const loaderSource = `const ProfileKey = "demo/profile"

type Profile struct {
	Name  string
	Loads int
}

var profile = mustModel(ProfileKey, &Profile{})

type profileLayout struct{}

func (profileLayout) Loading() app.UI {
	return app.Div().Style(LoaderBox).Text("Loading the profile…")
}

func (profileLayout) Error(err error) app.UI {
	return app.Div().Style(LoaderBox).Text(err.Error())
}

func (profileLayout) Draw(m *app.Model[Profile]) app.HTMLDiv {
	p := m.State()
	return app.Div().Style(LoaderBox).Text(p.Name + ", loaded " + strconv.Itoa(p.Loads) + " times")
}

func (l profileLayout) Body(...app.UI) app.Layout[Profile] { return l }

func fetchProfile(ctx app.Context, p *Profile) error {
	time.Sleep(600 * time.Millisecond)
	p.Name = "Ada"
	p.Loads++
	return nil
}

type loaderDemo struct {
	app.Compo
	runs int
}

func LoaderDemo() app.UI { return &loaderDemo{} }

func (d *loaderDemo) Render() app.UI {
	return app.Div().Body(
		app.NewScreen[Profile]("/demo/profile/"+strconv.Itoa(d.runs)).
			WithLayout(profileLayout{}).
			WithModel(profile).
			WithLoader(fetchProfile),
		app.Button().Type("button").Text("Load again").OnClick(d.again),
	)
}

func (d *loaderDemo) again(ctx app.Context, e app.Event) {
	d.runs++
}`

var LoaderExample = example.Example{Name: "loader.go", Source: loaderSource, Live: LoaderDemo}
