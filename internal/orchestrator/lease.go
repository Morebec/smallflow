package orchestrator

import (
	"context"
	"github.com/morebec/go-misas/misas"
	"time"
)

const defaultLeaseDuration = 5 * time.Minute

type WorkflowLease struct {
	WorkflowID string
	RunID      string
	expiresAt  time.Time
}

func (l WorkflowLease) IsExpired(currentTime time.Time) bool { return currentTime.After(l.expiresAt) }

type WorkflowLeaseRepository interface {
	Add(context.Context, WorkflowLease) error
	Update(context.Context, WorkflowLease) error
	Remove(ctx context.Context, workflowID string, runID string) error
	FindByWorkflowRunID(ctx context.Context, workflowID string, runID string) (*WorkflowLease, error)
}

type WorkflowLeaseManager struct {
	Clock      misas.Clock
	Repository WorkflowLeaseRepository
}

func (l WorkflowLeaseManager) TryAcquire(ctx context.Context, workflowID string, runID string) (bool, error) {
	lease, err := l.Repository.FindByWorkflowRunID(ctx, workflowID, runID)
	if err != nil {
		return false, err
	}

	if lease != nil && !lease.IsExpired(l.Clock.Now()) {
		// Lease is still valid
		return false, nil
	}

	wl := WorkflowLease{
		WorkflowID: workflowID,
		RunID:      runID,
		expiresAt:  l.Clock.Now().Add(defaultLeaseDuration),
	}
	if err := l.Repository.Add(ctx, wl); err != nil {
		return false, err
	}

	return true, nil
}

func (l WorkflowLeaseManager) Release(ctx context.Context, workflowID string, runID string) error {
	if err := l.Repository.Remove(ctx, workflowID, runID); err != nil {
		return err
	}

	return nil
}

func (l WorkflowLeaseManager) RenewLease(ctx context.Context, workflowID string, runID string) error {
	if err := l.Repository.Update(ctx, WorkflowLease{
		WorkflowID: workflowID,
		RunID:      runID,
		expiresAt:  l.Clock.Now().Add(defaultLeaseDuration),
	}); err != nil {
		return err
	}

	return nil
}
