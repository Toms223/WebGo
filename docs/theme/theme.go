package theme

import (
	"slices"

	"github.com/Toms223/WebGo/css"
	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	state "github.com/Toms223/WebGo/docs/pages/state/demos"
	styling "github.com/Toms223/WebGo/docs/pages/styling/demos"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const DarkAction = "theme/dark"

var narrow = media.MaxWidth(css.Rem(48))

var Layers = style.LayerOrder("base", "components")

var Base = style.Layer("base", Root, Reset, Body, LiveButton, LiveLabel, LiveInput)

var Components = style.Layer("components", withDemos(
	Pulse, Reveal, RevealAgain, TurnB,
	Button, ButtonPrimary, Pill,
	Shell, Side, Logo, LogoMark, SideMenu, ShellToggle, Content, ContentItem,
	SideNav, NavLabel, SideNavList, NavLink, NavLinkActive, MenuOpenNav,
	Topic, TopicHeader, TopicTitle, TopicLead, TopicIntro, TopicSection, SectionHeading,
	Rail, RailLabel, RailList, RailLink, RailLinkActive,
	Pager, PagerLink, PagerNext, PagerLabel,
	Doc, DocProse, DocHeading, DocSubheading, DocList, DocOrderedList, DocItem, DocNestedList,
	DocLink, DocStrong, DocCode, DocPre, DocPreGo, DocPreCSS, DocPreShell, DocPreMarkdown, DocPreHTML, DocPreCode,
	Note, NoteText, NoteCode, NoteLink, DocRule,
	Example, ExampleBar, ExampleName, ExampleTabs, ExampleTab, ExampleTabOn, ExampleCode, ExampleLabel, ExampleLive, ExampleLiveDoc,
	TopicLoading, TopicError, ShellLoading, ShellError,
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
	return slices.Concat(items, styling.Items(), applying.Items(), state.Items())
}

func darkRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && v
}

func lightRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && !v
}
