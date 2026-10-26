package news

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/theme"
)

func h(title string) app.UI {
	return app.H2().Style(theme.DocHeading).Text(title)
}

func homecoming() Article {
	return Article{
		Slug:      "0-1-0-homecoming",
		Version:   "0.1.0",
		Codename:  "Homecoming",
		Date:      "26 October 2026",
		Minutes:   5,
		Fill:      theme.FillSun,
		Signature: "app.Div().Style(theme.Card)",
		Title:     "We brought the engine home",
		Teaser:    "go-app now lives inside WebGo. One import, one handler, styles that go straight onto your elements, and updates that land on their own.",
		Short: []string{
			"go-app is now WebGo's own `app` package. Change one import and your code keeps working.",
			"Elements take your style rules directly: `app.Div().Style(theme.Card)`.",
			"`serve.Handler` is gone. `app.Handler` does the job and builds your stylesheets by itself.",
			"New builds reach open tabs on their own.",
			"A batch of fixes, one of them for security. Please update.",
		},
		SignOff: "That's it for this one. Grab it, break it, tell us what broke. See you next update.",
		Body: func() []app.UI {
			return []app.UI{
				doc.P("Hey everyone,"),
				doc.P("For its first month, WebGo was a layer on top of go-app. That got us a long way, but every time we wanted to change how the engine itself behaved, we hit the same wall: the engine wasn't ours to change. Update 0.1.0 takes that wall down. go-app's core now lives inside WebGo as the `app` package, with its MIT notice right where it belongs, and this whole update was about making the two feel like one thing."),
				doc.P("Here's what that means for you."),
				h("One import"),
				doc.P("Everything comes from one module now. Components, routing, state and actions keep their names and their behaviour. Only the address changes."),
				doc.Go(`import "github.com/Toms223/WebGo/app"`),
				doc.P("Coming from go-app, or from 0.0.5? The ", doc.A("/migrating", "migration guide"), " walks through every change, and most apps are done in a few minutes."),
				h("Styles go straight onto your elements"),
				doc.P("This is the one we've wanted since typed CSS showed up in 0.0.5. You used to copy a class name out of a rule and hand that string to an element. Now you hand over the rule itself."),
				&compare{
					BeforeLabel: "Before, 0.0.5",
					AfterLabel:  "Now, 0.1.0",
					Before:      `app.Div().Class(theme.Card.Name(), theme.Compact.Name())`,
					After:       `app.Div().Style(theme.Card, theme.Compact)`,
				},
				doc.P("The compiler also knows which elements can be styled at all. Put a style rule on a `<meta>` tag and the build stops right there, instead of you wondering later why nothing happened. When you really do need inline CSS, `RawStyle` is still around."),
				h("One handler"),
				doc.P("`serve.Handler` used to wrap go-app's handler just to add stylesheets. Now there's only `app.Handler`, and it takes your styles directly. It builds them the first time it needs them, so forgetting `Build()` no longer ships a page without its CSS. We still recommend calling it at startup: it's the quickest way to catch a missing file before your users do."),
				doc.Go(`handler := &app.Handler{
	Name:   "My app",
	Styles: theme.Styles(),
}
http.Handle("/", handler)`),
				h("Updates that actually arrive"),
				doc.P("If you've ever shipped a fix, opened a tab you'd left open, and found yesterday's build staring back at you, this one's for you. The app now looks for a new version when the page opens, when you come back online, when you switch back to its tab, and once a minute while it's open. When the new version is ready, the page reloads once and you're on it."),
				doc.P("Got a form where a surprise reload would hurt? Set `DisableAutoUpdate` and you decide when to switch. ", doc.A("/deploy#versions-and-updates", "Versions and updates"), " has the details."),
				h("Fixes we picked up on the way"),
				doc.P("Moving the engine in meant reading every line of it, and we found a few things worth fixing while we were there."),
				doc.List(
					doc.Item("Attribute values in server-rendered pages are escaped properly now. Before, a value with a quote in it could break out of its attribute. If you put user text into attributes, update today."),
					doc.Item("Setting `app.NotFound` finally works, on the server and in the browser, with a real 404."),
					doc.Item("Apps served under a sub-path, like a GitHub Pages project site, serve their files correctly from a live server too."),
					doc.Item("Your own icon is your favicon. No more surprise go-app logo."),
					doc.Item("Showing the install prompt no longer throws an error on Safari."),
					doc.Item("Static sites now include library stylesheets."),
					doc.Item("Server output is identical from run to run, so snapshot tests stop flaking."),
				),
				h("The docs grew up"),
				doc.P("There's a whole new Core section: components, routing, state and actions, concurrency, deployment, testing and migrating. A lot of it started from go-app's own documentation, rewritten in our voice and checked line by line against the code. Where a picture says it better than a paragraph, there's a diagram now. ", doc.A("/components", "Components"), " is a good place to start."),
				h("A word on the license"),
				doc.P("WebGo is now released under the GPL-3.0. In short: if you build on WebGo and ship it, your project stays open source too. The go-app code inside keeps its MIT license, and the NOTICE file credits it."),
				h("Get it"),
				doc.Sh(`go get github.com/Toms223/WebGo@v0.1.0`),
			}
		},
	}
}

func typedInk() Article {
	return Article{
		Slug:      "0-0-5-typed-ink",
		Version:   "0.0.5",
		Codename:  "Typed Ink",
		Date:      "11 October 2026",
		Minutes:   4,
		Fill:      theme.FillLime,
		Signature: `css.Padding(css.Rem(1))`,
		Title:     "Your CSS compiles now",
		Teaser:    "Typed styling lands. Write your CSS in Go, let the compiler catch the typos, and let your app's state drive the look.",
		Short: []string{
			"Every CSS property, keyword and unit is a Go value. Typos stop at the compiler.",
			"Hover states, nesting, media, container and supports queries, keyframes, fonts and layers.",
			"Stylesheets ship as files you wrote, files we write for you, or rules that live in the browser.",
			"Rules can follow your app's state, with no class juggling.",
			"The `model` package is now called `data`.",
		},
		SignOff: "Go write some CSS that can't break quietly. We'll see you next update.",
		Body: func() []app.UI {
			return []app.UI{
				doc.P("Hey everyone,"),
				doc.P("CSS has one trick it plays on everybody: it fails quietly. You type `colr: red`, nothing happens, no error, and a week later someone asks why the button is grey. Update 0.0.5 ends that."),
				h("CSS, written in Go"),
				doc.P("Every CSS property, keyword and unit is now a Go value. They're generated from the W3C's own specifications and the browser support data behind MDN, and we only include what every major browser can render today. If it compiles, browsers understand it."),
				doc.Go(`var Card = style.Class("card").
	Set(css.Padding(css.Rem(1)), css.Border(css.Px(2), kw.Solid, css.Var("line"))).
	Hover(css.Background(css.Var("sun")))`),
				doc.P("Misspell `Padding` and your editor tells you before you save. Pass a colour where a length belongs and the build stops."),
				h("The whole toolbox"),
				doc.P("Hover and focus states, nesting, `@media`, `@container` and `@supports`, keyframes, font faces, custom properties, cascade layers and `!important`. If you'd reach for it in a stylesheet, it's in there. The ", doc.A("/styling", "Styling"), " page has a live demo for each one."),
				h("Three ways to ship it"),
				doc.List(
					doc.Item("Static: link a stylesheet you already have."),
					doc.Item("Generated: write your rules in Go and get a real `.css` file at build time."),
					doc.Item("Dynamic: rules that live in the browser and come and go while the app runs."),
				),
				h("Styles that react"),
				doc.P("Bind a rule to an action and it switches itself on and off. Pair it with a ViewModel's `OnAction` and the look of your app follows its state on its own. The dark mode switch on this site is one of those rules."),
				h("One rename"),
				doc.P("The `model` package is now called `data`. Same API, clearer name. Update the import and you're done."),
				h("New docs to match"),
				doc.P("To show all of this off, the docs got rebuilt: a new look, a proper landing page, and live examples you can poke at on every styling page. Each one is a running WebGo component."),
				h("Get it"),
				doc.Sh(`go get github.com/Toms223/WebGo@v0.0.5`),
			}
		},
	}
}

func firstLight() Article {
	return Article{
		Slug:      "launch-first-light",
		Version:   "0.0.4",
		Label:     "0.0.1 – 0.0.4",
		Codename:  "First Light",
		Date:      "22 September 2026",
		Minutes:   3,
		Fill:      theme.FillCoral,
		Signature: `page.NewPage[AppState]("/settings")`,
		Title:     "WebGo is live",
		Teaser:    "Pages, view models and markdown for Go web apps. Four releases on day one, and the honest story of the first two.",
		Short: []string{
			"`Page` pairs a layout with a view model.",
			"`Base` picks the right page for the URL, even under a sub-path.",
			"`ViewModel` holds the typed state your pages share.",
			"Markdown rendering for pages written as text.",
			"Skip 0.0.1 and 0.0.2. Start at 0.0.4.",
		},
		SignOff: "Thanks for being here on day one. There's plenty more coming.",
		Body: func() []app.UI {
			return []app.UI{
				doc.P("Hey everyone,"),
				doc.P("Today WebGo goes public. It's a small layer on top of go-app that takes care of the parts every app ends up writing by hand: the pages, the shell that switches between them, and the state they share."),
				h("What's in the box"),
				doc.List(
					doc.Item("`Page`: a screen made of a layout and a view model. The layout draws, the model holds the data."),
					doc.Item("`Base`: the shell around your pages. It picks the right one for the URL."),
					doc.Item("`ViewModel`: typed state under a key, read from any page and changed with `Load`, `Apply` and `Set`."),
					doc.Item("Markdown rendering, for pages that read better as text."),
				),
				doc.Go(`p := page.NewPage[AppState]("/settings").
	WithLayout(&settingsLayout{}).
	WithModel(vm)`),
				h("Launch day, the honest version"),
				doc.P("Launch days are never clean, and ours wasn't either. 0.0.1 and 0.0.2 went out with a broken module setup, and Go's module proxy keeps every version exactly as it was first published, so those two will stay broken. Please skip them."),
				doc.P("0.0.3 added `ViewModel.State()`, so you can read the current state without loading it again. 0.0.4 taught `Base` to work when your app lives under a sub-path, like a GitHub Pages project site. Four releases in one day, and 0.0.4 is the one to start with."),
				h("The docs eat their own cooking"),
				doc.P("This documentation site is itself a WebGo app. If something on it breaks, that's a WebGo bug, and we'd love to hear about it on ", doc.A(base.Repository, "GitHub"), "."),
				h("Get it"),
				doc.Sh(`go get github.com/Toms223/WebGo@v0.0.4`),
			}
		},
	}
}
