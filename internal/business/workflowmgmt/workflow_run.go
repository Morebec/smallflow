package workflowmgmt

import (
	"context"

	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
)

type RunWorkflowCommandHandler struct {
	Clock              misas.Clock
	WorkflowRepository WorkflowRepository
	RunRepository      RunRepository
}

func (h RunWorkflowCommandHandler) Handle(ctx context.Context, cmd RunWorkflowCommand) misas.CommandResult {
	wf, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return mx.CommandResultFromError(err)
	}
	if wf == nil {
		return WorkflowNotFoundCommandResult(cmd.WorkflowID)
	}

	runner := WorkflowRunner{
		RunRepository:      h.RunRepository,
		WorkflowRepository: h.WorkflowRepository,
		Clock:              h.Clock,
	}

	report, err := runner.Run(ctx, wf, RunID(cmd.RunID))
	if err != nil {
		return mx.CommandResultFromError(err)
	}

	return misas.CommandResult{Payload: report}
}
