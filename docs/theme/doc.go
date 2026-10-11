package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
	"github.com/Toms223/WebGo/supports"
)

var Doc = style.Class("doc")

var DocFirst = Doc.Child(style.Any()).FirstChild(css.MarginTop(css.Px(0)))

var DocParagraph = Doc.Descendant(style.Element("p")).Set(
	css.Margin(css.Px(0), css.Px(0), css.Rem(1.15)),
	css.MaxWidth(css.Ch(68)),
)

var DocHeading = style.Class("doc-heading").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.6)),
		css.Margin(css.Rem(2.75), css.Px(0), css.Rem(1)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(1.65)),
		css.FontWeight(css.Num(800)),
		css.LetterSpacing(css.Em(-0.02)),
		css.LineHeight(css.Num(1.1)),
	).
	Before(
		css.Content(css.String("")),
		css.Flex(kw.None),
		css.Width(css.Em(0.5)),
		css.Height(css.Em(0.5)),
		css.Border(css.Var("border"), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Pct(50)),
		css.Background(css.Var("coral")),
	)

var DocSubheading = style.Class("doc-subheading").
	Set(label...).
	Set(
		css.Display(kw.InlineBlock),
		css.Margin(css.Rem(2), css.Px(0), css.Rem(0.75)),
		css.Padding(css.Rem(0.2), css.Rem(0.65)),
		line,
		css.BorderRadius(css.Px(999)),
		css.Background(css.Var("lime")),
		css.Color(css.Var("on-fill")),
	)

var docLists = []css.Property{
	css.Margin(css.Px(0), css.Px(0), css.Rem(1.15)),
	css.PaddingLeft(css.Rem(1.3)),
	css.MaxWidth(css.Ch(68)),
}

var DocList = Doc.Descendant(style.Element("ul")).Set(docLists...)

var DocOrderedList = Doc.Descendant(style.Element("ol")).Set(docLists...)

var DocItem = Doc.Descendant(style.Element("li")).
	Set(css.MarginBottom(css.Rem(0.4))).
	Marker(
		css.Color(css.Var("coral")),
		css.FontWeight(css.Num(900)),
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
		css.TextUnderlineOffset(css.Px(3)),
	).
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
)

var DocPre = Doc.Descendant(style.Element("pre")).Set(
	css.Position(kw.Relative),
	css.Margin(css.Px(0), css.Px(0), css.Rem(1.75)),
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
	css.Right(css.Px(0)),
	css.Padding(css.Rem(0.15), css.Rem(0.7)),
	css.BorderLeft(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderBottomLeftRadius(css.Px(10)),
	css.Background(css.Var("lilac")),
	css.Color(css.Var("on-fill")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.66)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.12)),
	css.LineHeight(css.Num(1.6)),
	css.TextTransform(kw.Uppercase),
}

var DocPreGo = DocPre.And(style.Class("lang-go")).Before(css.Content(css.String("go"))).Before(langTab...)

var DocPreCSS = DocPre.And(style.Class("lang-css")).Before(css.Content(css.String("css"))).Before(langTab...)

var DocPreShell = DocPre.And(style.Class("lang-sh")).Before(css.Content(css.String("sh"))).Before(langTab...)

var DocPreCode = DocPre.Descendant(style.Element("code")).Set(
	css.Padding(css.Px(0)),
	css.BorderRadius(css.Px(0)),
	css.Background(kw.None),
	css.Color(kw.Inherit),
	css.FontSize(kw.Inherit),
	css.FontWeight(kw.Inherit),
)

var DocQuote = Doc.Descendant(style.Element("blockquote")).
	Set(
		css.Position(kw.Relative),
		css.Margin(css.Px(0), css.Px(0), css.Rem(1.5)),
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

var DocQuoteLast = DocQuote.Descendant(style.Element("p")).LastChild(css.MarginBottom(css.Px(0)))

var DocQuoteCode = DocQuote.Descendant(style.Element("code")).Set(
	css.Background(css.RGBA(255, 255, 255, 0.6)),
	css.Color(css.Var("on-fill")),
)

var DocQuoteLink = DocQuote.Descendant(style.Element("a")).Set(css.Color(css.Var("on-fill")))

var DocRule = Doc.Descendant(style.Element("hr")).Set(
	css.Height(css.Var("border")),
	css.Margin(css.Rem(2.75), css.Px(0)),
	css.Border(css.Px(0)),
	css.BorderRadius(css.Px(2)),
	css.Background(css.Var("line")),
)
