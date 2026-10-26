package base

import (
	"strconv"
	"strings"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
)

const mainID = "main"

const (
	HomeRoute  = "/"
	NewsRoute  = "/news"
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
	return app.Div().Style(theme.Shell, theme.ShellLoading).Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Style(theme.Shell, theme.ShellError).Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	l.body = content
	return l
}

func (l *Layout) Draw(m *data.ViewModel[State]) app.HTMLDiv {
	state := m.State()
	if state.Active == HomeRoute || state.Active == NewsRoute || strings.HasPrefix(state.Active, NewsRoute+"/") {
		return l.landing(state)
	}
	classes := []app.ClassRule{theme.Shell}
	if state.Menu {
		classes = append(classes, theme.MenuOpen)
	}
	return app.Div().Style(classes...).Body(
		skip(),
		app.Aside().Style(theme.Side).Body(
			logo(),
			app.Button().Type("button").Style(theme.SideMenu).
				Aria("expanded", strconv.FormatBool(state.Menu)).
				Aria("controls", nav.ID).
				Text("Menu").OnClick(l.toggleMenu),
			nav.Render(l.entries, state.Active),
			l.toggleButton(state),
		),
		app.Main().Style(theme.Content).ID(mainID).TabIndex(-1).Body(l.body...),
	)
}

func (l *Layout) landing(state *State) app.HTMLDiv {
	links := []app.UI{}
	if len(l.entries) > 0 {
		links = append(links, app.A().Style(theme.Pill).Href(nav.Href(l.entries[0].Route)).Text("Docs"))
	}
	news := app.A().Style(theme.Pill).Href(nav.Href(NewsRoute)).Text("News")
	if state.Active != HomeRoute {
		news = news.Aria("current", "page")
	}
	links = append(links,
		news,
		app.A().Style(theme.Pill).Href(Repository).Text("GitHub"),
		l.toggleButton(state),
	)
	return app.Div().Style(theme.Landing).Body(
		skip(),
		app.Header().Style(theme.Topbar).Body(
			logo(),
			app.Nav().Style(theme.TopbarNav).Body(links...),
		),
		app.Main().ID(mainID).TabIndex(-1).Body(l.body...),
	)
}

func (l *Layout) toggleButton(state *State) app.UI {
	label := "Dark mode"
	if state.IsDark() {
		label = "Light mode"
	}
	return app.Button().Type("button").Style(theme.ShellToggle).Text(label).OnClick(l.toggle)
}

func skip() app.UI {
	return app.Button().Type("button").Style(theme.Skip).Text("Skip to content").OnClick(func(ctx app.Context, e app.Event) {
		if main := app.Window().GetElementByID(mainID); main.Truthy() {
			main.Call("focus")
		}
	})
}

func logo() app.UI {
	return app.A().Style(theme.Logo).Href(nav.Href(HomeRoute)).Body(
		app.Span().Style(theme.LogoMark).Text("W"),
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
