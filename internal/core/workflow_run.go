package core

import (
	"context"
	"fmt"
)

type RunWorkflowCommandHandler struct {
	Clock              Clock
	WorkflowRepository WorkflowRepository
	RunRepository      RunRepository
}

func (h RunWorkflowCommandHandler) Handle(ctx context.Context, cmd RunWorkflowCommand) error {
	run, err := h.RunRepository.FindByID(ctx, cmd.WorkflowID, cmd.RunID)
	if err != nil {
		return err
	}
	if run != nil {
		// Run already exists, nothing to do.
		return nil
	}

	wf, err := h.WorkflowRepository.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return err
	}
	if wf == nil {
		return fmt.Errorf("workflow not found: %s", cmd.WorkflowID)
	}

	run = StartRun(WorkflowID(cmd.WorkflowID), RunID(cmd.RunID), h.Clock.Now())
	if err := h.RunRepository.Add(ctx, run); err != nil {
		return err
	}

	for _, step := range wf.Definition().Steps {
		if err = h.runStep(ctx, run, step); err != nil {
			break
		}
	}

	run.End(h.Clock.Now())
	if err := h.RunRepository.Save(ctx, run); err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("workflow run failed: %s: %w", run.ID, err)
	}

	report := WorkflowRunReport{
		WorkflowID: string(run.WorkflowID),
		RunID:      string(run.ID),
		StartedAt:  run.StartedAt,
		EndedAt:    *run.EndedAt,
		Errors:     run.Errors(),
		Status:     string(run.Status),
	}

	fmt.Printf("Workflow run completed: %+v \n", report)

	return nil
}

func (h RunWorkflowCommandHandler) runStep(ctx context.Context, run *Run, step StepDefinition) error {
	if err := run.StartStep(step.ID, step.Action.ID(), step.IgnoreError, h.Clock.Now()); err != nil {
		return err
	}
	if err := h.RunRepository.Save(ctx, run); err != nil {
		return err
	}

	workflowErr := h.runStepAction(ctx, step)

	if err := run.EndStep(step.ID, workflowErr, h.Clock.Now()); err != nil {
		return err
	}
	if err := h.RunRepository.Save(ctx, run); err != nil {
		return err
	}

	return nil
}

func (h RunWorkflowCommandHandler) runStepAction(ctx context.Context, step StepDefinition) *WorkflowError {
	return step.Action.Run(ctx)
}
