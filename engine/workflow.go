package engine

import (
	"fmt"
)

type WorkflowFunc func(*Context) error

type WorkflowRunner struct {
	persistence *PersistenceLayer
}

func NewWorkflowRunner(dbPath string) (*WorkflowRunner, error) {
	persistence, err := NewPersistence(dbPath)
	if err != nil {
		return nil, err
	}
	return &WorkflowRunner{persistence: persistence}, nil
}

func (r *WorkflowRunner) Run(workflowID string, workflow WorkflowFunc) error {
	if err := r.persistence.CreateWorkflow(workflowID); err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	ctx := NewContext(workflowID, r.persistence)

	fmt.Printf("\n=== Starting workflow: %s ===\n\n", workflowID)

	if err := workflow(ctx); err != nil {
		return fmt.Errorf("workflow failed: %w", err)
	}

	fmt.Printf("\n=== Workflow completed successfully ===\n\n")
	return nil
}

func (r *WorkflowRunner) Close() error {
	return r.persistence.Close()
}
