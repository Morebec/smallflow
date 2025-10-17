package core

import (
	"context"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
)

type DisableWorkflowCommandHandler struct {
	Clock              misas.Clock
	WorkflowRepository WorkflowRepository
}

func (h DisableWorkflowCommandHandler) Handle(ctx context.Context, cmd DisableWorkflowCommand) misas.CommandResult {
	workflow, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return mx.CommandResultFromError(err)
	}
	if workflow == nil {
		return WorkflowNotFoundCommandResult(cmd.WorkflowID)
	}

	workflow.Disable(h.Clock.Now())

	return mx.CommandResultFromError(h.WorkflowRepository.Save(ctx, workflow))
}
