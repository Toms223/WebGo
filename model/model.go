package model

import (
	"errors"

	"github.com/maxence-charriere/go-app/v11/pkg/app"
)

type ViewModel[T any] struct {
	key   string
	state *T
}

func NewViewModel[T any](key string, value *T) (*ViewModel[T], error) {
	if value == nil {
		return nil, errors.New("value is nil")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	return &ViewModel[T]{key: key, state: value}, nil
}

func (v *ViewModel[T]) Get(ctx app.Context) *ViewModel[T] {
	ctx.GetState(v.key, v.state)
	return v
}

func (v *ViewModel[T]) Load(ctx app.Context) *Action[T] {
	ctx.GetState(v.key, v.state)
	return &Action[T]{ctx: ctx, model: v}
}

func (v *ViewModel[T]) Observe(ctx app.Context) app.Observer {
	return ctx.ObserveState(v.key, v.state)
}

type Action[T any] struct {
	ctx   app.Context
	model *ViewModel[T]
}

func (a *Action[T]) Apply(verb func(ctx app.Context, state *T) *T) *Action[T] {
	result := verb(a.ctx, a.model.state)
	if result == nil {
		return a
	}
	a.model.state = result
	return a
}

func (a *Action[T]) Set() *ViewModel[T] {
	a.ctx.SetState(a.model.key, a.model.state)
	return a.model
}

func (a *Action[T]) SetAndPersist(value *T) *ViewModel[T] {
	a.model.state = value
	a.ctx.SetState(a.model.key, value).Persist()
	return a.model
}
