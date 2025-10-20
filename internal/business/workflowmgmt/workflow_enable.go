package workflowmgmt

import (
	"context"

	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
)

type EnableWorkflowCommandHandler struct {
	Clock              misas.Clock
	WorkflowRepository WorkflowRepository
}

func (h EnableWorkflowCommandHandler) Handle(ctx context.Context, cmd EnableWorkflowCommand) misas.CommandResult {
	workflow, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return mx.CommandResultFromError(err)
	}
	if workflow == nil {
		return WorkflowNotFoundCommandResult(cmd.WorkflowID)
	}

	workflow.Enable(h.Clock.Now())

	return mx.CommandResultFromError(h.WorkflowRepository.Save(ctx, workflow))
}
