package style

import (
	"os"
	"strings"
	"testing"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/kw"
	"github.com/Toms223/WebGo/media"
)

var v1Light = []css.Property{
	css.Custom("signal", css.Hex("#FF5B23")),
	css.Custom("signal-dim", css.Hex("#C43D0E")),
	css.Custom("on-signal", css.Hex("#1A0A03")),
	css.Custom("coolant", css.Hex("#00C9BE")),
	css.Custom("coolant-dim", css.Hex("#00847C")),
	css.Custom("filament", css.Hex("#C8F53C")),
	css.Custom("surface", css.Hex("#F4F1EC")),
	css.Custom("surface-container-low", css.Hex("#EEEAE3")),
	css.Custom("surface-container", css.Hex("#E6E1D8")),
	css.Custom("surface-container-high", css.Hex("#DCD6CB")),
	css.Custom("surface-inverse", css.Hex("#14161A")),
	css.Custom("code-surface", css.Hex("#14161A")),
	css.Custom("code-on", css.Hex("#E9EAEC")),
	css.Custom("on-surface", css.Hex("#16181C")),
	css.Custom("on-surface-variant", css.Hex("#55585F")),
	css.Custom("on-surface-inverse", css.Hex("#E9EAEC")),
	css.Custom("outline", css.Hex("#B6B1A6")),
	css.Custom("outline-strong", css.Hex("#16181C")),
	css.Custom("error", css.Hex("#B3261E")),
	css.Custom("error-container", css.Hex("#F9DEDC")),
}

var v1Dark = []css.Property{
	css.Custom("signal", css.Hex("#FF6B37")),
	css.Custom("signal-dim", css.Hex("#FF9A74")),
	css.Custom("on-signal", css.Hex("#1A0A03")),
	css.Custom("coolant", css.Hex("#2BE8DC")),
	css.Custom("coolant-dim", css.Hex("#7DF3EB")),
	css.Custom("filament", css.Hex("#D6FF57")),
	css.Custom("surface", css.Hex("#0B0D10")),
	css.Custom("surface-container-low", css.Hex("#121519")),
	css.Custom("surface-container", css.Hex("#181C21")),
	css.Custom("surface-container-high", css.Hex("#21262D")),
	css.Custom("surface-inverse", css.Hex("#E9EAEC")),
	css.Custom("code-surface", css.Hex("#05070A")),
	css.Custom("code-on", css.Hex("#DCE3EA")),
	css.Custom("on-surface", css.Hex("#E7E9EC")),
	css.Custom("on-surface-variant", css.Hex("#9AA0A9")),
	css.Custom("on-surface-inverse", css.Hex("#14161A")),
	css.Custom("outline", css.Hex("#333941")),
	css.Custom("outline-strong", css.Hex("#4A525C")),
	css.Custom("error", css.Hex("#FF6B6B")),
	css.Custom("error-container", css.Hex("#2A1214")),
}

var v1Metrics = []css.Property{
	css.Custom("rail", css.Px(4)),
	css.Custom("shape-l", css.Px(24)),
	css.Custom("shape-m", css.Px(14)),
	css.Custom("shape-s", css.Px(8)),
	css.Custom("sans", css.Raw(`ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif`)),
	css.Custom("mono", css.Raw(`ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace`)),
}

var v1Narrow = media.MaxWidth(css.Rem(48))

var v1ColorFade = css.Transition(
	css.Raw("background-color 120ms linear"),
	css.Raw("color 120ms linear"),
	css.Raw("border-color 120ms linear"),
)

func v1Items() []Item {
	doc := Class("doc")
	return []Item{
		Element("html").
			Set(v1Light...).
			Set(v1Metrics...).
			Media(media.PrefersColorScheme(kw.Dark), v1Dark...),
		Any().Set(css.BoxSizing(kw.BorderBox)),
		Element("body").Set(
			css.Margin(css.Px(0)),
			css.BackgroundColor(css.Var("surface")),
			css.BackgroundImage(
				css.Raw("repeating-linear-gradient(90deg, transparent 0 63px, color-mix(in srgb, var(--outline) 28%, transparent) 63px 64px)"),
				css.Raw("repeating-linear-gradient(0deg, transparent 0 63px, color-mix(in srgb, var(--outline) 16%, transparent) 63px 64px)"),
			),
			css.Color(css.Var("on-surface")),
			css.FontFamily(css.Var("sans")),
			css.FontSize(css.Px(16)),
			css.LineHeight(css.Num(1.65)),
		),
		Class("shell").
			Set(
				css.MaxWidth(css.Rem(68)),
				css.Margin(css.Px(0), kw.Auto),
				css.Padding(css.Px(0), css.Rem(1.5), css.Rem(6)),
			).
			Media(v1Narrow, css.Padding(css.Px(0), css.Rem(1), css.Rem(4))),
		Class("shell-title").
			Set(
				css.Margin(css.Px(0)),
				css.FontSize(css.Clamp(css.Rem(2.25), css.Vw(7), css.Rem(3.75))),
				css.FontWeight(css.Num(800)),
				css.LetterSpacing(css.Em(-0.045)),
				css.LineHeight(css.Num(0.95)),
				css.TextTransform(kw.Uppercase),
			).
			Before(
				css.Content(css.String("GO \u002F\u002FWASM")),
				css.Display(kw.Block),
				css.MarginBottom(css.Rem(0.9)),
				css.FontFamily(css.Var("mono")),
				css.FontSize(css.Rem(0.7)),
				css.FontWeight(css.Num(600)),
				css.LetterSpacing(css.Em(0.32)),
				css.Color(css.Var("on-surface-variant")),
				css.TextTransform(kw.None),
			),
		Class("shell-toggle").
			Set(
				css.Position(kw.Absolute),
				css.Top(css.Rem(2.5)),
				css.Right(css.Px(0)),
				css.Padding(css.Rem(0.4), css.Rem(0.8)),
				css.Border(css.Px(1), kw.Solid, css.Var("outline")),
				css.BorderRadius(css.Var("shape-s")),
				css.Background(css.Var("surface-container-low")),
				css.Color(css.Var("on-surface-variant")),
				css.FontFamily(css.Var("mono")),
				css.FontSize(css.Rem(0.7)),
				css.FontWeight(css.Num(600)),
				css.LetterSpacing(css.Em(0.14)),
				css.TextTransform(kw.Uppercase),
				css.Cursor(kw.Pointer),
				v1ColorFade,
			).
			Hover(
				css.Color(css.Var("on-surface")),
				css.BorderColor(css.Var("outline-strong")),
			).
			FocusVisible(
				css.Outline(css.Px(2), kw.Solid, css.Var("signal")),
				css.OutlineOffset(css.Px(2)),
			).
			Media(v1Narrow, css.Top(css.Rem(1.5))),
		Class("nav").Descendant(Element("ul")).Set(
			css.Display(kw.Flex),
			css.FlexWrap(kw.Wrap),
			css.Gap(css.Rem(0.5)),
			css.Margin(css.Px(0)),
			css.Padding(css.Px(0)),
			css.ListStyle(kw.None),
		),
		Class("nav-link").
			Set(
				css.Position(kw.Relative),
				css.Display(kw.InlineFlex),
				css.AlignItems(kw.Center),
				css.Gap(css.Rem(0.6)),
				css.Padding(css.Rem(0.5), css.Rem(1), css.Rem(0.5), css.Rem(0.75)),
				css.Border(css.Px(1), kw.Solid, css.Var("outline")),
				css.BorderRadius(css.Px(999)),
				css.Color(css.Var("on-surface-variant")),
				css.FontSize(css.Rem(0.9)),
				css.FontWeight(css.Num(500)),
				css.TextDecoration(kw.None),
				css.Background(kw.Transparent),
				v1ColorFade,
			).
			Before(
				css.Content(css.Raw("counter(route, decimal-leading-zero)")),
				css.FontFamily(css.Var("mono")),
				css.FontSize(css.Rem(0.68)),
				css.FontWeight(css.Num(700)),
				css.LetterSpacing(css.Em(0.06)),
				css.Opacity(css.Num(0.65)),
			).
			Hover(
				css.Background(css.Raw("color-mix(in srgb, var(--on-surface) 8%, transparent)")),
				css.Color(css.Var("on-surface")),
				css.BorderColor(css.Var("outline-strong")),
			),
		doc.Child(Any()).FirstChild(css.MarginTop(css.Px(0))),
		doc.Descendant(Element("li")).
			Set(css.MarginBottom(css.Rem(0.4))).
			Marker(
				css.Color(css.Var("signal")),
				css.FontWeight(css.Num(700)),
			),
		doc.Descendant(Element("a")).
			Set(
				css.Color(css.Var("on-surface")),
				css.TextDecoration(kw.Underline),
				css.TextDecorationColor(css.Var("signal")),
				css.TextDecorationThickness(css.Px(2)),
				css.TextUnderlineOffset(css.Px(3)),
			).
			Hover(css.Background(css.Raw("color-mix(in srgb, var(--signal) 18%, transparent)"))),
	}
}

var v1Want = []string{
	`html {
	--signal: #FF5B23;
	--signal-dim: #C43D0E;
	--on-signal: #1A0A03;
	--coolant: #00C9BE;
	--coolant-dim: #00847C;
	--filament: #C8F53C;
	--surface: #F4F1EC;
	--surface-container-low: #EEEAE3;
	--surface-container: #E6E1D8;
	--surface-container-high: #DCD6CB;
	--surface-inverse: #14161A;
	--code-surface: #14161A;
	--code-on: #E9EAEC;
	--on-surface: #16181C;
	--on-surface-variant: #55585F;
	--on-surface-inverse: #E9EAEC;
	--outline: #B6B1A6;
	--outline-strong: #16181C;
	--error: #B3261E;
	--error-container: #F9DEDC;
	--rail: 4px;
	--shape-l: 24px;
	--shape-m: 14px;
	--shape-s: 8px;
	--sans: ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
	--mono: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace;
}`,
	`@media (prefers-color-scheme: dark) {
	html {
		--signal: #FF6B37;
		--signal-dim: #FF9A74;
		--on-signal: #1A0A03;
		--coolant: #2BE8DC;
		--coolant-dim: #7DF3EB;
		--filament: #D6FF57;
		--surface: #0B0D10;
		--surface-container-low: #121519;
		--surface-container: #181C21;
		--surface-container-high: #21262D;
		--surface-inverse: #E9EAEC;
		--code-surface: #05070A;
		--code-on: #DCE3EA;
		--on-surface: #E7E9EC;
		--on-surface-variant: #9AA0A9;
		--on-surface-inverse: #14161A;
		--outline: #333941;
		--outline-strong: #4A525C;
		--error: #FF6B6B;
		--error-container: #2A1214;
	}
}`,
	`* {
	box-sizing: border-box;
}`,
	`body {
	margin: 0px;
	background-color: var(--surface);
	background-image: repeating-linear-gradient(90deg, transparent 0 63px, color-mix(in srgb, var(--outline) 28%, transparent) 63px 64px), repeating-linear-gradient(0deg, transparent 0 63px, color-mix(in srgb, var(--outline) 16%, transparent) 63px 64px);
	color: var(--on-surface);
	font-family: var(--sans);
	font-size: 16px;
	line-height: 1.65;
}`,
	`.shell {
	max-width: 68rem;
	margin: 0px auto;
	padding: 0px 1.5rem 6rem;
}`,
	`@media (max-width: 48rem) {
	.shell {
		padding: 0px 1rem 4rem;
	}
}`,
	`.shell-title {
	margin: 0px;
	font-size: clamp(2.25rem, 7vw, 3.75rem);
	font-weight: 800;
	letter-spacing: -0.045em;
	line-height: 0.95;
	text-transform: uppercase;
}`,
	`.shell-title::before {
	content: "GO /` + `/WASM";
	display: block;
	margin-bottom: 0.9rem;
	font-family: var(--mono);
	font-size: 0.7rem;
	font-weight: 600;
	letter-spacing: 0.32em;
	color: var(--on-surface-variant);
	text-transform: none;
}`,
	`.shell-toggle {
	position: absolute;
	top: 2.5rem;
	right: 0px;
	padding: 0.4rem 0.8rem;
	border: 1px solid var(--outline);
	border-radius: var(--shape-s);
	background: var(--surface-container-low);
	color: var(--on-surface-variant);
	font-family: var(--mono);
	font-size: 0.7rem;
	font-weight: 600;
	letter-spacing: 0.14em;
	text-transform: uppercase;
	cursor: pointer;
	transition: background-color 120ms linear, color 120ms linear, border-color 120ms linear;
}`,
	`.shell-toggle:hover {
	color: var(--on-surface);
	border-color: var(--outline-strong);
}`,
	`.shell-toggle:focus-visible {
	outline: 2px solid var(--signal);
	outline-offset: 2px;
}`,
	`@media (max-width: 48rem) {
	.shell-toggle {
		top: 1.5rem;
	}
}`,
	`.nav ul {
	display: flex;
	flex-wrap: wrap;
	gap: 0.5rem;
	margin: 0px;
	padding: 0px;
	list-style: none;
}`,
	`.nav-link {
	position: relative;
	display: inline-flex;
	align-items: center;
	gap: 0.6rem;
	padding: 0.5rem 1rem 0.5rem 0.75rem;
	border: 1px solid var(--outline);
	border-radius: 999px;
	color: var(--on-surface-variant);
	font-size: 0.9rem;
	font-weight: 500;
	text-decoration: none;
	background: transparent;
	transition: background-color 120ms linear, color 120ms linear, border-color 120ms linear;
}`,
	`.nav-link::before {
	content: counter(route, decimal-leading-zero);
	font-family: var(--mono);
	font-size: 0.68rem;
	font-weight: 700;
	letter-spacing: 0.06em;
	opacity: 0.65;
}`,
	`.nav-link:hover {
	background: color-mix(in srgb, var(--on-surface) 8%, transparent);
	color: var(--on-surface);
	border-color: var(--outline-strong);
}`,
	`.doc > *:first-child {
	margin-top: 0px;
}`,
	`.doc li {
	margin-bottom: 0.4rem;
}`,
	`.doc li::marker {
	color: var(--signal);
	font-weight: 700;
}`,
	`.doc a {
	color: var(--on-surface);
	text-decoration: underline;
	text-decoration-color: var(--signal);
	text-decoration-thickness: 2px;
	text-underline-offset: 3px;
}`,
	`.doc a:hover {
	background: color-mix(in srgb, var(--signal) 18%, transparent);
}`,
}

func TestV1ThemeRendersByteIdentical(t *testing.T) {
	data, err := os.ReadFile("testdata/v1-theme.css")
	if err != nil {
		t.Fatal(err)
	}
	v1 := string(data)

	for _, block := range v1Want {
		if !strings.Contains(v1, block+"\n") {
			t.Errorf("testdata/v1-theme.css does not contain the expected block:\n%s", block)
		}
	}

	want := strings.Join(v1Want, "\n\n") + "\n"
	if got := render(v1Items()); got != want {
		t.Errorf("render(v1 subset) differs from the v1 theme slices\ngot:\n%s\nwant:\n%s", got, want)
	}
}

func TestV1PrefixedPropertyRendersPrefixFirst(t *testing.T) {
	got := render([]Item{
		Class("glass").
			Set(css.BackdropFilter(css.Raw("blur(4px)"))).
			Hover(css.BackdropFilter(css.Raw("blur(8px)"))).
			Media(v1Narrow, css.BackdropFilter(css.Raw("blur(2px)"))),
	})
	want := ".glass {\n\t-webkit-backdrop-filter: blur(4px);\n\tbackdrop-filter: blur(4px);\n}\n\n" +
		".glass:hover {\n\t-webkit-backdrop-filter: blur(8px);\n\tbackdrop-filter: blur(8px);\n}\n\n" +
		"@media (max-width: 48rem) {\n\t.glass {\n\t\t-webkit-backdrop-filter: blur(2px);\n\t\tbackdrop-filter: blur(2px);\n\t}\n}\n"
	if got != want {
		t.Errorf("render() = %q, want %q", got, want)
	}
}
