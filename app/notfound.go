package app

import "reflect"

var (
	// NotFound is the component shown for paths no route matches, both by the
	// Handler, which answers with status 404, and in the browser. Assign a
	// pointer to your own component type; each miss gets a new zero value of
	// that type.
	NotFound UI = &notFound{}
)

func notFoundComponent() Composer {
	if c, ok := NotFound.(Composer); ok {
		if t := reflect.TypeOf(c); t.Kind() == reflect.Pointer {
			return reflect.New(t.Elem()).Interface().(Composer)
		}
	}
	return &notFound{}
}

type notFound struct {
	Compo
	Icon string
}

func (n *notFound) OnMount(Context) {
	links := Window().Get("document").Call("getElementsByTagName", "link")

	for i := 0; i < links.Length(); i++ {
		link := links.Index(i)
		rel := link.Call("getAttribute", "rel")

		if rel.String() == "icon" {
			favicon := link.Call("getAttribute", "href")
			n.Icon = favicon.String()
			return
		}
	}
}

func (n *notFound) Render() UI {
	return Div().
		Class("goapp-app-info").
		Body(
			Div().
				Class("goapp-notfound-title").
				Body(
					Text("4"),
					Img().
						Class("goapp-logo").
						Alt("0").
						Src(n.Icon),
					Text("4"),
				),
			P().
				Class("goapp-label").
				Text("Not Found"),
		)
}
