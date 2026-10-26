package doc

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Toms223/WebGo/app"
	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/docs/theme"
)

const clampLines = 18

type Panel struct {
	app.Compo
	Example  example.Example
	showCSS  bool
	expanded bool
}

func Example(e example.Example) app.UI {
	return &Panel{Example: e}
}

func (p *Panel) Render() app.UI {
	e := p.Example
	shown, body := p.shown()
	lines := strings.Count(shown, "\n") + 1
	long := lines > clampLines

	tools := []app.UI{}
	if e.CSS != "" {
		tools = append(tools, app.Div().Style(theme.ExampleTabs).Body(
			p.tab("Go", false),
			p.tab("CSS", true),
		))
	}
	tools = append(tools, &Copy{Text: shown})

	codeClasses := []app.ClassRule{theme.ExampleCode}
	if long && !p.expanded {
		codeClasses = append(codeClasses, theme.ExampleCodeClamped)
	}
	children := []app.UI{
		app.Div().Style(theme.ExampleBar).Body(
			app.Span().Style(theme.ExampleName).Text(p.name()),
			app.Div().Style(theme.ExampleTools).Body(tools...),
		),
		app.Pre().Style(codeClasses...).Body(app.Code().Body(body...)),
	}
	if long {
		children = append(children, p.more(lines))
	}
	if e.Live != nil {
		children = append(children,
			app.Div().Style(theme.ExampleLabel).Text("Result"),
			app.Div().Style(theme.ExampleLive).Body(e.Live()),
		)
	}
	return app.Div().Style(theme.Example).Body(children...)
}

func (p *Panel) name() string {
	if p.showCSS {
		return strings.TrimSuffix(p.Example.Name, filepath.Ext(p.Example.Name)) + ".css"
	}
	return p.Example.Name
}

func (p *Panel) shown() (string, []app.UI) {
	if p.showCSS {
		return p.Example.CSS, HighlightCSS(p.Example.CSS)
	}
	if p.Example.Lang != "" && p.Example.Lang != "go" {
		return p.Example.Source, []app.UI{app.Text(p.Example.Source)}
	}
	return p.Example.Source, Highlight(p.Example.Source)
}

func (p *Panel) more(lines int) app.UI {
	label := "Show all " + strconv.Itoa(lines) + " lines"
	if p.expanded {
		label = "Show less"
	}
	return app.Button().Type("button").Style(theme.ExampleMore).Aria("expanded", strconv.FormatBool(p.expanded)).Text(label).
		OnClick(func(ctx app.Context, e app.Event) { p.expanded = !p.expanded })
}

func (p *Panel) tab(label string, css bool) app.UI {
	classes := []app.ClassRule{theme.ExampleTab}
	if p.showCSS == css {
		classes = append(classes, theme.ExampleTabOn)
	}
	return app.Button().Type("button").Style(classes...).Aria("pressed", strconv.FormatBool(p.showCSS == css)).Text(label).
		OnClick(func(ctx app.Context, e app.Event) { p.showCSS = css })
}
