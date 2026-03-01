package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveWorkflowAssignment_SingleMatch(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"},
	}
	workflows := map[string]WorkflowDefinition{
		"project_story": {
			Description:  "Project flow",
			InitialState: "proposed",
			Mode:         "enforce",
			Field:        "status",
			States: map[string]WorkflowStateDefinition{
				"proposed": {Schema: map[string]any{"type": "object"}, Transitions: []string{"planned"}},
			},
		},
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
	workflows := map[string]WorkflowDefinition{
		"project_story": {
			Description:  "Project flow",
			InitialState: "proposed",
			Mode:         "enforce",
			Field:        "status",
			States: map[string]WorkflowStateDefinition{
				"proposed": {Schema: map[string]any{"type": "object"}, Transitions: []string{"planned"}},
			},
		},
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

	result := ResolveWorkflowAssignment("stories/s-1.md", groups, map[string]WorkflowDefinition{})
	require.False(t, result.Resolved)
	require.NotEmpty(t, result.Diagnostics)
	assert.Equal(t, "workflow.assignment_conflict", result.Diagnostics[0].Code)
}

func TestResolveWorkflowAssignment_NoAssignment(t *testing.T) {
	groups := []NotebookGroup{{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"}}

	result := ResolveWorkflowAssignment("tasks/t-1.md", groups, map[string]WorkflowDefinition{})
	require.False(t, result.Resolved)
	require.NotEmpty(t, result.Diagnostics)
	assert.Equal(t, "workflow.assignment_not_found", result.Diagnostics[0].Code)
}

func TestResolveWorkflowAssignment_UnknownWorkflowID(t *testing.T) {
	groups := []NotebookGroup{{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"}}

	result := ResolveWorkflowAssignment("stories/s-1.md", groups, map[string]WorkflowDefinition{})
	require.False(t, result.Resolved)
	require.NotEmpty(t, result.Diagnostics)
	assert.Equal(t, "workflow.assignment_unknown_workflow", result.Diagnostics[0].Code)
}
