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
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.AlignItems(kw.FlexStart),
		css.Gap(css.Rem(1.25)),
		css.MinWidth(css.Px(0)),
	).
	Media(wide, css.GridColumn(css.Int(1)))

var TopicMain = style.Class("topic-main").
	Set(
		css.Display(kw.Flex),
		css.FlexDirection(kw.Column),
		css.Gap(css.Rem(2)),
		css.MinWidth(css.Px(0)),
	).
	Media(wide, css.GridColumn(css.Int(1)))

var TopicTitle = style.Class("topic-title").Set(
	css.Margin(css.Rem(0.5), css.Px(0), css.Px(0)),
	css.Padding(css.Em(0.1), css.Em(0.45)),
	css.Border(css.Var("border-l"), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Var("r-l")),
	css.BoxShadow(css.Var("shadow-l")),
	css.Background(css.Var("sun")),
	css.Color(css.Var("on-fill")),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Clamp(css.Rem(2), css.Vw(4.5), css.Rem(3.25))),
	css.FontWeight(css.Num(800)),
	css.LetterSpacing(css.Em(-0.035)),
	css.LineHeight(css.Num(1.05)),
	css.Transform(css.Raw("rotate(-1deg)")),
)

var TopicLead = style.Class("topic-lead").Set(
	css.Margin(css.Px(0)),
	css.MaxWidth(css.Ch(60)),
	css.Color(css.Var("ink-2")),
	css.FontSize(css.Clamp(css.Rem(1.05), css.Vw(1.5), css.Rem(1.25))),
	css.LineHeight(css.Num(1.5)),
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
		css.AlignItems(kw.Center),
		css.Gap(css.Rem(0.7)),
		css.Margin(css.Px(0)),
		css.FontFamily(css.Var("display")),
		css.FontSize(css.Clamp(css.Rem(1.8), css.Vw(3.2), css.Rem(2.5))),
		css.FontWeight(css.Num(800)),
		css.LetterSpacing(css.Em(-0.03)),
		css.LineHeight(css.Num(1.05)),
	).
	Before(
		css.Content(css.String("")),
		css.Flex(kw.None),
		css.Width(css.Em(0.5)),
		css.Height(css.Em(0.5)),
		css.Border(css.Var("border"), kw.Solid, css.Var("line")),
		css.BorderRadius(css.Pct(50)),
		css.Background(css.Var("coral")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Media(narrow, css.ScrollMarginTop(css.Rem(5.5)))

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
		css.Padding(css.Rem(0.9)),
		line,
		css.BorderRadius(css.Var("r-l")),
		css.Background(css.Var("panel")),
		css.BoxShadow(css.Var("shadow")),
	)

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
	Set(css.Color(css.Var("ink-2")), css.FontSize(css.Rem(0.7)), css.WhiteSpace(kw.Nowrap), css.MarginLeft(kw.Auto))

var RailPanel = style.Class("rail-panel").Set(
	css.Display(kw.Flex),
	css.FlexDirection(kw.Column),
	css.Gap(css.Rem(0.5)),
	css.Padding(css.Rem(0.75)),
	line,
	css.BorderRadius(css.Var("r-l")),
	css.Background(css.Var("panel")),
	css.BoxShadow(css.Var("shadow")),
)

var RailLabel = style.Class("rail-label").
	Set(label...).
	Set(css.Padding(css.Px(0), css.Rem(0.35)), css.Margin(css.Rem(0.35), css.Px(0), css.Px(0)), css.Color(css.Var("ink-2")), css.FontSize(css.Rem(0.72)))

var RailSearch = style.Class("rail-search").
	Set(
		css.Width(css.Pct(100)),
		css.MinHeight(css.Px(44)),
		css.Padding(css.Rem(0.4), css.Rem(0.9)),
		line,
		css.BorderRadius(css.Var("r")),
		css.Background(css.Var("ground")),
		css.Color(css.Var("ink")),
		css.FontFamily(css.Var("sans")),
		css.FontSize(css.Rem(0.95)),
	).
	FocusVisible(ring...)

var RailEmpty = style.Class("rail-empty").Set(
	css.Margin(css.Px(0)),
	css.Padding(css.Rem(0.35)),
	css.Color(css.Var("ink-2")),
	css.FontSize(css.Rem(0.9)),
)

var railLists = []css.Property{
	css.Display(kw.Flex),
	css.FlexWrap(kw.Wrap),
	css.Gap(css.Rem(0.4)),
	css.Margin(css.Px(0)),
	css.Padding(css.Px(0)),
	css.ListStyle(kw.None),
}

var RailList = Rail.Descendant(style.Element("ul")).
	Set(railLists...).
	Media(wide, css.FlexDirection(kw.Column), css.FlexWrap(kw.Nowrap), css.Gap(css.Rem(0.15)))

var RailLink = style.Class("rail-link").
	Set(pill...).
	Set(
		css.MinHeight(css.Px(44)),
		css.Padding(css.Rem(0.3), css.Rem(0.8)),
		css.Transition(css.Raw("transform 120ms ease, background 120ms ease")),
	).
	Hover(css.Background(css.Var("panel-2")), css.Transform(css.Raw("translateX(2px)"))).
	FocusVisible(ring...).
	Media(wide,
		css.Display(kw.Flex),
		css.MinHeight(css.Px(36)),
		css.Border(css.Var("border"), kw.Solid, kw.Transparent),
		css.BorderRadius(css.Var("r")),
		css.WhiteSpace(kw.Normal),
	)

var RailLinkSub = style.Class("rail-link-sub").
	Set(css.FontSize(css.Rem(0.82))).
	Media(wide, css.MinHeight(css.Px(32)), css.PaddingLeft(css.Rem(1.1)))

var RailWhere = style.Class("rail-where").
	Set(label...).
	Set(css.Display(kw.Block), css.Color(css.Var("ink-2")), css.FontSize(css.Rem(0.62)), css.LetterSpacing(css.Em(0.08)))

var RailLinkActive = style.Class("rail-link-active").
	Set(
		css.BorderColor(css.Var("line")),
		css.Background(css.Var("tab")),
		css.Color(css.Var("tab-on")),
		css.BoxShadow(css.Var("shadow-s")),
	).
	Hover(css.Background(css.Var("tab")), css.Transform(kw.None))

var Pager = style.Class("pager").
	Set(
		css.Display(kw.Flex),
		css.FlexWrap(kw.Wrap),
		css.JustifyContent(kw.SpaceBetween),
		css.Gap(css.Rem(1)),
		css.MinWidth(css.Px(0)),
		css.MarginTop(css.Rem(1)),
		css.PaddingTop(css.Rem(1.5)),
		css.BorderTop(css.Var("border"), kw.Solid, css.Var("line")),
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
