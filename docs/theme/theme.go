package theme

import (
	"slices"

	"github.com/Toms223/WebGo/css"
	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	styling "github.com/Toms223/WebGo/docs/pages/styling/demos"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const DarkAction = "theme/dark"

var narrow = media.MaxWidth(css.Rem(48))

var Layers = style.LayerOrder("base", "components")

var Base = style.Layer("base", Root, Reset, Body)

var Components = style.Layer("components", withDemos(
	Pulse,
	Button, ButtonPrimary, Pill,
	Shell, Side, Logo, LogoMark, SideMenu, ShellToggle, Content, ContentItem,
	SideNav, NavLabel, SideNavList, NavLink, NavLinkActive, MenuOpenNav,
	Topic, TopicItem, TopicTitle,
	Sections, SectionLink, SectionLinkActive,
	DocFirst, DocParagraph, DocHeading, DocSubheading, DocList, DocOrderedList, DocItem,
	DocLink, DocStrong, DocCode, DocPre, DocPreGo, DocPreCSS, DocPreShell, DocPreCode,
	DocQuote, DocQuoteLast, DocQuoteCode, DocQuoteLink, DocRule,
	Example, ExampleLabel, ExampleCode, ExampleLive,
	TopicLoading, TopicError, ShellLoading, ShellError, DocEmpty,
	Landing, Topbar, TopbarNav, TopbarToggle, Wrap,
	Hero, HeroCopy, HeroInstall, HeroSample, HeroTitle, Highlight, Lead, Actions, Chip,
	Code, CodeBar, CodeBody, CodeTilt, CodeQuiet, Keyword, Str, Fn, Typ, Dim,
	Eyebrow, Section, SectionTitle, Tiles, Card, CardIcon, CardTitle, CardText,
	FillSun, FillLime, FillLilac, FillCoral,
	Band, Marquee, MarqueeDot, Two, TwoText, TwoCode, Sticker, Points, PointsItem,
	Steps, Step, StepNumber, StepTitle, StepText, StepCode, StepCodeBody,
	Closing, ClosingText, ClosingLead, Footer,
	ReducedMotion,
)...)

func Styles() []style.Style {
	return []style.Style{
		style.Static("/web/app.css"),
		style.Generated("/web/theme.css", Fonts, Layers, Base, Components),
		style.Dynamic(append([]style.Item{ForceDark, ForceLight}, applying.Dynamic()...)...),
	}
}

func withDemos(items ...style.Item) []style.Item {
	return slices.Concat(items, styling.Items(), applying.Items())
}

func darkRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && v
}

func lightRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && !v
}
