package workflowmgmt

import (
	"context"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
)

type TriggerWorkflowCommandHandler struct {
	WorkflowRepository WorkflowRepository
	UUIDGenerator      muuid.UUIDGenerator
	Clock              misas.Clock
}

func (h TriggerWorkflowCommandHandler) Handle(ctx context.Context, cmd TriggerWorkflowCommand) misas.CommandResult {
	wf, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return mx.CommandResultFromError(err)
	}
	if wf == nil {
		return WorkflowNotFoundCommandResult(cmd.WorkflowID)
	}

	if cmd.RunID == "" {
		cmd.RunID = h.UUIDGenerator.Generate().String()
	}

	if err := wf.Trigger(RunID(cmd.RunID), h.Clock.Now()); err != nil {
		return mx.CommandResultFromError(err)
	}

	return mx.CommandResultFromError(h.WorkflowRepository.Save(ctx, wf))
}
