package orchestrator

import (
	"context"
	"github.com/morebec/smallflow/internal/core"
)

type WorkflowOrchestrator struct {
	Clock core.Clock
	API   *core.API
}

func (m WorkflowOrchestrator) HandleEvent(ctx context.Context, event any) error {
	switch e := event.(type) {
	case core.WorkflowTriggeredEvent:
		return m.runWorkflow(ctx, e)
	}

	return nil
}

func (m WorkflowOrchestrator) runWorkflow(ctx context.Context, e core.WorkflowTriggeredEvent) error {
	return m.API.HandleCommand(ctx, core.RunWorkflowCommand{
		WorkflowID: e.WorkflowID,
		RunID:      e.RunID,
	})
}
