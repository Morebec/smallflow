package core

import (
	"context"
	"fmt"
	"time"
)

type API struct {
	EnableWorkflowCommandHandler  EnableWorkflowCommandHandler
	DisableWorkflowCommandHandler DisableWorkflowCommandHandler
	TriggerWorkflowCommandHandler TriggerWorkflowCommandHandler
	RunWorkflowCommandHandler     RunWorkflowCommandHandler
}

func NewAPI(clock Clock, workflowRepo WorkflowRepository, repo RunRepository) *API {
	return &API{
		EnableWorkflowCommandHandler: EnableWorkflowCommandHandler{
			Clock:              clock,
			WorkflowRepository: workflowRepo,
		},
		DisableWorkflowCommandHandler: DisableWorkflowCommandHandler{
			Clock:              clock,
			WorkflowRepository: workflowRepo,
		},
		TriggerWorkflowCommandHandler: TriggerWorkflowCommandHandler{
			Clock:              clock,
			WorkflowRepository: workflowRepo,
		},
		RunWorkflowCommandHandler: RunWorkflowCommandHandler{
			Clock:              clock,
			WorkflowRepository: workflowRepo,
			RunRepository:      repo,
		},
	}
}

func (api API) HandleCommand(ctx context.Context, cmd any) error {
	switch c := cmd.(type) {
	case EnableWorkflowCommand:
		return api.EnableWorkflowCommandHandler.Handle(ctx, c)
	case DisableWorkflowCommand:
		return api.DisableWorkflowCommandHandler.Handle(ctx, c)
	case TriggerWorkflowCommand:
		return api.TriggerWorkflowCommandHandler.Handle(ctx, c)
	case RunWorkflowCommand:
		return api.RunWorkflowCommandHandler.Handle(ctx, c)
	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

// TriggerWorkflowCommand represents a command to trigger a new run of a
// workflow.
//
// This command is idempotent based on the RunID provided. If a run
// with the same RunID already exists for the given workflow, this command will
// have no effect and will succeed.
//
// If no RunID is provided, a new unique RunID
// will be generated.
//
// The workflow must be enabled for this command to succeed.
// If the workflow is disabled, this command will fail. If the workflow does not
// exist, this command will fail. If the concurrent runs limit has been reached,
// this command will fail.
type TriggerWorkflowCommand struct {
	WorkflowID string
	RunID      string
}

type WorkflowTriggeredEvent struct {
	WorkflowID  string
	RunID       string
	TriggeredAt time.Time
}

// EnableWorkflowCommand represents a command to enable a workflow. Enabling a
// workflow allows new runs to be triggered. If the workflow does not exist, this
// command will fail.
type EnableWorkflowCommand struct {
	WorkflowID string
}

// WorkflowEnabledEvent is emitted when a workflow is successfully enabled.

type WorkflowEnabledEvent struct {
	WorkflowID string
	EnabledAt  time.Time
}

// DisableWorkflowCommand represents a command to disable a workflow. Disabling a
// workflow prevents new runs from being triggered, but does not affect currently
// active runs.
type DisableWorkflowCommand struct {
	WorkflowID string
}

type WorkflowDisabledEvent struct {
	WorkflowID string
	DisabledAt time.Time
	ActiveRuns int
}

type WorkflowStartedEvent struct {
	WorkflowID string
	RunID      string
	StartedAt  time.Time
}

type WorkflowEndedEvent struct {
	WorkflowID string
	RunID      string
	EndedAt    time.Time
	StartedAt  time.Time
	Errors     map[string]*WorkflowError
	Status     string
}

type StepStartedEvent struct {
	WorkflowID   string
	RunID        string
	StepID       string
	ActionID     string
	StartedAt    time.Time
	IgnoreErrors bool
}

type StepEndedEvent struct {
	WorkflowID string
	RunID      string
	StepID     string
	ActionID   string
	EndedAt    time.Time
	Error      *WorkflowError
	Status     string
}

type WorkflowError struct {
	Kind    string         // e.g. "user", "system", "internal"
	Code    string         // e.g. "timeout", "network_error", "invalid_input"
	Message string         // human-readable message
	Details map[string]any // additional details, e.g. {"go_error": "error message"}
}

func (e WorkflowError) AsError() error {
	return fmt.Errorf("%s(%s): %s", e.Kind, e.Code, e.Message)
}

// RunWorkflowCommand runs a workflow synchronously.
// This command is idempotent based on the RunID provided:
// If a run with the same RunID already exists for the given workflow, this command will
// have no effect and will succeed.
// If no RunID is provided, this command will fail.
// This command is intended to be run after a workflow has been triggered.
//
// This command will not fail if the workflow or any of its steps fail, as these
// are expected outcomes of a workflow run. Instead, the errors will be recorded
// in the StepEndedEvent and WorkflowEndedEvent.
// This command will only return if internal errors have occurred, such as
// issues with the data storage.
type RunWorkflowCommand struct {
	WorkflowID string
	RunID      string
}

type WorkflowRunReport struct {
	WorkflowID string
	RunID      string
	StartedAt  time.Time
	EndedAt    time.Time
	Errors     map[string]*WorkflowError
	Status     string
}
