package services

import (
	"encoding/json"
	"fmt"
)

// WorkflowLifecycleResult is the outcome of enforcing workflow rules during a note mutation.
type WorkflowLifecycleResult struct {
	Enforced    bool                 `json:"enforced"`
	Allowed     bool                 `json:"allowed"`
	WorkflowID  string               `json:"workflow_id,omitempty"`
	GroupName   string               `json:"group_name,omitempty"`
	FromState   string               `json:"from_state,omitempty"`
	ToState     string               `json:"to_state,omitempty"`
	ResultState string               `json:"result_state,omitempty"`
	Diagnostics []WorkflowDiagnostic `json:"diagnostics,omitempty"`
}

// EnforceWorkflowOnMutation resolves the workflow for a note path and evaluates whether
// the mutation (create or edit) is allowed under the assigned workflow rules.
//
// For creates, fromState is "" and toState is the value of the workflow field in the
// provided metadata (defaulting to workflow initial_state if absent).
//
// For edits, fromState is the value of the workflow field in existingMetadata and
// toState is the value in newMetadata.
//
// If no workflow matches the note path, the mutation is allowed with enforced=false.
func EnforceWorkflowOnMutation(
	notePath string,
	groups []NotebookGroup,
	workflows map[string]json.RawMessage,
	existingMetadata map[string]any,
	newMetadata map[string]any,
	isCreate bool,
) WorkflowLifecycleResult {

	assignment := ResolveWorkflowAssignment(notePath, groups, workflows)
	if !assignment.Resolved {
		// Check if it's a "not found" case (no workflow applies) vs an error case
		if len(assignment.Diagnostics) > 0 && assignment.Diagnostics[0].Code == "workflow.assignment_not_found" {
			// No workflow applies — mutation allowed without enforcement
			return WorkflowLifecycleResult{
				Enforced: false,
				Allowed:  true,
			}
		}
		// Assignment error (conflict, unknown workflow, etc.) — block mutation
		return WorkflowLifecycleResult{
			Enforced:    true,
			Allowed:     false,
			WorkflowID:  assignment.WorkflowID,
			GroupName:   assignment.GroupName,
			Diagnostics: assignment.Diagnostics,
		}
	}

	def := assignment.Workflow
	field := def.Field

	// Determine from/to states
	var fromState, toState string

	if isCreate {
		// For creates: from is initial_state, to is metadata[field] or initial_state
		fromState = def.InitialState
		toState = metadataStringValue(newMetadata, field)
		if toState == "" {
			toState = def.InitialState
		}
		// Self-transition on create with initial state is always valid
		if fromState == toState {
			return WorkflowLifecycleResult{
				Enforced:    true,
				Allowed:     true,
				WorkflowID:  assignment.WorkflowID,
				GroupName:   assignment.GroupName,
				FromState:   fromState,
				ToState:     toState,
				ResultState: toState,
			}
		}
	} else {
		// For edits: from is existing value, to is new value
		fromState = metadataStringValue(existingMetadata, field)
		toState = metadataStringValue(newMetadata, field)
		if fromState == "" {
			fromState = def.InitialState
		}
		if toState == "" {
			toState = fromState
		}
		// No state change on edit — allow without evaluation
		if fromState == toState {
			return WorkflowLifecycleResult{
				Enforced:    true,
				Allowed:     true,
				WorkflowID:  assignment.WorkflowID,
				GroupName:   assignment.GroupName,
				FromState:   fromState,
				ToState:     toState,
				ResultState: toState,
			}
		}
	}

	// Evaluate the transition
	evalResult := EvaluateWorkflow(def, WorkflowExecutionRequest{
		Mode:      "apply",
		FromState: fromState,
		ToState:   toState,
		Metadata:  newMetadata,
	})

	result := WorkflowLifecycleResult{
		Enforced:    true,
		Allowed:     evalResult.Allowed,
		WorkflowID:  assignment.WorkflowID,
		GroupName:   assignment.GroupName,
		FromState:   fromState,
		ToState:     toState,
		ResultState: evalResult.ResultState,
		Diagnostics: evalResult.Diagnostics,
	}

	if !evalResult.Allowed {
		// Prepend a lifecycle-level diagnostic
		lifecycleDiag := WorkflowDiagnostic{
			Code:    "workflow.lifecycle_blocked",
			Message: fmt.Sprintf("workflow %s blocked mutation: transition %s -> %s not allowed", assignment.WorkflowID, fromState, toState),
			Path:    notePath,
		}
		result.Diagnostics = append([]WorkflowDiagnostic{lifecycleDiag}, result.Diagnostics...)
	}

	return result
}

// metadataStringValue extracts a string value from a metadata map.
func metadataStringValue(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	v, ok := metadata[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}
