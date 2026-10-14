package doc

import (
	"strconv"

	"github.com/Toms223/WebGo/docs/components/example"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type Panel struct {
	app.Compo
	Example example.Example
	showCSS bool
}

func Example(e example.Example) app.UI {
	return &Panel{Example: e}
}

func (p *Panel) Render() app.UI {
	e := p.Example
	bar := []app.UI{app.Span().Class(theme.ExampleName.Name()).Text(e.Name)}
	if e.CSS != "" {
		bar = append(bar, app.Div().Class(theme.ExampleTabs.Name()).Body(
			p.tab("Go", false),
			p.tab("CSS", true),
		))
	}
	children := []app.UI{
		app.Div().Class(theme.ExampleBar.Name()).Body(bar...),
		app.Pre().Class(theme.ExampleCode.Name()).Body(app.Code().Body(p.source()...)),
	}
	if e.Live != nil {
		children = append(children,
			app.Div().Class(theme.ExampleLabel.Name()).Text("Result"),
			app.Div().Class(theme.ExampleLive.Name()).Body(e.Live()),
		)
	}
	return app.Div().Class(theme.Example.Name()).Body(children...)
}

func (p *Panel) source() []app.UI {
	if p.showCSS {
		return []app.UI{app.Text(p.Example.CSS)}
	}
	if p.Example.Lang != "" && p.Example.Lang != "go" {
		return []app.UI{app.Text(p.Example.Source)}
	}
	return Highlight(p.Example.Source)
}

func (p *Panel) tab(label string, css bool) app.UI {
	classes := []string{theme.ExampleTab.Name()}
	if p.showCSS == css {
		classes = append(classes, theme.ExampleTabOn.Name())
	}
	return app.Button().Type("button").Class(classes...).Aria("pressed", strconv.FormatBool(p.showCSS == css)).Text(label).
		OnClick(func(ctx app.Context, e app.Event) { p.showCSS = css })
}
