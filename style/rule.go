package style

import (
	"cmp"
	"slices"
	"strings"

	"github.com/Toms223/WebGo/container"
	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/supports"

	"github.com/Toms223/WebGo/app"
)

// Rule is a selector with blocks of properties, written as the base block, the pseudo blocks, then the
// Container, Supports and Media blocks. Builder methods change the rule and return it, latching the
// first error for Err instead of panicking; combinators return a new rule.
type Rule struct {
	sel        selector
	blocks     []*block
	bindings   []binding
	initially  bool
	children   []*Rule
	owner      *Rule
	self       bool
	lastPseudo string
	err        error
}

type blockKind int

const (
	baseKind blockKind = iota
	pseudoKind
	containerKind
	supportsKind
	mediaKind
)

type block struct {
	kind   blockKind
	pseudo string
	cond   string
	under  string
	props  []css.Property
}

type query interface {
	CSS() string
	Err() error
}

// Self returns a rule for the parent selector & inside Nest: Self().Hover(...) writes &:hover. Used
// outside Nest, or given Set, it is ErrBadSelector.
func Self() *Rule { return &Rule{self: true} }

// Set adds properties to the base block, written under the bare selector. Nil properties are skipped.
func (r *Rule) Set(props ...css.Property) *Rule {
	return r.extend(block{}, props)
}

// Media adds properties to a block for the rule's selector wrapped in @media q. A query carrying an
// error latches ErrBadMediaQuery and leaves the block out. On a Self rule, the block attaches under
// the pseudo method called last.
func (r *Rule) Media(q media.Query, props ...css.Property) *Rule {
	return r.conditional(mediaKind, q, ErrBadMediaQuery, props)
}

// Container is Media for @container; a bad query latches ErrBadContainerQuery.
func (r *Rule) Container(q container.Query, props ...css.Property) *Rule {
	return r.conditional(containerKind, q, ErrBadContainerQuery, props)
}

// Supports is Media for @supports; a bad query latches ErrBadSupportsQuery.
func (r *Rule) Supports(q supports.Query, props ...css.Property) *Rule {
	return r.conditional(supportsKind, q, ErrBadSupportsQuery, props)
}

func (r *Rule) conditional(kind blockKind, q query, bad error, props []css.Property) *Rule {
	if q.Err() != nil {
		r.err = cmp.Or(r.err, bad)
		return r
	}
	key := block{kind: kind, cond: q.CSS()}
	if r.self {
		key.under = r.lastPseudo
	}
	return r.extend(key, props)
}

// Nest writes children inside r with CSS nesting and claims them. Passing a claimed child to another
// Nest, a combinator, Generated or Dynamic is ErrNested, and a child carrying a binding is
// ErrNestedBinding.
func (r *Rule) Nest(children ...*Rule) *Rule {
	for _, c := range children {
		switch {
		case c == nil:
			r.err = cmp.Or(r.err, ErrNoName)
		case c.owner != nil || r.within(c):
			r.err = cmp.Or(r.err, ErrNested)
		default:
			c.owner = r
			r.children = append(r.children, c)
		}
	}
	return r
}

// On binds r to the action named action: each time it fires, the engine inserts r when when returns
// true and removes it otherwise. Bind is the typed form.
func (r *Rule) On(action string, when func(app.Action) bool) *Rule {
	r.bindings, r.err = bind(r.bindings, r.err, binding{action: action, when: when}, true)
	return r
}

// Bind attaches a Binding from When or Compute; one from ComputeFrames is ErrBadBinding. For one
// action name, the last binding wins.
func (r *Rule) Bind(b Binding) *Rule {
	r.bindings, r.err = bind(r.bindings, r.err, binding{action: b.action, typed: &b}, b.frames == nil)
	return r
}

// CSSFor returns the CSS the engine would insert for r when the action named action carries v, or ""
// when r would be absent.
func (r *Rule) CSSFor(action string, v any) string { return cssFor(r, action, v) }

// Initially sets whether Run inserts the bound rule before any action fires. The default is absent.
func (r *Rule) Initially(present bool) *Rule {
	r.initially = present
	return r
}

// Err returns the first error latched on r or on any nested child.
func (r *Rule) Err() error {
	if err := r.ownErr(); err != nil {
		return err
	}
	for _, c := range r.children {
		if c.dynamic() {
			return ErrNestedBinding
		}
		if err := c.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *Rule) ownErr() error {
	switch {
	case r.self && len(r.base()) > 0:
		return ErrBadSelector
	case !r.self && len(r.sel.parts) == 0:
		return ErrNoName
	}
	return r.err
}

func (r *Rule) placementErr() error {
	switch {
	case r.owner != nil:
		return ErrNested
	case r.self:
		return ErrBadSelector
	}
	return nil
}

func (r *Rule) within(o *Rule) bool {
	for at := r; at != nil; at = at.owner {
		if at == o {
			return true
		}
	}
	return false
}

func (r *Rule) CSS() string { return render([]Item{r}) }

func (*Rule) item() {}

func (r *Rule) subject() string { return r.Selector() }

func (r *Rule) pseudo(name string, props []css.Property) *Rule {
	r.lastPseudo = name
	return r.extend(block{kind: pseudoKind, pseudo: name}, props)
}

func (r *Rule) blockTexts() []string { return r.textsWith(r.base()) }

func (r *Rule) textsWith(base []css.Property) []string {
	sel := r.Selector()
	if len(r.children) > 0 {
		if text := renderNested(sel, r, base, 0); text != "" {
			return []string{strings.TrimSuffix(text, "\n")}
		}
		return nil
	}
	var texts []string
	if len(base) > 0 {
		texts = append(texts, renderBlock(sel, base, 0))
	}
	for _, b := range r.blocks {
		if !b.base() && len(b.props) > 0 {
			texts = append(texts, b.text(sel))
		}
	}
	return texts
}

func (r *Rule) base() []css.Property {
	if b := r.find(block{}); b != nil {
		return b.props
	}
	return nil
}

func (r *Rule) dynamic() bool { return len(r.bindings) > 0 }

func (r *Rule) extend(key block, props []css.Property) *Rule {
	target := r.find(key)
	if target == nil {
		target = &key
		at := len(r.blocks)
		for at > 0 && r.blocks[at-1].rank() > key.rank() {
			at--
		}
		r.blocks = slices.Insert(r.blocks, at, target)
	}
	for _, p := range props {
		if p != nil {
			target.props = append(target.props, p)
		}
	}
	return r
}

func (r *Rule) find(key block) *block {
	for _, b := range r.blocks {
		if b.kind == key.kind && b.pseudo == key.pseudo && b.under == key.under && b.cond == key.cond {
			return b
		}
	}
	return nil
}

func (b *block) text(sel string) string {
	if b.conditional() {
		return renderConditional(b.prelude(), sel+b.under, b.props)
	}
	return renderBlock(sel+b.pseudo, b.props, 0)
}

func (r *Rule) lonePseudo() *block {
	var lone *block
	for _, b := range r.blocks {
		if len(b.props) == 0 {
			continue
		}
		if lone != nil || b.kind != pseudoKind {
			return nil
		}
		lone = b
	}
	if len(r.children) > 0 {
		return nil
	}
	return lone
}

func (b *block) base() bool { return b.kind == baseKind }

func (b *block) conditional() bool { return b.kind >= containerKind }

func (b *block) prelude() string {
	switch b.kind {
	case containerKind:
		return "@container " + b.cond
	case supportsKind:
		return "@supports " + b.cond
	case mediaKind:
		return "@media " + b.cond
	}
	return ""
}

func (b *block) rank() int { return int(b.kind) }
