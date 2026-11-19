package parity

import (
	"github.com/Toms223/WebGo/openfeature"
	"github.com/Toms223/WebGo/openfeature/memprovider"
)

func field(o object, k string) (any, bool) {
	for _, m := range o {
		if m.Key == k {
			return m.Value, true
		}
	}
	return nil, false
}

func memFlags(v any) []memprovider.Flag {
	var out []memprovider.Flag
	flags, _ := v.(object)
	for _, m := range flags {
		def, _ := m.Value.(object)
		f := memprovider.Flag{Key: m.Key}
		if vs, ok := field(def, "variants"); ok {
			f.Variants = map[string]any{}
			vo, _ := vs.(object)
			for _, e := range vo {
				f.Variants[e.Key] = e.Value
			}
		}
		if dv, ok := field(def, "defaultVariant"); ok {
			f.DefaultVariant, _ = dv.(string)
		}
		if d, ok := field(def, "disabled"); ok {
			f.Disabled, _ = d.(bool)
		}
		if ce, ok := field(def, "contextEvaluator"); ok {
			f.ContextEvaluator = memEvaluator(ce.(object), f.DefaultVariant)
		}
		out = append(out, f)
	}
	return out
}

func memEvaluator(def object, defaultVariant string) func(openfeature.Context) string {
	if match, ok := field(def, "match"); ok {
		variant, _ := field(def, "variant")
		return func(ctx openfeature.Context) string {
			for _, e := range match.(object) {
				var got any
				if e.Key == "targetingKey" {
					got = ctx.TargetingKey
				} else {
					got = ctx.Attributes[e.Key]
				}
				if got != e.Value {
					return defaultVariant
				}
			}
			s, _ := variant.(string)
			return s
		}
	}
	attr, _ := field(def, "attr")
	mp, _ := field(def, "map")
	return func(ctx openfeature.Context) string {
		a, _ := attr.(string)
		var v any
		if a == "targetingKey" {
			v = ctx.TargetingKey
		} else {
			v = ctx.Attributes[a]
		}
		if s, ok := v.(string); ok {
			if t, ok := field(mp.(object), s); ok {
				r, _ := t.(string)
				return r
			}
		}
		return defaultVariant
	}
}
