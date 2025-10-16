package core

import "context"

type ActionID string

type Action interface {
	Run(ctx context.Context) *WorkflowError
	ID() ActionID
}

type ActionFunc struct {
	fn func(ctx context.Context) *WorkflowError
	id ActionID
}

func (a *ActionFunc) Run(ctx context.Context) *WorkflowError { return a.fn(ctx) }
func (a *ActionFunc) ID() ActionID                           { return a.id }

func NewActionFunc(id ActionID, fn func(ctx context.Context) *WorkflowError) *ActionFunc {
	return &ActionFunc{id: id, fn: fn}
}
