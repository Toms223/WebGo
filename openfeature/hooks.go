package openfeature

import "maps"

// Hook runs around flag evaluation. Nil stages are skipped.
type Hook struct {
	Before  func(HookContext, HookHints) error
	After   func(HookContext, EvaluationDetails[any], HookHints) error
	Error   func(HookContext, error, HookHints)
	Finally func(HookContext, EvaluationDetails[any], HookHints)
}

// HookHints are read-only per-evaluation hints passed to every hook stage.
type HookHints struct {
	m map[string]any
}

// NewHookHints returns hints holding a shallow copy of m.
func NewHookHints(m map[string]any) HookHints {
	return HookHints{m: maps.Clone(m)}
}

// Value returns the hint stored under key, or nil.
func (h HookHints) Value(key string) any {
	return h.m[key]
}

// Map returns a shallow copy of the hints, or nil if none were set.
func (h HookHints) Map() map[string]any {
	return maps.Clone(h.m)
}

// HookContext is what a hook sees about the evaluation.
type HookContext struct {
	FlagKey          string
	DefaultValue     any
	FlagType         FlagType
	Context          Context
	ClientMetadata   ClientMetadata
	ProviderMetadata Metadata
	Logger           Logger
	HookData         *HookData
}

// HookData is per-hook, per-evaluation storage.
type HookData struct{ m map[string]any }

// NewHookData returns an empty HookData.
func NewHookData() *HookData { return &HookData{m: map[string]any{}} }

// Set stores v under k.
func (d *HookData) Set(k string, v any) { d.m[k] = v }

// Get returns the value under k, or nil.
func (d *HookData) Get(k string) any { return d.m[k] }

// Has reports whether k is set.
func (d *HookData) Has(k string) bool { _, ok := d.m[k]; return ok }

// Delete removes k and reports whether it was set.
func (d *HookData) Delete(k string) bool {
	_, ok := d.m[k]
	delete(d.m, k)
	return ok
}

// Clear removes every entry.
func (d *HookData) Clear() { clear(d.m) }
