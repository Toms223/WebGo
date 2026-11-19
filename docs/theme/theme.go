package theme

import (
	"slices"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/css"
	actions "github.com/Toms223/WebGo/docs/pages/actions/demos"
	applying "github.com/Toms223/WebGo/docs/pages/applying/demos"
	browser "github.com/Toms223/WebGo/docs/pages/browser/demos"
	components "github.com/Toms223/WebGo/docs/pages/components/demos"
	scripts "github.com/Toms223/WebGo/docs/pages/scripts/demos"
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
	Topic, TopicHeader, TopicMain, TopicTitle, TopicLead, TopicSection, SectionHeading, SectionNumber, TopicFolio,
	Rail, RailOpen, RailSummary, RailOpenChevron, RailCount, RailPanel, RailLabel, RailSearch, RailEmpty, RailList, RailLink, RailLinkSub, RailWhere, RailLinkActive, RailStamp, RailStampNumber,
	Pager, PagerLink, PagerNext, PagerLabel,
	Doc, CodeBlock, DocProse, DocHeading, DocSubheading, DocList, DocOrderedList, DocItem, DocNestedList,
	DocLink, DocStrong, DocCode, DocPre, DocPreGo, DocPreCSS, DocPreShell, DocPreMarkdown, DocPreHTML, DocPreCode,
	Note, NoteText, NoteCode, NoteLink, DocRule,
	Diagram, DiagramLane, DiagramTag, DiagramTitle, DiagramFlow, DiagramStep, DiagramStack, DiagramStackStep, DiagramBox, DiagramSun, DiagramLime, DiagramLilac, DiagramCoral,
	DiagramName, DiagramSub, DiagramArrow, DiagramGlyph, DiagramDown, DiagramVia, DiagramBranch, DiagramWays, DiagramWay, DiagramCaption,
	Example, ExampleBar, ExampleTools, ExamplePre, ExampleCodeClamped, ExampleMore, CopyButton, CopyFloat, ExampleName, ExampleTabs, ExampleTab, ExampleTabOn, ExampleCode, ExampleLabel, ExampleLive, ExampleLiveDoc,
	TopicLoading, TopicError, ShellLoading, ShellError,
	Landing, Topbar, TopbarNav, TopbarToggle, Wrap, Slug, SlugIssue, SlugWide,
	Hero, HeroInk, HeroDots, HeroHead, HeroCopy, HeroInstall, HeroSample, HeroTitle, Misregister, Mis, Lead, HeroLead, Actions, HandNote, Chip,
	Code, CodeBar, CodeNote, CodeFoot, CodeBody, CodeTilt, CodeLean, CodeQuiet, CodeQuietBody, Tape, TapeEnd, TapePink, Keyword, Str, Fn, Typ, Dim, Mark,
	Scribble, Draw, ScribblePath, ScribbleArrow, ScribbleWhole, ScribbleWholeArrow, ScribbleCircle, ScribbleRename, ScribbleRenameArrow,
	Spin, Sticker, StickerArt, StickerRing, Reg, RegEnd, Folio, FolioLow, InkPink, InkBlue, InkYellow,
	Eyebrow, Section, SectionTitle, Slab, Tiles, Card, Plate, PlateA, PlateB, PlateC, PlateD, PlateLetter, PlateLetterLight, PlateTag, CardTitle, CardText,
	FillSun, FillLime, FillLilac, FillCoral,
	Press, PressCopy, PressText, Bed, BedHead, BedCount, BedSheet, BedFoot, Thunk, ThunkAgain, Stamp, StampAgain, StampIdle, PressButton,
	Band, Marquee, MarqueeDot, Two, TwoTitle, TwoText, TwoCode, Points, PointsItem, PointsTick, Compiler, CompilerPrompt, CompilerError, Caption,
	Steps, Step, StepNext, StepNumber, StepTitle, StepText, StepCode, StepCodeBody,
	Closing, ClosingText, ClosingLead, ButtonBlue, Footer, FooterBig, FooterMis, FooterRow, ColorBar, FooterFine,
	News, NewsHead, NewsMasthead, NewsIntro, NewsCover, NewsInk, NewsBlock, NewsDots, NewsDisc, NewsCoverText, NewsCoverLabel, NewsCoverName, NewsCoverCard,
	NewsStamp, NewsStampNumber, NewsCoverStamp,
	NewsMeta, NewsFeature, NewsStory, NewsHeadline, NewsTeaser, NewsGo, NewsCatchUp, NewsSectionTitle, NewsAsk, NewsAskText, NewsPicks, NewsPick, NewsPickOn,
	NewsAnswer, NewsAnswerTitle, NewsMissed, NewsMissedLink, NewsCommand, NewsCommandText, NewsEarlier, NewsEarlierTitle, NewsRow, NewsRowTitle,
	NewsArticle, NewsBack, NewsIssue, NewsIssueDisc, NewsIssueBlock, NewsIssueStamp, NewsIssueMeta, NewsTitle, NewsSpread, NewsStaple, NewsStapleLow, NewsPage, NewsPageRight, NewsRunhead,
	NewsShort, NewsShortTitle, NewsShortList, NewsShortMark, NewsText, NewsTextFlow, NewsTextHeading, NewsDrop, NewsFolio, NewsFolioRight, NewsPull, NewsPullCode,
	NewsSign, NewsEndmark, NewsPager, NewsPagerLink, NewsPagerLabel, NewsPagerTitle, NewsCompare,
	LatestNews, LatestNewsCover, LatestNewsTitle, LatestNewsTitleLink,
	NewsMastheadIn, NewsIntroIn, NewsMissedNote, LatestNewsIn,
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
	return slices.Concat(items, styling.Items(), applying.Items(), state.Items(), components.Items(), actions.Items(), browser.Items(), scripts.Items())
}

func darkRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && v
}

func lightRequested(a app.Action) bool {
	v, ok := a.Value.(bool)
	return ok && !v
}
