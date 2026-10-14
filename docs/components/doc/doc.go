package doc

import (
	"strings"

	"github.com/Toms223/WebGo/docs/components/nav"
	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

func P(parts ...any) app.UI {
	return app.P().Body(Inline(parts...)...)
}

func H(title string) app.UI {
	return app.H3().Class(theme.DocSubheading.Name()).Text(title)
}

func B(text string) app.UI {
	return app.Strong().Text(text)
}

func C(text string) app.UI {
	return app.Code().Text(text)
}

func A(href, text string) app.UI {
	if strings.HasPrefix(href, "/") {
		return app.A().Href(nav.Href(href)).Text(text)
	}
	if strings.HasPrefix(href, "#") {
		return app.A().Href(href).Text(text)
	}
	return app.A().Href(href).Rel("noopener noreferrer").Target("_blank").Text(text)
}

func List(items ...app.UI) app.UI {
	return app.Ul().Body(items...)
}

func Ordered(items ...app.UI) app.UI {
	return app.Ol().Body(items...)
}

func Item(parts ...any) app.UI {
	return app.Li().Body(Inline(parts...)...)
}

func Note(parts ...any) app.UI {
	return app.Div().Class(theme.Note.Name()).Body(app.P().Body(Inline(parts...)...))
}

func Rule() app.UI {
	return app.Hr()
}

func Go(source string) app.UI {
	return Code(theme.LangGo, Highlight(source)...)
}

func CSS(source string) app.UI {
	return Code(theme.LangCSS, app.Text(source))
}

func Sh(source string) app.UI {
	return Code(theme.LangShell, app.Text(source))
}

func Markdown(source string) app.UI {
	return Code(theme.LangMarkdown, app.Text(source))
}

func HTML(source string) app.UI {
	return Code(theme.LangHTML, app.Text(source))
}

func Plain(source string) app.UI {
	return app.Pre().Body(app.Code().Text(source))
}

func Code(lang *style.Rule, body ...app.UI) app.UI {
	return app.Pre().Class(lang.Name()).Body(app.Code().Body(body...))
}

func Inline(parts ...any) []app.UI {
	var out []app.UI
	for _, part := range parts {
		switch v := part.(type) {
		case string:
			out = append(out, spans(v)...)
		case app.UI:
			out = append(out, v)
		case []app.UI:
			out = append(out, v...)
		}
	}
	return out
}

func spans(text string) []app.UI {
	var out []app.UI
	for i, piece := range strings.Split(text, "`") {
		if piece == "" {
			continue
		}
		if i%2 == 1 {
			out = append(out, app.Code().Text(piece))
			continue
		}
		out = append(out, app.Text(piece))
	}
	return out
}
