package orchestrator

import (
	"context"
	"fmt"
	"github.com/morebec/smallflow/internal/business/workflowmgmt"

	"github.com/alitto/pond/v2"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/go-misas/muuid"
	"github.com/morebec/go-misas/mx"
)

const defaultMaxConcurrentWorkers = 1000

type WorkflowOrchestrator struct {
	Clock         misas.Clock
	API           misas.BusinessAPI
	LeaseManager  WorkflowLeaseManager
	UUIDGenerator muuid.UUIDGenerator
	Pool          pond.Pool

	eventProcessor mx.EventProcessor
}

func NewWorkflowOrchestrator(
	clock misas.Clock,
	API misas.BusinessAPI,
	leaseManager WorkflowLeaseManager,
	uuidGenerator muuid.UUIDGenerator,
	eventStore misas.EventStore,
	checkpointStore misas.CheckpointStore,
) *WorkflowOrchestrator {
	wo := &WorkflowOrchestrator{
		Clock:         clock,
		API:           API,
		LeaseManager:  leaseManager,
		UUIDGenerator: uuidGenerator,
		Pool:          pond.NewPool(defaultMaxConcurrentWorkers),
	}

	wo.eventProcessor = *mx.NewEventProcessor(mx.EventProcessorConfig{
		ID:              "workflow-orchestrator",
		StreamID:        eventStore.GlobalStreamID(),
		EventStore:      eventStore,
		CheckpointStore: checkpointStore,
		CommitStrategy:  misas.CheckpointCommitStrategyAfterProcessing,
		Handler:         wo,
	})

	return wo
}

func (m *WorkflowOrchestrator) Start() {
	err := m.eventProcessor.Start()
	if err != nil {
		panic(err)
	}
}

func (m *WorkflowOrchestrator) Stop() {
	m.eventProcessor.Stop()
}

func (m *WorkflowOrchestrator) IsRunning() bool {
	return m.eventProcessor.IsRunning()
}

func (m *WorkflowOrchestrator) HandleEvent(ctx context.Context, event misas.Event) misas.Error {
	switch e := event.(type) {
	case workflowmgmt.WorkflowTriggeredEvent:
		m.runWorkflow(ctx, e)
		return nil
	}

	return nil
}

func (m *WorkflowOrchestrator) runWorkflow(_ context.Context, e workflowmgmt.WorkflowTriggeredEvent) {
	m.Pool.Submit(func() {
		runner := Worker{
			Clock:        m.Clock,
			API:          m.API,
			LeaseManager: m.LeaseManager,
			InstanceID:   m.UUIDGenerator.Generate().String(),
		}
		if err := runner.Run(context.Background(), e.WorkflowID, e.RunID); err != nil {
			fmt.Println("workflow runner error:", err)
		}
	})
}
