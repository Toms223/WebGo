package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var wide = media.MinWidth(css.Rem(78))

var Topic = style.Class("topic").
	Set(
		css.Display(kw.Grid),
		css.GridTemplateColumns(css.Raw("minmax(0, 1fr)")),
		css.AlignItems(kw.Start),
		css.ColumnGap(css.Rem(2.5)),
		css.RowGap(css.Rem(1.5)),
	).
	Media(wide, css.GridTemplateColumns(css.Raw("minmax(0, 1fr) 15rem")), css.RowGap(css.Rem(2)))

var TopicHeader = style.Class("topic-header").
	Set(
		css.Position(kw.Relative),
		css.Isolation(kw.Isolate),
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.FlexStart),
		css.Gap(css.Rem(1.4)),
		css.MinWidth(css.Px(0)),
		css.PaddingTop(css.Rem(0.75)),
	).
	Before(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Top(css.Rem(1.5)),
		css.Right(css.Pct(9)),
		css.Width(css.Pct(24)),
		css.Height(css.Clamp(css.Rem(6), css.Vw(12), css.Rem(10.5))),
		css.ZIndex(css.Int(-1)),
		css.Background(css.Var("ink-yellow")),
		css.Transform(css.Raw("rotate(3deg)")),
		css.MixBlendMode(css.Var("blend")),
	).
	After(
		css.Content(css.String("")),
		css.Position(kw.Absolute),
		css.Top(css.Px(0)),
		css.Right(css.Px(0)),
		css.Width(css.Pct(34)),
		css.Height(css.Clamp(css.Rem(6.5), css.Vw(13), css.Rem(11))),
		css.ZIndex(css.Int(-1)),
		css.BackgroundImage(css.Var("halftone")),
		css.BackgroundSize(css.Raw("10px 10px")),
		css.MixBlendMode(css.Var("blend")),
	).
	Media(wide, css.GridColumn(css.Int(1))).
	Nest(style.Self().Before().Media(narrow, css.Top(css.Rem(3.5)), css.Width(css.Pct(22))))

var TopicMain = style.Class("topic-main").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(2)),
		css.MinWidth(css.Px(0)),
	).
	Media(wide, css.GridColumn(css.Int(1)))

var TopicTitle = style.Class("topic-title").
	Set(misprint...).
	Set(
		css.Margin(css.Px(0)),
		css.MaxWidth(css.Ch(16)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Clamp(css.Rem(2.5), css.Vw(5.4), css.Rem(5))),
		css.FontWeight(css.Num(400)),
		css.LetterSpacing(css.Em(0.005)),
		css.LineHeight(css.Num(0.98)),
		css.TextTransform(kw.Uppercase),
		css.TextWrap(kw.Balance),
	)

var TopicLead = style.Class("topic-lead").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Em(36)),
	css.Padding(css.Rem(0.85), css.Rem(1.1)),
	css.Border(css.Px(2), kw.Solid, css.Var("line")),
	css.Background(css.Var("sun")),
	css.BoxShadow(css.Var("shadow-s")),
	css.Color(css.Var("on-fill")),
	css.FontSize(css.Clamp(css.Rem(1.1), css.Vw(1.5), css.Rem(1.3))),
	css.FontWeight(css.Num(600)),
	css.LineHeight(css.Num(1.45)),
	css.Transform(css.Raw("rotate(-0.5deg)")),
)

var Reveal = style.Keyframes("reveal").
	From(css.Opacity(css.Num(0)), css.Transform(css.Raw("translateY(10px)"))).
	To(css.Opacity(css.Num(1)), css.Transform(kw.None))

var RevealAgain = style.Keyframes("reveal-again").
	From(css.Opacity(css.Num(0)), css.Transform(css.Raw("translateY(10px)"))).
	To(css.Opacity(css.Num(1)), css.Transform(kw.None))

var TurnB = style.Class("turn-b").Set(css.AnimationName(css.Ident(RevealAgain.Name())))

var TopicSection = style.Class("topic-section").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.FlexStart),
		css.Gap(css.Rem(1.1)),
		css.MinWidth(css.Px(0)),
		css.AnimationName(css.Ident(Reveal.Name())),
		css.AnimationDuration(css.Ms(320)),
		css.AnimationTimingFunction(css.Raw("cubic-bezier(0.16, 1, 0.3, 1)")),
		css.AnimationFillMode(kw.Both),
	)

var SectionHeading = style.Class("section-heading").
	Set(
		css.ScrollMarginTop(css.Rem(1)),
		css.Display(kw.Flex),
		css.AlignItems(kw.Baseline),
		css.Gap(css.Rem(0.85)),
		css.Margin(css.Px(0)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Clamp(css.Rem(2), css.Vw(3.2), css.Rem(2.9))),
		css.FontWeight(css.Num(400)),
		css.LetterSpacing(css.Em(0.01)),
		css.LineHeight(css.Num(1)),
		css.TextTransform(kw.Uppercase),
	).
	Media(narrow, css.ScrollMarginTop(css.Rem(5.5)))

var SectionNumber = style.Class("section-number").Set(
	css.Flex(kw.None),
	css.Padding(css.Px(5), css.Px(7)),
	css.Background(css.Var("tab")),
	css.Color(css.Var("tab-on")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.85)),
	css.FontWeight(css.Num(700)),
	css.LineHeight(css.Num(1)),
	css.Transform(css.Raw("translateY(-0.35em) rotate(-3deg)")),
)

var TopicFolio = style.Class("topic-folio").Set(
	css.AlignSelf(kw.FlexEnd),
	css.Margin(css.Rem(-0.5), css.Px(0), css.Px(0)),
	css.Color(css.Var("coral")),
	css.WebkitTextStroke(css.Px(2), css.Var("line")),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(4.5), css.Vw(9), css.Rem(9))),
	css.LineHeight(css.Num(0.8)),
	css.MixBlendMode(css.Var("blend")),
	css.UserSelect(kw.None),
	css.PointerEvents(kw.None),
)

var stickyNote = []css.Property{
	css.Position(kw.Relative),
	css.Custom("focus", css.Var("on-fill")),
	css.Padding(css.Rem(1.1), css.Rem(1), css.Rem(1)),
	line,
	css.Background(css.Var("sun")),
	css.BoxShadow(css.Var("shadow")),
	css.Color(css.Var("on-fill")),
}

var railTape = []css.Property{
	css.Content(css.String("")),
	css.Position(kw.Absolute),
	css.Top(css.Px(0)),
	css.Left(css.Pct(30)),
	css.Width(css.Px(86)),
	css.Height(css.Px(20)),
	css.Background(css.Raw("color-mix(in srgb, var(--ink-pink) 55%, transparent)")),
	css.Transform(css.Raw("rotate(-4deg)")),
	css.PointerEvents(kw.None),
}

var Rail = style.Class("rail").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(0.4)),
		css.MinWidth(css.Px(0)),
	).
	Media(wide,
		css.Position(kw.Sticky),
		css.Top(css.Rem(1.5)),
		css.GridColumn(css.Int(2)),
		css.GridRow(css.Raw("1 / span 2")),
		css.MaxHeight(css.Raw("calc(100vh - 3rem)")),
		css.OverflowY(kw.Auto),
		css.Custom("focus", css.Var("on-fill")),
		css.Padding(css.Rem(1.4), css.Rem(1), css.Rem(1.1)),
		line,
		css.Background(css.Var("sun")),
		css.BoxShadow(css.Var("shadow")),
		css.Color(css.Var("on-fill")),
		css.Transform(css.Raw("rotate(1.2deg)")),
	).
	Nest(style.Self().Before().Media(wide, railTape...))

var RailOpen = style.Class("rail-open")

var RailSummary = style.Class("rail-summary").
	Set(pill...).
	Set(
		css.Display(kw.Flex),
		css.JustifyContent(kw.SpaceBetween),
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.75)),
		css.MinHeight(css.Px(44)),
		css.Width(css.Pct(100)),
		css.BoxShadow(css.Var("shadow-s")),
		css.FontWeight(css.Num(700)),
		css.TextAlign(kw.Left),
	).
	Hover(css.Background(css.Var("panel-2"))).
	FocusVisible(ring...).
	After(
		css.Content(css.String("")),
		css.Flex(kw.None),
		css.Width(css.Px(8)),
		css.Height(css.Px(8)),
		css.BorderRight(css.Px(2), kw.Solid, css.Var("ink")),
		css.BorderBottom(css.Px(2), kw.Solid, css.Var("ink")),
		css.Transform(css.Raw("translateY(-2px) rotate(45deg)")),
		css.Transition(css.Raw("transform 160ms ease")),
	)

var RailOpenChevron = RailOpen.Descendant(RailSummary).After(css.Transform(css.Raw("translateY(2px) rotate(-135deg)")))

var RailCount = style.Class("rail-count").
	Set(label...).
	Set(
		css.MarginLeft(kw.Auto),
		css.Padding(css.Px(2), css.Px(6)),
		css.Border(css.Px(2), kw.Solid, kw.Currentcolor),
		css.Color(css.Var("mis-ink")),
		css.FontSize(css.Rem(0.75)),
		css.WhiteSpace(kw.Nowrap),
		css.Transform(css.Raw("rotate(-4deg)")),
	)

var RailPanel = style.Class("rail-panel").
	Set(stickyNote...).
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(0.5)),
		css.MarginTop(css.Rem(0.5)),
		css.PaddingTop(css.Rem(1.4)),
	).
	Before(railTape...)

var RailLabel = style.Class("rail-label").Set(
	css.Margin(css.Rem(0.4), css.Px(0), css.Px(0)),
	css.Color(css.Var("ink-pink-deep")),
	css.FontFamily(css.Var("hand")),
	css.FontSize(css.Rem(1.5)),
	css.LineHeight(css.Num(1.1)),
	css.Transform(css.Raw("rotate(-1.5deg)")),
)

var RailSearch = style.Class("rail-search").
	Set(
		css.Width(css.Pct(100)),
		css.MinHeight(css.Px(44)),
		css.Padding(css.Rem(0.4), css.Rem(0.8)),
		css.Border(css.Px(2), kw.Solid, css.Var("on-fill")),
		css.BorderRadius(css.Var("r")),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Raw("3px 3px 0 var(--on-fill)")),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.95)),
	).
	FocusVisible(ring...)

var RailEmpty = style.Class("rail-empty").Set(
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(0.35)),
	css.FontSize(css.Rem(1)),
)

var RailList = Rail.Descendant(style.Element("ul")).Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Margin(css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
	css.CounterReset(css.Ident("rail")),
)

var RailLink = style.Class("rail-link").
	Set(
		css.Display(kw.Flex),
		css.AlignItems(kw.Baseline),
		css.Gap(css.Rem(0.5)),
		css.MinHeight(css.Px(44)),
		css.Padding(css.Rem(0.3), css.Rem(0.45)),
		css.Border(css.Px(2), kw.Solid, kw.Transparent),
		css.Color(kw.Inherit),
		css.FontSize(css.Rem(1)),
		css.LineHeight(css.Num(1.3)),
		css.TextDecoration(kw.None),
		css.CounterIncrement(css.Ident("rail")),
	).
	Before(
		css.Content(css.Raw("counter(rail, decimal-leading-zero)")),
		css.Flex(kw.None),
		css.FontFamily(css.Var("mono")),
		css.FontSize(css.Rem(0.8)),
		css.FontWeight(css.Num(700)),
	).
	Hover(
		css.TextDecorationLine(kw.Underline),
		css.TextDecorationStyle(kw.Wavy),
		css.TextDecorationColor(css.Var("ink-pink-deep")),
		css.TextUnderlineOffset(css.Px(4)),
	).
	FocusVisible(ring...).
	Media(wide, css.MinHeight(css.Px(30)), css.PaddingTop(css.Rem(0.1)), css.PaddingBottom(css.Rem(0.1)))

var RailLinkSub = style.Class("rail-link-sub").
	Set(css.FontSize(css.Rem(0.95)), css.PaddingLeft(css.Rem(1))).
	Before(css.Content(css.String("↳")), css.FontWeight(css.Num(400))).
	Media(wide, css.MinHeight(css.Px(28)))

var RailWhere = style.Class("rail-where").
	Set(label...).
	Set(css.Display(kw.Block), css.FontSize(css.Rem(0.72)), css.LetterSpacing(css.Em(0.06)))

var RailLinkActive = style.Class("rail-link-active").
	Set(
		css.BorderColor(css.Var("on-fill")),
		css.Background(css.Var("panel")),
		css.Color(css.Var("ink")),
		css.BoxShadow(css.Raw("3px 3px 0 var(--on-fill)")),
		css.FontWeight(css.Num(700)),
		css.Transform(css.Raw("rotate(-1deg)")),
	).
	Before(css.Color(css.Var("ink-pink-deep"))).
	Hover(css.TextDecorationLine(kw.None))

var RailStamp = style.Class("rail-stamp").
	Set(label...).
	Set(
		css.Display(kw.InlineGrid),
		css.JustifyItems(kw.Center),
		css.AlignSelf(kw.FlexStart),
		css.Gap(css.Px(2)),
		css.Margin(css.Rem(1.4), css.Px(0), css.Rem(0.6), css.Rem(0.75)),
		css.Padding(css.Rem(0.5), css.Rem(0.9), css.Rem(0.4)),
		css.Border(css.Px(4), kw.Solid, kw.Currentcolor),
		css.Outline(css.Px(2), kw.Solid, kw.Currentcolor),
		css.OutlineOffset(css.Px(3)),
		css.Color(css.Var("ink-blue")),
		css.FontSize(css.Rem(0.75)),
		css.LineHeight(css.Num(1.2)),
		css.Transform(css.Raw("rotate(-7deg)")),
	)

var RailStampNumber = style.Class("rail-stamp-number").Set(
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(2.6)),
	css.FontWeight(css.Num(400)),
	css.LetterSpacing(css.Em(0.02)),
	css.LineHeight(css.Num(0.95)),
)

var Pager = style.Class("pager").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.JustifyContent(kw.SpaceBetween),
		css.Gap(css.Rem(1)),
		css.MinWidth(css.Px(0)),
		css.MarginTop(css.Rem(1)),
		css.PaddingTop(css.Rem(1.5)),
		css.BorderTop(css.Px(2), kw.Dashed, css.Var("line")),
	)

var PagerLink = style.Class("pager-link").
	Set(button...).
	Set(
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.FlexStart),
		css.Gap(css.Rem(0.2)),
		css.MinWidth(css.Rem(10)),
		css.MaxWidth(css.Pct(100)),
		css.Padding(css.Rem(0.7), css.Rem(1.1)),
	).
	Hover(lift...).
	Active(press...).
	FocusVisible(ring...)

var PagerNext = style.Class("pager-next").Set(
	css.MarginLeft(kw.Auto),
	css.AlignItems(kw.FlexEnd),
	css.TextAlign(kw.Right),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
)

var PagerLabel = style.Class("pager-label").
	Set(label...).
	Set(css.FontSize(css.Rem(0.72)))
