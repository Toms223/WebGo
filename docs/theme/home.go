package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Landing = style.Class("landing").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.MinHeight(css.Vh(100)),
	css.OverflowX(kw.Clip),
)

var Topbar = style.Class("topbar").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.AlignItems(kw.Center),
		css.JustifyContent(kw.SpaceBetween),
		css.Gap(css.Rem(1)),
		css.Padding(css.Rem(1), css.Rem(2)),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
		css.Background(css.Var("panel")),
	).
	Media(narrow, css.Padding(css.Rem(0.75), css.Rem(1)))

var TopbarNav = style.Class("topbar-nav").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.6)),
)

var TopbarToggle = Topbar.Descendant(ShellToggle).Set(css.MarginTop(css.Px(0)), css.AlignSelf(kw.Center))

var Wrap = style.Class("wrap").
	Set(
		css.Width(css.Pct(100)),
		css.MaxWidth(css.Px(1180)),
		css.Margin(css.Px(0), kw.Auto),
		css.Padding(css.Px(0), css.Rem(1.5)),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
	).
	Media(narrow, css.Padding(css.Px(0), css.Rem(1)))

func dots(ink string) css.RawValue {
	return css.Raw("radial-gradient(circle, " + ink + " 34%, transparent 37%)")
}

var bleed = css.Margin(css.Px(0), css.Raw("calc(50% - 50vw)"))

var rule = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Absolute),
	css.Bottom(css.Px(0)),
	css.Left(css.Raw("calc(50% - 50vw)")),
	css.Width(css.Vw(100)),
	css.Height(css.Var("border")),
	css.Background(css.Var("line")),
}

var Slug = style.Class("slug").
	Set(label...).
	Set(
		bleed,
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.JustifyContent(kw.SpaceBetween),
		css.Gap(css.Rem(0.4), css.Rem(1.4)),
		css.Padding(css.Rem(0.5), css.Raw("max(1.5rem, calc(50vw - 575px))")),
		css.Background(css.Var("on-fill")),
		css.Color(css.Var("on-blue")),
		css.FontSize(css.Rem(0.85)),
		css.LetterSpacing(css.Em(0.06)),
	).
	Media(narrow, css.Padding(css.Rem(0.5), css.Rem(1)))

var SlugIssue = Slug.Descendant(style.Element("b")).Set(css.Color(css.Var("ink-yellow")))

var SlugWide = style.Class("slug-wide").Media(narrow, css.Display(kw.None))

var Hero = style.Class("hero").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 0.9fr) minmax(0, 1.1fr)")),
		css.Gap(css.Px(0), css.Rem(3.5)),
		css.AlignItems(kw.Start),
		css.Padding(css.Rem(3.5), css.Px(0), css.Rem(6)),
	).
	After(rule...).
	Media(narrow, css.GridTemplateColumns(css.Raw("minmax(0, 1fr)")), css.Padding(css.Rem(2.25), css.Px(0), css.Rem(5)))

var HeroInk = style.Class("hero-ink").
	Set(
		css.Position(kw.Absolute),
		css.ZIndex(css.Int(0)),
		css.PointerEvents(kw.None),
		css.Width(css.Pct(66)),
		css.Height(css.Px(300)),
		css.Left(css.Pct(-6)),
		css.Top(css.Px(96)),
		css.Background(css.Var("code-mark")),
		css.MixBlendMode(css.Var("blend")),
		css.Transform(css.Raw("rotate(-2deg)")),
	).
	Media(narrow, css.Width(css.Pct(105)), css.Height(css.Px(210)), css.Top(css.Px(70)))

var HeroDots = style.Class("hero-dots").
	Set(
		css.Position(kw.Absolute),
		css.ZIndex(css.Int(0)),
		css.PointerEvents(kw.None),
		css.Width(css.Pct(48)),
		css.Height(css.Px(400)),
		css.Right(css.Pct(-8)),
		css.Top(css.Px(20)),
		css.BackgroundImage(dots("var(--ink-pink)")),
		css.BackgroundSize(css.Raw("9px 9px")),
		css.MaskImage(css.Raw("linear-gradient(115deg, #000 15%, transparent 85%)")),
		css.MixBlendMode(css.Var("blend")),
	).
	Media(narrow, css.Width(css.Pct(80)), css.Top(kw.Auto), css.Bottom(css.Px(0)))

var HeroHead = style.Class("hero-head").Set(
	css.GridColumn(css.Raw("1 / -1")),
	css.Position(kw.Relative),
	css.ZIndex(css.Int(2)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.5)),
)

var HeroCopy = style.Class("hero-copy").Set(
	css.Position(kw.Relative),
	css.ZIndex(css.Int(1)),
	css.MinWidth(css.Px(0)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1.75)),
)

var HeroInstall = HeroCopy.Child(Code).Set(
	css.AlignSelf(kw.Stretch),
	css.MaxWidth(css.Em(36)),
	css.MinWidth(css.Px(0)),
	css.Transform(css.Raw("rotate(0.8deg)")),
)

var HeroSample = style.Class("hero-sample").
	Set(
		css.Position(kw.Relative),
		css.ZIndex(css.Int(1)),
		css.MinWidth(css.Px(0)),
		css.MarginTop(css.Rem(0.5)),
	).
	Media(narrow, css.MarginTop(css.Rem(3)))

var HeroTitle = style.Class("hero-title").Set(
	css.Margin(css.Rem(0.4), css.Px(0), css.Rem(1)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2.75), css.Vw(9), css.Rem(8.25))),
	css.FontWeight(css.Num(400)),
	css.LetterSpacing(css.Em(0.004)),
	css.LineHeight(css.Num(0.9)),
	css.TextTransform(kw.Uppercase),
)

var Misregister = style.Keyframes("misregister").
	From(css.Transform(css.Raw("translate(0.14em, 0.1em)"))).
	To(css.Transform(css.Raw("translate(0.035em, 0.022em)")))

var Mis = style.Class("mis").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.InlineBlock),
		css.Color(css.Var("mis-ink")),
	).
	After(
		css.Content(css.Raw("attr(data-ink)")),
		css.Position(kw.Absolute),
		css.Inset(css.Px(0)),
		css.Color(css.Var("ink-pink")),
		css.Transform(css.Raw("translate(0.035em, 0.022em)")),
		css.MixBlendMode(css.Var("blend")),
		css.PointerEvents(kw.None),
		css.AnimationName(css.Ident(Misregister.Name())),
		css.AnimationDuration(css.Ms(900)),
		css.AnimationTimingFunction(css.Raw("cubic-bezier(0.16, 1, 0.3, 1)")),
		css.AnimationDelay(css.Ms(150)),
		css.AnimationFillMode(kw.Both),
	)

var Lead = style.Class("lead").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(54)),
	css.Color(css.Var("ink")),
	css.FontSize(css.Clamp(css.Rem(1.05), css.Vw(1.6), css.Rem(1.2))),
	css.LineHeight(css.Num(1.6)),
)

var HeroLead = HeroCopy.Child(Lead).
	Set(
		css.Position(kw.Relative),
		css.MaxWidth(css.Em(36)),
		css.Padding(css.Rem(1.1), css.Rem(1.25)),
		css.Border(css.Px(2), kw.Solid, css.Var("line")),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	After(
		css.Content(css.String("p.1 / lede")),
		css.Position(kw.Absolute),
		css.Right(css.Px(-2)),
		css.Bottom(css.Px(-24)),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.85)),
		css.Color(css.Var("ink-2")),
	)

var Actions = style.Class("actions").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(1)),
)

var HandNote = style.Class("hand-note").Set(
	css.FontFamily(css.Var("hand")),
	css.FontSize(css.Rem(1.15)),
	css.LineHeight(css.Num(1.1)),
	css.Color(css.Var("mis-ink")),
	css.Transform(css.Raw("rotate(-4deg)")),
)

var Chip = style.Class("chip").
	Set(label...).
	Set(
		css.Display(kw.InlineFlex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.Padding(css.Rem(0.3), css.Rem(0.65)),
		css.Border(css.Px(2), kw.Solid, css.Var("line")),
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.FontSize(css.Rem(0.85)),
		css.LetterSpacing(css.Em(0.1)),
		css.LineHeight(css.Num(1.4)),
	).
	Before(
		css.Content(css.String("")),
		css.Width(css.Px(22)),
		css.Height(css.Px(10)),
		css.Border(css.Px(2), kw.Solid, css.Var("line")),
		css.Background(css.Var("ink-pink")),
	)

var Code = style.Class("code").
	Set(
		css.Position(kw.Relative),
		css.Border(css.Px(2), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Var("r-l")),
		css.Background(css.Var("code")),
		css.Color(css.Var("code-on")),
		css.BoxShadow(css.Var("shadow")),
	).
	Before(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Inset(css.Px(0)),
		css.BackgroundImage(css.Var("noise")),
		css.Opacity(css.Num(0.3)),
		css.MixBlendMode(css.Var("blend")),
		css.PointerEvents(kw.None),
	)

var CodeBar = style.Class("code-bar").Set(
	css.Display(kw.Flex),
	css.JustifyContent(kw.SpaceBetween),
	css.AlignItems(kw.Baseline),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.6), css.Rem(1), css.Rem(0.5)),
	css.BorderBottom(css.Px(2), kw.Dashed, css.Var("line")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.875)),
	css.FontWeight(css.Num(700)),
)

var CodeNote = style.Class("code-note").Set(css.Color(css.Var("code-dim")), css.FontWeight(css.Num(400)))

var CodeFoot = style.Class("code-foot").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(0.5)),
	css.Padding(css.Rem(0.4), css.Rem(1), css.Rem(0.55)),
	css.BorderTop(css.Px(2), kw.Dashed, css.Var("line")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.85)),
	css.Color(css.Var("code-dim")),
)

var CodeBody = Code.Descendant(style.Element("pre")).Set(
	css.Position(kw.Relative),
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(1), css.Rem(1.15), css.Rem(1.1)),
	css.OverflowX(kw.Auto),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.9)),
	css.LineHeight(css.Num(1.65)),
	css.TabSize(css.Num(4)),
)

var CodeTilt = style.Class("code-tilt").Set(css.Transform(css.Raw("rotate(1.4deg)")))

var CodeLean = style.Class("code-lean").Set(css.Transform(css.Raw("rotate(-1deg)")))

var CodeQuiet = style.Class("code-quiet").Set(css.BoxShadow(kw.None))

var CodeQuietBody = CodeQuiet.Descendant(style.Element("pre")).Set(
	css.WhiteSpace(kw.PreWrap),
	css.OverflowWrap(kw.Anywhere),
)

var tapeStrip = []css.Property{
	css.Position(kw.Absolute),
	css.ZIndex(css.Int(2)),
	css.Width(css.Px(118)),
	css.Height(css.Px(30)),
	css.Background(css.Var("tape")),
	css.ClipPath(css.Raw("polygon(0 8%, 6% 0, 12% 10%, 20% 2%, 30% 9%, 42% 0, 55% 8%, 68% 1%, 80% 9%, 92% 0, 100% 7%, 100% 92%, 94% 100%, 84% 91%, 72% 100%, 60% 92%, 46% 100%, 34% 91%, 22% 100%, 10% 92%, 0 100%)")),
	css.PointerEvents(kw.None),
}

var Tape = style.Class("tape").Set(tapeStrip...).Set(css.Top(css.Px(-15)), css.Left(css.Px(22)), css.Transform(css.Raw("rotate(-6deg)")))

var TapeEnd = style.Class("tape-end").Set(tapeStrip...).Set(css.Bottom(css.Px(-14)), css.Right(css.Px(40)), css.Transform(css.Raw("rotate(-4deg)")))

var TapePink = style.Class("tape-pink").Set(tapeStrip...).Set(
	css.Top(css.Px(-13)),
	css.Right(css.Pct(42)),
	css.Background(css.Var("code-mark")),
	css.Transform(css.Raw("rotate(5deg)")),
)

var Keyword = style.Class("k").Set(css.Color(css.Var("code-key")), css.FontWeight(css.Num(700)))

var Str = style.Class("s").Set(css.Color(css.Var("code-str")))

var Fn = style.Class("f").Set(css.Background(css.Raw("linear-gradient(transparent 58%, var(--code-mark) 58%)")))

var Typ = style.Class("t").Set(css.Color(css.Var("code-type")))

var Dim = style.Class("c").Set(css.Color(css.Var("code-dim")))

var Mark = style.Class("mark").Set(
	css.Background(css.Raw("color-mix(in srgb, var(--ink-pink) 22%, transparent)")),
	css.BoxShadow(css.Raw("-14px 0 0 color-mix(in srgb, var(--ink-pink) 22%, transparent), 14px 0 0 color-mix(in srgb, var(--ink-pink) 22%, transparent)")),
)

var Scribble = style.Class("scribble").
	Set(
		css.Position(kw.Absolute),
		css.ZIndex(css.Int(3)),
		css.PointerEvents(kw.None),
		css.FontFamily(css.Var("hand")),
		css.FontSize(css.Rem(1.3)),
		css.LineHeight(css.Num(1.15)),
		css.Color(css.Var("code-str")),
	).
	Media(narrow, css.Display(kw.None))

var Draw = style.Keyframes("draw").
	From(css.StrokeDashoffset(css.Num(600))).
	To(css.StrokeDashoffset(css.Num(0)))

var ScribblePath = Scribble.Descendant(style.Element("path")).Set(
	css.Fill(kw.None),
	css.Stroke(css.Raw("currentColor")),
	css.StrokeWidth(css.Num(3)),
	css.StrokeLinecap(kw.Round),
	css.StrokeLinejoin(kw.Round),
	css.StrokeDasharray(css.Num(600)),
	css.AnimationName(css.Ident(Draw.Name())),
	css.AnimationDuration(css.Ms(1100)),
	css.AnimationDelay(css.Ms(600)),
	css.AnimationFillMode(kw.Backwards),
)

var ScribbleArrow = Scribble.Descendant(style.Element("svg")).Set(css.Position(kw.Absolute), css.Overflow(kw.Visible))

var ScribbleWhole = style.Class("scribble-whole").Set(css.Right(css.Px(30)), css.Bottom(css.Px(-58)), css.Width(css.Px(230)), css.Transform(css.Raw("rotate(-4deg)")))

var ScribbleWholeArrow = ScribbleWhole.Descendant(style.Element("svg")).Set(
	css.Left(css.Px(-110)),
	css.Top(css.Px(-70)),
	css.Width(css.Px(120)),
	css.Height(css.Px(90)),
	css.Transform(css.Raw("scaleX(-1) rotate(10deg)")),
)

var ScribbleCircle = style.Class("scribble-circle").Set(css.Left(css.Px(4)), css.Top(css.Px(124)), css.Width(css.Px(400)), css.Height(css.Px(110)), css.Color(css.Var("ink-pink")))

var ScribbleRename = style.Class("scribble-rename").Set(css.Right(css.Px(-8)), css.Top(css.Px(96)), css.Width(css.Px(190)), css.TextAlign(kw.Right), css.Transform(css.Raw("rotate(5deg)")))

var ScribbleRenameArrow = ScribbleRename.Descendant(style.Element("svg")).Set(css.Left(css.Px(-10)), css.Top(css.Px(16)), css.Width(css.Px(90)), css.Height(css.Px(40)))

var Spin = style.Keyframes("spin").To(css.Transform(css.Raw("rotate(360deg)")))

var Sticker = style.Class("sticker").
	Set(
		css.Position(kw.Absolute),
		css.ZIndex(css.Int(4)),
		css.Right(css.Px(-18)),
		css.Top(css.Px(-34)),
		css.Width(css.Px(150)),
		css.Filter(css.Raw("drop-shadow(4px 4px 0 var(--shadow-ink))")),
		css.Transform(css.Raw("rotate(12deg)")),
		css.Transition(css.Raw("transform 350ms cubic-bezier(0.3, 1.6, 0.5, 1)")),
	).
	Hover(css.Transform(css.Raw("rotate(30deg) scale(1.06)"))).
	Media(narrow, css.Width(css.Px(104)), css.Right(css.Px(-6)), css.Top(css.Px(-52)))

var StickerArt = Sticker.Child(style.Element("svg")).Set(css.Display(kw.Block), css.Width(css.Pct(100)), css.Height(kw.Auto))

var StickerRing = Sticker.Descendant(style.Element("g")).Set(
	css.TransformOrigin(css.Px(80), css.Px(80)),
	css.Animation(css.Raw("spin 40s linear infinite")),
)

var Reg = style.Class("reg").
	Set(
		css.Position(kw.Absolute),
		css.Top(css.Px(18)),
		css.Left(css.Px(-6)),
		css.Width(css.Px(26)),
		css.Height(css.Px(26)),
		css.Border(css.Px(1.5), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Pct(50)),
		css.PointerEvents(kw.None),
	).
	Before(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Left(css.Pct(50)),
		css.Top(css.Px(-6)),
		css.Width(css.Px(1.5)),
		css.Height(css.Raw("calc(100% + 12px)")),
		css.Background(css.Var("line")),
		css.Transform(css.Raw("translateX(-50%)")),
	).
	After(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Left(css.Pct(50)),
		css.Top(css.Px(-6)),
		css.Width(css.Px(1.5)),
		css.Height(css.Raw("calc(100% + 12px)")),
		css.Background(css.Var("line")),
		css.Transform(css.Raw("translateX(-50%) rotate(90deg)")),
	).
	Media(narrow, css.Display(kw.None))

var RegEnd = style.Class("reg-end").Set(css.Left(kw.Auto), css.Right(css.Px(-6)))

var Folio = style.Class("folio").
	Set(
		css.Position(kw.Absolute),
		css.ZIndex(css.Int(0)),
		css.Right(css.Px(0)),
		css.Top(css.Rem(2.5)),
		css.PointerEvents(kw.None),
		css.UserSelect(kw.None),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Clamp(css.Rem(4), css.Vw(9), css.Rem(9.4))),
		css.FontWeight(css.Num(400)),
		css.LineHeight(css.Num(0.8)),
		css.WebkitTextStroke(css.Px(2), css.Var("line")),
		css.MixBlendMode(css.Var("blend")),
	).
	Media(narrow, css.Top(css.Rem(1.25)))

var FolioLow = style.Class("folio-low").
	Set(css.Top(kw.Auto), css.Bottom(css.Rem(-0.8))).
	Media(narrow, css.Top(kw.Auto), css.Bottom(css.Rem(1)))

var InkPink = style.Class("ink-pink").Set(css.Color(css.Var("ink-pink")))

var InkBlue = style.Class("ink-blue").Set(css.Color(css.Var("mis-ink")))

var InkYellow = style.Class("ink-yellow").Set(css.Color(css.Var("ink-yellow")), css.WebkitTextStroke(css.Px(2), css.Var("line")))

var Eyebrow = style.Class("eyebrow").
	Set(label...).
	Set(
		css.Display(kw.InlineFlex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.Color(css.Var("ink")),
		css.FontSize(css.Rem(0.85)),
		css.LetterSpacing(css.Em(0.1)),
	).
	Before(
		css.Content(css.String("")),
		css.Width(css.Px(28)),
		css.Height(css.Px(12)),
		css.Border(css.Px(2), kw.Solid, css.Var("line")),
		css.Background(css.Var("ink-pink")),
	)

var Section = style.Class("section").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.FlexStart),
		css.Gap(css.Rem(1.25)),
		css.Padding(css.Rem(6), css.Px(0)),
	).
	After(rule...).
	Media(narrow, css.Padding(css.Rem(4.5), css.Px(0)))

var SectionTitle = style.Class("section-title").Set(
	css.Position(kw.Relative),
	css.ZIndex(css.Int(1)),
	css.Margin(css.Px(0), css.Px(0), css.Rem(0.75)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2.1), css.Vw(5), css.Rem(4.75))),
	css.FontWeight(css.Num(400)),
	css.LetterSpacing(css.Em(0.005)),
	css.LineHeight(css.Num(0.98)),
	css.TextTransform(kw.Uppercase),
)

var Slab = style.Class("slab").Set(
	css.Display(kw.Inline),
	css.Padding(css.Px(0), css.Px(10)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow-s")),
	css.BoxDecorationBreak(kw.Clone),
	css.LineHeight(css.Num(1.18)),
)

var Tiles = style.Class("tiles").Set(
	css.AlignSelf(kw.Stretch),
	css.Display(kw.Grid),
	css.GridTemplateColumns(css.Raw("repeat(auto-fit, minmax(230px, 1fr))")),
	css.Gap(css.Rem(1.6)),
	css.MarginTop(css.Rem(1.5)),
)

var Card = style.Class("card").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(0.5)),
		css.PaddingBottom(css.Rem(1.25)),
		line,
		css.BorderRadius(css.Var("r-l")),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow")),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		css.Transform(css.Raw("rotate(var(--tilt, 0deg))")),
		css.Transition(css.Raw("transform 250ms cubic-bezier(0.3, 1.5, 0.5, 1), box-shadow 200ms ease")),
	).
	Hover(css.Transform(css.Raw("translate(-3px, -5px)")), css.BoxShadow(css.Var("shadow-l"))).
	FocusVisible(ring...).
	NthChild(css.An(4, 1), css.Custom("tilt", css.Deg(-1.2))).
	NthChild(css.An(4, 2), css.Custom("tilt", css.Deg(0.9))).
	NthChild(css.An(4, 3), css.Custom("tilt", css.Deg(-0.5))).
	NthChild(css.An(4, 4), css.Custom("tilt", css.Deg(1.4)))

var Plate = style.Class("plate").Set(
	css.Position(kw.Relative),
	css.Height(css.Px(128)),
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	css.Overflow(kw.Hidden),
	css.Isolation(kw.Isolate),
)

var plateLayer = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Absolute),
	css.MixBlendMode(kw.Multiply),
}

var PlateA = style.Class("plate-a").
	Set(css.Background(css.Var("ink-yellow"))).
	Before(plateLayer...).
	Before(css.Inset(css.Px(0), css.Px(0), css.Px(0), css.Pct(40)), css.BackgroundImage(dots("var(--ink-blue)")), css.BackgroundSize(css.Raw("9px 9px")))

var PlateB = style.Class("plate-b").
	Set(css.Background(css.Raw("linear-gradient(90deg, var(--ink-pink) 70%, var(--on-blue) 70%)"))).
	Before(plateLayer...).
	Before(css.Inset(css.Pct(20), css.Px(0), css.Px(0), css.Pct(45)), css.BorderRadius(css.Pct(50)), css.Background(css.Var("ink-yellow")))

var PlateC = style.Class("plate-c").
	Set(css.BackgroundColor(css.Var("on-blue")), css.BackgroundImage(dots("var(--ink-pink)")), css.BackgroundSize(css.Raw("7px 7px"))).
	Before(plateLayer...).
	Before(css.Inset(css.Pct(30), css.Pct(10), css.Px(0), css.Pct(30)), css.Background(css.Var("ink-blue")), css.Transform(css.Raw("rotate(-8deg)")))

var PlateD = style.Class("plate-d").
	Set(css.Background(css.Raw("linear-gradient(var(--ink-blue) 30%, #3B1E6E 30% 50%, var(--ink-pink) 50%)"))).
	Before(plateLayer...).
	Before(css.Inset(css.Px(0)), css.BackgroundImage(dots("var(--ink-yellow)")), css.BackgroundSize(css.Raw("12px 12px")), css.MixBlendMode(kw.Normal), css.Opacity(css.Num(0.85)))

var PlateLetter = style.Class("plate-letter").Set(
	css.Position(kw.Absolute),
	css.ZIndex(css.Int(1)),
	css.Left(css.Px(12)),
	css.Bottom(css.Px(4)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(4)),
	css.LineHeight(css.Num(1)),
	css.Color(css.Var("on-fill")),
)

var PlateLetterLight = style.Class("plate-letter-light").Set(css.Color(css.Var("on-blue")))

var PlateTag = style.Class("plate-tag").
	Set(label...).
	Set(
		css.Position(kw.Absolute),
		css.ZIndex(css.Int(2)),
		css.Right(css.Px(10)),
		css.Top(css.Px(10)),
		css.Padding(css.Px(4), css.Px(6)),
		css.Border(css.Px(2), kw.Solid, css.Var("on-fill")),
		css.Background(css.Var("on-blue")),
		css.Color(css.Var("on-fill")),
		css.FontSize(css.Rem(0.82)),
		css.LineHeight(css.Num(1)),
	)

var CardTitle = style.Class("card-title").Set(
	css.Margin(css.Rem(0.9), css.Rem(1.1), css.Px(0)),
	css.FontFamily(css.Var("sans")),
	css.FontSize(css.Rem(1.3)),
	css.FontWeight(css.Num(700)),
	css.LineHeight(css.Num(1.2)),
)

var CardText = style.Class("card-text").Set(
	css.Margin(css.Px(0), css.Rem(1.1)),
	css.Color(css.Var("ink")),
)

var FillSun = style.Class("fill-sun").Set(css.Background(css.Var("sun")))

var FillLime = style.Class("fill-lime").Set(css.Background(css.Var("lime")))

var FillLilac = style.Class("fill-lilac").Set(css.Background(css.Var("lilac")))

var FillCoral = style.Class("fill-coral").Set(css.Background(css.Var("coral")))

var Press = style.Class("press").Set(
	css.AlignSelf(kw.Stretch),
	css.Display(kw.Grid),
	css.GridTemplateColumns(css.Raw("repeat(auto-fit, minmax(min(100%, 360px), 1fr))")),
	css.Gap(css.Rem(2.5)),
	css.AlignItems(kw.Center),
)

var PressCopy = style.Class("press-copy").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1)),
)

var PressText = style.Class("press-text").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Em(32)),
	css.Padding(css.Rem(1), css.Rem(1.15)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow-s")),
)

var Bed = style.Class("bed").Set(
	css.Position(kw.Relative),
	line,
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow-l")),
	css.Transform(css.Raw("rotate(-0.8deg)")),
	css.Overflow(kw.Hidden),
)

var BedHead = style.Class("bed-head").Set(
	css.Display(kw.Flex),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(1)),
	css.Padding(css.Rem(0.6), css.Rem(1)),
	css.Background(css.Var("on-fill")),
	css.Color(css.Var("on-blue")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.85)),
	css.FontWeight(css.Num(700)),
)

var BedCount = BedHead.Descendant(style.Element("output")).Set(css.Color(css.Var("ink-yellow")))

var BedSheet = style.Class("bed-sheet").Set(
	css.Position(kw.Relative),
	css.Display(kw.Grid),
	css.PlaceItems(kw.Center),
	css.Height(css.Px(260)),
	css.BackgroundImage(dots("color-mix(in srgb, var(--mis-ink) 22%, transparent)")),
	css.BackgroundSize(css.Raw("12px 12px")),
	css.Overflow(kw.Hidden),
)

var BedFoot = style.Class("bed-foot").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.75), css.Rem(1)),
	css.BorderTop(css.Px(2), kw.Dashed, css.Var("line")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.875)),
	css.Color(css.Var("ink-2")),
)

var Thunk = style.Keyframes("thunk").
	From(css.Transform(css.Raw("rotate(var(--r)) scale(1.9)")), css.Opacity(css.Num(0))).
	To(css.Transform(css.Raw("rotate(var(--r)) scale(1)")), css.Opacity(css.Num(1)))

var ThunkAgain = style.Keyframes("thunk-again").
	From(css.Transform(css.Raw("rotate(var(--r)) scale(1.9)")), css.Opacity(css.Num(0))).
	To(css.Transform(css.Raw("rotate(var(--r)) scale(1)")), css.Opacity(css.Num(1)))

var stampLayer = []css.Property{
	css.Content(css.Raw("attr(data-ink)")),
	css.Position(kw.Absolute),
	css.Left(css.Px(0)),
	css.Top(css.Px(0)),
	css.MixBlendMode(css.Var("blend")),
}

var Stamp = style.Class("stamp").
	Set(
		css.Position(kw.Relative),
		css.Custom("r", css.Deg(-6)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(10)),
		css.LineHeight(css.Num(1)),
		css.Color(css.Var("mis-ink")),
		css.Transform(css.Raw("rotate(var(--r))")),
		css.AnimationName(css.Ident(Thunk.Name())),
		css.AnimationDuration(css.Ms(380)),
		css.AnimationTimingFunction(css.Raw("cubic-bezier(0.2, 1.4, 0.4, 1)")),
	).
	Before(stampLayer...).
	Before(css.Color(css.Var("ink-pink")), css.Transform(css.Raw("translate(0.05em, 0.03em)"))).
	After(stampLayer...).
	After(css.Color(css.Var("ink-yellow")), css.Transform(css.Raw("translate(-0.06em, -0.02em)")))

var StampAgain = style.Class("stamp-again").Set(css.Custom("r", css.Deg(5)), css.AnimationName(css.Ident(ThunkAgain.Name())))

var StampIdle = style.Class("stamp-idle").Set(css.Opacity(css.Num(0.2)), css.AnimationName(kw.None))

var PressButton = style.Class("press-button").Set(
	css.FontSize(css.Rem(1.5)),
	css.FontWeight(css.Num(400)),
	css.TextTransform(kw.Uppercase),
	css.LetterSpacing(css.Em(0.04)),
	css.Background(css.Var("ink-yellow")),
	css.Color(css.Var("on-fill")),
)

var Band = style.Class("band").Set(
	bleed,
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	css.Background(css.Var("lime")),
	css.Color(css.Var("on-fill")),
)

var Marquee = style.Class("marquee").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.JustifyContent(kw.Center),
	css.Gap(css.Rem(0.6), css.Rem(2)),
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(0.9), css.Rem(1.5)),
	css.ListStyle(kw.None),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.3)),
	css.FontWeight(css.Num(400)),
	css.TextTransform(kw.Uppercase),
	css.LetterSpacing(css.Em(0.02)),
)

var MarqueeDot = style.Class("marquee-dot").Set(css.Color(css.Var("on-fill")))

var Two = style.Class("two").Set(
	css.AlignSelf(kw.Stretch),
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(4)),
)

var TwoTitle = Two.Descendant(SectionTitle).Set(css.FontSize(css.Clamp(css.Rem(2.1), css.Vw(4), css.Rem(3.6))))

var TwoText = style.Class("two-text").Set(
	css.Flex(css.Raw("0.95 1 340px")),
	css.MinWidth(css.Px(0)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1)),
)

var TwoCode = style.Class("two-code").Set(
	css.Position(kw.Relative),
	css.Flex(css.Raw("1.05 1 400px")),
	css.MinWidth(css.Px(0)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(1.5)),
	css.PaddingTop(css.Rem(1.5)),
)

var Points = style.Class("points").Set(
	css.Display(kw.Grid),
	css.Gap(css.Rem(1)),
	css.Margin(css.Rem(1), css.Px(0), css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
)

var PointsItem = Points.Child(style.Element("li")).Set(
	css.Display(kw.Grid),
	css.GridTemplateColumns(css.Raw("36px 1fr")),
	css.Gap(css.Rem(0.75)),
	css.AlignItems(kw.Start),
	css.Padding(css.Rem(0.85), css.Rem(1)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow-s")),
)

var PointsTick = PointsItem.Descendant(style.Element("path")).Set(
	css.Fill(kw.None),
	css.Stroke(css.Var("code-str")),
	css.StrokeWidth(css.Num(4)),
	css.StrokeLinecap(kw.Round),
	css.StrokeLinejoin(kw.Round),
)

var Compiler = style.Class("compiler").Set(
	css.AlignSelf(kw.FlexEnd),
	css.MaxWidth(css.Pct(100)),
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(0.85), css.Rem(1)),
	line,
	css.Background(css.Var("on-fill")),
	css.Color(css.Var("on-blue")),
	css.BoxShadow(css.Var("shadow")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.9)),
	css.LineHeight(css.Num(1.6)),
	css.WhiteSpace(kw.PreWrap),
	css.OverflowWrap(kw.Anywhere),
	css.Transform(css.Raw("rotate(-1deg)")),
)

var CompilerPrompt = style.Class("compiler-prompt").Set(css.Color(css.Var("ink-yellow")))

var CompilerError = style.Class("compiler-error").Set(css.Color(css.Var("ink-pink")))

var Caption = style.Class("caption").Set(
	css.Margin(css.Px(0)),
	css.AlignSelf(kw.FlexEnd),
	css.TextAlign(kw.Right),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.875)),
	css.Color(css.Var("ink-2")),
)

var Steps = style.Class("steps").
	Set(
		css.AlignSelf(kw.Stretch),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("repeat(3, minmax(0, 1fr))")),
		css.MarginTop(css.Rem(1.5)),
		line,
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow-l")),
	).
	Media(narrow, css.GridTemplateColumns(css.Raw("minmax(0, 1fr)")))

var Step = style.Class("step").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.6)),
	css.MinWidth(css.Px(0)),
	css.Padding(css.Rem(1.6), css.Rem(1.4), css.Rem(1.5)),
)

var StepNext = Step.Next(Step).
	Set(css.BorderLeft(css.Var("border"), kw.Solid, css.Var("line"))).
	Media(narrow, css.BorderLeft(css.Px(0), kw.Solid, css.Var("line")), css.BorderTop(css.Var("border"), kw.Solid, css.Var("line")))

var StepNumber = style.Class("step-number").Set(
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(6)),
	css.FontWeight(css.Num(400)),
	css.LineHeight(css.Num(0.8)),
	css.MarginBottom(css.Rem(0.5)),
)

var StepTitle = style.Class("step-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("sans")),
	css.FontSize(css.Rem(1.35)),
	css.FontWeight(css.Num(700)),
	css.LineHeight(css.Num(1.2)),
)

var StepText = style.Class("step-text").Set(
	css.Margin(css.Px(0), css.Px(0), css.Rem(0.5)),
	css.Color(css.Var("ink")),
)

var StepCode = Step.Descendant(Code).Set(
	css.AlignSelf(kw.Stretch),
	css.Border(css.Px(2), kw.Dashed, css.Var("line")),
	css.Background(css.Var("ground")),
)

var StepCodeBody = Step.Descendant(CodeBody).Set(css.Padding(css.Rem(0.6), css.Rem(0.8)), css.FontSize(css.Rem(0.85)))

var Closing = style.Class("closing").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.AlignItems(kw.Center),
		css.JustifyContent(kw.SpaceBetween),
		css.Gap(css.Rem(1.5)),
		css.Padding(css.Rem(5.5), css.Px(0)),
	).
	Before(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Top(css.Px(0)),
		css.Bottom(css.Px(0)),
		css.Left(css.Raw("calc(50% - 50vw)")),
		css.Width(css.Vw(100)),
		css.ZIndex(css.Int(-1)),
		css.BackgroundImage(dots("var(--ink-blue)")),
		css.BackgroundSize(css.Raw("10px 10px")),
		css.MaskImage(css.Raw("linear-gradient(100deg, transparent 30%, #000 95%)")),
		css.Opacity(css.Num(0.5)),
	)

var ClosingText = style.Class("closing-text").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.5)),
	css.MaxWidth(css.Ch(60)),
)

var ClosingLead = style.Class("closing-lead").Set(
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(0.75), css.Rem(1)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("panel")),
	css.FontSize(css.Rem(1.05)),
)

var ButtonBlue = style.Class("button-blue").Set(css.Background(css.Var("ink-blue")), css.Color(css.Var("on-blue")))

var Footer = style.Class("footer").
	Set(
		bleed,
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(1.25)),
		css.Padding(css.Rem(2.75), css.Raw("max(1.5rem, calc(50vw - 575px))"), css.Rem(2)),
		css.Background(css.Var("on-fill")),
		css.Color(css.Var("on-blue")),
	).
	Media(narrow, css.Padding(css.Rem(2.25), css.Rem(1), css.Rem(1.75)))

var FooterBig = style.Class("footer-big").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(3.25), css.Vw(10), css.Rem(9.4))),
	css.FontWeight(css.Num(400)),
	css.LineHeight(css.Num(0.85)),
	css.TextTransform(kw.Uppercase),
)

var FooterMis = FooterBig.Descendant(Mis).Set(css.Color(css.Var("on-blue"))).After(css.MixBlendMode(kw.Screen))

var FooterRow = style.Class("footer-row").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(1)),
)

var ColorBar = style.Class("color-bar").Set(
	css.Display(kw.Block),
	css.Width(css.Px(132)),
	css.Height(css.Px(16)),
	css.Border(css.Px(2), kw.Solid, css.Var("on-blue")),
	css.Background(css.Raw("linear-gradient(90deg, var(--ink-pink) 0 20%, var(--ink-blue) 20% 40%, var(--ink-yellow) 40% 60%, var(--lime) 60% 80%, var(--lilac) 80%)")),
)

var FooterFine = style.Class("footer-fine").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(0.75)),
	css.PaddingTop(css.Rem(0.9)),
	css.BorderTop(css.Px(1), kw.Dashed, css.Raw("color-mix(in srgb, var(--on-blue) 40%, transparent)")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.85)),
	css.Color(css.Raw("color-mix(in srgb, var(--on-blue) 80%, transparent)")),
)
