package core

import (
	"context"
	"time"
)

type WorkflowRepository interface {
	FindByID(ctx context.Context, workflowID string) (*Workflow, error)
	Save(ctx context.Context, wf *Workflow) error
}

type RunRepository interface {
	FindByID(ctx context.Context, workflowID string, runID string) (*Run, error)
	Add(ctx context.Context, run *Run) error
	Save(ctx context.Context, r *Run) error
}

type Clock interface{ Now() time.Time }
