package main

import (
	"context"
	"fmt"
	"github.com/morebec/smallflow/internal/core"
	"github.com/morebec/smallflow/internal/core/adapters"
	"github.com/morebec/smallflow/internal/orchestrator"
)

func main() {
	fmt.Println("Build, run, and observe workflows without the overhead!")

	eventStore := &adapters.InMemoryEventStore{}
	clock := adapters.RealTimeClock{}
	workflowRepo := &adapters.EventStoreWorkflowRepository{
		EventStore: eventStore,
	}
	runRepo := &adapters.EventStoreRunRepository{
		EventStore: eventStore,
	}

	api := core.NewAPI(clock, workflowRepo, runRepo)
	orch := orchestrator.WorkflowOrchestrator{Clock: clock, API: api}

	ctx := context.Background()

	fmt.Println("Enabling workflow...")
	if err := api.HandleCommand(ctx, core.EnableWorkflowCommand{
		WorkflowID: "my-workflow",
	}); err != nil {
		panic(err)
	}

	fmt.Println("Triggering workflow...")
	if err := api.HandleCommand(ctx, core.TriggerWorkflowCommand{
		WorkflowID: "my-workflow",
		RunID:      "run-1",
	}); err != nil {
		panic(err)
	}

	fmt.Println("Triggering workflow...")
	if err := api.HandleCommand(ctx, core.TriggerWorkflowCommand{
		WorkflowID: "my-workflow",
		RunID:      "run-1",
	}); err != nil {
		panic(err)
	}

	fmt.Println("Disabling workflow...")
	if err := api.HandleCommand(ctx, core.DisableWorkflowCommand{
		WorkflowID: "my-workflow",
	}); err != nil {
		panic(err)
	}

	fmt.Println("Dispatching events through the orchestrator...")
	for _, event := range eventStore.Events() {
		if err := orch.HandleEvent(ctx, event); err != nil {
			panic(err)
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
