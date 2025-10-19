subsystem "workflowmgmt" {
  description = "Business subsystem responsible for core business logic related to managing workflows."
  type        = "business"
}


command "workflowmgmt.EnableWorkflow" {
  description = "EnableWorkflowCommand represents a command to enable a workflow."
  field "WorkflowID" {
    description = "ID of the workflow to enable."
    type        = "identifier"
  }
}

event "workflowmgmt.WorkflowEnabled" {
  description = "Event emitted when a workflow is successfully enabled."
  field "WorkflowID" {
    description = "ID of the workflow that was enabled."
    type        = "identifier"
  }
  field "EnabledAt" {
    description = "Timestamp when the workflow was enabled."
    type        = "datetime"
  }
}

command "workflowmgmt.DisableWorkflow" {
  description = "DisableWorkflowCommand represents a command to disable a workflow."
  field "WorkflowID" {
    description = "ID of the workflow to disable."
    type        = "identifier"
  }
}

event "workflowmgmt.WorkflowDisabled" {
  description = "Event emitted when a workflow is successfully disabled."
  field "WorkflowID" {
    description = "ID of the workflow that was disabled."
    type        = "identifier"
  }
  field "DisabledAt" {
    description = "Timestamp when the workflow was disabled."
    type        = "datetime"
  }
}

command "workflowmgmt.TriggerWorkflow" {
  description = <<EOT
TriggerWorkflowCommand represents a command to trigger a new run of a
workflow.

This command is idempotent based on the RunID provided. If a run
with the same RunID already exists for the given workflow, this command will
have no effect and will succeed.

If no RunID is provided, a new unique RunID
will be generated.

The workflow must be enabled for this command to succeed.
If the workflow is disabled, this command will fail.
If the workflow does not
exist, this command will fail. If the concurrent runs limit has been reached,
this command will fail.
EOT
  field "WorkflowID" {
    description = "ID of the workflow to trigger."
    type        = "identifier"
  }
  field "RunID" {
    description = "Optional ID for the workflow run. If not provided, a new unique ID will be generated."
    type        = "identifier"
    required    = false
  }
}

event "workflowmgmt.WorkflowTriggered" {
  description = "Event emitted when a workflow is successfully triggered."
  field "WorkflowID" {
    description = "ID of the workflow that was triggered."
    type        = "identifier"
  }
  field "RunID" {
    description = "ID of the workflow run that was created."
    type        = "identifier"
  }
  field "TriggeredAt" {
    description = "Timestamp when the workflow was triggered."
    type        = "datetime"
  }
}


command "workflowmgmt.RunWorkflow" {
  description = <<EOT
Runs a workflow synchronously.
This command is intended to be run after a workflow has been triggered, and will fail
if no such workflow run exists.

This command is idempotent based on the RunID provided:
  - If a run with the same RunID already exists for the given workflow, this command will
    have no effect and will succeed.
  - If no RunID is provided, this command will fail.

This command will not fail if the workflow or any of its steps fail, as these
are expected outcomes of a workflow run. Instead, the errors will be recorded
in the StepEndedEvent and WorkflowEndedEvent.
This command will only return errors if internal errors have occurred while attempting
to run the workflow, outside of the workflow's own logic.

EOT

    field "WorkflowID" {
        description = "ID of the workflow to run."
        type        = "identifier"
    }
    field "RunID" {
        description = "ID of the workflow run to execute."
        type        = "identifier"
    }
}

event "workflowmgmt.WorkflowStarted" {
  description = "Event emitted when a workflow run is started."
  result = "workflowmgmt.WorkflowRunReport"
  field "WorkflowID" {
    description = "ID of the workflow that is being run."
    type        = "identifier"
  }
  field "RunID" {
    description = "ID of the workflow run that was started."
    type        = "identifier"
  }
  field "StartedAt" {
    description = "Timestamp when the workflow run was started."
    type        = "datetime"
  }
}

event "workflowmgmt.WorkflowEnded" {
  description = "Event emitted when a workflow run is ended."
  result = "workflowmgmt.WorkflowRunReport"
  field "WorkflowID" {
    description = "ID of the workflow that was run."
    type        = "identifier"
  }
  field "RunID" {
    description = "ID of the workflow run that was ended."
    type        = "identifier"
  }
  field "EndedAt" {
    description = "Timestamp when the workflow run ended."
    type        = "datetime"
  }
  field "Status" {
    description = "Final status of the workflow run (e.g., Success, Failed)."
    type        = "string"
  }
  field "Errors" {
    description = <<-EOT
List of errors encountered during the workflow run at every step.
This can include errors from steps that still succeeded because of their ignore errors setting.
EOT
    type        = "map[identifier]workflowmgmt.WorkflowError"
  }
}

event "workflowmgmt.StepStarted" {
  description = "Event emitted when a workflow step is started."
  field "WorkflowID" {
    description = "ID of the workflow that contains the step."
    type        = "identifier"
  }
  field "RunID" {
    description = "ID of the workflow run."
    type        = "identifier"
  }
  field "StepID" {
    description = "ID of the step that was started."
    type        = "identifier"
  }
  field "StartedAt" {
    description = "Timestamp when the step was started."
    type        = "datetime"
  }
}

event "workflowmgmt.StepEnded" {
  description = "Event emitted when a workflow step is ended."
  field "WorkflowID" {
    description = "ID of the workflow that contains the step."
    type        = "identifier"
  }
  field "RunID" {
    description = "ID of the workflow run."
    type        = "identifier"
  }
  field "StepID" {
    description = "ID of the step that was ended."
    type        = "identifier"
  }
  field "EndedAt" {
    description = "Timestamp when the step ended."
    type        = "datetime"
  }
  field "Status" {
    description = "Final status of the step (e.g., Success, Failed)."
    type        = "string"
  }
  field "Error" {
    description = "Error encountered during the step, if any."
    type        = "workflowmgmt.WorkflowError"
    required    = false
  }
}


struct "workflowmgmt.WorkflowRunReport" {
  description = "Report generated after a workflow run is completed."
  field "WorkflowID" {
    description = "ID of the workflow that was run."
    type        = "identifier"
  }
  field "RunID" {
    description = "ID of the workflow run."
    type        = "identifier"
  }
  field "StartedAt" {
    description = "Timestamp when the workflow run started."
    type        = "datetime"
  }
  field "EndedAt" {
    description = "Timestamp when the workflow run ended."
    type        = "datetime"
  }
  field "Status" {
    description = "Final status of the workflow run (e.g., Success, Failed)."
    type        = "string"
  }
  field "Errors" {
    description = <<-EOT
List of errors encountered during the workflow run at every step.
This can include errors from steps that still succeeded because of their ignore errors setting.
EOT
    type        = "map[identifier]string"
  }
}

struct "workflowmgmt.WorkflowError" {
  description = "Represents an error encountered during a workflow run."
  field "Kind" {
    description = "Kind of error. (e.g. internal, timeout, not_found, unauthorized, etc.) See the MISAS error documentation for more details."
    type        = "string"
  }
  field "Code" {
    description = "Specific error code. See the MISAS error code documentation for more details."
    type        = "string"
  }
  field "Message" {
    description = "Detailed error message."
    type        = "string"
  }
  fiel "Details" {
    description = "Additional details about the error."
    type        = "map[string]any"
  }
}

type "workflowmgmt.WorkflowState" {
  description = "Represents a workflow definition."
  field "ID" {
    description = "Unique identifier for the workflow."
    type        = "identifier"
  }
}


service_provider "workflowmgmt.WorkflowRepository" {
  description = "Service provider responsible for persisting and retrieving workflows, their definition and their state."
  operation "FindByID" {
    description = "Finds a workflow by its ID."
    field "workflowID" {
      description = "ID of the workflow to find."
      type        = "identifier"
    }
    result = "workflowmgmt.Workflow"
  }
  operation "Save" {
    description = "Saves a workflow in this store."
    field "workflow" {
      description = "Workflow to save."
        type        = "workflowmgmt.WorkflowState"
      type        = "[]byte"
    }
  }
}
