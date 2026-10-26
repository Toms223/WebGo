package topic

import (
	"strconv"
	"strings"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/theme"
)

const (
	wideAt     = 1248
	maxResults = 8
	railPanel  = "rail-panel"
)

type Rail struct {
	app.Compo
	Sections []section
	Active   int
	Visit    func(app.Context, string)
	wide     bool
	open     bool
	query    string
}

func NewRail(sections []section, active int, visit func(app.Context, string)) *Rail {
	r := &Rail{Sections: sections, Active: active, Visit: visit}
	r.measure()
	return r
}

func (r *Rail) measure() {
	width, _ := app.Window().Size()
	r.wide = width >= wideAt
}

func (r *Rail) OnResize(ctx app.Context) {
	was := r.wide
	r.measure()
	if r.wide != was {
		ctx.Update()
	}
}

func (r *Rail) Render() app.UI {
	if r.wide {
		return app.Nav().Style(theme.Rail).Aria("label", "Sections").Body(r.panel()...)
	}
	classes := []app.ClassRule{theme.Rail}
	if r.open {
		classes = append(classes, theme.RailOpen)
	}
	body := []app.UI{
		app.Button().Type("button").Style(theme.RailSummary).
			Aria("expanded", strconv.FormatBool(r.open)).Aria("controls", railPanel).
			OnClick(func(ctx app.Context, e app.Event) { r.open = !r.open }).
			Body(
				app.Span().Text(r.Sections[r.Active].Title),
				app.Span().Style(theme.RailCount).Text(strconv.Itoa(r.Active+1)+" of "+strconv.Itoa(len(r.Sections))),
			),
	}
	if r.open {
		body = append(body, app.Div().ID(railPanel).Style(theme.RailPanel).Body(r.panel()...))
	}
	return app.Nav().Style(classes...).Aria("label", "Sections").Body(body...)
}

func (r *Rail) panel() []app.UI {
	items := []app.UI{r.search()}
	if strings.TrimSpace(r.query) != "" {
		return append(items, r.results())
	}
	items = append(items, app.Span().Style(theme.RailLabel).Text("Sections"), r.sectionList())
	if subs := r.Sections[r.Active].Subs; len(subs) > 0 {
		items = append(items, app.Span().Style(theme.RailLabel).Text("In this section"), r.subList(subs))
	}
	return items
}

func (r *Rail) search() app.UI {
	return app.Input().Type("search").Style(theme.RailSearch).
		Aria("label", "Search sections and headings").Placeholder("Search sections").
		Value(r.query).OnInput(r.input).OnKeyDown(r.key)
}

func (r *Rail) input(ctx app.Context, e app.Event) {
	r.query = ctx.JSSrc().Get("value").String()
}

func (r *Rail) key(ctx app.Context, e app.Event) {
	if e.Get("key").String() != "Enter" {
		return
	}
	if found := r.matches(); len(found) > 0 {
		r.pick(ctx, found[0].ID)
	}
}

func (r *Rail) sectionList() app.UI {
	links := make([]app.UI, 0, len(r.Sections))
	for i, s := range r.Sections {
		link := r.link(s.Slug, s.Title, theme.RailLink)
		if i == r.Active {
			link = link.Style(theme.RailLinkActive).Aria("current", "true")
		}
		links = append(links, app.Li().Body(link))
	}
	return app.Ul().Body(links...)
}

func (r *Rail) subList(subs []sub) app.UI {
	links := make([]app.UI, 0, len(subs))
	for _, h := range subs {
		links = append(links, app.Li().Body(r.link(h.ID, h.Title, theme.RailLink, theme.RailLinkSub)))
	}
	return app.Ul().Body(links...)
}

type match struct {
	ID    string
	Title string
	Where string
}

func (r *Rail) matches() []match {
	needle := strings.ToLower(strings.TrimSpace(r.query))
	var found []match
	for _, s := range r.Sections {
		if strings.Contains(strings.ToLower(s.Title), needle) {
			found = append(found, match{ID: s.Slug, Title: s.Title})
		}
		for _, h := range s.Subs {
			if strings.Contains(strings.ToLower(h.Title), needle) {
				found = append(found, match{ID: h.ID, Title: h.Title, Where: s.Title})
			}
		}
	}
	if len(found) > maxResults {
		found = found[:maxResults]
	}
	return found
}

func (r *Rail) results() app.UI {
	found := r.matches()
	if len(found) == 0 {
		return app.P().Style(theme.RailEmpty).Text("No section or heading matches \"" + strings.TrimSpace(r.query) + "\".")
	}
	links := make([]app.UI, 0, len(found))
	for _, m := range found {
		link := r.link(m.ID, m.Title, theme.RailLink)
		if m.Where != "" {
			link = link.Body(app.Span().Text(m.Title), app.Span().Style(theme.RailWhere).Text(m.Where))
		}
		links = append(links, app.Li().Body(link))
	}
	return app.Ul().Aria("label", "Search results").Body(links...)
}

func (r *Rail) link(id, title string, classes ...app.ClassRule) app.HTMLA {
	return app.A().Style(classes...).Href("#" + id).Text(title).OnClick(func(ctx app.Context, e app.Event) {
		e.PreventDefault()
		r.pick(ctx, id)
	})
}

func (r *Rail) pick(ctx app.Context, id string) {
	r.query = ""
	r.open = false
	r.Visit(ctx, id)
}
