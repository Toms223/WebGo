package style

import (
	"cmp"
	"slices"
	"strings"

	"github.com/Toms223/WebGo/css"

	"github.com/Toms223/WebGo/app"
)

// Binding is a typed reaction to one action, attached with Bind. The action's value must be a *T; any
// other value is logged and leaves the item unchanged.
type Binding struct {
	action string
	props  func(any) ([]css.Property, bool)
	frames func(any) ([]Frame, bool)
	when   func(any) (bool, bool)
}

// When returns a binding that inserts the item while when returns true for the action's *T and removes
// it otherwise.
func When[T any](action string, when func(*T) bool) Binding {
	return Binding{action: action, when: typed(when)}
}

// Compute returns a binding whose function gives a Rule's base block for the action's *T. Pseudo,
// conditional and nested blocks stay as declared, and a nil or empty result removes the rule.
func Compute[T any](action string, props func(*T) []css.Property) Binding {
	return Binding{action: action, props: typed(props)}
}

// ComputeFrames is Compute for a KeyframesRule: its function gives the frames, and no frame with
// properties removes the keyframes.
func ComputeFrames[T any](action string, frames func(*T) []Frame) Binding {
	return Binding{action: action, frames: typed(frames)}
}

func typed[T, R any](f func(*T) R) func(any) (R, bool) {
	if f == nil {
		return nil
	}
	return func(v any) (R, bool) {
		t, ok := v.(*T)
		if !ok {
			var zero R
			return zero, false
		}
		return f(t), true
	}
}

type binding struct {
	action string
	when   func(app.Action) bool
	typed  *Binding
}

func (b binding) valid() bool {
	if b.action == "" {
		return false
	}
	if b.typed == nil {
		return b.when != nil
	}
	return b.typed.when != nil || b.typed.props != nil || b.typed.frames != nil
}

func bind(bindings []binding, err error, b binding, fits bool) ([]binding, error) {
	if !fits || !b.valid() {
		return bindings, cmp.Or(err, ErrBadBinding)
	}
	return append(bindings, b), err
}

type verdict struct {
	present  bool
	computes bool
	props    []css.Property
	frames   []Frame
}

func (b binding) judge(a app.Action) (verdict, bool) {
	switch {
	case b.typed == nil:
		return verdict{present: b.when(a)}, true
	case b.typed.when != nil:
		present, ok := b.typed.when(a.Value)
		return verdict{present: present}, ok
	case b.typed.props != nil:
		props, ok := b.typed.props(a.Value)
		props = withoutNilProps(props)
		return verdict{present: len(props) > 0, computes: true, props: props}, ok
	default:
		frames, ok := b.typed.frames(a.Value)
		frames = slices.DeleteFunc(slices.Clone(frames), func(f Frame) bool { return len(f.props) == 0 })
		return verdict{present: len(frames) > 0, computes: true, frames: frames}, ok
	}
}

func judge(bindings []binding, a app.Action) (verdict, bool) {
	for _, b := range slices.Backward(bindings) {
		if b.action == a.Name {
			return b.judge(a)
		}
	}
	return verdict{}, false
}

type outcome struct {
	texts    []string
	present  bool
	computes bool
	handled  bool
}

func (r *Rule) textsFor(a app.Action) outcome {
	return evaluate(r.bindings, a, r.blockTexts, func(v verdict) []string { return r.textsWith(v.props) })
}

func (k *KeyframesRule) textsFor(a app.Action) outcome {
	return evaluate(k.bindings, a, k.blockTexts, func(v verdict) []string { return k.textsWith(v.frames) })
}

func evaluate(bindings []binding, a app.Action, declared func() []string, computed func(verdict) []string) outcome {
	v, handled := judge(bindings, a)
	o := outcome{present: v.present, computes: v.computes, handled: handled}
	switch {
	case !handled || !v.present:
	case v.computes:
		o.texts = computed(v)
	default:
		o.texts = declared()
	}
	return o
}

func (r *Rule) startTexts() ([]string, bool) { return r.blockTexts(), r.initially }

func (k *KeyframesRule) startTexts() ([]string, bool) { return k.blockTexts(), k.initially }

func (r *Rule) actions() []string { return actionsOf(r.bindings) }

func (k *KeyframesRule) actions() []string { return actionsOf(k.bindings) }

func actionsOf(bindings []binding) []string {
	var names []string
	for _, b := range bindings {
		if !slices.Contains(names, b.action) {
			names = append(names, b.action)
		}
	}
	return names
}

func cssFor(i dynamicItem, action string, v any) string {
	o := i.textsFor(app.Action{Name: action, Value: v})
	if !o.handled || !o.present || len(o.texts) == 0 {
		return ""
	}
	return strings.Join(o.texts, "\n\n") + "\n"
}

func withoutNilProps(props []css.Property) []css.Property {
	return slices.DeleteFunc(slices.Clone(props), func(p css.Property) bool { return p == nil })
}
