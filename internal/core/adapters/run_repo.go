package adapters

import (
	"context"
	"github.com/morebec/smallflow/internal/core"
)

type EventStoreRunRepository struct {
	EventStore *InMemoryEventStore
}

func (r EventStoreRunRepository) Add(ctx context.Context, run *core.Run) error {
	return r.Save(ctx, run)
}

func (r EventStoreRunRepository) Save(ctx context.Context, run *core.Run) error {
	for _, event := range run.UncommittedEvents() {
		if err := r.EventStore.Record(ctx, event); err != nil {
			return err
		}
	}
	run.Commit()

	return nil
}

func (r EventStoreRunRepository) FindByID(_ context.Context, workflowID string, runID string) (*core.Run, error) {
	var run *core.Run
	for _, event := range r.EventStore.events {
		switch ev := event.(type) {
		case core.WorkflowStartedEvent:
			run = &core.Run{
				ID:         core.RunID(runID),
				WorkflowID: core.WorkflowID(workflowID),
			}
			if ev.RunID == runID && ev.WorkflowID == workflowID {
				run.Apply([]any{ev})
			}
		}
	}

	return run, nil

}
