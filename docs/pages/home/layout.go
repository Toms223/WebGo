package home

import (
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/markdown"
	"github.com/Toms223/WebGo/docs/pages/start"
	"github.com/Toms223/WebGo/docs/pages/state"
	"github.com/Toms223/WebGo/docs/pages/styling"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const install = "go get github.com/Toms223/WebGo@v0.0.5"

type Layout struct{}

func (l *Layout) Loading() app.UI {
	return app.Div().Class(theme.Wrap.Name(), theme.ShellLoading.Name()).Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Class(theme.Wrap.Name(), theme.ShellError.Name()).Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	return l
}

func (l *Layout) Draw(m *data.ViewModel[State]) app.HTMLDiv {
	return app.Div().Class(theme.Wrap.Name()).Body(
		hero(),
		features(),
		band(),
		feel(),
		steps(),
		closing(),
		footer(),
	)
}

func hero() app.UI {
	return app.Section().Class(theme.Hero.Name()).Body(
		app.Div().Class(theme.HeroCopy.Name()).Body(
			app.Span().Class(theme.Chip.Name()).Text("v0.0.5 · built on go-app"),
			app.H1().Class(theme.HeroTitle.Name()).Body(
				app.Text("Build the "),
				app.Span().Class(theme.Highlight.Name()).Text("whole"),
				app.Text(" web app in Go."),
			),
			app.P().Class(theme.Lead.Name()).Text("WebGo adds pages, view models, markdown and typed CSS to go-app. Routing, state and stylesheets are plain Go that compiles to WebAssembly, so the compiler checks your UI the way it checks your server."),
			app.Div().Class(theme.Actions.Name()).Body(
				app.A().Class(theme.Button.Name(), theme.ButtonPrimary.Name()).Href(start.Route).Body(app.Text("Get started"), arrow()),
				app.A().Class(theme.Button.Name()).Href(base.Repository).Text("Star on GitHub"),
			),
			app.Div().Class(theme.Code.Name(), theme.CodeQuiet.Name()).Body(
				app.Pre().Body(dim("$ "), app.Text(install)),
			),
		),
		app.Div().Class(theme.HeroSample.Name()).Body(
			sample("theme.go", theme.CodeTilt.Name(),
				kw("var"), tx(" Button = style."), fn("Class"), tx("("), str(`"button"`), tx(").\n"),
				tx("\t"), fn("Set"), tx("(\n"),
				tx("\t\tcss."), fn("Padding"), tx("(css."), fn("Rem"), tx("(0.6), css."), fn("Rem"), tx("(1.1)),\n"),
				tx("\t\tcss."), fn("Border"), tx("(css."), fn("Px"), tx("(2), kw.Solid, css."), fn("Var"), tx("("), str(`"line"`), tx(")),\n"),
				tx("\t\tcss."), fn("BoxShadow"), tx("(css."), fn("Var"), tx("("), str(`"shadow"`), tx(")),\n"),
				tx("\t).\n"),
				tx("\t"), fn("Hover"), tx("(css."), fn("Transform"), tx("(css."), fn("Raw"), tx("("), str(`"translate(-2px, -2px)"`), tx("))).\n"),
				tx("\t"), fn("Media"), tx("(media."), fn("MaxWidth"), tx("(css."), fn("Rem"), tx("(48)), css."), fn("Width"), tx("(css."), fn("Pct"), tx("(100)))\n\n"),
				kw("func"), tx(" (b *Buy) "), fn("Render"), tx("() app.UI {\n"),
				tx("\t"), kw("return"), tx(" app."), fn("Button"), tx("()."), fn("Class"), tx("(Button."), fn("Name"), tx("())."), fn("Text"), tx("("), str(`"Buy"`), tx(")\n"),
				tx("}"),
			),
		),
	)
}

func features() app.UI {
	return app.Section().Class(theme.Section.Name()).Body(
		app.Span().Class(theme.Eyebrow.Name()).Text("What you get"),
		app.Div().Class(theme.Tiles.Name()).Body(
			card(state.Route, theme.FillSun, iconPages(), "Pages and routing", "A Base shell hosts Screens. Routes, base paths and fragments resolve in the browser, with no server round trip."),
			card(state.Route, theme.FillLime, iconState(), "View models", "Typed state with Load, Apply and Set. Observe it from any component, or react to it from a stylesheet."),
			card(styling.Route, theme.FillLilac, iconBrush(), "Typed CSS", "Properties, keywords, media and container queries, keyframes, nesting and layers as Go values. Dead rules never ship."),
			card(markdown.Route, theme.FillCoral, iconDoc(), "Markdown with live demos", "Render docs from markdown and drop a running Go component into any fence. This site is built that way."),
		),
	)
}

func band() app.UI {
	return app.Div().Class(theme.Band.Name()).Body(
		app.Ul().Class(theme.Marquee.Name()).Body(
			app.Li().Text("Go all the way down"),
			separator(),
			app.Li().Text("Compiles to WebAssembly"),
			separator(),
			app.Li().Text("Static site in one command"),
			separator(),
			app.Li().Text("Typed stylesheets"),
			separator(),
			app.Li().Text("Zero JavaScript to write"),
		),
	)
}

func feel() app.UI {
	return app.Section().Class(theme.Two.Name()).Body(
		app.Div().Class(theme.TwoText.Name()).Body(
			app.Span().Class(theme.Sticker.Name()).Text("How it feels"),
			app.H2().Class(theme.SectionTitle.Name()).Text("One language, one mental model."),
			app.P().Class(theme.Lead.Name()).Text("A screen is a Page with a Layout and a ViewModel. The layout draws from the model, handlers apply changes, and the page re-renders. Styles are values you can pass around, test and reuse."),
			app.Ul().Class(theme.Points.Name()).Body(
				app.Li().Text("Rename a CSS property and the build breaks, not the page."),
				app.Li().Text("Bind a class to state and the dynamic sheet recomputes on its own."),
				app.Li().Text("Ship the generated site to any static host, GitHub Pages included."),
			),
		),
		app.Div().Class(theme.TwoCode.Name()).Body(
			sample("counter.go", "",
				kw("type"), tx(" "), typ("Counter"), tx(" "), kw("struct"), tx("{ Count "), typ("int"), tx(" }\n\n"),
				kw("var"), tx(" model, _ = data."), fn("NewViewModel"), tx("("), str(`"counter"`), tx(", &"), typ("Counter"), tx("{})\n\n"),
				kw("func"), tx(" (l *Layout) "), fn("Draw"), tx("(m *data.ViewModel["), typ("Counter"), tx("]) app.HTMLDiv {\n"),
				tx("\t"), kw("return"), tx(" app."), fn("Div"), tx("()."), fn("Body"), tx("(\n"),
				tx("\t\tapp."), fn("H1"), tx("()."), fn("Textf"), tx("("), str(`"%d"`), tx(", m."), fn("State"), tx("().Count),\n"),
				tx("\t\tapp."), fn("Button"), tx("()."), fn("Text"), tx("("), str(`"+1"`), tx(")."), fn("OnClick"), tx("(l.bump),\n"),
				tx("\t)\n}\n\n"),
				kw("func"), tx(" (l *Layout) "), fn("bump"), tx("(ctx app.Context, _ app.Event) {\n"),
				tx("\tmodel."), fn("Load"), tx("(ctx)."), fn("Apply"), tx("("), kw("func"), tx("(_ app.Context, c *"), typ("Counter"), tx(") *"), typ("Counter"), tx(" {\n"),
				tx("\t\tc.Count++\n"),
				tx("\t\t"), kw("return"), tx(" c\n"),
				tx("\t})."), fn("Set"), tx("()\n}"),
			),
		),
	)
}

func steps() app.UI {
	return app.Section().Class(theme.Section.Name()).Body(
		app.Span().Class(theme.Eyebrow.Name()).Text("Three steps"),
		app.H2().Class(theme.SectionTitle.Name()).Text("From go get to a running page"),
		app.Div().Class(theme.Steps.Name()).Body(
			step("1", theme.FillSun, "Install", "Add WebGo to your module. go-app comes with it.", install),
			step("2", theme.FillLime, "Write a page", "A component, a route and app.RunWhenOnBrowser. Then grow into Pages and ViewModels.", `app.Route("/", func() app.Composer { return &hello{} })`),
			step("3", theme.FillLilac, "Build and ship", "Compile to wasm and generate a static site, or serve it from the same binary.", "GOARCH=wasm GOOS=js go build -o web/app.wasm ."),
		),
	)
}

func closing() app.UI {
	return app.Section().Class(theme.Closing.Name()).Body(
		app.Div().Class(theme.ClosingText.Name()).Body(
			app.H2().Class(theme.SectionTitle.Name()).Text("Ready when you are."),
			app.P().Class(theme.ClosingLead.Name()).Text("The getting started guide takes you from an empty module to a static site. Every example on this site is live."),
		),
		app.A().Class(theme.Button.Name()).Href(start.Route).Body(app.Text("Open the docs"), arrow()),
	)
}

func footer() app.UI {
	return app.Footer().Class(theme.Footer.Name()).Body(
		app.Span().Text("Built with WebGo and go-app. Docs rendered from markdown, styled with typed CSS."),
		app.A().Class(theme.Pill.Name()).Href(base.Repository).Text("github.com/Toms223/WebGo"),
	)
}

func card(href string, fill *style.Rule, icon app.UI, title, text string) app.UI {
	return app.A().Class(theme.Card.Name()).Href(href).Body(
		app.Span().Class(theme.CardIcon.Name(), fill.Name()).Body(icon),
		app.H3().Class(theme.CardTitle.Name()).Text(title),
		app.P().Class(theme.CardText.Name()).Text(text),
	)
}

func step(number string, fill *style.Rule, title, text, code string) app.UI {
	return app.Div().Class(theme.Step.Name()).Body(
		app.Span().Class(theme.StepNumber.Name(), fill.Name()).Text(number),
		app.H3().Class(theme.StepTitle.Name()).Text(title),
		app.P().Class(theme.StepText.Name()).Text(text),
		app.Div().Class(theme.Code.Name(), theme.CodeQuiet.Name()).Body(app.Pre().Text(code)),
	)
}

func sample(name, extra string, body ...app.UI) app.UI {
	classes := []string{theme.Code.Name()}
	if extra != "" {
		classes = append(classes, extra)
	}
	return app.Div().Class(classes...).Body(
		app.Div().Class(theme.CodeBar.Name()).Text(name),
		app.Pre().Body(body...),
	)
}

func separator() app.UI {
	return app.Li().Class(theme.MarqueeDot.Name()).Aria("hidden", "true").Text("·")
}

func kw(text string) app.UI { return app.Span().Class(theme.Keyword.Name()).Text(text) }

func fn(text string) app.UI { return app.Span().Class(theme.Fn.Name()).Text(text) }

func str(text string) app.UI { return app.Span().Class(theme.Str.Name()).Text(text) }

func typ(text string) app.UI { return app.Span().Class(theme.Typ.Name()).Text(text) }

func dim(text string) app.UI { return app.Span().Class(theme.Dim.Name()).Text(text) }

func tx(text string) app.UI { return app.Text(text) }

func arrow() app.UI {
	return app.Raw(`<svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M13 6l6 6-6 6"/></svg>`)
}

func iconPages() app.UI {
	return app.Raw(`<svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 5h16v14H4z"/><path d="M4 10h16M9 10v9"/></svg>`)
}

func iconState() app.UI {
	return app.Raw(`<svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3v18M3 12h18"/><circle cx="12" cy="12" r="9"/></svg>`)
}

func iconBrush() app.UI {
	return app.Raw(`<svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M4 20l4-1 10-10-3-3L5 16l-1 4z"/><path d="M13 7l3 3"/></svg>`)
}

func iconDoc() app.UI {
	return app.Raw(`<svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 4h14v16H5z"/><path d="M8 9h8M8 13h8M8 17h5"/></svg>`)
}
