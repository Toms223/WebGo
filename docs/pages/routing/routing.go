package routing

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
)

const (
	Title = "Routing"
	Route = "/routing"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "routing",
		Path:  Route,
		Title: Title,
		Lead:  "Routes map URL paths to the component that renders them, and navigation switches between them without reloading the page.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Registering routes", Body: registering},
			{Title: "Regular expressions", Body: regexps},
			{Title: "Navigating", Body: navigating},
			{Title: "Links and internal URLs", Body: links},
			{Title: "The page", Body: thePage},
			{Title: "Fragments and OnNav", Body: fragments},
			{Title: "Serving under a prefix", Body: prefix},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("This page covers the `app` package's router: registering routes, moving between them from code and from links, reading and changing the current page, and hosting the app under a path prefix such as a GitHub Pages project site."),
		doc.P("Components themselves, and the `OnNav` hook among their lifecycle events, are on the ", doc.A("/components", "Components"), " page. WebGo's `page.Base`, which switches screens by route inside one component, is on the ", doc.A("/state#base", "Pages, layouts and state"), " page."),
	}
}

func registering() []app.UI {
	return []app.UI{
		doc.Go(`func Route(path string, newComponent func() Composer)`),
		doc.P("`Route` maps a path to a function that returns a new component. Register every route before `app.RunWhenOnBrowser`, in the `main` shared by the server and the browser: the server needs the routes to prerender a page, the browser to mount it."),
		doc.Go(`package main

import (
	"net/http"

	"github.com/Toms223/WebGo/app"
)

type home struct{ app.Compo }

func (h *home) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Home"),
		app.A().Href("/about").Text("About"),
	)
}

type about struct{ app.Compo }

func (a *about) Render() app.UI {
	return app.H1().Text("About")
}

func main() {
	app.Route("/", func() app.Composer { return &home{} })
	app.Route("/about", func() app.Composer { return &about{} })
	app.RunWhenOnBrowser()

	http.Handle("/", &app.Handler{Name: "Routes"})
	http.ListenAndServe(":8080", nil)
}`),
		doc.List(
			doc.Item("The path is matched exactly against the URL path, without query or fragment. `/about` does not match `/about/`."),
			doc.Item("Registering the same path again replaces the earlier function."),
			doc.Item("Exact routes are tried first, then ", doc.A("#regular-expressions", "regular-expression routes"), " in the order they were registered."),
		),
		doc.H("What the function is for"),
		doc.P("The function is called on every navigation to a new path, and its component is compared with the root that is mounted, by the rules in ", doc.A("/components#writing-a-component", "Writing a component"), ". If the types differ, the old root is dismounted and the new one mounted. If they are the same type, the mounted root stays: differing exported fields are copied from the fresh value, unexported state is kept, and `OnMount` does not run again. Return a fresh value every time, never a shared one."),
		doc.P("WebGo's docs use that second case on purpose: every route returns a `page.Base`, which stays mounted and swaps its screen in `OnNav`."),
		doc.H("Unmatched paths"),
		doc.P("A path with no route shows `app.NotFound`, a built-in \"Not Found\" component by default. On the server, `app.Handler` prerenders the page with it and answers `404`; in the browser, it is mounted like any root."),
		doc.Go(`type missing struct{ app.Compo }

func (m *missing) Render() app.UI {
	return app.Main().Body(app.H1().Text("Page not found"))
}

func main() {
	app.NotFound = &missing{}
	app.Route("/", app.NewZeroComponentFactory(&home{}))
	app.RunWhenOnBrowser()
}`),
		doc.P("Assign a pointer to your own component type before `RunWhenOnBrowser`. Each miss mounts a new zero value of that type, so fields set on the assigned value are not kept."),
	}
}

func regexps() []app.UI {
	return []app.UI{
		doc.Go(`func RouteWithRegexp(pattern string, newComponent func() Composer)
func NewZeroComponentFactory(c Composer) func() Composer`),
		doc.P("`RouteWithRegexp` maps every path that matches a Go regular expression. The pattern is compiled when the route is registered, and an invalid one panics. Matching is not anchored, so `/products` also matches `/old/products/1`; start patterns with `^` and end them with `$`."),
		doc.P("`NewZeroComponentFactory(&product{})` returns a function that makes a new zero value of the pointed-to type each time. It must be given a pointer."),
		doc.Go(`type product struct {
	app.Compo
	id string
}

func (p *product) OnNav(ctx app.Context) {
	p.id = path.Base(ctx.Page().URL().Path)
	ctx.Page().SetTitle("Product " + p.id)
}

func (p *product) Render() app.UI {
	return app.H1().Textf("Product %s", p.id)
}

type missing struct{ app.Compo }

func (m *missing) Render() app.UI {
	return app.P().Text("There is nothing here.")
}

func main() {
	app.Route("/", func() app.Composer { return &home{} })
	app.RouteWithRegexp("^/products/[0-9]+$", app.NewZeroComponentFactory(&product{}))
	app.RouteWithRegexp(".*", app.NewZeroComponentFactory(&missing{}))
	app.RunWhenOnBrowser()
}`),
		doc.P("The router hands nothing to the component. It reads the path itself, in `OnNav`, which runs on the server while prerendering and in the browser on every navigation, so the served HTML already has the right heading and title. `ctx.Page().URL().Path` is the browser's path, mount prefix included, which is why the example takes its last segment instead of trimming `/products/`."),
		doc.P("A catch-all route counts as a route: the server prerenders every path with `200` instead of answering `404`. For missing pages, prefer `app.NotFound`, under ", doc.A("#unmatched-paths", "Unmatched paths"), "."),
		doc.Note("`app.GenerateStaticWebsite` writes one HTML file per exact route. It cannot list the paths a pattern matches, so pass them as its `pages` arguments: `app.GenerateStaticWebsite(\".\", h, \"/products/1\", \"/products/2\")`."),
	}
}

func navigating() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) Navigate(rawURL string)
func (ctx Context) NavigateTo(u *url.URL)
func (ctx Context) Reload()`),
		doc.P("`Navigate` parses `rawURL` and calls `NavigateTo`. If it does not parse, the error is logged and nothing happens. A URL without a host is given the app's host. `NavigateTo` then takes the first case that applies:"),
		navigationDiagram(),
		doc.Ordered(
			doc.Item("A URL that starts with one of `Handler.InternalURLs`, a `mailto:` URL or a `tel:` URL is loaded by the browser in the same tab, leaving the app."),
			doc.Item("A URL on another host is opened in a new window."),
			doc.Item("A URL equal to the last one navigated to does nothing: no history entry, no `OnNav`."),
			doc.Item("A URL with the same path as the last one and a different fragment is ", doc.A("#fragments-and-onnav", "a fragment change"), ": the root is not touched."),
			doc.Item("Any other URL creates the routed component and loads it as the root."),
		),
		doc.P("In the last two cases, a history entry is pushed, the URL is recorded as the last one, `OnNav` runs in every mounted component that has it, and, if the URL has a fragment, the element with that ID is scrolled into view after the next render."),
		doc.Go(`func (c *login) submit(ctx app.Context, e app.Event) {
	e.PreventDefault()
	if c.signIn() {
		ctx.Navigate("/account")
	}
}`),
		doc.List(
			doc.Item("Back and forward follow the same rules for the browser's new URL, without pushing a history entry."),
			doc.Item("JavaScript on the page can call `goappNav(url)` to navigate the same way."),
			doc.Item("`Reload` reloads the page from the server. On the server it does nothing."),
		),
	}
}

func links() []app.UI {
	return []app.UI{
		doc.P("A plain `app.A().Href(\"/about\")` is enough to navigate inside the app. The app listens for clicks on the whole window, and a click on an `<a>`, or on anything inside one, is turned into `ctx.Navigate(href)` with the browser's default action prevented. The browser handles the click itself when:"),
		doc.List(
			doc.Item("Ctrl or Cmd is held, to open a new tab;"),
			doc.Item("the link has a `download` attribute;"),
			doc.Item("the link has `target=\"_blank\"`."),
		),
		doc.P("Every other link goes through the rules under ", doc.A("#navigating", "Navigating"), ", so a link to another host opens a new window. For an external link, setting `Target(\"_blank\")` and `Rel(\"noopener\")` yourself leaves it entirely to the browser."),
		doc.Go(`app.Nav().Body(
	app.A().Href("/").Text("Home"),
	app.A().Href("/about#team").Text("Team"),
	app.A().Href("https://github.com/Toms223/WebGo").Target("_blank").Rel("noopener").Text("Source"),
	app.A().Href("/report.pdf").Download("report.pdf").Text("Report"),
)`),
		doc.H("InternalURLs"),
		doc.Go(`type Handler struct {
	InternalURLs []string
}`),
		doc.P("`InternalURLs` lists URL prefixes, compared with the start of the full URL, that leave the app in the same tab instead of being routed or opened in a new window. Use it for pages that are not part of the app: a sign-in provider, or another site on the same domain."),
		doc.Go(`&app.Handler{
	Name:         "Shop",
	InternalURLs: []string{"https://accounts.example.com/"},
}`),
	}
}

func thePage() []app.UI {
	return []app.UI{
		doc.Go(`func (ctx Context) Page() Page

type Page interface {
	URL() *url.URL
	ReplaceURL(v *url.URL)
	Title() string
	SetTitle(v string)
	SetTitlef(format string, v ...any)
	Lang() string
	SetLang(v string)
	Description() string
	SetDescription(v string)
	SetDescriptionf(format string, v ...any)
	Author() string
	SetAuthor(v string)
	SetAuthorf(format string, v ...any)
	Keywords() string
	SetKeywords(v ...string)
	Image() string
	SetImage(v string)
	SetImagef(format string, v ...any)
	Preloads() []Preload
	SetPreloads(v ...Preload)
	SetLoadingLabel(v string)
	SetLoadingLabelf(format string, v ...any)
	SetTwitterCard(v TwitterCard)
	SetCanonicalLink(v string)
	SetCanonicalLinkf(format string, v ...any)
	Size() (w int, h int)
}`),
		doc.P("`ctx.Page()` is the current page, and what it changes depends on where the code runs:"),
		doc.List(
			doc.Item("On the server, while a page is prerendered, it is the page being served. Setters change its `<head>`: title, description, Open Graph and Twitter tags, canonical link and preloads. They start from the `app.Handler` fields of the same names, so a component overrides them per route, typically in `OnPreRender` or `OnNav`. `URL` is the request URL and `Size` is `0, 0`."),
			doc.Item("In the browser, it is the live document. `SetTitle` changes the tab title and `og:title`, the other setters their `<meta>` tags. `URL` is the address bar's URL and `Size` the window's inner size."),
		),
		doc.P("Paths given to `SetImage` and `SetTwitterCard` go through `Handler.Resources`, like other `/web/` resources."),
		doc.H("ReplaceURL"),
		doc.P("`ReplaceURL(u)` replaces the current history entry with `u`, and `og:url` with it. It does not navigate: no component is created and `OnNav` does not run. Use it to keep the address bar in step with state, such as a filter in the query string. On the server it only changes the URL the rest of the prerender sees."),
		doc.Go(`func (c *search) filter(ctx app.Context, e app.Event) {
	u := *ctx.Page().URL()
	u.RawQuery = url.Values{"q": {c.query}}.Encode()
	ctx.Page().ReplaceURL(&u)
}`),
		doc.P("The router does not see the change: the last URL it navigated to stays the one before `ReplaceURL`. The consequence is under ", doc.A("#fragments-and-onnav", "Fragments and OnNav"), "."),
	}
}

func fragments() []app.UI {
	return []app.UI{
		doc.P("`OnNav` is how a mounted component learns that the URL changed. It runs in every mounted component that implements `app.Navigator`, top-down, after the new root (if any) is mounted, on the first load and on every navigation that is not a no-op."),
		doc.P("When only the fragment changes, the path is the same as the last navigation, so the root stays as it is and its fields are not touched. A history entry is still pushed, `OnNav` still runs, and the element whose ID is the fragment is scrolled into view after the next render. A component can read the fragment in `OnNav` with `ctx.Page().URL().Fragment`, for example to switch tabs."),
		doc.Go(`func (d *docs) OnNav(ctx app.Context) {
	d.section = ctx.Page().URL().Fragment
}`),
		doc.H("When a fragment link reloads the root"),
		doc.P("The fragment case compares paths with the last URL the router navigated to, not with the address bar. If the address bar was changed by `ReplaceURL` (by your code, or by `page.Base` restoring a mount prefix), a link to the address bar's path with a fragment has a different path from the last navigation. The router then creates the routed component again and loads it as the root, and the mounted root's exported fields are reset to the fresh value's."),
		doc.P("Keep the paths of your links equal to what ends up in the address bar. Under a mount prefix that means prefixing every link, as shown in ", doc.A("#serving-under-a-prefix", "Serving under a prefix"), "."),
	}
}

func prefix() []app.UI {
	return []app.UI{
		doc.P("A GitHub Pages project site is served under the repository name, such as `https://user.github.io/shop/`. The app is generated with `app.GitHubPages`, which prefixes every resource with `/shop`:"),
		doc.Go(`func main() {
	app.Route("/", func() app.Composer { return &home{} })
	app.Route("/about", func() app.Composer { return &about{} })
	app.RunWhenOnBrowser()

	err := app.GenerateStaticWebsite("dist", &app.Handler{
		Name:      "Shop",
		Resources: app.GitHubPages("shop"),
	})
	if err != nil {
		log.Fatal(err)
	}
}`),
		doc.List(
			doc.Item("`GitHubPages(repo)` is `PrefixedLocation(\"/\" + repo)`. `/web/app.wasm` and every other `/web/` path is written as `/shop/web/...`."),
			doc.Item("The prefix is passed to the browser as the environment variable `GOAPP_ROOT_PREFIX`, and the router removes it from the path before matching routes, so `/shop/about` matches `/about`. Without a prefix the variable is `/`."),
			doc.Item("`GenerateStaticWebsite` writes `index.html` and one `<route>.html` per exact route, which GitHub Pages serves at `/shop/` and `/shop/about`. The files are build output; see ", doc.A("/start#generate-the-static-site", "Generate the static site"), "."),
			doc.Item("`app.Handler` matches page requests against routes with the path as it arrives, prefix included, so a mount prefix is meant for a generated static site."),
		),
		doc.H("Links under a prefix"),
		doc.P("Routing ignores the prefix, but the address bar does not. A link to `/about` matches the route, then leaves `/about` in the address bar, outside the site, so a reload or a shared link is a GitHub `404`. Build in-app links from the prefix instead:"),
		doc.Go(`func href(route string) string {
	return strings.TrimSuffix(app.Getenv("GOAPP_ROOT_PREFIX"), "/") + route
}

app.A().Href(href("/about")).Text("About")`),
		doc.P("Trimming the slash matters: without a prefix the variable is `/`, and `\"/\" + \"/about\"` is `//about`, a link to a host named `about`."),
		doc.H("page.Base"),
		doc.P("WebGo's `page.Base` works out the prefix by itself: on mount, it compares the address bar's path with its screens' routes, keeps what comes before the matching route as its base path, matches screens with that base removed, and restores the base in the address bar with `ReplaceURL` when a link left it out. Links still need the prefix, for the reason under ", doc.A("#fragments-and-onnav", "Fragments and OnNav"), ". See ", doc.A("/state#base", "Base"), "."),
	}
}
