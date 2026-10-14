package applying

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	Title = "Applying"
	Route = "/applying"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "applying",
		Path:  Route,
		Title: Title,
		Lead:  "Deliver rules as static, generated and dynamic stylesheets, and drive them from actions.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Stylesheets", Body: stylesheets},
			{Title: "Handler and Build", Body: handler},
			{Title: "Run", Body: run},
			{Title: "Actions", Body: actions},
			{Title: "ViewModel and OnAction", Body: viewModel},
			{Title: "Errors", Body: errs},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("This page covers delivering rules to the browser: the three kinds of stylesheet, the `serve` handler that writes them, the engine that inserts dynamic rules, the actions that drive it, and the errors. The ", doc.A("/styling", "Styling"), " page covers writing rules."),
		doc.P("Every example on this page is live. Its bound rule is in this site's `Dynamic` style, and the controls change it through the same actions your app would fire. The CSS tab shows the text of the rules for the demo's starting state."),
	}
}

func stylesheets() []app.UI {
	return []app.UI{
		doc.Go(`type Style interface{ style() }

type Item interface {
	item()
	renderable
}

func Static(path string) Style
func Generated(path string, items ...Item) Style
func Dynamic(items ...Item) Style`),
		doc.P("`Style` has an unexported method: these three constructors are the only way to make one. The `serve` package turns a list of them into go-app's handler, under ", doc.A("#handler-and-build", "Handler and Build"), ". `Item` has unexported methods too. `*Rule`, `*KeyframesRule`, `*FontFaceRule`, `*PropertyRule`, `*ImportRule`, `*LayerRule` and `*LayerOrderRule` implement it, and each has `Err() error`. A nil `Style` in a handler's list, and a nil item passed to `Generated` or `Dynamic`, are ignored."),
		doc.P("The three kinds are chosen per stylesheet, and one handler can mix them:"),
		doc.List(
			doc.Item("`Static(path)` links a `.css` file that already exists.", doc.List(
				doc.Item("A path starting with `/web/` is looked up in `WebDir`. If the file is missing, `Build` returns `serve.ErrMissingStyle`."),
				doc.Item("A URL starting with `http://`, `https://` or `//` is linked as given and is not checked."),
				doc.Item("Any other path is `serve.ErrPathOutsideWeb`."),
				doc.Item("The content of the file is never read or validated."),
			)),
			doc.Item("`Generated(path, items...)` writes a `.css` file from the items. The path must start with `/web/` (otherwise `serve.ErrPathOutsideWeb`) and end in `.css` (otherwise `serve.ErrNotCSSPath`). `Build` renders the items and writes the text to `<WebDir>/…`, creating directories as needed, only if the bytes differ from what is already on disk. The file is then linked like a Static one. A Generated style with no items, or whose items have no properties, still writes an empty file."),
			doc.Item("`Dynamic(items...)` adds no link and writes no file. It takes `*Rule` and `*KeyframesRule` values that carry a binding. `Build` only validates them. They are inserted into and removed from the page in the browser, as described under ", doc.A("#run", "Run"), " and ", doc.A("#actions", "Actions"), "."),
		),
		doc.P("A Generated file writes every `@layer` order statement first, then every `@import`, then everything else, each group in slice order, with one blank line between items."),
		doc.P("Paths are cleaned before use, so `/web/./app.css` and `/web/x/../app.css` both mean `/web/app.css`, and the link written into the page is the cleaned path. A path that leaves `/web/` once cleaned, such as `/web/../secret.css`, is `serve.ErrPathOutsideWeb`. URLs are not cleaned."),
		doc.P("Use `Static` for CSS that has no typed form, such as a property the generator leaves out, and for a stylesheet hosted elsewhere."),
		doc.H("Resolve"),
		doc.Go(`type Kind int

const (
	StaticStyle Kind = iota + 1
	GeneratedStyle
	DynamicStyle
)

type Source struct {
	Kind    Kind
	Path    string
	Content []byte
}

func Resolve(s Style) (Source, error)`),
		doc.P("`Resolve` is what the handler calls on each style. It checks the items and returns the style's kind, its path as given, and for a Generated style the rendered bytes, which are empty but not nil when the style has no items. A Static style is not checked beyond that, and a Dynamic one has no path. The errors it returns are the item errors under ", doc.A("#errors", "Errors"), ", and on an error the `Source` is zero. A nil `Style` resolves to a zero `Source` with no error. You need `Resolve` only to deliver styles through something other than `serve.Handler`."),
		doc.H("A complete app"),
		doc.P("Two rules, delivered in a Generated file. The rules live in a package of their own, `theme`, which also holds the one function that lists the app's stylesheets:"),
		doc.Go(`package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var Card = style.Class("card").
	Set(css.Padding(css.Rem(1)), css.BorderRadius(css.Px(8)))

var Link = style.Class("nav-link").
	Set(css.Color(css.Var("--ink"))).
	Hover(css.TextDecoration(kw.Underline)).
	Media(media.Screen().And(media.MaxWidth(css.Px(600))), css.FontSize(css.Rem(0.9)))

func Styles() []style.Style {
	return []style.Style{
		style.Generated("/web/theme.css", Card, Link),
	}
}`),
		doc.P("The program that uses it, in a module named `example.com/hello`:"),
		doc.Go(`package main

import (
	"log"
	"net/http"

	"example.com/hello/theme"
	"github.com/Toms223/WebGo/serve"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type hello struct {
	app.Compo
}

func (h *hello) Render() app.UI {
	return app.Div().Class(theme.Card.Name()).Body(
		app.A().Class(theme.Link.Name()).Href("/").Text("Home"),
	)
}

func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	style.Run(theme.Styles()...)
	app.RunWhenOnBrowser()

	handler := &serve.Handler{
		Handler: app.Handler{Name: "Hello"},
		Styles:  theme.Styles(),
	}
	built, err := handler.Build()
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", built)
	log.Fatal(http.ListenAndServe(":8080", nil))
}`),
		doc.P("`Build` writes `web/theme.css` and puts `/web/theme.css` in the embedded `app.Handler`'s `Styles`. The file contains:"),
		doc.CSS(`.card {
	padding: 1rem;
	border-radius: 8px;
}

.nav-link {
	color: var(--ink);
}

.nav-link:hover {
	text-decoration: underline;
}

@media screen and (max-width: 600px) {
	.nav-link {
		font-size: 0.9rem;
	}
}`),
	}
}

func handler() []app.UI {
	return []app.UI{
		doc.Go(`package serve

type Handler struct {
	app.Handler
	Styles []style.Style
	WebDir string
}`),
		doc.P("The `serve` package, at `github.com/Toms223/WebGo/serve`, is the server side of styling: it holds the handler, `Build` and the path and file errors. `style` holds the rules and `Run`, and does not import `serve`."),
		doc.P("`serve.Handler` is go-app's `app.Handler` with one field changed: `Styles` is `[]style.Style` instead of `[]string`. The outer `Styles` shadows the embedded one, which is still reachable as `h.Handler.Styles`. `Handler` also has unexported fields, so a composite literal must name its fields."),
		doc.Go(`handler := &serve.Handler{
	Handler: app.Handler{Name: "Docs", Title: "Docs"},
	Styles: []style.Style{
		style.Static("/web/app.css"),
		style.Generated("/web/theme.css", theme.Card, theme.Link),
		style.Dynamic(theme.Compact),
	},
	WebDir: "web",
}`),
		doc.P("`theme.Compact` stands for any bound rule, such as the ones under ", doc.A("#actions", "Actions"), "."),
		doc.H("Build"),
		doc.Go(`func (h *Handler) Build() (*app.Handler, error)`),
		doc.P("`Build` resolves every style in the order of `Styles`:"),
		doc.List(
			doc.Item("It returns the first problem it finds as a `nil` handler and an error. It first checks every style, in the order of `Styles`: its items, through `style.Resolve`, and then its path. Only then does it look up the Static files, so a missing Static file is reported only when no style has an item or path error, wherever each sits in the list. Generated files are written last, so a failed check leaves the web directory as it was."),
			doc.Item("After any failure, `h.Handler.Styles` is left as it was."),
			doc.Item("On success it sets the embedded `Handler.Styles` to the paths of the Static and Generated styles, in order, followed by whatever `h.Handler.Styles` held when `Build` was first called. It returns `&h.Handler`. Dynamic styles contribute no path."),
			doc.Item("Calling `Build` again gives the same paths. Nothing is added twice."),
			doc.Item("`Build` never copies the handler. `app.Handler` carries a `sync.Once` and must not be copied, so the pointer it returns is the handler's own embedded value."),
			doc.Item("go-app still passes every path through the handler's `Resources` resolver. With `Resources: app.GitHubPages(\"WebGo\")`, the link to `/web/theme.css` is written as `/WebGo/web/theme.css`. Everything downstream, `http.Handle` and `app.GenerateStaticWebsite`, is unchanged."),
		),
		doc.P("`Build` has a pointer receiver. On a variable, `handler.Build()` works. A composite literal is not addressable, so `serve.Handler{…}.Build()` does not compile: write `(&serve.Handler{…}).Build()`, or keep the handler in a variable first:"),
		doc.Go(`built, err := (&serve.Handler{Styles: theme.Styles()}).Build()`),
		doc.P("`Build` is where generation happens, wherever the handler is built: in a build step, such as a command that calls `app.GenerateStaticWebsite`, or at the start of a server that runs go-app server-side. `app.GenerateStaticWebsite` does not copy `WebDir` into its output, so a static site publishes a generated stylesheet the same way it publishes `app.wasm`: by copying `WebDir` to `/web/`."),
		doc.H("WebDir"),
		doc.P("`WebDir` is the local directory go-app serves at `/web/`. `Build` uses it to find Static files and to write Generated ones: `/web/x/y.css` maps to `<WebDir>/x/y.css`. It defaults to `\"web\"`. It does not change where go-app serves files from, so keep the two in agreement. A relative `WebDir` is resolved against the working directory of the process that calls `Build`."),
		doc.H("Determinism"),
		doc.P("The text a Generated style writes depends only on its items: the same items give the same bytes, with no timestamp and no banner. Items are written in slice order, blocks in rule order and properties in call order. `Build` leaves a file alone when its bytes already match, so building twice does not touch the file's modification time. The writer's format is fixed:"),
		doc.List(
			doc.Item("Tabs for indentation, one blank line between blocks and a newline at the end."),
			doc.Item("A numeric value is written in its shortest round-trip form: `0.9rem`, not `.9rem`. A zero keeps its unit as given: `css.Px(0)` is `0px`."),
			doc.Item("A prefixed declaration comes immediately before the unprefixed one: `-webkit-backdrop-filter: blur(4px);` and then `backdrop-filter: blur(4px);`."),
		),
		doc.P("What is or is not committed to version control is up to you."),
	}
}
