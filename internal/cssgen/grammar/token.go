package grammar

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

type tokenKind int

const (
	tokBar tokenKind = iota
	tokBars
	tokAmps
	tokOpen
	tokClose
	tokBang
	tokCloseParen
	tokMult
	tokLiteral
	tokKeyword
	tokType
	tokProp
	tokFunc
)

type token struct {
	kind     tokenKind
	text     string
	raw      string
	pos      int
	min, max int
	comma    bool
}

var punctuation = map[byte]token{
	'|': {kind: tokBar},
	'[': {kind: tokOpen},
	']': {kind: tokClose},
	'!': {kind: tokBang},
	')': {kind: tokCloseParen},
	',': {kind: tokLiteral, text: ","},
	'/': {kind: tokLiteral, text: "/"},
	'?': {kind: tokMult, min: 0, max: 1},
	'*': {kind: tokMult, min: 0, max: -1},
	'+': {kind: tokMult, min: 1, max: -1},
	'#': {kind: tokMult, min: 1, max: -1, comma: true},
}

func tokenize(src string) ([]token, error) {
	var toks []token
	for i := 0; i < len(src); {
		rest := src[i:]
		t, n := token{}, 1
		var err error
		switch c := rest[0]; {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
			continue
		case c == '[' && len(toks) > 0 && toks[len(toks)-1].kind == tokType && rangeLen(rest) > 0:
			i += rangeLen(rest)
			continue
		case strings.HasPrefix(rest, "||"):
			t, n = token{kind: tokBars}, 2
		case strings.HasPrefix(rest, "&&"):
			t, n = token{kind: tokAmps}, 2
		case c == '{' || strings.HasPrefix(rest, "#{"):
			t, n, err = braces(rest)
		case c == '<':
			t, n, err = angle(rest)
		case identChar(c):
			t, n = word(rest)
		default:
			var ok bool
			if t, ok = punctuation[c]; !ok {
				r, _ := utf8.DecodeRuneInString(rest)
				err = fmt.Errorf("unexpected %q", r)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("grammar %q: %v at offset %d", src, err, i)
		}
		t.raw, t.pos = rest[:n], i
		toks = append(toks, t)
		i += n
	}
	return toks, nil
}

func braces(s string) (token, int, error) {
	t := token{kind: tokMult, comma: s[0] == '#'}
	end := strings.IndexByte(s, '}')
	if end < 0 {
		return t, 0, errors.New("unterminated multiplier")
	}
	from, to, ranged := strings.Cut(s[strings.IndexByte(s, '{')+1:end], ",")
	lo, okLo := count(from)
	hi, okHi := lo, true
	if ranged {
		hi = -1
		if strings.TrimSpace(to) != "" {
			hi, okHi = count(to)
		}
	}
	if !okLo || !okHi || (hi >= 0 && hi < lo) {
		return t, 0, fmt.Errorf("bad multiplier %q", s[:end+1])
	}
	t.min, t.max = lo, hi
	return t, end + 1, nil
}

func count(s string) (int, bool) {
	s = strings.TrimSpace(s)
	if s == "" || strings.Trim(s, "0123456789") != "" {
		return 0, false
	}
	n, err := strconv.Atoi(s)
	return n, err == nil
}

func angle(s string) (token, int, error) {
	end := strings.IndexByte(s, '>')
	if end < 0 {
		return token{}, 0, errors.New("unterminated <")
	}
	body := strings.TrimSpace(s[1:end])
	if quoted, ok := strings.CutPrefix(body, "'"); ok {
		name, ok := strings.CutSuffix(quoted, "'")
		if !ok || !isIdent(name) {
			return token{}, 0, fmt.Errorf("bad property reference %q", s[:end+1])
		}
		return token{kind: tokProp, text: name}, end + 1, nil
	}
	name, annotation := body, ""
	if i := strings.IndexAny(body, " ["); i >= 0 {
		name, annotation = body[:i], strings.TrimSpace(body[i:])
	}
	if !isIdent(strings.TrimSuffix(name, "()")) || (annotation != "" && rangeLen(annotation) != len(annotation)) {
		return token{}, 0, fmt.Errorf("bad type %q", s[:end+1])
	}
	return token{kind: tokType, text: name}, end + 1, nil
}

func rangeLen(s string) int {
	end := strings.IndexByte(s, ']')
	if !strings.HasPrefix(s, "[") || end < 0 {
		return 0
	}
	lo, hi, ok := strings.Cut(s[1:end], ",")
	if !ok || !isBound(lo) || !isBound(hi) {
		return 0
	}
	return end + 1
}

func isBound(s string) bool {
	s = strings.TrimPrefix(strings.TrimSpace(s), "-")
	if s == "∞" {
		return true
	}
	unit := strings.TrimLeft(s, "0123456789.")
	return unit != s && (unit == "%" || isLetters(unit))
}

func isLetters(s string) bool {
	for i := 0; i < len(s); i++ {
		if c := s[i]; (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}
	return true
}

func word(s string) (token, int) {
	n := 0
	for n < len(s) && identChar(s[n]) {
		n++
	}
	if n < len(s) && s[n] == '(' {
		return token{kind: tokFunc, text: s[:n]}, n + 1
	}
	return token{kind: tokKeyword, text: s[:n]}, n
}

func identChar(c byte) bool {
	return c == '-' || c == '_' || '0' <= c && c <= '9' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isIdent(s string) bool {
	for i := 0; i < len(s); i++ {
		if !identChar(s[i]) {
			return false
		}
	}
	return s != ""
}
