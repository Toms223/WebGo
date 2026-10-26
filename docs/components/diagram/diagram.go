package diagram

import (
	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/theme"
)

type Fill int

const (
	Plain Fill = iota
	Sun
	Lime
	Lilac
	Coral
)

type Part struct {
	ui  app.UI
	via string
}

type Way struct {
	label string
	parts []Part
}

func Figure(caption string, body ...app.UI) app.UI {
	return app.Figure().Style(theme.Diagram).Body(append(body, app.FigCaption().Style(theme.DiagramCaption).Text(caption))...)
}

func Lane(tag, title string, body ...app.UI) app.UI {
	head := []app.UI{}
	if tag != "" {
		head = append(head, app.Span().Style(theme.DiagramTag).Text(tag))
	}
	if title != "" {
		head = append(head, app.P().Style(theme.DiagramTitle).Text(title))
	}
	return app.Div().Style(theme.DiagramLane).Body(append(head, body...)...)
}

func Box(fill Fill, name, sub string) Part {
	body := []app.UI{app.Strong().Style(theme.DiagramName).Text(name)}
	if sub != "" {
		body = append(body, app.Span().Style(theme.DiagramSub).Text(sub))
	}
	return Part{ui: app.Div().Style(boxStyle(fill)).Body(body...)}
}

func Via(label string) Part {
	return Part{via: label}
}

func Flow(parts ...Part) app.UI {
	return app.Div().Style(theme.DiagramFlow).Body(joined(parts, theme.DiagramGlyph)...)
}

func Stack(parts ...Part) app.UI {
	return app.Div().Style(theme.DiagramStack).Body(joined(parts, theme.DiagramDown)...)
}

func When(label string, parts ...Part) Way {
	return Way{label: label, parts: parts}
}

func Branch(question Part, ways ...Way) app.UI {
	rows := make([]app.UI, 0, len(ways))
	for _, w := range ways {
		row := joined(w.parts, theme.DiagramGlyph)
		if len(row) > 0 {
			row[0] = step(arrow(w.label, theme.DiagramGlyph), row[0])
		}
		rows = append(rows, app.Div().Style(theme.DiagramWay).Body(row...))
	}
	return app.Div().Style(theme.DiagramBranch).Body(
		question.ui,
		app.Div().Style(theme.DiagramWays).Body(rows...),
	)
}

func joined(parts []Part, glyph ...app.ClassRule) []app.UI {
	var out []app.UI
	via := ""
	for _, p := range parts {
		if p.ui == nil {
			via = p.via
			continue
		}
		if len(out) > 0 {
			out = append(out, step(arrow(via, glyph...), p.ui))
		} else {
			out = append(out, p.ui)
		}
		via = ""
	}
	return out
}

func step(arrow, box app.UI) app.UI {
	return app.Span().Style(theme.DiagramStep).Body(arrow, box)
}

func arrow(label string, glyph ...app.ClassRule) app.UI {
	body := []app.UI{}
	if label != "" {
		body = append(body, app.Span().Style(theme.DiagramVia).Text(label))
	}
	body = append(body, app.Span().Style(glyph...).Aria("hidden", "true").Text("→"))
	return app.Span().Style(theme.DiagramArrow).Body(body...)
}

func boxStyle(fill Fill) app.ClassRule {
	switch fill {
	case Sun:
		return theme.DiagramSun
	case Lime:
		return theme.DiagramLime
	case Lilac:
		return theme.DiagramLilac
	case Coral:
		return theme.DiagramCoral
	}
	return theme.DiagramBox
}
