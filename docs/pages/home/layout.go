package home

import (
	"strconv"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/markdown"
	"github.com/Toms223/WebGo/docs/pages/news"
	"github.com/Toms223/WebGo/docs/pages/start"
	"github.com/Toms223/WebGo/docs/pages/state"
	"github.com/Toms223/WebGo/docs/pages/styling"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/style"
)

const install = "go get github.com/Toms223/WebGo@v1.0.0"

type Layout struct {
	model *app.Model[State]
}

func (l *Layout) Loading() app.UI {
	return app.Div().Style(theme.Wrap, theme.ShellLoading).Text("Loading")
}

func (l *Layout) Error(err error) app.UI {
	return app.Div().Style(theme.Wrap, theme.ShellError).Text(err.Error())
}

func (l *Layout) Body(content ...app.UI) app.Layout[State] {
	return l
}

func (l *Layout) Draw(m *app.Model[State]) app.HTMLDiv {
	l.model = m
	return app.Div().Style(theme.Wrap).Body(
		slug(),
		hero(),
		l.press(m.State().Count),
		features(),
		band(),
		feel(),
		steps(),
		latestNews(),
		closing(),
		footer(),
	)
}

func slug() app.UI {
	return app.Div().Style(theme.Slug).Body(
		app.Span().Body(app.Text("WebGo zine "), app.B().Text("No. 1.0.0")),
		app.Span().Style(theme.SlugWide).Text("Printed in three inks on app.wasm"),
		app.Span().Style(theme.SlugWide).Text("Free. Take one."),
		app.Span().Text("Nov 2026"),
	)
}

func mis(text string) app.UI {
	return app.Span().Style(theme.Mis).DataSet("ink", text).Text(text)
}

func heading(text string) app.UI {
	return app.H2().Style(theme.SectionTitle).Body(app.Span().Style(theme.Slab).Text(text))
}

func folio(number string, ink app.ClassRule) app.UI {
	return app.Span().Style(theme.Folio, ink).Aria("hidden", "true").Text(number)
}

func hero() app.UI {
	return app.Section().Style(theme.Hero).Body(
		app.Span().Style(theme.Reg).Aria("hidden", "true"),
		app.Span().Style(theme.Reg, theme.RegEnd).Aria("hidden", "true"),
		app.Div().Style(theme.HeroInk).Aria("hidden", "true"),
		app.Div().Style(theme.HeroDots).Aria("hidden", "true"),
		app.Div().Style(theme.HeroHead).Body(
			app.Span().Style(theme.Chip).Text("v1.0.0 · First Edition"),
			app.H1().Style(theme.HeroTitle).Body(mis("Build the whole web app in Go.")),
		),
		app.Div().Style(theme.HeroCopy).Body(
			app.P().Style(theme.Lead).Text("WebGo gives you pages, view models, markdown and typed CSS on its own app engine. Routing, state and stylesheets are plain Go that compiles to WebAssembly, so the compiler checks your UI the way it checks your server."),
			app.Div().Style(theme.Code, theme.CodeQuiet).Body(
				app.Pre().Body(str("$ "), app.Text(install)),
			),
			app.Div().Style(theme.Actions).Body(
				app.A().Style(theme.Button, theme.ButtonPrimary).Href(nav.Href(start.Route)).Body(app.Text("Get started"), arrow()),
				app.A().Style(theme.Button).Href(base.Repository).Text("Star on GitHub"),
				app.Span().Style(theme.HandNote).Text("one file ships: app.wasm"),
			),
		),
		app.Div().Style(theme.HeroSample).Body(
			sticker(),
			sample("counter.go", "gofmt, tabs", "fig. 1, a screen is a layout and a model", theme.CodeTilt, []app.UI{
				app.Span().Style(theme.Tape).Aria("hidden", "true"),
				app.Span().Style(theme.TapeEnd).Aria("hidden", "true"),
			},
				kw("type"), tx(" "), typ("Counter"), tx(" "), kw("struct"), tx("{ Count "), typ("int"), tx(" }\n\n"),
				kw("var"), tx(" model, _ = app."), fn("NewModel"), tx("("), str(`"counter"`), tx(", &"), typ("Counter"), tx("{})\n\n"),
				kw("func"), tx(" (l *Layout) "), fn("Draw"), tx("(m *app.Model["), typ("Counter"), tx("]) app.HTMLDiv {\n"),
				tx("\t"), kw("return"), tx(" app.Div().Body(\n"),
				tx("\t\tapp.H1().Textf("), str(`"%d"`), tx(", m.State().Count),\n"),
				tx("\t\tapp.Button().Text("), str(`"+1"`), tx(").OnClick(l.bump),\n"),
				tx("\t)\n}\n\n"),
				kw("func"), tx(" (l *Layout) "), fn("bump"), tx("(ctx app.Context, _ app.Event) {\n"),
				tx("\tmodel.Load(ctx).Apply("), kw("func"), tx("(_ app.Context, c *"), typ("Counter"), tx(") *"), typ("Counter"), tx(" {\n"),
				tx("\t\tc.Count++\n"),
				tx("\t\t"), kw("return"), tx(" c\n"),
				tx("\t}).Set()\n}"),
			),
			app.Div().Style(theme.Scribble, theme.ScribbleCircle).Aria("hidden", "true").Body(
				app.Raw(`<svg viewBox="0 0 380 110" width="100%" height="100%" preserveAspectRatio="none"><path d="M30 62 C 24 22, 200 6, 330 18 C 380 26, 372 84, 300 96 C 200 110, 60 104, 26 80 C 10 66, 30 44, 60 38"/></svg>`),
			),
			app.Div().Style(theme.Scribble, theme.ScribbleWhole).Aria("hidden", "true").Body(
				app.Text("this is the whole app"),
				app.Raw(`<svg viewBox="0 0 120 90"><path d="M4 70 C 40 80, 80 60, 108 14 M108 14 L 90 18 M108 14 L 110 34"/></svg>`),
			),
		),
		app.Span().Style(theme.Folio, theme.FolioLow, theme.InkYellow).Aria("hidden", "true").Text("01"),
	)
}

func sticker() app.UI {
	return app.Div().Style(theme.Sticker).Aria("hidden", "true").Body(
		app.Raw(`<svg viewBox="0 0 160 160"><defs><path id="sticker-ring" d="M80 80 m-58 0 a58 58 0 1 1 116 0 a58 58 0 1 1 -116 0"/></defs><circle cx="80" cy="80" r="76" style="fill:var(--ink-blue);stroke:var(--on-fill)" stroke-width="4"/><circle cx="80" cy="80" r="44" style="fill:var(--ink-yellow);stroke:var(--on-fill)" stroke-width="3"/><g><text font-family="Space Mono, monospace" font-weight="700" font-size="15" letter-spacing="3" style="fill:var(--on-blue)"><textPath href="#sticker-ring">BUILT WITH GO · BUILT WITH GO · </textPath></text></g><text x="80" y="96" text-anchor="middle" font-family="Anton, Impact, sans-serif" font-size="46" style="fill:var(--on-fill)">GO</text></svg>`),
	)
}

func (l *Layout) press(count int) app.UI {
	n := strconv.Itoa(count)
	stamp := []app.ClassRule{theme.Stamp}
	switch {
	case count == 0:
		stamp = append(stamp, theme.StampIdle)
	case count%2 == 0:
		stamp = append(stamp, theme.StampAgain)
	}
	log := "Ready. Paper loaded."
	if count > 0 {
		log = "Load, Apply, Set. Pull " + n + " printed."
	}
	return app.Section().Style(theme.Section).Aria("label", "Pull the press").Body(
		app.Span().Style(theme.Eyebrow).Text("Try it, it runs"),
		app.Div().Style(theme.Press).Body(
			app.Div().Style(theme.PressCopy).Body(
				heading("Pull the press"),
				app.P().Style(theme.PressText).Text("Every bump of the counter goes through the three model verbs. Load takes the state, Apply changes it, Set stores it and redraws whatever reads it. This landing page is a Screen, so the press below is its own model."),
			),
			app.Div().Style(theme.Bed).Body(
				app.Div().Style(theme.BedHead).Body(
					app.Span().Text(`model "home"`),
					app.Span().Body(app.Text("Count = "), app.Output().Aria("live", "polite").Text(n)),
				),
				app.Div().Style(theme.BedSheet).Aria("hidden", "true").Body(
					app.Span().Style(stamp...).DataSet("ink", n).Text(n),
				),
				app.Div().Style(theme.BedFoot).Body(
					app.Span().Text(log),
					app.Button().Type("button").Style(theme.Button, theme.PressButton).Text("Press +1").OnClick(l.bump),
				),
			),
		),
		folio("02", theme.InkPink),
	)
}

func (l *Layout) bump(ctx app.Context, e app.Event) {
	l.model.Load(ctx).Apply(func(_ app.Context, s *State) *State {
		s.Count++
		return s
	}).Set()
}

func features() app.UI {
	return app.Section().Style(theme.Section).Aria("label", "What you get").Body(
		app.Span().Style(theme.Eyebrow).Text("What you get"),
		heading("Four things in the box"),
		app.Div().Style(theme.Tiles).Body(
			card(nav.Href(state.Route), theme.PlateA, "A", "app.Shell", false, "Pages and routing", "A Shell hosts Screens. Routes, base paths and fragments resolve in the browser, with no server round trip."),
			card(nav.Href(state.Route), theme.PlateB, "B", "app.Model", false, "View models", "Typed state with Load, Apply and Set. Observe it from any component, or react to it from a stylesheet."),
			card(nav.Href(styling.Route), theme.PlateC, "C", "style.Class", false, "Typed CSS", "Properties, keywords, media and container queries, keyframes, nesting and layers as Go values. Dead rules never ship."),
			card(nav.Href(markdown.Route), theme.PlateD, "D", "```go", true, "Markdown with live demos", "Render docs from markdown and drop a running Go component into any fence. Every example on this site runs live."),
		),
		folio("03", theme.InkBlue),
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
	return app.Section().Style(theme.Section).Aria("label", "How it feels").Body(
		app.Div().Style(theme.Two).Body(
			app.Div().Style(theme.TwoText).Body(
				app.Span().Style(theme.Eyebrow).Text("How it feels"),
				heading("One language, one mental model."),
				app.P().Style(theme.Lead).Text("A screen is a Layout and a Model. The layout draws from the model, handlers apply changes, and the screen re-renders. Styles are values you can pass around, test and reuse."),
				app.Ul().Style(theme.Points).Body(
					point("Rename a CSS property and the build breaks, not the page."),
					point("Bind a class to state and the dynamic sheet recomputes on its own."),
					point("Ship the generated site to any static host, GitHub Pages included."),
				),
			),
			app.Div().Style(theme.TwoCode).Body(
				sample("theme.go", "package theme", "fig. 2, a stylesheet as Go values", theme.CodeLean, []app.UI{
					app.Span().Style(theme.TapePink).Aria("hidden", "true"),
				},
					kw("var"), tx(" Button = style."), fn("Class"), tx("("), str(`"button"`), tx(").\n"),
					tx("\t"), fn("Set"), tx("(\n"),
					app.Span().Style(theme.Mark).Body(tx("\t\tcss."), fn("Padding"), tx("(css.Rem(0.6), css.Rem(1.1)),")), tx("\n"),
					tx("\t\tcss.Border(css.Px(2), kw.Solid, css.Var("), str(`"line"`), tx(")),\n"),
					tx("\t\tcss.BoxShadow(css.Var("), str(`"shadow"`), tx(")),\n"),
					tx("\t).\n"),
					tx("\t"), fn("Hover"), tx("(css.Transform(css.Raw("), str(`"translate(-2px, -2px)"`), tx("))).\n"),
					tx("\t"), fn("Media"), tx("(media.MaxWidth(css.Rem(48)), css.Width(css.Pct(100)))\n\n"),
					kw("func"), tx(" (b *Buy) "), fn("Render"), tx("() app.UI {\n"),
					tx("\t"), kw("return"), tx(" app.Button().Style(Button).Text("), str(`"Buy"`), tx(")\n"),
					tx("}"),
				),
				app.Div().Style(theme.Scribble, theme.ScribbleRename).Aria("hidden", "true").Body(
					app.Text("drop one letter here..."),
					app.Raw(`<svg viewBox="0 0 90 40"><path d="M86 6 C 60 4, 30 14, 8 30 M8 30 L 22 32 M8 30 L 14 18"/></svg>`),
				),
				app.Pre().Style(theme.Compiler).Aria("label", "Compiler output").Body(
					app.Span().Style(theme.CompilerPrompt).Text("$ "),
					app.Text("GOARCH=wasm GOOS=js go build -o web/app.wasm .\n"),
					app.Span().Style(theme.CompilerError).Text("./theme.go:14:7: undefined: css.Paddin"),
				),
				app.P().Style(theme.Caption).Text("...and the compiler catches it before a browser ever sees it."),
			),
		),
		app.Span().Style(theme.Folio, theme.FolioLow, theme.InkYellow).Aria("hidden", "true").Text("04"),
	)
}

func point(text string) app.UI {
	return app.Li().Body(
		app.Raw(`<svg aria-hidden="true" width="36" height="30" viewBox="0 0 36 30"><path d="M3 16 L 13 26 L 33 3"/></svg>`),
		app.Span().Text(text),
	)
}

func steps() app.UI {
	return app.Section().Style(theme.Section).Aria("label", "Three steps").Body(
		app.Span().Style(theme.Eyebrow).Text("Three steps"),
		heading("From go get to a running page"),
		app.Div().Style(theme.Steps).Body(
			step("1", theme.InkPink, "Install", "Add WebGo to your module. Everything comes from this one import.", install),
			step("2", theme.InkBlue, "Write a page", "A component, a route and app.RunWhenOnBrowser. Then grow into Screens and Models.", `app.Route("/", func() app.Composer { return &hello{} })`),
			step("3", theme.InkYellow, "Build and ship", "Compile to wasm and generate a static site, or serve it from the same binary.", "GOARCH=wasm GOOS=js go build -o web/app.wasm ."),
		),
		app.Span().Style(theme.Folio, theme.FolioLow, theme.InkPink).Aria("hidden", "true").Text("05"),
	)
}

func latestNews() app.UI {
	latest := news.Latest()
	return app.Section().Style(theme.Section).Aria("label", "Latest news").Body(
		app.Span().Style(theme.Eyebrow).Text("Latest news"),
		heading("Fresh off the drum"),
		app.Div().Style(theme.LatestNews).Body(
			app.A().Style(theme.LatestNewsCover).Href(nav.Href(latest.Route())).TabIndex(-1).Aria("hidden", "true").Body(news.Cover(latest, false)),
			app.Div().Style(theme.NewsStory).Body(
				app.H3().Style(theme.LatestNewsTitle).Body(app.A().Href(nav.Href(latest.Route())).Text(latest.Title)),
				app.P().Style(theme.NewsTeaser).Text(latest.Teaser),
				news.Meta(latest),
				app.Div().Style(theme.Actions).Body(
					app.A().Style(theme.Button, theme.ButtonPrimary).Href(nav.Href(latest.Route())).Body(app.Text("Read the update"), arrow()),
					app.A().Style(theme.Button).Href(nav.Href(news.Route)).Text("All news"),
				),
			),
		),
		folio("06", theme.InkBlue),
	)
}

func closing() app.UI {
	return app.Section().Style(theme.Closing).Body(
		app.Div().Style(theme.ClosingText).Body(
			heading("Ready when you are."),
			app.P().Style(theme.ClosingLead).Text("The getting started guide takes you from an empty module to a static site. Every example on this site is live."),
		),
		app.A().Style(theme.Button, theme.ButtonBlue).Href(nav.Href(start.Route)).Body(app.Text("Open the docs"), arrow()),
	)
}

func footer() app.UI {
	return app.Footer().Style(theme.Footer).Body(
		app.P().Style(theme.FooterBig).Body(mis("Built with WebGo")),
		app.Div().Style(theme.FooterRow).Body(
			app.Span().Text("Docs written in Go, styled with typed CSS. Printed in fluorescent pink, federal blue and yellow, with black for the words."),
			app.A().Style(theme.Pill).Href(base.Repository).Text("github.com/Toms223/WebGo"),
		),
		app.Span().Style(theme.ColorBar).Aria("hidden", "true"),
		app.Div().Style(theme.FooterFine).Body(
			app.Span().Text("GOARCH=wasm GOOS=js"),
			app.Span().Text("Zine No. 1.0.0 · page 06 of 06"),
		),
	)
}

func card(href string, plate *style.Rule, letter, tag string, light bool, title, text string) app.UI {
	letterClasses := []app.ClassRule{theme.PlateLetter}
	if light {
		letterClasses = append(letterClasses, theme.PlateLetterLight)
	}
	return app.A().Style(theme.Card).Href(href).Body(
		app.Div().Style(theme.Plate, plate).Aria("hidden", "true").Body(
			app.Span().Style(letterClasses...).Text(letter),
			app.Span().Style(theme.PlateTag).Text(tag),
		),
		app.H3().Style(theme.CardTitle).Text(title),
		app.P().Style(theme.CardText).Text(text),
	)
}

func step(number string, ink *style.Rule, title, text, code string) app.UI {
	return app.Div().Style(theme.Step).Body(
		app.Span().Style(theme.StepNumber, ink).Aria("hidden", "true").Text(number),
		app.H3().Style(theme.StepTitle).Text(title),
		app.P().Style(theme.StepText).Text(text),
		app.Div().Style(theme.Code, theme.CodeQuiet).Body(app.Pre().Text(code)),
	)
}

func sample(name, note, caption string, extra app.ClassRule, tapes []app.UI, body ...app.UI) app.UI {
	return app.Div().Style(theme.Code, extra).Body(
		append(tapes,
			app.Div().Style(theme.CodeBar).Body(app.Span().Text(name), app.Span().Style(theme.CodeNote).Text(note)),
			app.Pre().Body(body...),
			app.Div().Style(theme.CodeFoot).Text(caption),
		)...,
	)
}

func separator() app.UI {
	return app.Li().Style(theme.MarqueeDot).Aria("hidden", "true").Text("·")
}

func kw(text string) app.UI { return app.Span().Style(theme.Keyword).Text(text) }

func fn(text string) app.UI { return app.Span().Style(theme.Fn).Text(text) }

func str(text string) app.UI { return app.Span().Style(theme.Str).Text(text) }

func typ(text string) app.UI { return app.Span().Style(theme.Typ).Text(text) }

func tx(text string) app.UI { return app.Text(text) }

func arrow() app.UI {
	return app.Raw(`<svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14M13 6l6 6-6 6"/></svg>`)
}
