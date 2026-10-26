// Package data holds view models: typed state kept under a key in the app state store and shared
// between components.
package data

import (
	"errors"

	"github.com/Toms223/WebGo/app"
)

// ViewModel binds a *T to a key in the app state store. It holds the pointer given to NewViewModel,
// without copying it.
type ViewModel[T any] struct {
	key    string
	state  *T
	source app.UI
}

// NewViewModel returns a model for value under key. It returns an error when value is nil or key is
// empty, and does not touch the store until Get or Load.
func NewViewModel[T any](key string, value *T) (*ViewModel[T], error) {
	if value == nil {
		return nil, errors.New("value is nil")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	return &ViewModel[T]{key: key, state: value}, nil
}

// Get reads the stored value for the key into the model's state and returns the model. A missing key
// or a value of another type leaves the state unchanged.
func (v *ViewModel[T]) Get(ctx app.Context) *ViewModel[T] {
	ctx.GetState(v.key, v.state)
	return v
}

// State returns the in-memory state without reading the store.
func (v *ViewModel[T]) State() *T {
	return v.state
}

// Load refreshes the state from the store, as Get does, and returns an Action to change and publish
// it.
func (v *ViewModel[T]) Load(ctx app.Context) *Action[T] {
	ctx.GetState(v.key, v.state)
	return &Action[T]{ctx: ctx, model: v}
}

// Observe returns an observer of the key. It fires on Set and SetAndPersist, not on direct writes to
// the state.
func (v *ViewModel[T]) Observe(ctx app.Context) app.Observer {
	return ctx.ObserveState(v.key, v.state)
}

// OnAction makes every Set and SetAndPersist fire an app action named after the key and carrying the
// model's *T, for style.When and style.Compute. Call it from lifecycle hooks, never Render; it
// registers once per model until the registering component dismounts.
func (v *ViewModel[T]) OnAction(ctx app.Context) {
	if v.source != nil && v.source.Mounted() {
		return
	}
	v.source = ctx.Src()
	ctx.ObserveState(v.key, v.state).OnChange(func() {
		ctx.NewActionWithValue(v.key, v.state)
	})
}

// Action is a pending change to a ViewModel's state, made by Load.
type Action[T any] struct {
	ctx   app.Context
	model *ViewModel[T]
}

// Apply calls verb with the current state and replaces the state with the result unless it is nil.
// Return the pointer verb was given: a new pointer detaches observers holding the old one. Apply
// changes memory only; Set publishes.
func (a *Action[T]) Apply(verb func(ctx app.Context, state *T) *T) *Action[T] {
	result := verb(a.ctx, a.model.state)
	if result == nil {
		return a
	}
	a.model.state = result
	return a
}

// Set writes the current state to the store, notifying observers, without persisting it.
func (a *Action[T]) Set() *ViewModel[T] {
	a.ctx.SetState(a.model.key, a.model.state)
	return a.model
}

// SetAndPersist replaces the state with value, discarding earlier Apply calls, writes it to the store
// and persists a snapshot to local storage.
func (a *Action[T]) SetAndPersist(value *T) *ViewModel[T] {
	a.model.state = value
	a.ctx.SetState(a.model.key, value).Persist()
	return a.model
}
