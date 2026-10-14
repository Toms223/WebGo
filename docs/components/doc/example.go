package doc

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
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
		tools = append(tools, app.Div().Class(theme.ExampleTabs.Name()).Body(
			p.tab("Go", false),
			p.tab("CSS", true),
		))
	}
	tools = append(tools, &Copy{Text: shown})

	codeClasses := []string{theme.ExampleCode.Name()}
	if long && !p.expanded {
		codeClasses = append(codeClasses, theme.ExampleCodeClamped.Name())
	}
	children := []app.UI{
		app.Div().Class(theme.ExampleBar.Name()).Body(
			app.Span().Class(theme.ExampleName.Name()).Text(p.name()),
			app.Div().Class(theme.ExampleTools.Name()).Body(tools...),
		),
		app.Pre().Class(codeClasses...).Body(app.Code().Body(body...)),
	}
	if long {
		children = append(children, p.more(lines))
	}
	if e.Live != nil {
		children = append(children,
			app.Div().Class(theme.ExampleLabel.Name()).Text("Result"),
			app.Div().Class(theme.ExampleLive.Name()).Body(e.Live()),
		)
	}
	return app.Div().Class(theme.Example.Name()).Body(children...)
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
	return app.Button().Type("button").Class(theme.ExampleMore.Name()).Aria("expanded", strconv.FormatBool(p.expanded)).Text(label).
		OnClick(func(ctx app.Context, e app.Event) { p.expanded = !p.expanded })
}

func (p *Panel) tab(label string, css bool) app.UI {
	classes := []string{theme.ExampleTab.Name()}
	if p.showCSS == css {
		classes = append(classes, theme.ExampleTabOn.Name())
	}
	return app.Button().Type("button").Class(classes...).Aria("pressed", strconv.FormatBool(p.showCSS == css)).Text(label).
		OnClick(func(ctx app.Context, e app.Event) { p.showCSS = css })
}
