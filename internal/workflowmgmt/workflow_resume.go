package workflowmgmt

import (
	"context"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
)

type ResumeWorkflowRunCommandHandler struct {
	WorkflowRepository WorkflowRepository
	RunRepository      RunRepository
	Runner             WorkflowRunner
}

func (h ResumeWorkflowRunCommandHandler) Handle(ctx context.Context, cmd ResumeWorkflowRunCommand) misas.CommandResult {
	wf, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return mx.CommandResultFromError(err)
	}
	if wf == nil {
		return WorkflowNotFoundCommandResult(cmd.WorkflowID)
	}

	report, err := h.Runner.ResumeFromStep(ctx, wf, RunID(cmd.RunID))
	if err != nil {
		return mx.CommandResultFromError(err)
	}

	return misas.CommandResult{Payload: report}
}
