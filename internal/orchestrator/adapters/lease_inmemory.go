package adapters

import (
	"context"
	"github.com/morebec/smallflow/internal/orchestrator"
	"sync"
)

type InMemoryWorkflowLeaseRepository struct {
	mu     sync.Mutex
	leases map[string]orchestrator.WorkflowLease
}

func NewInMemoryWorkflowLeaseRepository() *InMemoryWorkflowLeaseRepository {
	return &InMemoryWorkflowLeaseRepository{leases: make(map[string]orchestrator.WorkflowLease)}
}

func (r *InMemoryWorkflowLeaseRepository) Add(_ context.Context, lease orchestrator.WorkflowLease) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leases[lease.WorkflowID+lease.RunID] = lease
	return nil
}

func (r *InMemoryWorkflowLeaseRepository) Update(_ context.Context, lease orchestrator.WorkflowLease) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leases[lease.WorkflowID+lease.RunID] = lease
	return nil
}

func (r *InMemoryWorkflowLeaseRepository) Remove(_ context.Context, workflowID string, runID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.leases, workflowID+runID)
	return nil
}

func (r *InMemoryWorkflowLeaseRepository) FindByWorkflowRunID(_ context.Context, workflowID string, runID string) (*orchestrator.WorkflowLease, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	lease, ok := r.leases[workflowID+runID]
	if !ok {
		return nil, nil
	}
	return &lease, nil
}
