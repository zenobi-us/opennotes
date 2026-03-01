package services

import (
	"fmt"
	"sort"
	"strings"
)

// WorkflowAssignmentResult contains deterministic assignment resolution outcome.
type WorkflowAssignmentResult struct {
	Resolved    bool                 `json:"resolved"`
	WorkflowID  string               `json:"workflow_id,omitempty"`
	GroupName   string               `json:"group_name,omitempty"`
	Workflow    WorkflowDefinition   `json:"workflow,omitempty"`
	Diagnostics []WorkflowDiagnostic `json:"diagnostics,omitempty"`
}

// ResolveWorkflowAssignment finds the workflow assigned to a note path based on group matches.
func ResolveWorkflowAssignment(notePath string, groups []NotebookGroup, workflows map[string]WorkflowDefinition) WorkflowAssignmentResult {
	type match struct {
		groupName  string
		workflowID string
	}

	matches := make([]match, 0)
	workflowIDs := make(map[string]struct{})
	for _, g := range groups {
		if g.WorkflowID == "" {
			continue
		}

		matched := false
		for _, glob := range g.Globs {
			if globMatch(glob, notePath) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}

		matches = append(matches, match{groupName: g.Name, workflowID: g.WorkflowID})
		workflowIDs[g.WorkflowID] = struct{}{}
	}

	if len(matches) == 0 {
		return WorkflowAssignmentResult{
			Resolved: false,
			Diagnostics: []WorkflowDiagnostic{{
				Code:    "workflow.assignment_not_found",
				Message: fmt.Sprintf("no workflow assignment found for note path: %s", notePath),
				Path:    notePath,
			}},
		}
	}

	if len(workflowIDs) > 1 {
		groupLabels := make([]string, 0, len(matches))
		for _, m := range matches {
			groupLabels = append(groupLabels, fmt.Sprintf("%s:%s", m.groupName, m.workflowID))
		}
		sort.Strings(groupLabels)

		return WorkflowAssignmentResult{
			Resolved: false,
			Diagnostics: []WorkflowDiagnostic{{
				Code:    "workflow.assignment_conflict",
				Message: fmt.Sprintf("conflicting workflow assignments for note path %s: %s", notePath, strings.Join(groupLabels, ", ")),
				Path:    notePath,
			}},
		}
	}

	selected := matches[0]
	def, ok := workflows[selected.workflowID]
	if !ok {
		return WorkflowAssignmentResult{
			Resolved: false,
			Diagnostics: []WorkflowDiagnostic{{
				Code:    "workflow.assignment_unknown_workflow",
				Message: fmt.Sprintf("group %s references unknown workflow_id: %s", selected.groupName, selected.workflowID),
				Path:    selected.workflowID,
			}},
		}
	}

	if strings.TrimSpace(def.Field) == "" {
		return WorkflowAssignmentResult{
			Resolved: false,
			Diagnostics: []WorkflowDiagnostic{{
				Code:    "workflow.missing_field_selector",
				Message: fmt.Sprintf("workflow %s is missing required field selector", selected.workflowID),
				Path:    selected.workflowID + ".field",
			}},
		}
	}

	return WorkflowAssignmentResult{
		Resolved:    true,
		WorkflowID:  selected.workflowID,
		GroupName:   selected.groupName,
		Workflow:    def,
		Diagnostics: nil,
	}
}
