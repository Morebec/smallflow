package workflowmgmt

import (
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mx"
	"time"
)

type WorkflowID string

type Workflow struct {
	enabled    bool
	definition WorkflowDefinition

	runIds     map[RunID]struct{}
	activeRuns map[RunID]struct{}

	events []misas.Event
}

func NewWorkflow(d WorkflowDefinition, enabled bool) *Workflow {
	return &Workflow{
		definition: d,
		enabled:    enabled,
		runIds:     make(map[RunID]struct{}),
		activeRuns: make(map[RunID]struct{}),
	}
}

func (w *Workflow) ID() WorkflowID { return w.definition.ID }

func (w *Workflow) Trigger(id RunID, currentTime time.Time) misas.Error {
	if _, exists := w.runIds[id]; exists {
		return nil // idempotent
	}

	workflowID := w.ID()
	if !w.enabled {
		return mx.ErrConflict.WithMessage(fmt.Sprintf("workflow is not enabled: %s", workflowID))
	}

	if w.definition.ConcurrencyLimit != ConcurrencyLimitNone &&
		len(w.activeRuns) >= int(w.definition.ConcurrencyLimit) {
		return mx.ErrConflict.WithMessage(fmt.Sprintf("workflow concurrency limit reached: %s", workflowID))
	}

	w.record(WorkflowTriggeredEvent{
		WorkflowID:  string(workflowID),
		RunID:       string(id),
		TriggeredAt: currentTime,
	})

	return nil
}

func (w *Workflow) Enable(now time.Time) {
	if w.enabled {
		return
	}

	w.record(WorkflowEnabledEvent{
		WorkflowID: string(w.ID()),
		EnabledAt:  now,
	})
}

func (w *Workflow) Disable(now time.Time) {
	if !w.enabled {
		return
	}

	w.record(WorkflowDisabledEvent{
		WorkflowID: string(w.ID()),
		DisabledAt: now,
	})
}

func (w *Workflow) record(event misas.Event) {
	w.Apply([]misas.Event{event})
	w.events = append(w.events, event)
}

func (w *Workflow) UncommittedEvents() []misas.Event { return w.events }

func (w *Workflow) Apply(events []misas.Event) {
	for _, event := range events {
		switch e := event.(type) {
		case WorkflowTriggeredEvent:
			w.runIds[RunID(e.RunID)] = struct{}{}
			w.activeRuns[RunID(e.RunID)] = struct{}{}
		case WorkflowEndedEvent:
			delete(w.activeRuns, RunID(e.RunID))

		case WorkflowEnabledEvent:
			w.enabled = true

		case WorkflowDisabledEvent:
			w.enabled = false
		}
	}
}

func (w *Workflow) Definition() WorkflowDefinition { return w.definition }

func (w *Workflow) Commit() { w.events = nil }
