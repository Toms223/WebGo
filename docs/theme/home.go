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
	css.OverflowX(kw.Hidden),
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
		css.Gap(css.Rem(4)),
	).
	Media(narrow, css.Padding(css.Px(0), css.Rem(1)), css.Gap(css.Rem(3)))

var Hero = style.Class("hero").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(2.5)),
		css.Padding(css.Rem(4.5), css.Px(0), css.Rem(1)),
	).
	Media(narrow, css.Padding(css.Rem(2.5), css.Px(0), css.Px(0)))

var HeroCopy = style.Class("hero-copy").Set(
	css.Flex(css.Raw("1 1 440px")),
	css.MinWidth(css.Px(0)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1.5)),
)

var HeroInstall = HeroCopy.Child(Code).Set(
	css.AlignSelf(kw.Stretch),
	css.MaxWidth(css.Px(560)),
	css.MinWidth(css.Px(0)),
)

var HeroSample = style.Class("hero-sample").Set(
	css.Flex(css.Raw("1 1 400px")),
	css.MinWidth(css.Px(0)),
)

var HeroTitle = style.Class("hero-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2.6), css.Vw(6.2), css.Rem(5))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.04)),
	css.LineHeight(css.Num(0.98)),
)

var Highlight = style.Class("highlight").Set(
	css.Display(kw.InlineBlock),
	css.Padding(css.Px(0), css.Em(0.18)),
	css.Border(css.Var("border-l"), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Px(14)),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.BoxShadow(css.Var("shadow")),
	css.Transform(css.Raw("rotate(-1.5deg)")),
)

var Lead = style.Class("lead").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(54)),
	css.Color(css.Var("ink-2")),
	css.FontSize(css.Clamp(css.Rem(1.05), css.Vw(1.6), css.Rem(1.3))),
	css.LineHeight(css.Num(1.5)),
)

var Actions = style.Class("actions").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.9)),
)

var Chip = style.Class("chip").
	Set(label...).
	Set(
		css.Display(kw.InlineFlex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.5)),
		css.Padding(css.Rem(0.25), css.Rem(0.7)),
		line,
		css.BorderRadius(css.Px(999)),
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.LetterSpacing(css.Em(0.08)),
		css.LineHeight(css.Num(1.4)),
	).
	Before(
		css.Content(css.String("")),
		css.Width(css.Px(9)),
		css.Height(css.Px(9)),
		css.Border(css.Px(1.5), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Pct(50)),
		css.Background(css.Var("coral")),
	)

var Code = style.Class("code").Set(
	line,
	css.BorderRadius(css.Var("r-l")),
	css.Background(css.Var("code")),
	css.Color(css.Var("code-on")),
	css.BoxShadow(css.Var("shadow")),
	css.Overflow(kw.Hidden),
)

var CodeBar = style.Class("code-bar").
	Set(label...).
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.75)),
		css.Padding(css.Rem(0.55), css.Rem(1)),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
		css.Background(css.Var("lilac")),
		css.Color(css.Var("on-fill")),
		css.LetterSpacing(css.Em(0.1)),
	)

var CodeBody = Code.Descendant(style.Element("pre")).Set(
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(1.1), css.Rem(1.25)),
	css.OverflowX(kw.Auto),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.875)),
	css.FontWeight(css.Num(500)),
	css.LineHeight(css.Num(1.7)),
	css.TabSize(css.Num(4)),
)

var CodeTilt = style.Class("code-tilt").Set(
	css.Transform(css.Raw("rotate(1deg)")),
	css.BoxShadow(css.Var("shadow-l")),
)

var CodeQuiet = style.Class("code-quiet").Set(
	css.BorderRadius(css.Var("r")),
	css.BoxShadow(kw.None),
)

var Keyword = style.Class("k").Set(css.Color(css.Var("code-key")))

var Str = style.Class("s").Set(css.Color(css.Var("code-str")))

var Fn = style.Class("f").Set(css.Color(css.Var("code-fn")))

var Typ = style.Class("t").Set(css.Color(css.Var("code-type")))

var Dim = style.Class("c").Set(css.Color(css.Var("code-dim")))

var Eyebrow = style.Class("eyebrow").
	Set(label...).
	Set(css.Color(css.Var("ink-2")), css.LetterSpacing(css.Em(0.14)))

var Section = style.Class("section").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(1.25)),
)

var SectionTitle = style.Class("section-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(1.8), css.Vw(4), css.Rem(2.8))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.03)),
	css.LineHeight(css.Num(1.05)),
)

var Tiles = style.Class("tiles").Set(
	css.Display(kw.Grid),
	css.GridTemplateColumns(css.Raw("repeat(auto-fit, minmax(240px, 1fr))")),
	css.Gap(css.Rem(1.25)),
)

var Card = style.Class("card").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(0.75)),
		css.Padding(css.Rem(1.4)),
		line,
		css.BorderRadius(css.Var("r-l")),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow")),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		css.Transition(css.Raw("transform 120ms ease, box-shadow 120ms ease")),
	).
	Hover(lift...).
	FocusVisible(ring...)

var CardIcon = style.Class("card-icon").Set(
	css.Display(kw.Grid),
	css.PlaceItems(kw.Center),
	css.Width(css.Px(44)),
	css.Height(css.Px(44)),
	line,
	css.BorderRadius(css.Var("r")),
	css.Color(css.Var("on-fill")),
)

var CardTitle = style.Class("card-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.35)),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.02)),
	css.LineHeight(css.Num(1.1)),
)

var CardText = style.Class("card-text").Set(
	css.Margin(css.Px(0)),
	css.Color(css.Var("ink-2")),
)

var FillSun = style.Class("fill-sun").Set(css.Background(css.Var("sun")))

var FillLime = style.Class("fill-lime").Set(css.Background(css.Var("lime")))

var FillLilac = style.Class("fill-lilac").Set(css.Background(css.Var("lilac")))

var FillCoral = style.Class("fill-coral").Set(css.Background(css.Var("coral")))

var Band = style.Class("band").Set(
	css.Margin(css.Rem(4), css.Raw("calc(50% - 50vw)"), css.Px(0)),
	css.BorderTop(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	css.Background(css.Var("lime")),
	css.Color(css.Var("on-fill")),
)

var Marquee = style.Class("marquee").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.JustifyContent(kw.Center),
		css.Gap(css.Rem(0.9), css.Rem(2.25)),
		css.Margin(css.Px(0)),
		css.Padding(css.Rem(1), css.Rem(1.5)),
		css.ListStyle(kw.None),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(1.1)),
		css.FontWeight(css.Num(800)),
		css.LetterSpacing(css.Em(-0.01)),
	)

var MarqueeDot = style.Class("marquee-dot").Set(css.Color(css.Var("line")))

var Two = style.Class("two").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(2.5)),
	css.PaddingTop(css.Rem(4)),
)

var TwoText = style.Class("two-text").Set(
	css.Flex(css.Raw("1 1 360px")),
	css.MinWidth(css.Px(0)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1)),
)

var TwoCode = style.Class("two-code").Set(
	css.Flex(css.Raw("1.2 1 420px")),
	css.MinWidth(css.Px(0)),
)

var Sticker = style.Class("sticker").Set(
	css.Display(kw.InlineBlock),
	css.Padding(css.Rem(0.25), css.Rem(0.75)),
	line,
	css.BorderRadius(css.Var("r")),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.BoxShadow(css.Var("shadow-s")),
	css.FontFamily(css.Var("display")),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.02)),
	css.Transform(css.Raw("rotate(-3deg)")),
)

var Points = style.Class("points").Set(
	css.Margin(css.Px(0)),
	css.PaddingLeft(css.Rem(1.3)),
	css.Color(css.Var("ink-2")),
)

var PointsItem = Points.Child(style.Element("li")).
	Set(css.MarginBottom(css.Rem(0.4))).
	Marker(css.Color(css.Var("coral")), css.FontWeight(css.Num(900)))

var Steps = style.Class("steps").Set(
	css.Display(kw.Grid),
	css.GridTemplateColumns(css.Raw("repeat(auto-fit, minmax(260px, 1fr))")),
	css.Gap(css.Rem(1.25)),
)

var Step = style.Class("step").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.6)),
	css.Padding(css.Rem(1.4)),
	line,
	css.BorderRadius(css.Var("r-l")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow")),
)

var StepNumber = style.Class("step-number").Set(
	css.Display(kw.Grid),
	css.PlaceItems(kw.Center),
	css.Width(css.Px(40)),
	css.Height(css.Px(40)),
	line,
	css.BorderRadius(css.Pct(50)),
	css.Color(css.Var("on-fill")),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.1)),
	css.FontWeight(css.Num(800)),
)

var StepTitle = style.Class("step-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.25)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(-0.02)),
)

var StepText = style.Class("step-text").Set(
	css.Margin(css.Px(0)),
	css.Color(css.Var("ink-2")),
)

var StepCode = Step.Descendant(Code).Set(css.AlignSelf(kw.Stretch))

var StepCodeBody = Step.Descendant(CodeBody).Set(css.Padding(css.Rem(0.6), css.Rem(0.9)), css.FontSize(css.Rem(0.8)))

var Closing = style.Class("closing").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(1.5)),
	css.Padding(css.Rem(2.5)),
	css.Border(css.Var("border-l"), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Var("r-l")),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.BoxShadow(css.Var("shadow")),
)

var ClosingText = style.Class("closing-text").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.5)),
	css.MaxWidth(css.Ch(56)),
)

var ClosingLead = style.Class("closing-lead").Set(
	css.Margin(css.Px(0)),
	css.FontSize(css.Rem(1.05)),
)

var Footer = style.Class("footer").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.Center),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(1)),
	css.Padding(css.Rem(1.75), css.Px(0), css.Rem(3)),
	css.Color(css.Var("ink-2")),
)
