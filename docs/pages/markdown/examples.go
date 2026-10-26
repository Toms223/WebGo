package markdown

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/docs/theme"
	md "github.com/Toms223/WebGo/markdown"
)

const tick = "`"

var docStyle = md.Style{
	Heading:      theme.DocHeading.Name(),
	Subheading:   theme.DocSubheading.Name(),
	Example:      theme.Example.Name(),
	ExampleCode:  theme.ExampleCode.Name(),
	ExampleLive:  theme.ExampleLive.Name(),
	ExampleLabel: theme.ExampleLabel.Name(),
}

func render(source string, style md.Style) app.UI {
	return app.Div().Style(theme.Doc).Body(md.RenderMarkdown(source, style)...)
}

func mdExample(name, source string) example.Example {
	return example.Example{Name: name, Source: source, Lang: "md", Live: func() app.UI { return render(source, docStyle) }}
}

var paragraphExample = mdExample("paragraph.md", "A paragraph with *emphasis* and **strong** text.")

var headingsExample = mdExample("headings.md", `# Top
## Second
### Third`)

var listsExample = mdExample("lists.md", `1. First
2. Second

- Apples
- Pears
  - Conference
  - Williams`)

var blockquoteExample = mdExample("blockquote.md", `> A quoted paragraph.
>
> And a second one inside the same quote.`)

var codeBlocksExample = mdExample("code-blocks.md", tick+tick+tick+"\nfmt.Println(\"hi\")\n"+tick+tick+tick+"\n\n    indented := true")

var breaksExample = mdExample("breaks.md", `Above the break.

---

Below the break.`)

var inlineExample = mdExample("inline.md", "Single *emphasis* or _emphasis_, double **strong**, triple ***strong*** and a "+tick+"code span"+tick+".")

var lineBreaksExample = mdExample("line-breaks.md", "A hard break  \nstarts a new line.\n\nA soft break\njoins the lines.")

var imageExample = mdExample("image.md", `An image: ![A diagram of the pipeline](/img/pipeline.png)`)

var rawHTMLExample = mdExample("raw-html.md", `<script>alert("never runs")</script>

Inline <img src=x onerror=alert(1)> stays text.`)

var linksExample = mdExample("links.md", `A [safe link](https://example.com), an autolink <https://example.com> and an [unsafe one](javascript:alert(1)).`)

var liveDemoSource = "A fence naming a registered demo:\n\n" + tick + tick + tick + "go box\napp.Span().Text(\"a running component\")\n" + tick + tick + tick

var liveDemoExample = example.Example{Name: "live-demo.md", Source: liveDemoSource, Lang: "md", Live: func() app.UI {
	style := docStyle
	style.Demos = map[string]func() app.UI{"box": func() app.UI { return app.Span().Text("a running component") }}
	return render(liveDemoSource, style)
}}

var unsupportedExample = mdExample("unsupported.md", `| Name | Value |
|------|-------|
| a    | 1     |

~~struck~~ text and a task: - [ ] todo`)
