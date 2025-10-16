package core

import (
	"context"
	"fmt"
)

type TriggerWorkflowCommandHandler struct {
	WorkflowRepository WorkflowRepository
	Clock              Clock
}

func (h TriggerWorkflowCommandHandler) Handle(ctx context.Context, cmd TriggerWorkflowCommand) error {
	wf, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return err
	}
	if wf == nil {
		return fmt.Errorf("workflow not found: %s", cmd.WorkflowID)
	}

	if cmd.WorkflowID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	if cmd.RunID == "" {
		cmd.RunID = generateRunID()
	}

	if err := wf.Trigger(RunID(cmd.RunID), h.Clock.Now()); err != nil {
		return err
	}

	return h.WorkflowRepository.Save(ctx, wf)
}

func generateRunID() string { return "some-generated-run-id" }
