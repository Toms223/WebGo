package deploy

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func resources() []app.UI {
	return []app.UI{
		doc.P("Files the app links to (images, icons, stylesheets, scripts and `app.wasm` itself) live in a folder named `web`, next to where the server runs. Code always names them by a path that starts with `/web/`, wherever they are actually hosted:"),
		doc.Go(`app.Img().Src("/web/logo.png").Alt("Logo")`),
		doc.P("The `src`, `href`, `srcset`, `cite` and `data` attributes are resolved through the handler's `Resources`, on the server and in the browser, so the same element works with every resolver below. Paths outside `/web/` and `http://` or `https://` URLs are left as they are."),
		doc.H("Resources"),
		doc.Go(`Resources ResourceResolver

type ResourceResolver interface {
	Resolve(string) string
}

func LocalDir(directory string) ResourceResolver
func RemoteBucket(url string) ResourceResolver
func PrefixedLocation(prefix string) ResourceResolver
func GitHubPages(repositoryName string) ResourceResolver`),
		doc.P("`Resolve` maps a `/web/` path to the URL the page uses. The handler also asks it where the page root is, which matters for the manifest, the service worker and the static generator."),
		doc.List(
			doc.Item("`LocalDir(\"\")` is the default. The handler serves `./web` at `/web/`. Given a directory, it serves `<directory>/web` at `/web/`, and links stay `/web/...`: `LocalDir(\"site\")` serves `site/web/logo.png` as `/web/logo.png`. `WebDir` follows this setting when it is empty."),
			doc.Item("`RemoteBucket(url)` turns `/web/logo.png` into `<url>/web/logo.png`. The handler stops serving `/web/` itself, so upload the whole `web` folder, `app.wasm` included. The bucket must allow cross-origin requests from the app's origin. The service worker stores only same-origin files, so bucket files are not available offline."),
			doc.Item("`PrefixedLocation(prefix)` puts `prefix` in front of every `/web/` path and of the page root: with `\"/assets\"`, `/web/main.css` becomes `/assets/web/main.css`. `GitHubPages(repositoryName)` is `PrefixedLocation(\"/\" + repositoryName)`."),
		),
		doc.Note("`PrefixedLocation` and `GitHubPages` are for sites that something else serves under a sub-path, such as a static site on GitHub Pages. A handler that uses them still serves `/web/` from `./web`, so it also runs behind a proxy that strips the prefix. For GitHub Pages, use them with `app.GenerateStaticWebsite`, under ", doc.A("#static-sites-and-github-pages", "Static sites and GitHub Pages"), "."),
		doc.H("ResolveStaticResource"),
		doc.Go(`func (ctx Context) ResolveStaticResource(v string) string`),
		doc.P("`ResolveStaticResource` applies the same rule outside the attributes above: in a `url()` inside an inline style, a `fetch`, or a value passed to JavaScript. It returns `v` unchanged when it is not a `/web/` path."),
		doc.Go(`type banner struct {
	app.Compo
	background string
}

func (b *banner) OnNav(ctx app.Context) {
	b.background = ctx.ResolveStaticResource("/web/banner.jpg")
}

func (b *banner) Render() app.UI {
	return app.Div().
		RawStyle("background-image", "url("+b.background+")").
		Body(app.Img().Src("/web/logo.png").Alt("Logo"))
}`),
		doc.P("`OnNav` runs on the server and in the browser, so the prerendered HTML has the resolved URL too."),
		doc.H("ProxyResources"),
		doc.Go(`ProxyResources []ProxyResource

type ProxyResource struct {
	Path         string
	ResourcePath string
}`),
		doc.P("A `ProxyResource` serves the `/web/` file `ResourcePath` at `Path`, for files that must sit at the site root:"),
		doc.Go(`handler := &app.Handler{
	Name: "Notes",
	ProxyResources: []app.ProxyResource{
		{Path: "/.well-known/security.txt", ResourcePath: "/web/security.txt"},
	},
}`),
		doc.List(
			doc.Item("`/robots.txt`, `/sitemap.xml` and `/ads.txt` are proxied to the file of the same name under `/web/` by default. An entry with the same `Path` replaces the default."),
			doc.Item("An entry is ignored when `Path` does not start with `/`, when `ResourcePath` does not start with `/web/`, or when `Path` is one of the handler's own paths, such as `/`, `/app.js` or `/manifest.webmanifest`."),
			doc.Item("When the file does not exist, the response is a 404."),
			doc.Item("`app.GenerateStaticWebsite` does not write proxied files. A static site serves them from wherever you copy them."),
		),
	}
}
