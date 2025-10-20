package workflowmgmt

type StepID string

type StepDefinition struct {
	ID          StepID
	IgnoreError bool
	Action      Action
}

type ConcurrencyLimit int

const (
	ConcurrencyLimitNone ConcurrencyLimit = 0
)

type WorkflowDefinition struct {
	ID               WorkflowID
	ConcurrencyLimit ConcurrencyLimit
	Steps            []StepDefinition
}
