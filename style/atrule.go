package style

import (
	"cmp"
	"slices"
	"strconv"
	"strings"

	"github.com/Toms223/WebGo/css"
	"github.com/Toms223/WebGo/media"
	"github.com/Toms223/WebGo/supports"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type KeyframesRule struct {
	name      string
	frames    []Frame
	bindings  []binding
	initially bool
	err       error
}

type Frame struct {
	stop  string
	props []css.Property
}

func FrameAt(stop css.Percentage, props ...css.Property) Frame {
	if !withinStops(stop) {
		return Frame{}
	}
	return newFrame(stop.CSS(), props)
}

func FrameFrom(props ...css.Property) Frame { return newFrame("from", props) }

func FrameTo(props ...css.Property) Frame { return newFrame("to", props) }

func newFrame(stop string, props []css.Property) Frame {
	return Frame{stop: stop, props: withoutNilProps(props)}
}

func withinStops(stop css.Percentage) bool {
	n, err := strconv.ParseFloat(strings.TrimSuffix(stop.CSS(), "%"), 64)
	return err == nil && n >= 0 && n <= 100
}

func Keyframes(name string) *KeyframesRule { return &KeyframesRule{name: name} }

func (k *KeyframesRule) At(stop css.Percentage, props ...css.Property) *KeyframesRule {
	if !withinStops(stop) {
		k.err = cmp.Or(k.err, ErrBadAtRule)
	}
	return k.frame(FrameAt(stop, props...))
}

func (k *KeyframesRule) From(props ...css.Property) *KeyframesRule {
	return k.frame(FrameFrom(props...))
}

func (k *KeyframesRule) To(props ...css.Property) *KeyframesRule {
	return k.frame(FrameTo(props...))
}

func (k *KeyframesRule) Name() string { return k.name }

func (k *KeyframesRule) CSS() string { return render([]Item{k}) }

func (k *KeyframesRule) On(action string, when func(app.Action) bool) *KeyframesRule {
	k.bindings, k.err = bind(k.bindings, k.err, binding{action: action, when: when}, true)
	return k
}

func (k *KeyframesRule) Bind(b Binding) *KeyframesRule {
	k.bindings, k.err = bind(k.bindings, k.err, binding{action: b.action, typed: &b}, b.props == nil)
	return k
}

func (k *KeyframesRule) CSSFor(action string, v any) string { return cssFor(k, action, v) }

func (k *KeyframesRule) Initially(present bool) *KeyframesRule {
	k.initially = present
	return k
}

func (k *KeyframesRule) Err() error {
	if k.name == "" {
		return ErrNoName
	}
	return k.err
}

func (k *KeyframesRule) frame(f Frame) *KeyframesRule {
	if len(f.props) > 0 {
		k.frames = append(k.frames, f)
	}
	return k
}

func (k *KeyframesRule) blockTexts() []string { return k.textsWith(k.frames) }

func (k *KeyframesRule) textsWith(frames []Frame) []string {
	if len(frames) == 0 {
		return nil
	}
	texts := make([]string, len(frames))
	for i, f := range frames {
		texts[i] = renderBlock(f.stop, f.props, 1)
	}
	return []string{"@keyframes " + k.name + " {\n" + strings.Join(texts, "\n") + "\n}"}
}

func (k *KeyframesRule) subject() string { return atName("@keyframes", k.name) }

func (k *KeyframesRule) dynamic() bool { return len(k.bindings) > 0 }

type FontFaceRule struct{ descs []css.Descriptor }

func FontFace(descs ...css.Descriptor) *FontFaceRule { return (&FontFaceRule{}).Set(descs...) }

func (f *FontFaceRule) Set(descs ...css.Descriptor) *FontFaceRule {
	f.descs = appendDescriptors(f.descs, descs)
	return f
}

func (f *FontFaceRule) CSS() string { return render([]Item{f}) }

func (f *FontFaceRule) Err() error { return require(f.descs, "font-family", "src") }

func (f *FontFaceRule) blockTexts() []string { return descriptorBlock("@font-face", f.descs) }

func (*FontFaceRule) subject() string { return "@font-face" }

type PropertyRule struct {
	name  string
	descs []css.Descriptor
}

func Property(name string, descs ...css.Descriptor) *PropertyRule {
	if !strings.HasPrefix(name, "--") {
		name = "--" + name
	}
	return (&PropertyRule{name: name}).Set(descs...)
}

func (p *PropertyRule) Set(descs ...css.Descriptor) *PropertyRule {
	p.descs = appendDescriptors(p.descs, descs)
	return p
}

func (p *PropertyRule) Name() string { return p.name }

func (p *PropertyRule) CSS() string { return render([]Item{p}) }

func (p *PropertyRule) Err() error {
	if p.name == "--" {
		return ErrNoName
	}
	return require(p.descs, "syntax", "inherits")
}

func (p *PropertyRule) blockTexts() []string { return descriptorBlock("@property "+p.name, p.descs) }

func (p *PropertyRule) subject() string { return "@property " + p.name }

type ImportRule struct {
	target   css.Value
	layer    string
	supports string
	media    string
	err      error
}

type ImportOption func(*ImportRule)

func Import(target css.Value, opts ...ImportOption) *ImportRule {
	r := &ImportRule{target: target}
	switch target.(type) {
	case nil, css.UrlValue, css.StringValue:
	default:
		r.err = ErrBadAtRule
	}
	for _, opt := range opts {
		if opt != nil {
			opt(r)
		}
	}
	return r
}

func InLayer(name string) ImportOption {
	return func(r *ImportRule) {
		if name == "" {
			r.err = cmp.Or(r.err, ErrNoName)
			return
		}
		r.layer = name
	}
}

func IfSupports(q supports.Query) ImportOption {
	return func(r *ImportRule) {
		if q.Err() != nil {
			r.err = cmp.Or(r.err, ErrBadSupportsQuery)
			return
		}
		r.supports = q.CSS()
	}
}

func ForMedia(q media.Query) ImportOption {
	return func(r *ImportRule) {
		if q.Err() != nil {
			r.err = cmp.Or(r.err, ErrBadMediaQuery)
			return
		}
		r.media = q.CSS()
	}
}

func (r *ImportRule) CSS() string { return render([]Item{r}) }

func (r *ImportRule) Err() error {
	if r.target == nil {
		return ErrNoName
	}
	return r.err
}

func (r *ImportRule) blockTexts() []string {
	if r.target == nil {
		return nil
	}
	text := "@import " + r.target.CSS()
	if r.layer != "" {
		text += " layer(" + r.layer + ")"
	}
	if r.supports != "" {
		text += " supports(" + unwrapped(r.supports) + ")"
	}
	if r.media != "" {
		text += " " + r.media
	}
	return []string{text + ";"}
}

func (r *ImportRule) subject() string {
	if r.target == nil {
		return "@import"
	}
	return "@import " + r.target.CSS()
}

type LayerRule struct {
	name  string
	items []Item
}

func Layer(name string, items ...Item) *LayerRule {
	return &LayerRule{name: name, items: withoutNil(items)}
}

func (l *LayerRule) CSS() string { return render([]Item{l}) }

func (l *LayerRule) Err() error {
	if err := l.ownErr(); err != nil {
		return err
	}
	for _, i := range l.items {
		if err := i.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (l *LayerRule) ownErr() error {
	if l.name == "" {
		return ErrNoName
	}
	for _, i := range l.items {
		switch i.(type) {
		case *ImportRule, *LayerOrderRule:
			return ErrBadAtRule
		}
	}
	return nil
}

func (l *LayerRule) blockTexts() []string {
	var texts []string
	for _, i := range l.items {
		texts = append(texts, i.blockTexts()...)
	}
	if len(texts) == 0 {
		return nil
	}
	return []string{"@layer " + l.name + " {\n" + indent(strings.Join(texts, "\n\n"), 1) + "\n}"}
}

func (l *LayerRule) subject() string { return atName("@layer", l.name) }

type LayerOrderRule struct{ names []string }

func LayerOrder(names ...string) *LayerOrderRule {
	return &LayerOrderRule{names: slices.Clone(names)}
}

func (o *LayerOrderRule) CSS() string { return render([]Item{o}) }

func (o *LayerOrderRule) Err() error {
	if len(o.names) == 0 || slices.Contains(o.names, "") {
		return ErrNoName
	}
	return nil
}

func (o *LayerOrderRule) blockTexts() []string {
	if len(o.names) == 0 {
		return nil
	}
	return []string{"@layer " + strings.Join(o.names, ", ") + ";"}
}

func (o *LayerOrderRule) subject() string { return atName("@layer", strings.Join(o.names, ", ")) }

func (*KeyframesRule) item()  {}
func (*FontFaceRule) item()   {}
func (*PropertyRule) item()   {}
func (*ImportRule) item()     {}
func (*LayerRule) item()      {}
func (*LayerOrderRule) item() {}

func atName(keyword, name string) string {
	if name == "" {
		return keyword
	}
	return keyword + " " + name
}

func appendDescriptors(to, descs []css.Descriptor) []css.Descriptor {
	for _, d := range descs {
		if d != nil {
			to = append(to, d)
		}
	}
	return to
}

func require(descs []css.Descriptor, names ...string) error {
	for _, name := range names {
		if !slices.ContainsFunc(descs, func(d css.Descriptor) bool { return d.Name() == name }) {
			return ErrBadAtRule
		}
	}
	return nil
}

func descriptorBlock(head string, descs []css.Descriptor) []string {
	if len(descs) == 0 {
		return nil
	}
	return []string{head + " {\n" + renderDeclarationsOf(descs, 1) + "}"}
}

func unwrapped(condition string) string {
	if !strings.HasPrefix(condition, "(") {
		return condition
	}
	depth := 0
	for i, c := range condition {
		switch c {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 && i != len(condition)-1 {
				return condition
			}
		}
	}
	if depth != 0 {
		return condition
	}
	return condition[1 : len(condition)-1]
}
