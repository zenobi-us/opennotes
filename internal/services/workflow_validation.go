package services

import "fmt"

// WorkflowDefinition defines one workflow contract from .jot.json workflows.<key>.
type WorkflowDefinition struct {
	Description  string                             `json:"description"`
	InitialState string                             `json:"initial_state"`
	Mode         string                             `json:"mode"`
	Field        string                             `json:"field"`
	States       map[string]WorkflowStateDefinition `json:"states"`
}

// WorkflowStateDefinition captures one workflow state contract.
type WorkflowStateDefinition struct {
	Schema      map[string]any `json:"schema"`
	Transitions []string       `json:"transitions"`
}

// WorkflowDiagnostic is a stable machine-readable validation issue.
type WorkflowDiagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

// WorkflowValidationResult contains validation outcome and diagnostics.
type WorkflowValidationResult struct {
	Valid       bool                 `json:"valid"`
	Diagnostics []WorkflowDiagnostic `json:"diagnostics,omitempty"`
}

// WorkflowExecutionRequest defines one workflow evaluation request.
type WorkflowExecutionRequest struct {
	Mode      string         `json:"mode"`
	FromState string         `json:"from_state"`
	ToState   string         `json:"to_state"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// WorkflowExecutionResult defines stable machine-readable workflow evaluation output.
type WorkflowExecutionResult struct {
	Valid       bool                 `json:"valid"`
	Allowed     bool                 `json:"allowed"`
	Applied     bool                 `json:"applied"`
	Mode        string               `json:"mode"`
	FromState   string               `json:"from_state"`
	ToState     string               `json:"to_state"`
	ResultState string               `json:"result_state"`
	Diagnostics []WorkflowDiagnostic `json:"diagnostics,omitempty"`
}

// EvaluateWorkflow validates and evaluates workflow execution behavior for apply/dry-run modes.
func EvaluateWorkflow(def WorkflowDefinition, req WorkflowExecutionRequest) WorkflowExecutionResult {
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}

	if req.Mode != "apply" && req.Mode != "dry-run" {
		return WorkflowExecutionResult{
			Valid:       false,
			Allowed:     false,
			Applied:     false,
			Mode:        req.Mode,
			FromState:   req.FromState,
			ToState:     req.ToState,
			ResultState: req.FromState,
			Diagnostics: []WorkflowDiagnostic{{
				Code:    "workflow.invalid_execution_mode",
				Message: fmt.Sprintf("unknown workflow execution mode: %s", req.Mode),
				Path:    "mode",
			}},
		}
	}

	validation := ValidateWorkflowTransition(def, req.FromState, req.ToState, req.Metadata)
	result := WorkflowExecutionResult{
		Valid:       validation.Valid,
		Allowed:     validation.Valid,
		Applied:     false,
		Mode:        req.Mode,
		FromState:   req.FromState,
		ToState:     req.ToState,
		ResultState: req.FromState,
		Diagnostics: validation.Diagnostics,
	}

	if req.Mode == "apply" {
		if validation.Valid {
			result.Applied = true
			result.ResultState = req.ToState
		} else {
			result.Allowed = false
		}
	}

	return result
}

// ValidateWorkflowTransition validates transition legality and target-state required metadata.
func ValidateWorkflowTransition(def WorkflowDefinition, fromState, toState string, metadata map[string]any) WorkflowValidationResult {
	diagnostics := make([]WorkflowDiagnostic, 0)

	from, fromOK := def.States[fromState]
	if !fromOK {
		diagnostics = append(diagnostics, WorkflowDiagnostic{
			Code:    "workflow.unknown_state",
			Message: fmt.Sprintf("unknown source state: %s", fromState),
			Path:    fromState,
		})
	}

	_, toOK := def.States[toState]
	if !toOK {
		diagnostics = append(diagnostics, WorkflowDiagnostic{
			Code:    "workflow.unknown_state",
			Message: fmt.Sprintf("unknown target state: %s", toState),
			Path:    toState,
		})
	}

	if fromOK && toOK && !containsTransition(from.Transitions, toState) {
		diagnostics = append(diagnostics, WorkflowDiagnostic{
			Code:    "workflow.invalid_transition",
			Message: fmt.Sprintf("transition %s -> %s is not allowed", fromState, toState),
			Path:    toState,
		})
	}

	if toOK {
		required := requiredFields(def.States[toState].Schema)
		for _, field := range required {
			if _, ok := metadata[field]; !ok {
				diagnostics = append(diagnostics, WorkflowDiagnostic{
					Code:    "workflow.missing_required_field",
					Message: fmt.Sprintf("required field missing for state %s: %s", toState, field),
					Path:    field,
				})
			}
		}
	}

	return WorkflowValidationResult{Valid: len(diagnostics) == 0, Diagnostics: diagnostics}
}

func requiredFields(schema map[string]any) []string {
	if schema == nil {
		return nil
	}

	raw, ok := schema["required"]
	if !ok {
		return nil
	}

	switch v := raw.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

func containsTransition(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
