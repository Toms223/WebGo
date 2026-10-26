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
		css.FontSize(css.Clamp(css.Rem(1.6), css.Vw(2.6), css.Rem(2.1))),
		css.FontWeight(css.Num(800)),
		css.LetterSpacing(css.Em(-0.025)),
		css.LineHeight(css.Num(1.1)),
	).
	Before(
		css.Content(css.String("")),
		css.Flex(kw.None),
		css.Width(css.Em(0.45)),
		css.Height(css.Em(0.45)),
		css.Border(css.Var("border"), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Pct(50)),
		css.Background(css.Var("coral")),
	)

var DocSubheading = style.Class("doc-subheading").
	Set(label...).
	Set(
		css.ScrollMarginTop(css.Rem(1)),
		css.Display(kw.InlineBlock),
		css.Margin(css.Rem(1), css.Px(0), css.Px(0)),
		css.Padding(css.Rem(0.2), css.Rem(0.65)),
		line,
		css.BorderRadius(css.Px(999)),
		css.Background(css.Var("lime")),
		css.Color(css.Var("on-fill")),
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
		css.BorderRadius(css.Px(4)),
		css.TextDecoration(kw.Underline),
		css.TextDecorationColor(css.Var("coral")),
		css.TextDecorationThickness(css.Px(3)),
		css.TextUnderlineOffset(css.Em(0.3)),
		css.TextDecorationSkipInk(kw.None),
	).
	FocusVisible(ring...).
	Nest(style.Self().
		Hover(css.Background(css.Var("lilac"))).
		Supports(colorMix, css.Background(css.Raw("color-mix(in srgb, var(--coral) 30%, transparent)"))))

var DocStrong = Doc.Descendant(style.Element("strong")).Set(css.FontWeight(css.Num(700)))

var DocCode = Doc.Descendant(style.Element("code")).Set(
	css.Padding(css.Em(0.12), css.Em(0.45)),
	css.BorderRadius(css.Var("r-s")),
	css.Background(css.Var("code-inline")),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Em(0.86)),
	css.FontWeight(css.Num(600)),
	css.OverflowWrap(kw.Anywhere),
)

var CodeBlock = style.Class("codeblock").Set(
	css.Position(kw.Relative),
	css.AlignSelf(kw.Stretch),
	css.MinWidth(css.Px(0)),
)

var DocPre = Doc.Descendant(style.Element("pre")).Set(
	css.Position(kw.Relative),
	css.AlignSelf(kw.Stretch),
	css.Margin(css.Rem(0.35), css.Px(0)),
	css.Padding(css.Rem(1.6), css.Rem(1.25), css.Rem(1.1)),
	css.OverflowX(kw.Auto),
	line,
	css.BorderRadius(css.Var("r-l")),
	css.BoxShadow(css.Var("shadow")),
	css.Background(css.Var("code")),
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
	css.Left(css.Px(18)),
	css.Padding(css.Rem(0.15), css.Rem(0.7)),
	css.BorderLeft(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderRight(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderBottomLeftRadius(css.Px(10)),
	css.BorderBottomRightRadius(css.Px(10)),
	css.Background(css.Var("lilac")),
	css.Color(css.Var("on-fill")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.66)),
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
	css.BorderRadius(css.Px(0)),
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
		css.Margin(css.Rem(0.35), css.Px(0)),
		css.Padding(css.Rem(0.9), css.Rem(1.1), css.Rem(0.9), css.Rem(3.4)),
		line,
		css.BorderRadius(css.Var("r")),
		css.BoxShadow(css.Var("shadow")),
		css.Background(css.Var("lilac")),
		css.Color(css.Var("on-fill")),
	).
	Before(
		css.Content(css.String("!")),
		css.Position(kw.Absolute),
		css.Top(css.Rem(0.85)),
		css.Left(css.Rem(1.1)),
		css.Display(kw.Grid),
		css.PlaceItems(kw.Center),
		css.Width(css.Px(28)),
		css.Height(css.Px(28)),
		line,
		css.BorderRadius(css.Pct(50)),
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("display")),
		css.FontWeight(css.Num(800)),
		css.LineHeight(css.Num(1)),
	)

var NoteText = Note.Child(style.Element("p")).Set(css.Margin(css.Px(0)), css.MaxWidth(kw.None))

var NoteCode = Note.Descendant(style.Element("code")).Set(
	css.Background(css.RGBA(255, 255, 255, 0.6)),
	css.Color(css.Var("on-fill")),
)

var NoteLink = Note.Descendant(style.Element("a")).Set(css.Color(css.Var("on-fill")))

var DocRule = Doc.Descendant(style.Element("hr")).Set(
	css.AlignSelf(kw.Stretch),
	css.Height(css.Var("border")),
	css.Margin(css.Rem(1), css.Px(0)),
	css.Border(css.Px(0)),
	css.BorderRadius(css.Px(2)),
	css.Background(css.Var("line")),
)
