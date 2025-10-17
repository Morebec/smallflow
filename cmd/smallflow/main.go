package main

import (
	"context"
	"fmt"
	"github.com/alitto/pond/v2"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
	"github.com/morebec/smallflow/internal/core"
	"github.com/morebec/smallflow/internal/core/adapters"
	"github.com/morebec/smallflow/internal/orchestrator"
	adapters2 "github.com/morebec/smallflow/internal/orchestrator/adapters"
	"time"
)

func main() {
	fmt.Println("Build, run, and observe workflows without the overhead!")

	eventStore := &adapters.InMemoryEventStore{}
	clock := mx.NewRealTimeClock(time.UTC)
	workflowRepo := &adapters.EventStoreWorkflowRepository{
		EventStore: eventStore,
	}
	runRepo := &adapters.EventStoreRunRepository{
		EventStore: eventStore,
	}

	api := core.NewSubsystem(clock, workflowRepo, runRepo, muuid.NewRandomUUIDGenerator()).API
	orch := &orchestrator.WorkflowOrchestrator{
		Clock: clock,
		API:   api,
		LeaseManager: orchestrator.WorkflowLeaseManager{
			Clock:      clock,
			Repository: adapters2.NewInMemoryWorkflowLeaseRepository(),
		},
		Pool: pond.NewPool(10),
	}

	ctx := context.Background()

	fmt.Println("Enabling workflow...")
	if result := api.HandleCommand(ctx, core.EnableWorkflowCommand{
		WorkflowID: "my-workflow",
	}); result.Error != nil {
		panic(result.Error)
	}

	fmt.Println("Triggering workflow...")
	if result := api.HandleCommand(ctx, core.TriggerWorkflowCommand{
		WorkflowID: "my-workflow",
		RunID:      muuid.NewRandomUUIDGenerator().Generate().String(),
	}); result.Error != nil {
		panic(result.Error)
	}

	fmt.Println("Disabling workflow...")
	if result := api.HandleCommand(ctx, core.DisableWorkflowCommand{
		WorkflowID: "my-workflow",
	}); result.Error != nil {
		panic(result.Error)
	}

	fmt.Println("Dispatching events through the orchestrator...")
	orch.Start()
	defer orch.Stop()

	for _, event := range eventStore.Events() {
		if err := orch.HandleEvent(ctx, event); err != nil {
			panic(err)
		}
	}

	for orch.IsRunning() {
		select {
		case <-time.After(15 * time.Second):
			fmt.Println("Stopping orchestrator after 15 seconds...")
			orch.Stop()
		}
	}

	fmt.Println("Current events in the event store:")
	for i, event := range eventStore.Events() {
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
