package adapters

import (
	"context"
	"fmt"

	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
	"github.com/morebec/smallflow/internal/workflowmgmt"
	"github.com/samber/lo"
)

type EventStoreWorkflowRepository struct {
	EventStore    misas.EventStore
	UUIDGenerator muuid.UUIDGenerator
}

func (r EventStoreWorkflowRepository) FindByID(ctx context.Context, workflowID string) (*workflowmgmt.Workflow, misas.Error) {
	stream, err := r.EventStore.ReadFromStream(
		ctx,
		misas.EventStreamID("workflows/"+workflowID),
		misas.ReadFromEventStreamOptions{}.
			FromStart().
			Forward(),
	)
	if err != nil {
		if !misas.ErrorHasCode(err, misas.ErrEventStreamNotFoundErrorCode) {
			return nil, mx.NewInternalErrorFrom(err)
		}
	}

	wf := workflowmgmt.NewWorkflow(workflowmgmt.WorkflowDefinition{
		ID:               workflowmgmt.WorkflowID(workflowID),
		ConcurrencyLimit: workflowmgmt.ConcurrencyLimitNone,
		Steps: []workflowmgmt.StepDefinition{
			{
				ID:          "step-1",
				IgnoreError: false,
				Action: workflowmgmt.NewActionFunc("my-action", func(ctx context.Context) *workflowmgmt.WorkflowError {
					fmt.Println("Executing my-action")
					return nil
				}),
			},
			{
				ID:          "step-2",
				IgnoreError: false,
				Action: workflowmgmt.NewActionFunc("my-action-2", func(ctx context.Context) *workflowmgmt.WorkflowError {
					fmt.Println("Executing my-action 2")
					return &workflowmgmt.WorkflowError{
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

	var events []misas.Event
	for _, record := range stream.Events {
		events = append(events, record.Data)
	}

	wf.Apply(events)

	return wf, nil
}

func (r EventStoreWorkflowRepository) Save(ctx context.Context, wf *workflowmgmt.Workflow) misas.Error {
	descriptors := lo.Map(wf.UncommittedEvents(), func(event misas.Event, _ int) misas.EventDescriptor {
		return misas.EventDescriptor{
			ID:       r.UUIDGenerator.Generate().String(),
			TypeName: event.TypeName(),
			Data:     event,
		}
	})

	err := r.EventStore.AppendToStream(
		ctx,
		misas.EventStreamID("workflows/"+wf.ID()),
		descriptors,
		misas.AppendToEventStreamOptions{}, //TODO expected versioning
	)
	if err != nil {
		return mx.NewInternalErrorFrom(err)
	}

	wf.Commit()

	return nil
}
