package doc

import (
	"go/scanner"
	"go/token"

	"github.com/Toms223/WebGo/docs/theme"
	"github.com/Toms223/WebGo/style"
	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type piece struct {
	start, end int
	tok        token.Token
	text       string
}

var builtinTypes = map[string]bool{
	"any": true, "bool": true, "byte": true, "error": true, "float64": true, "int": true, "rune": true, "string": true, "uint8": true,
}

func Highlight(source string) []app.UI {
	pieces := tokens(source)
	var out []app.UI
	last := 0
	for i, p := range pieces {
		if p.start > last {
			out = append(out, app.Text(source[last:p.start]))
		}
		if class := classify(pieces, i); class != nil {
			out = append(out, app.Span().Class(class.Name()).Text(p.text))
		} else {
			out = append(out, app.Text(p.text))
		}
		last = p.end
	}
	if last < len(source) {
		out = append(out, app.Text(source[last:]))
	}
	return out
}

func tokens(source string) []piece {
	fset := token.NewFileSet()
	file := fset.AddFile("", -1, len(source))
	var s scanner.Scanner
	s.Init(file, []byte(source), nil, scanner.ScanComments)
	var pieces []piece
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}
		text := lit
		if text == "" {
			text = tok.String()
		}
		start := file.Offset(pos)
		pieces = append(pieces, piece{start: start, end: start + len(text), tok: tok, text: text})
	}
	return pieces
}

func classify(pieces []piece, i int) *style.Rule {
	p := pieces[i]
	switch {
	case p.tok.IsKeyword():
		return theme.Keyword
	case p.tok == token.STRING || p.tok == token.CHAR:
		return theme.Str
	case p.tok == token.COMMENT:
		return theme.Dim
	case p.tok != token.IDENT:
		return nil
	case i+1 < len(pieces) && pieces[i+1].tok == token.LPAREN:
		return theme.Fn
	case builtinTypes[p.text] && (i == 0 || pieces[i-1].tok != token.PERIOD):
		return theme.Typ
	case i > 0 && isTypePrefix(pieces[i-1].tok):
		return theme.Typ
	}
	return nil
}

func isTypePrefix(tok token.Token) bool {
	switch tok {
	case token.TYPE, token.MUL, token.AND, token.RBRACK, token.LBRACK, token.CHAN:
		return true
	}
	return false
}
