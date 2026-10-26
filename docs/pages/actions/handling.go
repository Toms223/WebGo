package actions

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
)

func actions() []app.UI {
	return []app.UI{
		doc.Go(`type Action struct {
	Name  string
	Value any
	Tags  Tags
}

func (ctx Context) NewAction(action string, tags ...Tagger)
func (ctx Context) NewActionWithValue(action string, v any, tags ...Tagger)`),
		doc.P("An action is a named event. Unlike a state it is not stored: creating one runs the handlers registered for its name at that moment, and a handler registered later never sees it."),
		doc.List(
			doc.Item("`NewAction(name, tags...)` is `NewActionWithValue(name, nil, tags...)`."),
			doc.Item("`v` is passed to every handler as it is, without a copy. A pointer is shared by all of them."),
			doc.Item("The tags are merged into one `Tags` map in the order given, and a later tag replaces an earlier one with the same name. With no tags, `Tags` is nil."),
			doc.Item("If no handler is registered for the name, nothing happens."),
			doc.Item("Both methods can be called from any goroutine."),
		),
		doc.H("Tags"),
		doc.Go(`type Tags map[string]string

func (t Tags) Get(name string) string
func (t Tags) Set(name string, v any)
func (t Tags) Tags() Tags

type Tagger interface {
	Tags() Tags
}

type Tag struct {
	Name  string
	Value string
}

func T(name string, value any) Tag`),
		doc.P("Tags are string metadata about an action, such as where it came from. `T` and `Tags.Set` turn the value into a string: strings and byte slices as they are, `int` and `bool` in decimal and `true`/`false`, `float64` with four decimals, `nil` as an empty string, and other types through `fmt`. Both `Tag` and `Tags` implement `Tagger`, so a whole map can be passed as one argument."),
		doc.Go(`func (s *searchBox) onInput(ctx app.Context, e app.Event) {
	query := ctx.JSSrc().Get("value").String()
	ctx.NewActionWithValue("search", query, app.T("source", "header"))
}`),
	}
}

func handling() []app.UI {
	return []app.UI{
		doc.Go(`type ActionHandler func(Context, Action)

func Handle(actionName string, h ActionHandler)
func (ctx Context) Handle(action string, h ActionHandler)`),
		doc.P("There are two kinds of handler. `app.Handle` registers a global handler for business logic. `ctx.Handle` registers a handler that belongs to a component and updates it."),
		handlersDiagram(),
		doc.H("Global handlers"),
		doc.List(
			doc.Item("`app.Handle` stores the handler in a package-level map. Call it in `main`, before `app.RunWhenOnBrowser`: the engine reads the map when it loads its first component, and later registrations are not seen."),
			doc.Item("There is one global handler per name. A second `app.Handle` with the same name replaces the first. WebGo's style engine registers its own names this way, see ", doc.A("/applying#run", "Run"), "."),
			doc.Item("Each action runs the handler on a new goroutine, through `ctx.Async`, so it can block on network or disk. It must not change component fields."),
			doc.Item("Its `ctx` belongs to no component. `ctx.Dispatch` still runs a function on the UI goroutine, but it re-renders nothing. To change what is shown, set a state or create another action."),
			doc.Item("Global handlers also run on the server while a page is prerendered, if something creates their action there. A `TestEngine` does not use them."),
		),
		doc.H("Component handlers"),
		doc.List(
			doc.Item("`ctx.Handle` registers `h` for the source of `ctx`, normally the component in `OnMount`."),
			doc.Item("It runs on the UI goroutine, through `ctx.Dispatch`, so it can change fields, and the component and its ancestors re-render afterwards."),
			doc.Item("It is removed when its component is dismounted."),
			doc.Item("A component can register several functions for one name. Registering the same function again, such as the same method or a closure from the same function literal, replaces the earlier registration rather than adding one."),
			doc.Item("The order in which the handlers of one action run is not specified."),
			doc.Item("`OnMount` runs only in the browser, so handlers registered there never run during prerendering."),
		),
		doc.H("Example: a search"),
		doc.Go(`func main() {
	app.Handle("search", runSearch)
	app.Route("/", func() app.Composer { return &results{} })
	app.RunWhenOnBrowser()
}

func runSearch(ctx app.Context, a app.Action) {
	query, ok := a.Value.(string)
	if !ok {
		return
	}
	hits, err := lookup(query)
	if err != nil {
		app.Log(err)
		return
	}
	ctx.SetState("search-hits", hits)
}

type results struct {
	app.Compo
	query string
	hits  []string
}

func (r *results) OnMount(ctx app.Context) {
	ctx.Handle("search", r.onSearch)
	ctx.ObserveState("search-hits", &r.hits)
}

func (r *results) onSearch(ctx app.Context, a app.Action) {
	r.query, _ = a.Value.(string)
}

func (r *results) Render() app.UI {
	return app.Div().Body(
		app.P().Text("Results for " + r.query),
		app.Ul().Body(
			app.Range(r.hits).Slice(func(i int) app.UI {
				return app.Li().Text(r.hits[i])
			}),
		),
	)
}`),
		doc.P("`lookup` stands for any blocking call of your own. The `search` action from the search box above reaches both handlers. `results.onSearch` runs on the UI goroutine and shows the query at once. `runSearch` runs on its own goroutine, waits for `lookup`, and hands the hits back through a state that `results` observes. The server part of `main` is left out."),
	}
}

func choosing() []app.UI {
	return []app.UI{
		doc.List(
			doc.Item(doc.B("A state"), " is for a value that has a current value: the signed-in user, a filter, a cart. A component mounted later can still read it, and it can be persisted and broadcast."),
			doc.Item(doc.B("An action"), " is for something that happened: a click handled somewhere else, or a request to do work off the UI goroutine. It is not stored, so a component mounted later does not see it."),
			doc.Item(doc.B("A ViewModel"), ", `data.ViewModel[T]`, is for typed state shared by a page. It binds one key to a `*T`, so components share the pointer instead of each declaring a receiver, and changes go through `Load(ctx).Apply(...).Set()`. Underneath it is `GetState`, `SetState` and `ObserveState`, and `SetAndPersist` is `SetState(...).Persist()`. It is on ", doc.A("/state#model", "Pages, layouts and state"), "."),
		),
		doc.P("Dynamic styles join the two. `ViewModel.OnAction` observes the model's key and, on each change, creates an action named after the key with the model's `*T` as its value. `style.When` and `style.Compute` bind rules to that name, so the page's CSS follows the state. This is under ", doc.A("/applying#viewmodel-and-onaction", "ViewModel and OnAction"), "."),
		doc.Note("Action names that dynamic styles are bound to belong to WebGo's style engine. Handle them with `ctx.Handle` in a component, never with `app.Handle`, or one of the two handlers is lost."),
	}
}
