package core

import (
	"context"
	"fmt"
)

type EnableWorkflowCommandHandler struct {
	Clock              Clock
	WorkflowRepository WorkflowRepository
}

func (h EnableWorkflowCommandHandler) Handle(ctx context.Context, cmd EnableWorkflowCommand) error {
	workflow, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return err
	}
	if workflow == nil {
		return fmt.Errorf("workflow not found: %s", cmd.WorkflowID)
	}

	workflow.Enable(h.Clock.Now())

	return h.WorkflowRepository.Save(ctx, workflow)
}
