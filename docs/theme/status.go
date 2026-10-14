package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/style"
)

var Pulse = style.Keyframes("pulse").
	At(css.Pct(0), css.Opacity(css.Num(1))).
	At(css.Pct(50), css.Opacity(css.Num(0.15))).
	At(css.Pct(100), css.Opacity(css.Num(1)))

var statusBox = []css.Property{
	css.JustifySelf(kw.Start),
	css.AlignSelf(kw.FlexStart),
	css.Display(kw.InlineFlex),
	css.AlignItems(kw.Center),
	css.Gap(css.Rem(0.7)),
	css.MinHeight(kw.Auto),
	css.Margin(css.Rem(2), css.Px(0)),
	css.Padding(css.Rem(0.65), css.Rem(1)),
	line,
	css.BorderRadius(css.Var("r")),
	css.BoxShadow(css.Var("shadow-s")),
	css.Background(css.Var("panel")),
	css.Color(css.Var("ink")),
	css.FontFamily(css.Var("mono")),
	css.FontSize(css.Rem(0.8)),
	css.FontWeight(css.Num(700)),
	css.LetterSpacing(css.Em(0.08)),
	css.TextTransform(kw.Uppercase),
}

var dot = []css.Property{
	css.Content(css.String("")),
	css.Width(css.Px(10)),
	css.Height(css.Px(10)),
	css.Border(css.Px(1.5), kw.Solid, css.Var("line")),
	css.BorderRadius(css.Pct(50)),
	css.Background(css.Var("coral")),
}

var statusPulse = []css.Property{
	css.AnimationName(css.Ident(Pulse.Name())),
	css.AnimationDuration(css.S(1.1)),
	css.AnimationTimingFunction(css.Raw("steps(2, jump-none)")),
	css.AnimationIterationCount(kw.Infinite),
}

var statusError = []css.Property{
	css.Background(css.Var("coral")),
	css.Color(css.Var("on-fill")),
	css.TextTransform(kw.None),
	css.LetterSpacing(css.Px(0)),
}

var statusErrorMark = []css.Property{
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
