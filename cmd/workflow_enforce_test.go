package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/zenobi-us/jot/internal/services"
)

func testNotebookWithWorkflow(t *testing.T, root string) *services.Notebook {
	t.Helper()

	wfDef, _ := json.Marshal(map[string]any{
		"description":   "Test flow",
		"initial_state": "draft",
		"mode":          "enforce",
		"field":         "status",
		"states": map[string]any{
			"draft": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title"}},
				"transitions": []string{"review"},
			},
			"review": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title"}},
				"transitions": []string{"published"},
			},
			"published": map[string]any{
				"schema":      map[string]any{"type": "object", "required": []string{"title"}},
				"transitions": []string{},
			},
		},
	})

	return &services.Notebook{
		Config: services.NotebookConfig{
			StoredNotebookConfig: services.StoredNotebookConfig{
				Root: root,
				Name: "test",
				Groups: []services.NotebookGroup{
					{Name: "Docs", Globs: []string{"docs/*.md"}, WorkflowID: "doc_flow"},
				},
				Workflows: map[string]json.RawMessage{
					"doc_flow": wfDef,
				},
			},
		},
	}
}

func TestEnforceWorkflowForCreate_AllowsInitialState(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	notePath := filepath.Join(root, "docs", "new.md")
	metadata := map[string]any{"title": "New doc", "status": "draft"}

	err := enforceWorkflowForCreate(nb, notePath, metadata)
	if err != nil {
		t.Fatalf("expected create to be allowed: %v", err)
	}
}

func TestEnforceWorkflowForCreate_BlocksInvalidInitialState(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	notePath := filepath.Join(root, "docs", "bad.md")
	metadata := map[string]any{"title": "Bad doc", "status": "published"}

	err := enforceWorkflowForCreate(nb, notePath, metadata)
	if err == nil {
		t.Fatalf("expected create to be blocked")
	}
	if ExitCode(err) != ExitCodeWorkflowBlocked {
		t.Fatalf("expected exit code %d, got %d", ExitCodeWorkflowBlocked, ExitCode(err))
	}
}

func TestEnforceWorkflowForCreate_AllowsNoWorkflowMatch(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	notePath := filepath.Join(root, "notes", "random.md")
	metadata := map[string]any{"title": "Random"}

	err := enforceWorkflowForCreate(nb, notePath, metadata)
	if err != nil {
		t.Fatalf("expected no-workflow note to be allowed: %v", err)
	}
}

func TestEnforceWorkflowForCreate_AllowsNoGroups(t *testing.T) {
	root := t.TempDir()
	nb := &services.Notebook{
		Config: services.NotebookConfig{
			StoredNotebookConfig: services.StoredNotebookConfig{
				Root: root,
				Name: "empty",
			},
		},
	}

	err := enforceWorkflowForCreate(nb, filepath.Join(root, "any.md"), map[string]any{})
	if err != nil {
		t.Fatalf("expected no-groups notebook to allow create: %v", err)
	}
}

func TestEnforceWorkflowForEdit_AllowsValidTransition(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	notePath := filepath.Join(root, "docs", "doc1.md")
	if err := os.MkdirAll(filepath.Dir(notePath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	existingContent := []byte("---\ntitle: Doc1\nstatus: draft\n---\n\n# Doc1\n")
	newMetadata := map[string]any{"title": "Doc1", "status": "review"}

	err := enforceWorkflowForEdit(nb, notePath, existingContent, newMetadata)
	if err != nil {
		t.Fatalf("expected edit to be allowed: %v", err)
	}
}

func TestEnforceWorkflowForEdit_BlocksInvalidTransition(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	notePath := filepath.Join(root, "docs", "doc1.md")
	existingContent := []byte("---\ntitle: Doc1\nstatus: draft\n---\n\n# Doc1\n")
	newMetadata := map[string]any{"title": "Doc1", "status": "published"} // draft->published invalid

	err := enforceWorkflowForEdit(nb, notePath, existingContent, newMetadata)
	if err == nil {
		t.Fatalf("expected edit to be blocked")
	}
	if ExitCode(err) != ExitCodeWorkflowBlocked {
		t.Fatalf("expected exit code %d, got %d", ExitCodeWorkflowBlocked, ExitCode(err))
	}
}

func TestEnforceWorkflowForEdit_AllowsNoStateChange(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	notePath := filepath.Join(root, "docs", "doc1.md")
	existingContent := []byte("---\ntitle: Doc1\nstatus: draft\n---\n\n# Doc1\n")
	newMetadata := map[string]any{"title": "Doc1 Renamed", "status": "draft"} // same state

	err := enforceWorkflowForEdit(nb, notePath, existingContent, newMetadata)
	if err != nil {
		t.Fatalf("expected no-state-change edit to be allowed: %v", err)
	}
}

func TestExtractFrontmatterMetadata_ParsesYAML(t *testing.T) {
	content := []byte("---\ntitle: Test\nstatus: draft\n---\n\n# Body\n")
	meta := extractFrontmatterMetadata(content)

	if meta["title"] != "Test" {
		t.Fatalf("expected title=Test, got %v", meta["title"])
	}
	if meta["status"] != "draft" {
		t.Fatalf("expected status=draft, got %v", meta["status"])
	}
}

func TestExtractFrontmatterMetadata_NoFrontmatter(t *testing.T) {
	content := []byte("# Just a heading\n\nSome body.\n")
	meta := extractFrontmatterMetadata(content)

	if len(meta) != 0 {
		t.Fatalf("expected empty metadata, got %v", meta)
	}
}

func TestUpdateNoteFileWithWorkflow_BlocksInvalidCreate(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	content := []byte("---\ntitle: Bad\nstatus: published\n---\n\n# Bad\n")
	_, err := updateNoteFileWithWorkflow(root, "docs/bad", content, true, nb)
	if err == nil {
		t.Fatalf("expected workflow-blocked create to fail")
	}
	if ExitCode(err) != ExitCodeWorkflowBlocked {
		t.Fatalf("expected exit code %d, got %d", ExitCodeWorkflowBlocked, ExitCode(err))
	}
}

func TestUpdateNoteFileWithWorkflow_AllowsValidCreate(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	content := []byte("---\ntitle: Good\nstatus: draft\n---\n\n# Good\n")
	result, err := updateNoteFileWithWorkflow(root, "docs/good", content, true, nb)
	if err != nil {
		t.Fatalf("expected valid create to succeed: %v", err)
	}
	if result.Action != "created" {
		t.Fatalf("expected action=created, got %s", result.Action)
	}
}

func TestUpdateNoteFileWithWorkflow_BlocksInvalidEdit(t *testing.T) {
	root := t.TempDir()
	nb := testNotebookWithWorkflow(t, root)

	// Create existing note
	notePath := filepath.Join(root, "docs", "existing.md")
	if err := os.MkdirAll(filepath.Dir(notePath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(notePath, []byte("---\ntitle: Existing\nstatus: draft\n---\n\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	// Try to update with invalid transition
	newContent := []byte("---\ntitle: Existing\nstatus: published\n---\n\n")
	_, err := updateNoteFileWithWorkflow(root, "docs/existing", newContent, false, nb)
	if err == nil {
		t.Fatalf("expected workflow-blocked edit to fail")
	}
	if ExitCode(err) != ExitCodeWorkflowBlocked {
		t.Fatalf("expected exit code %d, got %d", ExitCodeWorkflowBlocked, ExitCode(err))
	}
}

func TestUpdateNoteFileWithWorkflow_AllowsWithoutNotebook(t *testing.T) {
	root := t.TempDir()

	content := []byte("---\ntitle: No Workflow\n---\n\n")
	result, err := updateNoteFileWithWorkflow(root, "anything", content, true, nil)
	if err != nil {
		t.Fatalf("expected nil-notebook to allow: %v", err)
	}
	if result.Action != "created" {
		t.Fatalf("expected action=created, got %s", result.Action)
	}
}
