// Package memprovider is an in-memory OpenFeature provider, a port of the
// Web SDK's TypedInMemoryProvider.
package memprovider

import (
	"fmt"
	"math"
	"reflect"
	"slices"
	"strconv"
	"sync"

	"github.com/Toms223/WebGo/openfeature"
)

// Flag configures one flag.
type Flag struct {
	Key              string
	Variants         map[string]any
	DefaultVariant   string
	Disabled         bool
	ContextEvaluator func(openfeature.Context) string
}

// Provider serves flags from memory.
type Provider struct {
	mu     sync.Mutex
	keys   []string
	flags  map[string]Flag
	events *openfeature.Emitter
}

// New returns a Provider holding flags. A repeated key keeps its first
// position and its last value.
func New(flags ...Flag) *Provider {
	p := &Provider{events: openfeature.NewEmitter()}
	p.keys, p.flags = index(flags)
	return p
}

func index(flags []Flag) ([]string, map[string]Flag) {
	keys := []string{}
	m := make(map[string]Flag, len(flags))
	for _, f := range flags {
		if _, ok := m[f.Key]; !ok {
			keys = append(keys, f.Key)
		}
		m[f.Key] = f
	}
	return keys, m
}

// PutConfiguration replaces all flags and emits ConfigurationChanged listing
// the new keys in order, then the keys only the old configuration had.
func (p *Provider) PutConfiguration(flags ...Flag) error {
	keys, m := index(flags)
	p.mu.Lock()
	changed := append([]string{}, keys...)
	for _, k := range p.keys {
		if _, ok := m[k]; !ok {
			changed = append(changed, k)
		}
	}
	p.keys, p.flags = keys, m
	p.mu.Unlock()
	changed = jsKeyOrder(changed)
	p.events.Emit(openfeature.EventConfigurationChanged, openfeature.EventDetails{FlagsChanged: changed})
	return nil
}

func arrayIndex(k string) (uint64, bool) {
	n, err := strconv.ParseUint(k, 10, 32)
	return n, err == nil && n < math.MaxUint32 && strconv.FormatUint(n, 10) == k
}

func jsKeyOrder(keys []string) []string {
	var idx, rest []string
	for _, k := range keys {
		if _, ok := arrayIndex(k); ok {
			idx = append(idx, k)
		} else {
			rest = append(rest, k)
		}
	}
	slices.SortFunc(idx, func(a, b string) int {
		x, _ := arrayIndex(a)
		y, _ := arrayIndex(b)
		return int(x) - int(y)
	})
	return append(idx, rest...)
}

var objectPrototypeKeys = []string{
	"constructor", "toString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable",
	"toLocaleString", "__proto__", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__",
}

// Metadata returns the provider's metadata.
func (p *Provider) Metadata() openfeature.Metadata { return openfeature.Metadata{Name: "in-memory"} }

// RunsOn reports "client".
func (p *Provider) RunsOn() string { return "client" }

// Events returns the provider's emitter.
func (p *Provider) Events() *openfeature.Emitter { return p.events }

func jsTypeof(v any) string {
	switch reflect.ValueOf(v).Kind() {
	case reflect.Bool:
		return "boolean"
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return "number"
	}
	return "object"
}

func resolve[T any](p *Provider, key string, def any, ctx openfeature.Context, log openfeature.Logger, conv func(any) T) openfeature.ResolutionDetails[T] {
	fail := func(code openfeature.ErrorCode, msg string) openfeature.ResolutionDetails[T] {
		var zero T
		return openfeature.ResolutionDetails[T]{Value: zero, Reason: openfeature.ReasonError, ErrorCode: code, ErrorMessage: msg}
	}
	p.mu.Lock()
	f, ok := p.flags[key]
	p.mu.Unlock()
	if !ok && slices.Contains(objectPrototypeKeys, key) {
		msg := fmt.Sprintf("no value associated with variant provided for %s found", key)
		log.Error(msg)
		return fail(openfeature.CodeGeneral, msg)
	}
	if !ok {
		msg := "no flag found with key " + key
		log.Debug(msg)
		return fail(openfeature.CodeFlagNotFound, msg)
	}
	if f.Disabled {
		return openfeature.ResolutionDetails[T]{Value: conv(def), Reason: openfeature.ReasonDisabled}
	}
	variant, reason := f.DefaultVariant, openfeature.ReasonStatic
	if f.ContextEvaluator != nil {
		variant, reason = f.ContextEvaluator(ctx), openfeature.ReasonTargetingMatch
	}
	var value any = ""
	if variant != "" {
		v, ok := f.Variants[variant]
		switch {
		case f.Variants == nil:
			return fail(openfeature.CodeGeneral, fmt.Sprintf("Cannot read properties of undefined (reading '%s')", variant))
		case ok:
			value = v
		case variant == "__proto__":
			value = openfeature.Object{}
		case slices.Contains(objectPrototypeKeys, variant):
			return fail(openfeature.CodeTypeMismatch, "")
		default:
			msg := fmt.Sprintf("no value associated with variant provided for %s found", key)
			log.Error(msg)
			return fail(openfeature.CodeGeneral, msg)
		}
	}
	if jsTypeof(value) != jsTypeof(def) {
		return fail(openfeature.CodeTypeMismatch, "")
	}
	return openfeature.ResolutionDetails[T]{Value: conv(value), Variant: variant, Reason: reason}
}

func toFloat(v any) float64 {
	rv := reflect.ValueOf(v)
	switch {
	case rv.CanInt():
		return float64(rv.Int())
	case rv.CanUint():
		return float64(rv.Uint())
	}
	return rv.Float()
}

// ResolveBoolean resolves a boolean flag.
func (p *Provider) ResolveBoolean(flag string, def bool, ctx openfeature.Context, log openfeature.Logger) openfeature.ResolutionDetails[bool] {
	return resolve(p, flag, def, ctx, log, func(v any) bool { return reflect.ValueOf(v).Bool() })
}

// ResolveString resolves a string flag.
func (p *Provider) ResolveString(flag string, def string, ctx openfeature.Context, log openfeature.Logger) openfeature.ResolutionDetails[string] {
	return resolve(p, flag, def, ctx, log, func(v any) string { return reflect.ValueOf(v).String() })
}

// ResolveNumber resolves a number flag.
func (p *Provider) ResolveNumber(flag string, def float64, ctx openfeature.Context, log openfeature.Logger) openfeature.ResolutionDetails[float64] {
	return resolve(p, flag, def, ctx, log, toFloat)
}

// ResolveObject resolves an object flag; the value is returned as configured.
func (p *Provider) ResolveObject(flag string, def any, ctx openfeature.Context, log openfeature.Logger) openfeature.ResolutionDetails[any] {
	return resolve(p, flag, def, ctx, log, func(v any) any { return v })
}
