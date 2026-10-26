package app

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/Toms223/WebGo/internal/errors"
)

const (
	defaultThemeColor = "#2d2c2c"
)

// Handler is an http.Handler that serves the app: server-rendered pages, the
// files under /web/, app.wasm and the PWA resources. The zero value works.
// Fields are read on the first request, or by Build, so set them before
// serving.
type Handler struct {
	Name string

	ShortName string

	// Icon is the favicon, PWA and loading-screen icon. The favicon is Icon.SVG,
	// or Icon.Default when SVG is empty; an empty Default uses go-app's logo.
	Icon Icon

	BackgroundColor string

	ThemeColor string

	// LoadingLabel is shown while app.wasm loads; {progress} becomes the
	// percentage. The default is "{progress}%".
	LoadingLabel string

	Lang string

	// Libraries are component libraries whose stylesheets are served and linked
	// on every page.
	Libraries []Library

	Title string

	Description string

	Domain string

	Author string

	Keywords []string

	Image string

	// Styles are linked in <head> in slice order, after app.css and the
	// Libraries. Build checks static files and writes generated ones; the slice
	// itself is never modified.
	Styles []Stylesheet

	// WebDir is the directory on disk behind /web/, used by Build and the
	// default Version. When empty it is dir/web if Resources is LocalDir(dir),
	// and "web" otherwise.
	WebDir string

	// DisableAutoUpdate stops the page reloading itself after a new version is
	// installed. Update checks still run, and AppUpdater components are still
	// notified.
	DisableAutoUpdate bool

	Fonts []string

	// Scripts are JS files loaded in <head>. A path may be followed by a space
	// and async, defer or crossorigin to set that attribute.
	Scripts []string

	// CacheableResources are extra files precached for offline use. Icon,
	// Styles, Libraries, Scripts and Fonts are precached already.
	CacheableResources []string

	// RawHeaders are inserted into <head> as raw HTML.
	RawHeaders []string

	// HTML creates the <html> element of each page, for example to set
	// attributes on it. The default is app.Html.
	HTML func() HTMLHtml

	// Body creates the <body> element of each page. The default is app.Body.
	Body func() HTMLBody

	// Env holds variables that Getenv reads on both sides. They are embedded in
	// app.js, so never put secrets here. GOAPP_VERSION,
	// GOAPP_STATIC_RESOURCES_URL, GOAPP_ROOT_PREFIX and GOAPP_INTERNAL_URLS are
	// reserved.
	Env map[string]string

	// InternalURLs are URL prefixes that load in the current tab with a full
	// page load, instead of being routed in the app or opened in a new tab.
	InternalURLs []string

	Preconnect []string

	// ProxyResources serve /web/ files at other paths. /robots.txt,
	// /sitemap.xml and /ads.txt map to /web/ by default.
	ProxyResources []ProxyResource

	// Resources resolves /web/ paths. The default is LocalDir("").
	Resources ResourceResolver

	StartURL string

	// Version identifies a release; when it changes, installed clients update.
	// When empty it is a hash of the settings, embedded assets and the local
	// /web/ files the page uses, so it only changes with the content.
	Version string

	// WasmContentLength is the uncompressed size of app.wasm in bytes, for the
	// loading progress when the response has no usable Content-Length.
	WasmContentLength string

	// WasmContentLengthHeader names the response header that carries the size
	// of app.wasm. The default is Content-Length.
	WasmContentLengthHeader string

	// ServiceWorkerTemplate replaces DefaultAppWorkerJS as the app-worker.js
	// template. The automatic reload relies on the default worker's install and
	// activate behaviour.
	ServiceWorkerTemplate string

	once                 sync.Once
	etag                 string
	libraries            map[string][]byte
	proxyResources       map[string]ProxyResource
	cachedProxyResources *memoryCache
	cachedPWAResources   *memoryCache
	initErr              error
	styleLinks           []string
	libraryLinks         []string
	built                bool
}

func (h *Handler) init() {
	if !h.built {
		if err := h.Build(); err != nil {
			h.initErr = err
			return
		}
	}
	h.initVersion()
	h.initStaticResources()
	h.initLibraries()
	h.initLinks()
	h.initServiceWorker()
	h.initIcon()
	h.initPWA()
	h.initPageContent()
	h.initPWAResources()
	h.initProxyResources()
}

func (h *Handler) initVersion() {
	if h.Version == "" {
		h.Version = h.contentVersion()
	}
	h.etag = `"` + h.Version + `"`
}

func (h *Handler) initStaticResources() {
	if h.Resources == nil {
		h.Resources = LocalDir("")
	}
}

func (h *Handler) initLibraries() {
	libs := make(map[string][]byte)
	for _, l := range h.Libraries {
		path, styles := l.Styles()
		if !strings.HasPrefix(path, "/") || len(styles) == 0 {
			continue
		}
		libs[path] = []byte(styles)
	}
	h.libraries = libs
}

func (h *Handler) initLinks() {
	h.libraryLinks = make([]string, 0, len(h.libraries))
	for path := range h.libraries {
		h.libraryLinks = append(h.libraryLinks, path)
	}
	slices.Sort(h.libraryLinks)
}

func (h *Handler) updaterScript() string {
	return strings.ReplaceAll(updaterJS, "{{.AutoUpdate}}", strconv.FormatBool(!h.DisableAutoUpdate))
}

func (h *Handler) headLinks() []string {
	return slices.Concat([]string{"/app.css"}, h.libraryLinks, h.styleLinks)
}

func (h *Handler) initServiceWorker() {
	if h.ServiceWorkerTemplate == "" {
		h.ServiceWorkerTemplate = DefaultAppWorkerJS
	}
}

func (h *Handler) initIcon() {
	if h.Icon.Default == "" {
		h.Icon.Default = "https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.png"
		h.Icon.Large = "https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.png"
		if h.Icon.SVG == "" {
			h.Icon.SVG = "https://raw.githubusercontent.com/maxence-charriere/go-app/master/docs/web/icon.svg"
		}
	}

	if h.Icon.Maskable == "" {
		h.Icon.Maskable = h.Icon.Default
	}
}

func (h *Handler) initPWA() {
	if h.Name == "" && h.ShortName == "" && h.Title == "" {
		h.Name = "App PWA"
	}
	if h.ShortName == "" {
		h.ShortName = h.Name
	}
	if h.Name == "" {
		h.Name = h.ShortName
	}

	if h.BackgroundColor == "" {
		h.BackgroundColor = defaultThemeColor
	}
	if h.ThemeColor == "" {
		h.ThemeColor = defaultThemeColor
	}

	if h.Lang == "" {
		h.Lang = "en"
	}

	if h.LoadingLabel == "" {
		h.LoadingLabel = "{progress}%"
	}

	if h.StartURL == "" {
		h.StartURL = "/"
	}
}

func (h *Handler) initPageContent() {
	if h.HTML == nil {
		h.HTML = Html
	}

	if h.Body == nil {
		h.Body = Body
	}

}

func (h *Handler) initPWAResources() {
	h.cachedPWAResources = newMemoryCache(5)

	h.cachedPWAResources.Set(cacheItem{
		Path:        "/wasm_exec.js",
		ContentType: "application/javascript",
		Body:        []byte(wasmExecJS()),
	})

	h.cachedPWAResources.Set(cacheItem{
		Path:        "/app.js",
		ContentType: "application/javascript",
		Body:        h.makeAppJS(),
	})

	h.cachedPWAResources.Set(cacheItem{
		Path:        "/app-worker.js",
		ContentType: "application/javascript",
		Body:        h.makeAppWorkerJS(),
	})

	h.cachedPWAResources.Set(cacheItem{
		Path:        "/manifest.webmanifest",
		ContentType: "application/manifest+json",
		Body:        h.makeManifestJSON(),
	})

	h.cachedPWAResources.Set(cacheItem{
		Path:        "/app.css",
		ContentType: "text/css",
		Body:        []byte(appCSS),
	})
}

func (h *Handler) makeAppJS() []byte {
	if h.Env == nil {
		h.Env = make(map[string]string)
	}
	internalURLs, _ := json.Marshal(h.InternalURLs)
	h.Env["GOAPP_INTERNAL_URLS"] = string(internalURLs)
	h.Env["GOAPP_VERSION"] = h.Version
	h.Env["GOAPP_STATIC_RESOURCES_URL"] = h.Resources.Resolve("/web")
	h.Env["GOAPP_ROOT_PREFIX"] = h.Resources.Resolve("/")

	for k, v := range h.Env {
		if err := os.Setenv(k, v); err != nil {
			Log(errors.New("setting app env variable failed").
				WithTag("name", k).
				WithTag("value", v).
				Wrap(err))
		}
	}

	s := appJS
	s = strings.ReplaceAll(s, "{{.Env}}", jsonString(h.Env))
	s = strings.ReplaceAll(s, "{{.LoadingLabel}}", h.LoadingLabel)
	s = strings.ReplaceAll(s, "{{.Wasm}}", h.Resources.Resolve("/web/app.wasm"))
	s = strings.ReplaceAll(s, "{{.WasmContentLength}}", h.WasmContentLength)
	s = strings.ReplaceAll(s, "{{.WasmContentLengthHeader}}", h.WasmContentLengthHeader)
	s = strings.ReplaceAll(s, "{{.WorkerJS}}", h.Resources.Resolve("/app-worker.js"))
	return []byte(s)
}

func (h *Handler) makeAppWorkerJS() []byte {
	resources := make(map[string]struct{})
	setResources := func(res ...string) {
		for _, r := range res {
			if resource := parseHTTPResource(r); resource.URL != "" {
				resources[resource.URL] = struct{}{}
			}
		}
	}
	setResources(
		"/app.css",
		"/app.js",
		"/manifest.webmanifest",
		"/wasm_exec.js",
		"/web/app.wasm",
		"/",
	)
	setResources(h.Fonts...)
	setResources(h.Icon.Default, h.Icon.Large, h.Icon.Maskable)
	setResources(h.headLinks()...)
	setResources(h.Scripts...)
	setResources(h.CacheableResources...)

	resourcesTocache := make([]string, 0, len(resources))
	for k := range resources {
		resolved := h.Resources.Resolve(k)
		if k == "/" && !strings.HasSuffix(resolved, "/") {
			resolved += "/"
		}
		resourcesTocache = append(resourcesTocache, resolved)
	}
	sort.Slice(resourcesTocache, func(a, b int) bool {
		return strings.Compare(resourcesTocache[a], resourcesTocache[b]) > 0
	})

	s := h.ServiceWorkerTemplate
	s = strings.ReplaceAll(s, "{{.Version}}", h.Version)
	s = strings.ReplaceAll(s, "{{.ResourcesToCache}}", jsonString(resourcesTocache))
	return []byte(s)
}

func (h *Handler) makeManifestJSON() []byte {
	type manifestIcon struct {
		Src     string `json:"src"`
		Type    string `json:"type"`
		Purpose string `json:"purpose"`
		Sizes   string `json:"sizes"`
	}

	var icons []manifestIcon
	if h.Icon.Maskable != "" {
		iconURL := h.Resources.Resolve(h.Icon.Maskable)
		icons = append(icons, manifestIcon{
			Src:     iconURL,
			Type:    "image/png",
			Purpose: "maskable",
			Sizes:   "512x512",
		})
	}
	if h.Icon.SVG != "" {
		iconURL := h.Resources.Resolve(h.Icon.SVG)
		icons = append(icons, manifestIcon{
			Src:   iconURL,
			Type:  "image/svg+xml",
			Sizes: "any",
		})
	}
	if h.Icon.Large != "" {
		iconURL := h.Resources.Resolve(h.Icon.Large)
		icons = append(icons, manifestIcon{
			Src:   iconURL,
			Type:  "image/png",
			Sizes: "512x512",
		})
	}
	icons = append(icons, manifestIcon{
		Src:   h.Resources.Resolve(h.Icon.Default),
		Type:  "image/png",
		Sizes: "192x192",
	})

	manifest := struct {
		ShortName       string         `json:"short_name"`
		Name            string         `json:"name"`
		Description     string         `json:"description"`
		Scope           string         `json:"scope"`
		StartURL        string         `json:"start_url"`
		BackgroundColor string         `json:"background_color"`
		ThemeColor      string         `json:"theme_color"`
		Display         string         `json:"display"`
		Icons           []manifestIcon `json:"icons"`
	}{
		ShortName:       h.ShortName,
		Name:            h.Name,
		Description:     h.Description,
		Scope:           "/",
		StartURL:        h.Resources.Resolve(h.StartURL),
		BackgroundColor: h.BackgroundColor,
		ThemeColor:      h.ThemeColor,
		Display:         "standalone",
		Icons:           icons,
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		panic(errors.New("encoding manifest failed").Wrap(err))
	}
	return manifestJSON
}

func (h *Handler) initProxyResources() {
	h.cachedProxyResources = newMemoryCache(len(h.ProxyResources))
	resources := make(map[string]ProxyResource)

	for _, r := range h.ProxyResources {
		switch r.Path {
		case "/wasm_exec.js",
			"/goapp.js",
			"/app.js",
			"/app-worker.js",
			"/manifest.json",
			"/manifest.webmanifest",
			"/app.css",
			"/app.wasm",
			"/goapp.wasm",
			"/":
			continue

		default:
			if strings.HasPrefix(r.Path, "/") && strings.HasPrefix(r.ResourcePath, "/web/") {
				resources[r.Path] = r
			}
		}
	}

	if _, ok := resources["/robots.txt"]; !ok {
		resources["/robots.txt"] = ProxyResource{
			Path:         "/robots.txt",
			ResourcePath: "/web/robots.txt",
		}
	}
	if _, ok := resources["/sitemap.xml"]; !ok {
		resources["/sitemap.xml"] = ProxyResource{
			Path:         "/sitemap.xml",
			ResourcePath: "/web/sitemap.xml",
		}
	}
	if _, ok := resources["/ads.txt"]; !ok {
		resources["/ads.txt"] = ProxyResource{
			Path:         "/ads.txt",
			ResourcePath: "/web/ads.txt",
		}
	}

	h.proxyResources = resources
}

// ServeHTTP serves the app. The first call runs Build, unless it already
// succeeded, and fills in the defaults; if Build fails there, every request is
// answered with 500 and the error is logged. The Version ETag applies to pages
// and PWA files, not to /web/ files.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.once.Do(h.init)
	if h.initErr != nil {
		Log(h.initErr)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Cache-Control", "no-cache")

	path := r.URL.Path
	if strings.HasPrefix(path, "/"+h.Version+"/") {
		path = strings.TrimPrefix(path, "/"+h.Version)
	}

	fileHandler, isServingStaticResources := h.Resources.(http.Handler)
	if isServingStaticResources && strings.HasPrefix(path, "/web/") {
		fileHandler.ServeHTTP(w, r)
		return
	}

	if proxyResource, ok := h.proxyResources[path]; ok {
		h.serveProxyResource(proxyResource, w, r)
		return
	}

	w.Header().Set("ETag", h.etag)
	if r.Header.Get("If-None-Match") == h.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	switch path {
	case "/goapp.js":
		path = "/app.js"

	case "/manifest.json":
		path = "/manifest.webmanifest"

	case "/app.wasm", "/goapp.wasm":
		if isServingStaticResources {
			r2 := *r
			r2.URL.Path = h.Resources.Resolve("/web/app.wasm")
			fileHandler.ServeHTTP(w, &r2)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		return

	}

	if res, ok := h.cachedPWAResources.Get(path); ok {
		h.serveCachedItem(w, res)
		return
	}

	if library, ok := h.libraries[path]; ok {
		h.serveLibrary(w, r, library)
		return
	}

	h.servePage(w, r)
}

func (h *Handler) serveCachedItem(w http.ResponseWriter, i cacheItem) {
	w.Header().Set("Content-Length", strconv.Itoa(i.Len()))
	w.Header().Set("Content-Type", i.ContentType)

	if i.ContentEncoding != "" {
		w.Header().Set("Content-Encoding", i.ContentEncoding)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(i.Body)
}

func (h *Handler) serveProxyResource(resource ProxyResource, w http.ResponseWriter, r *http.Request) {
	var u string
	if _, ok := h.Resources.(http.Handler); ok {
		var protocol string
		if r.TLS != nil {
			protocol = "https://"
		} else {
			protocol = "http://"
		}
		u = protocol + r.Host + resource.ResourcePath
	} else {
		u = h.Resources.Resolve(resource.ResourcePath)
	}

	if i, ok := h.cachedProxyResources.Get(resource.Path); ok {
		h.serveCachedItem(w, i)
		return
	}

	res, err := http.Get(u)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Log(errors.New("getting proxy static resource failed").
			WithTag("url", u).
			WithTag("proxy-path", resource.Path).
			WithTag("static-resource-path", resource.ResourcePath).
			Wrap(err),
		)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		Log(errors.New("reading proxy static resource failed").
			WithTag("url", u).
			WithTag("proxy-path", resource.Path).
			WithTag("static-resource-path", resource.ResourcePath).
			Wrap(err),
		)
		return
	}

	item := cacheItem{
		Path:            resource.Path,
		ContentType:     res.Header.Get("Content-Type"),
		ContentEncoding: res.Header.Get("Content-Encoding"),
		Body:            body,
	}
	h.cachedProxyResources.Set(item)
	h.serveCachedItem(w, item)
}

func (h *Handler) servePage(w http.ResponseWriter, r *http.Request) {
	status := http.StatusOK
	if !routes.routed(r.URL.Path) {
		status = http.StatusNotFound
	}

	ctx := context.Background()

	origin := *r.URL
	origin.Scheme = "http"

	page := makeRequestPage(&origin, h.Resources.Resolve)
	page.SetTitle(h.Title)
	links := h.headLinks()
	page.SetLang(h.Lang)
	page.SetDescription(h.Description)
	page.SetAuthor(h.Author)
	page.SetKeywords(h.Keywords...)
	page.SetLoadingLabel(strings.ReplaceAll(h.LoadingLabel, "{progress}", "0"))
	page.SetImage(h.Image)

	engine := newEngine(ctx,
		&routes,
		h.Resources.Resolve,
		&page,
		actionHandlers,
	)
	engine.Navigate(page.URL(), false)
	engine.ConsumeAll()

	icon := h.Icon.SVG
	if icon == "" {
		icon = h.Icon.Default
	}

	var b bytes.Buffer
	err := engine.Encode(&b, h.HTML().
		Lang(page.Lang()).
		privateBody(
			Head().Body(
				Meta().Charset("UTF-8"),
				Meta().
					Name("author").
					Content(page.Author()),
				Meta().
					Name("description").
					Content(page.Description()),
				If(page.Keywords() != "", func() UI {
					return Meta().
						Name("keywords").
						Content(page.Keywords())
				}),
				Meta().
					Name("theme-color").
					Content(h.ThemeColor),
				Meta().
					Name("viewport").
					Content("width=device-width, initial-scale=1, viewport-fit=cover"),
				Meta().
					Name("mobile-web-app-capable").
					Content("yes"),
				Meta().
					Property("og:url").
					Content(resolveOGResource(h.Domain, h.Resources.Resolve(page.URL().Path))),
				Meta().
					Property("og:title").
					Content(page.Title()),
				Meta().
					Property("og:description").
					Content(page.Description()),
				Meta().
					Property("og:type").
					Content("website"),
				Meta().
					Property("og:image").
					Content(resolveOGResource(h.Domain, page.Image())),
				Range(page.twitterCardMap).Map(func(k string) UI {
					v := page.twitterCardMap[k]
					if v == "" {
						return nil
					}
					if k == "twitter:image" {
						v = resolveOGResource(h.Domain, v)
					}
					return Meta().
						Name(k).
						Content(v)
				}),
				Title().Text(page.Title()),
				If(page.canonicalLink != "", func() UI {
					return Link().
						Rel("canonical").
						Href(resolveOGResource(h.Domain, page.canonicalLink))
				}),
				Range(h.Preconnect).Slice(func(i int) UI {
					if resource := parseHTTPResource(h.Preconnect[i]); resource.URL != "" {
						return resource.toLink().Rel("preconnect")
					}
					return nil
				}),
				Range(h.Fonts).Slice(func(i int) UI {
					if resource := parseHTTPResource(h.Fonts[i]); resource.URL != "" {
						return resource.toLink().
							Type("font/" + strings.TrimPrefix(filepath.Ext(resource.URL), ".")).
							Rel("preload").
							As("font")
					}
					return nil
				}),
				Range(page.Preloads()).Slice(func(i int) UI {
					p := page.Preloads()[i]
					if p.Href == "" || p.As == "" {
						return nil
					}

					if resource := parseHTTPResource(p.Href); resource.URL != "" {
						return resource.toLink().
							Type(p.Type).
							Rel("preload").
							As(p.As).
							FetchPriority(p.FetchPriority)
					}
					return nil
				}),
				Link().
					Rel("icon").
					Href(icon),
				Link().
					Rel("apple-touch-icon").
					Href(h.Icon.Maskable),
				Link().
					Rel("manifest").
					Href("/manifest.webmanifest"),
				Range(links).Slice(func(i int) UI {
					if resource := parseHTTPResource(links[i]); resource.URL != "" {
						return resource.toLink().
							Type("text/css").
							Rel("stylesheet")
					}
					return nil
				}),
				Raw("<script>"+h.updaterScript()+"</script>"),
				Script().
					Defer(true).
					Src("/wasm_exec.js"),
				Script().
					Defer(true).
					Src("/app.js"),
				Range(h.Scripts).Slice(func(i int) UI {
					if resource := parseHTTPResource(h.Scripts[i]); resource.URL != "" {
						return resource.toScript()
					}
					return nil

				}),
				Range(h.RawHeaders).Slice(func(i int) UI {
					return Raw(h.RawHeaders[i])
				}),
			),
			h.Body().privateBody(
				Aside().
					ID("app-wasm-loader").
					Class("goapp-app-info").
					Body(
						Img().
							ID("app-wasm-loader-icon").
							Class("goapp-logo goapp-spin").
							Alt("wasm loader icon").
							Src(h.Icon.Default),
						P().
							ID("app-wasm-loader-label").
							Class("goapp-label").
							Text(page.loadingLabel),
					),
			),
		))
	if err != nil {
		Log(errors.New("encoding html document failed").Wrap(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	w.Write(b.Bytes())
}

func (h *Handler) serveLibrary(w http.ResponseWriter, r *http.Request, library []byte) {
	w.Header().Set("Content-Length", strconv.Itoa(len(library)))
	w.Header().Set("Content-Type", "text/css")
	w.Write(library)
}

type Icon struct {
	Default string

	Large string

	SVG string

	Maskable string
}

func isRemoteLocation(path string) bool {
	return strings.HasPrefix(path, "https://") ||
		strings.HasPrefix(path, "http://")
}

func isStaticResourcePath(path string) bool {
	return strings.HasPrefix(path, "/web/") ||
		strings.HasPrefix(path, "web/")
}

type httpResource struct {
	URL         string
	LoadingMode string
	CrossOrigin string
}

func (r httpResource) toLink() HTMLLink {
	link := Link().Href(r.URL)
	if r.CrossOrigin != "" {
		link = link.CrossOrigin(strings.Trim(r.CrossOrigin, "true"))
	}
	return link
}

func (r httpResource) toScript() HTMLScript {
	script := Script().Src(r.URL)
	if r.CrossOrigin != "" {
		script = script.CrossOrigin(strings.Trim(r.CrossOrigin, "true"))
	}

	switch r.LoadingMode {
	case "defer":
		script = script.Defer(true)

	case "async":
		script = script.Async(true)
	}
	return script
}

func parseHTTPResource(v string) httpResource {
	var res httpResource
	for _, part := range strings.Split(v, " ") {
		if part = strings.TrimSpace(part); part == "" {
			continue
		}

		switch normalizedPart := strings.ToLower(part); {
		case normalizedPart == "crossorigin":
			res.CrossOrigin = "true"

		case strings.HasPrefix(normalizedPart, "crossorigin="):
			res.CrossOrigin = strings.TrimPrefix(normalizedPart, "crossorigin=")

		case normalizedPart == "defer",
			normalizedPart == "async":
			res.LoadingMode = normalizedPart

		default:
			res.URL = part
		}
	}
	return res
}
