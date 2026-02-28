package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustWorkflowRaw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	require.NoError(t, err)
	return data
}

func TestResolveWorkflowAssignment_SingleMatch(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"},
	}
	workflows := map[string]json.RawMessage{
		"project_story": mustWorkflowRaw(t, map[string]any{
			"description":   "Project flow",
			"initial_state": "proposed",
			"mode":          "enforce",
			"field":         "status",
			"states": map[string]any{
				"proposed": map[string]any{"schema": map[string]any{"type": "object"}, "transitions": []string{"planned"}},
			},
		}),
	}

	result := ResolveWorkflowAssignment("stories/s-1.md", groups, workflows)
	require.True(t, result.Resolved)
	assert.Equal(t, "project_story", result.WorkflowID)
	assert.Equal(t, "Stories", result.GroupName)
	assert.Equal(t, "status", result.Workflow.Field)
	assert.Empty(t, result.Diagnostics)
}

func TestResolveWorkflowAssignment_MultiMatchSameWorkflowID(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"},
		{Name: "All", Globs: []string{"**/*.md"}, WorkflowID: "project_story"},
	}
	workflows := map[string]json.RawMessage{
		"project_story": mustWorkflowRaw(t, map[string]any{
			"description":   "Project flow",
			"initial_state": "proposed",
			"mode":          "enforce",
			"field":         "status",
			"states": map[string]any{
				"proposed": map[string]any{"schema": map[string]any{"type": "object"}, "transitions": []string{"planned"}},
			},
		}),
	}

	result := ResolveWorkflowAssignment("stories/s-1.md", groups, workflows)
	require.True(t, result.Resolved)
	assert.Equal(t, "project_story", result.WorkflowID)
	assert.Empty(t, result.Diagnostics)
}

func TestResolveWorkflowAssignment_ConflictAcrossGroups(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"},
		{Name: "All", Globs: []string{"**/*.md"}, WorkflowID: "task_flow"},
	}

	result := ResolveWorkflowAssignment("stories/s-1.md", groups, map[string]json.RawMessage{})
	require.False(t, result.Resolved)
	require.NotEmpty(t, result.Diagnostics)
	assert.Equal(t, "workflow.assignment_conflict", result.Diagnostics[0].Code)
}

func TestResolveWorkflowAssignment_NoAssignment(t *testing.T) {
	groups := []NotebookGroup{{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"}}

	result := ResolveWorkflowAssignment("tasks/t-1.md", groups, map[string]json.RawMessage{})
	require.False(t, result.Resolved)
	require.NotEmpty(t, result.Diagnostics)
	assert.Equal(t, "workflow.assignment_not_found", result.Diagnostics[0].Code)
}

func TestResolveWorkflowAssignment_UnknownWorkflowID(t *testing.T) {
	groups := []NotebookGroup{{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"}}

	result := ResolveWorkflowAssignment("stories/s-1.md", groups, map[string]json.RawMessage{})
	require.False(t, result.Resolved)
	require.NotEmpty(t, result.Diagnostics)
	assert.Equal(t, "workflow.assignment_unknown_workflow", result.Diagnostics[0].Code)
}
