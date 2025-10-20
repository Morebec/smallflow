package workflowmgmt

import "github.com/morebec/go-misas/misas"

func NewWorkflowNotFoundError(workflowID any) misas.Error {
	switch id := workflowID.(type) {
	case WorkflowID:
		return ErrWorkflowNotFound.WithAppendedMessage(string(id))
	case string:
		return ErrWorkflowNotFound.WithAppendedMessage(id)
	}
	panic("invalid type for workflowID")
}

func WorkflowNotFoundCommandResult(workflowID any) misas.CommandResult {
	return misas.CommandResult{
		Error: NewWorkflowNotFoundError(workflowID),
	}
}
