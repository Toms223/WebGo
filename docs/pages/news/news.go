package news

import (
	"strconv"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/data"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/page"
	"github.com/Toms223/WebGo/style"
)

const (
	Title = "News"
	Route = base.NewsRoute
)

type Article struct {
	Slug      string
	Version   string
	Label     string
	Codename  string
	Date      string
	Minutes   int
	Fill      *style.Rule
	Signature string
	Title     string
	Teaser    string
	Short     []string
	SignOff   string
	Body      func() []app.UI
}

func (a Article) Route() string { return Route + "/" + a.Slug }

func (a Article) Versions() string {
	if a.Label != "" {
		return a.Label
	}
	return a.Version
}

func Articles() []Article {
	return []Article{homecoming(), typedInk(), firstLight()}
}

func Latest() Article { return Articles()[0] }

func Routes() []string {
	routes := []string{Route}
	for _, a := range Articles() {
		routes = append(routes, a.Route())
	}
	return routes
}

type State struct{}

type Screen struct {
	*page.Page[State]
	title string
}

func (s *Screen) OnMount(ctx app.Context) page.Screen {
	s.Page.OnMount(ctx)
	ctx.Page().SetTitle(s.title)
	base.Activate(ctx, s.Route())
	return s
}

func newScreen(key, route, title string, layout page.Layout[State]) (*Screen, error) {
	viewModel, err := data.NewViewModel(key, &State{})
	if err != nil {
		return nil, err
	}
	built := page.NewPage[State](route).WithLayout(layout).WithModel(viewModel)
	if err := built.Err(); err != nil {
		return nil, err
	}
	return &Screen{Page: built, title: title}, nil
}

func Screens() ([]page.Screen, error) {
	feed, err := newScreen("news", Route, "News · WebGo", &feedLayout{})
	if err != nil {
		return nil, err
	}
	screens := []page.Screen{feed}
	articles := Articles()
	for i, a := range articles {
		layout := &articleLayout{article: a}
		if i > 0 {
			newer := articles[i-1]
			layout.newer = &newer
		}
		if i+1 < len(articles) {
			older := articles[i+1]
			layout.older = &older
		}
		screen, err := newScreen("news/"+a.Slug, a.Route(), a.Title+" · WebGo news", layout)
		if err != nil {
			return nil, err
		}
		screens = append(screens, screen)
	}
	return screens, nil
}

func Cover(a Article, wide bool) app.UI {
	classes := []app.ClassRule{theme.NewsCover, a.Fill}
	if wide {
		classes = []app.ClassRule{theme.NewsCoverWide, a.Fill}
	}
	return app.Div().Style(classes...).Aria("hidden", "true").Body(
		app.Div().Body(
			app.P().Style(theme.NewsCoverVersion).Text(a.Version),
			app.P().Style(theme.NewsCoverName).Text(a.Codename),
		),
		app.Span().Style(theme.NewsCoverCard).Text(a.Signature),
	)
}

func Meta(a Article, extra ...string) app.UI {
	items := []app.UI{
		app.Li().Text("Update " + a.Versions()),
		app.Li().Body(app.Time().Text(a.Date)),
		app.Li().Text(strconv.Itoa(a.Minutes) + " min read"),
	}
	for _, e := range extra {
		items = append(items, app.Li().Text(e))
	}
	return app.Ul().Style(theme.NewsMeta).Body(items...)
}

type layout struct{}

func (layout) Loading() app.UI {
	return app.Div().Style(theme.News, theme.ShellLoading).Text("Loading")
}
func (layout) Error(err error) app.UI {
	return app.Div().Style(theme.News, theme.ShellError).Text(err.Error())
}

type feedLayout struct{ layout }

func (l *feedLayout) Body(...app.UI) page.Layout[State] { return l }

func (l *feedLayout) Draw(*data.ViewModel[State]) app.HTMLDiv {
	articles := Articles()
	latest := articles[0]
	rows := []app.UI{app.H2().Style(theme.NewsSectionTitle).Text("Earlier updates")}
	for _, a := range articles[1:] {
		rows = append(rows, app.A().Style(theme.NewsRow).Href(nav.Href(a.Route())).Body(
			Cover(a, false),
			app.Div().Style(theme.NewsStory).Body(
				app.H3().Style(theme.NewsRowTitle).Text(a.Title),
				app.P().Style(theme.NewsTeaser).Text(a.Teaser),
				Meta(a),
			),
		))
	}
	return app.Div().Style(theme.News).Body(
		app.Header().Style(theme.NewsHead).Body(
			app.H1().Style(theme.NewsMasthead).Text("News"),
			app.P().Style(theme.NewsIntro).Text("What we shipped, why it matters, and how to get it. Straight from the team that builds WebGo."),
		),
		app.A().Style(theme.NewsFeature).Href(nav.Href(latest.Route())).Body(
			Cover(latest, false),
			app.Div().Style(theme.NewsStory).Body(
				app.H2().Style(theme.NewsHeadline).Text(latest.Title),
				app.P().Style(theme.NewsTeaser).Text(latest.Teaser),
				Meta(latest),
				app.Span().Style(theme.NewsGo).Text("Read the update →"),
			),
		),
		&catchUp{},
		app.Section().Style(theme.NewsEarlier).Body(rows...),
	)
}

type articleLayout struct {
	layout
	article Article
	newer   *Article
	older   *Article
}

func (l *articleLayout) Body(...app.UI) page.Layout[State] { return l }

func (l *articleLayout) Draw(*data.ViewModel[State]) app.HTMLDiv {
	a := l.article
	short := make([]app.UI, 0, len(a.Short))
	for _, s := range a.Short {
		short = append(short, doc.Item(s))
	}
	pager := []app.UI{}
	if l.older != nil {
		pager = append(pager, pagerLink("Older update", *l.older))
	}
	if l.newer != nil {
		pager = append(pager, pagerLink("Newer update", *l.newer))
	}
	return app.Div().Style(theme.NewsArticle).Body(
		app.A().Style(theme.NewsBack).Href(nav.Href(Route)).Text("← All news"),
		Cover(a, true),
		app.Article().Style(theme.NewsColumn).Body(
			app.H1().Style(theme.NewsTitle).Text(a.Title),
			app.P().Style(theme.NewsIntro).Text(a.Teaser),
			Meta(a, a.Codename, "By the WebGo team"),
			app.Div().Style(theme.NewsShort).Body(
				app.P().Style(theme.NewsShortTitle).Text("The short version"),
				app.Ul().Body(short...),
			),
			app.Div().Style(theme.Doc, theme.NewsBody).Body(a.Body()...),
			app.P().Style(theme.NewsSign).Text(a.SignOff),
			app.Nav().Style(theme.NewsPager).Aria("label", "More updates").Body(pager...),
		),
	)
}

func pagerLink(label string, a Article) app.UI {
	return app.A().Style(theme.NewsPagerLink).Href(nav.Href(a.Route())).Body(
		app.Span().Style(theme.NewsPagerLabel).Text(label+" · "+a.Versions()),
		app.Span().Style(theme.NewsPagerTitle).Text(a.Title),
	)
}

type compare struct {
	app.Compo
	BeforeLabel string
	AfterLabel  string
	Before      string
	After       string
	before      bool
}

func (c *compare) Render() app.UI {
	source := c.After
	if c.before {
		source = c.Before
	}
	return app.Div().Style(theme.NewsCompare).Body(
		app.Div().Style(theme.NewsPicks).Aria("label", "Compare versions").Body(
			c.pick(c.BeforeLabel, true),
			c.pick(c.AfterLabel, false),
		),
		doc.Go(source),
	)
}

func (c *compare) pick(label string, before bool) app.UI {
	classes := []app.ClassRule{theme.NewsPick}
	if c.before == before {
		classes = []app.ClassRule{theme.NewsPickOn}
	}
	return app.Button().Type("button").Style(classes...).
		Aria("pressed", strconv.FormatBool(c.before == before)).
		Text(label).
		OnClick(func(ctx app.Context, e app.Event) { c.before = before })
}

type catchUp struct {
	app.Compo
	from int
}

func (c *catchUp) Render() app.UI {
	articles := Articles()
	picks := make([]app.UI, 0, len(articles))
	for i := len(articles) - 1; i >= 0; i-- {
		label := articles[i].Version
		if i == len(articles)-1 {
			label = articles[i].Version + " or older"
		}
		picks = append(picks, c.pick(label, i))
	}
	return app.Section().Style(theme.NewsCatchUp).Body(
		app.Div().Style(theme.NewsAsk).Body(
			app.H2().Style(theme.NewsSectionTitle).Text("What did I miss?"),
			app.P().Style(theme.NewsAskText).Text("Pick the version you're on. We'll show you what changed since, and the one line that gets you current."),
			app.Div().Style(theme.NewsPicks).Aria("label", "Your version").Body(picks...),
		),
		c.answer(articles),
	)
}

func (c *catchUp) pick(label string, i int) app.UI {
	on := c.selected(len(Articles())) == i
	classes := []app.ClassRule{theme.NewsPick}
	if on {
		classes = []app.ClassRule{theme.NewsPickOn}
	}
	return app.Button().Type("button").Style(classes...).
		Aria("pressed", strconv.FormatBool(on)).
		Text(label).
		OnClick(func(ctx app.Context, e app.Event) { c.from = i + 1 })
}

func (c *catchUp) selected(count int) int {
	if c.from == 0 {
		return count - 1
	}
	return c.from - 1
}

func (c *catchUp) answer(articles []Article) app.UI {
	missed := articles[:c.selected(len(articles))]
	if len(missed) == 0 {
		return app.Div().Style(theme.NewsAnswer).Aria("live", "polite").Body(
			app.P().Style(theme.NewsAnswerTitle).Text("You're up to date."),
			app.P().Style(theme.NewsAskText).Text("Nothing new since "+articles[0].Version+". We'll have more soon."),
		)
	}
	title := "You missed one update."
	if len(missed) > 1 {
		title = "You missed " + strconv.Itoa(len(missed)) + " updates."
	}
	links := make([]app.UI, 0, len(missed))
	for _, a := range missed {
		links = append(links, app.Li().Body(
			app.A().Href(nav.Href(a.Route())).Text(a.Title),
			app.Span().Style(theme.NewsMissedNote).Text(" · "+a.Version+", "+a.Codename),
		))
	}
	command := "go get github.com/Toms223/WebGo@v" + articles[0].Version
	return app.Div().Style(theme.NewsAnswer).Aria("live", "polite").Body(
		app.P().Style(theme.NewsAnswerTitle).Text(title),
		app.Ul().Style(theme.NewsMissed).Body(links...),
		app.Div().Style(theme.NewsCommand).Body(
			app.Code().Style(theme.NewsCommandText).Text(command),
			&doc.Copy{Text: command},
		),
	)
}
