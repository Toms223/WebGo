package app

import "errors"

// Model binds a *T to a key in the app state store. It holds the pointer given to NewModel, without
// copying it.
type Model[T any] struct {
	key   string
	state *T
}

// NewModel returns a model for value under key. It returns an error when value is nil or key is
// empty, and does not touch the store until Get or Load.
func NewModel[T any](key string, value *T) (*Model[T], error) {
	if value == nil {
		return nil, errors.New("value is nil")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	return &Model[T]{key: key, state: value}, nil
}

// Get reads the stored value for the key into the model's state and returns the model. A missing key
// or a value of another type leaves the state unchanged.
func (m *Model[T]) Get(ctx Context) *Model[T] {
	ctx.GetState(m.key, m.state)
	return m
}

// State returns the in-memory state without reading the store.
func (m *Model[T]) State() *T {
	return m.state
}

// Load refreshes the state from the store, as Get does, and returns a Change to apply and publish it.
func (m *Model[T]) Load(ctx Context) *Change[T] {
	ctx.GetState(m.key, m.state)
	return &Change[T]{ctx: ctx, model: m}
}

// Observe returns an observer of the key. It fires on Set and SetAndPersist, not on direct writes to
// the state.
func (m *Model[T]) Observe(ctx Context) Observer {
	return ctx.ObserveState(m.key, m)
}

func (m *Model[T]) target() any { return m.state }

// Change is a pending change to a Model's state, made by Load.
type Change[T any] struct {
	ctx   Context
	model *Model[T]
}

// Apply calls verb with the current state and replaces the state with the result unless it is nil.
// The result may be a new pointer. Apply changes memory only; Set publishes.
func (c *Change[T]) Apply(verb func(ctx Context, state *T) *T) *Change[T] {
	result := verb(c.ctx, c.model.state)
	if result == nil {
		return c
	}
	c.model.state = result
	return c
}

// Set writes the current state to the store, notifying observers, without persisting it.
func (c *Change[T]) Set() *Model[T] {
	c.ctx.SetState(c.model.key, c.model.state)
	return c.model
}

// SetAndPersist replaces the state with value, discarding earlier Apply calls, writes it to the store
// and persists a snapshot to local storage.
func (c *Change[T]) SetAndPersist(value *T) *Model[T] {
	c.model.state = value
	c.ctx.SetState(c.model.key, value).Persist()
	return c.model
}
