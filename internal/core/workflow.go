package core

import (
	"fmt"
	"time"
)

type WorkflowID string

type Workflow struct {
	enabled    bool
	definition WorkflowDefinition

	runIds     map[RunID]struct{}
	activeRuns map[RunID]struct{}

	events []any
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

func (w *Workflow) Trigger(id RunID, currentTime time.Time) error {
	if _, exists := w.runIds[id]; exists {
		return nil // idempotent
	}

	if !w.enabled {
		return fmt.Errorf("workflow is not enabled: %s", w.ID())
	}

	if w.definition.ConcurrencyLimit != ConcurrencyLimitNone &&
		len(w.activeRuns) >= int(w.definition.ConcurrencyLimit) {
		return fmt.Errorf("workflow concurrency limit reached: %s", w.ID())
	}

	w.record(WorkflowTriggeredEvent{
		WorkflowID:  string(w.ID()),
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

func (w *Workflow) record(event any) {
	w.Apply([]any{event})
	w.events = append(w.events, event)
}

func (w *Workflow) UncommittedEvents() []any { return w.events }

func (w *Workflow) Apply(events []any) {
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
