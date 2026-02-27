package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Text Search E2E Tests
// ============================================================================

func TestE2E_TextSearch_BasicSearch(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Search for "meeting" - should find meeting notes
	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "meeting")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "meeting-notes.md", "should find meeting-notes.md")
}

func TestE2E_TextSearch_NoResults(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Search for non-existent term
	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "xyz123nonexistent")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "No notes found", "should indicate no results")
}

func TestE2E_TextSearch_CaseInsensitive(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Search with different case
	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "MEETING")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "meeting-notes.md", "case-insensitive search should work")
}

func TestE2E_TextSearch_ListAllNotes(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Search without term lists all notes
	stdout, stderr, code := env.runInDir(nbDir, "notes", "search")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "meeting-notes.md", "should list meeting-notes.md")
	assert.Contains(t, stdout, "project-plan.md", "should list project-plan.md")
	assert.Contains(t, stdout, "active-task.md", "should list active-task.md")
}

// ============================================================================
// Removed Fuzzy Flag E2E Tests
// ============================================================================

func TestE2E_Search_RejectsRemovedFuzzyFlag(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	_, stderr, code := env.runInDir(nbDir, "notes", "search", "--fuzzy", "mtng")

	assert.NotEqual(t, 0, code, "removed --fuzzy flag should fail")
	assert.Contains(t, stderr, "unknown flag: --fuzzy", "should clearly report removed flag")
}

// ============================================================================
// DSL Filter E2E Tests
// ============================================================================

func TestE2E_DSLFilter_SingleAnd(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "tag:workflow")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "active-task.md", "should find note with workflow tag")
}

func TestE2E_DSLFilter_MultipleAnd(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "tag:workflow status:active")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "active-task.md", "should find active workflow note")
}

func TestE2E_DSLFilter_OrConditions(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "tag:workflow OR tag:meeting")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "active-task.md", "should include workflow note")
	assert.Contains(t, stdout, "meeting-notes.md", "should include meeting note")
}

func TestE2E_DSLFilter_CombinedConditions(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "tag:epic status:active")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "epic1.md", "should find active epic")
	assert.NotContains(t, stdout, "epic2.md", "should exclude archived epic")
}

func TestE2E_DSLFilter_TitleMatch(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "title:Epic")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	assert.Contains(t, stdout, "epic1.md", "should find epic notes by title")
}

// ============================================================================
// Link Query E2E Tests
//
// NOTE: These tests are currently skipped because DuckDB's markdown extension
// does not properly parse YAML arrays in frontmatter. The `links` field comes
// back as null/empty instead of as an array of strings.
//
// The link query implementation is correct and tested at the unit level,
// but requires proper array support from the markdown extension.
// ============================================================================

func TestE2E_LinkQuery_LinksTo(t *testing.T) {
	t.Skip("SKIP: DuckDB markdown extension does not parse YAML arrays - links field returns null")

	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Find notes that link to tasks/task1.md
	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "links-to:tasks/task1.md")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	// epic1.md has links to tasks/task1.md
	assert.Contains(t, stdout, "epic1.md", "should find epic that links to task")
}

func TestE2E_LinkQuery_LinksToGlob(t *testing.T) {
	t.Skip("SKIP: DuckDB markdown extension does not parse YAML arrays - links field returns null")

	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Find notes that link to any task
	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "links-to:tasks/*.md")

	assert.Equal(t, 0, code, "exit code should be 0, stderr: %s", stderr)
	// Should find epics that link to tasks
	assert.NotContains(t, stdout, "No notes found", "should find notes linking to tasks")
}

// ============================================================================
// Semantic Search Command E2E Tests
// ============================================================================

func TestE2E_SemanticSearch_InvalidMode(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	_, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting", "--mode", "invalid")

	assert.NotEqual(t, 0, code, "invalid mode should fail")
	assert.Contains(t, stderr, "invalid mode", "should explain mode validation failure")
}

func TestE2E_SemanticSearch_KeywordMode_WithDSLFilters(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(
		nbDir,
		"notes", "search", "semantic", "task",
		"--mode", "keyword",
		"--and", "data.status=active",
	)

	assert.Equal(t, 0, code, "keyword semantic command should succeed, stderr: %s", stderr)
	assert.Contains(t, stdout, "active-task.md", "should include active task note")
	assert.Contains(t, stdout, "tasks/task1.md", "should include active task1 note")
	assert.NotContains(t, stdout, "tasks/task2.md", "should exclude done task via filter")
}

func TestE2E_SemanticSearch_HybridFallbackWarning(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting")

	assert.Equal(t, 0, code, "hybrid fallback should still succeed, stderr: %s", stderr)
	assert.Contains(t, stdout, "Warning: semantic backend unavailable", "should warn about fallback")
	assert.Contains(t, stdout, "meeting-notes.md", "should still return keyword results")
}

func TestE2E_SemanticSearch_SemanticModeUnavailable(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting", "--mode", "semantic")

	assert.Equal(t, 0, code, "semantic mode unavailability should be a non-fatal warning, stderr: %s", stderr)
	assert.Contains(t, stdout, "Semantic backend unavailable", "should explain why semantic mode cannot run")
}

// ============================================================================
// Error Handling E2E Tests
// ============================================================================

func TestE2E_ErrorHandling_InvalidField(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Invalid field name via condition flags on semantic command
	_, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting",
		"--and", "invalid.field=value")

	assert.NotEqual(t, 0, code, "should fail with invalid field")
	assert.Contains(t, stderr, "invalid field", "should report invalid field error")
}

func TestE2E_ErrorHandling_InvalidFormat(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Missing equals sign
	_, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting",
		"--and", "data.tag-workflow")

	assert.NotEqual(t, 0, code, "should fail with invalid format")
	assert.Contains(t, stderr, "expected field=value", "should report format error")
}

func TestE2E_ErrorHandling_ValueTooLong(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Value exceeds max length (1000 chars)
	longValue := strings.Repeat("a", 2000)
	_, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting",
		"--and", "data.tag="+longValue)

	assert.NotEqual(t, 0, code, "should fail with long value")
	assert.Contains(t, stderr, "too long", "should report value too long error")
}

func TestE2E_ErrorHandling_EmptyValue(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	// Empty value
	_, stderr, code := env.runInDir(nbDir, "notes", "search", "semantic", "meeting",
		"--and", "data.tag=")

	assert.NotEqual(t, 0, code, "should fail with empty value")
	assert.Contains(t, stderr, "cannot be empty", "should report empty value error")
}

func TestE2E_ErrorHandling_QueryTreatedAsSearchTerm(t *testing.T) {
	env := newTestEnv(t)
	nbDir := setupSearchNotebook(t, env)

	stdout, stderr, code := env.runInDir(nbDir, "notes", "search", "query")

	assert.Equal(t, 0, code, "query token should be treated as regular search term, stderr: %s", stderr)
	assert.NotContains(t, stderr, "unknown command", "query token is not a subcommand")
	assert.Contains(t, stdout, "No notes found", "should run normal fieldless search flow")
}

// ============================================================================
// Security E2E Tests
// ============================================================================

// NOTE: TestE2E_Security_SQLInjectionPrevention removed as part of Phase 5.
// SQL injection is no longer possible because we use Bleve (not SQL) for search.
// Query parameters are validated at the parser level, not SQL level.

// ============================================================================
// CLI Help Text E2E Tests
// ============================================================================

func TestE2E_HelpText_SearchCommand(t *testing.T) {
	env := newTestEnv(t)

	stdout, _, code := env.run("notes", "search", "--help")

	assert.Equal(t, 0, code, "help should succeed")
	// Verify help text reflects fuzzy and legacy query command removal
	assert.NotContains(t, stdout, "--fuzzy", "should not mention removed fuzzy flag")
	assert.NotContains(t, stdout, "search query", "should not mention removed query subcommand")
	assert.Contains(t, stdout, "Search notes", "should have search description")
}

func TestE2E_HelpText_QueryTokenWithHelp(t *testing.T) {
	env := newTestEnv(t)

	stdout, stderr, code := env.run("notes", "search", "query", "--help")

	assert.Equal(t, 0, code, "--help should show search help")
	assert.NotContains(t, stderr, "unknown command", "query token is not a subcommand")
	assert.Contains(t, stdout, "Search notes", "should show search command help")
}

// ============================================================================
// Helper Functions
// ============================================================================

// setupSearchNotebook creates a test notebook with notes for search testing.
func setupSearchNotebook(t *testing.T, env *testEnv) string {
	t.Helper()

	nbDir := filepath.Join(env.tmpDir, "search-notebook")
	require.NoError(t, os.MkdirAll(nbDir, 0755))

	// Create notebook config
	config := `{
		"name": "Search Test Notebook",
		"version": "1.0.0"
	}`
	require.NoError(t, os.WriteFile(filepath.Join(nbDir, ".jot.json"), []byte(config), 0644))

	// Create notes directory structure
	require.NoError(t, os.MkdirAll(filepath.Join(nbDir, "epics"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(nbDir, "tasks"), 0755))

	// Create test notes with different metadata
	notes := map[string]string{
		"meeting-notes.md": `---
title: Meeting Notes
tag: meeting
status: active
---

# Meeting Notes

Team meeting discussion about the project.
`,
		"project-plan.md": `---
title: Project Plan
tag: planning
status: active
priority: high
---

# Project Plan

This is the project planning document.
`,
		"active-task.md": `---
title: Active Task
tag: workflow
status: active
priority: medium
---

# Active Task

A task that is currently active.
`,
		"epics/epic1.md": `---
title: Epic 1
tag: epic
status: active
priority: high
links:
  - tasks/task1.md
  - tasks/task2.md
---

# Epic 1

An active epic that links to tasks.
`,
		"epics/epic2.md": `---
title: Epic 2
tag: epic
status: archived
---

# Epic 2

An archived epic.
`,
		"tasks/task1.md": `---
title: Task 1
tag: task
status: active
---

# Task 1

A simple task.
`,
		"tasks/task2.md": `---
title: Task 2
tag: task
status: done
---

# Task 2

A completed task.
`,
	}

	for path, content := range notes {
		fullPath := filepath.Join(nbDir, path)
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	return nbDir
}
