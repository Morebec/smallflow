package orchestrator

import (
	"context"
	"fmt"
	"github.com/morebec/go-misas/misas"
	"github.com/morebec/smallflow/internal/workflowmgmt"
	"time"
)

const defaultLeaseHeartbeatDuration = time.Second * 30

type Worker struct {
	Clock        misas.Clock
	API          misas.BusinessAPI
	LeaseManager WorkflowLeaseManager
	InstanceID   string
}

func (r Worker) acquireLease(ctx context.Context, workflowID, runID string) (context.Context, context.CancelFunc, error) {
	acquired, err := r.LeaseManager.TryAcquire(ctx, workflowID, runID)
	if err != nil {
		return nil, nil, err
	}

	if !acquired {
		// Could not acquire lease, another runner instance is likely handling this workflow run.
		fmt.Printf("Could not acquire lease for workflow run: {workflowID: %s, runID: %s}\n", workflowID, runID)
		return nil, nil, nil
	}
	fmt.Printf("Lease acquired for workflow run: {workflowID: %s, runID: %s}\n", workflowID, runID)

	ctx, cancel := context.WithCancel(ctx)

	ticker := time.NewTicker(defaultLeaseHeartbeatDuration)
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := r.LeaseManager.RenewLease(ctx, workflowID, runID); err != nil {
					fmt.Printf(
						"Failed to renew lease for workflow run: {workflowID: %s, runID: %s}: %s\n",
						workflowID,
						runID,
						err,
					)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	stopHeartbeat := func() {
		ticker.Stop()
		cancel()
		// Release the lease when stopping the heartbeat
		// use context.Background() given we just canceled the original context.
		if err := r.LeaseManager.Release(context.Background(), workflowID, runID); err != nil {
			fmt.Printf(
				"Failed to release lease for workflow run: {workflowID: %s, runID: %s}: %s\n",
				workflowID,
				runID,
				err,
			)
		}
		fmt.Printf("Lease released for workflow run: {workflowID: %s, runID: %s}\n", workflowID, runID)
	}

	return ctx, stopHeartbeat, nil
}

func (r Worker) Run(ctx context.Context, workflowID, runID string) error {
	ctx, releaseLease, err := r.acquireLease(ctx, workflowID, runID)
	if err != nil {
		return err
	}
	if releaseLease == nil {
		// Lease not acquired, another runner is handling this workflow run.
		return nil
	}
	defer releaseLease()

	result := r.API.HandleCommand(ctx, workflowmgmt.RunWorkflowCommand{
		WorkflowID: workflowID,
		RunID:      runID,
	})

	return result.Error
}
