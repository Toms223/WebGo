package style

import (
	"slices"

	"github.com/Toms223/WebGo/app"
)

type dynamicItem interface {
	Item
	textsFor(a app.Action) outcome
	startTexts() (texts []string, present bool)
	actions() []string
}

type engine struct {
	sheet   sheet
	rules   []dynamicItem
	present map[Item]entry
	bound   map[string][]dynamicItem
}

type entry struct {
	indices []int
	texts   []string
}

func newEngine(s sheet) *engine {
	return &engine{
		sheet:   s,
		present: make(map[Item]entry),
		bound:   make(map[string][]dynamicItem),
	}
}

func (e *engine) register(i dynamicItem) {
	if !slices.Contains(e.rules, i) {
		e.rules = append(e.rules, i)
	}
	for _, action := range i.actions() {
		if !slices.Contains(e.bound[action], i) {
			e.bound[action] = append(e.bound[action], i)
		}
	}
}

func (e *engine) start() {
	for _, i := range e.rules {
		if texts, present := i.startTexts(); present {
			e.set(i, texts, true)
		}
	}
}

func (e *engine) apply(a app.Action) {
	for _, i := range e.bound[a.Name] {
		o := i.textsFor(a)
		switch {
		case !o.handled:
			app.Logf("style: action %q carries a %T, %s unchanged", a.Name, a.Value, i.subject())
		case o.present && o.computes:
			e.replace(i, o.texts)
		default:
			e.set(i, o.texts, o.present)
		}
	}
}

func (e *engine) set(i Item, texts []string, present bool) {
	_, isPresent := e.present[i]
	switch {
	case present && !isPresent:
		e.add(i, texts)
	case !present && isPresent:
		e.drop(i)
	}
}

func (e *engine) replace(i Item, texts []string) {
	if current, ok := e.present[i]; ok && slices.Equal(current.texts, texts) {
		return
	}
	e.drop(i)
	e.add(i, texts)
}

func (e *engine) add(i Item, texts []string) {
	var indices []int
	for _, text := range texts {
		index := e.sheet.length()
		if e.sheet.insert(text, index) {
			indices = append(indices, index)
		}
	}
	if len(indices) > 0 {
		e.present[i] = entry{indices: indices, texts: texts}
	}
}

func (e *engine) drop(i Item) {
	indices := e.present[i].indices
	delete(e.present, i)
	for _, index := range slices.Backward(indices) {
		e.sheet.remove(index)
		e.shiftDown(index)
	}
}

func (e *engine) shiftDown(removed int) {
	for _, current := range e.present {
		for i := range current.indices {
			if current.indices[i] > removed {
				current.indices[i]--
			}
		}
	}
}

func (e *engine) handler() app.ActionHandler {
	return func(ctx app.Context, a app.Action) {
		ctx.Dispatch(func(app.Context) { e.apply(a) })
	}
}

// Run starts the browser engine for the Dynamic styles among styles: it registers one app.Handle
// handler per bound action name and inserts the items set Initially(true). It returns at once on the
// server. Call it once, before app.RunWhenOnBrowser, with every Dynamic style, since a second call
// replaces the first one's handlers for shared action names.
func Run(styles ...Style) {
	if app.IsServer {
		return
	}
	e := newEngine(&jsSheet{})
	for _, i := range dynamicItems(styles) {
		if d, ok := i.(dynamicItem); ok {
			e.register(d)
		}
	}
	for name := range e.bound {
		app.Handle(name, e.handler())
	}
	e.start()
}
