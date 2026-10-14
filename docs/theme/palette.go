package theme

import (
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

var fills = []css.Property{
	css.Custom("sun", css.Hex("#FFD60A")),
	css.Custom("lime", css.Hex("#C6F432")),
	css.Custom("lilac", css.Hex("#C9B3FF")),
	css.Custom("coral", css.Hex("#FF6F59")),
	css.Custom("on-fill", css.Hex("#15131A")),
	css.Custom("code-key", css.Hex("#FFA394")),
	css.Custom("code-str", css.Hex("#93EBC5")),
	css.Custom("code-fn", css.Hex("#FFE27A")),
	css.Custom("code-type", css.Hex("#C9B3FF")),
}

var light = []css.Property{
	css.Custom("ground", css.Hex("#F4F0E8")),
	css.Custom("panel", css.Hex("#FFFFFF")),
	css.Custom("panel-2", css.Hex("#EAE4D8")),
	css.Custom("ink", css.Hex("#15131A")),
	css.Custom("ink-2", css.Hex("#4C4857")),
	css.Custom("line", css.Hex("#15131A")),
	css.Custom("shadow-ink", css.Hex("#15131A")),
	css.Custom("focus", css.Hex("#15131A")),
	css.Custom("code", css.Hex("#1A1826")),
	css.Custom("code-on", css.Hex("#F1EEFF")),
	css.Custom("code-dim", css.Hex("#8E88AA")),
	css.Custom("code-inline", css.Hex("#FFF0A6")),
	css.Custom("tab", css.Hex("#15131A")),
	css.Custom("tab-on", css.Hex("#F4F0E8")),
	css.Custom("grid", css.RGBA(21, 19, 26, 0.07)),
	css.Custom("surface", css.Hex("#F4F0E8")),
	css.Custom("side", css.Hex("#FFD60A")),
	css.Custom("accent", css.Hex("#FF6F59")),
	css.Custom("mint", css.Hex("#5FE3B0")),
	css.Custom("pink", css.Hex("#C9B3FF")),
}

var dark = []css.Property{
	css.Custom("ground", css.Hex("#17161C")),
	css.Custom("panel", css.Hex("#23222B")),
	css.Custom("panel-2", css.Hex("#3A3946")),
	css.Custom("ink", css.Hex("#F3F0EA")),
	css.Custom("ink-2", css.Hex("#B5B0C2")),
	css.Custom("line", css.Hex("#9791AE")),
	css.Custom("shadow-ink", css.Hex("#4A4760")),
	css.Custom("focus", css.Hex("#F3F0EA")),
	css.Custom("code", css.Hex("#0F0E15")),
	css.Custom("code-on", css.Hex("#EDEAFB")),
	css.Custom("code-dim", css.Hex("#7E789A")),
	css.Custom("code-inline", css.Hex("#534924")),
	css.Custom("tab", css.Hex("#C9B3FF")),
	css.Custom("tab-on", css.Hex("#15131A")),
	css.Custom("grid", css.RGBA(243, 240, 234, 0.06)),
	css.Custom("surface", css.Hex("#17161C")),
	css.Custom("side", css.Hex("#7A6600")),
	css.Custom("accent", css.Hex("#FF6F59")),
	css.Custom("mint", css.Hex("#1F6B52")),
	css.Custom("pink", css.Hex("#6E5AA8")),
}

var metrics = []css.Property{
	css.Custom("border", css.Px(2)),
	css.Custom("border-l", css.Px(3)),
	css.Custom("r-s", css.Px(6)),
	css.Custom("r", css.Px(12)),
	css.Custom("r-l", css.Px(18)),
	css.Custom("shadow", css.Raw("4px 4px 0 var(--shadow-ink)")),
	css.Custom("shadow-l", css.Raw("6px 6px 0 var(--shadow-ink)")),
	css.Custom("shadow-s", css.Raw("2px 2px 0 var(--shadow-ink)")),
	css.Custom("display", css.Raw(`"Bricolage Grotesque", "Archivo", system-ui, sans-serif`)),
	css.Custom("sans", css.Raw(`"Archivo", ui-sans-serif, system-ui, -apple-system, "Segoe UI", sans-serif`)),
	css.Custom("mono", css.Raw(`"JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, Consolas, monospace`)),
}

var Fonts = style.Import(css.Url("https://fonts.googleapis.com/css2?family=Archivo:wght@400;500;600;700&family=Bricolage+Grotesque:opsz,wght@12..96,700;12..96,800&family=JetBrains+Mono:wght@500;700&display=swap"))

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
