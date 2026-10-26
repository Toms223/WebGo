package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func prerendering() []app.UI {
	return []app.UI{
		doc.P("When a browser or a crawler asks for a routed path, the handler renders the page on the server before it answers:"),
		doc.Ordered(
			doc.Item("It creates the routed component and mounts it with its children, calling `OnInit` and `OnPreRender` on each before its first `Render`."),
			doc.Item("It calls `OnNav` on every component that has it."),
			doc.Item("It runs every function passed to `ctx.Dispatch` and `ctx.Defer`, waits for every `ctx.Async` goroutine, and renders again, until nothing is left."),
			doc.Item("It writes the document. The page's HTML is the first child of `<body>`, followed by the loading box."),
		),
		doc.P("In the browser, `app.js` loads `app.wasm`, and `app.RunWhenOnBrowser` renders the page again and takes the body over. No component state travels from the server: data that `OnPreRender` loaded has to be loaded again in the browser, in `OnMount` or `OnNav`."),
		doc.H("Which hooks run where"),
		doc.List(
			doc.Item("`OnInit`: on both sides."),
			doc.Item("`OnPreRender`: on the server only, once per request."),
			doc.Item("`OnMount`: in the browser only."),
			doc.Item("`OnNav`: on both sides. In the browser it runs on the first load and on every navigation inside the app."),
		),
		doc.P("Anything else that only makes sense on one side can test the constants `app.IsServer` and `app.IsClient`."),
		doc.H("The generated document"),
		doc.P("For a component that renders `<h1>Hi</h1>` and a handler with only `Name` set, the response is, shortened:"),
		doc.HTML(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <meta name="author">
    <meta name="description">
    <meta content="#2d2c2c" name="theme-color">
    <meta content="width=device-width, initial-scale=1, viewport-fit=cover" name="viewport">
    <meta content="yes" name="mobile-web-app-capable">
    <meta content="https://" property="og:url">
    <meta property="og:title">
    <meta property="og:description">
    <meta content="website" property="og:type">
    <meta content="https://" property="og:image">
    <title></title>
    <link href="https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.svg" rel="icon">
    <link href="https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.png" rel="apple-touch-icon">
    <link href="/manifest.webmanifest" rel="manifest">
    <link href="/app.css" rel="stylesheet" type="text/css">
    <script>…</script>
    <script defer src="/wasm_exec.js"></script>
    <script defer src="/app.js"></script>
  </head>
  <body>
    <h1>Hi</h1>
    <aside class="goapp-app-info" id="app-wasm-loader">
      <img alt="wasm loader icon" class="goapp-logo goapp-spin" id="app-wasm-loader-icon" src="https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.png">
      <p class="goapp-label" id="app-wasm-loader-label">0%</p>
    </aside>
  </body>
</html>`),
		doc.List(
			doc.Item("Attributes are written in name order and escaped as HTML."),
			doc.Item("Stylesheet links come after `/app.css`: library stylesheets sorted by path, then `Styles` in order."),
			doc.Item("The inline `<script>` is the update checker, under ", doc.A("#versions-and-updates", "Versions and updates"), "."),
			doc.Item("`Scripts` follow `app.js`, and `RawHeaders` close the `<head>`."),
			doc.Item("Preload, preconnect, font, canonical and Twitter tags appear only when they are set."),
		),
		doc.H("Page metadata"),
		doc.Go(`func (ctx Context) Page() Page

type Page interface {
	Title() string
	SetTitle(v string)
	Lang() string
	SetLang(v string)
	Description() string
	SetDescription(v string)
	Author() string
	SetAuthor(v string)
	Keywords() string
	SetKeywords(v ...string)
	Image() string
	SetImage(v string)
	Preloads() []Preload
	SetPreloads(v ...Preload)
	SetLoadingLabel(v string)
	SetTwitterCard(v TwitterCard)
	SetCanonicalLink(v string)
	URL() *url.URL
	ReplaceURL(v *url.URL)
	Size() (w int, h int)
}`),
		doc.P("`SetTitle`, `SetDescription`, `SetAuthor`, `SetImage`, `SetLoadingLabel` and `SetCanonicalLink` also have an `f` variant that takes a format string: `SetTitlef(\"%s · Notes\", name)`."),
		doc.List(
			doc.Item("On the server, the setters change the document the handler writes. The handler's own fields are applied first, so a setter overrides them for that page only."),
			doc.Item("In the browser, `SetTitle`, `SetLang`, `SetDescription`, `SetAuthor`, `SetKeywords` and `SetImage` update the live document."),
			doc.Item("In the browser, `SetPreloads`, `SetLoadingLabel` and `SetCanonicalLink` do nothing, and `Preloads` returns nil. `SetTwitterCard` appends a new set of `<meta>` tags on every call."),
			doc.Item("`SetImage`, `SetCanonicalLink` and the card's `Image` are resolved through `Resources`, and the handler makes them absolute with `Domain`."),
			doc.Item("`URL` is the requested URL on the server and the window's URL in the browser. `ReplaceURL` changes the browser's address without a navigation; on the server it only changes what `URL` returns."),
			doc.Item("`Size` is the window's inner size in the browser and `0, 0` on the server."),
		),
		doc.P("Set the title and description in `OnNav`, which runs for the first load and for every navigation inside the app. Set the canonical link, preloads and the Twitter card in `OnPreRender`, because only the server's copy of them counts:"),
		doc.Go(`type article struct {
	app.Compo
	slug string
}

func (a *article) OnPreRender(ctx app.Context) {
	ctx.Page().SetCanonicalLink(ctx.Page().URL().Path)
	ctx.Page().SetPreloads(app.Preload{
		Type:          "image/png",
		As:            "image",
		Href:          "/web/hero.png",
		FetchPriority: "high",
	})
	ctx.Page().SetTwitterCard(app.TwitterCard{
		Card:     "summary_large_image",
		Site:     "notesapp",
		Image:    "/web/card.png",
		ImageAlt: "The Notes app on a laptop",
	})
}

func (a *article) OnNav(ctx app.Context) {
	a.slug = strings.TrimPrefix(ctx.Page().URL().Path, "/articles/")
	ctx.Page().SetTitlef("%s · Notes", a.slug)
	ctx.Page().SetDescription("Notes about " + a.slug)
}

func (a *article) Render() app.UI {
	return app.Article().Body(app.H1().Text(a.slug))
}`),
		doc.List(
			doc.Item("`TwitterCard.Card` is `summary`, `summary_large_image`, `app` or `player`. `Site` and `Creator` get a leading `@` when they lack one. Empty fields write no tag."),
			doc.Item("A `Preload` with an empty `Href` or `As` is skipped."),
			doc.Item("The title also sets `og:title`, the description `og:description`, and the image `og:image`."),
		),
		doc.P("There is no prerender cache. Each request renders the page again, so keep `OnPreRender` and the server side of `OnNav` fast."),
	}
}
