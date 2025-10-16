package core

import (
	"context"
	"fmt"
)

type DisableWorkflowCommandHandler struct {
	Clock              Clock
	WorkflowRepository WorkflowRepository
}

func (h DisableWorkflowCommandHandler) Handle(ctx context.Context, cmd DisableWorkflowCommand) error {
	workflow, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return err
	}
	if workflow == nil {
		return fmt.Errorf("workflow not found: %s", cmd.WorkflowID)
	}

	workflow.Disable(h.Clock.Now())

	return h.WorkflowRepository.Save(ctx, workflow)
}
