package orchestrator

import (
	"context"
	"fmt"
	"github.com/alitto/pond/v2"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/smallflow/internal/core"
)

type WorkflowOrchestrator struct {
	Clock        misas.Clock
	API          misas.BusinessAPI
	LeaseManager WorkflowLeaseManager
	Pool         pond.Pool

	ctx       context.Context
	cancelCtx context.CancelFunc
}

func (m *WorkflowOrchestrator) Start() {
	m.ctx, m.cancelCtx = context.WithCancel(context.Background())
}

func (m *WorkflowOrchestrator) Stop() { m.cancelCtx() }

func (m *WorkflowOrchestrator) IsRunning() bool {
	if m.ctx == nil {
		return false
	}
	return m.ctx.Err() == nil
}

func (m *WorkflowOrchestrator) HandleEvent(ctx context.Context, event any) error {
	switch e := event.(type) {
	case core.WorkflowTriggeredEvent:
		return m.runWorkflow(ctx, e)
	}

	return nil
}

func (m *WorkflowOrchestrator) runWorkflow(_ context.Context, e core.WorkflowTriggeredEvent) error {
	pond.Submit(func() {
		runner := Worker{
			Clock:        m.Clock,
			API:          m.API,
			LeaseManager: m.LeaseManager,
			InstanceID:   "instance-1", // TODO: generate unique instance ID
		}
		if err := runner.Run(context.Background(), e.WorkflowID, e.RunID); err != nil {
			fmt.Println("workflow runner error:", err)
		}
	})
	return nil
}
