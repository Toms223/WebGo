package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var arrive = []css.Property{
	css.AnimationName(css.Ident(Reveal.Name())),
	css.AnimationDuration(css.Ms(420)),
	css.AnimationTimingFunction(css.Raw("cubic-bezier(0.16, 1, 0.3, 1)")),
	css.AnimationFillMode(kw.Both),
}

var News = style.Class("news").
	Set(
		css.Width(css.Pct(100)),
		css.MaxWidth(css.Px(1180)),
		css.Margin(css.Px(0), kw.Auto),
		css.Padding(css.Rem(3.5), css.Rem(1.5), css.Rem(5)),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(4.5)),
	).
	Media(narrow, css.Padding(css.Rem(2), css.Rem(1), css.Rem(3)), css.Gap(css.Rem(3)))

var NewsHead = style.Class("news-head").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.AlignItems(kw.FlexEnd),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(1.5)),
)

var NewsMasthead = style.Class("news-masthead").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(4), css.Vw(13), css.Rem(10.5))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.06)),
	css.LineHeight(css.Num(0.82)),
)

var NewsIntro = style.Class("news-intro").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(36)),
	css.FontSize(css.Clamp(css.Rem(1.1), css.Vw(1.7), css.Rem(1.4))),
	css.LineHeight(css.Num(1.4)),
	css.Color(css.Var("ink")),
)

var NewsCover = style.Class("news-cover").
	Set(
		css.ContainerType(kw.InlineSize),
		css.Position(kw.Relative),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.JustifyContent(kw.FlexEnd),
		css.Gap(css.Rem(0.5)),
		css.AspectRatio(css.Raw("4 / 3")),
		css.Padding(css.Raw("5cqi")),
		line,
		css.BorderRadius(css.Var("r")),
		css.Color(css.Var("on-fill")),
		css.Overflow(kw.Hidden),
		css.MinWidth(css.Px(0)),
	)

var NewsCoverWide = NewsCover.And(style.Class("news-cover-wide")).Set(css.AspectRatio(css.Raw("16 / 7"))).
	Media(narrow, css.AspectRatio(css.Raw("4 / 3")))

var NewsCoverVersion = style.Class("news-cover-version").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Raw("33cqi")),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.07)),
	css.LineHeight(css.Num(0.8)),
)

var NewsCoverWideVersion = NewsCoverWide.Descendant(NewsCoverVersion).Set(css.FontSize(css.Raw("24cqi")))

var NewsCoverName = style.Class("news-cover-name").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Raw("8.5cqi")),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.03)),
	css.LineHeight(css.Num(1)),
)

var NewsCoverWideName = NewsCoverWide.Descendant(NewsCoverName).Set(css.FontSize(css.Raw("5.5cqi")))

var NewsCoverCard = style.Class("news-cover-card").Set(
	css.Position(kw.Absolute),
	css.Right(css.Raw("5cqi")),
	css.Top(css.Raw("7cqi")),
	css.MaxWidth(css.Pct(62)),
	css.Padding(css.Raw("1.6cqi 2.4cqi")),
	line,
	css.BorderRadius(css.Px(10)),
	css.Background(css.Var("code")),
	css.Color(css.Var("code-on")),
	css.BoxShadow(css.Var("shadow")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Raw("max(0.82rem, 2.6cqi)")),
	css.WhiteSpace(kw.Nowrap),
	css.Overflow(kw.Hidden),
	css.TextOverflow(kw.Ellipsis),
	css.Transform(css.Raw("var(--tilt, rotate(4deg))")),
	css.Transition(css.Raw("transform 220ms cubic-bezier(0.34, 1.56, 0.64, 1)")),
)

var tilt = css.Custom("tilt", css.Raw("rotate(-3deg) translateY(-6px) scale(1.05)"))

var NewsCoverWideCard = NewsCoverWide.Descendant(NewsCoverCard).Set(css.FontSize(css.Raw("max(0.82rem, 1.7cqi)")), css.MaxWidth(css.Pct(48)))

var NewsMeta = style.Class("news-meta").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.Gap(css.Rem(0.35), css.Rem(0.9)),
	css.Margin(css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
	css.FontSize(css.Rem(0.95)),
	css.Color(css.Var("ink-2")),
)

var NewsFeature = style.Class("news-feature").
	Set(box...).
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1.15fr) minmax(0, 1fr)")),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(2.5)),
		css.Padding(css.Rem(1.1)),
		css.PaddingRight(css.Rem(2.5)),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		css.BoxShadow(css.Var("shadow-l")),
		css.Transition(css.Raw("transform 160ms ease, box-shadow 160ms ease")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(140))).
	Hover(css.Transform(css.Raw("translate(-3px, -3px)")), css.BoxShadow(css.Raw("9px 9px 0 var(--shadow-ink)")), tilt).
	Active(press...).
	FocusVisible(ring...).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.Padding(css.Rem(0.9)), css.Gap(css.Rem(1.25)))

var NewsStory = style.Class("news-story").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1)),
	css.MinWidth(css.Px(0)),
)

var NewsHeadline = style.Class("news-headline").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2), css.Vw(3.6), css.Rem(3.1))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.035)),
	css.LineHeight(css.Num(1)),
)

var NewsTeaser = style.Class("news-teaser").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(48)),
	css.FontSize(css.Rem(1.12)),
	css.LineHeight(css.Num(1.5)),
	css.Color(css.Var("ink")),
)

var NewsGo = style.Class("news-go").
	Set(button...).
	Set(css.Background(css.Var("sun")), css.Color(css.Var("on-fill")))

var NewsCatchUp = style.Class("news-catchup").
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1fr) minmax(0, 1.2fr)")),
		css.Gap(css.Rem(2.5)),
		css.Padding(css.Rem(2.25)),
		line,
		css.BorderRadius(css.Var("r-l")),
		css.Background(css.Var("lilac")),
		css.Color(css.Var("on-fill")),
		css.BoxShadow(css.Var("shadow-l")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(220))).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.Padding(css.Rem(1.25)), css.Gap(css.Rem(1.5)))

var NewsSectionTitle = style.Class("news-section-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(1.9), css.Vw(3.4), css.Rem(2.8))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.035)),
	css.LineHeight(css.Num(1)),
)

var NewsAsk = style.Class("news-ask").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(1)),
)

var NewsAskText = style.Class("news-ask-text").Set(
	css.Margin(css.Px(0)),
	css.FontSize(css.Rem(1.1)),
	css.LineHeight(css.Num(1.45)),
)

var NewsPicks = style.Class("news-picks").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.Gap(css.Rem(0.6)),
)

var NewsPick = style.Class("news-pick").
	Set(pill...).
	Set(
		css.MinHeight(css.Px(44)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Rem(1)),
		css.FontWeight(css.Num(700)),
		css.BoxShadow(css.Var("shadow-s")),
		css.Transition(css.Raw("transform 120ms ease, box-shadow 120ms ease")),
	).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...)

var NewsPickOn = NewsPick.And(style.Class("news-pick-on")).Set(
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.BorderColor(css.Var("line")),
)

var NewsAnswer = style.Class("news-answer").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.9)),
	css.Padding(css.Rem(1.4)),
	line,
	css.BorderRadius(css.Var("r")),
	css.Background(css.Var("panel")),
	css.Color(css.Var("ink")),
	css.MinWidth(css.Px(0)),
)

var NewsAnswerTitle = style.Class("news-answer-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.5)),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.02)),
	css.LineHeight(css.Num(1.1)),
)

var NewsMissed = style.Class("news-missed").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.5)),
	css.Margin(css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
)

var NewsMissedLink = NewsMissed.Descendant(style.Element("a")).
	Set(
		css.Color(css.Var("ink")),
		css.FontWeight(css.Num(700)),
		css.TextDecoration(kw.Underline),
		css.TextDecorationColor(css.Var("coral")),
		css.TextDecorationThickness(css.Px(3)),
		css.TextUnderlineOffset(css.Em(0.3)),
		css.TextDecorationSkipInk(kw.None),
	).
	FocusVisible(ring...)

var NewsCommand = style.Class("news-command").Set(
	css.Display(kw.Flex),
	css.AlignItems(kw.Center),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(0.6), css.Rem(0.6), css.Rem(0.6), css.Rem(1)),
	css.BorderRadius(css.Px(10)),
	css.Background(css.Var("code")),
	css.Color(css.Var("code-on")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.9)),
	css.MinWidth(css.Px(0)),
)

var NewsCommandText = style.Class("news-command-text").Set(
	css.OverflowWrap(kw.Anywhere),
	css.MinWidth(css.Px(0)),
)

var NewsEarlier = style.Class("news-earlier").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(1.5)),
)

var NewsRow = style.Class("news-row").
	Set(box...).
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 300px) minmax(0, 1fr)")),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(1.75)),
		css.Padding(css.Rem(0.9)),
		css.PaddingRight(css.Rem(2)),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		css.Transition(css.Raw("transform 160ms ease, box-shadow 160ms ease")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(300))).
	Hover(lift...).
	Hover(tilt).
	Active(press...).
	FocusVisible(ring...).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.Gap(css.Rem(1)), css.PaddingRight(css.Rem(0.9)))

var NewsRowTitle = style.Class("news-row-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(1.5), css.Vw(2.4), css.Rem(2))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.03)),
	css.LineHeight(css.Num(1.05)),
)

var NewsArticle = style.Class("news-article").
	Set(
		css.Width(css.Pct(100)),
		css.MaxWidth(css.Px(1180)),
		css.Margin(css.Px(0), kw.Auto),
		css.Padding(css.Rem(2.5), css.Rem(1.5), css.Rem(5)),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(2.25)),
	).
	Media(narrow, css.Padding(css.Rem(1.5), css.Rem(1), css.Rem(3)))

var NewsColumn = style.Class("news-column").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(1.75)),
	css.Width(css.Pct(100)),
	css.MaxWidth(css.Px(760)),
	css.MarginInline(kw.Auto),
)

var NewsBack = style.Class("news-back").
	Set(pill...).
	Set(css.AlignSelf(kw.FlexStart), css.BoxShadow(css.Var("shadow-s"))).
	Hover(lift...).
	FocusVisible(ring...)

var NewsTitle = style.Class("news-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2.6), css.Vw(5.6), css.Rem(4.4))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.045)),
	css.LineHeight(css.Num(0.98)),
)

var NewsShort = style.Class("news-short").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(1.4), css.Rem(1.6)),
	line,
	css.BorderRadius(css.Var("r-l")),
	css.Background(css.Var("lime")),
	css.Color(css.Var("on-fill")),
	css.BoxShadow(css.Var("shadow")),
)

var NewsShortTitle = style.Class("news-short-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.35)),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.02)),
)

var NewsShortList = NewsShort.Child(style.Element("ul")).Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.45)),
	css.Margin(css.Px(0)),
	css.PaddingLeft(css.Rem(1.2)),
	css.FontSize(css.Rem(1.05)),
	css.LineHeight(css.Num(1.45)),
)

var NewsBody = style.Class("news-body").Set(css.FontSize(css.Rem(1.08)))

var NewsSign = style.Class("news-sign").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.25)),
	css.FontWeight(css.Num(700)),
)

var NewsPager = style.Class("news-pager").
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("repeat(auto-fit, minmax(240px, 1fr))")),
		css.Gap(css.Rem(1)),
	)

var NewsPagerLink = style.Class("news-pager-link").
	Set(box...).
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(0.3)),
		css.Padding(css.Rem(1.1), css.Rem(1.3)),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		css.Transition(css.Raw("transform 160ms ease, box-shadow 160ms ease")),
	).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...)

var NewsPagerLabel = style.Class("news-pager-label").Set(
	css.FontSize(css.Rem(0.95)),
	css.Color(css.Var("ink-2")),
)

var NewsPagerTitle = style.Class("news-pager-title").Set(
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(1.3)),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.02)),
)

var NewsCompare = style.Class("news-compare").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.Stretch),
	css.Gap(css.Rem(0.75)),
	css.Width(css.Pct(100)),
)

var LatestNews = style.Class("latest-news").
	Set(box...).
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 340px) minmax(0, 1fr)")),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(2)),
		css.Padding(css.Rem(1)),
		css.PaddingRight(css.Rem(2.25)),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow-l")),
	).
	Hover(tilt).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.PaddingRight(css.Rem(1)), css.Gap(css.Rem(1.25)))

var LatestNewsCover = style.Class("latest-news-cover").
	Set(css.Display(kw.Block), css.BorderRadius(css.Var("r")), css.TextDecoration(kw.None)).
	FocusVisible(ring...)

var LatestNewsTitle = style.Class("latest-news-title").Set(
	css.Margin(css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(1.6), css.Vw(2.8), css.Rem(2.3))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.03)),
	css.LineHeight(css.Num(1.05)),
)

var LatestNewsTitleLink = LatestNewsTitle.Child(style.Element("a")).
	Set(css.Color(css.Var("ink")), css.TextDecoration(kw.None)).
	Hover(css.TextDecoration(kw.Underline), css.TextDecorationColor(css.Var("coral")), css.TextDecorationThickness(css.Px(4)), css.TextUnderlineOffset(css.Em(0.15)), css.TextDecorationSkipInk(kw.None)).
	FocusVisible(ring...)

var NewsMastheadIn = NewsHead.Child(NewsMasthead).Set(arrive...)

var NewsIntroIn = NewsHead.Child(NewsIntro).Set(arrive...).Set(css.AnimationDelay(css.Ms(70)))

var NewsMissedNote = style.Class("news-missed-note").Set(css.Color(css.Var("ink-2")))

var NewsCoverIn = NewsArticle.Child(NewsCoverWide).Set(arrive...)

var NewsTitleIn = NewsColumn.Child(NewsTitle).Set(arrive...).Set(css.AnimationDelay(css.Ms(80)))

var NewsLeadIn = NewsColumn.Child(NewsIntro).Set(arrive...).Set(css.AnimationDelay(css.Ms(140)))

var NewsShortIn = NewsColumn.Child(NewsShort).Set(arrive...).Set(css.AnimationDelay(css.Ms(200)))

var NewsBodyIn = NewsColumn.Child(NewsBody).Set(arrive...).Set(css.AnimationDelay(css.Ms(260)))

var LatestNewsIn = Wrap.Child(LatestNews).Set(arrive...).Set(css.AnimationDelay(css.Ms(120)))
