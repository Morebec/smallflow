package adapters

import (
	"context"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
	"github.com/morebec/smallflow/internal/workflowmgmt"
	"github.com/samber/lo"
)

type EventStoreRunRepository struct {
	UUIDGenerator muuid.UUIDGenerator
	EventStore    misas.EventStore
	EventRegistry *mx.MessageRegistry[misas.EventTypeName, misas.Event]
}

func (r EventStoreRunRepository) Add(ctx context.Context, run *workflowmgmt.Run) misas.Error {
	descriptors := lo.Map(run.UncommittedEvents(), func(event misas.Event, _ int) misas.EventDescriptor {
		return misas.EventDescriptor{
			ID:       r.UUIDGenerator.Generate().String(),
			TypeName: event.TypeName(),
			Data:     event,
		}
	})

	err := r.EventStore.AppendToStream(
		ctx,
		misas.EventStreamID("runs/"+run.ID),
		descriptors,
		misas.AppendToEventStreamOptions{}.
			ExpectNotExist(),
	)
	if err != nil {
		return mx.NewInternalErrorFrom(err)
	}

	run.Commit()

	return nil

}

func (r EventStoreRunRepository) Save(ctx context.Context, run *workflowmgmt.Run) misas.Error {
	descriptors := lo.Map(run.UncommittedEvents(), func(event misas.Event, _ int) misas.EventDescriptor {
		return misas.EventDescriptor{
			ID:       r.UUIDGenerator.Generate().String(),
			TypeName: event.TypeName(),
			Data:     event,
		}
	})

	err := r.EventStore.AppendToStream(
		ctx,
		misas.EventStreamID("runs/"+run.ID),
		descriptors,
		misas.AppendToEventStreamOptions{}, //TODO expected versioning
	)
	if err != nil {
		return mx.NewInternalErrorFrom(err)
	}

	run.Commit()

	return nil
}

func (r EventStoreRunRepository) FindByID(ctx context.Context, workflowID string, runID string) (*workflowmgmt.Run, misas.Error) {
	var run *workflowmgmt.Run

	stream, err := r.EventStore.ReadFromStream(
		ctx,
		misas.EventStreamID("runs/"+runID),
		misas.ReadFromEventStreamOptions{}.
			FromStart().
			Forward(),
	)
	if err != nil {
		if misas.ErrorHasCode(err, misas.ErrEventStreamNotFoundErrorCode) {
			return nil, nil
		}

		return nil, mx.NewInternalErrorFrom(err)
	}

	run = &workflowmgmt.Run{
		ID:         workflowmgmt.RunID(runID),
		WorkflowID: workflowmgmt.WorkflowID(workflowID),
	}

	var events []misas.Event
	for _, record := range stream.Events {
		events = append(events, record.Data)
	}

	run.Apply(events)

	return run, nil

}
