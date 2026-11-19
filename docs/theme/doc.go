package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/Toms223/WebGo/supports"
)

var Doc = style.Class("doc").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1.1)),
	css.MinWidth(css.Px(0)),
)

var DocProse = Doc.Child(style.Element("p")).Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(68)),
	css.FontSize(css.Rem(1.02)),
)

var DocHeading = style.Class("doc-heading").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.Margin(css.Rem(0.5), css.Px(0), css.Px(0)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Clamp(css.Rem(1.7), css.Vw(2.8), css.Rem(2.3))),
		css.FontWeight(css.Num(400)),
		css.LetterSpacing(css.Em(0.01)),
		css.LineHeight(css.Num(1.05)),
		css.TextTransform(kw.Uppercase),
	).
	Before(
		css.Content(css.String("")),
		css.Flex(kw.None),
		css.Width(css.Em(0.42)),
		css.Height(css.Em(0.42)),
		css.Border(css.Var("border"), kw.Solid, css.Var("line")),
		css.Background(css.Var("coral")),
		css.Rotate(css.Deg(-8)),
	)

var DocSubheading = style.Class("doc-subheading").
	Set(label...).
	Set(
		css.ScrollMarginTop(css.Rem(1)),
		css.Display(kw.InlineBlock),
		css.Margin(css.Rem(1.2), css.Px(0), css.Rem(0.1)),
		css.Padding(css.Rem(0.3), css.Rem(0.7)),
		line,
		css.Background(css.Var("sun")),
		css.BoxShadow(css.Var("shadow-s")),
		css.Color(css.Var("on-fill")),
		css.FontSize(css.Rem(0.82)),
		css.Rotate(css.Deg(-1.2)),
	).
	Media(narrow, css.ScrollMarginTop(css.Rem(5.5)))

var docLists = []css.Property{
	css.Margin(css.Px(0)),
	css.PaddingLeft(css.Rem(1.3)),
	css.MaxWidth(css.Ch(68)),
}

var DocList = Doc.Descendant(style.Element("ul")).Set(docLists...)

var DocOrderedList = Doc.Descendant(style.Element("ol")).Set(docLists...)

var DocItem = Doc.Descendant(style.Element("li")).
	Set(css.Margin(css.Rem(0.2), css.Px(0))).
	Marker(
		css.Color(css.Var("coral")),
		css.FontWeight(css.Num(900)),
	)

var DocNestedList = Doc.Descendant(style.Element("li")).Child(style.Element("ul")).Set(
	css.MarginTop(css.Rem(0.35)),
)

var colorMix = supports.Property(css.Color(css.Raw("color-mix(in srgb, red, blue)")))

var DocLink = Doc.Descendant(style.Element("a")).
	Set(
		css.Color(css.Var("ink")),
		css.FontWeight(css.Num(600)),
		css.TextDecoration(kw.Underline),
		css.TextDecorationColor(css.Var("coral")),
		css.TextDecorationThickness(css.Px(3)),
		css.TextUnderlineOffset(css.Em(0.3)),
		css.TextDecorationSkipInk(kw.None),
	).
	FocusVisible(ring...).
	Nest(style.Self().
		Hover(css.TextDecorationStyle(kw.Wavy), css.TextDecorationThickness(css.Px(2))).
		Supports(colorMix, css.Background(css.Raw("linear-gradient(transparent 55%, color-mix(in srgb, var(--sun) 55%, transparent) 55%)"))))

var DocStrong = Doc.Descendant(style.Element("strong")).Set(css.FontWeight(css.Num(700)))

var DocCode = Doc.Descendant(style.Element("code")).Set(
	css.Padding(css.Em(0.08), css.Em(0.38)),
	css.Border(css.Px(1), kw.Solid, css.Var("line")),
	css.Background(css.Var("code-inline")),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Em(0.86)),
	css.FontWeight(css.Num(600)),
	css.OverflowWrap(kw.Anywhere),
)

var tape = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Absolute),
	css.Top(css.Px(-13)),
	css.ZIndex(css.Int(2)),
	css.Width(css.Px(104)),
	css.Height(css.Px(28)),
	css.Background(css.Var("tape")),
	css.BorderLeft(css.Px(1), kw.Dashed, css.Var("line")),
	css.BorderRight(css.Px(1), kw.Dashed, css.Var("line")),
	css.ClipPath(css.Raw("polygon(0 8%, 6% 0, 12% 10%, 20% 2%, 30% 9%, 42% 0, 55% 8%, 68% 1%, 80% 9%, 92% 0, 100% 7%, 100% 92%, 94% 100%, 84% 91%, 72% 100%, 60% 92%, 46% 100%, 34% 91%, 22% 100%, 10% 92%, 0 100%)")),
	css.PointerEvents(kw.None),
}

var CodeBlock = style.Class("codeblock").
	Set(
		css.Position(kw.Relative),
		css.AlignSelf(kw.Stretch),
		css.MinWidth(css.Px(0)),
		css.Margin(css.Rem(0.6), css.Px(0), css.Rem(0.3)),
		css.Rotate(css.Deg(-0.4)),
	).
	Before(tape...).
	Before(css.Left(css.Pct(18)), css.Rotate(css.Deg(-5))).
	After(tape...).
	After(css.Right(css.Pct(28)), css.Rotate(css.Deg(4)), css.Background(css.Var("coral")), css.Opacity(css.Num(0.55))).
	Media(narrow, css.Rotate(kw.None))

var printout = []css.Property{
	css.BackgroundColor(css.Var("code")),
	css.BackgroundImage(css.Raw("linear-gradient(color-mix(in srgb, var(--code) 72%, transparent), color-mix(in srgb, var(--code) 72%, transparent)), var(--noise)")),
}

var DocPre = Doc.Descendant(style.Element("pre")).Set(
	css.Position(kw.Relative),
	css.AlignSelf(kw.Stretch),
	css.Margin(css.Rem(0.35), css.Px(0)),
	css.Padding(css.Rem(2.4), css.Rem(1.25), css.Rem(1.1)),
	css.OverflowX(kw.Auto),
	line,
	css.BoxShadow(css.Var("shadow")),
).Set(printout...).Set(
	css.Color(css.Var("code-on")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.875)),
	css.FontWeight(css.Num(500)),
	css.LineHeight(css.Num(1.7)),
	css.TabSize(css.Num(4)),
)

var langTab = []css.Property{
	css.Position(kw.Absolute),
	css.Top(css.Px(0)),
	css.Left(css.Px(0)),
	css.Padding(css.Rem(0.2), css.Rem(0.75)),
	css.BorderRight(css.Px(2), kw.Dashed, css.Var("line")),
	css.BorderBottom(css.Px(2), kw.Dashed, css.Var("line")),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.8125)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.12)),
	css.LineHeight(css.Num(1.6)),
	css.TextTransform(kw.Uppercase),
}

var LangGo = style.Class("lang-go")

var LangCSS = style.Class("lang-css")

var LangShell = style.Class("lang-sh")

var LangMarkdown = style.Class("lang-md")

var LangPlain = style.Class("lang-plain")

var LangHTML = style.Class("lang-html")

var DocPreGo = DocPre.And(LangGo).Before(css.Content(css.String("go"))).Before(langTab...)

var DocPreCSS = DocPre.And(LangCSS).Before(css.Content(css.String("css"))).Before(langTab...)

var DocPreShell = DocPre.And(LangShell).Before(css.Content(css.String("sh"))).Before(langTab...)

var DocPreMarkdown = DocPre.And(LangMarkdown).Before(css.Content(css.String("md"))).Before(langTab...)

var DocPreHTML = DocPre.And(LangHTML).Before(css.Content(css.String("html"))).Before(langTab...)

var DocPreCode = DocPre.Descendant(style.Element("code")).Set(
	css.Padding(css.Px(0)),
	css.Border(css.Px(0)),
	css.Background(kw.None),
	css.Color(kw.Inherit),
	css.FontSize(kw.Inherit),
	css.FontWeight(kw.Inherit),
	css.OverflowWrap(kw.Normal),
)

var Note = style.Class("note").
	Set(
		css.Position(kw.Relative),
		css.AlignSelf(kw.Stretch),
		css.MaxWidth(css.Ch(76)),
		css.Margin(css.Rem(0.9), css.Px(0), css.Rem(0.5)),
		css.Padding(css.Rem(1.1), css.Rem(1.2), css.Rem(1), css.Rem(3.5)),
		line,
		css.BoxShadow(css.Var("shadow")),
		css.Background(css.Var("sun")),
		css.Color(css.Var("on-fill")),
		css.Rotate(css.Deg(0.5)),
	).
	Before(
		css.Content(css.String("!")),
		css.Position(kw.Absolute),
		css.Top(css.Rem(0.95)),
		css.Left(css.Rem(1.1)),
		css.Display(kw.Grid),
		css.PlaceItems(kw.Center),
		css.Width(css.Px(28)),
		css.Height(css.Px(28)),
		css.Background(css.Var("on-fill")),
		css.Color(css.Var("sun")),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(1.1)),
		css.FontWeight(css.Num(400)),
		css.LineHeight(css.Num(1)),
		css.Rotate(css.Deg(-6)),
	).
	After(tape...).
	After(css.Left(css.Pct(40)), css.Width(css.Px(80)), css.Height(css.Px(22)), css.Top(css.Px(-11)), css.Rotate(css.Deg(-3)), css.Background(css.Var("coral")), css.Opacity(css.Num(0.6))).
	Media(narrow, css.Rotate(kw.None))

var NoteText = Note.Child(style.Element("p")).Set(css.Margin(css.Px(0)), css.MaxWidth(kw.None))

var NoteCode = Note.Descendant(style.Element("code")).Set(
	css.Background(css.Var("code")),
	css.Color(css.Var("code-on")),
)

var NoteLink = Note.Descendant(style.Element("a")).Set(css.Color(css.Var("on-fill")))

var DocRule = Doc.Descendant(style.Element("hr")).Set(
	css.AlignSelf(kw.Stretch),
	css.Height(css.Px(0)),
	css.Margin(css.Rem(1), css.Px(0)),
	css.Border(css.Px(0)),
	css.BorderTop(css.Var("border"), kw.Dashed, css.Var("line")),
)
