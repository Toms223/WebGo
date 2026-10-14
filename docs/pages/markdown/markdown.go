package markdown

import (
	"github.com/Toms223/WebGo/docs/components/doc"
	"github.com/Toms223/WebGo/docs/components/topic"
	"github.com/Toms223/WebGo/docs/pages/base"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

const (
	Title = "Markdown rendering"
	Route = "/markdown"
)

func New() (*topic.Screen, error) {
	return topic.New(topic.Config{
		Key:   "markdown",
		Path:  Route,
		Title: Title,
		Lead:  "Render a safe subset of markdown into go-app elements, with live components in fences.",
		Intro: intro,
		Sections: []topic.Section{
			{Title: "RenderMarkdown", Body: renderMarkdown},
			{Title: "Style", Body: styleSection},
			{Title: "Headings", Body: headings},
			{Title: "Lists", Body: lists},
			{Title: "Blockquotes", Body: blockquotes},
			{Title: "Code blocks", Body: codeBlocks},
			{Title: "Thematic breaks", Body: thematicBreaks},
			{Title: "Inline emphasis and code spans", Body: inlineEmphasis},
			{Title: "Line breaks", Body: lineBreaks},
			{Title: "Images", Body: images},
			{Title: "Raw HTML", Body: rawHTML},
			{Title: "Links", Body: links},
			{Title: "Live demos", Body: liveDemos},
			{Title: "Unsupported syntax", Body: unsupported},
		},
		OnActivate: base.Activate,
	})
}

func intro() []app.UI {
	return []app.UI{
		doc.P("The `markdown` package renders Markdown source into a tree of `go-app` `app.UI` elements. It is a subset renderer: it supports the constructs listed on this page and treats everything else as plain text."),
	}
}

func renderMarkdown() []app.UI {
	return []app.UI{
		doc.Go(`func RenderMarkdown(source string, style Style) []app.UI`),
		doc.P("`RenderMarkdown` parses `source` and returns a slice of `app.UI` block elements, one per top-level Markdown block (paragraph, heading, list, blockquote, code block, and so on)."),
		doc.P("If `source` is empty or contains only whitespace, `RenderMarkdown` returns `nil`."),
		doc.P("Minimal usage:"),
		doc.Go(`package main

import (
	"github.com/Toms223/WebGo/markdown"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Body(source string) []app.UI {
	return markdown.RenderMarkdown(source, markdown.Style{
		Heading:    "heading",
		Subheading: "subheading",
	})
}`),
		doc.P("A paragraph rendered through it:"),
		doc.Example(paragraphExample),
	}
}

func styleSection() []app.UI {
	return []app.UI{
		doc.Go(`type Style struct {
	Heading      string
	Subheading   string
	Demos        map[string]func() app.UI
	Example      string
	ExampleCode  string
	ExampleLive  string
	ExampleLabel string
}`),
		doc.List(
			doc.Item("`Heading` is set as the CSS class on every level 1 and level 2 heading."),
			doc.Item("`Subheading` is set as the CSS class on every level 3 and deeper heading."),
			doc.Item("`Demos` maps a demo name to the function that builds its live component. A fenced block whose info string names a registered demo renders as a panel; see ", doc.A("#live-demos", "Live demos"), "."),
			doc.Item("`Example` is set as the CSS class on the panel that wraps a demo fence."),
			doc.Item("`ExampleCode` is set as the CSS class on the panel element holding the fence's code block."),
			doc.Item("`ExampleLive` is set as the CSS class on the panel element holding the live component."),
			doc.Item("`ExampleLabel` is set as the CSS class on the panel's \"Go\" and \"Result\" labels."),
		),
	}
}

func headings() []app.UI {
	return []app.UI{
		doc.P("A `#` or `##` heading renders as an `h3` carrying the `Style.Heading` class."),
		doc.P("A `###` heading, or any deeper level, renders as an `h4` carrying the `Style.Subheading` class."),
		doc.P("Example:"),
		doc.Example(headingsExample),
		doc.P("renders as:"),
		doc.HTML(`<h3 class="heading">Top</h3>
<h3 class="heading">Second</h3>
<h4 class="subheading">Third</h4>`),
	}
}

func lists() []app.UI {
	return []app.UI{
		doc.List(
			doc.Item("An ordered list (`1.`, `2.`, ...) renders as `ol`."),
			doc.Item("An unordered list (`-`, `*`, `+`) renders as `ul`."),
			doc.Item("Each list item renders as `li`."),
		),
		doc.P("A tight list item (no blank line between items) puts its text directly inside `li`, without wrapping it in a `p`. A list item that contains its own nested block content (a nested list, a blockquote, and so on) renders that nested content as child elements of `li`."),
		doc.Example(listsExample),
	}
}

func blockquotes() []app.UI {
	return []app.UI{
		doc.P("A `>` blockquote renders as `blockquote`, containing the rendered form of whatever blocks are inside it."),
		doc.Example(blockquoteExample),
	}
}

func codeBlocks() []app.UI {
	return []app.UI{
		doc.P("Both a fenced code block (a block wrapped in a line of three or more backticks, or three or more tildes) and an indented code block (a block indented by four spaces) render the same way: a `pre` element containing a `code` element with the block's text."),
		doc.P("A fenced block and an indented block containing:"),
		doc.Example(codeBlocksExample),
		doc.P("The fenced block renders as:"),
		doc.HTML(`<pre><code>fmt.Println("hi")</code></pre>`),
	}
}

func thematicBreaks() []app.UI {
	return []app.UI{
		doc.P("A thematic break (`---`, `***`, or `___` on its own line) renders as `hr`."),
		doc.Example(breaksExample),
	}
}

func inlineEmphasis() []app.UI {
	return []app.UI{
		doc.List(
			doc.Item("Single-level emphasis (`*text*` or `_text_`) renders as `em`."),
			doc.Item("Double or higher emphasis (`**text**`, `***text***`, and so on) renders as `strong`."),
			doc.Item("An inline code span (", doc.C("`text`"), ") renders as `code`."),
		),
		doc.Example(inlineExample),
	}
}

func lineBreaks() []app.UI {
	return []app.UI{
		doc.List(
			doc.Item("A hard line break (a line ending in two or more trailing spaces, then a newline) renders as the preceding text followed by a `br` element."),
			doc.Item("A soft line break (a plain newline inside a paragraph) does not render a `br`. It renders as a single space appended to the preceding text, joining the two lines."),
		),
		doc.Example(lineBreaksExample),
	}
}

func images() []app.UI {
	return []app.UI{
		doc.P("An image (`![alt](url)`) does not render as `img`. It renders as `em`, with the image's alt text as the emphasis body. The image URL is discarded."),
		doc.Example(imageExample),
	}
}

func rawHTML() []app.UI {
	return []app.UI{
		doc.P("Raw HTML, whether a block of HTML or an inline HTML tag, is never rendered as markup. It is rendered as escaped text, so tags such as `<script>` or `<img>` appear as visible, inert text rather than being interpreted by the browser."),
		doc.Example(rawHTMLExample),
	}
}

func links() []app.UI {
	return []app.UI{
		doc.P("A link (`[text](destination)`) or autolink (`<https://example.com>`) renders as an anchor (`a`) only if its destination passes a safety check. A safe anchor always carries:"),
		doc.List(
			doc.Item("`rel=\"noopener noreferrer\"`"),
			doc.Item("`target=\"_blank\"`"),
		),
		doc.P("A destination is treated as safe, and rendered as `a`, when it is one of:"),
		doc.List(
			doc.Item("A relative path starting with `/`"),
			doc.Item("An anchor reference starting with `#`"),
			doc.Item("A URL starting with `http://`"),
			doc.Item("A URL starting with `https://`"),
			doc.Item("A URL starting with `mailto:`"),
		),
		doc.P("The scheme check is case-insensitive, so `HTTPS://Example.com` is still treated as safe."),
		doc.P("Every other destination renders as a plain `span` instead of `a`, with the link's body text preserved. This includes:"),
		doc.List(
			doc.Item("A bare relative reference with no leading `/`, `#`, or scheme, such as `[guide](guide.md)`"),
			doc.Item("Protocol-relative URLs starting with `//` (for example `//evil.example/path`)"),
			doc.Item("Any other scheme, such as `javascript:`, `data:`, or `vbscript:`"),
			doc.Item("An empty or whitespace-only destination"),
		),
		doc.Example(linksExample),
	}
}

func liveDemos() []app.UI {
	return []app.UI{
		doc.P("A fenced block whose info string has a second word, as in ", doc.C("```go name"), ", names a demo. When `Style.Demos[name]` is registered, the block renders as a panel carrying the `Example` class, holding a \"Go\" label (`ExampleLabel`), the code (`ExampleCode`), a \"Result\" label and the live component returned by the registered function (`ExampleLive`). A name that is not registered renders as a plain code block."),
		doc.P("Here `box` is registered to a function returning `app.Span().Text(\"a running component\")`:"),
		doc.Example(liveDemoExample),
		doc.Go(`func DemoNames(source string) []string`),
		doc.P("`DemoNames` lists the demo names a source references, so a caller can check that every fence has a registered component before rendering."),
	}
}

func unsupported() []app.UI {
	return []app.UI{
		doc.P("Constructs the renderer has no specific handling for, such as tables, strikethrough (`~~text~~`), and task list checkboxes (`- [ ]`), are not rendered as their own elements. Their text content still appears in the output, carried through as plain text or inside the surrounding block."),
		doc.Example(unsupportedExample),
	}
}
