package core

import (
	"context"
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
)

type WorkflowRunner struct {
	RunRepository      RunRepository
	WorkflowRepository WorkflowRepository
	Clock              misas.Clock
}

func (r WorkflowRunner) Run(ctx context.Context, wf *Workflow, rID RunID) (WorkflowRunReport, misas.Error) {
	workflowID := wf.ID()
	run, err := r.RunRepository.FindByID(ctx, string(workflowID), string(rID))
	if err != nil {
		return WorkflowRunReport{}, err
	}
	if run != nil {
		// Run already exists, nothing to do.
		return WorkflowRunReport{}, nil
	}

	run = StartRun(workflowID, rID, r.Clock.Now())
	if err := r.RunRepository.Add(ctx, run); err != nil {
		return WorkflowRunReport{}, err
	}

	if err := r.executeRun(ctx, run, wf.Definition().Steps); err != nil {
		return WorkflowRunReport{}, err
	}

	return newWorkflowRunReport(run), nil
}

func (r WorkflowRunner) ResumeFromStep(ctx context.Context, wf *Workflow, rID RunID) (WorkflowRunReport, misas.Error) {
	workflowID := wf.ID()
	run, err := r.RunRepository.FindByID(ctx, string(workflowID), string(rID))
	if err != nil {
		return WorkflowRunReport{}, err
	}
	if run == nil {
		// TODO: better error message.
		return WorkflowRunReport{}, mx.ErrNotFound.WithMessage(fmt.Sprintf("workflow run not found: %s", rID))
	}

	steps := r.remainingStepsFrom(wf.Definition().Steps, run.CurrentStepID)
	if err := r.executeRun(ctx, run, steps); err != nil {
		return WorkflowRunReport{}, err
	}

	return newWorkflowRunReport(run), nil
}

func (r WorkflowRunner) runSteps(ctx context.Context, run *Run, steps []StepDefinition) misas.Error {
	// If a start step ID is provided, set from to false until we encounter it.
	for _, step := range steps {
		if err := r.runStep(ctx, run, step); err != nil {
			return err
		}
	}

	return nil
}

func (r WorkflowRunner) runStep(ctx context.Context, run *Run, step StepDefinition) misas.Error {
	if err := run.StartStep(step.ID, step.Action.ID(), step.IgnoreError, r.Clock.Now()); err != nil {
		return err
	}
	if err := r.RunRepository.Save(ctx, run); err != nil {
		return err
	}

	workflowErr := r.runStepAction(ctx, step)

	if err := run.EndStep(step.ID, workflowErr, r.Clock.Now()); err != nil {
		return err
	}
	if err := r.RunRepository.Save(ctx, run); err != nil {
		return err
	}

	return nil
}

func (r WorkflowRunner) runStepAction(ctx context.Context, step StepDefinition) *WorkflowError {
	return step.Action.Run(ctx)
}

func (r WorkflowRunner) executeRun(ctx context.Context, run *Run, steps []StepDefinition) misas.Error {
	err := r.runSteps(ctx, run, steps)

	run.End(r.Clock.Now())
	if repoErr := r.RunRepository.Save(ctx, run); repoErr != nil {
		err = misas.ErrorGroup{err, repoErr}
	}

	return err
}

func (r WorkflowRunner) remainingStepsFrom(steps []StepDefinition, start StepID) []StepDefinition {
	var remaining []StepDefinition
	include := false
	for _, step := range steps {
		if !include && step.ID == start {
			include = true
		}
		if include {
			remaining = append(remaining, step)
		}
	}

	return remaining
}

func newWorkflowRunReport(run *Run) WorkflowRunReport {
	return WorkflowRunReport{
		WorkflowID: string(run.WorkflowID),
		RunID:      string(run.ID),
		StartedAt:  run.StartedAt,
		EndedAt:    *run.EndedAt,
		Errors:     run.Errors(),
		Status:     string(run.Status),
	}
}
