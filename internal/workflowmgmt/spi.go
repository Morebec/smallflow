package workflowmgmt

import (
	"context"
	"github.com/morebec/go-misas/misas"
)

type WorkflowRepository interface {
	FindByID(ctx context.Context, workflowID string) (*Workflow, misas.Error)
	Save(ctx context.Context, wf *Workflow) misas.Error
}

type RunRepository interface {
	FindByID(ctx context.Context, workflowID string, runID string) (*Run, misas.Error)
	Add(ctx context.Context, run *Run) misas.Error
	Save(ctx context.Context, r *Run) misas.Error
}
