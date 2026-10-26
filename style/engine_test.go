package style

import (
	"errors"
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/media"

	"github.com/Toms223/WebGo/app"
)

type fakeSheet struct {
	rules    []string
	rejects  map[string]bool
	attempts int
}

func (f *fakeSheet) insert(text string, index int) bool {
	f.attempts++
	if f.rejects[text] {
		return false
	}
	f.rules = slices.Insert(f.rules, index, text)
	return true
}

func (f *fakeSheet) remove(index int) {
	f.rules = slices.Delete(f.rules, index, index+1)
}

func (f *fakeSheet) length() int { return len(f.rules) }

func rejecting(texts ...string) *fakeSheet {
	fake := &fakeSheet{rejects: make(map[string]bool)}
	for _, text := range texts {
		fake.rejects[text] = true
	}
	return fake
}

type host struct {
	app.Compo
	Tick     int
	captured *app.Context
}

func (h *host) OnUpdate(ctx app.Context) { *h.captured = ctx }

func (h *host) Render() app.UI { return app.Div() }

func newContext(t *testing.T) (app.Context, app.TestEngine) {
	t.Helper()

	testEngine := app.NewTestEngine()
	var ctx app.Context

	if err := testEngine.Load(&host{captured: &ctx, Tick: 0}); err != nil {
		t.Fatalf("mounting the host component failed: %v", err)
	}
	if err := testEngine.Load(&host{Tick: 1}); err != nil {
		t.Fatalf("updating the host component failed: %v", err)
	}
	testEngine.ConsumeAll()

	if ctx.Src() == nil {
		t.Fatal("the host was never updated, so no Context was captured")
	}
	return ctx, testEngine
}

type fakeDOM struct {
	created  int
	tag      string
	attrs    map[string]string
	attached bool
	noSheet  bool
	throws   string
	rules    []string
}

func newFakeDOM() *fakeDOM { return &fakeDOM{attrs: make(map[string]string)} }

func (d *fakeDOM) document() app.Value { return fakeNode{dom: d, kind: "document"} }

func (d *fakeDOM) throwOn(member string) {
	if member == d.throws {
		panic(errors.New("boom"))
	}
}

type fakeNode struct {
	app.Value
	dom  *fakeDOM
	kind string
}

func (n fakeNode) Truthy() bool { return true }

func (n fakeNode) Length() int { return len(n.dom.rules) }

func (n fakeNode) Get(name string) app.Value {
	n.dom.throwOn(n.kind + "." + name)
	switch n.kind + "." + name {
	case "document.head":
		return fakeNode{dom: n.dom, kind: "head"}
	case "element.sheet":
		if n.dom.noSheet || !n.dom.attached {
			return fakeNull{}
		}
		return fakeNode{dom: n.dom, kind: "sheet"}
	case "sheet.cssRules":
		return fakeNode{dom: n.dom, kind: "rules"}
	}
	panic("unexpected Get of " + name + " on " + n.kind)
}

func (n fakeNode) Call(name string, args ...any) app.Value {
	n.dom.throwOn(n.kind + "." + name)
	switch n.kind + "." + name {
	case "document.createElement":
		n.dom.created++
		n.dom.tag = args[0].(string)
		return fakeNode{dom: n.dom, kind: "element"}
	case "element.setAttribute":
		n.dom.attrs[args[0].(string)] = args[1].(string)
	case "head.appendChild":
		n.dom.attached = true
	case "sheet.insertRule":
		n.dom.rules = slices.Insert(n.dom.rules, args[1].(int), args[0].(string))
	case "sheet.deleteRule":
		index := args[0].(int)
		n.dom.rules = slices.Delete(n.dom.rules, index, index+1)
	default:
		panic("unexpected Call of " + name + " on " + n.kind)
	}
	return fakeNull{}
}

type fakeNull struct{ app.Value }

func (fakeNull) Truthy() bool { return false }

func (fakeNull) Length() int { panic(errors.New("Length of null")) }

func (fakeNull) Get(name string) app.Value { panic(fmt.Errorf("Get(%q) of null", name)) }

func (fakeNull) Call(name string, _ ...any) app.Value { panic(fmt.Errorf("Call(%q) of null", name)) }

func captureLog(t *testing.T) func() []string {
	t.Helper()
	var lines []string
	previous := app.DefaultLogger
	app.DefaultLogger = func(format string, v ...any) {
		lines = append(lines, fmt.Sprintf(format, v...))
	}
	t.Cleanup(func() { app.DefaultLogger = previous })
	return func() []string { return lines }
}

func truthy(a app.Action) bool {
	on, _ := a.Value.(bool)
	return on
}

func toggled(name string) *Rule {
	return Class(name).Set(css.Custom("--n", css.Px(1))).On(name, truthy)
}

func baseBlock(name string) string {
	return "." + name + " {\n\t--n: 1px;\n}"
}

func screenBlock(name string) string {
	return "@media screen {\n\t." + name + " {\n\t\t--m: 2px;\n\t}\n}"
}

func withScreen(r *Rule) *Rule {
	return r.Media(media.Screen(), css.Custom("--m", css.Px(2)))
}

func fire(e *engine, name string, on bool) {
	e.apply(app.Action{Name: name, Value: on})
}

func expectRules(t *testing.T, fake *fakeSheet, want ...string) {
	t.Helper()
	if !slices.Equal(fake.rules, want) {
		t.Errorf("rules = %q, want %q", fake.rules, want)
	}
}

func TestInsertsOnTrueAndRemovesOnFalse(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("dark").Set(css.Custom("--w", css.Px(1))).On("theme/dark", func(a app.Action) bool { return a.Value.(bool) }))

	e.apply(app.Action{Name: "theme/dark", Value: true})
	expectRules(t, fake, ".dark {\n\t--w: 1px;\n}")

	e.apply(app.Action{Name: "theme/dark", Value: false})
	expectRules(t, fake)
}

func TestAppliesEveryBlockAsOneCSSOMRule(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("nav").
		Set(css.Custom("--a", css.Px(1))).
		pseudo(":hover", []css.Property{css.Custom("--b", css.Px(2))}).
		Media(media.Screen(), css.Custom("--c", css.Px(3))).
		On("nav", truthy))

	fire(e, "nav", true)
	expectRules(t, fake,
		".nav {\n\t--a: 1px;\n}",
		".nav:hover {\n\t--b: 2px;\n}",
		"@media screen {\n\t.nav {\n\t\t--c: 3px;\n\t}\n}",
	)

	fire(e, "nav", false)
	expectRules(t, fake)
}

func TestIdempotentOnRepeatedValue(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(toggled("a"))

	fire(e, "a", true)
	fire(e, "a", true)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "a", false)
	fire(e, "a", false)
	expectRules(t, fake)
}

func TestRemovingMiddleRuleShiftsIndices(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	for _, name := range []string{"a", "b", "c"} {
		e.register(toggled(name))
		fire(e, name, true)
	}
	expectRules(t, fake, baseBlock("a"), baseBlock("b"), baseBlock("c"))

	fire(e, "b", false)
	expectRules(t, fake, baseBlock("a"), baseBlock("c"))

	fire(e, "c", false)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "a", false)
	expectRules(t, fake)
}

func TestRemovingMultiBlockRulesKeepsTheRestAligned(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(toggled("a"))
	e.register(withScreen(toggled("b")))
	e.register(withScreen(toggled("c")))
	for _, name := range []string{"a", "b", "c"} {
		fire(e, name, true)
	}
	expectRules(t, fake, baseBlock("a"), baseBlock("b"), screenBlock("b"), baseBlock("c"), screenBlock("c"))

	fire(e, "b", false)
	expectRules(t, fake, baseBlock("a"), baseBlock("c"), screenBlock("c"))

	fire(e, "c", false)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "b", true)
	expectRules(t, fake, baseBlock("a"), baseBlock("b"), screenBlock("b"))

	fire(e, "a", false)
	expectRules(t, fake, baseBlock("b"), screenBlock("b"))

	fire(e, "b", false)
	expectRules(t, fake)
}

func TestRandomTogglingAndRejectionsMatchAnOrderedModel(t *testing.T) {
	names := []string{"a", "b", "c", "d", "e", "f"}
	rules := make(map[string]*Rule)
	var texts []string
	for i, name := range names {
		rule := toggled(name)
		if i%2 == 1 {
			rule = withScreen(rule)
		}
		if i%3 == 0 {
			rule = rule.pseudo(":hover", []css.Property{css.Custom("--h", css.Px(3))})
		}
		rules[name] = rule
		texts = append(texts, rule.blockTexts()...)
	}

	fake := rejecting()
	e := newEngine(fake)
	for _, name := range names {
		e.register(rules[name])
	}

	rng := rand.New(rand.NewPCG(1, 2))
	var order []string
	shown := make(map[string][]string)
	for step := range 4000 {
		if rng.IntN(4) == 0 {
			text := texts[rng.IntN(len(texts))]
			fake.rejects[text] = !fake.rejects[text]
		}
		name := names[rng.IntN(len(names))]
		on := rng.IntN(2) == 0
		fire(e, name, on)

		present := slices.Contains(order, name)
		switch {
		case on && !present:
			accepted := slices.DeleteFunc(rules[name].blockTexts(), func(text string) bool { return fake.rejects[text] })
			if len(accepted) > 0 {
				order = append(order, name)
				shown[name] = accepted
			}
		case !on && present:
			order = slices.DeleteFunc(order, func(n string) bool { return n == name })
			delete(shown, name)
		}

		var want []string
		for _, n := range order {
			want = append(want, shown[n]...)
		}
		if !slices.Equal(fake.rules, want) {
			t.Fatalf("step %d (%s=%v): rules = %q, want %q", step, name, on, fake.rules, want)
		}
	}
}

func TestRejectedBlockIsNotRecorded(t *testing.T) {
	fake := rejecting(baseBlock("b"))
	e := newEngine(fake)
	e.register(toggled("a"))
	e.register(withScreen(toggled("b")))
	e.register(withScreen(toggled("c")))
	for _, name := range []string{"a", "b", "c"} {
		fire(e, name, true)
	}
	expectRules(t, fake, baseBlock("a"), screenBlock("b"), baseBlock("c"), screenBlock("c"))

	fire(e, "b", false)
	expectRules(t, fake, baseBlock("a"), baseBlock("c"), screenBlock("c"))

	fire(e, "c", false)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "a", false)
	expectRules(t, fake)
}

func TestRuleWithOnlyRejectedBlockIsRetried(t *testing.T) {
	fake := rejecting(baseBlock("a"))
	e := newEngine(fake)
	e.register(toggled("a").Initially(true))

	e.start()
	expectRules(t, fake)

	fire(e, "a", true)
	expectRules(t, fake)

	clear(fake.rejects)
	fire(e, "a", true)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "a", false)
	expectRules(t, fake)
}

func TestUnboundActionIsIgnored(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.apply(app.Action{Name: "nope"})
	expectRules(t, fake)

	e.register(toggled("a"))
	fire(e, "a", true)
	e.apply(app.Action{Name: "nope"})
	expectRules(t, fake, baseBlock("a"))
}

func TestInitially(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(toggled("a").Initially(true))
	e.register(toggled("b").Initially(false))
	e.register(toggled("c").Initially(true))
	expectRules(t, fake)

	e.start()
	expectRules(t, fake, baseBlock("a"), baseBlock("c"))

	fire(e, "a", false)
	expectRules(t, fake, baseBlock("c"))

	fire(e, "b", true)
	expectRules(t, fake, baseBlock("c"), baseBlock("b"))
}

func TestLatestBindingWins(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("a").
		Set(css.Custom("--n", css.Px(1))).
		On("a", func(app.Action) bool { return true }).
		On("b", func(app.Action) bool { return false }))

	e.apply(app.Action{Name: "a"})
	expectRules(t, fake, baseBlock("a"))

	e.apply(app.Action{Name: "b"})
	expectRules(t, fake)

	e.apply(app.Action{Name: "a"})
	expectRules(t, fake, baseBlock("a"))
}

func TestLastBindingOfOneActionWins(t *testing.T) {
	yes := func(app.Action) bool { return true }
	no := func(app.Action) bool { return false }

	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("on").Set(css.Custom("--n", css.Px(1))).On("x", no).On("x", yes))
	e.register(Class("off").Set(css.Custom("--n", css.Px(1))).On("x", yes).On("x", no))

	e.apply(app.Action{Name: "x"})
	expectRules(t, fake, baseBlock("on"))
}

type alpha struct{ on bool }

type beta struct{ on bool }

func TestLastBindingDecidesIncludingItsType(t *testing.T) {
	logged := captureLog(t)
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("ab").
		Set(css.Custom("--n", css.Px(1))).
		Bind(When("x", func(*alpha) bool { return true })).
		Bind(When("x", func(b *beta) bool { return b.on })))

	e.apply(app.Action{Name: "x", Value: &beta{on: true}})
	expectRules(t, fake, baseBlock("ab"))

	e.apply(app.Action{Name: "x", Value: &alpha{on: false}})
	expectRules(t, fake, baseBlock("ab"))

	e.apply(app.Action{Name: "x", Value: &beta{on: false}})
	expectRules(t, fake)

	e.apply(app.Action{Name: "x", Value: &alpha{on: true}})
	expectRules(t, fake)

	lines := logged()
	if len(lines) != 2 {
		t.Fatalf("logged %d messages, want 2: %q", len(lines), lines)
	}
	for _, line := range lines {
		if !strings.Contains(line, "*style.alpha") || !strings.Contains(line, ".ab") {
			t.Errorf("log message %q does not name the unhandled *style.alpha and .ab", line)
		}
	}
}

func TestOnlyTheLastBindingRunsOncePerAction(t *testing.T) {
	var first, second int
	rule := Class("a").
		Set(css.Custom("--n", css.Px(1))).
		On("x", func(app.Action) bool { first++; return true }).
		On("x", func(app.Action) bool { second++; return false })

	e := newEngine(&fakeSheet{})
	e.register(rule)
	e.register(rule)
	e.apply(app.Action{Name: "x"})

	if first != 0 || second != 1 {
		t.Errorf("predicates ran %d and %d times, want 0 and 1", first, second)
	}
}

func TestRegisteringARuleTwiceInsertsItOnce(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	rule := toggled("a").Initially(true)
	e.register(rule)
	e.register(rule)

	e.start()
	fire(e, "a", true)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "a", false)
	expectRules(t, fake)
}

func TestRegisteringARuleTwiceVisitsItOnce(t *testing.T) {
	fake := rejecting(baseBlock("a"))
	e := newEngine(fake)
	rule := toggled("a").Initially(true)
	e.register(rule)
	e.register(rule)

	e.start()
	if fake.attempts != 1 {
		t.Errorf("start() tried to insert the rule %d times, want once", fake.attempts)
	}
}

func TestRunIsNoopOnServer(t *testing.T) {
	Run(Dynamic(toggled("a").Initially(true)))
}

func TestJSSheetCreatesOneStyleElementAndInsertsIntoIt(t *testing.T) {
	logged := captureLog(t)
	dom := newFakeDOM()
	s := &jsSheet{document: dom.document()}

	if got := s.length(); got != 0 {
		t.Errorf("length before the first insert = %d, want 0", got)
	}
	if dom.created != 0 {
		t.Fatalf("%d elements created before the first insert, want none", dom.created)
	}

	for _, text := range []string{baseBlock("a"), baseBlock("b")} {
		if !s.insert(text, s.length()) {
			t.Fatalf("insert(%q) reported failure", text)
		}
	}
	if dom.created != 1 || dom.tag != "style" {
		t.Errorf("created %d %q elements, want exactly one style element", dom.created, dom.tag)
	}
	if !maps.Equal(dom.attrs, map[string]string{"data-webgo": "dynamic"}) {
		t.Errorf("attributes = %v, want only data-webgo=dynamic", dom.attrs)
	}
	if !dom.attached {
		t.Error("the style element was never appended to the head")
	}
	if want := []string{baseBlock("a"), baseBlock("b")}; !slices.Equal(dom.rules, want) || s.length() != 2 {
		t.Errorf("rules = %q, length = %d, want %q and 2", dom.rules, s.length(), want)
	}

	s.remove(0)
	if want := []string{baseBlock("b")}; !slices.Equal(dom.rules, want) || s.length() != 1 {
		t.Errorf("rules after remove(0) = %q, length = %d, want %q and 1", dom.rules, s.length(), want)
	}
	if lines := logged(); len(lines) != 0 {
		t.Errorf("logged %q although nothing failed", lines)
	}
}

func TestJSSheetWithoutASheetFailsSafely(t *testing.T) {
	logged := captureLog(t)
	dom := newFakeDOM()
	dom.noSheet = true
	s := &jsSheet{document: dom.document()}

	for range 2 {
		if s.insert(baseBlock("a"), s.length()) {
			t.Fatal("insert reported success although the browser gave the style element no sheet")
		}
	}
	s.remove(0)

	if got := s.length(); got != 0 {
		t.Errorf("length = %d, want 0", got)
	}
	if dom.created != 1 {
		t.Errorf("%d elements created, want 1: a retry must not append another style element", dom.created)
	}
	lines := logged()
	if len(lines) != 2 {
		t.Fatalf("logged %d messages, want 2: %q", len(lines), lines)
	}
	for _, line := range lines {
		for _, part := range []string{strconv.Quote(baseBlock("a")), errNoSheet.Error()} {
			if !strings.Contains(line, part) {
				t.Errorf("log message %q does not mention %s", line, part)
			}
		}
	}
}

func TestJSSheetInsertSurvivesAJavaScriptException(t *testing.T) {
	logged := captureLog(t)
	dom := newFakeDOM()
	dom.throws = "sheet.insertRule"
	s := &jsSheet{document: dom.document()}

	if s.insert(baseBlock("a"), s.length()) {
		t.Error("insert reported success although the browser threw")
	}
	lines := logged()
	if len(lines) != 1 {
		t.Fatalf("logged %d messages, want 1: %q", len(lines), lines)
	}
	for _, part := range []string{strconv.Quote(baseBlock("a")), "boom"} {
		if !strings.Contains(lines[0], part) {
			t.Errorf("log message %q does not mention %s", lines[0], part)
		}
	}

	dom.throws = ""
	if !s.insert(baseBlock("b"), s.length()) {
		t.Fatal("a later insert failed")
	}
	if want := []string{baseBlock("b")}; dom.created != 1 || !slices.Equal(dom.rules, want) {
		t.Errorf("created %d elements holding %q, want one element holding %q", dom.created, dom.rules, want)
	}
}

func TestJSSheetRemoveAndLengthSurviveAJavaScriptException(t *testing.T) {
	throwing := func(t *testing.T, member string) (*jsSheet, func() []string) {
		t.Helper()
		logged := captureLog(t)
		dom := newFakeDOM()
		s := &jsSheet{document: dom.document()}
		if !s.insert(baseBlock("a"), s.length()) {
			t.Fatal("insert reported failure")
		}
		dom.throws = member
		return s, logged
	}
	expectOneBoom := func(t *testing.T, lines []string) {
		t.Helper()
		if len(lines) != 1 || !strings.Contains(lines[0], "boom") {
			t.Errorf("logged %q, want one message mentioning boom", lines)
		}
	}

	t.Run("deleteRule", func(t *testing.T) {
		s, logged := throwing(t, "sheet.deleteRule")
		s.remove(0)
		expectOneBoom(t, logged())
	})

	t.Run("cssRules", func(t *testing.T) {
		s, logged := throwing(t, "sheet.cssRules")
		if got := s.length(); got != 0 {
			t.Errorf("length = %d, want 0", got)
		}
		expectOneBoom(t, logged())
	})
}

func TestHandlerAppliesThroughDispatch(t *testing.T) {
	ctx, testEngine := newContext(t)
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(toggled("a"))

	e.handler()(ctx, app.Action{Name: "a", Value: true})
	expectRules(t, fake)

	testEngine.ConsumeAll()
	expectRules(t, fake, baseBlock("a"))
}

func TestEndToEndThroughGoApp(t *testing.T) {
	ctx, testEngine := newContext(t)
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("dark").Set(css.Custom("--w", css.Px(1))).On("theme/dark", truthy))
	ctx.Handle("theme/dark", e.handler())

	ctx.NewActionWithValue("theme/dark", true)
	testEngine.ConsumeAll()
	expectRules(t, fake, ".dark {\n\t--w: 1px;\n}")

	ctx.NewActionWithValue("theme/dark", false)
	testEngine.ConsumeAll()
	expectRules(t, fake)
}

func TestNestedRootOccupiesOneSheetIndex(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(withScreen(Class("card").
		Set(css.Custom("--n", css.Px(1))).
		Nest(Class("title").Set(css.Custom("--t", css.Px(3)))).
		On("card", truthy)))
	e.register(toggled("after"))

	fire(e, "card", true)
	fire(e, "after", true)
	expectRules(t, fake,
		".card {\n\t--n: 1px;\n\t@media screen {\n\t\t--m: 2px;\n\t}\n\t& .title {\n\t\t--t: 3px;\n\t}\n}",
		baseBlock("after"),
	)
	if got := e.present[e.rules[0]].indices; !slices.Equal(got, []int{0}) {
		t.Errorf("present indices = %v, want [0]", got)
	}

	fire(e, "card", false)
	expectRules(t, fake, baseBlock("after"))
	if got := e.present[e.rules[1]].indices; !slices.Equal(got, []int{0}) {
		t.Errorf("present indices after removal = %v, want [0]", got)
	}
}

func fireSettings(e *engine, value any) {
	e.apply(app.Action{Name: "settings", Value: value})
}

func accentBlock(name, color string) string {
	return "." + name + " {\n\t--c: " + color + ";\n}"
}

func TestComputeRerendersFromEachReturn(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("accent").Bind(Compute("settings", accent)))

	fireSettings(e, &settings{Accent: "#0A84FF"})
	expectRules(t, fake, accentBlock("accent", "#0A84FF"))

	fireSettings(e, &settings{Accent: "#FF453A"})
	expectRules(t, fake, accentBlock("accent", "#FF453A"))

	fireSettings(e, &settings{Accent: "#FF453A"})
	expectRules(t, fake, accentBlock("accent", "#FF453A"))
}

func TestComputeAbsentRemovesTheRule(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(withScreen(Class("accent").Bind(Compute("settings", accent))))

	fireSettings(e, &settings{Accent: "#0A84FF"})
	expectRules(t, fake, accentBlock("accent", "#0A84FF"), screenBlock("accent"))

	fireSettings(e, &settings{})
	expectRules(t, fake)

	fireSettings(e, (*settings)(nil))
	expectRules(t, fake)
}

func TestWrongTypeLeavesTheRuleUntouched(t *testing.T) {
	cases := []struct {
		name  string
		rule  *Rule
		value any
	}{
		{"Compute", Class("accent").Bind(Compute("settings", accent)), &settings{Accent: "#0A84FF"}},
		{"When", Class("accent").Set(css.Custom("--c", css.Hex("#0A84FF"))).Bind(When("settings", compact)), &settings{Compact: true}},
		{"On before When", Class("accent").Set(css.Custom("--c", css.Hex("#0A84FF"))).On("settings", func(app.Action) bool { return false }).Bind(When("settings", compact)), &settings{Compact: true}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			logged := captureLog(t)
			fake := &fakeSheet{}
			e := newEngine(fake)
			e.register(c.rule)
			e.register(toggled("other"))
			fire(e, "other", true)
			fireSettings(e, c.value)
			want := []string{baseBlock("other"), accentBlock("accent", "#0A84FF")}
			expectRules(t, fake, want...)

			for _, wrong := range []any{"wrong", settings{Compact: true}, nil, true} {
				fireSettings(e, wrong)
				expectRules(t, fake, want...)
			}
			lines := logged()
			if len(lines) != 4 {
				t.Fatalf("logged %d messages, want 4: %q", len(lines), lines)
			}
			for _, part := range []string{`"settings"`, "string", ".accent"} {
				if !strings.Contains(lines[0], part) {
					t.Errorf("log message %q does not mention %s", lines[0], part)
				}
			}
		})
	}
}

func TestKeyframesInsertAndRemove(t *testing.T) {
	pulse := "@keyframes pulse {\n\tfrom {\n\t\topacity: 0;\n\t}\n}"
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Keyframes("pulse").From(css.Opacity(css.Num(0))).On("pulse", truthy).Initially(true))
	e.register(toggled("a"))

	e.start()
	expectRules(t, fake, pulse)

	fire(e, "a", true)
	fire(e, "pulse", false)
	expectRules(t, fake, baseBlock("a"))

	fire(e, "pulse", true)
	fire(e, "pulse", true)
	expectRules(t, fake, baseBlock("a"), pulse)
	if !strings.Contains(fake.rules[1], "@keyframes") {
		t.Errorf("rule %q is not a keyframes rule", fake.rules[1])
	}

	fire(e, "pulse", false)
	expectRules(t, fake, baseBlock("a"))
}

func TestComputedKeyframesRecompute(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Keyframes("fade").Bind(ComputeFrames("settings", fading)))

	fireSettings(e, &settings{})
	expectRules(t, fake, "@keyframes fade {\n\tfrom {\n\t\topacity: 1;\n\t}\n\tto {\n\t\topacity: 0.4;\n\t}\n}")

	fireSettings(e, &settings{Compact: true})
	expectRules(t, fake)
}

func TestComputeKeepsOtherRulesAligned(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	b := withScreen(Class("b").Bind(Compute("settings", accent)))
	e.register(toggled("a"))
	e.register(b)
	e.register(toggled("c"))

	fire(e, "a", true)
	fireSettings(e, &settings{Accent: "#111"})
	fire(e, "c", true)
	expectRules(t, fake, baseBlock("a"), accentBlock("b", "#111"), screenBlock("b"), baseBlock("c"))

	fireSettings(e, &settings{Accent: "#222"})
	expectRules(t, fake, baseBlock("a"), baseBlock("c"), accentBlock("b", "#222"), screenBlock("b"))

	fire(e, "a", false)
	expectRules(t, fake, baseBlock("c"), accentBlock("b", "#222"), screenBlock("b"))
	if got := e.present[b].indices; !slices.Equal(got, []int{1, 2}) {
		t.Errorf("present indices of b = %v, want [1 2]", got)
	}

	fireSettings(e, &settings{Accent: "#333"})
	fire(e, "c", false)
	expectRules(t, fake, accentBlock("b", "#333"), screenBlock("b"))

	fire(e, "a", true)
	fireSettings(e, &settings{})
	expectRules(t, fake, baseBlock("a"))

	fire(e, "a", false)
	expectRules(t, fake)
}

func TestEndToEndTypedBinding(t *testing.T) {
	ctx, testEngine := newContext(t)
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("shell").Set(css.Custom("--p", css.Px(1))).Bind(When("settings", compact)))
	e.register(Class("accent").Bind(Compute("settings", accent)))
	ctx.Handle("settings", e.handler())

	ctx.NewActionWithValue("settings", &settings{Compact: true})
	testEngine.ConsumeAll()
	expectRules(t, fake, ".shell {\n\t--p: 1px;\n}")

	ctx.NewActionWithValue("settings", &settings{Accent: "#0A84FF"})
	testEngine.ConsumeAll()
	expectRules(t, fake, accentBlock("accent", "#0A84FF"))
}

func TestUnchangedRecomputeKeepsItsPosition(t *testing.T) {
	fake := &fakeSheet{}
	e := newEngine(fake)
	a := withScreen(Class("a").Bind(Compute("settings", accent)))
	e.register(a)
	e.register(toggled("b"))

	fireSettings(e, &settings{Accent: "#111"})
	fire(e, "b", true)
	expectRules(t, fake, accentBlock("a", "#111"), screenBlock("a"), baseBlock("b"))
	inserts := fake.attempts

	fireSettings(e, &settings{Accent: "#111"})
	expectRules(t, fake, accentBlock("a", "#111"), screenBlock("a"), baseBlock("b"))
	if got := e.present[a].indices; !slices.Equal(got, []int{0, 1}) {
		t.Errorf("present indices of a = %v, want [0 1]", got)
	}
	if fake.attempts != inserts {
		t.Errorf("an unchanged recompute inserted %d rules, want none", fake.attempts-inserts)
	}

	fireSettings(e, &settings{Accent: "#222"})
	expectRules(t, fake, baseBlock("b"), accentBlock("a", "#222"), screenBlock("a"))
	if got := e.present[a].indices; !slices.Equal(got, []int{1, 2}) {
		t.Errorf("present indices of a = %v, want [1 2]", got)
	}
}

func TestComputeIgnoresLaterMutationOfTheReturnedSlice(t *testing.T) {
	shared := []css.Property{css.Custom("--c", css.Hex("#111"))}
	fake := &fakeSheet{}
	e := newEngine(fake)
	e.register(Class("alias").Bind(Compute("settings", func(*settings) []css.Property { return shared })))

	fireSettings(e, &settings{})
	expectRules(t, fake, accentBlock("alias", "#111"))

	shared[0] = css.Custom("--c", css.Hex("#222"))
	expectRules(t, fake, accentBlock("alias", "#111"))
	if got := e.present[e.rules[0]].texts; !slices.Equal(got, []string{accentBlock("alias", "#111")}) {
		t.Errorf("recorded texts = %q, want the text rendered when the action fired", got)
	}

	fireSettings(e, &settings{})
	expectRules(t, fake, accentBlock("alias", "#222"))
}
