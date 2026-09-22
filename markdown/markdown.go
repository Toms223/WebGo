package news_feed

import (
	"strings"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	textm "github.com/yuin/goldmark/text"
)

var markdownParser = goldmark.New().Parser()

type Style struct {
	Heading    string
	Subheading string
}

func RenderMarkdown(source string, style Style) []app.UI {
	if strings.TrimSpace(source) == "" {
		return nil
	}
	src := []byte(source)
	return blocks(markdownParser.Parse(textm.NewReader(src)), src, style)
}

func blocks(parent ast.Node, src []byte, style Style) []app.UI {
	var out []app.UI
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		if ui := block(node, src, style); ui != nil {
			out = append(out, ui)
		}
	}
	return out
}

func block(node ast.Node, src []byte, style Style) app.UI {
	switch typed := node.(type) {
	case *ast.Heading:
		return heading(typed, src, style)
	case *ast.List:
		return list(typed, src, style)
	case *ast.ListItem:
		return listItem(typed, src, style)
	case *ast.Blockquote:
		return app.Blockquote().Body(blocks(node, src, style)...)
	case *ast.FencedCodeBlock:
		return codeBlock(typed.Lines(), src)
	case *ast.CodeBlock:
		return codeBlock(typed.Lines(), src)
	case *ast.ThematicBreak:
		return app.Hr()
	case *ast.HTMLBlock:
		return app.P().Text(lines(typed.Lines(), src))
	default:
		return app.P().Body(inlines(node, src)...)
	}
}

func heading(node *ast.Heading, src []byte, style Style) app.UI {
	if node.Level <= 2 {
		return app.H3().Class(style.Heading).Body(inlines(node, src)...)
	}
	return app.H4().Class(style.Subheading).Body(inlines(node, src)...)
}

func list(node *ast.List, src []byte, style Style) app.UI {
	items := blocks(node, src, style)
	if node.IsOrdered() {
		return app.Ol().Body(items...)
	}
	return app.Ul().Body(items...)
}

func listItem(node *ast.ListItem, src []byte, style Style) app.UI {
	var body []app.UI
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if _, tight := child.(*ast.TextBlock); tight {
			body = append(body, inlines(child, src)...)
			continue
		}
		if ui := block(child, src, style); ui != nil {
			body = append(body, ui)
		}
	}
	return app.Li().Body(body...)
}

func codeBlock(segments *textm.Segments, src []byte) app.UI {
	return app.Pre().Body(
		app.Code().Text(lines(segments, src)),
	)
}

func lines(segments *textm.Segments, src []byte) string {
	var out strings.Builder
	for i := 0; i < segments.Len(); i++ {
		segment := segments.At(i)
		out.Write(segment.Value(src))
	}
	return out.String()
}

func inlines(parent ast.Node, src []byte) []app.UI {
	var out []app.UI
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		out = append(out, inline(node, src)...)
	}
	return out
}

func inline(node ast.Node, src []byte) []app.UI {
	switch typed := node.(type) {
	case *ast.Text:
		return textNode(typed, src)
	case *ast.String:
		return []app.UI{app.Text(string(typed.Value))}
	case *ast.CodeSpan:
		return []app.UI{app.Code().Body(inlines(node, src)...)}
	case *ast.Emphasis:
		if typed.Level >= 2 {
			return []app.UI{app.Strong().Body(inlines(node, src)...)}
		}
		return []app.UI{app.Em().Body(inlines(node, src)...)}
	case *ast.Link:
		return []app.UI{link(string(typed.Destination), inlines(node, src))}
	case *ast.AutoLink:
		url := string(typed.URL(src))
		return []app.UI{link(url, []app.UI{app.Text(url)})}
	case *ast.Image:
		return []app.UI{app.Em().Body(inlines(node, src)...)}
	case *ast.RawHTML:
		return []app.UI{app.Text(lines(typed.Segments, src))}
	default:
		return inlines(node, src)
	}
}

func textNode(node *ast.Text, src []byte) []app.UI {
	value := string(node.Segment.Value(src))
	switch {
	case node.HardLineBreak():
		return []app.UI{app.Text(value), app.Br()}
	case node.SoftLineBreak():
		return []app.UI{app.Text(value + " ")}
	default:
		return []app.UI{app.Text(value)}
	}
}

func link(destination string, body []app.UI) app.UI {
	href, ok := safeHref(destination)
	if !ok {
		return app.Span().Body(body...)
	}
	return app.A().
		Href(href).
		Rel("noopener noreferrer").
		Target("_blank").
		Body(body...)
}

func safeHref(destination string) (string, bool) {
	href := strings.TrimSpace(destination)
	if href == "" {
		return "", false
	}
	if strings.HasPrefix(href, "//") {
		return "", false
	}
	if strings.HasPrefix(href, "/") || strings.HasPrefix(href, "#") {
		return href, true
	}
	lowered := strings.ToLower(href)
	for _, scheme := range []string{"http://", "https://", "mailto:"} {
		if strings.HasPrefix(lowered, scheme) {
			return href, true
		}
	}
	return "", false
}
