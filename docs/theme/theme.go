package theme

import (
	"slices"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	actions "github.com/Toms223/WebGo/docs/pages/actions/demos"
	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	browser "github.com/Toms223/WebGo/docs/pages/browser/demos"
	components "github.com/Toms223/WebGo/docs/pages/components/demos"
	state "github.com/Toms223/WebGo/docs/pages/state/demos"
	styling "github.com/Toms223/WebGo/docs/pages/styling/demos"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/style"
)

const DarkAction = "theme/dark"

var narrow = media.MaxWidth(css.Rem(48))

var Layers = style.LayerOrder("base", "components")

var Base = style.Layer("base", Root, Reset, Body, LiveButton, LiveLabel, LiveInput)

var Components = style.Layer("components", withDemos(
	Pulse, Reveal, RevealAgain, TurnB,
	Button, ButtonPrimary, Pill, Skip,
	Shell, Side, Logo, LogoMark, SideMenu, ShellToggle, Content, ContentItem,
	SideNav, NavLabel, SideNavList, NavGroupGap, NavLink, NavLinkActive, MenuOpenNav,
	Topic, TopicHeader, TopicMain, TopicTitle, TopicLead, TopicSection, SectionHeading,
	Rail, RailOpen, RailSummary, RailOpenChevron, RailCount, RailPanel, RailLabel, RailSearch, RailEmpty, RailList, RailLink, RailLinkSub, RailWhere, RailLinkActive,
	Pager, PagerLink, PagerNext, PagerLabel,
	Doc, CodeBlock, DocProse, DocHeading, DocSubheading, DocList, DocOrderedList, DocItem, DocNestedList,
	DocLink, DocStrong, DocCode, DocPre, DocPreGo, DocPreCSS, DocPreShell, DocPreMarkdown, DocPreHTML, DocPreCode,
	Note, NoteText, NoteCode, NoteLink, DocRule,
	Diagram, DiagramLane, DiagramTag, DiagramTitle, DiagramFlow, DiagramStep, DiagramStack, DiagramStackStep, DiagramBox, DiagramSun, DiagramLime, DiagramLilac, DiagramCoral,
	DiagramName, DiagramSub, DiagramArrow, DiagramGlyph, DiagramDown, DiagramVia, DiagramBranch, DiagramWays, DiagramWay, DiagramCaption,
	Example, ExampleBar, ExampleTools, ExamplePre, ExampleCodeClamped, ExampleMore, CopyButton, CopyFloat, ExampleName, ExampleTabs, ExampleTab, ExampleTabOn, ExampleCode, ExampleLabel, ExampleLive, ExampleLiveDoc,
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
	return slices.Concat(items, styling.Items(), applying.Items(), state.Items(), components.Items(), actions.Items(), browser.Items())
}

func darkRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && v
}

func lightRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && !v
}
