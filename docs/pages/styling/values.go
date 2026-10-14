package styling

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/styling/demos"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func values() []app.UI {
	return []app.UI{
		doc.Go(`type Value interface{ CSS() string }

type Property interface {
	Name() string
	Value() string
	Prefixes() []string
	Important() Property
	property()
}`),
		doc.P("A `Value` is anything that renders as CSS text. A `Property` is one declaration: a name, a value and the vendor prefixes it needs. `Property` has an unexported method, so only the `css` package makes properties. Every property function returns one, and `Rule.Set` and the pseudo methods take them. `Important` is described under ", doc.A("#important", "Important"), "."),
		doc.H("Kinds"),
		doc.List(
			doc.Item("`Length`: `Px`, `Rem`, `Em`, `Ch`, `Ex`, `Vw`, `Vh`, `Vmin`, `Vmax`, `Dvh`, `Svh`, `Lvh`, `Dvw`, `Svw`, `Lvw`, `Cm`, `Mm`, `In`, `Pt`, `Pc` and `Q`. Each takes a `float64`: `css.Px(8)` is `8px`."),
			doc.Item("`Percentage`: `Pct(50)` is `50%`."),
			doc.Item("`Number`: `Num(1.5)` is `1.5`. `Integer`: `Int(3)` is `3`."),
			doc.Item("`Angle`: `Deg`, `Rad`, `Grad` and `Turn`: `Deg(90)` is `90deg`."),
			doc.Item("`Time`: `S` and `Ms`: `Ms(250)` is `250ms`."),
			doc.Item("`Resolution`: `Dpi`, `Dpcm` and `Dppx`: `Dppx(2)` is `2dppx`."),
			doc.Item("`Fraction`: `Fr(1)` is `1fr`."),
			doc.Item("`RatioValue`: `Ratio(16, 9)` is `16/9`."),
			doc.Item("`ColorLiteral`: `Hex`, `RGB`, `RGBA`, `HSL` and `HSLA`. Named colors are keywords in `kw`."),
			doc.Item("`StringValue`: `String`. `UrlValue`: `Url`. `IdentValue`: `Ident`, written as given."),
			doc.Item("`VarValue`: `Var`. `CalcValue`: `Calc`, `Min`, `Max` and `Clamp`. `RawValue`: `Raw`."),
			doc.Item("`Nth`: `An`, `Odd` and `Even`, for the `Nth…` pseudo methods."),
		),
		doc.P("The color functions:"),
		doc.Go(`func Hex(s string) ColorLiteral
func RGB(r, g, b uint8) ColorLiteral
func RGBA(r, g, b uint8, a float64) ColorLiteral
func HSL(h, s, l float64) ColorLiteral
func HSLA(h, s, l, a float64) ColorLiteral`),
		doc.List(
			doc.Item("`Hex` adds the `#` when it is missing: `Hex(\"fff\")` and `Hex(\"#fff\")` are both `#fff`."),
			doc.Item("`RGB(255, 0, 0)` is `rgb(255, 0, 0)` and `RGBA(255, 0, 0, 0.5)` is `rgba(255, 0, 0, 0.5)`."),
			doc.Item("`HSL(200, 50, 40)` is `hsl(200, 50%, 40%)`: saturation and lightness are plain numbers, written as percentages. `HSLA(200, 50, 40, 0.5)` is `hsla(200, 50%, 40%, 0.5)`."),
		),
		doc.P("Numbers are written in their shortest round-trip form, without an exponent. A zero keeps its unit: `css.Px(0)` is `0px`."),
		doc.H("Property functions"),
		doc.P("Each generated CSS property is one function in `css`, named after the property in CamelCase: `css.Padding`, `css.BackgroundColor`. What a function accepts depends on its tier:"),
		doc.List(
			doc.Item("Tier 1, exact. The property's grammar is made of keywords, the kinds above and simple repetition. The generator emits a marker interface for it, named after the property with a `Value` suffix, and the function takes it. The interface is satisfied by the kinds the grammar accepts, the keywords the grammar names, the CSS-wide keywords `inherit`, `initial`, `unset`, `revert` and `revert-layer`, and, when the grammar accepts a `ColorLiteral`, every named color, `currentcolor` and `transparent`. Any other kind or keyword is a compile error."),
			doc.Item("Tier 2, loose. The grammar is more than keywords, the kinds above and simple repetition: that of `border-radius`, for one, has a `/` between two repeated groups. The function takes `(first css.Value, rest ...css.Value)` and accepts any `Value`. It writes the values separated by commas when the top level of the property's grammar is a comma-separated list, as `font-family`'s is, a choice whose every alternative other than a keyword is such a list, as `text-shadow`'s `none | <shadow>#` is, or a sequence with a literal comma between its parts, as `background`'s `<bg-layer>#? , <final-bg-layer>` is, and separated by spaces otherwise. A value that needs any other arrangement goes in one `css.Raw`."),
		),
		doc.Go(`func Width(v WidthValue) Property
func Padding(first PaddingValue, rest ...PaddingValue) Property`),
		doc.P("`Width` takes one value. `Padding` is a repeated property, `{1,4}` in its grammar, so it takes a first value and more after it, and writes them separated by a space. A property whose grammar repeats with `#` writes its values separated by commas. The compiler does not enforce the repeat count. `css.Width(kw.Auto)` and `css.Width(css.Px(10))` compile. `css.Width(kw.Flex)` does not: `flex` is not among the keywords of the `width` grammar. A property's signature shows its tier: `css.Width`, `css.Padding` and `css.Color` are tier 1, and `css.BorderRadius`, `css.TextDecoration` and `css.Flex` are tier 2."),
		doc.P("Tier 1 functions also accept `css.Var` and `css.Raw`, and accept `css.Calc` where the property takes a length, percentage, number, integer, angle, time or fraction."),
		doc.P("A property that needs vendor prefixes carries them in the `Property` it returns, and the writer puts each prefixed declaration before the unprefixed one."),
		doc.H("Raw"),
		doc.Go(`func Raw(s string) RawValue`),
		doc.P("`Raw` writes its text as given. It is accepted by every property function and is the escape hatch for a value the typed grammar does not cover: `css.Width(css.Raw(\"fit-content(20em)\"))` writes `width: fit-content(20em);`. `Raw` is not checked."),
		doc.H("Var"),
		doc.Go(`func Var(name string, fallback ...Value) VarValue`),
		doc.P("`Var` writes a `var()` reference. It adds the `--` when it is missing, and joins several fallback values with `, `:"),
		doc.List(
			doc.Item("`Var(\"--ink\")` is `var(--ink)`."),
			doc.Item("`Var(\"ink\")` is `var(--ink)`."),
			doc.Item("`Var(\"--ink\", Hex(\"#000\"))` is `var(--ink, #000)`."),
			doc.Item("`Var(\"--a\", Px(1), Pct(2))` is `var(--a, 1px, 2%)`."),
		),
		doc.H("Calc"),
		doc.Go(`func Calc(e Expr) CalcValue
func Min(first Expr, rest ...Expr) CalcValue
func Max(first Expr, rest ...Expr) CalcValue
func Clamp(lo, preferred, hi Expr) CalcValue
func Add(a, b Expr) Expr
func Sub(a, b Expr) Expr
func Mul(a Expr, n Number) Expr
func Div(a Expr, n Number) Expr`),
		doc.P("`Expr` has an unexported method. It is implemented by `Length`, `Percentage`, `Number`, `Integer`, `Angle`, `Time`, `Fraction`, `VarValue`, `CalcValue` and the results of `Add`, `Sub`, `Mul` and `Div`."),
		doc.List(
			doc.Item("`Calc(Sub(Pct(100), Px(20)))` is `calc(100% - 20px)`."),
			doc.Item("`Calc(Mul(Add(Pct(100), Px(2)), Num(2)))` is `calc((100% + 2px) * 2)`. The result of `Add`, `Sub`, `Mul` or `Div` used as the left operand of `Mul` or `Div`, or as the right operand of `Sub`, is wrapped in parentheses."),
			doc.Item("`Calc(Div(Var(\"--w\"), Num(3)))` is `calc(var(--w) / 3)`."),
			doc.Item("`Min(Px(10), Pct(5))` is `min(10px, 5%)`, and `Max` is the same with `max`."),
			doc.Item("`Clamp(Rem(1), Vw(2), Rem(3))` is `clamp(1rem, 2vw, 3rem)`."),
		),
		doc.H("Custom"),
		doc.Go(`func Custom(name string, v Value) Property`),
		doc.P("`Custom` defines a custom property. It adds the `--` when it is missing, and it carries no vendor prefixes. Read it back with `Var`:"),
		doc.Example(demos.ValuesExample),
		doc.H("Strings"),
		doc.Go(`func String(s string) StringValue
func Url(s string) UrlValue`),
		doc.P("`String` writes a double-quoted CSS string, and `Url` writes `url(\"…\")` with the same quoting. Inside the quotes:"),
		doc.List(
			doc.Item("a backslash is written `\\\\`;"),
			doc.Item("a double quote is written `\\\"`;"),
			doc.Item("a newline is written `\\a `, a carriage return `\\d ` and a form feed `\\c `, each followed by the space that ends the escape."),
		),
		doc.P("So `css.String(\"→\")` is `\"→\"`, a string holding a double quote between two letters is written `\"a\\\"b\"`, a string holding a newline between two letters is written `\"a\\a b\"`, and `css.Url(\"x.png\")` is `url(\"x.png\")`."),
	}
}

func keywords() []app.UI {
	return []app.UI{
		doc.P("The `kw` package has one constant for every keyword that appears in the grammar of a generated property, descriptor or query feature, plus every named color, `currentcolor` and `transparent`, the CSS-wide keywords, and `ltr` and `rtl` for `Rule.Dir`. The name is the keyword in CamelCase: `inline-flex` is `kw.InlineFlex`, `x-small` is `kw.XSmall` and `rebeccapurple` is `kw.Rebeccapurple`. The feature and descriptor keywords are here too: `kw.Dark`, `kw.Portrait`, `kw.Woff2` and `kw.Swap`."),
		doc.P("Each keyword is a constant of its own type: `kw.InlineFlex` has the unexported type `kInlineFlex`, whose underlying type is `string`. A tier 1 property function accepts the keywords its grammar names, and each keyword has its own type, so `kw.Flex` is not accepted where only `kw.Auto` is. `css.Display(kw.InlineFlex)` and `css.Width(kw.Auto)` compile. `css.Width(kw.Flex)` does not. Every tier 1 function also accepts the CSS-wide keywords `kw.Inherit`, `kw.Initial`, `kw.Unset`, `kw.Revert` and `kw.RevertLayer`, and one that takes a color also accepts every named color, `kw.Currentcolor` and `kw.Transparent`: `css.Width(kw.Inherit)` and `css.Color(kw.Transparent)` compile. A tier 2 property function takes any `Value`, and so any keyword."),
		doc.H("Names shared with properties"),
		doc.P("`flex`, `grid`, `content` and `all` are each both a property and a keyword in CSS: `flex: 1` and `display: flex`. In `css`, `css.Flex` is the property function. In `kw`, `kw.Flex` is the keyword. The separate package keeps both names free of suffixes such as `FlexKeyword`."),
	}
}
