package postgres

import (
	"context"
	"github.com/morebec/smallflow/internal/application/orchestrator"

	"github.com/morebec/go-misas/mpostgres"
)

type WorkflowLeaseRepository struct {
	conn       mpostgres.DB
	collection mpostgres.Collection
}

func NewWorkflowLeaseRepository(conn mpostgres.DB) (WorkflowLeaseRepository, error) {
	ctx := context.Background()
	docStore, err := mpostgres.NewDocumentStore(ctx, conn)
	if err != nil {
		return WorkflowLeaseRepository{}, err
	}

	collection, err := docStore.Collection("workflow_leases")
	if err != nil {
		return WorkflowLeaseRepository{}, err
	}
	if err := collection.Create(ctx); err != nil {
		return WorkflowLeaseRepository{}, err
	}

	return WorkflowLeaseRepository{conn: conn, collection: collection}, nil
}

func (r WorkflowLeaseRepository) Add(ctx context.Context, lease orchestrator.WorkflowLease) error {
	doc, err := mpostgres.NewDocument(r.workflowLeasID(lease.WorkflowID, lease.RunID), lease)
	if err != nil {
		return err
	}

	if err := r.collection.Add(ctx, doc); err != nil {
		return err
	}

	return nil
}

func (r WorkflowLeaseRepository) workflowLeasID(workflowID, runID string) string {
	return workflowID + "/" + runID
}

func (r WorkflowLeaseRepository) Update(ctx context.Context, lease orchestrator.WorkflowLease) error {
	doc, err := mpostgres.NewDocument(r.workflowLeasID(lease.WorkflowID, lease.RunID), lease)
	if err != nil {
		return err
	}

	if _, err := r.collection.Update(ctx, doc); err != nil {
		return err
	}

	return nil
}

func (r WorkflowLeaseRepository) Remove(ctx context.Context, workflowID string, runID string) error {
	if _, err := r.collection.RemoveByID(ctx, r.workflowLeasID(workflowID, runID)); err != nil {
		return err
	}

	return nil
}

func (r WorkflowLeaseRepository) FindByWorkflowRunID(ctx context.Context, workflowID string, runID string) (*orchestrator.WorkflowLease, error) {
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
