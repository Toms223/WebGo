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

func firstEdition() Article {
	return Article{
		Slug:      "1-0-0-first-edition",
		Version:   "1.0.0",
		Codename:  "First Edition",
		Date:      "19 November 2026",
		Minutes:   9,
		Fill:      theme.FillLilac,
		Signature: "go get WebGo@v1.0.0",
		Title:     "WebGo 1.0, and the release we skipped",
		Teaser:    "Loading before drawing, pages in the engine, typed scripts, feature flags and a dev server. All of it lands at once as v1.0.0, and v0.2.0 never happens. Here's the whole story.",
		Short: []string{
			"There is no v0.2.0. Everything planned for it, and everything planned after it, is in v1.0.0.",
			"Any component can implement `OnLoad`. It is not drawn until its load is done.",
			"`app.Screen`, `app.Shell` and `app.Model` replace the `page` and `data` packages.",
			"Scripts: list ES modules on `Handler.Imports` and call them from Go through `script.Use`.",
			"Feature flags: an OpenFeature SDK in Go that runs in the browser.",
			"`dev.RunDevServer` serves your site locally and rebuilds it when you save.",
			"WebGo is MIT now, and the docs got a new coat of ink.",
			"Breaking: `page` and `data` are deleted. One list gets you across, and you only cross once.",
		},
		SignOff: "That's the first edition. Print something with it and tell us what smudged. See you next update.",
		Body: func() []app.UI {
			return []app.UI{
				doc.P("Hey everyone,"),
				doc.P("This is the big one. When we brought go-app's engine into WebGo in 0.1.0, we wrote down what had to be true before we'd call it 1.0: keep everything that worked, move pages into the engine, let components load before they draw, give JavaScript modules a typed door, and ship feature flags that run in the browser. Every line of that list is done. So this update is 1.0.0."),
				h("Wait, what happened to 0.2.0?"),
				doc.P("If you've been reading the README or these docs over the last two weeks, you saw `v0.2.0` everywhere. It even had a name, Moving Day, and an article sitting right here. It was real, it just never got a tag."),
				doc.P("The plan was two steps. 0.2.0 would move pages into the engine and add loading and scripts. Then feature flags would follow, and that would be 1.0. We finished the first step on 9 November and rewrote the docs for it the next day. But we held the tag, because the flags work was already under way and it reached into the same places."),
				doc.P("Then the flags work finished too. Two days after the docs said 0.2.0, the last thing between it and 1.0 was done, and the dev server we'd filed under extras showed up as well. At that point 0.2.0 would have been a tag with a shelf life of about a week, and it would have cost you something real: it deletes two packages. You'd migrate to it, and then update again right after."),
				doc.P("So we skipped it. One breaking release instead of two, and one migration guide. If you're on 0.1.0, your next stop is 1.0.0. Nothing was dropped on the way: everything that article promised is below."),
				h("Load before you draw"),
				doc.P("This is the one that started it. A screen that wanted its data before it drew used to fake it with flags, because the engine wasn't ours to teach. Now any component can implement `OnLoad`. It runs in the browser, on its own goroutine, after `OnMount` and again after `OnUpdate`, and the component isn't drawn until it returns. Give it a `Loading()` to show something meanwhile, and an `Error(err)` for when the load fails."),
				doc.Go(`func (p *profileCard) OnLoad(ctx app.Context) error {
	return fetchProfile(ctx, &p.profile)
}

func (p *profileCard) Loading() app.UI {
	return app.Text("Loading…")
}`),
				doc.P("The wait is per component. Its siblings and its parents keep rendering, so parts of one page can load on their own clocks. The ", doc.A("/state#loading-side-by-side", "State page"), " has a live demo of two of them side by side."),
				doc.P("One thing it doesn't do yet: keep the old screen up while the next one loads. You get the new screen's `Loading()` in between. That part isn't built yet."),
				h("Pages in the core"),
				doc.P("The `page` and `data` packages were written on top of go-app, from the outside. Their ideas now live in `app`, next to `Compo`, where the engine can drive them. A screen fetches its data with `WithLoader`:"),
				doc.Go(`profileScreen := app.NewScreen[Profile]("/profile").
	WithLayout(&profileLayout{}).
	WithModel(profile).
	WithLoader(fetchProfile)`),
				doc.P("One `Shell` is the root for the whole app, and `app.Serve` registers every screen's route for you. In the browser the shell is built once and kept across navigations: moving between pages switches the screen, it doesn't rebuild everything."),
				doc.P("Every `Set` on a model now fires the action named after its key, so styles bound to that key follow the state with no `OnAction` call to remember."),
				h("Scripts, typed"),
				doc.P("Need a JavaScript library? Declare it as a module: a Go type that says where it comes from and what Go gets back. List it on `Handler.Imports`, and the handler writes one importmap so the browser fetches it while `app.wasm` downloads."),
				doc.Go(`type Confetti struct{}

func (Confetti) Source() app.ModuleSource { return script.CDN("npm", "canvas-confetti", "1.9.3") }

func (Confetti) Init(ctx app.Context, mod app.Value) Cannon { return Cannon{fire: mod.Get("default")} }`),
				doc.P("From then on `script.Use(ctx, Confetti{})` hands you a `Cannon`, typed, with a `bool` that is `false` when the module didn't load. `script.Local` serves a file from your own `web` directory the same way. The ", doc.A("/scripts", "Scripts page"), " has the whole thing, with a button that fires confetti."),
				h("Feature flags"),
				doc.P("This is the piece 0.2.0 was waiting for. OpenFeature is the vendor-neutral standard for feature flags, and its official Go SDK only runs on servers. So we wrote one from scratch that runs in the browser: one API, clients, hooks, events and an evaluation context."),
				doc.Go(`on := openfeature.GetClient().Boolean("new-nav", false)`),
				doc.P("If anything goes wrong, you get the default you passed. Three providers come with it: `memprovider` keeps flags in memory for tests and local work, `jsprovider` wraps any JavaScript OpenFeature web provider, and `flipt` loads the official Flipt one for you."),
				doc.P("We were stubborn about this one. It follows the OpenFeature specification v0.9.0, and a test walks every requirement in it. On top of that, the same scenarios run against the official Web SDK and against ours, and what each one evaluates, emits and logs has to match. The handful of places where Go has to differ are listed on the ", doc.A("/feature-flags#parity", "Feature flags page"), ", one by one. That's most of why it took as long as it did, and why 0.2.0 kept waiting."),
				h("A dev server"),
				doc.P("We had this down as an extra for after 1.0. It got done early, so here it is. `dev.RunDevServer` builds your static site, serves it, and rebuilds when you save a Go file or anything in `web`."),
				doc.Go(`err := dev.RunDevServer(ctx, ".dev", "localhost:8000", func(dir string) *exec.Cmd {
	return exec.Command("go", "run", ".", "-static", dir)
})`),
				doc.P("Break the build and it tells you, and keeps serving the last one that worked. Fix it, switch back to your tab, and the page moves to the new build by itself. This site's own `go run . dev` is built on it. ", doc.A("/deploy#running-locally", "Running locally"), " has a full `main` to copy."),
				h("A new coat of ink"),
				doc.P("You may have noticed the place looks different. The docs are a riso zine now: paper grain, off-register titles, taped-on code printouts. The pages themselves are written in Go instead of markdown, which is how every example on them can be a live component and not a screenshot. All of it is styled with WebGo's own typed CSS."),
				h("A word on the license, again"),
				doc.P("In 0.1.0 we moved to the GPL-3.0. We've changed our minds: WebGo is MIT as of this release. Use it in anything, open source or not. The go-app code inside was MIT all along, and the NOTICE file still credits it."),
				h("What's gone"),
				doc.List(
					doc.Item("The `page` and `data` packages."),
					doc.Item("`OnAction`, and the `Content()` method screens had. A screen is a component now, and it draws with `Render`."),
					doc.Item("The shell working out its own mount point. It reads `GOAPP_ROOT_PREFIX`, which `app.Handler` sets from `Resources`, so `BasePath()` is gone too."),
				),
				h("How to update"),
				doc.List(
					doc.Item("`page.NewPage` becomes `app.NewScreen`."),
					doc.Item("`page.NewBase(layout, model, []page.Screen{...})` becomes `app.NewShell(layout, model, screens...)`, returned from the function you pass to `app.Serve`."),
					doc.Item("`data.NewViewModel` and `data.ViewModel[T]` become `app.NewModel` and `app.Model[T]`. `data.Action[T]` becomes `app.Change[T]`."),
					doc.Item("`page.Layout[T]` becomes `app.Layout[T]`, and `page.Screen` becomes `app.Routed`."),
					doc.Item("Delete every `OnAction` call. If you used `BasePath()`, read `app.Getenv(\"GOAPP_ROOT_PREFIX\")` instead."),
				),
				doc.P("The ", doc.A("/migrating#from-v0-1-0", "migration guide"), " has the before and after. The import path stays the same."),
				h("What 1.0 means"),
				doc.P("It means the list we wrote down is finished, and that we're done moving your furniture. From here, a change that breaks your code needs a 2.0, and we're in no hurry to write one."),
				h("Get it"),
				doc.Sh(`go get github.com/Toms223/WebGo@v1.0.0`),
			}
		},
	}
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
