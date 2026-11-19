package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var fills = []css.Property{
	css.Custom("sun", css.Hex("#FFE600")),
	css.Custom("lime", css.Hex("#FF9A3D")),
	css.Custom("lilac", css.Hex("#A9BCF0")),
	css.Custom("coral", css.Hex("#FF48B0")),
	css.Custom("on-fill", css.Hex("#1B1A24")),
	css.Custom("ink-pink", css.Hex("#FF48B0")),
	css.Custom("ink-pink-deep", css.Hex("#C8127C")),
	css.Custom("ink-blue", css.Hex("#2C4FA8")),
	css.Custom("ink-yellow", css.Hex("#FFE600")),
	css.Custom("on-blue", css.Hex("#FFFDF7")),
}

var light = []css.Property{
	css.Custom("ground", css.Hex("#F7F2E6")),
	css.Custom("panel", css.Hex("#FFFDF7")),
	css.Custom("panel-2", css.Hex("#EFE7D4")),
	css.Custom("ink", css.Hex("#1B1A24")),
	css.Custom("ink-2", css.Hex("#45424F")),
	css.Custom("line", css.Hex("#1B1A24")),
	css.Custom("shadow-ink", css.Hex("#1B1A24")),
	css.Custom("focus", css.Hex("#2C4FA8")),
	css.Custom("code", css.Hex("#FFFDF7")),
	css.Custom("code-on", css.Hex("#1B1A24")),
	css.Custom("code-dim", css.Hex("#6E6A78")),
	css.Custom("code-inline", css.Hex("#FFF27A")),
	css.Custom("code-key", css.Hex("#2C4FA8")),
	css.Custom("code-str", css.Hex("#C8127C")),
	css.Custom("code-fn", css.Hex("#1B1A24")),
	css.Custom("code-type", css.Hex("#3A2370")),
	css.Custom("code-mark", css.RGBA(255, 230, 0, 0.75)),
	css.Custom("tab", css.Hex("#1B1A24")),
	css.Custom("tab-on", css.Hex("#F7F2E6")),
	css.Custom("grid", css.RGBA(27, 26, 36, 0.07)),
	css.Custom("tape", css.RGBA(255, 244, 196, 0.78)),
	css.Custom("mis", css.Hex("#FF48B0")),
	css.Custom("mis-ink", css.Hex("#2C4FA8")),
	css.Custom("blend", kw.Multiply),
	css.Custom("grain-opacity", css.Num(0.33)),
	css.Custom("surface", css.Hex("#F4F0E8")),
	css.Custom("side", css.Hex("#FFD60A")),
	css.Custom("accent", css.Hex("#FF6F59")),
	css.Custom("mint", css.Hex("#5FE3B0")),
	css.Custom("pink", css.Hex("#C9B3FF")),
}

var dark = []css.Property{
	css.Custom("ground", css.Hex("#16151B")),
	css.Custom("panel", css.Hex("#211F28")),
	css.Custom("panel-2", css.Hex("#2E2B37")),
	css.Custom("ink", css.Hex("#F4EEDF")),
	css.Custom("ink-2", css.Hex("#BDB6C9")),
	css.Custom("line", css.Hex("#CFC8DA")),
	css.Custom("shadow-ink", css.Hex("#C8127C")),
	css.Custom("focus", css.Hex("#FFE600")),
	css.Custom("code", css.Hex("#1F1D26")),
	css.Custom("code-on", css.Hex("#F4EEDF")),
	css.Custom("code-dim", css.Hex("#8F8A9C")),
	css.Custom("code-inline", css.Hex("#4A4214")),
	css.Custom("code-key", css.Hex("#8FAAFF")),
	css.Custom("code-str", css.Hex("#FF7CC6")),
	css.Custom("code-fn", css.Hex("#F4EEDF")),
	css.Custom("code-type", css.Hex("#FFE600")),
	css.Custom("code-mark", css.RGBA(255, 72, 176, 0.38)),
	css.Custom("tab", css.Hex("#FFE600")),
	css.Custom("tab-on", css.Hex("#1B1A24")),
	css.Custom("grid", css.RGBA(244, 238, 223, 0.06)),
	css.Custom("tape", css.RGBA(255, 230, 0, 0.32)),
	css.Custom("mis", css.Hex("#C8127C")),
	css.Custom("mis-ink", css.Hex("#8FAAFF")),
	css.Custom("blend", kw.Screen),
	css.Custom("grain-opacity", css.Num(0.5)),
	css.Custom("surface", css.Hex("#17161C")),
	css.Custom("side", css.Hex("#7A6600")),
	css.Custom("accent", css.Hex("#FF6F59")),
	css.Custom("mint", css.Hex("#1F6B52")),
	css.Custom("pink", css.Hex("#6E5AA8")),
}

var metrics = []css.Property{
	css.Custom("border", css.Px(3)),
	css.Custom("border-l", css.Px(4)),
	css.Custom("r-s", css.Px(0)),
	css.Custom("r", css.Px(0)),
	css.Custom("r-l", css.Px(0)),
	css.Custom("shadow", css.Raw("5px 5px 0 var(--shadow-ink)")),
	css.Custom("shadow-l", css.Raw("8px 8px 0 var(--shadow-ink)")),
	css.Custom("shadow-s", css.Raw("3px 3px 0 var(--shadow-ink)")),
	css.Custom("display", css.Raw(`"Anton", "Impact", "Arial Narrow", sans-serif`)),
	css.Custom("sans", css.Raw(`"Work Sans", ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif`)),
	css.Custom("mono", css.Raw(`"Space Mono", ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`)),
	css.Custom("hand", css.Raw(`"Permanent Marker", "Comic Sans MS", cursive`)),
	css.Custom("noise", css.Raw(`url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='220' height='220'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='.85' numOctaves='3' stitchTiles='stitch'/%3E%3CfeColorMatrix values='0 0 0 0 .1 0 0 0 0 .1 0 0 0 0 .15 0 0 0 .55 0'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E")`)),
	css.Custom("halftone", css.Raw("radial-gradient(circle, var(--ink-pink) 34%, transparent 37%)")),
	css.Custom("mis-shadow", css.Raw("0.06em 0.04em 0 var(--mis)")),
}

var Fonts = style.Import(css.Url("https://fonts.googleapis.com/css2?family=Anton&family=Permanent+Marker&family=Space+Mono:wght@400;700&family=Work+Sans:wght@400;500;600;700&display=swap"))

var Root = style.Element("html").
	Set(light...).
	Set(fills...).
	Set(metrics...).
	Media(media.PrefersColorScheme(kw.Dark), dark...)

var ForceDark = style.Element("html").
	Set(dark...).
	On(DarkAction, darkRequested)

var ForceLight = style.Element("html").
	Set(light...).
	On(DarkAction, lightRequested)
