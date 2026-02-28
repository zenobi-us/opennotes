package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadWorkflowFixture(t *testing.T, name string) WorkflowDefinition {
	t.Helper()

	path := filepath.Join("testdata", "workflows", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var def WorkflowDefinition
	require.NoError(t, json.Unmarshal(data, &def))
	return def
}

func assertDiagnosticCodes(t *testing.T, diagnostics []WorkflowDiagnostic, expected ...string) {
	t.Helper()
	actual := make([]string, 0, len(diagnostics))
	for _, d := range diagnostics {
		actual = append(actual, d.Code)
	}
	assert.ElementsMatch(t, expected, actual)
}

func TestWorkflowValidationHarness_ValidTransition(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := ValidateWorkflowTransition(def, "proposed", "planned", map[string]any{
		"title":   "My story",
		"epic_id": "b2f4e6a8",
	})

	assert.True(t, result.Valid)
	assert.Empty(t, result.Diagnostics)
}

func TestWorkflowValidationHarness_InvalidTransition(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := ValidateWorkflowTransition(def, "proposed", "completed", map[string]any{
		"title":        "My story",
		"epic_id":      "b2f4e6a8",
		"completed_at": "2026-02-28T17:00:00Z",
	})

	assert.False(t, result.Valid)
	assertDiagnosticCodes(t, result.Diagnostics, "workflow.invalid_transition")
}

func TestWorkflowValidationHarness_MissingRequiredMetadata(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := ValidateWorkflowTransition(def, "proposed", "planned", map[string]any{
		"title": "My story",
	})

	assert.False(t, result.Valid)
	assertDiagnosticCodes(t, result.Diagnostics, "workflow.missing_required_field")
}

func TestWorkflowEvaluationEntryPoint_ApplyModeValidTransition(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := EvaluateWorkflow(def, WorkflowExecutionRequest{
		Mode:      "apply",
		FromState: "proposed",
		ToState:   "planned",
		Metadata: map[string]any{
			"title":   "My story",
			"epic_id": "b2f4e6a8",
		},
	})

	assert.True(t, result.Valid)
	assert.True(t, result.Allowed)
	assert.True(t, result.Applied)
	assert.Equal(t, "apply", result.Mode)
	assert.Equal(t, "proposed", result.FromState)
	assert.Equal(t, "planned", result.ToState)
	assert.Equal(t, "planned", result.ResultState)
	assert.Empty(t, result.Diagnostics)
}

func TestWorkflowEvaluationEntryPoint_ApplyModeBlocksInvalidTransition(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := EvaluateWorkflow(def, WorkflowExecutionRequest{
		Mode:      "apply",
		FromState: "proposed",
		ToState:   "completed",
		Metadata: map[string]any{
			"title":        "My story",
			"epic_id":      "b2f4e6a8",
			"completed_at": "2026-02-28T17:00:00Z",
		},
	})

	assert.False(t, result.Valid)
	assert.False(t, result.Allowed)
	assert.False(t, result.Applied)
	assert.Equal(t, "proposed", result.ResultState)
	assertDiagnosticCodes(t, result.Diagnostics, "workflow.invalid_transition")
}

func TestWorkflowEvaluationEntryPoint_ApplyModeBlocksMissingMetadata(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := EvaluateWorkflow(def, WorkflowExecutionRequest{
		Mode:      "apply",
		FromState: "proposed",
		ToState:   "planned",
		Metadata: map[string]any{
			"title": "My story",
		},
	})

	assert.False(t, result.Valid)
	assert.False(t, result.Allowed)
	assert.False(t, result.Applied)
	assert.Equal(t, "proposed", result.ResultState)
	assertDiagnosticCodes(t, result.Diagnostics, "workflow.missing_required_field")
}

func TestWorkflowEvaluationEntryPoint_DryRunDoesNotApplyStateMutation(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")

	result := EvaluateWorkflow(def, WorkflowExecutionRequest{
		Mode:      "dry-run",
		FromState: "proposed",
		ToState:   "planned",
		Metadata: map[string]any{
			"title":   "My story",
			"epic_id": "b2f4e6a8",
		},
	})

	assert.True(t, result.Valid)
	assert.True(t, result.Allowed)
	assert.False(t, result.Applied)
	assert.Equal(t, "proposed", result.ResultState)
	assert.Empty(t, result.Diagnostics)
}

func TestWorkflowEvaluationEntryPoint_ApplyModeTreatsWarnAsEnforce(t *testing.T) {
	def := loadWorkflowFixture(t, "project_story.json")
	def.Mode = "warn"

	result := EvaluateWorkflow(def, WorkflowExecutionRequest{
		Mode:      "apply",
		FromState: "proposed",
		ToState:   "completed",
		Metadata: map[string]any{
			"title":        "My story",
			"epic_id":      "b2f4e6a8",
			"completed_at": "2026-02-28T17:00:00Z",
		},
	})

	assert.False(t, result.Valid)
	assert.False(t, result.Allowed)
	assert.False(t, result.Applied)
	assert.Equal(t, "proposed", result.ResultState)
	assertDiagnosticCodes(t, result.Diagnostics, "workflow.invalid_transition")
}
