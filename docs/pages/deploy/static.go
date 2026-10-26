package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func staticSites() []app.UI {
	return []app.UI{
		doc.P("A WebGo app does not need a Go server in production. The handler can write every page and support file once, and any static host can serve the result."),
		staticDiagram(),
		doc.Go(`func GenerateStaticWebsite(dir string, h *Handler, pages ...string) error`),
		doc.Ordered(
			doc.Item("It calls `h.Build()` and returns its error. Generated stylesheets are written to `WebDir` here."),
			doc.Item("It starts `h` on a local test server and requests `/`, `/app.js`, `/app-worker.js`, `/manifest.webmanifest`, `/wasm_exec.js`, `/app.css`, every path registered with `app.Route`, every path in `pages`, and the library stylesheets."),
			doc.Item("It writes each response under `dir`, which defaults to `.`. `/` becomes `index.html`, and a path with no extension gets `.html`: `/about` becomes `about.html`, and `/notes/first` becomes `notes/first.html`."),
			doc.Item("It creates an empty `dir/web` folder."),
		),
		doc.List(
			doc.Item("Paths registered with `app.RouteWithRegexp` cannot be listed. Pass the concrete paths you want as `pages`; a leading `/` is added when it is missing."),
			doc.Item("It does not check status codes. A path in `pages` that no route matches is written with the `app.NotFound` page."),
			doc.Item("It does not copy `WebDir`. Copy `app.wasm`, the generated stylesheets and every other `/web/` file into `dir/web` yourself."),
			doc.Item("It writes library stylesheets at their paths, under ", doc.A("#libraries", "Libraries"), ". It does not write proxied files such as `robots.txt`, under ", doc.A("#static-resources", "Static resources"), "."),
		),
		doc.P("One `main` can serve the app during development and generate the site for deployment:"),
		doc.Go(`package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/Toms223/WebGo/app"
)

type page struct {
	app.Compo
	text string
}

func (p *page) Render() app.UI {
	return app.H1().Text(p.text)
}

func main() {
	app.Route("/", func() app.Composer { return &page{text: "Home"} })
	app.Route("/about", func() app.Composer { return &page{text: "About"} })
	app.RunWhenOnBrowser()

	output := flag.String("static", "", "write a static site to this directory")
	repository := flag.String("github-pages", "", "repository name of a GitHub Pages project site")
	flag.Parse()

	handler := &app.Handler{Name: "Notes", Title: "Notes"}
	if *repository != "" {
		handler.Resources = app.GitHubPages(*repository)
	}

	if *output != "" {
		if err := app.GenerateStaticWebsite(*output, handler); err != nil {
			log.Fatal(err)
		}
		return
	}

	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}`),
		doc.Sh(`GOARCH=wasm GOOS=js go build -o web/app.wasm .
go run . -static dist
mkdir -p dist/web
cp -r web/. dist/web/`),
		doc.Plain(`dist
├── index.html
├── about.html
├── app.css
├── app.js
├── app-worker.js
├── manifest.webmanifest
├── wasm_exec.js
└── web
    └── app.wasm`),
		doc.H("GitHub Pages"),
		doc.P("A project site is served from `https://<user>.github.io/<repository>/`, not from the domain root. `Resources: app.GitHubPages(\"<repository>\")` writes every link, the manifest, the worker's file list and the app's own routing with that prefix. A user site (`<user>.github.io`) or a custom domain is served from the root and keeps the default `Resources`."),
		doc.P("This documentation site is deployed that way. Its workflow, run from the `docs` folder on every push to `main`:"),
		doc.Sh(`GOARCH=wasm GOOS=js go build -o web/app.wasm .
go run . static -o dist --github-pages "${{ github.event.repository.name }}"
mkdir -p dist/web
cp -r web/. dist/web/
cp dist/index.html dist/404.html
touch dist/.nojekyll`),
		doc.List(
			doc.Item("`static` is the site's own command around `app.GenerateStaticWebsite`, with the repository name passed to `app.GitHubPages`."),
			doc.Item("`cp -r web/. dist/web/` publishes `app.wasm`, the icons and the generated `theme.css`."),
			doc.Item("`404.html` is what GitHub Pages serves for a path with no file. As a copy of `index.html`, it loads the app, which then routes the URL in the browser, so regexp routes and paths that were not generated still work. The response status is still 404, and the prerendered HTML is the home page's."),
			doc.Item("`.nojekyll` makes GitHub Pages publish the folder as it is instead of running it through Jekyll."),
		),
		doc.P("The folder is then uploaded with `actions/upload-pages-artifact` and published with `actions/deploy-pages`."),
	}
}

func libraries() []app.UI {
	return []app.UI{
		doc.P("A library lets a Go package of components ship its own CSS, so an app that uses it only adds one value to its handler."),
		doc.Go(`Libraries []Library

type Library interface {
	Styles() (path, styles string)
}`),
		doc.Go(`type chartLibrary struct{}

func (chartLibrary) Styles() (string, string) {
	return "/charts.css", ".chart { display: grid; gap: 4px; }"
}

handler := &app.Handler{
	Name:      "Dashboard",
	Libraries: []app.Library{chartLibrary{}},
}`),
		doc.List(
			doc.Item("The handler serves `styles` as `text/css` at `path`, from memory. The file is not under `/web/` and is not checked by `Build`."),
			doc.Item("A library is ignored when `path` does not start with `/` or `styles` is empty. When two libraries use the same path, the last one wins."),
			doc.Item("Every page links the libraries after `/app.css`, sorted by path, and before `Styles`. The worker stores them, and their CSS is part of the version."),
			doc.Item("`app.GenerateStaticWebsite` does not write them, so a static site links files that do not exist."),
		),
		doc.P("In a WebGo app, `style.Generated` usually does this job better: a package exports its rules or a `style.Style`, and the app lists them in `Styles`, where `Build` checks them and static sites get the file. See ", doc.A("/applying#stylesheets", "Stylesheets"), "."),
	}
}
