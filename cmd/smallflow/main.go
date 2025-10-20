package main

import (
	"context"
	"fmt"
	orchestrator2 "github.com/morebec/smallflow/internal/application/orchestrator"
	"github.com/morebec/smallflow/internal/business/workflowmgmt"
	"github.com/morebec/smallflow/internal/integration/postgres"
	"time"

	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/mpostgres"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
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

	mx.EventRegistry.Register(workflowmgmt.WorkflowEnabledEventTypeName, workflowmgmt.WorkflowEnabledEvent{})
	mx.EventRegistry.Register(workflowmgmt.WorkflowDisabledEventTypeName, workflowmgmt.WorkflowDisabledEvent{})
	mx.EventRegistry.Register(workflowmgmt.WorkflowTriggeredEventTypeName, workflowmgmt.WorkflowTriggeredEvent{})
	mx.EventRegistry.Register(workflowmgmt.WorkflowStartedEventTypeName, workflowmgmt.WorkflowStartedEvent{})
	mx.EventRegistry.Register(workflowmgmt.WorkflowEndedEventTypeName, workflowmgmt.WorkflowEndedEvent{})
	mx.EventRegistry.Register(workflowmgmt.StepStartedEventTypeName, workflowmgmt.StepStartedEvent{})
	mx.EventRegistry.Register(workflowmgmt.StepEndedEventTypeName, workflowmgmt.StepEndedEvent{})

	eventStore = mx.NewEventStoreDeserializerDecorator(eventStore)

	workflowRepo := &postgres.EventStoreWorkflowRepository{
		EventStore:    eventStore,
		UUIDGenerator: muuid.NewRandomUUIDGenerator(),
	}
	runRepo := &postgres.EventStoreRunRepository{
		EventStore:    eventStore,
		UUIDGenerator: muuid.NewRandomUUIDGenerator(),
	}

	api := workflowmgmt.NewSubsystem(clock, workflowRepo, runRepo, muuid.NewRandomUUIDGenerator()).API
	workflowLeaseRepository, err := postgres.NewWorkflowLeaseRepository(dbConn)
	if err != nil {
		panic(err)
	}

	leaseManager := orchestrator2.WorkflowLeaseManager{
		Clock:      clock,
		Repository: workflowLeaseRepository,
	}

	checkpointStore, err := mpostgres.NewPostgreSQLCheckpointStore(dbConn)
	if err != nil {
		panic(err)
	}
	orch := orchestrator2.NewWorkflowOrchestrator(
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
