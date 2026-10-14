package styling

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/Toms223/WebGo/docs/pages/styling/demos"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	Title = "Styling"
	Route = "/styling"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "styling",
		Path:  Route,
		Title: Title,
		Lead:  "Properties, keywords, selectors, queries and at-rules as Go values the compiler checks.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "Rules", Body: rules},
			{Title: "Selectors", Body: selectors},
			{Title: "Pseudo-classes", Body: pseudoClasses},
			{Title: "Nesting", Body: nesting},
			{Title: "Values", Body: values},
			{Title: "Keywords", Body: keywords},
			{Title: "Media", Body: mediaQueries},
			{Title: "Container", Body: containerQueries},
			{Title: "Supports", Body: supportsQueries},
			{Title: "Keyframes", Body: keyframes},
			{Title: "Font faces", Body: fontFaces},
			{Title: "Properties", Body: properties},
			{Title: "Layers and imports", Body: layers},
			{Title: "Important", Body: important},
			{Title: "Generated code", Body: generated},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("The `style` package describes CSS as Go values. The `css`, `kw`, `media`, `container` and `supports` packages supply the typed values and queries those rules are made of. A property that does not exist or a misspelled keyword fails to compile, and so does a value of the wrong kind wherever the property's grammar can be typed exactly (see ", doc.A("#values", "Values"), ")."),
		doc.P("This page covers writing rules. The ", doc.A("/applying", "Applying"), " page covers delivering them: stylesheets, the handler, the browser engine, actions and errors."),
		doc.List(
			doc.Item("`style` holds `Rule` and the at-rule types, `Style` and `Resolve`, and `Run`."),
			doc.Item("`serve` holds `Handler`, which turns a list of styles into go-app's handler."),
			doc.Item("`css` holds the value types (`Px`, `Rem`, `Hex`, `Var`, `Calc`, `Raw` and the rest), one function per CSS property, such as `css.Padding`, and one function per at-rule descriptor, such as `css.Src`."),
			doc.Item("`kw` holds one constant per CSS keyword and per named color, such as `kw.Underline` and `kw.Rebeccapurple`."),
			doc.Item("`media`, `container` and `supports` hold the queries of `@media`, `@container` and `@supports`."),
		),
		doc.P("The import paths are `github.com/Toms223/WebGo/style`, `github.com/Toms223/WebGo/serve`, `github.com/Toms223/WebGo/css`, `github.com/Toms223/WebGo/kw`, `github.com/Toms223/WebGo/media`, `github.com/Toms223/WebGo/container` and `github.com/Toms223/WebGo/supports`. None of them imports `page`, `data` or `markdown`, so they work in any go-app application."),
		doc.P("Every example on this page is rendered live by the rule shown. Its CSS tab is that rule's `CSS()` output."),
	}
}

func rules() []app.UI {
	return []app.UI{
		doc.P("A `Rule` is one selector and an ordered list of blocks of properties. It has no exported fields. It is normally a package-level variable, and its `Name()` is the class or id you pass to a go-app element."),
		doc.Go(`func (r *Rule) Set(props ...css.Property) *Rule
func (r *Rule) Media(q media.Query, props ...css.Property) *Rule
func (r *Rule) Container(q container.Query, props ...css.Property) *Rule
func (r *Rule) Supports(q supports.Query, props ...css.Property) *Rule
func (r *Rule) Nest(children ...*Rule) *Rule
func (r *Rule) CSS() string
func (r *Rule) Err() error`),
		doc.Example(demos.CardExample),
		doc.P("The element is `app.Div().Class(Card.Name())`."),
		doc.List(
			doc.Item("`Set` adds properties to the base block, the one written under the bare selector."),
			doc.Item("A pseudo method adds properties to a block for the selector plus that pseudo-class or pseudo-element: `Hover` writes `.demo-card:hover`."),
			doc.Item("`Media`, `Container` and `Supports` add properties to a block wrapped in that at-rule. The block holds the rule's own selector: `@media print { .demo-card { … } }`."),
			doc.Item("`Nest` puts other rules inside this one. See ", doc.A("#nesting", "Nesting"), "."),
			doc.Item("`Set`, `Media`, `Container`, `Supports`, `Nest`, the pseudo methods and the binding methods change the rule they are called on and return it, so calls chain. The combinators return a new rule."),
		),
		doc.H("Block order"),
		doc.List(
			doc.Item("A rule owns the base block, then one block per pseudo in the order each was first called, then the `Container` blocks, the `Supports` blocks and the `Media` blocks, each kind in the order its query was first called."),
			doc.Item("Calling the same pseudo method, or a conditional method with the same query, again appends to that block."),
			doc.Item("The writer emits blocks in that order, whatever order the calls were made in."),
			doc.Item("A block with no properties is not written."),
			doc.Item("Properties are written in call order. Duplicates are written as given, and CSS's last-wins rule does the rest. A nil `css.Property` is skipped."),
			doc.Item("Pseudo and conditional blocks apply to the subject of the rule, its rightmost compound selector: `.doc a:hover`, never `.doc:hover a`."),
		),
		doc.H("CSS"),
		doc.P("`CSS` returns the rule's full text, every block, exactly as the writer emits it into a stylesheet. Blocks are separated by one blank line and the text ends with a newline. A rule without properties returns `\"\"`."),
		doc.P("`Err` returns the first error any call latched. A rule never panics on bad input. The errors are listed on the Applying page under ", doc.A("/applying#errors", "Errors"), "."),
	}
}

func selectors() []app.UI {
	return []app.UI{
		doc.Go(`func Class(name string) *Rule
func Id(name string) *Rule
func Element(tag string) *Rule
func Any() *Rule`),
		doc.P("Each constructor returns a new rule with no properties:"),
		doc.List(
			doc.Item("`Class(\"card\")` selects `.card`."),
			doc.Item("`Id(\"nav\")` selects `#nav`."),
			doc.Item("`Element(\"a\")` selects `a`."),
			doc.Item("`Any()` selects `*`."),
		),
		doc.P("`Class(\"\")`, `Id(\"\")` and `Element(\"\")` give a rule whose error is `ErrNoName`."),
		doc.H("Combinators"),
		doc.Go(`func (r *Rule) Descendant(o *Rule) *Rule
func (r *Rule) Child(o *Rule) *Rule
func (r *Rule) Next(o *Rule) *Rule
func (r *Rule) Sibling(o *Rule) *Rule
func (r *Rule) And(o *Rule) *Rule`),
		doc.Example(demos.SelectorsExample),
		doc.P("`List` has no properties. It exists only to build `ListItem`, and only `ListItem` goes into a stylesheet."),
		doc.List(
			doc.Item("`Descendant` selects `.doc a`: an `a` anywhere inside `.doc`."),
			doc.Item("`Child` selects `.doc > *`: an element directly inside `.doc`."),
			doc.Item("`Next` selects `.card + .card`: a `.card` immediately after a `.card`."),
			doc.Item("`Sibling` selects `.card ~ .card`: a `.card` anywhere after a `.card`."),
			doc.Item("`And` selects `.card.active` or `a.nav-link`: both selectors on the same element."),
		),
		doc.P("A combinator returns a new rule whose selector is the receiver's selector, the combinator and the argument's selector. Neither operand is changed and neither has to be in any stylesheet. The properties, bindings and `Initially` setting of the operands are not carried over, so the new rule starts empty."),
		doc.P("Chains nest to any depth: `style.Class(\"doc\").Child(style.Element(\"ul\")).Descendant(style.Element(\"a\"))` selects `.doc > ul a`."),
		doc.P("`And` joins two selectors into one compound selector:"),
		doc.List(
			doc.Item("At most one of the two may be an element or `Any`. When there is one, it is written first: `a.nav-link`. `Any()` is dropped when it is joined with an id or a class: `style.Any().And(style.Id(\"main\"))` selects `#main`."),
			doc.Item("The result writes the element, then the id, then the classes in call order: `style.Class(\"a\").And(style.Id(\"b\")).And(style.Element(\"li\"))` selects `li#b.a`."),
			doc.Item("When the receiver is a chain, `And` joins onto its last selector: `style.Class(\"doc\").Child(style.Class(\"x\")).And(style.Class(\"y\"))` selects `.doc > .x.y`."),
			doc.Item("The argument must be a single selector. Joining two selectors that are each an element or `Any`, joining two ids, or passing an argument built with a combinator is `ErrBadSelector`."),
		),
		doc.P("A selector list such as `.a, .b { … }` has no form. Write one rule per selector and share the property slice between them."),
		doc.H("Name and Selector"),
		doc.Go(`func (r *Rule) Name() string
func (r *Rule) Selector() string`),
		doc.P("`Name` returns the class of a `Class` rule and the id of an `Id` rule: `app.Div().Class(card.Name())`, `app.Nav().ID(menu.Name())`. It returns `\"\"` for an element rule, for `Any`, for `Self()`, and for any rule built with a combinator or `And`: those are matched by structure, not applied by name."),
		doc.P("`Selector` returns the selector text:"),
		doc.List(
			doc.Item("`style.Class(\"card\").Selector()` is `.card`."),
			doc.Item("`style.Id(\"nav\").Selector()` is `#nav`."),
			doc.Item("`style.Element(\"a\").Selector()` is `a`."),
			doc.Item("`style.Class(\"doc\").Child(style.Any()).Selector()` is `.doc > *`."),
			doc.Item("`style.Self().Selector()` is `&`."),
		),
	}
}

func pseudoClasses() []app.UI {
	return []app.UI{
		doc.P("Each pseudo-class and pseudo-element is a method on `Rule`, generated from the same data as the properties. Most take only properties:"),
		doc.Go(`func (r *Rule) Hover(props ...css.Property) *Rule
func (r *Rule) Active(props ...css.Property) *Rule
func (r *Rule) FocusVisible(props ...css.Property) *Rule
func (r *Rule) FirstChild(props ...css.Property) *Rule
func (r *Rule) Before(props ...css.Property) *Rule
func (r *Rule) After(props ...css.Property) *Rule`),
		doc.Example(demos.PseudoExample),
		doc.P("A method name is the pseudo's name in CamelCase: `:focus-visible` is `FocusVisible`, `::before` is `Before`. The functional pseudos with simple arguments take that argument first:"),
		doc.Go(`func (r *Rule) NthChild(n css.Nth, props ...css.Property) *Rule
func (r *Rule) Lang(tag string, props ...css.Property) *Rule
func (r *Rule) Dir(d DirValue, props ...css.Property) *Rule
func (r *Rule) Highlight(name string, props ...css.Property) *Rule`),
		doc.List(
			doc.Item("`NthChild`, `NthLastChild`, `NthOfType` and `NthLastOfType` take a `css.Nth`: `css.An(2, 1)` is `2n+1`, `css.Odd` is `odd`, `css.Even` is `even`. The ", doc.A("#selectors", "Selectors"), " example uses `NthChild(css.Odd, …)`."),
			doc.Item("`Lang` takes the language tag as a string and writes `:lang(tag)`."),
			doc.Item("`Dir` takes a `DirValue`, which `kw.Ltr` and `kw.Rtl` satisfy, and writes `:dir(ltr)` or `:dir(rtl)`."),
			doc.Item("`Highlight` takes the name of a custom highlight as a string and writes `::highlight(name)`."),
		),
		doc.P("Pseudos that take a selector as their argument, such as `:not()`, `:is()`, `:where()` and `:has()`, have no method. `supports.Selector` takes raw selector text for `@supports selector(…)`."),
		doc.P("A pseudo method called with no properties opens its block without declarations. On a `Self()` rule that block can hold conditional blocks. See ", doc.A("#nesting", "Nesting"), "."),
	}
}

func nesting() []app.UI {
	return []app.UI{
		doc.Go(`func (r *Rule) Nest(children ...*Rule) *Rule
func Self() *Rule`),
		doc.P("`Nest` writes CSS nesting syntax: each child goes inside the parent's braces."),
		doc.Example(demos.NestingExample),
		doc.List(
			doc.Item("A child is written inside the parent as `& <child selector> { … }`. Its own pseudo and conditional blocks are written one level deeper. A child may itself nest."),
			doc.Item("A child with no base declarations, exactly one pseudo block, no conditional block and no children is written as one selector: `& a:hover { … }`."),
			doc.Item("`Self()` is the parent itself. `Self().Hover(…)` writes `&:hover { … }`."),
			doc.Item("A `Self()` child with only conditional blocks writes them directly inside the parent, with no `& { }` wrapper: the `@media` block above."),
			doc.Item("A child built with `And` on the parent renders `&.on`: `card.Nest(card.And(style.Class(\"on\")))` writes `&.on { … }`."),
			doc.Item("A combinator child built from the parent drops the parent's part: `card.Nest(card.Child(style.Element(\"a\")))` writes `& > a { … }`. Any other child selector gets `& ` in front."),
			doc.Item("`Self()` outside `Nest`, or `Self()` with `Set`, is `ErrBadSelector`."),
			doc.Item("A nested child is claimed by its parent. Its `Name()` is unchanged, so `PanelTitle.Name()` still gives the class for the element. Passing a claimed child to `Generated`, `Dynamic`, another `Nest`, or a combinator is `ErrNested`."),
			doc.Item("A nested child cannot carry bindings: `ErrNestedBinding`. Bindings belong to the root."),
			doc.Item("A rule with no children is written flat, as `.x:hover { … }` and a top-level `@media` block."),
		),
		doc.H("Conditional blocks under a pseudo"),
		doc.P("`Media`, `Container` and `Supports` called after a pseudo method on a `Self()` rule attach under that pseudo:"),
		doc.Example(demos.HoverMediaExample),
		doc.P("With an empty `Hover()`, `style.Self().Hover().Media(q, …)` writes `&:hover { @media … { … } }` with no declarations of its own under `&:hover`."),
	}
}
