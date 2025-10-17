package adapters

import (
	"context"
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/smallflow/internal/core"
)

type EventStoreWorkflowRepository struct {
	EventStore *InMemoryEventStore
}

func (e EventStoreWorkflowRepository) FindByID(_ context.Context, workflowID string) (*core.Workflow, misas.Error) {
	events := e.EventStore.Events()
	wf := core.NewWorkflow(core.WorkflowDefinition{
		ID:               core.WorkflowID(workflowID),
		ConcurrencyLimit: core.ConcurrencyLimitNone,
		Steps: []core.StepDefinition{
			{
				ID:          "step-1",
				IgnoreError: false,
				Action: core.NewActionFunc("my-action", func(ctx context.Context) *core.WorkflowError {
					fmt.Println("Executing my-action")
					return nil
				}),
			},
			{
				ID:          "step-2",
				IgnoreError: false,
				Action: core.NewActionFunc("my-action-2", func(ctx context.Context) *core.WorkflowError {
					fmt.Println("Executing my-action 2")
					return &core.WorkflowError{
						Kind:    "internal",
						Code:    "not_implemented",
						Message: "this action is not implemented",
						Details: map[string]any{"action_id": "my-action-2"},
					}
					//return nil
				}),
			},
		},
	}, false)

	for _, event := range events {
		switch ev := event.(type) {
		case core.WorkflowEnabledEvent, core.WorkflowDisabledEvent, core.WorkflowTriggeredEvent:
			wf.Apply([]any{ev})
		}
	}

	return wf, nil
}

func (e EventStoreWorkflowRepository) Save(ctx context.Context, wf *core.Workflow) misas.Error {
	for _, event := range wf.UncommittedEvents() {
		if err := e.EventStore.Record(ctx, event); err != nil {
			return err
		}
	}

	wf.Commit()

	return nil
}
