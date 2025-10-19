package main

import (
	"context"
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mpostgres"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
	"github.com/morebec/smallflow/internal/orchestrator"
	adapters2 "github.com/morebec/smallflow/internal/orchestrator/adapters"
	"github.com/morebec/smallflow/internal/workflowmgmt"
	"github.com/morebec/smallflow/internal/workflowmgmt/adapters"
	"time"
)

func main() {
	fmt.Println("Build, run, and observe workflows without the overhead!")

	clock := mx.NewRealTimeClock(time.UTC)

	dbConn, err := mpostgres.OpenConn("postgres://smallflow:smallflow@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		panic(err)
	}

	var eventStore misas.EventStore
	eventStore, err = mpostgres.NewEventStore(clock, dbConn)
	if err != nil {
		panic(err)
	}

	eventRegistry := mx.NewMessageRegistry[misas.EventTypeName, misas.Event]()

	eventRegistry.Register(workflowmgmt.WorkflowEnabledEventTypeName, workflowmgmt.WorkflowEnabledEvent{})
	eventRegistry.Register(workflowmgmt.WorkflowDisabledEventTypeName, workflowmgmt.WorkflowDisabledEvent{})
	eventRegistry.Register(workflowmgmt.WorkflowTriggeredEventTypeName, workflowmgmt.WorkflowTriggeredEvent{})
	eventRegistry.Register(workflowmgmt.WorkflowStartedEventTypeName, workflowmgmt.WorkflowStartedEvent{})
	eventRegistry.Register(workflowmgmt.WorkflowEndedEventTypeName, workflowmgmt.WorkflowEndedEvent{})
	eventRegistry.Register(workflowmgmt.StepStartedEventTypeName, workflowmgmt.StepStartedEvent{})
	eventRegistry.Register(workflowmgmt.StepEndedEventTypeName, workflowmgmt.StepEndedEvent{})

	eventStore = mx.NewEventStoreDeserializerDecorator(eventStore, eventRegistry)

	workflowRepo := &adapters.EventStoreWorkflowRepository{
		EventStore:    eventStore,
		EventRegistry: eventRegistry,
		UUIDGenerator: muuid.NewRandomUUIDGenerator(),
	}
	runRepo := &adapters.EventStoreRunRepository{
		EventStore:    eventStore,
		EventRegistry: eventRegistry,
		UUIDGenerator: muuid.NewRandomUUIDGenerator(),
	}

	api := workflowmgmt.NewSubsystem(clock, workflowRepo, runRepo, muuid.NewRandomUUIDGenerator()).API
	workflowLeaseRepository, err := adapters2.NewPostgresWorkflowLeaseRepository(dbConn)
	if err != nil {
		panic(err)
	}

	leaseManager := orchestrator.WorkflowLeaseManager{
		Clock:      clock,
		Repository: workflowLeaseRepository,
	}

	checkpointStore, err := mpostgres.NewPostgreSQLCheckpointStore(dbConn)
	if err != nil {
		panic(err)
	}
	orch := orchestrator.NewWorkflowOrchestrator(
		clock,
		api,
		leaseManager,
		muuid.NewRandomUUIDGenerator(),
		eventStore,
		checkpointStore,
	)
	orch.Start()
	defer orch.Stop()

	ctx := context.Background()

	fmt.Println("Enabling workflow...")
	if result := api.HandleCommand(ctx, workflowmgmt.EnableWorkflowCommand{
		WorkflowID: "my-workflow",
	}); result.Error != nil {
		panic(result.Error)
	}

	for i := range 1 {
		fmt.Printf("Triggering workflow #%d...\n", i+1)
		if result := api.HandleCommand(ctx, workflowmgmt.TriggerWorkflowCommand{
			WorkflowID: "my-workflow",
			RunID:      muuid.NewRandomUUIDGenerator().Generate().String(),
		}); result.Error != nil {
			panic(result.Error)
		}
	}

	<-time.After(30 * time.Second)
	fmt.Println("Current events in the event store:")
	stream, err := eventStore.ReadFromStream(ctx, eventStore.GlobalStreamID(), misas.ReadFromEventStreamOptions{}.FromStart().Forward())
	if err != nil {
		panic(err)
	}

	for i, event := range stream.Events {
		fmt.Printf("Event %d: %T → %+v\n", i, event, event)
	}
}

func registerActions() {
	//actionRegistry := definition.ActionRegistry{}
	//actionRegistry.Register(definition.NewActionFunc("my_action", func(ctx definition.Action) *definition.ActionError {
	//	fmt.Println("Hello, World!")
	//	return nil
	//}))
}
