The `style` package describes CSS as Go values. The `css`, `kw`, `media`, `container` and `supports` packages supply the typed values and queries those rules are made of. A property that does not exist or a misspelled keyword fails to compile, and so does a value of the wrong kind wherever the property's grammar can be typed exactly (see Values).

This page covers writing rules. The Applying page covers delivering them: stylesheets, the handler, the browser engine, actions and errors.

- `style` holds `Rule` and the at-rule types, `Style` and `Resolve`, and `Run`.
- `serve` holds `Handler`, which turns a list of styles into go-app's handler.
- `css` holds the value types (`Px`, `Rem`, `Hex`, `Var`, `Calc`, `Raw` and the rest), one function per CSS property, such as `css.Padding`, and one function per at-rule descriptor, such as `css.Src`.
- `kw` holds one constant per CSS keyword and per named color, such as `kw.Underline` and `kw.Rebeccapurple`.
- `media`, `container` and `supports` hold the queries of `@media`, `@container` and `@supports`.

The import paths are `github.com/Toms223/WebGo/style`, `github.com/Toms223/WebGo/serve`, `github.com/Toms223/WebGo/css`, `github.com/Toms223/WebGo/kw`, `github.com/Toms223/WebGo/media`, `github.com/Toms223/WebGo/container` and `github.com/Toms223/WebGo/supports`. None of them imports `page`, `data` or `markdown`, so they work in any go-app application.

Every example with a Result panel is rendered live on this page by the rule shown. The CSS under an example is that rule's `CSS()` output, pinned by a test.

## Rules

A `Rule` is one selector and an ordered list of blocks of properties. It has no exported fields. It is normally a package-level variable, and its `Name()` is the class or id you pass to a go-app element.

```go
func (r *Rule) Set(props ...css.Property) *Rule
func (r *Rule) Media(q media.Query, props ...css.Property) *Rule
func (r *Rule) Container(q container.Query, props ...css.Property) *Rule
func (r *Rule) Supports(q supports.Query, props ...css.Property) *Rule
func (r *Rule) Nest(children ...*Rule) *Rule
func (r *Rule) CSS() string
func (r *Rule) Err() error
```

```go card
var Card = style.Class("demo-card").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Hover(css.Background(css.Var("--pink")))
```

```css
.demo-card {
	padding: 1rem;
	background: var(--mint);
}

.demo-card:hover {
	background: var(--pink);
}
```

The element is `app.Div().Class(Card.Name())`.

- `Set` adds properties to the base block, the one written under the bare selector.
- A pseudo method adds properties to a block for the selector plus that pseudo-class or pseudo-element: `Hover` writes `.demo-card:hover`.
- `Media`, `Container` and `Supports` add properties to a block wrapped in that at-rule. The block holds the rule's own selector: `@media print { .demo-card { … } }`.
- `Nest` puts other rules inside this one. See Nesting.
- `Set`, `Media`, `Container`, `Supports`, `Nest`, the pseudo methods and the binding methods change the rule they are called on and return it, so calls chain. The combinators return a new rule.

### Block order

- A rule owns the base block, then one block per pseudo in the order each was first called, then the `Container` blocks, the `Supports` blocks and the `Media` blocks, each kind in the order its query was first called.
- Calling the same pseudo method, or a conditional method with the same query, again appends to that block.
- The writer emits blocks in that order, whatever order the calls were made in.
- A block with no properties is not written.
- Properties are written in call order. Duplicates are written as given, and CSS's last-wins rule does the rest. A nil `css.Property` is skipped.
- Pseudo and conditional blocks apply to the subject of the rule, its rightmost compound selector: `.doc a:hover`, never `.doc:hover a`.

### CSS

`CSS` returns the rule's full text, every block, exactly as the writer emits it into a stylesheet. Blocks are separated by one blank line and the text ends with a newline. A rule without properties returns `""`.

`Err` returns the first error any call latched. A rule never panics on bad input. The errors are listed on the Applying page under Errors.

## Selectors

```go
func Class(name string) *Rule
func Id(name string) *Rule
func Element(tag string) *Rule
func Any() *Rule
```

Each constructor returns a new rule with no properties:

- `Class("card")` selects `.card`.
- `Id("nav")` selects `#nav`.
- `Element("a")` selects `a`.
- `Any()` selects `*`.

`Class("")`, `Id("")` and `Element("")` give a rule whose error is `ErrNoName`.

### Combinators

```go
func (r *Rule) Descendant(o *Rule) *Rule
func (r *Rule) Child(o *Rule) *Rule
func (r *Rule) Next(o *Rule) *Rule
func (r *Rule) Sibling(o *Rule) *Rule
func (r *Rule) And(o *Rule) *Rule
```

```go selectors
var List = style.Class("demo-list")

var ListItem = List.Child(style.Element("li")).
	Set(css.Padding(css.Rem(0.25), css.Rem(0.75))).
	NthChild(css.Odd, css.Background(css.Var("--mint")))
```

```css
.demo-list > li {
	padding: 0.25rem 0.75rem;
}

.demo-list > li:nth-child(odd) {
	background: var(--mint);
}
```

`List` has no properties. It exists only to build `ListItem`, and only `ListItem` goes into a stylesheet.

- `Descendant` selects `.doc a`: an `a` anywhere inside `.doc`.
- `Child` selects `.doc > *`: an element directly inside `.doc`.
- `Next` selects `.card + .card`: a `.card` immediately after a `.card`.
- `Sibling` selects `.card ~ .card`: a `.card` anywhere after a `.card`.
- `And` selects `.card.active` or `a.nav-link`: both selectors on the same element.

A combinator returns a new rule whose selector is the receiver's selector, the combinator and the argument's selector. Neither operand is changed and neither has to be in any stylesheet. The properties, bindings and `Initially` setting of the operands are not carried over, so the new rule starts empty.

Chains nest to any depth: `style.Class("doc").Child(style.Element("ul")).Descendant(style.Element("a"))` selects `.doc > ul a`.

`And` joins two selectors into one compound selector:

- At most one of the two may be an element or `Any`. When there is one, it is written first: `a.nav-link`. `Any()` is dropped when it is joined with an id or a class: `style.Any().And(style.Id("main"))` selects `#main`.
- The result writes the element, then the id, then the classes in call order: `style.Class("a").And(style.Id("b")).And(style.Element("li"))` selects `li#b.a`.
- When the receiver is a chain, `And` joins onto its last selector: `style.Class("doc").Child(style.Class("x")).And(style.Class("y"))` selects `.doc > .x.y`.
- The argument must be a single selector. Joining two selectors that are each an element or `Any`, joining two ids, or passing an argument built with a combinator is `ErrBadSelector`.

A selector list such as `.a, .b { … }` has no form. Write one rule per selector and share the property slice between them.

### Name and Selector

```go
func (r *Rule) Name() string
func (r *Rule) Selector() string
```

`Name` returns the class of a `Class` rule and the id of an `Id` rule: `app.Div().Class(card.Name())`, `app.Nav().ID(menu.Name())`. It returns `""` for an element rule, for `Any`, for `Self()`, and for any rule built with a combinator or `And`: those are matched by structure, not applied by name.

`Selector` returns the selector text:

- `style.Class("card").Selector()` is `.card`.
- `style.Id("nav").Selector()` is `#nav`.
- `style.Element("a").Selector()` is `a`.
- `style.Class("doc").Child(style.Any()).Selector()` is `.doc > *`.
- `style.Self().Selector()` is `&`.

## Pseudo-classes

Each pseudo-class and pseudo-element is a method on `Rule`, generated from the same data as the properties. Most take only properties:

```go
func (r *Rule) Hover(props ...css.Property) *Rule
func (r *Rule) Active(props ...css.Property) *Rule
func (r *Rule) FocusVisible(props ...css.Property) *Rule
func (r *Rule) FirstChild(props ...css.Property) *Rule
func (r *Rule) Before(props ...css.Property) *Rule
func (r *Rule) After(props ...css.Property) *Rule
```

```go pseudo
var Button = style.Class("demo-button").
	Set(
		css.Padding(css.Rem(0.5), css.Rem(1)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
		css.Background(css.Var("--surface")),
	).
	Before(css.Content(css.String("→ "))).
	Hover(css.Background(css.Var("--side"))).
	Active(css.Transform(css.Raw("translate(2px, 2px)"))).
	FocusVisible(css.Outline(css.Px(3), kw.Dashed, css.Var("--accent")))
```

```css
.demo-button {
	padding: 0.5rem 1rem;
	border: 3px solid var(--ink);
	background: var(--surface);
}

.demo-button::before {
	content: "→ ";
}

.demo-button:hover {
	background: var(--side);
}

.demo-button:active {
	transform: translate(2px, 2px);
}

.demo-button:focus-visible {
	outline: 3px dashed var(--accent);
}
```

A method name is the pseudo's name in CamelCase: `:focus-visible` is `FocusVisible`, `::before` is `Before`. The functional pseudos with simple arguments take that argument first:

```go
func (r *Rule) NthChild(n css.Nth, props ...css.Property) *Rule
func (r *Rule) Lang(tag string, props ...css.Property) *Rule
func (r *Rule) Dir(d DirValue, props ...css.Property) *Rule
func (r *Rule) Highlight(name string, props ...css.Property) *Rule
```

- `NthChild`, `NthLastChild`, `NthOfType` and `NthLastOfType` take a `css.Nth`: `css.An(2, 1)` is `2n+1`, `css.Odd` is `odd`, `css.Even` is `even`. The Selectors example uses `NthChild(css.Odd, …)`.
- `Lang` takes the language tag as a string and writes `:lang(tag)`.
- `Dir` takes a `DirValue`, which `kw.Ltr` and `kw.Rtl` satisfy, and writes `:dir(ltr)` or `:dir(rtl)`.
- `Highlight` takes the name of a custom highlight as a string and writes `::highlight(name)`.

Pseudos that take a selector as their argument, such as `:not()`, `:is()`, `:where()` and `:has()`, have no method. `supports.Selector` takes raw selector text for `@supports selector(…)`.

A pseudo method called with no properties opens its block without declarations. On a `Self()` rule that block can hold conditional blocks. See Nesting.

## Nesting

```go
func (r *Rule) Nest(children ...*Rule) *Rule
func Self() *Rule
```

`Nest` writes CSS nesting syntax: each child goes inside the parent's braces.

```go nesting
var PanelTitle = style.Class("demo-panel-title").
	Set(css.FontWeight(css.Num(900)))

var Panel = style.Class("demo-panel").
	Set(css.Padding(css.Rem(1)), css.Border(css.Px(3), kw.Solid, css.Var("--ink"))).
	Nest(
		PanelTitle,
		style.Element("a").Hover(css.Background(css.Var("--side"))),
		style.Self().Hover(css.Background(css.Var("--mint"))),
		style.Self().Media(media.MaxWidth(css.Rem(48)), css.Padding(css.Rem(0.5))),
	)
```

```css
.demo-panel {
	padding: 1rem;
	border: 3px solid var(--ink);
	& .demo-panel-title {
		font-weight: 900;
	}
	& a:hover {
		background: var(--side);
	}
	&:hover {
		background: var(--mint);
	}
	@media (max-width: 48rem) {
		padding: 0.5rem;
	}
}
```

- A child is written inside the parent as `& <child selector> { … }`. Its own pseudo and conditional blocks are written one level deeper. A child may itself nest.
- A child with no base declarations, exactly one pseudo block, no conditional block and no children is written as one selector: `& a:hover { … }`.
- `Self()` is the parent itself. `Self().Hover(…)` writes `&:hover { … }`.
- A `Self()` child with only conditional blocks writes them directly inside the parent, with no `& { }` wrapper: the `@media` block above.
- A child built with `And` on the parent renders `&.on`: `card.Nest(card.And(style.Class("on")))` writes `&.on { … }`.
- A combinator child built from the parent drops the parent's part: `card.Nest(card.Child(style.Element("a")))` writes `& > a { … }`. Any other child selector gets `& ` in front.
- `Self()` outside `Nest`, or `Self()` with `Set`, is `ErrBadSelector`.
- A nested child is claimed by its parent. Its `Name()` is unchanged, so `PanelTitle.Name()` still gives the class for the element. Passing a claimed child to `Generated`, `Dynamic`, another `Nest`, or a combinator is `ErrNested`.
- A nested child cannot carry bindings: `ErrNestedBinding`. Bindings belong to the root.
- A rule with no children is written flat, as `.x:hover { … }` and a top-level `@media` block.

### Conditional blocks under a pseudo

`Media`, `Container` and `Supports` called after a pseudo method on a `Self()` rule attach under that pseudo:

```go hover-media
var Tile = style.Class("demo-tile").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Nest(
		style.Self().
			Hover(css.Background(css.Var("--pink"))).
			Media(media.MaxWidth(css.Rem(48)), css.Background(css.Var("--side"))),
	)
```

```css
.demo-tile {
	padding: 1rem;
	background: var(--mint);
	&:hover {
		background: var(--pink);
		@media (max-width: 48rem) {
			background: var(--side);
		}
	}
}
```

With an empty `Hover()`, `style.Self().Hover().Media(q, …)` writes `&:hover { @media … { … } }` with no declarations of its own under `&:hover`.

## Values

```go
type Value interface{ CSS() string }

type Property interface {
	Name() string
	Value() string
	Prefixes() []string
	Important() Property
	property()
}
```

A `Value` is anything that renders as CSS text. A `Property` is one declaration: a name, a value and the vendor prefixes it needs. `Property` has an unexported method, so only the `css` package makes properties. Every property function returns one, and `Rule.Set` and the pseudo methods take them. `Important` is described under Important.

### Kinds

- `Length`: `Px`, `Rem`, `Em`, `Ch`, `Ex`, `Vw`, `Vh`, `Vmin`, `Vmax`, `Dvh`, `Svh`, `Lvh`, `Dvw`, `Svw`, `Lvw`, `Cm`, `Mm`, `In`, `Pt`, `Pc` and `Q`. Each takes a `float64`: `css.Px(8)` is `8px`.
- `Percentage`: `Pct(50)` is `50%`.
- `Number`: `Num(1.5)` is `1.5`. `Integer`: `Int(3)` is `3`.
- `Angle`: `Deg`, `Rad`, `Grad` and `Turn`: `Deg(90)` is `90deg`.
- `Time`: `S` and `Ms`: `Ms(250)` is `250ms`.
- `Resolution`: `Dpi`, `Dpcm` and `Dppx`: `Dppx(2)` is `2dppx`.
- `Fraction`: `Fr(1)` is `1fr`.
- `RatioValue`: `Ratio(16, 9)` is `16/9`.
- `ColorLiteral`: `Hex`, `RGB`, `RGBA`, `HSL` and `HSLA`. Named colors are keywords in `kw`.
- `StringValue`: `String`. `UrlValue`: `Url`. `IdentValue`: `Ident`, written as given.
- `VarValue`: `Var`. `CalcValue`: `Calc`, `Min`, `Max` and `Clamp`. `RawValue`: `Raw`.
- `Nth`: `An`, `Odd` and `Even`, for the `Nth…` pseudo methods.

The color functions:

```go
func Hex(s string) ColorLiteral
func RGB(r, g, b uint8) ColorLiteral
func RGBA(r, g, b uint8, a float64) ColorLiteral
func HSL(h, s, l float64) ColorLiteral
func HSLA(h, s, l, a float64) ColorLiteral
```

- `Hex` adds the `#` when it is missing: `Hex("fff")` and `Hex("#fff")` are both `#fff`.
- `RGB(255, 0, 0)` is `rgb(255, 0, 0)` and `RGBA(255, 0, 0, 0.5)` is `rgba(255, 0, 0, 0.5)`.
- `HSL(200, 50, 40)` is `hsl(200, 50%, 40%)`: saturation and lightness are plain numbers, written as percentages. `HSLA(200, 50, 40, 0.5)` is `hsla(200, 50%, 40%, 0.5)`.

Numbers are written in their shortest round-trip form, without an exponent. A zero keeps its unit: `css.Px(0)` is `0px`.

### Property functions

Each generated CSS property is one function in `css`, named after the property in CamelCase: `css.Padding`, `css.BackgroundColor`. What a function accepts depends on its tier:

- Tier 1, exact. The property's grammar is made of keywords, the kinds above and simple repetition. The generator emits a marker interface for it, named after the property with a `Value` suffix, and the function takes it. The interface is satisfied by the kinds the grammar accepts, the keywords the grammar names, the CSS-wide keywords `inherit`, `initial`, `unset`, `revert` and `revert-layer`, and, when the grammar accepts a `ColorLiteral`, every named color, `currentcolor` and `transparent`. Any other kind or keyword is a compile error.
- Tier 2, loose. The grammar is more than keywords, the kinds above and simple repetition: that of `border-radius`, for one, has a `/` between two repeated groups. The function takes `(first css.Value, rest ...css.Value)` and accepts any `Value`. It writes the values separated by commas when the top level of the property's grammar is a comma-separated list, as `font-family`'s is, a choice whose every alternative other than a keyword is such a list, as `text-shadow`'s `none | <shadow>#` is, or a sequence with a literal comma between its parts, as `background`'s `<bg-layer>#? , <final-bg-layer>` is, and separated by spaces otherwise. A value that needs any other arrangement goes in one `css.Raw`.

```go
func Width(v WidthValue) Property
func Padding(first PaddingValue, rest ...PaddingValue) Property
```

`Width` takes one value. `Padding` is a repeated property, `{1,4}` in its grammar, so it takes a first value and more after it, and writes them separated by a space. A property whose grammar repeats with `#` writes its values separated by commas. The compiler does not enforce the repeat count. `css.Width(kw.Auto)` and `css.Width(css.Px(10))` compile. `css.Width(kw.Flex)` does not: `flex` is not among the keywords of the `width` grammar. A property's signature shows its tier: `css.Width`, `css.Padding` and `css.Color` are tier 1, and `css.BorderRadius`, `css.TextDecoration` and `css.Flex` are tier 2.

Tier 1 functions also accept `css.Var` and `css.Raw`, and accept `css.Calc` where the property takes a length, percentage, number, integer, angle, time or fraction.

A property that needs vendor prefixes carries them in the `Property` it returns, and the writer puts each prefixed declaration before the unprefixed one.

### Raw

```go
func Raw(s string) RawValue
```

`Raw` writes its text as given. It is accepted by every property function and is the escape hatch for a value the typed grammar does not cover: `css.Width(css.Raw("fit-content(20em)"))` writes `width: fit-content(20em);`. `Raw` is not checked.

### Var

```go
func Var(name string, fallback ...Value) VarValue
```

`Var` writes a `var()` reference. It adds the `--` when it is missing, and joins several fallback values with `, `:

- `Var("--ink")` is `var(--ink)`.
- `Var("ink")` is `var(--ink)`.
- `Var("--ink", Hex("#000"))` is `var(--ink, #000)`.
- `Var("--a", Px(1), Pct(2))` is `var(--a, 1px, 2%)`.

### Calc

```go
func Calc(e Expr) CalcValue
func Min(first Expr, rest ...Expr) CalcValue
func Max(first Expr, rest ...Expr) CalcValue
func Clamp(lo, preferred, hi Expr) CalcValue
func Add(a, b Expr) Expr
func Sub(a, b Expr) Expr
func Mul(a Expr, n Number) Expr
func Div(a Expr, n Number) Expr
```

`Expr` has an unexported method. It is implemented by `Length`, `Percentage`, `Number`, `Integer`, `Angle`, `Time`, `Fraction`, `VarValue`, `CalcValue` and the results of `Add`, `Sub`, `Mul` and `Div`.

- `Calc(Sub(Pct(100), Px(20)))` is `calc(100% - 20px)`.
- `Calc(Mul(Add(Pct(100), Px(2)), Num(2)))` is `calc((100% + 2px) * 2)`. The result of `Add`, `Sub`, `Mul` or `Div` used as the left operand of `Mul` or `Div`, or as the right operand of `Sub`, is wrapped in parentheses.
- `Calc(Div(Var("--w"), Num(3)))` is `calc(var(--w) / 3)`.
- `Min(Px(10), Pct(5))` is `min(10px, 5%)`, and `Max` is the same with `max`.
- `Clamp(Rem(1), Vw(2), Rem(3))` is `clamp(1rem, 2vw, 3rem)`.

### Custom

```go
func Custom(name string, v Value) Property
```

`Custom` defines a custom property. It adds the `--` when it is missing, and it carries no vendor prefixes. Read it back with `Var`:

```go values
var Swatch = style.Class("demo-swatch").
	Set(
		css.Custom("demo-size", css.Rem(4)),
		css.Width(css.Var("--demo-size")),
		css.Height(css.Calc(css.Div(css.Var("--demo-size"), css.Num(2)))),
		css.Background(css.HSL(200, 80, 60)),
		css.Border(css.Px(3), kw.Solid, css.Var("--ink")),
	)
```

```css
.demo-swatch {
	--demo-size: 4rem;
	width: var(--demo-size);
	height: calc(var(--demo-size) / 2);
	background: hsl(200, 80%, 60%);
	border: 3px solid var(--ink);
}
```

### Strings

```go
func String(s string) StringValue
func Url(s string) UrlValue
```

`String` writes a double-quoted CSS string, and `Url` writes `url("…")` with the same quoting. Inside the quotes:

- a backslash is written `\\`;
- a double quote is written `\"`;
- a newline is written `\a `, a carriage return `\d ` and a form feed `\c `, each followed by the space that ends the escape.

So `css.String("→")` is `"→"`, a string holding a double quote between two letters is written `"a\"b"`, a string holding a newline between two letters is written `"a\a b"`, and `css.Url("x.png")` is `url("x.png")`.

## Keywords

The `kw` package has one constant for every keyword that appears in the grammar of a generated property, descriptor or query feature, plus every named color, `currentcolor` and `transparent`, the CSS-wide keywords, and `ltr` and `rtl` for `Rule.Dir`. The name is the keyword in CamelCase: `inline-flex` is `kw.InlineFlex`, `x-small` is `kw.XSmall` and `rebeccapurple` is `kw.Rebeccapurple`. The feature and descriptor keywords are here too: `kw.Dark`, `kw.Portrait`, `kw.Woff2` and `kw.Swap`.

Each keyword is a constant of its own type: `kw.InlineFlex` has the unexported type `kInlineFlex`, whose underlying type is `string`. A tier 1 property function accepts the keywords its grammar names, and each keyword has its own type, so `kw.Flex` is not accepted where only `kw.Auto` is. `css.Display(kw.InlineFlex)` and `css.Width(kw.Auto)` compile. `css.Width(kw.Flex)` does not. Every tier 1 function also accepts the CSS-wide keywords `kw.Inherit`, `kw.Initial`, `kw.Unset`, `kw.Revert` and `kw.RevertLayer`, and one that takes a color also accepts every named color, `kw.Currentcolor` and `kw.Transparent`: `css.Width(kw.Inherit)` and `css.Color(kw.Transparent)` compile. A tier 2 property function takes any `Value`, and so any keyword.

### Names shared with properties

`flex`, `grid`, `content` and `all` are each both a property and a keyword in CSS: `flex: 1` and `display: flex`. In `css`, `css.Flex` is the property function. In `kw`, `kw.Flex` is the keyword. The separate package keeps both names free of suffixes such as `FlexKeyword`.

## Media

`Query` is a struct with unexported fields. Build one from a media type or a media feature:

```go
func Screen() Query
func Print() Query
func All() Query
func (q Query) And(o Query) Query
func (q Query) Or(o Query) Query
func Not(q Query) Query
func (q Query) CSS() string
func (q Query) Err() error
```

`Rule.Media` takes a `media.Query`. `CSS` writes the query text, as the writer puts it after `@media`.

```go media
var Wide = style.Class("demo-wide").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint"))).
	Media(media.MaxWidth(css.Rem(48)), css.Background(css.Var("--pink")))
```

```css
.demo-wide {
	padding: 1rem;
	background: var(--mint);
}

@media (max-width: 48rem) {
	.demo-wide {
		background: var(--pink);
	}
}
```

### Combining queries

- `media.Screen()` writes `screen`.
- `media.Screen().And(media.MaxWidth(css.Px(600)))` writes `screen and (max-width: 600px)`.
- `media.MinWidth(css.Px(400)).And(media.MaxWidth(css.Px(800)))` writes `(min-width: 400px) and (max-width: 800px)`.
- `media.Screen().Or(media.Print())` writes `screen, print`. `Or` joins queries into a comma-separated list, which CSS reads as any of them.
- `media.Not(media.Screen())` writes `not screen`.
- `media.Not(media.PrefersColorScheme(kw.Dark))` writes `not (prefers-color-scheme: dark)`.
- `media.Screen().And(media.Not(media.PrefersColorScheme(kw.Dark)))` writes `screen and (not (prefers-color-scheme: dark))`.

`Not` negates the whole operand. A media type may not appear inside parentheses and may not be joined with `or`: `Or` writes a comma list and `Not` is written in front of the query.

### Features

Media features are generated, one set of functions per feature, from the same data as the properties.

A range feature takes its value kind directly, not a marker interface, and comes in four forms. For a feature whose Go name is `X` they are `X`, `MinX`, `MaxX` and `XBetween`. At the current pins the range features are `width`, `height`, `device-width` and `device-height`, which take a `css.Length`; `aspect-ratio` and `device-aspect-ratio`, which take a `css.RatioValue`; `resolution`, which takes a `css.Resolution`; and `color` and `monochrome`, which take a `css.Integer`. For `width`:

```go
func Width(v css.Length) Query
func MinWidth(v css.Length) Query
func MaxWidth(v css.Length) Query
func WidthBetween(lo, hi css.Length) Query
```

- `media.Width(css.Px(600))` writes `(width: 600px)`.
- `media.MinWidth(css.Px(600))` writes `(min-width: 600px)`, and `MaxWidth` writes `(max-width: 600px)`.
- `media.WidthBetween(css.Px(400), css.Px(600))` writes `(400px <= width <= 600px)`, the range syntax. The generator emits it only when the latest release of every considered browser can render it without a vendor prefix; otherwise it emits the `min-` and `max-` pair instead: `(min-width: 400px) and (max-width: 600px)`. The choice is made once, when the package is generated, for every `Between` function. It is the range syntax at the current pins.

A discrete feature takes the keywords it allows, typed in the same way as a property:

```go
media.PrefersColorScheme(kw.Dark)
media.Orientation(kw.Portrait)
media.Pointer(kw.Coarse)
```

These write `(prefers-color-scheme: dark)`, `(orientation: portrait)` and `(pointer: coarse)`. A keyword the feature does not allow is a compile error. A discrete feature also accepts `css.Raw`, but not `css.Var`: a browser does not substitute `var()` inside a media query.

A combination that cannot be written as one media query carries `media.ErrBadQuery`, a `var ErrBadQuery error`. `container` and `supports` export the same sentinel. The Applying page lists the cases under Errors.

## Container

A container query matches the size of the nearest ancestor that is a container. Make an element a container with `css.ContainerType(kw.InlineSize)`, and name it with `css.ContainerName` when a query should target that one by name.

```go
func Named(name string) Query
func (q Query) And(o Query) Query
func (q Query) Or(o Query) Query
func Not(q Query) Query
func (q Query) CSS() string
func (q Query) Err() error
```

`Rule.Container` takes a `container.Query`. The features are generated from the `@container` descriptors:

- Range features, in the four forms `X`, `MinX`, `MaxX` and `XBetween`: `Width`, `Height`, `InlineSize` and `BlockSize`, which take a `css.Length`, and `AspectRatio`, which takes a `css.RatioValue`.
- One discrete feature: `Orientation`, which takes `kw.Portrait` or `kw.Landscape`.

```go container
var Frame = style.Class("demo-frame").
	Set(
		css.ContainerType(kw.InlineSize),
		css.Width(css.Rem(22)),
		css.MaxWidth(css.Pct(100)),
		css.Padding(css.Rem(0.5)),
		css.Border(css.Px(3), kw.Dashed, css.Var("--ink")),
	)

var Narrow = style.Class("demo-frame-narrow")

var NarrowFrame = Frame.And(Narrow).
	Set(css.Width(css.Rem(12)))

var Row = style.Class("demo-row").
	Set(css.Display(kw.Flex), css.Gap(css.Rem(0.5))).
	Container(container.MaxWidth(css.Rem(16)), css.FlexDirection(kw.Column))

var Cell = Row.Child(style.Element("span")).
	Set(css.Padding(css.Rem(0.25), css.Rem(0.5)), css.Background(css.Var("--mint")))
```

```css
.demo-frame {
	container-type: inline-size;
	width: 22rem;
	max-width: 100%;
	padding: 0.5rem;
	border: 3px dashed var(--ink);
}

.demo-frame.demo-frame-narrow {
	width: 12rem;
}

.demo-row {
	display: flex;
	gap: 0.5rem;
}

@container (max-width: 16rem) {
	.demo-row {
		flex-direction: column;
	}
}

.demo-row > span {
	padding: 0.25rem 0.5rem;
	background: var(--mint);
}
```

Both frames hold the same row. The wide one is over `16rem`, so its row stays a row. The narrow one is under it, so its row becomes a column.

- `container.Named("card").And(container.MinWidth(css.Px(400)))` writes `card (min-width: 400px)`. `Named` may appear once, first.
- `container.MinWidth(css.Rem(30)).Or(container.Orientation(kw.Portrait))` writes `(min-width: 30rem) or (orientation: portrait)`.
- `container.Not(container.MaxWidth(css.Rem(20)))` writes `not (max-width: 20rem)`.
- `container.WidthBetween(css.Rem(10), css.Rem(20))` writes `(10rem <= width <= 20rem)`. Container queries always use the range syntax.
- A `Container` block on a rule that is not nested is written flat, as `@container … { .demo-row { … } }`. On a nested rule or a `Self()` child it is written inside the rule's braces, with bare declarations.

## Supports

```go
func Property(p css.Property) Query
func Selector(text string) Query
func FontFormat(f css.FormatValue) Query
func FontTech(t css.Value) Query
func (q Query) And(o Query) Query
func (q Query) Or(o Query) Query
func Not(q Query) Query
func (q Query) CSS() string
func (q Query) Err() error
```

`Rule.Supports` takes a `supports.Query`. Write the fallback in `Set` and the feature in `Supports`:

```go supports
var Mixed = style.Class("demo-mixed").
	Set(css.Padding(css.Rem(1)), css.BackgroundColor(css.Var("--mint"))).
	Supports(
		supports.Property(css.BackgroundColor(css.Raw("color-mix(in srgb, red, blue)"))),
		css.BackgroundColor(css.Raw("color-mix(in srgb, var(--mint), var(--pink))")),
	)
```

```css
.demo-mixed {
	padding: 1rem;
	background-color: var(--mint);
}

@supports (background-color: color-mix(in srgb, red, blue)) {
	.demo-mixed {
		background-color: color-mix(in srgb, var(--mint), var(--pink));
	}
}
```

- `supports.Property(css.Display(kw.Grid))` writes `(display: grid)`. A prefixed property writes only its unprefixed declaration: `supports.Property(css.BackdropFilter(css.Raw("blur(4px)")))` writes `(backdrop-filter: blur(4px))`.
- `supports.Selector(":has(a)")` writes `selector(:has(a))`. The text is raw selector text.
- `supports.FontFormat(kw.Woff2)` writes `font-format(woff2)`. `FontTech` writes `font-tech(…)`.
- `And` and `Or` join two conditions, and a joined operand is wrapped in parentheses: `supports.Property(css.Display(kw.Grid)).Or(supports.Property(css.Display(kw.Flex))).And(supports.Selector(":has(a)"))` writes `((display: grid) or (display: flex)) and selector(:has(a))`.
- `supports.Not(supports.Property(css.Display(kw.Grid)))` writes `not (display: grid)`.

## Keyframes

```go
func Keyframes(name string) *KeyframesRule
func (k *KeyframesRule) At(stop css.Percentage, props ...css.Property) *KeyframesRule
func (k *KeyframesRule) From(props ...css.Property) *KeyframesRule
func (k *KeyframesRule) To(props ...css.Property) *KeyframesRule
func (k *KeyframesRule) Name() string
func (k *KeyframesRule) CSS() string
func (k *KeyframesRule) Err() error
```

```go keyframes
var Pulse = style.Keyframes("demo-pulse").
	From(css.Transform(css.Raw("scale(1)"))).
	To(css.Transform(css.Raw("scale(1.2)")))

var Pulsing = style.Class("demo-pulsing").
	Set(
		css.Width(css.Rem(3)),
		css.Height(css.Rem(3)),
		css.Background(css.Var("--accent")),
		css.AnimationName(css.Ident(Pulse.Name())),
		css.AnimationDuration(css.Ms(800)),
		css.AnimationIterationCount(kw.Infinite),
		css.AnimationDirection(kw.Alternate),
	)
```

```css
@keyframes demo-pulse {
	from {
		transform: scale(1);
	}
	to {
		transform: scale(1.2);
	}
}

.demo-pulsing {
	width: 3rem;
	height: 3rem;
	background: var(--accent);
	animation-name: demo-pulse;
	animation-duration: 800ms;
	animation-iteration-count: infinite;
	animation-direction: alternate;
}
```

- `Name` returns the animation name. Pass it to `css.AnimationName(css.Ident(…))`.
- Frames are written in call order. The same stop twice writes two frames.
- A frame with no properties is not written.
- Stops run from 0% to 100%; a stop outside that range is `ErrBadAtRule`.
- `Keyframes("")` gives a rule whose error is `ErrNoName`.
- A `KeyframesRule` goes into `Generated` like a rule. Bound to an action, it goes into `Dynamic`. The Applying page covers `On`, `Bind`, `Initially` and `CSSFor` on keyframes.

## Font faces

```go
func FontFace(descs ...css.Descriptor) *FontFaceRule
func (f *FontFaceRule) Set(descs ...css.Descriptor) *FontFaceRule
func (f *FontFaceRule) CSS() string
func (f *FontFaceRule) Err() error
```

```go
type Descriptor interface {
	Name() string
	Value() string
	descriptor()
}
```

A `Descriptor` is one declaration inside an at-rule. Only the `css` package makes them. The `@font-face` descriptors are generated: `FaceFontFamily`, `Src`, `FontDisplay`, `FaceFontWeight`, `FaceFontStyle`, `FaceFontFeatureSettings`, `SizeAdjust` and `UnicodeRange`. A descriptor whose name is also a property takes a `Face` prefix: `css.FaceFontFamily` writes `font-family` inside `@font-face`, and `css.FontFamily` stays the property.

```go
var Inter = style.FontFace(
	css.FaceFontFamily(css.String("Inter")),
	css.Src(css.Url("/web/inter.woff2"), css.Format(kw.Woff2)),
	css.FontDisplay(kw.Swap),
	css.FaceFontWeight(css.Num(400)),
)
```

```css
@font-face {
	font-family: "Inter";
	src: url("/web/inter.woff2") format("woff2");
	font-display: swap;
	font-weight: 400;
}
```

- `css.Format(kw.Woff2)` writes `format("woff2")`.
- A font face without `FaceFontFamily` or without `Src` is `ErrBadAtRule`.
- `css.Var` is not accepted on descriptors. `var()` is not substituted inside `@font-face` or `@property`.
- `font-variant`, `ascent-override`, `descent-override`, `line-gap-override`, `font-variation-settings` and `font-width` are not generated at the current pins. Write those in a `Static` file.

This page has no live font face: the site ships no font file.

## Properties

```go
func Property(name string, descs ...css.Descriptor) *PropertyRule
func (p *PropertyRule) Set(descs ...css.Descriptor) *PropertyRule
func (p *PropertyRule) Name() string
func (p *PropertyRule) CSS() string
func (p *PropertyRule) Err() error
```

`Property` writes `@property`: a registered custom property with a type, an inheritance flag and an initial value. Its descriptors are `css.Syntax`, `css.Inherits` and `css.InitialValue`. A registered property animates: the gradient angle below turns on hover.

```go property
var Angle = style.Property("demo-angle",
	css.Syntax(css.SyntaxAngle),
	css.Inherits(kw.False),
	css.InitialValue(css.Deg(0)),
)

var Turning = style.Class("demo-turning").
	Set(
		css.Width(css.Rem(10)),
		css.Height(css.Rem(4)),
		css.Background(css.Raw("linear-gradient(var(--demo-angle), var(--mint), var(--pink))")),
		css.Transition(css.Raw("--demo-angle 600ms")),
	).
	Hover(css.Custom(Angle.Name(), css.Deg(180)))
```

```css
@property --demo-angle {
	syntax: "<angle>";
	inherits: false;
	initial-value: 0deg;
}

.demo-turning {
	width: 10rem;
	height: 4rem;
	background: linear-gradient(var(--demo-angle), var(--mint), var(--pink));
	transition: --demo-angle 600ms;
}

.demo-turning:hover {
	--demo-angle: 180deg;
}
```

- `Property` adds the `--` when the name lacks it. `Name` returns the name with it.
- `css.Syntax` takes a `css.SyntaxValue`: `SyntaxLength`, `SyntaxNumber`, `SyntaxPercentage`, `SyntaxLengthPercentage`, `SyntaxColor`, `SyntaxImage`, `SyntaxUrl`, `SyntaxInteger`, `SyntaxAngle`, `SyntaxTime`, `SyntaxResolution`, `SyntaxTransformFunction`, `SyntaxTransformList`, `SyntaxCustomIdent` and `SyntaxAny`. `css.SyntaxRaw("<length> | auto")` writes any other syntax string.
- `css.Inherits` takes `kw.True` or `kw.False`.
- A property rule without `Syntax` or without `Inherits` is `ErrBadAtRule`. `Property("")` is `ErrNoName`.

## Layers and imports

```go
func Layer(name string, items ...Item) *LayerRule
func LayerOrder(names ...string) *LayerOrderRule
func Import(target css.Value, opts ...ImportOption) *ImportRule
func InLayer(name string) ImportOption
func IfSupports(q supports.Query) ImportOption
func ForMedia(q media.Query) ImportOption
```

A later layer wins over an earlier one, whatever the specificity of the selectors:

```go layers
var LayeredId = style.Id("demo-layered").
	Set(css.Background(css.Var("--pink")))

var Layered = style.Class("demo-layered").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--mint")))

var Under = style.Layer("demo-under", LayeredId)

var Over = style.Layer("demo-over", Layered)
```

```css
@layer demo-under {
	#demo-layered {
		background: var(--pink);
	}
}

@layer demo-over {
	.demo-layered {
		padding: 1rem;
		background: var(--mint);
	}
}
```

- `Layer` holds items and writes them one tab deeper inside `@layer name { … }`. A layer may hold another layer.
- `LayerOrder` writes `@layer a, b, c;`, which fixes the order of the named layers.
- `Import` writes `@import`. The target is `css.Url` or `css.String`; anything else is `ErrBadAtRule`. The options are written in the order layer, supports, media: `style.Import(css.Url("/web/print.css"), style.InLayer("print"), style.IfSupports(supports.Property(css.Display(kw.Grid))), style.ForMedia(media.Print()))` writes `@import url("/web/print.css") layer(print) supports(display: grid) print;`
- An `Import` or a `LayerOrder` inside a `Layer` is `ErrBadAtRule`. A `Layer` in `Dynamic` is `ErrNotDynamic`.
- A selector may repeat across layers. `ErrDuplicateRule` checks siblings only: the top level of a Generated file, and the direct children of each layer.

A Generated file writes every `@layer` order statement first, then every `@import`, then everything else, each group in slice order:

```go
var Base = style.Layer("base", style.Element("body").Set(css.Margin(css.Px(0))))

var Reset = style.Import(css.Url("/web/reset.css"), style.InLayer("reset"))

var Order = style.LayerOrder("reset", "base")

var Site = style.Generated("/web/site.css", Base, Reset, Order)
```

`Site` writes:

```css
@layer reset, base;

@import url("/web/reset.css") layer(reset);

@layer base {
	body {
		margin: 0px;
	}
}
```

## Important

```go
Important() Property
```

`Important` is a method of every `css.Property`. It returns a copy that the writer ends with ` !important`. Calling it twice is the same as calling it once.

```go important
var Quiet = style.Id("demo-quiet").
	Set(css.Background(css.Var("--mint")))

var Loud = style.Class("demo-loud").
	Set(css.Padding(css.Rem(1)), css.Background(css.Var("--pink")).Important())
```

```css
#demo-quiet {
	background: var(--mint);
}

.demo-loud {
	padding: 1rem;
	background: var(--pink) !important;
}
```

The id selector is more specific, and the important declaration still wins. A prefixed property gets ` !important` on every copy: `css.BackdropFilter(css.Raw("blur(4px)")).Important()` writes

```css
.glass {
	-webkit-backdrop-filter: blur(4px) !important;
	backdrop-filter: blur(4px) !important;
}
```

## Generated code

The generator writes:

- the property functions and their marker interfaces, in `css/properties_gen.go` and `css/markers_gen.go`;
- the at-rule descriptor functions, in `css/descriptors_gen.go`;
- the keyword constants and their marker methods, in `kw/*_gen.go`;
- the media features, in `media/features_gen.go`, and the container features, in `container/features_gen.go`;
- the pseudo-class and pseudo-element methods of `Rule`, in `style/pseudo_gen.go`.

Every generated file starts with the line `// Code generated by cssgen. DO NOT EDIT.` and contains no other comment. Browser support is not recorded in the code. Do not edit a generated file by hand: regenerate it.

A property, descriptor, feature or pseudo that the filters below leave out has no generated form, and `css` has no function that writes a property by name (`Custom` writes custom properties only). Such CSS stays in a `.css` file loaded with `Static`. At the current pins that includes `-webkit-font-smoothing`, `-webkit-text-size-adjust`, `resize` and the `::selection` pseudo-element, and this site's own `app.css` keeps the two prefixed properties and `::selection`.

### Which properties are generated

Property grammars come from W3C webref and support data from MDN browser-compat-data (BCD). A property is generated when:

- webref defines it and gives it a grammar;
- BCD marks it as standard-track and not deprecated;
- the latest release of every considered browser can render it, prefixed or unprefixed, according to BCD.

The considered browsers are Chrome, Chrome Android, Edge, Firefox, Firefox Android, Safari and Safari on iOS. A property that any of them cannot render is not generated. Media features, container features, descriptors, pseudo-classes and pseudo-elements go through the same support check. A standard-track property that exists only prefixed is generated under its prefixed name.

Generated names follow fixed rules:

- A name is the CSS name in CamelCase: `background-color` is `BackgroundColor`.
- The leading `-` of a prefixed name is dropped and the vendor kept: `-webkit-text-stroke` is named `WebkitTextStroke`.
- The marker interface of a tier 1 property is its name followed by `Value`: `css.WidthValue`. A tier 1 descriptor's is its name followed by `Descriptor`: `css.FontDisplayDescriptor`.
- A descriptor whose name is also a property takes the `Face` prefix.
- A generated name that would collide with a hand-written identifier makes the generator fail. So does a pseudo-class or pseudo-element whose method would collide with one of `Name`, `Selector`, `Set`, `Media`, `On`, `Initially`, `CSS`, `Err`, `Descendant`, `Child`, `Next`, `Sibling`, `And`, `Nest`, `Bind`, `Container`, `Supports`, `Self` and `CSSFor`.

### Vendor prefixes

A vendor prefix is emitted for a property when some version of a considered browser that was released within the last 30 months supports the property only with that prefix. The 30 months are counted back from the pin date. Only property names are prefixed, never values.

### Regenerating

From the root of the WebGo repository:

```sh
go run ./internal/cssgen
```

The command:

- downloads the pinned npm packages `@webref/css` and `@mdn/browser-compat-data` into a `webgo-cssgen` directory under the user cache directory, one directory per package and version, and reuses them on later runs;
- extracts the JSON it needs, applies the filters above, and writes every generated file;
- prints a report with these lines:
  - `properties: N (tier 1: a, tier 2: b)`, the number of property functions it generated, split into tier 1 and tier 2;
  - `dropped for <browser>: …`, the properties the support filter removed. Each is listed once, under the first of these browsers that cannot render it, named as BCD names them: `chrome`, `chrome_android`, `edge`, `firefox`, `firefox_android`, `safari` and `safari_ios`. A browser that is first for no property has no line;
  - `media features: N (dropped: …)` and `pseudos: N (dropped: …)`, how many it generated and the names it left out, including the pseudos that take a selector and `:before` and `:after`, which `::before` and `::after` replace;
  - `range syntax: true` or `range syntax: false`, the decision the media `Between` functions follow;
  - `descriptors: N (font-face: a, property: b; dropped: …)`, the descriptor functions per at-rule and the descriptors left out;
  - `container features: N (dropped: …)`;
  - `at-rules renderable: …`, the support verdict for `@container`, `@property`, `@layer`, `@supports`, `@import`, `@font-face`, `@keyframes` and nesting. Each is `true` at the current pins.

The pins are Go constants in `internal/cssgen/pins/pins.go`: `WebrefCSSVersion`, `BCDVersion`, `Date` and `FloorMonths`. `Date` is the pin date, and the window for vendor prefixes is `FloorMonths` months back from it. Bumping a pin and regenerating is the whole upgrade procedure.

The output is deterministic. The same pins give the same bytes: emission is sorted, the inputs are pinned, and nothing in the output depends on the date the command runs.
