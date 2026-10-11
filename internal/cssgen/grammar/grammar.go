package grammar

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

type Kind int

const (
	Alt Kind = iota
	AnyOrder
	AllOrder
	Seq
	Group
	Keyword
	Type
	PropRef
	Func
)

type Node struct {
	Kind     Kind
	Text     string
	Children []*Node
	Min, Max int
	Comma    bool
}

type Shape struct {
	Kinds, Keywords []string
	Min, Max        int
	Comma           bool
}

var BasicKinds = map[string][]string{
	"length": {"Length"}, "percentage": {"Percentage"}, "length-percentage": {"Length", "Percentage"},
	"number": {"Number"}, "integer": {"Integer"}, "angle": {"Angle"}, "time": {"Time"}, "resolution": {"Resolution"},
	"flex": {"Fraction"}, "ratio": {"RatioValue"}, "string": {"StringValue"}, "url": {"UrlValue"},
	"custom-ident": {"IdentValue"}, "dashed-ident": {"IdentValue"}, "ident": {"IdentValue"}, "color": {"ColorLiteral"},
}

func Parse(src string) (*Node, error) {
	toks, err := tokenize(src)
	if err != nil {
		return nil, err
	}
	p := &parser{src: src, toks: toks}
	n, err := p.alt()
	if err == nil && p.at < len(p.toks) {
		err = p.fail("the end of the grammar")
	}
	if err != nil {
		return nil, err
	}
	return n, nil
}

func MergeAlternatives(values []string) (*Node, error) {
	var alternatives []*Node
	for _, v := range values {
		n, err := Parse(v)
		if err != nil {
			return nil, err
		}
		if n.Kind == Alt {
			alternatives = append(alternatives, n.Children...)
		} else {
			alternatives = append(alternatives, n)
		}
	}
	if len(alternatives) == 0 {
		return nil, errors.New("grammar: no values to merge")
	}
	return wrap(Alt, alternatives), nil
}

type parser struct {
	src  string
	toks []token
	at   int
}

var leaves = map[tokenKind]Kind{tokKeyword: Keyword, tokLiteral: Keyword, tokType: Type, tokProp: PropRef}

func (p *parser) alt() (*Node, error)      { return p.join(Alt, tokBar, p.anyOrder) }
func (p *parser) anyOrder() (*Node, error) { return p.join(AnyOrder, tokBars, p.allOrder) }
func (p *parser) allOrder() (*Node, error) { return p.join(AllOrder, tokAmps, p.seq) }

func (p *parser) join(kind Kind, sep tokenKind, operand func() (*Node, error)) (*Node, error) {
	var children []*Node
	for {
		n, err := operand()
		if err != nil {
			return nil, err
		}
		children = append(children, n)
		if !p.peek(sep) {
			return wrap(kind, children), nil
		}
		p.at++
	}
}

func (p *parser) seq() (*Node, error) {
	var children []*Node
	for p.startsComponent() {
		n, err := p.component()
		if err != nil {
			return nil, err
		}
		children = append(children, n)
	}
	if len(children) == 0 {
		return nil, p.fail("a component")
	}
	return wrap(Seq, children), nil
}

func (p *parser) startsComponent() bool {
	if p.at >= len(p.toks) {
		return false
	}
	k := p.toks[p.at].kind
	_, leaf := leaves[k]
	return leaf || k == tokOpen || k == tokFunc
}

func (p *parser) component() (*Node, error) {
	n, err := p.primary()
	if err != nil {
		return nil, err
	}
	for stacked := false; p.peek(tokMult); stacked = true {
		t := p.toks[p.at]
		p.at++
		if stacked {
			n = &Node{Kind: Group, Children: []*Node{n}}
		}
		n.Min, n.Max, n.Comma = t.min, t.max, t.comma
	}
	return n, nil
}

func (p *parser) primary() (*Node, error) {
	t := p.toks[p.at]
	p.at++
	switch t.kind {
	case tokOpen:
		inner, err := p.enclosed(tokClose, "]")
		if err != nil {
			return nil, err
		}
		if p.peek(tokBang) {
			p.at++
		}
		return &Node{Kind: Group, Children: []*Node{inner}, Min: 1, Max: 1}, nil
	case tokFunc:
		n := &Node{Kind: Func, Text: t.text, Min: 1, Max: 1}
		if p.peek(tokCloseParen) {
			p.at++
			return n, nil
		}
		inner, err := p.enclosed(tokCloseParen, ")")
		if err != nil {
			return nil, err
		}
		n.Children = []*Node{inner}
		return n, nil
	}
	return &Node{Kind: leaves[t.kind], Text: t.text, Min: 1, Max: 1}, nil
}

func (p *parser) enclosed(end tokenKind, want string) (*Node, error) {
	inner, err := p.alt()
	if err != nil {
		return nil, err
	}
	if !p.peek(end) {
		return nil, p.fail(want)
	}
	p.at++
	return inner, nil
}

func (p *parser) peek(k tokenKind) bool {
	return p.at < len(p.toks) && p.toks[p.at].kind == k
}

func (p *parser) fail(want string) error {
	if p.at < len(p.toks) {
		t := p.toks[p.at]
		return fmt.Errorf("grammar %q: expected %s, found %q at offset %d", p.src, want, t.raw, t.pos)
	}
	return fmt.Errorf("grammar %q: expected %s at the end", p.src, want)
}

func wrap(kind Kind, children []*Node) *Node {
	if len(children) == 1 {
		return children[0]
	}
	return &Node{Kind: kind, Children: children, Min: 1, Max: 1}
}

const opaque Kind = -1

const maxDepth = 8

type multiplicity struct {
	min, max int
	comma    bool
}

var (
	once     = multiplicity{min: 1, max: 1}
	optional = multiplicity{min: 0, max: 1}
)

func multiplicityOf(n *Node) multiplicity {
	return multiplicity{min: n.Min, max: n.Max, comma: n.Comma}
}

func Classify(n *Node, props, types func(name string) (*Node, bool)) (Shape, bool) {
	root := expander{props: props, types: types}.expand(n, nil)
	alternatives := []*Node{root}
	if root.Kind == Alt && multiplicityOf(root) == once {
		alternatives = root.Children
	}
	var c collector
	lo, hi, multiplied, comma := 0, 0, 0, false
	for i, a := range alternatives {
		aLo, aHi, m, ok := c.alternative(a)
		if !ok {
			return Shape{}, false
		}
		if m != once {
			multiplied++
			comma = m.comma
		}
		if i == 0 || aLo < lo {
			lo = aLo
		}
		if hi != -1 && (aHi == -1 || aHi > hi) {
			hi = aHi
		}
	}
	if multiplied > 1 || lo < 1 || len(c.kinds)+len(c.keywords) == 0 {
		return Shape{}, false
	}
	return Shape{Kinds: sorted(c.kinds), Keywords: sorted(c.keywords), Min: lo, Max: hi, Comma: comma}, true
}

func Keywords(n *Node, props, types func(name string) (*Node, bool)) []string {
	var words []string
	var walk func(*Node)
	walk = func(n *Node) {
		if n.Kind == Keyword && isIdent(n.Text) {
			words = append(words, n.Text)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(expander{props: props, types: types}.expand(n, nil))
	return sorted(words)
}

type expander struct {
	props, types func(name string) (*Node, bool)
}

func (e expander) expand(n *Node, path []string) *Node {
	if n == nil {
		return &Node{Kind: Keyword, Min: 1, Max: 1}
	}
	switch n.Kind {
	case Type:
		if _, basic := BasicKinds[n.Text]; !basic {
			if strings.HasSuffix(n.Text, "()") {
				return copyAs(n, opaque)
			}
			return e.reference(n, "<"+n.Text+">", e.types, path)
		}
	case PropRef:
		return e.reference(n, "<'"+n.Text+"'>", e.props, path)
	case Func:
		return copyAs(n, opaque)
	}
	out := copyAs(n, n.Kind)
	for _, c := range n.Children {
		out.Children = append(out.Children, e.expand(c, path))
	}
	return normalize(out)
}

func (e expander) reference(n *Node, key string, resolve func(string) (*Node, bool), path []string) *Node {
	if slices.Contains(path, key) || len(path) >= maxDepth {
		return copyAs(n, opaque)
	}
	if resolve != nil {
		if target, ok := resolve(n.Text); ok && target != nil {
			return withMultiplier(e.expand(target, append(slices.Clip(path), key)), n)
		}
	}
	if n.Kind == Type {
		return copyAs(n, opaque)
	}
	return copyAs(n, PropRef)
}

func copyAs(n *Node, kind Kind) *Node {
	return &Node{Kind: kind, Text: n.Text, Min: n.Min, Max: n.Max, Comma: n.Comma}
}

func withMultiplier(n, from *Node) *Node {
	m := multiplicityOf(from)
	switch {
	case m == once:
		return n
	case multiplicityOf(n) == once:
		n.Min, n.Max, n.Comma = m.min, m.max, m.comma
		return n
	}
	return &Node{Kind: Group, Children: []*Node{n}, Min: m.min, Max: m.max, Comma: m.comma}
}

func normalize(n *Node) *Node {
	switch {
	case n.Kind == Group && len(n.Children) == 1:
		return withMultiplier(n.Children[0], n)
	case n.Kind == Alt:
		var flat []*Node
		for _, c := range n.Children {
			if c.Kind == Alt && multiplicityOf(c) == once {
				flat = append(flat, c.Children...)
			} else {
				flat = append(flat, c)
			}
		}
		n.Children = flat
	}
	return n
}

type collector struct {
	kinds, keywords []string
}

func (c *collector) alternative(a *Node) (lo, hi int, m multiplicity, ok bool) {
	m = multiplicityOf(a)
	if atomic(a) || atomicAlt(a) {
		c.add(a)
		return m.min, m.max, m, true
	}
	if (a.Kind != AnyOrder && a.Kind != AllOrder && a.Kind != Seq) || m != once {
		return 0, 0, m, false
	}
	required := 0
	for _, member := range a.Children {
		mm := multiplicityOf(member)
		if !(atomic(member) || atomicAlt(member)) || (mm != once && mm != optional) {
			return 0, 0, m, false
		}
		if mm == once {
			required++
		}
		c.add(member)
	}
	lo = 1
	if a.Kind == Seq {
		lo = required
	}
	return lo, len(a.Children), once, true
}

func (c *collector) add(n *Node) {
	switch n.Kind {
	case Alt:
		for _, child := range n.Children {
			c.add(child)
		}
	case Keyword:
		c.keywords = append(c.keywords, n.Text)
	case Type:
		c.kinds = append(c.kinds, BasicKinds[n.Text]...)
	}
}

func atomic(n *Node) bool {
	_, basic := BasicKinds[n.Text]
	return n.Kind == opaque || (n.Kind == Keyword && isIdent(n.Text)) || (n.Kind == Type && basic)
}

func atomicAlt(n *Node) bool {
	if n.Kind != Alt || len(n.Children) == 0 {
		return false
	}
	for _, c := range n.Children {
		if !atomic(c) || multiplicityOf(c) != once {
			return false
		}
	}
	return true
}

func sorted(s []string) []string {
	slices.Sort(s)
	return slices.Compact(s)
}
