package css

import "strings"

var stringEscaper = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	"\n", `\a `,
	"\r", `\d `,
	"\f", `\c `,
)

func quote(s string) string {
	return `"` + stringEscaper.Replace(s) + `"`
}

type StringValue struct{ text string }

func String(s string) StringValue { return StringValue{text: quote(s)} }

func (s StringValue) CSS() string { return s.text }

type UrlValue struct{ text string }

func Url(s string) UrlValue { return UrlValue{text: "url(" + quote(s) + ")"} }

func (u UrlValue) CSS() string { return u.text }

type IdentValue struct{ text string }

func Ident(s string) IdentValue { return IdentValue{text: s} }

func (i IdentValue) CSS() string { return i.text }

type RawValue struct{ text string }

func Raw(s string) RawValue { return RawValue{text: s} }

func (r RawValue) CSS() string { return r.text }
