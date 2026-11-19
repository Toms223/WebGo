package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var arrive = []css.Property{
	css.AnimationName(css.Ident(Reveal.Name())),
	css.AnimationDuration(css.Ms(420)),
	css.AnimationTimingFunction(css.Raw("cubic-bezier(0.16, 1, 0.3, 1)")),
	css.AnimationFillMode(kw.Both),
}

var spreadNarrow = media.MaxWidth(css.Rem(64))

var newsMis = []css.Property{
	css.FontFamily(css.Var("display")),
	css.FontWeight(css.Num(400)),
	css.TextTransform(kw.Uppercase),
	css.Color(css.Var("mis-ink")),
	css.TextShadow(css.Var("mis-shadow")),
}

var newsSlab = []css.Property{
	css.FontFamily(css.Var("display")),
	css.FontWeight(css.Num(400)),
	css.TextTransform(kw.Uppercase),
	css.LetterSpacing(css.Em(0.01)),
}

var newsStack = css.BoxShadow(css.Raw("6px 6px 0 -2px var(--panel), 6px 6px 0 1px var(--line), 12px 12px 0 -2px var(--panel), 12px 12px 0 1px var(--line), 18px 18px 0 var(--shadow-ink)"))

var newsStackUp = css.BoxShadow(css.Raw("8px 8px 0 -2px var(--panel), 8px 8px 0 1px var(--line), 16px 16px 0 -2px var(--panel), 16px 16px 0 1px var(--line), 23px 23px 0 var(--shadow-ink)"))

var newsMetal = css.Background(css.Raw("linear-gradient(#C9CCD3, #8A8E98 55%, #E3E5EA)"))

var newsStapleBar = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Absolute),
	css.Top(css.Px(-5)),
	css.Left(css.Raw("calc(50% - 23px)")),
	css.Width(css.Px(46)),
	css.Height(css.Px(7)),
	newsMetal,
	css.BoxShadow(css.Raw("0 1px 0 rgba(0, 0, 0, 0.35), inset 0 -1px 0 rgba(0, 0, 0, 0.25)")),
	css.ZIndex(css.Int(3)),
}

var newsStampMask = css.Raw(`url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='160' height='160'%3E%3Cfilter id='m'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.9' numOctaves='2'/%3E%3CfeColorMatrix values='0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 -3 2.3'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23m)'/%3E%3C/svg%3E")`)

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
	css.PaddingBottom(css.Rem(1.5)),
	css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
)

var NewsMasthead = style.Class("news-masthead").
	Set(newsMis...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Clamp(css.Rem(5), css.Vw(15), css.Rem(12))),
		css.LetterSpacing(css.Em(0.01)),
		css.LineHeight(css.Num(0.82)),
	)

var NewsIntro = style.Class("news-intro").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(36)),
	css.FontSize(css.Clamp(css.Rem(1.1), css.Vw(1.7), css.Rem(1.4))),
	css.LineHeight(css.Num(1.4)),
	css.Color(css.Var("ink")),
)

var NewsCover = style.Class("news-cover").Set(
	css.ContainerType(kw.InlineSize),
	css.Position(kw.Relative),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.JustifyContent(kw.FlexEnd),
	css.AspectRatio(css.Raw("4 / 3")),
	css.Padding(css.Raw("6cqi")),
	line,
	css.Background(css.Var("panel")),
	css.Color(css.Var("ink")),
	css.Overflow(kw.Hidden),
	css.MinWidth(css.Px(0)),
)

var NewsInk = style.Class("news-ink").Set(
	css.Position(kw.Absolute),
	css.PointerEvents(kw.None),
	css.MixBlendMode(css.Var("blend")),
)

var NewsBlock = style.Class("news-block").Set(
	css.Left(css.Pct(-8)),
	css.Top(css.Pct(-8)),
	css.Width(css.Pct(62)),
	css.Height(css.Pct(58)),
	css.Transform(css.Raw("rotate(-3deg)")),
)

var NewsDots = style.Class("news-dots").Set(
	css.Right(css.Pct(-6)),
	css.Top(css.Pct(-8)),
	css.Width(css.Pct(48)),
	css.Height(css.Pct(94)),
	css.BackgroundImage(css.Var("halftone")),
	css.BackgroundSize(css.Raw("11px 11px")),
)

var NewsDisc = style.Class("news-disc").Set(
	css.Right(css.Pct(-6)),
	css.Bottom(css.Pct(-22)),
	css.Width(css.Pct(40)),
	css.AspectRatio(css.Raw("1")),
	css.BorderRadius(css.Pct(50)),
	css.Background(css.Var("ink-blue")),
	css.Opacity(css.Num(0.85)),
)

var NewsCoverText = style.Class("news-cover-text").Set(
	css.Position(kw.Relative),
	css.ZIndex(css.Int(1)),
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.AlignItems(kw.FlexStart),
	css.Gap(css.Rem(0.6)),
	css.MinWidth(css.Px(0)),
)

var NewsCoverLabel = style.Class("news-cover-label").Set(
	css.Margin(css.Px(0)),
	css.Padding(css.Em(0.35), css.Em(0.6)),
	css.Background(css.Var("ink")),
	css.Color(css.Var("ground")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Raw("clamp(0.82rem, 2.4cqi, 0.95rem)")),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.14)),
	css.LineHeight(css.Num(1.25)),
	css.TextTransform(kw.Uppercase),
)

var NewsCoverName = style.Class("news-cover-name").
	Set(newsMis...).
	Set(
		css.Margin(css.Px(0)),
		css.MaxWidth(css.Pct(70)),
		css.FontSize(css.Raw("15cqi")),
		css.LetterSpacing(css.Em(0.01)),
		css.LineHeight(css.Num(0.88)),
	)

var NewsCoverCard = style.Class("news-cover-card").Set(
	css.Position(kw.Absolute),
	css.Left(css.Raw("6cqi")),
	css.Top(css.Raw("7cqi")),
	css.ZIndex(css.Int(1)),
	css.MaxWidth(css.Pct(56)),
	css.Padding(css.Raw("1.4cqi 2.2cqi")),
	css.Border(css.Px(2), kw.Dashed, css.Var("line")),
	css.Background(css.Var("code")),
	css.Color(css.Var("code-on")),
	css.BoxShadow(css.Var("shadow-s")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Raw("max(0.82rem, 2.6cqi)")),
	css.WhiteSpace(kw.Nowrap),
	css.Overflow(kw.Hidden),
	css.TextOverflow(kw.Ellipsis),
	css.Transform(css.Raw("var(--tilt, rotate(-2deg))")),
	css.Transition(css.Raw("transform 220ms cubic-bezier(0.34, 1.56, 0.64, 1)")),
)

var tilt = css.Custom("tilt", css.Raw("rotate(2deg) translateY(-4px) scale(1.04)"))

var NewsStamp = style.Class("news-stamp").Set(
	css.Display(kw.InlineGrid),
	css.Gap(css.Px(2)),
	css.TextAlign(kw.Center),
	css.Color(css.Var("code-str")),
	css.Border(css.Px(4), kw.Solid, kw.Currentcolor),
	css.Outline(css.Px(2), kw.Solid, kw.Currentcolor),
	css.OutlineOffset(css.Px(4)),
	css.Padding(css.Em(0.7), css.Em(1.1), css.Em(0.55)),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.9)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.12)),
	css.LineHeight(css.Num(1.2)),
	css.TextTransform(kw.Uppercase),
	css.WhiteSpace(kw.Nowrap),
	css.Transform(css.Raw("rotate(-7deg)")),
	css.MixBlendMode(css.Var("blend")),
	css.MaskImage(newsStampMask),
	css.MaskSize(css.Px(160)),
)

var NewsStampNumber = NewsStamp.Child(style.Element("strong")).Set(
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Em(3.1)),
	css.FontWeight(css.Num(400)),
	css.LetterSpacing(css.Em(0.02)),
	css.LineHeight(css.Num(0.95)),
)

var NewsCoverStamp = NewsCover.Child(NewsStamp).Set(
	css.Position(kw.Absolute),
	css.Right(css.Raw("7cqi")),
	css.Top(css.Raw("10cqi")),
	css.ZIndex(css.Int(1)),
	css.FontSize(css.Raw("max(0.82rem, 2.5cqi)")),
)

var NewsMeta = style.Class("news-meta").Set(
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.Gap(css.Rem(0.35), css.Rem(0.9)),
	css.Margin(css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.9)),
	css.Color(css.Var("ink-2")),
)

var NewsFeature = style.Class("news-feature").
	Set(box...).
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1.15fr) minmax(0, 1fr)")),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(2.5)),
		css.Padding(css.Rem(1.1)),
		css.PaddingRight(css.Rem(2.5)),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		newsStack,
		css.Transition(css.Raw("transform 160ms ease, box-shadow 160ms ease")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(140))).
	Before(newsStapleBar...).
	Hover(css.Transform(css.Raw("translate(-3px, -3px)")), newsStackUp, tilt).
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

var NewsHeadline = style.Class("news-headline").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Clamp(css.Rem(2), css.Vw(3.6), css.Rem(3.3))),
		css.LineHeight(css.Num(1)),
	)

var NewsTeaser = style.Class("news-teaser").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(48)),
	css.FontSize(css.Rem(1.08)),
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
		css.BackgroundImage(css.Raw("radial-gradient(circle, color-mix(in srgb, var(--ink-blue) 22%, transparent) 34%, transparent 37%)")),
		css.BackgroundSize(css.Raw("14px 14px")),
		css.Color(css.Var("on-fill")),
		css.BoxShadow(css.Var("shadow-l")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(220))).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.Padding(css.Rem(1.25)), css.Gap(css.Rem(1.5)))

var NewsSectionTitle = style.Class("news-section-title").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.Padding(css.Em(0.12), css.Em(0.35), css.Em(0.08)),
		line,
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.BoxShadow(css.Var("shadow")),
		css.FontSize(css.Clamp(css.Rem(1.9), css.Vw(3.4), css.Rem(2.8))),
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
		css.FontFamily(css.Var("mono")),
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
	css.BoxShadow(css.Var("shadow")),
	css.MinWidth(css.Px(0)),
)

var NewsAnswerTitle = style.Class("news-answer-title").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Rem(1.6)),
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
	css.Border(css.Px(2), kw.Dashed, css.Var("line")),
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

var NewsEarlier = style.Class("news-earlier").
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("repeat(auto-fill, minmax(min(100%, 300px), 1fr))")),
		css.AlignItems(kw.Start),
		css.Gap(css.Rem(3), css.Rem(2.5)),
	)

var NewsEarlierTitle = NewsEarlier.Child(NewsSectionTitle).Set(
	css.GridColumn(css.Raw("1 / -1")),
	css.JustifySelf(kw.Start),
)

var NewsRow = style.Class("news-row").
	Set(box...).
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(1.1)),
		css.Padding(css.Rem(0.8), css.Rem(0.8), css.Rem(1.3)),
		css.Color(css.Var("ink")),
		css.TextDecoration(kw.None),
		newsStack,
		css.Transform(css.Raw("var(--lean, rotate(-1deg))")),
		css.Transition(css.Raw("transform 160ms ease, box-shadow 160ms ease")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(300))).
	Before(newsStapleBar...).
	NthOfType(css.An(2, 0), css.Custom("lean", css.Raw("rotate(1.2deg)"))).
	Hover(css.Transform(css.Raw("translate(-3px, -3px) rotate(0deg)")), newsStackUp, tilt).
	Active(press...).
	FocusVisible(ring...)

var NewsRowTitle = style.Class("news-row-title").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Clamp(css.Rem(1.6), css.Vw(2.4), css.Rem(2.1))),
		css.LineHeight(css.Num(1.02)),
	)

var NewsArticle = style.Class("news-article").
	Set(
		css.Width(css.Pct(100)),
		css.MaxWidth(css.Px(1180)),
		css.Margin(css.Px(0), kw.Auto),
		css.Padding(css.Rem(2.5), css.Rem(1.5), css.Rem(5)),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(2.5)),
	).
	Media(narrow, css.Padding(css.Rem(1.5), css.Rem(1), css.Rem(3)), css.Gap(css.Rem(1.75)))

var NewsBack = style.Class("news-back").
	Set(pill...).
	Set(
		css.AlignSelf(kw.FlexStart),
		css.FontFamily(css.Var("mono")),
		css.FontWeight(css.Num(700)),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Hover(css.Transform(css.Raw("translate(-2px, -2px) rotate(-1deg)")), css.BoxShadow(css.Var("shadow"))).
	FocusVisible(ring...)

var NewsIssue = style.Class("news-issue").
	Set(
		css.ContainerType(kw.InlineSize),
		css.Position(kw.Relative),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1fr) auto")),
		css.AlignItems(kw.End),
		css.Gap(css.Rem(2.5)),
		css.Padding(css.Rem(3), css.Rem(3), css.Rem(2.5)),
		line,
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.BoxShadow(css.Var("shadow-l")),
		css.Overflow(kw.Hidden),
	).
	Set(arrive...).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.Padding(css.Rem(1.75), css.Rem(1.25)), css.Gap(css.Rem(1.75)))

var NewsIssueDisc = NewsIssue.Child(NewsDisc).Set(
	css.Width(css.Px(260)),
	css.Right(css.Pct(6)),
	css.Bottom(css.Px(-90)),
)

var NewsIssueBlock = NewsIssue.Child(NewsBlock).Set(
	css.Top(css.Pct(80)),
	css.Left(css.Pct(-4)),
	css.Width(css.Pct(58)),
	css.Height(css.Pct(70)),
)

var NewsIssueStamp = NewsIssue.Child(NewsStamp).
	Set(css.Position(kw.Relative), css.ZIndex(css.Int(1)), css.MarginBottom(css.Rem(1)), css.MarginRight(css.Rem(1))).
	Media(narrow, css.JustifySelf(kw.Start), css.MarginLeft(css.Rem(0.75)))

var NewsIssueMeta = NewsIssue.Descendant(NewsMeta).Set(
	css.Padding(css.Rem(0.35), css.Rem(0.7)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("panel")),
	css.Color(css.Var("ink")),
)

var NewsTitle = style.Class("news-title").
	Set(newsMis...).
	Set(
		css.Margin(css.Px(0), css.Px(0), css.Rem(0.6)),
		css.MaxWidth(css.Ch(13)),
		css.FontSize(css.Clamp(css.Rem(2.6), css.Vw(7.6), css.Rem(7.5))),
		css.LetterSpacing(css.Em(0.005)),
		css.LineHeight(css.Num(0.92)),
	).
	Media(narrow, css.MaxWidth(kw.None))

var NewsSpread = style.Class("news-spread").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1fr) minmax(0, 1fr)")),
		line,
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow-l")),
	).
	Set(arrive...).
	Set(css.AnimationDelay(css.Ms(120))).
	Before(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Top(css.Px(0)),
		css.Bottom(css.Px(0)),
		css.Left(css.Raw("calc(50% - 20px)")),
		css.Width(css.Px(40)),
		css.Background(css.Raw("linear-gradient(90deg, transparent, color-mix(in srgb, var(--ink) 8%, transparent) 45%, color-mix(in srgb, var(--ink) 16%, transparent) 50%, color-mix(in srgb, var(--ink) 8%, transparent) 55%, transparent)")),
		css.PointerEvents(kw.None),
	).
	Media(spreadNarrow, css.GridTemplateColumns(css.Raw("minmax(0, 1fr)"))).
	Nest(style.Self().Before().Media(spreadNarrow, css.Display(kw.None)))

var NewsStaple = style.Class("news-staple").
	Set(
		css.Position(kw.Absolute),
		css.Left(css.Raw("calc(50% - 4px)")),
		css.Top(css.Pct(22)),
		css.Width(css.Px(8)),
		css.Height(css.Px(46)),
		css.Background(css.Raw("linear-gradient(90deg, #C9CCD3, #8A8E98 55%, #E3E5EA)")),
		css.BoxShadow(css.Raw("1px 0 0 rgba(0, 0, 0, 0.35)")),
		css.ZIndex(css.Int(3)),
	).
	Media(spreadNarrow, css.Display(kw.None))

var NewsStapleLow = NewsStaple.And(style.Class("news-staple-low")).Set(css.Top(css.Pct(72)))

var NewsPage = style.Class("news-page").
	Set(
		css.Position(kw.Relative),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(1.75)),
		css.Padding(css.Rem(2.5), css.Rem(2.75), css.Rem(7)),
		css.MinWidth(css.Px(0)),
	).
	Media(narrow, css.Padding(css.Rem(1.75), css.Rem(1.25), css.Rem(6)))

var NewsPageRight = NewsPage.Next(NewsPage).
	Set(css.BorderLeft(css.Px(2), kw.Dashed, css.Raw("color-mix(in srgb, var(--ink) 35%, transparent)"))).
	Media(spreadNarrow, css.BorderLeft(kw.None), css.BorderTop(css.Px(2), kw.Dashed, css.Raw("color-mix(in srgb, var(--ink) 35%, transparent)")))

var NewsRunhead = style.Class("news-runhead").Set(
	css.Display(kw.Flex),
	css.JustifyContent(kw.SpaceBetween),
	css.Gap(css.Rem(1)),
	css.PaddingBottom(css.Rem(0.6)),
	css.BorderBottom(css.Px(2), kw.Solid, css.Var("line")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.85)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.12)),
	css.LineHeight(css.Num(1.2)),
	css.TextTransform(kw.Uppercase),
	css.Color(css.Var("ink-2")),
)

var NewsShort = style.Class("news-short").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.75)),
	css.Padding(css.Rem(1.4), css.Rem(1.5), css.Rem(1.1)),
	line,
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.BoxShadow(css.Var("shadow")),
	css.Transform(css.Raw("rotate(-1deg)")),
)

var NewsShortTitle = style.Class("news-short-title").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Rem(1.9)),
		css.LineHeight(css.Num(1)),
	)

var NewsShortList = NewsShort.Child(style.Element("ul")).Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.45)),
	css.Margin(css.Px(0)),
	css.PaddingLeft(css.Rem(1.4)),
	css.FontSize(css.Rem(1.02)),
	css.LineHeight(css.Num(1.45)),
)

var NewsShortMark = NewsShortList.Child(style.Element("li")).Marker(
	css.Content(css.String("■  ")),
	css.Color(css.Var("ink-pink-deep")),
)

var NewsText = Doc.And(style.Class("news-text")).Set(
	css.Display(kw.Block),
	css.FontSize(css.Rem(1.06)),
)

var NewsTextFlow = NewsText.Child(style.Any()).Set(css.Margin(css.Px(0), css.Px(0), css.Rem(1.1)))

var NewsTextHeading = NewsText.Child(DocHeading).Set(css.MarginTop(css.Rem(2.2)))

var NewsDrop = style.Class("news-drop").FirstLetter(
	css.Float(kw.Left),
	css.Margin(css.Rem(0.35), css.Rem(0.7), css.Px(0), css.Px(0)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(5.8)),
	css.FontWeight(css.Num(400)),
	css.LineHeight(css.Num(0.8)),
	css.Color(css.Var("mis-ink")),
	css.TextShadow(css.Raw("4px 3px 0 var(--mis)")),
)

var NewsFolio = style.Class("news-folio").Set(
	css.Position(kw.Absolute),
	css.Bottom(css.Rem(0.75)),
	css.Left(css.Rem(1.25)),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(3.5), css.Vw(5), css.Rem(4.75))),
	css.FontWeight(css.Num(400)),
	css.LineHeight(css.Num(1)),
	css.Color(css.Var("ink-pink")),
	css.MixBlendMode(css.Var("blend")),
	css.PointerEvents(kw.None),
	css.UserSelect(kw.None),
)

var NewsFolioRight = NewsFolio.And(style.Class("news-folio-right")).Set(
	css.Left(kw.Auto),
	css.Right(css.Rem(1.25)),
	css.Color(css.Var("mis-ink")),
)

var NewsPull = style.Class("news-pull").
	Set(newsSlab...).
	Set(
		css.Position(kw.Relative),
		css.Margin(css.Rem(1), css.Px(0), css.Px(0)),
		css.Padding(css.Rem(1.1), css.Rem(3.5), css.Rem(1.1), css.Px(0)),
		css.BorderTop(css.Var("border"), kw.Solid, css.Var("line")),
		css.BorderBottom(css.Var("border"), kw.Solid, css.Var("line")),
		css.Color(css.Var("mis-ink")),
		css.FontSize(css.Clamp(css.Rem(1.25), css.Vw(1.7), css.Rem(1.45))),
		css.LineHeight(css.Num(1.12)),
	).
	Before(
		css.Content(css.String("“")),
		css.Position(kw.Absolute),
		css.Right(css.Px(0)),
		css.Top(css.Em(-0.3)),
		css.FontSize(css.Em(3.6)),
		css.LineHeight(css.Num(1)),
		css.Color(css.Var("ink-pink")),
		css.MixBlendMode(css.Var("blend")),
	)

var NewsSign = style.Class("news-sign").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Rem(1.4)),
		css.LineHeight(css.Num(1.15)),
	)

var NewsEndmark = style.Class("news-endmark").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(1)),
		css.Margin(css.Px(0)),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.9)),
		css.FontWeight(css.Num(700)),
		css.LetterSpacing(css.Em(0.1)),
		css.TextTransform(kw.Uppercase),
	).
	Before(
		css.Content(css.String("")),
		css.Width(css.Px(18)),
		css.Height(css.Px(18)),
		css.Background(css.Var("ink-blue")),
		css.BoxShadow(css.Raw("4px 4px 0 var(--ink-pink)")),
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
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.9)),
	css.Color(css.Var("ink-2")),
)

var NewsPagerTitle = style.Class("news-pager-title").
	Set(newsSlab...).
	Set(css.FontSize(css.Rem(1.45)), css.LineHeight(css.Num(1.1)))

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
		css.Position(kw.Relative),
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 340px) minmax(0, 1fr)")),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(2)),
		css.Padding(css.Rem(1)),
		css.PaddingRight(css.Rem(2.25)),
		css.Background(css.Var("panel")),
		newsStack,
	).
	Before(newsStapleBar...).
	Hover(tilt).
	Media(narrow, css.GridTemplateColumns(css.Fr(1)), css.PaddingRight(css.Rem(1)), css.Gap(css.Rem(1.25)))

var LatestNewsCover = style.Class("latest-news-cover").
	Set(css.Display(kw.Block), css.BorderRadius(css.Var("r")), css.TextDecoration(kw.None)).
	FocusVisible(ring...)

var LatestNewsTitle = style.Class("latest-news-title").
	Set(newsSlab...).
	Set(
		css.Margin(css.Px(0)),
		css.FontSize(css.Clamp(css.Rem(1.8), css.Vw(3), css.Rem(2.6))),
		css.LineHeight(css.Num(1)),
	)

var LatestNewsTitleLink = LatestNewsTitle.Child(style.Element("a")).
	Set(css.Color(css.Var("ink")), css.TextDecoration(kw.None)).
	Hover(css.TextDecoration(kw.Underline), css.TextDecorationColor(css.Var("coral")), css.TextDecorationThickness(css.Px(4)), css.TextUnderlineOffset(css.Em(0.15)), css.TextDecorationSkipInk(kw.None)).
	FocusVisible(ring...)

var NewsMastheadIn = NewsHead.Child(NewsMasthead).Set(arrive...)

var NewsIntroIn = NewsHead.Child(NewsIntro).Set(arrive...).Set(css.AnimationDelay(css.Ms(70)))

var NewsMissedNote = style.Class("news-missed-note").Set(css.Color(css.Var("ink-2")))

var LatestNewsIn = Wrap.Child(LatestNews).Set(arrive...).Set(css.AnimationDelay(css.Ms(120)))

var NewsPullCode = NewsPull.Descendant(style.Element("code")).Set(
	css.Padding(css.Px(0)),
	css.Border(kw.None),
	css.Background(kw.None),
	css.BoxShadow(kw.None),
	css.Color(kw.Inherit),
	css.FontFamily(kw.Inherit),
	css.FontSize(kw.Inherit),
)
