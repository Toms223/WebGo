package home

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/markdown"
	"github.com/Toms223/WebGo/docs/pages/start"
	"github.com/Toms223/WebGo/docs/pages/state"
	"github.com/Toms223/WebGo/docs/pages/styling"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/Toms223/WebGo/style"
)

const install = "go get github.com/Toms223/WebGo@v0.1.0"

type Layout struct{}

func (l *Layout) Loading() app.UI {
	return app.Div().Style(theme.Wrap, theme.ShellLoading).Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Style(theme.Wrap, theme.ShellError).Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) page.Layout[State] {
	return l
}

func (l *Layout) Draw(m *data.ViewModel[State]) app.HTMLDiv {
	return app.Div().Style(theme.Wrap).Body(
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
	return app.Section().Style(theme.Hero).Body(
		app.Div().Style(theme.HeroCopy).Body(
			app.Span().Style(theme.Chip).Text("v0.1.0 · built from go-app"),
			app.H1().Style(theme.HeroTitle).Body(
				app.Text("Build the "),
				app.Span().Style(theme.Highlight).Text("whole"),
				app.Text(" web app in Go."),
			),
			app.P().Style(theme.Lead).Text("WebGo adds pages, view models, markdown and typed CSS to go-app. Routing, state and stylesheets are plain Go that compiles to WebAssembly, so the compiler checks your UI the way it checks your server."),
			app.Div().Style(theme.Actions).Body(
				app.A().Style(theme.Button, theme.ButtonPrimary).Href(nav.Href(start.Route)).Body(app.Text("Get started"), arrow()),
				app.A().Style(theme.Button).Href(base.Repository).Text("Star on GitHub"),
			),
			app.Div().Style(theme.Code, theme.CodeQuiet).Body(
				app.Pre().Body(dim("$ "), app.Text(install)),
			),
		),
		app.Div().Style(theme.HeroSample).Body(
			sample("theme.go", theme.CodeTilt,
				kw("var"), tx(" Button = style."), fn("Class"), tx("("), str(`"button"`), tx(").\n"),
				tx("\t"), fn("Set"), tx("(\n"),
				tx("\t\tcss."), fn("Padding"), tx("(css."), fn("Rem"), tx("(0.6), css."), fn("Rem"), tx("(1.1)),\n"),
				tx("\t\tcss."), fn("Border"), tx("(css."), fn("Px"), tx("(2), kw.Solid, css."), fn("Var"), tx("("), str(`"line"`), tx(")),\n"),
				tx("\t\tcss."), fn("BoxShadow"), tx("(css."), fn("Var"), tx("("), str(`"shadow"`), tx(")),\n"),
				tx("\t).\n"),
				tx("\t"), fn("Hover"), tx("(css."), fn("Transform"), tx("(css."), fn("Raw"), tx("("), str(`"translate(-2px, -2px)"`), tx("))).\n"),
				tx("\t"), fn("Media"), tx("(media."), fn("MaxWidth"), tx("(css."), fn("Rem"), tx("(48)), css."), fn("Width"), tx("(css."), fn("Pct"), tx("(100)))\n\n"),
				kw("func"), tx(" (b *Buy) "), fn("Render"), tx("() app.UI {\n"),
				tx("\t"), kw("return"), tx(" app."), fn("Button"), tx("()."), fn("Style"), tx("(Button)."), fn("Text"), tx("("), str(`"Buy"`), tx(")\n"),
				tx("}"),
			),
		),
	)
}

func features() app.UI {
	return app.Section().Style(theme.Section).Body(
		app.Span().Style(theme.Eyebrow).Text("What you get"),
		app.Div().Style(theme.Tiles).Body(
			card(nav.Href(state.Route), theme.FillSun, iconPages(), "Pages and routing", "A Base shell hosts Screens. Routes, base paths and fragments resolve in the browser, with no server round trip."),
			card(nav.Href(state.Route), theme.FillLime, iconState(), "View models", "Typed state with Load, Apply and Set. Observe it from any component, or react to it from a stylesheet."),
			card(nav.Href(styling.Route), theme.FillLilac, iconBrush(), "Typed CSS", "Properties, keywords, media and container queries, keyframes, nesting and layers as Go values. Dead rules never ship."),
			card(nav.Href(markdown.Route), theme.FillCoral, iconDoc(), "Markdown with live demos", "Render docs from markdown and drop a running Go component into any fence. Every example on this site runs live."),
		),
	)
}

func band() app.UI {
	return app.Div().Style(theme.Band).Body(
		app.Ul().Style(theme.Marquee).Body(
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
	return app.Section().Style(theme.Two).Body(
		app.Div().Style(theme.TwoText).Body(
			app.Span().Style(theme.Sticker).Text("How it feels"),
			app.H2().Style(theme.SectionTitle).Text("One language, one mental model."),
			app.P().Style(theme.Lead).Text("A screen is a Page with a Layout and a ViewModel. The layout draws from the model, handlers apply changes, and the page re-renders. Styles are values you can pass around, test and reuse."),
			app.Ul().Style(theme.Points).Body(
				app.Li().Text("Rename a CSS property and the build breaks, not the page."),
				app.Li().Text("Bind a class to state and the dynamic sheet recomputes on its own."),
				app.Li().Text("Ship the generated site to any static host, GitHub Pages included."),
			),
		),
		app.Div().Style(theme.TwoCode).Body(
			sample("counter.go", nil,
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
	return app.Section().Style(theme.Section).Body(
		app.Span().Style(theme.Eyebrow).Text("Three steps"),
		app.H2().Style(theme.SectionTitle).Text("From go get to a running page"),
		app.Div().Style(theme.Steps).Body(
			step("1", theme.FillSun, "Install", "Add WebGo to your module. go-app comes with it.", install),
			step("2", theme.FillLime, "Write a page", "A component, a route and app.RunWhenOnBrowser. Then grow into Pages and ViewModels.", `app.Route("/", func() app.Composer { return &hello{} })`),
			step("3", theme.FillLilac, "Build and ship", "Compile to wasm and generate a static site, or serve it from the same binary.", "GOARCH=wasm GOOS=js go build -o web/app.wasm ."),
		),
	)
}

func closing() app.UI {
	return app.Section().Style(theme.Closing).Body(
		app.Div().Style(theme.ClosingText).Body(
			app.H2().Style(theme.SectionTitle).Text("Ready when you are."),
			app.P().Style(theme.ClosingLead).Text("The getting started guide takes you from an empty module to a static site. Every example on this site is live."),
		),
		app.A().Style(theme.Button).Href(nav.Href(start.Route)).Body(app.Text("Open the docs"), arrow()),
	)
}

func footer() app.UI {
	return app.Footer().Style(theme.Footer).Body(
		app.Span().Text("Built with WebGo and go-app. Docs written in Go, styled with typed CSS."),
		app.A().Style(theme.Pill).Href(base.Repository).Text("github.com/Toms223/WebGo"),
	)
}

func card(href string, fill *style.Rule, icon app.UI, title, text string) app.UI {
	return app.A().Style(theme.Card).Href(href).Body(
		app.Span().Style(theme.CardIcon, fill).Body(icon),
		app.H2().Style(theme.CardTitle).Text(title),
		app.P().Style(theme.CardText).Text(text),
	)
}

func step(number string, fill *style.Rule, title, text, code string) app.UI {
	return app.Div().Style(theme.Step).Body(
		app.Span().Style(theme.StepNumber, fill).Text(number),
		app.H3().Style(theme.StepTitle).Text(title),
		app.P().Style(theme.StepText).Text(text),
		app.Div().Style(theme.Code, theme.CodeQuiet).Body(app.Pre().Text(code)),
	)
}

func sample(name string, extra app.ClassRule, body ...app.UI) app.UI {
	classes := []app.ClassRule{theme.Code}
	if extra != nil {
		classes = append(classes, extra)
	}
	return app.Div().Style(classes...).Body(
		app.Div().Style(theme.CodeBar).Text(name),
		app.Pre().Body(body...),
	)
}

func separator() app.UI {
	return app.Li().Style(theme.MarqueeDot).Aria("hidden", "true").Text("·")
}

func kw(text string) app.UI { return app.Span().Style(theme.Keyword).Text(text) }

func fn(text string) app.UI { return app.Span().Style(theme.Fn).Text(text) }

func str(text string) app.UI { return app.Span().Style(theme.Str).Text(text) }

func typ(text string) app.UI { return app.Span().Style(theme.Typ).Text(text) }

func dim(text string) app.UI { return app.Span().Style(theme.Dim).Text(text) }

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
