package adapters

import (
	"context"

	"github.com/morebec/go-misas/mpostgres"
	"github.com/morebec/smallflow/internal/orchestrator"
)

type PostgresWorkflowLeaseRepository struct {
	conn       mpostgres.DB
	collection mpostgres.Collection
}

func NewPostgresWorkflowLeaseRepository(conn mpostgres.DB) (PostgresWorkflowLeaseRepository, error) {
	ctx := context.Background()
	docStore, err := mpostgres.NewDocumentStore(ctx, conn)
	if err != nil {
		return PostgresWorkflowLeaseRepository{}, err
	}

	collection, err := docStore.Collection("workflow_leases")
	if err != nil {
		return PostgresWorkflowLeaseRepository{}, err
	}
	if err := collection.Create(ctx); err != nil {
		return PostgresWorkflowLeaseRepository{}, err
	}

	return PostgresWorkflowLeaseRepository{conn: conn, collection: collection}, nil
}

func (r PostgresWorkflowLeaseRepository) Add(ctx context.Context, lease orchestrator.WorkflowLease) error {
	doc, err := mpostgres.NewDocument(r.workflowLeasID(lease.WorkflowID, lease.RunID), lease)
	if err != nil {
		return err
	}

	if err := r.collection.Add(ctx, doc); err != nil {
		return err
	}

	return nil
}

func (r PostgresWorkflowLeaseRepository) workflowLeasID(workflowID, runID string) string {
	return workflowID + "/" + runID
}

func (r PostgresWorkflowLeaseRepository) Update(ctx context.Context, lease orchestrator.WorkflowLease) error {
	doc, err := mpostgres.NewDocument(r.workflowLeasID(lease.WorkflowID, lease.RunID), lease)
	if err != nil {
		return err
	}

	if _, err := r.collection.Update(ctx, doc); err != nil {
		return err
	}

	return nil
}

func (r PostgresWorkflowLeaseRepository) Remove(ctx context.Context, workflowID string, runID string) error {
	if _, err := r.collection.RemoveByID(ctx, r.workflowLeasID(workflowID, runID)); err != nil {
		return err
	}

	return nil
}

func (r PostgresWorkflowLeaseRepository) FindByWorkflowRunID(ctx context.Context, workflowID string, runID string) (*orchestrator.WorkflowLease, error) {
	doc, err := r.collection.FindByID(ctx, r.workflowLeasID(workflowID, runID))
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}

	var wl *orchestrator.WorkflowLease
	if err := doc.Unmarshal(&wl); err != nil {
		return nil, err
	}

	return wl, nil
}
