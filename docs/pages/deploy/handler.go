package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func handler() []app.UI {
	return []app.UI{
		doc.P("Every field is optional. `Handler` has unexported fields, so a composite literal must name the fields it sets, and the handler must be used through a pointer. A typical production handler:"),
		doc.Go(`handler := &app.Handler{
	Name:        "Notes",
	ShortName:   "Notes",
	Title:       "Notes",
	Description: "A notebook that works offline",
	Author:      "Ana Silva",
	Keywords:    []string{"notes", "offline"},
	Image:       "/web/card.png",
	Domain:      "notes.example.com",
	Icon: app.Icon{
		Default:  "/web/icon-192.png",
		Large:    "/web/icon-512.png",
		SVG:      "/web/icon.svg",
		Maskable: "/web/icon-maskable.png",
	},
	BackgroundColor: "#ffffff",
	ThemeColor:      "#1a1a1a",
	Styles:          []style.Style{style.Static("/web/main.css")},
	Scripts:         []string{"/web/chart.js defer"},
	Fonts:           []string{"/web/inter.woff2"},
}
http.Handle("/", handler)`),
		doc.P("The handler fills in its defaults on the first request, or when `app.GenerateStaticWebsite` runs it. Changing a field after that has no effect."),
		doc.H("Page metadata"),
		doc.Go(`Name         string
ShortName    string
Title        string
Description  string
Author       string
Keywords     []string
Image        string
Domain       string
Lang         string
LoadingLabel string`),
		doc.List(
			doc.Item("`Name` and `ShortName` go into the manifest. Each defaults to the other. If `Name`, `ShortName` and `Title` are all empty, `Name` is `\"App PWA\"`."),
			doc.Item("`Title`, `Description`, `Author`, `Keywords` and `Image` are the defaults for every page's `<title>`, `<meta>` and Open Graph tags. A component overrides them per page through `ctx.Page()`, under ", doc.A("#prerendering-and-seo", "Prerendering and SEO"), ". The keywords tag is written only when `Keywords` is not empty."),
			doc.Item("`Domain` makes social URLs absolute: `og:url`, `og:image`, `twitter:image` and the canonical link become `https://<Domain><path>`. With no `Domain` they become `https://<path>`, which no crawler can use, so set it whenever pages are shared."),
			doc.Item("`Lang` is the `lang` attribute of `<html>`. It defaults to `\"en\"`."),
			doc.Item("`LoadingLabel` is the text under the icon while `app.wasm` downloads. `{progress}` is replaced with the percentage. It defaults to `\"{progress}%\"`, and the server writes it with `0`."),
		),
		doc.H("Icons and colours"),
		doc.Go(`Icon            Icon
BackgroundColor string
ThemeColor      string

type Icon struct {
	Default  string
	Large    string
	SVG      string
	Maskable string
}`),
		doc.List(
			doc.Item("`Default` is a 192×192 PNG. It is the loading image and a manifest icon. If it is empty, `Default`, `Large` and an empty `SVG` become go-app's logo."),
			doc.Item("`Large` is a 512×512 PNG for the manifest. An empty `Large` is left out of it."),
			doc.Item("`SVG` is the favicon (`<link rel=\"icon\">`) and the manifest's scalable icon. When it is empty, the favicon is `Default` and the manifest has no SVG icon."),
			doc.Item("`Maskable` is a 512×512 PNG with a safe zone for adaptive icons. It is also the `apple-touch-icon`. It falls back to `Default`."),
			doc.Item("`BackgroundColor` and `ThemeColor` go into the manifest, and `ThemeColor` also into `<meta name=\"theme-color\">`. Both default to `#2d2c2c`."),
		),
		doc.H("Head resources"),
		doc.Go(`Styles             []Stylesheet
Scripts            []string
Fonts              []string
Preconnect         []string
CacheableResources []string
RawHeaders         []string`),
		doc.List(
			doc.Item("`Styles` is covered under ", doc.A("/applying#handler-and-build", "Handler and Build"), "."),
			doc.Item("`Scripts` become `<script src>` tags after `app.js`, in order. An entry may end with `defer`, `async`, `crossorigin` or `crossorigin=<value>` after a space: `\"/web/chart.js defer\"`."),
			doc.Item("`Fonts` become `<link rel=\"preload\" as=\"font\">` tags whose `type` is `font/` plus the file extension."),
			doc.Item("`Preconnect` entries become `<link rel=\"preconnect\">` tags, one per origin."),
			doc.Item("`CacheableResources` are extra files the service worker stores for offline use. Icons, fonts, stylesheets, scripts and `app.wasm` are stored already."),
			doc.Item("`RawHeaders` are written as they are, unescaped, at the end of `<head>`. Use them for third-party snippets such as analytics tags."),
		),
		doc.H("Document"),
		doc.Go(`HTML func() HTMLHtml
Body func() HTMLBody`),
		doc.P("These build the `<html>` and `<body>` elements. They default to `app.Html` and `app.Body`. Use them to add attributes; the handler sets `lang` and the children itself:"),
		doc.Go(`handler := &app.Handler{
	Name: "Notes",
	HTML: func() app.HTMLHtml { return app.Html().DataSet("theme", "dark") },
}`),
		doc.H("Runtime"),
		doc.Go(`StartURL                string
Env                     map[string]string
InternalURLs            []string
ProxyResources          []ProxyResource
Resources               ResourceResolver
WasmContentLength       string
WasmContentLengthHeader string
ServiceWorkerTemplate   string`),
		doc.List(
			doc.Item("`StartURL` is the manifest's `start_url`, the page an installed app opens on. It defaults to `\"/\"`."),
			doc.Item("`Env` is a set of variables for the app. `app.Getenv(k)` reads them in the browser, and on the server, where the handler also sets them in the process environment when it initialises. They are written into `app.js`, so never put secrets in `Env`. The handler overwrites four keys of its own: `GOAPP_VERSION`, `GOAPP_STATIC_RESOURCES_URL`, `GOAPP_ROOT_PREFIX` and `GOAPP_INTERNAL_URLS`."),
			doc.Item("`InternalURLs` are URL prefixes that open with a full page load in the same tab. Without them, a link to another host opens in a new tab, and a link to the same host is routed inside the app."),
			doc.Item("`ProxyResources` and `Resources` are covered under ", doc.A("#static-resources", "Static resources"), "."),
			doc.Item("`WasmContentLength` is the size of `app.wasm` in bytes, as a string. The loader uses it for the progress label when the response has no usable length, for example when it is compressed. Otherwise the loader reads the header named by `WasmContentLengthHeader`, then `Content-Length`."),
			doc.Item("`ServiceWorkerTemplate` replaces the service worker. It defaults to `app.DefaultAppWorkerJS`. A custom template loses WebGo's update handling, under ", doc.A("#versions-and-updates", "Versions and updates"), ", unless it copies it."),
		),
		doc.H("WebGo fields"),
		doc.Go(`WebDir            string
DisableAutoUpdate bool
Version           string

func (h *Handler) Build() error`),
		doc.P("`WebDir` and `Build` are covered under ", doc.A("/applying#handler-and-build", "Handler and Build"), ". `Version` and `DisableAutoUpdate` are covered under ", doc.A("#versions-and-updates", "Versions and updates"), "."),
		doc.H("What it serves"),
		doc.List(
			doc.Item("`/app.js`, `/app-worker.js`, `/manifest.webmanifest`, `/wasm_exec.js` and `/app.css`, generated from the fields above. `/goapp.js` and `/manifest.json` are aliases."),
			doc.Item("`/web/...` from disk, when `Resources` is a local directory. `/app.wasm` is an alias of `/web/app.wasm`."),
			doc.Item("`/robots.txt`, `/sitemap.xml`, `/ads.txt` and every `ProxyResources` path."),
			doc.Item("Library stylesheets, under ", doc.A("#libraries", "Libraries"), "."),
			doc.Item("Any other path is a page. A path no `app.Route` or `app.RouteWithRegexp` matches gets a plain 404."),
		),
		doc.P("Every response has `Cache-Control: no-cache`. Everything except `/web/` files and proxied files also has the version as its `ETag`, and a request that sends it back gets `304 Not Modified`. `/web/` files are validated by their modification time instead."),
	}
}
