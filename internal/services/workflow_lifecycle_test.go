package services

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func storyWorkflowRaw(t *testing.T) json.RawMessage {
	t.Helper()
	return mustWorkflowRaw(t, map[string]any{
		"description":   "Project flow",
		"initial_state": "proposed",
		"mode":          "enforce",
		"field":         "status",
		"states": map[string]any{
			"proposed": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title"}},
				"transitions": []string{"planned", "cancelled"},
			},
			"planned": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title", "epic_id"}},
				"transitions": []string{"in-progress", "cancelled"},
			},
			"in-progress": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title", "epic_id"}},
				"transitions": []string{"completed", "cancelled"},
			},
			"completed": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title", "epic_id", "completed_at"}},
				"transitions": []string{},
			},
			"cancelled": map[string]any{
				"schema":      map[string]any{"type": "object"},
				"transitions": []string{},
			},
		},
	})
}

func defaultTestGroups() []NotebookGroup {
	return []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"},
	}
}

func defaultTestWorkflows(t *testing.T) map[string]json.RawMessage {
	return map[string]json.RawMessage{
		"project_story": storyWorkflowRaw(t),
	}
}

// --- Create path tests ---

func TestEnforceWorkflowOnMutation_CreateWithInitialState_Allowed(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/new-story.md",
		groups, workflows,
		nil, // no existing metadata (create)
		map[string]any{"title": "New Story", "status": "proposed"},
		true, // isCreate
	)

	assert.True(t, result.Enforced)
	assert.True(t, result.Allowed)
	assert.Equal(t, "project_story", result.WorkflowID)
	assert.Equal(t, "proposed", result.FromState)
	assert.Equal(t, "proposed", result.ToState)
	assert.Empty(t, result.Diagnostics)
}

func TestEnforceWorkflowOnMutation_CreateWithNoStatusField_DefaultsToInitial(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/another.md",
		groups, workflows,
		nil,
		map[string]any{"title": "Another Story"}, // no "status" key
		true,
	)

	assert.True(t, result.Enforced)
	assert.True(t, result.Allowed)
	assert.Equal(t, "proposed", result.ToState)
	assert.Empty(t, result.Diagnostics)
}

func TestEnforceWorkflowOnMutation_CreateWithInvalidInitialState_Blocked(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/bad.md",
		groups, workflows,
		nil,
		map[string]any{"title": "Bad Story", "status": "completed"}, // skip to completed
		true,
	)

	assert.True(t, result.Enforced)
	assert.False(t, result.Allowed)
	assert.Equal(t, "proposed", result.FromState)
	assert.Equal(t, "completed", result.ToState)
	require.NotEmpty(t, result.Diagnostics)
	// Should contain lifecycle_blocked and invalid_transition
	codes := diagCodes(result.Diagnostics)
	assert.Contains(t, codes, "workflow.lifecycle_blocked")
	assert.Contains(t, codes, "workflow.invalid_transition")
}

// --- Edit path tests ---

func TestEnforceWorkflowOnMutation_EditValidTransition_Allowed(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/s1.md",
		groups, workflows,
		map[string]any{"title": "S1", "status": "proposed"},                 // existing
		map[string]any{"title": "S1", "status": "planned", "epic_id": "e1"}, // new
		false,
	)

	assert.True(t, result.Enforced)
	assert.True(t, result.Allowed)
	assert.Equal(t, "proposed", result.FromState)
	assert.Equal(t, "planned", result.ToState)
	assert.Empty(t, result.Diagnostics)
}

func TestEnforceWorkflowOnMutation_EditInvalidTransition_Blocked(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/s1.md",
		groups, workflows,
		map[string]any{"title": "S1", "status": "proposed"},
		map[string]any{"title": "S1", "status": "completed", "epic_id": "e1", "completed_at": "now"},
		false,
	)

	assert.True(t, result.Enforced)
	assert.False(t, result.Allowed)
	codes := diagCodes(result.Diagnostics)
	assert.Contains(t, codes, "workflow.lifecycle_blocked")
}

func TestEnforceWorkflowOnMutation_EditNoStateChange_Allowed(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/s1.md",
		groups, workflows,
		map[string]any{"title": "S1", "status": "proposed"},
		map[string]any{"title": "S1 Renamed", "status": "proposed"}, // same state
		false,
	)

	assert.True(t, result.Enforced)
	assert.True(t, result.Allowed)
	assert.Equal(t, "proposed", result.FromState)
	assert.Equal(t, "proposed", result.ToState)
	assert.Empty(t, result.Diagnostics)
}

// --- No workflow match tests ---

func TestEnforceWorkflowOnMutation_NoWorkflowMatch_AllowedNotEnforced(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"docs/random.md", // doesn't match stories/*.md
		groups, workflows,
		nil,
		map[string]any{"title": "Random"},
		true,
	)

	assert.False(t, result.Enforced)
	assert.True(t, result.Allowed)
	assert.Empty(t, result.Diagnostics)
}

// --- Assignment error tests ---

func TestEnforceWorkflowOnMutation_ConflictingWorkflows_Blocked(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "project_story"},
		{Name: "All", Globs: []string{"**/*.md"}, WorkflowID: "other_flow"},
	}
	workflows := defaultTestWorkflows(t)

	result := EnforceWorkflowOnMutation(
		"stories/conflict.md",
		groups, workflows,
		nil,
		map[string]any{"title": "Conflict"},
		true,
	)

	assert.True(t, result.Enforced)
	assert.False(t, result.Allowed)
	codes := diagCodes(result.Diagnostics)
	assert.Contains(t, codes, "workflow.assignment_conflict")
}

func TestEnforceWorkflowOnMutation_UnknownWorkflowRef_Blocked(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Stories", Globs: []string{"stories/*.md"}, WorkflowID: "nonexistent"},
	}

	result := EnforceWorkflowOnMutation(
		"stories/s1.md",
		groups,
		map[string]json.RawMessage{}, // empty workflows
		nil,
		map[string]any{"title": "S1"},
		true,
	)

	assert.True(t, result.Enforced)
	assert.False(t, result.Allowed)
	codes := diagCodes(result.Diagnostics)
	assert.Contains(t, codes, "workflow.assignment_unknown_workflow")
}

// --- Missing metadata tests ---

func TestEnforceWorkflowOnMutation_CreateMissingRequiredMetadata_Blocked(t *testing.T) {
	groups := defaultTestGroups()
	workflows := defaultTestWorkflows(t)

	// Transition proposed -> planned requires "title" and "epic_id"
	result := EnforceWorkflowOnMutation(
		"stories/missing-meta.md",
		groups, workflows,
		nil,
		map[string]any{"status": "planned"}, // missing title and epic_id
		true,
	)

	assert.True(t, result.Enforced)
	assert.False(t, result.Allowed)
	codes := diagCodes(result.Diagnostics)
	assert.Contains(t, codes, "workflow.missing_required_field")
}

// --- Helper ---

func diagCodes(diagnostics []WorkflowDiagnostic) []string {
	codes := make([]string, 0, len(diagnostics))
	for _, d := range diagnostics {
		codes = append(codes, d.Code)
	}
	return codes
}
