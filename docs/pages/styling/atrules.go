package styling

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/pages/styling/demos"
)

func keyframes() []app.UI {
	return []app.UI{
		doc.Go(`func Keyframes(name string) *KeyframesRule
func (k *KeyframesRule) At(stop css.Percentage, props ...css.Property) *KeyframesRule
func (k *KeyframesRule) From(props ...css.Property) *KeyframesRule
func (k *KeyframesRule) To(props ...css.Property) *KeyframesRule
func (k *KeyframesRule) Name() string
func (k *KeyframesRule) CSS() string
func (k *KeyframesRule) Err() error`),
		doc.Example(demos.KeyframesExample),
		doc.List(
			doc.Item("`Name` returns the animation name. Pass it to `css.AnimationName(css.Ident(…))`."),
			doc.Item("Frames are written in call order. The same stop twice writes two frames."),
			doc.Item("A frame with no properties is not written."),
			doc.Item("Stops run from 0% to 100%; a stop outside that range is `ErrBadAtRule`."),
			doc.Item("`Keyframes(\"\")` gives a rule whose error is `ErrNoName`."),
			doc.Item("A `KeyframesRule` goes into `Generated` like a rule. Bound to an action, it goes into `Dynamic`. The ", doc.A("/applying", "Applying"), " page covers `On`, `Bind`, `Initially` and `CSSFor` on keyframes."),
		),
	}
}

func fontFaces() []app.UI {
	return []app.UI{
		doc.Go(`func FontFace(descs ...css.Descriptor) *FontFaceRule
func (f *FontFaceRule) Set(descs ...css.Descriptor) *FontFaceRule
func (f *FontFaceRule) CSS() string
func (f *FontFaceRule) Err() error`),
		doc.Go(`type Descriptor interface {
	Name() string
	Value() string
	descriptor()
}`),
		doc.P("A `Descriptor` is one declaration inside an at-rule. Only the `css` package makes them. The `@font-face` descriptors are generated: `FaceFontFamily`, `Src`, `FontDisplay`, `FaceFontWeight`, `FaceFontStyle`, `FaceFontFeatureSettings`, `SizeAdjust` and `UnicodeRange`. A descriptor whose name is also a property takes a `Face` prefix: `css.FaceFontFamily` writes `font-family` inside `@font-face`, and `css.FontFamily` stays the property."),
		doc.Go(`var Inter = style.FontFace(
	css.FaceFontFamily(css.String("Inter")),
	css.Src(css.Url("/web/inter.woff2"), css.Format(kw.Woff2)),
	css.FontDisplay(kw.Swap),
	css.FaceFontWeight(css.Num(400)),
)`),
		doc.CSS(`@font-face {
	font-family: "Inter";
	src: url("/web/inter.woff2") format("woff2");
	font-display: swap;
	font-weight: 400;
}`),
		doc.List(
			doc.Item("`css.Format(kw.Woff2)` writes `format(\"woff2\")`."),
			doc.Item("A font face without `FaceFontFamily` or without `Src` is `ErrBadAtRule`."),
			doc.Item("`css.Var` is not accepted on descriptors. `var()` is not substituted inside `@font-face` or `@property`."),
			doc.Item("`font-variant`, `ascent-override`, `descent-override`, `line-gap-override`, `font-variation-settings` and `font-width` are not generated at the current pins. Write those in a `Static` file."),
		),
		doc.P("This page has no live font face: the site ships no font file."),
	}
}

func properties() []app.UI {
	return []app.UI{
		doc.Go(`func Property(name string, descs ...css.Descriptor) *PropertyRule
func (p *PropertyRule) Set(descs ...css.Descriptor) *PropertyRule
func (p *PropertyRule) Name() string
func (p *PropertyRule) CSS() string
func (p *PropertyRule) Err() error`),
		doc.P("`Property` writes `@property`: a registered custom property with a type, an inheritance flag and an initial value. Its descriptors are `css.Syntax`, `css.Inherits` and `css.InitialValue`. A registered property animates: the gradient angle below turns on hover."),
		doc.Example(demos.PropertyExample),
		doc.List(
			doc.Item("`Property` adds the `--` when the name lacks it. `Name` returns the name with it."),
			doc.Item("`css.Syntax` takes a `css.SyntaxValue`: `SyntaxLength`, `SyntaxNumber`, `SyntaxPercentage`, `SyntaxLengthPercentage`, `SyntaxColor`, `SyntaxImage`, `SyntaxUrl`, `SyntaxInteger`, `SyntaxAngle`, `SyntaxTime`, `SyntaxResolution`, `SyntaxTransformFunction`, `SyntaxTransformList`, `SyntaxCustomIdent` and `SyntaxAny`. `css.SyntaxRaw(\"<length> | auto\")` writes any other syntax string."),
			doc.Item("`css.Inherits` takes `kw.True` or `kw.False`."),
			doc.Item("A property rule without `Syntax` or without `Inherits` is `ErrBadAtRule`. `Property(\"\")` is `ErrNoName`."),
		),
	}
}

func layers() []app.UI {
	return []app.UI{
		doc.Go(`func Layer(name string, items ...Item) *LayerRule
func LayerOrder(names ...string) *LayerOrderRule
func Import(target css.Value, opts ...ImportOption) *ImportRule
func InLayer(name string) ImportOption
func IfSupports(q supports.Query) ImportOption
func ForMedia(q media.Query) ImportOption`),
		doc.P("A later layer wins over an earlier one, whatever the specificity of the selectors:"),
		doc.Example(demos.LayersExample),
		doc.List(
			doc.Item("`Layer` holds items and writes them one tab deeper inside `@layer name { … }`. A layer may hold another layer."),
			doc.Item("`LayerOrder` writes `@layer a, b, c;`, which fixes the order of the named layers."),
			doc.Item("`Import` writes `@import`. The target is `css.Url` or `css.String`; anything else is `ErrBadAtRule`. The options are written in the order layer, supports, media: `style.Import(css.Url(\"/web/print.css\"), style.InLayer(\"print\"), style.IfSupports(supports.Property(css.Display(kw.Grid))), style.ForMedia(media.Print()))` writes `@import url(\"/web/print.css\") layer(print) supports(display: grid) print;`"),
			doc.Item("An `Import` or a `LayerOrder` inside a `Layer` is `ErrBadAtRule`. A `Layer` in `Dynamic` is `ErrNotDynamic`."),
			doc.Item("A selector may repeat across layers. `ErrDuplicateRule` checks siblings only: the top level of a Generated file, and the direct children of each layer."),
		),
		doc.P("A Generated file writes every `@layer` order statement first, then every `@import`, then everything else, each group in slice order:"),
		doc.Go(`var Base = style.Layer("base", style.Element("body").Set(css.Margin(css.Px(0))))

var Reset = style.Import(css.Url("/web/reset.css"), style.InLayer("reset"))

var Order = style.LayerOrder("reset", "base")

var Site = style.Generated("/web/site.css", Base, Reset, Order)`),
		doc.P("`Site` writes:"),
		doc.CSS(`@layer reset, base;

@import url("/web/reset.css") layer(reset);

@layer base {
	body {
		margin: 0px;
	}
}`),
	}
}

func important() []app.UI {
	return []app.UI{
		doc.Go(`Important() Property`),
		doc.P("`Important` is a method of every `css.Property`. It returns a copy that the writer ends with ` !important`. Calling it twice is the same as calling it once."),
		doc.Example(demos.ImportantExample),
		doc.P("The id selector is more specific, and the important declaration still wins. A prefixed property gets ` !important` on every copy: `css.BackdropFilter(css.Raw(\"blur(4px)\")).Important()` writes"),
		doc.CSS(`.glass {
	-webkit-backdrop-filter: blur(4px) !important;
	backdrop-filter: blur(4px) !important;
}`),
	}
}
