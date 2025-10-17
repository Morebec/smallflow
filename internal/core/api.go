package core

import (
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
	"time"
)

func NewSubsystem(
	clock misas.Clock,
	workflowRepo WorkflowRepository,
	runRepository RunRepository,
	uidg muuid.UUIDGenerator,
) misas.BusinessSubsystem {
	return mx.NewBusinessSubsystemAssembler().
		WithCommandHandler(EnableWorkflowCommandTypeName, mx.NewTypedCommandHandler(EnableWorkflowCommandHandler{
			WorkflowRepository: workflowRepo,
			Clock:              clock,
		})).
		WithCommandHandler(DisableWorkflowCommandTypeName, mx.NewTypedCommandHandler(DisableWorkflowCommandHandler{
			WorkflowRepository: workflowRepo,
			Clock:              clock,
		})).
		WithCommandHandler(TriggerWorkflowCommandTypeName, mx.NewTypedCommandHandler(TriggerWorkflowCommandHandler{
			WorkflowRepository: workflowRepo,
			UUIDGenerator:      uidg,
			Clock:              clock,
		})).
		WithCommandHandler(RunWorkflowCommandTypeName, mx.NewTypedCommandHandler(RunWorkflowCommandHandler{
			WorkflowRepository: workflowRepo,
			RunRepository:      runRepository,
			Clock:              clock,
		})).
		WithCommandHandler(ResumeWorkflowRunCommandTypeName, mx.NewTypedCommandHandler(ResumeWorkflowRunCommandHandler{
			WorkflowRepository: workflowRepo,
			RunRepository:      runRepository,
		})).
		Assemble()
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

const TriggerWorkflowCommandTypeName = "TriggerWorkflowCommand"

func (TriggerWorkflowCommand) TypeName() misas.CommandTypeName { return TriggerWorkflowCommandTypeName }

type WorkflowTriggeredEvent struct {
	WorkflowID  string
	RunID       string
	TriggeredAt time.Time
}

const WorkflowTriggeredEventTypeName = "WorkflowTriggeredEvent"

func (WorkflowTriggeredEvent) TypeName() misas.EventTypeName { return WorkflowTriggeredEventTypeName }

// EnableWorkflowCommand represents a command to enable a workflow. Enabling a
// workflow allows new runs to be triggered. If the workflow does not exist, this
// command will fail.
type EnableWorkflowCommand struct {
	WorkflowID string
}

const EnableWorkflowCommandTypeName = "EnableWorkflowCommand"

func (EnableWorkflowCommand) TypeName() misas.CommandTypeName { return EnableWorkflowCommandTypeName }

// WorkflowEnabledEvent is emitted when a workflow is successfully enabled.

type WorkflowEnabledEvent struct {
	WorkflowID string
	EnabledAt  time.Time
}

const WorkflowEnabledEventTypeName = "WorkflowEnabledEvent"

func (WorkflowEnabledEvent) TypeName() misas.EventTypeName { return WorkflowEnabledEventTypeName }

// DisableWorkflowCommand represents a command to disable a workflow. Disabling a
// workflow prevents new runs from being triggered, but does not affect currently
// active runs.
type DisableWorkflowCommand struct {
	WorkflowID string
}

const DisableWorkflowCommandTypeName = "DisableWorkflowCommand"

func (DisableWorkflowCommand) TypeName() misas.CommandTypeName { return DisableWorkflowCommandTypeName }

type WorkflowDisabledEvent struct {
	WorkflowID string
	DisabledAt time.Time
	ActiveRuns int
}

const WorkflowDisabledEventTypeName = "WorkflowDisabledEvent"

func (WorkflowDisabledEvent) TypeName() misas.EventTypeName { return WorkflowDisabledEventTypeName }

type WorkflowStartedEvent struct {
	WorkflowID string
	RunID      string
	StartedAt  time.Time
}

const WorkflowStartedEventTypeName = "WorkflowStartedEvent"

func (WorkflowStartedEvent) TypeName() misas.EventTypeName { return WorkflowStartedEventTypeName }

type WorkflowEndedEvent struct {
	WorkflowID string
	RunID      string
	EndedAt    time.Time
	StartedAt  time.Time
	Errors     map[string]*WorkflowError
	Status     string
}

const WorkflowEndedEventTypeName = "WorkflowEndedEvent"

func (WorkflowEndedEvent) TypeName() misas.EventTypeName { return WorkflowEndedEventTypeName }

type StepStartedEvent struct {
	WorkflowID   string
	RunID        string
	StepID       string
	ActionID     string
	StartedAt    time.Time
	IgnoreErrors bool
}

const StepStartedEventTypeName = "StepStartedEvent"

func (StepStartedEvent) TypeName() misas.EventTypeName { return StepStartedEventTypeName }

type StepEndedEvent struct {
	WorkflowID string
	RunID      string
	StepID     string
	ActionID   string
	EndedAt    time.Time
	Error      *WorkflowError
	Status     string
}

const StepEndedEventTypeName = "StepEndedEvent"

func (StepEndedEvent) TypeName() misas.EventTypeName { return StepEndedEventTypeName }

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

const RunWorkflowCommandTypeName = "RunWorkflowCommand"

func (RunWorkflowCommand) TypeName() misas.CommandTypeName { return RunWorkflowCommandTypeName }

type ResumeWorkflowRunCommand struct {
	WorkflowID string
	RunID      string
}

const ResumeWorkflowRunCommandTypeName = "ResumeWorkflowRunCommand"

func (ResumeWorkflowRunCommand) TypeName() misas.CommandTypeName {
	return ResumeWorkflowRunCommandTypeName
}

type WorkflowRunReport struct {
	WorkflowID string
	RunID      string
	StartedAt  time.Time
	EndedAt    time.Time
	Errors     map[string]*WorkflowError
	Status     string
}

const ErrorCodeWorkflowNotFound misas.ErrorCode = "workflow_not_found"

var ErrWorkflowNotFound = mx.ErrNotFound.WithCode(ErrorCodeWorkflowNotFound).WithMessage("workflow not found")

const ErrorCodeRunNotFound misas.ErrorCode = "workflow_run_not_found"

var ErrWorkflowRunNotFound = mx.ErrNotFound.WithCode(ErrorCodeRunNotFound).WithMessage("workflow run not found")
