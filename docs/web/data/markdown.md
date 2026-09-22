The `markdown` package renders Markdown source into a tree of `go-app` `app.UI` elements. It is a subset renderer: it supports the constructs listed on this page and treats everything else as plain text.

## RenderMarkdown

```go
func RenderMarkdown(source string, style Style) []app.UI
```

`RenderMarkdown` parses `source` and returns a slice of `app.UI` block elements, one per top-level Markdown block (paragraph, heading, list, blockquote, code block, and so on).

If `source` is empty or contains only whitespace, `RenderMarkdown` returns `nil`.

Minimal usage:

```go
package main

import (
	"github.com/Toms223/WebGo/markdown"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func Body(source string) []app.UI {
	return markdown.RenderMarkdown(source, markdown.Style{
		Heading:    "heading",
		Subheading: "subheading",
	})
}
```

## Style

```go
type Style struct {
	Heading    string
	Subheading string
}
```

- `Heading` is set as the CSS class on every level 1 and level 2 heading.
- `Subheading` is set as the CSS class on every level 3 and deeper heading.

## Headings

A `#` or `##` heading renders as an `h3` carrying the `Style.Heading` class.

A `###` heading, or any deeper level, renders as an `h4` carrying the `Style.Subheading` class.

Example:

```
# Top
## Second
### Third
```

renders as:

```
<h3 class="heading">Top</h3>
<h3 class="heading">Second</h3>
<h4 class="subheading">Third</h4>
```

## Lists

- An ordered list (`1.`, `2.`, ...) renders as `ol`.
- An unordered list (`-`, `*`, `+`) renders as `ul`.
- Each list item renders as `li`.

A tight list item (no blank line between items) puts its text directly inside `li`, without wrapping it in a `p`. A list item that contains its own nested block content (a nested list, a blockquote, and so on) renders that nested content as child elements of `li`.

## Blockquotes

A `>` blockquote renders as `blockquote`, containing the rendered form of whatever blocks are inside it.

## Code blocks

Both a fenced code block (a block wrapped in a line of three or more backticks, or three or more tildes) and an indented code block (a block indented by four spaces) render the same way: a `pre` element containing a `code` element with the block's text.

A fenced block containing:

```
fmt.Println("hi")
```

renders as:

```
<pre><code>fmt.Println("hi")</code></pre>
```

## Thematic breaks

A thematic break (`---`, `***`, or `___` on its own line) renders as `hr`.

## Inline emphasis and code spans

- Single-level emphasis (`*text*` or `_text_`) renders as `em`.
- Double or higher emphasis (`**text**`, `***text***`, and so on) renders as `strong`.
- An inline code span (`` `text` ``) renders as `code`.

## Line breaks

- A hard line break (a line ending in two or more trailing spaces, then a newline) renders as the preceding text followed by a `br` element.
- A soft line break (a plain newline inside a paragraph) does not render a `br`. It renders as a single space appended to the preceding text, joining the two lines.

## Images

An image (`![alt](url)`) does not render as `img`. It renders as `em`, with the image's alt text as the emphasis body. The image URL is discarded.

## Raw HTML

Raw HTML, whether a block of HTML or an inline HTML tag, is never rendered as markup. It is rendered as escaped text, so tags such as `<script>` or `<img>` appear as visible, inert text rather than being interpreted by the browser.

## Links

A link (`[text](destination)`) or autolink (`<https://example.com>`) renders as an anchor (`a`) only if its destination passes a safety check. A safe anchor always carries:

- `rel="noopener noreferrer"`
- `target="_blank"`

A destination is treated as safe, and rendered as `a`, when it is one of:

- A relative path starting with `/`
- An anchor reference starting with `#`
- A URL starting with `http://`
- A URL starting with `https://`
- A URL starting with `mailto:`

The scheme check is case-insensitive, so `HTTPS://Example.com` is still treated as safe.

Every other destination renders as a plain `span` instead of `a`, with the link's body text preserved. This includes:

- A bare relative reference with no leading `/`, `#`, or scheme, such as `[guide](guide.md)`
- Protocol-relative URLs starting with `//` (for example `//evil.example/path`)
- Any other scheme, such as `javascript:`, `data:`, or `vbscript:`
- An empty or whitespace-only destination

## Unsupported syntax

Constructs the renderer has no specific handling for, such as tables, strikethrough (`~~text~~`), and task list checkboxes (`- [ ]`), are not rendered as their own elements. Their text content still appears in the output, carried through as plain text or inside the surrounding block.
