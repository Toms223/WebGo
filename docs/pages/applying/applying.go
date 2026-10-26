package applying

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
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
		doc.P("This page covers delivering rules to the browser: the three kinds of stylesheet, the handler that writes them, the engine that inserts dynamic rules, the actions that drive it, and the errors. The ", doc.A("/styling", "Styling"), " page covers writing rules."),
		doc.P("Every example on this page is live. Its bound rule is in this site's `Dynamic` style, and the controls change it through the same actions your app would fire. The CSS tab shows the text of the rules for the demo's starting state."),
	}
}

func stylesheets() []app.UI {
	return []app.UI{
		doc.Go(`type Style = app.Stylesheet

func Static(path string) Style
func Generated(path string, items ...Item) Style
func Dynamic(items ...Item) Style`),
		doc.P("`Style` is `app.Stylesheet`, the interface `app.Handler.Styles` takes, and these three constructors make the kinds the handler understands. The handler turns a list of them into links and files, under ", doc.A("#handler-and-build", "Handler and Build"), ". `Item` has unexported methods too. `*Rule`, `*KeyframesRule`, `*FontFaceRule`, `*PropertyRule`, `*ImportRule`, `*LayerRule` and `*LayerOrderRule` implement it, and each has `Err() error`. A nil `Style` in a handler's list, and a nil item passed to `Generated` or `Dynamic`, are ignored."),
		stylesheetsDiagram(),
		doc.P("The three kinds are chosen per stylesheet, and one handler can mix them:"),
		doc.List(
			doc.Item("`Static(path)` links a `.css` file that already exists.", doc.List(
				doc.Item("A path starting with `/web/` is looked up in `WebDir`. If the file is missing, `Build` returns `app.ErrMissingStyle`."),
				doc.Item("A URL starting with `http://`, `https://` or `//` is linked as given and is not checked."),
				doc.Item("Any other path is `app.ErrPathOutsideWeb`."),
				doc.Item("The content of the file is never read or validated."),
			)),
			doc.Item("`Generated(path, items...)` writes a `.css` file from the items. The path must start with `/web/` (otherwise `app.ErrPathOutsideWeb`) and end in `.css` (otherwise `app.ErrNotCSSPath`). `Build` renders the items and writes the text to `<WebDir>/…`, creating directories as needed, only if the bytes differ from what is already on disk. The file is then linked like a Static one. A Generated style with no items, or whose items have no properties, still writes an empty file."),
			doc.Item("`Dynamic(items...)` adds no link and writes no file. It takes `*Rule` and `*KeyframesRule` values that carry a binding. `Build` only validates them. They are inserted into and removed from the page in the browser, as described under ", doc.A("#run", "Run"), " and ", doc.A("#actions", "Actions"), "."),
		),
		doc.P("A Generated file writes every `@layer` order statement first, then every `@import`, then everything else, each group in slice order, with one blank line between items."),
		doc.P("Paths are cleaned before use, so `/web/./app.css` and `/web/x/../app.css` both mean `/web/app.css`, and the link written into the page is the cleaned path. A path that leaves `/web/` once cleaned, such as `/web/../secret.css`, is `app.ErrPathOutsideWeb`. URLs are not cleaned."),
		doc.P("Use `Static` for CSS that has no typed form, such as a property the generator leaves out, and for a stylesheet hosted elsewhere."),
		doc.H("Resolve"),
		doc.Go(`type Kind = app.StyleKind

const (
	StaticStyle    = app.StaticStyle
	GeneratedStyle = app.GeneratedStyle
	DynamicStyle   = app.DynamicStyle
)

type Source = app.StyleSource

func Resolve(s Style) (Source, error)`),
		doc.P("`Resolve` is what the handler calls on each style. It checks the items and returns the style's kind, its path as given, and for a Generated style the rendered bytes, which are empty but not nil when the style has no items. A Static style is not checked beyond that, and a Dynamic one has no path. The errors it returns are the item errors under ", doc.A("#errors", "Errors"), ", and on an error the `Source` is zero. A nil `Style` resolves to a zero `Source` with no error. It is the same as calling the style's own `StyleSource` method. You need `Resolve` only to deliver styles through something other than `app.Handler`."),
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
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/style"
)

type hello struct {
	app.Compo
}

func (h *hello) Render() app.UI {
	return app.Div().Style(theme.Card).Body(
		app.A().Style(theme.Link).Href("/").Text("Home"),
	)
}

func main() {
	app.Route("/", func() app.Composer { return &hello{} })
	style.Run(theme.Styles()...)
	app.RunWhenOnBrowser()

	handler := &app.Handler{
		Name:   "Hello",
		Styles: theme.Styles(),
	}
	if err := handler.Build(); err != nil {
		log.Fatal(err)
	}
	http.Handle("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}`),
		doc.P("`Build` writes `web/theme.css` and links `/web/theme.css` in every page. Calling it yourself surfaces errors before the server starts; without the call, the first request builds, and on an error every request answers with a 500. The file contains:"),
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
		doc.Go(`Styles            []Stylesheet
WebDir            string
DisableAutoUpdate bool`),
		doc.P("There is one handler, `app.Handler`. Next to go-app's fields (`Name`, `Title`, `Icon`, `Resources`, `Version` and the rest) it has the three above. Its `Styles` takes the styles from `style.Static`, `style.Generated` and `style.Dynamic` directly, because `style.Style` is `app.Stylesheet`. A URL is a `style.Static` too: `style.Static(\"https://cdn.example.com/x.css\")`. `Handler` has unexported fields, so a composite literal must name its fields."),
		doc.Go(`handler := &app.Handler{
	Name:  "Docs",
	Title: "Docs",
	Styles: []style.Style{
		style.Static("/web/app.css"),
		style.Generated("/web/theme.css", theme.Card, theme.Link),
		style.Dynamic(theme.Compact),
	},
	WebDir: "web",
}`),
		doc.P("`theme.Compact` stands for any bound rule, such as the ones under ", doc.A("#actions", "Actions"), "."),
		doc.H("Build"),
		doc.Go(`func (h *Handler) Build() error`),
		doc.P("`Build` resolves every style in the order of `Styles`:"),
		doc.List(
			doc.Item("It returns the first problem it finds. It first checks every style, in the order of `Styles`: its items, through `StyleSource`, and then its path. Only then does it look up the Static files, so a missing Static file is reported only when no style has an item or path error, wherever each sits in the list. Generated files are written last, so a failed check leaves the web directory as it was."),
			doc.Item("On success the page links `/app.css`, then the libraries sorted by path, then the paths of the Static and Generated styles in order. Dynamic styles contribute no path."),
			doc.Item("Calling `Build` again gives the same links. Nothing is added twice."),
			doc.Item("You don't have to call it. The first request, and `app.GenerateStaticWebsite`, build the handler if nobody did. `GenerateStaticWebsite` returns the error; a server logs it and answers every request with a 500."),
			doc.Item("Every path still goes through the handler's `Resources` resolver. With `Resources: app.GitHubPages(\"WebGo\")`, the link to `/web/theme.css` is written as `/WebGo/web/theme.css`."),
		),
		doc.P("`app.GenerateStaticWebsite` does not copy `WebDir` into its output, so a static site publishes a generated stylesheet the same way it publishes `app.wasm`: by copying `WebDir` to `/web/`."),
		doc.H("WebDir"),
		doc.P("`WebDir` is where `Build` finds Static files and writes Generated ones: `/web/x/y.css` maps to `<WebDir>/x/y.css`. It must be the folder the handler serves at `/web/`. When it is empty, it follows `Resources`: `<dir>/web` for `app.LocalDir(dir)`, and `\"web\"` otherwise, which is what the default `Resources` serves. A relative `WebDir` is resolved against the working directory of the process that calls `Build`."),
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
