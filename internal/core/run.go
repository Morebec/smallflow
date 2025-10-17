package core

import (
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
	"time"
)

type WorkflowStatus string

const (
	WorkflowStatusFailed    WorkflowStatus = "failed"
	WorkflowStatusSucceeded WorkflowStatus = "succeeded"
)

type RunID string

type Run struct {
	ID            RunID
	WorkflowID    WorkflowID
	StartedAt     time.Time
	EndedAt       *time.Time
	Error         *WorkflowError
	CurrentStepID StepID
	Steps         map[StepID]*StepRun

	events []any
	Status WorkflowStatus
}

func StartRun(workflowID WorkflowID, id RunID, startedAt time.Time) *Run {
	run := &Run{}
	run.record(WorkflowStartedEvent{
		RunID:      string(id),
		WorkflowID: string(workflowID),
		StartedAt:  startedAt,
	})
	return run
}

func (r *Run) StartStep(stepID StepID, id ActionID, ignoreErrors bool, currentTime time.Time) misas.Error {
	if r.CurrentStepID == stepID {
		// already running, idempotent
		return nil
	}

	if r.CurrentStepID != "" {
		return mx.ErrConflict.WithMessage(fmt.Sprintf(
			"workflow error: %s: cannot start step %s: step %s is currently running",
			r.WorkflowID,
			stepID,
			r.CurrentStepID,
		))
	}

	r.record(StepStartedEvent{
		RunID:        string(r.ID),
		WorkflowID:   string(r.WorkflowID),
		StepID:       string(stepID),
		ActionID:     string(id),
		StartedAt:    currentTime,
		IgnoreErrors: ignoreErrors,
	})

	return nil
}

func (r *Run) EndStep(stepID StepID, err *WorkflowError, currentTime time.Time) misas.Error {
	step := r.Steps[stepID]
	if step.EndedAt != nil {
		// already ended, idempotent
		return nil
	}

	if r.CurrentStepID != stepID {
		return mx.ErrConflict.WithMessage(fmt.Sprintf(
			"workflow error: %s: cannot start step %s: step %s is currently running",
			r.WorkflowID,
			stepID,
			r.CurrentStepID,
		))
	}

	status := StepStatusSucceeded
	if err != nil && !step.IgnoreError {
		status = StepStatusFailed
	}

	r.record(StepEndedEvent{
		RunID:      string(r.ID),
		WorkflowID: string(r.WorkflowID),
		StepID:     string(stepID),
		ActionID:   string(step.ActionID),
		EndedAt:    currentTime,
		Error:      err,
		Status:     string(status),
	})

	return nil
}

func (r *Run) End(currentTime time.Time) {
	if r.EndedAt != nil {
		// Already ended
		return
	}

	// Collect step errors
	var stepErrors map[string]WorkflowError
	workflowStatus := WorkflowStatusSucceeded
	for _, s := range r.Steps {
		if s.Error != nil {
			if stepErrors == nil {
				stepErrors = make(map[string]WorkflowError)
			}
			stepErrors[string(s.ID)] = *s.Error
		}
		if s.Status == StepStatusFailed {
			workflowStatus = WorkflowStatusFailed
		}
	}

	// Compute Status

	r.record(WorkflowEndedEvent{
		RunID:      string(r.ID),
		WorkflowID: string(r.WorkflowID),
		StartedAt:  r.StartedAt,
		EndedAt:    currentTime,
		Status:     string(workflowStatus),
	})
}

func (r *Run) Errors() map[string]*WorkflowError {
	var stepErrors map[string]*WorkflowError
	for _, s := range r.Steps {
		if stepErrors == nil {
			stepErrors = make(map[string]*WorkflowError)
		}
		stepErrors[string(s.ID)] = s.Error
	}
	return stepErrors
}

func (r *Run) Apply(events []any) {
	for _, event := range events {
		switch e := event.(type) {
		case WorkflowStartedEvent:
			r.ID = RunID(e.RunID)
			r.WorkflowID = WorkflowID(e.WorkflowID)
			r.Steps = make(map[StepID]*StepRun)
			r.StartedAt = e.StartedAt

		case WorkflowEndedEvent:
			r.EndedAt = &e.EndedAt
			r.Status = WorkflowStatus(e.Status)

		case StepStartedEvent:
			r.CurrentStepID = StepID(e.StepID)
			r.Steps[StepID(e.StepID)] = &StepRun{
				ID:          StepID(e.StepID),
				StartedAt:   e.StartedAt,
				ActionID:    ActionID(e.ActionID),
				IgnoreError: e.IgnoreErrors,
			}

		case StepEndedEvent:
			r.CurrentStepID = ""
			step := r.Steps[StepID(e.StepID)]
			step.EndedAt = &e.EndedAt
			step.Error = e.Error
			step.Status = StepStatus(e.Status)
		}
	}
}

func (r *Run) UncommittedEvents() []any { return r.events }

func (r *Run) record(event any) {
	r.events = append(r.events, event)
	r.Apply([]any{event})
}

func (r *Run) Commit() { r.events = nil }

type StepStatus string

const (
	StepStatusFailed    StepStatus = "failed"
	StepStatusSucceeded StepStatus = "succeeded"
)

type StepRun struct {
	ID          StepID
	StartedAt   time.Time
	EndedAt     *time.Time
	Error       *WorkflowError
	ActionID    ActionID
	Status      StepStatus
	IgnoreError bool
}
