package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Pulse = style.Keyframes("pulse").
	From(css.Transform(css.Raw("rotate(0deg)"))).
	To(css.Transform(css.Raw("rotate(360deg)")))

var statusBox = []css.Property{
	css.JustifySelf(kw.Start),
	css.AlignSelf(kw.FlexStart),
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.9)),
	css.MinHeight(kw.Auto),
	css.Margin(css.Rem(2), css.Px(0)),
	css.Padding(css.Rem(0.8), css.Rem(1.2)),
	line,
	css.BorderRadius(css.Var("r")),
	css.BoxShadow(css.Var("shadow")),
	css.Background(css.Var("panel")),
	css.Color(css.Var("mis-ink")),
	css.TextShadow(css.Var("mis-shadow")),
	css.FontFamily(css.Var("display")),
	css.FontSize(css.Rem(2)),
	css.FontWeight(css.Num(400)),
	css.LineHeight(css.Num(1)),
	css.LetterSpacing(css.Em(0.04)),
	css.TextTransform(kw.Uppercase),
}

var dot = []css.Property{
	css.Content(css.String("")),
	css.Flex(kw.None),
	css.Width(css.Px(34)),
	css.Height(css.Px(34)),
	css.Border(css.Var("border"), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Pct(50)),
	css.Background(css.Raw("repeating-conic-gradient(var(--ink-pink) 0 30deg, var(--panel) 30deg 60deg)")),
}

var statusPulse = []css.Property{
	css.AnimationName(css.Ident(Pulse.Name())),
	css.AnimationDuration(css.S(0.9)),
	css.AnimationTimingFunction(kw.Linear),
	css.AnimationIterationCount(kw.Infinite),
}

var statusError = []css.Property{
	css.Background(css.Var("coral")),
	css.Color(css.Var("on-fill")),
	css.TextShadow(kw.None),
	css.FontFamily(css.Var("sans")),
	css.FontSize(css.Rem(1)),
	css.FontWeight(css.Num(600)),
	css.LineHeight(css.Num(1.4)),
	css.TextTransform(kw.None),
	css.LetterSpacing(css.Px(0)),
}

var statusErrorMark = []css.Property{
	css.Width(css.Px(22)),
	css.Height(css.Px(22)),
	css.Background(css.Var("panel")),
}

var shellStatus = []css.Property{
	css.MarginLeft(css.Rem(2)),
}

var shellStatusNarrow = []css.Property{
	css.Margin(css.Rem(1)),
}

var TopicLoading = style.Class("topic-loading").
	Set(statusBox...).
	Before(dot...).
	Before(statusPulse...)

var TopicError = style.Class("topic-error").
	Set(statusBox...).
	Set(statusError...).
	Before(dot...).
	Before(statusErrorMark...)

var ShellLoading = style.Class("shell-loading").
	Set(statusBox...).
	Set(shellStatus...).
	Before(dot...).
	Before(statusPulse...).
	Media(narrow, shellStatusNarrow...)

var ShellError = style.Class("shell-error").
	Set(statusBox...).
	Set(statusError...).
	Set(shellStatus...).
	Before(dot...).
	Before(statusErrorMark...).
	Media(narrow, shellStatusNarrow...)
