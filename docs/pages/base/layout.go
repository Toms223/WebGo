package base

import (
	"strconv"

	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	HomeRoute  = "/"
	Repository = "https://github.com/Toms223/WebGo"
)

type Layout struct {
	entries []nav.Entry
	body    []app.UI
}

func NewLayout(entries []nav.Entry) *Layout {
	return &Layout{entries: entries}
}

func (l *Layout) Loading() app.UI {
	return app.Div().Class(theme.Shell.Name(), theme.ShellLoading.Name()).Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Class(theme.Shell.Name(), theme.ShellError.Name()).Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	l.body = content
	return l
}

func (l *Layout) Draw(m *data.ViewModel[State]) app.HTMLDiv {
	state := m.State()
	if state.Active == HomeRoute {
		return l.landing(state)
	}
	classes := []string{theme.Shell.Name()}
	if state.Menu {
		classes = append(classes, theme.MenuOpen.Name())
	}
	return app.Div().Class(classes...).Body(
		app.Aside().Class(theme.Side.Name()).Body(
			logo(),
			app.Button().Type("button").Class(theme.SideMenu.Name()).
				Aria("expanded", strconv.FormatBool(state.Menu)).
				Aria("controls", nav.ID).
				Text("Menu").OnClick(l.toggleMenu),
			nav.Render(l.entries, state.Active),
			l.toggleButton(state),
		),
		app.Main().Class(theme.Content.Name()).Body(l.body...),
	)
}

func (l *Layout) landing(state *State) app.HTMLDiv {
	links := []app.UI{}
	if len(l.entries) > 0 {
		links = append(links, app.A().Class(theme.Pill.Name()).Href(l.entries[0].Route).Text("Docs"))
	}
	links = append(links,
		app.A().Class(theme.Pill.Name()).Href(Repository).Text("GitHub"),
		l.toggleButton(state),
	)
	return app.Div().Class(theme.Landing.Name()).Body(
		app.Header().Class(theme.Topbar.Name()).Body(
			logo(),
			app.Nav().Class(theme.TopbarNav.Name()).Body(links...),
		),
		app.Main().Body(l.body...),
	)
}

func (l *Layout) toggleButton(state *State) app.UI {
	label := "Dark mode"
	if state.IsDark() {
		label = "Light mode"
	}
	return app.Button().Type("button").Class(theme.ShellToggle.Name()).Text(label).OnClick(l.toggle)
}

func logo() app.UI {
	return app.A().Class(theme.Logo.Name()).Href(HomeRoute).Body(
		app.Span().Class(theme.LogoMark.Name()).Text("W"),
		app.Text("WebGo"),
	)
}

func (l *Layout) toggle(ctx app.Context, e app.Event) {
	next := !shell.Get(ctx).State().IsDark()
	shell.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Dark = next
		state.Chosen = true
		return state
	}).Set()
	ctx.NewActionWithValue(theme.DarkAction, next)
}

func (l *Layout) toggleMenu(ctx app.Context, e app.Event) {
	shell.Load(ctx).Apply(func(ctx app.Context, state *State) *State {
		state.Menu = !state.Menu
		return state
	}).Set()
}
