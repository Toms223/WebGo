package demos

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var ListBox = style.Class("demo-components-list").Set(
	css.Display(kw.Grid),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(1)),
	css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	css.Background(css.Var("--mint")),
)

type entry struct {
	ID   int
	Text string
}

type listDemo struct {
	app.Compo
	draft string
	next  int
	items []entry
}

func ListDemo() app.UI {
	return &listDemo{next: 3, items: []entry{{1, "Read the Components page"}, {2, "Remove me"}}}
}

func (d *listDemo) Render() app.UI {
	return app.Div().Style(ListBox).Body(
		app.Div().Body(
			app.Input().Type("text").Value(d.draft).Placeholder("New item").OnInput(d.ValueTo(&d.draft)),
			app.Button().Type("button").Text("Add").OnClick(d.add),
		),
		app.If(len(d.items) == 0, func() app.UI {
			return app.P().Text("Nothing left.")
		}).Else(func() app.UI {
			return app.Ul().Body(
				app.Range(d.items).Slice(func(i int) app.UI {
					item := d.items[i]
					return app.Li().Body(
						app.Text(item.Text+" "),
						app.Button().Type("button").Text("Remove").
							OnClick(func(ctx app.Context, e app.Event) { d.remove(item.ID) }, app.EventScope(item.ID)),
					)
				}),
			)
		}),
		app.Small().Textf("%d items", len(d.items)),
	)
}

func (d *listDemo) add(ctx app.Context, e app.Event) {
	if d.draft == "" {
		return
	}
	d.items = append(d.items, entry{d.next, d.draft})
	d.next++
	d.draft = ""
}

func (d *listDemo) remove(id int) {
	for i, item := range d.items {
		if item.ID == id {
			d.items = append(d.items[:i], d.items[i+1:]...)
			return
		}
	}
}

func Items() []style.Item {
	return []style.Item{ListBox}
}

const listSource = `type entry struct {
	ID   int
	Text string
}

type listDemo struct {
	app.Compo
	draft string
	next  int
	items []entry
}

func ListDemo() app.UI {
	return &listDemo{next: 3, items: []entry{{1, "Read the Components page"}, {2, "Remove me"}}}
}

func (d *listDemo) Render() app.UI {
	return app.Div().Style(ListBox).Body(
		app.Div().Body(
			app.Input().Type("text").Value(d.draft).Placeholder("New item").OnInput(d.ValueTo(&d.draft)),
			app.Button().Type("button").Text("Add").OnClick(d.add),
		),
		app.If(len(d.items) == 0, func() app.UI {
			return app.P().Text("Nothing left.")
		}).Else(func() app.UI {
			return app.Ul().Body(
				app.Range(d.items).Slice(func(i int) app.UI {
					item := d.items[i]
					return app.Li().Body(
						app.Text(item.Text+" "),
						app.Button().Type("button").Text("Remove").
							OnClick(func(ctx app.Context, e app.Event) { d.remove(item.ID) }, app.EventScope(item.ID)),
					)
				}),
			)
		}),
		app.Small().Textf("%d items", len(d.items)),
	)
}

func (d *listDemo) add(ctx app.Context, e app.Event) {
	if d.draft == "" {
		return
	}
	d.items = append(d.items, entry{d.next, d.draft})
	d.next++
	d.draft = ""
}

func (d *listDemo) remove(id int) {
	for i, item := range d.items {
		if item.ID == id {
			d.items = append(d.items[:i], d.items[i+1:]...)
			return
		}
	}
}`

var ListExample = example.Example{Name: "list.go", Source: listSource, CSS: ListBox.CSS(), Live: ListDemo}

func Examples() []example.Example {
	return []example.Example{ListExample}
}
