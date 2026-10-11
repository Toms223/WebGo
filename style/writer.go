package style

import (
	"strings"

	"github.com/Toms223/WebGo/css"
)

func render(items []Item) string {
	var texts []string
	for _, i := range ordered(items) {
		texts = append(texts, i.blockTexts()...)
	}
	if len(texts) == 0 {
		return ""
	}
	return strings.Join(texts, "\n\n") + "\n"
}

func ordered(items []Item) []Item {
	var orders, imports, rest []Item
	for _, i := range items {
		switch i.(type) {
		case *LayerOrderRule:
			orders = append(orders, i)
		case *ImportRule:
			imports = append(imports, i)
		default:
			rest = append(rest, i)
		}
	}
	return append(append(orders, imports...), rest...)
}

func renderBlock(selector string, props []css.Property, depth int) string {
	return braces(selector, declarations(props, depth+1), depth)
}

func renderConditional(prelude, selector string, props []css.Property) string {
	return braces(prelude, renderBlock(selector, props, 1)+"\n", 0)
}

func renderNested(selector string, r *Rule, base []css.Property, depth int) string {
	return wrap(selector, nestedBody(r, base, depth+1), depth)
}

func nestedBody(r *Rule, base []css.Property, depth int) string {
	text := declarations(base, depth)
	for _, b := range r.blocks {
		if b.kind == pseudoKind {
			text += wrap("&"+b.pseudo, pseudoBody(r, b, depth+1), depth)
		}
	}
	text += conditionals(r, "", depth)
	for _, c := range r.children {
		text += nestedChild(r, c, depth)
	}
	return text
}

func pseudoBody(r *Rule, pseudo *block, depth int) string {
	return declarations(pseudo.props, depth) + conditionals(r, pseudo.pseudo, depth)
}

func conditionals(r *Rule, under string, depth int) string {
	var text string
	for _, b := range r.blocks {
		if b.conditional() && b.under == under {
			text += wrap(b.prelude(), declarations(b.props, depth+1), depth)
		}
	}
	return text
}

func nestedChild(parent, child *Rule, depth int) string {
	if child.self {
		return nestedBody(child, child.base(), depth)
	}
	sel := nestedSelector(parent, child)
	if lone := child.lonePseudo(); lone != nil {
		return wrap(sel+lone.pseudo, pseudoBody(child, lone, depth+1), depth)
	}
	return renderNested(sel, child, child.base(), depth)
}

func nestedSelector(parent, child *Rule) string {
	if child.self {
		return "&"
	}
	for parent.self && parent.owner != nil {
		parent = parent.owner
	}
	sel := child.Selector()
	for _, combinator := range []string{" > ", " + ", " ~ ", " "} {
		if rest, ok := strings.CutPrefix(sel, parent.Selector()+combinator); ok {
			return "&" + combinator + rest
		}
	}
	if rest, ok := strings.CutPrefix(sel, parent.Selector()); ok && rest != "" && strings.ContainsRune(".#:", rune(rest[0])) {
		return "&" + rest
	}
	return "& " + sel
}

func wrap(head, body string, depth int) string {
	if body == "" {
		return ""
	}
	return braces(head, body, depth) + "\n"
}

func braces(head, body string, depth int) string {
	tabs := strings.Repeat("\t", depth)
	return tabs + head + " {\n" + body + tabs + "}"
}

func declarations(props []css.Property, depth int) string {
	var b strings.Builder
	for _, p := range props {
		b.WriteString(renderDeclarations(p.Name(), p.Value(), p.Prefixes(), depth))
	}
	return b.String()
}

func renderDeclarations(name, value string, prefixes []string, depth int) string {
	tabs := strings.Repeat("\t", depth)
	var b strings.Builder
	for _, prefix := range prefixes {
		b.WriteString(tabs + prefix + name + ": " + value + ";\n")
	}
	b.WriteString(tabs + name + ": " + value + ";\n")
	return b.String()
}

func renderDeclarationsOf(descs []css.Descriptor, depth int) string {
	var b strings.Builder
	for _, d := range descs {
		b.WriteString(renderDeclarations(d.Name(), d.Value(), nil, depth))
	}
	return b.String()
}

func indent(text string, depth int) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = strings.Repeat("\t", depth) + line
		}
	}
	return strings.Join(lines, "\n")
}
