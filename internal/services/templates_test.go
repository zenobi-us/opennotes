package services

import (
	"strings"
	"testing"
)

func TestTuiRender_NoteList_WithNotes(t *testing.T) {
	// Create notes with the anonymous struct format
	note1 := Note{}
	note1.File.Relative = "notes/test1.md"
	note1.File.Filepath = "/path/to/notes/test1.md"

	note2 := Note{}
	note2.File.Relative = "notes/test2.md"
	note2.File.Filepath = "/path/to/notes/test2.md"

	ctx := map[string]any{
		"Notes": []Note{note1, note2},
	}

	result, err := TuiRender(Templates.NoteList, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	// Check for note count
	if !strings.Contains(result, "2") {
		t.Errorf("TuiRender() result = %q, want to contain note count '2'", result)
	}

	// Check for note files
	if !strings.Contains(result, "test1.md") {
		t.Errorf("TuiRender() result = %q, want to contain 'test1.md'", result)
	}

	if !strings.Contains(result, "test2.md") {
		t.Errorf("TuiRender() result = %q, want to contain 'test2.md'", result)
	}
}

func TestTuiRender_NoteList_Empty(t *testing.T) {
	ctx := map[string]any{
		"Notes": []Note{},
	}

	result, err := TuiRender(Templates.NoteList, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	if !strings.Contains(result, "No notes found") {
		t.Errorf("TuiRender() result = %q, want to contain 'No notes found'", result)
	}
}

func TestTuiRender_NotebookInfo_AllFields(t *testing.T) {
	ctx := map[string]any{
		"Config": NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Name:     "Test Notebook",
				Root:     "/path/to/notebook",
				Contexts: []string{"/context1", "/context2"},
				Groups: []NotebookGroup{
					{Name: "docs", Globs: []string{"*.md", "docs/**/*.md"}},
				},
			},
			Path: "/path/to/config.json",
		},
	}

	result, err := TuiRender(Templates.NotebookInfo, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	// Check for notebook name
	if !strings.Contains(result, "Test Notebook") {
		t.Errorf("TuiRender() result = %q, want to contain 'Test Notebook'", result)
	}

	// Check for contexts
	if !strings.Contains(result, "context1") {
		t.Errorf("TuiRender() result = %q, want to contain 'context1'", result)
	}

	// Check for groups
	if !strings.Contains(result, "docs") {
		t.Errorf("TuiRender() result = %q, want to contain 'docs'", result)
	}
}

func TestTuiRender_NotebookInfo_MinimalFields(t *testing.T) {
	ctx := map[string]any{
		"Config": NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Name: "Minimal Notebook",
				Root: "/path/to/notebook",
			},
			Path: "/path/to/config.json",
		},
	}

	result, err := TuiRender(Templates.NotebookInfo, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	if !strings.Contains(result, "Minimal Notebook") {
		t.Errorf("TuiRender() result = %q, want to contain 'Minimal Notebook'", result)
	}

	// Should not contain Contexts or Groups headers if they're empty
	// (template conditionally renders them)
}

func TestTuiRender_NotebookList_WithNotebooks(t *testing.T) {
	nb1 := &Notebook{
		Config: NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Name:     "Notebook 1",
				Root:     "/path/to/nb1",
				Contexts: []string{"/ctx1"},
			},
			Path: "/path/to/nb1/config.json",
		},
	}

	nb2 := &Notebook{
		Config: NotebookConfig{
			StoredNotebookConfig: StoredNotebookConfig{
				Name: "Notebook 2",
				Root: "/path/to/nb2",
			},
			Path: "/path/to/nb2/config.json",
		},
	}

	ctx := map[string]any{
		"Notebooks": []*Notebook{nb1, nb2},
	}

	result, err := TuiRender(Templates.NotebookList, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	// Check for notebook count
	if !strings.Contains(result, "2") {
		t.Errorf("TuiRender() result = %q, want to contain notebook count '2'", result)
	}

	// Check for notebook names
	if !strings.Contains(result, "Notebook 1") {
		t.Errorf("TuiRender() result = %q, want to contain 'Notebook 1'", result)
	}

	if !strings.Contains(result, "Notebook 2") {
		t.Errorf("TuiRender() result = %q, want to contain 'Notebook 2'", result)
	}
}

func TestTuiRender_NotebookList_Empty(t *testing.T) {
	ctx := map[string]any{
		"Notebooks": []*Notebook{},
	}

	result, err := TuiRender(Templates.NotebookList, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	if !strings.Contains(result, "No notebooks found") {
		t.Errorf("TuiRender() result = %q, want to contain 'No notebooks found'", result)
	}
}

func TestTuiRender_NoteDetail(t *testing.T) {
	// Create note with file info
	note := Note{
		Content:  "This is the note content.",
		Metadata: map[string]any{"author": "Test User", "tags": "test, example"},
	}
	note.File.Relative = "notes/my-note.md"
	note.File.Filepath = "/path/to/notes/my-note.md"

	ctx := map[string]any{
		"Title":    "My Note",
		"File":     note.File,
		"Metadata": note.Metadata,
		"Content":  note.Content,
	}

	result, err := TuiRender(Templates.NoteDetail, ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	if !strings.Contains(result, "My Note") {
		t.Errorf("TuiRender() result = %q, want to contain 'My Note'", result)
	}

	if !strings.Contains(result, "my-note.md") {
		t.Errorf("TuiRender() result = %q, want to contain 'my-note.md'", result)
	}

	if !strings.Contains(result, "note content") {
		t.Errorf("TuiRender() result = %q, want to contain 'note content'", result)
	}
}

func TestTemplates_NotNil(t *testing.T) {
	// Ensure all templates are non-empty
	if Templates.NoteList == "" {
		t.Error("Templates.NoteList is empty")
	}

	if Templates.NoteDetail == "" {
		t.Error("Templates.NoteDetail is empty")
	}

	if Templates.NotebookInfo == "" {
		t.Error("Templates.NotebookInfo is empty")
	}

	if Templates.NotebookList == "" {
		t.Error("Templates.NotebookList is empty")
	}
}

func TestTuiRender_InvalidTemplate_Fallback(t *testing.T) {
	// Invalid template syntax
	tmpl := "# {{ .Title"
	ctx := map[string]string{"Title": "Test"}

	result, err := TuiRender(tmpl, ctx)
	if err != nil {
		t.Fatalf("TuiRender() should not fail on invalid template: %v", err)
	}

	// Should fallback to template as-is
	if result != tmpl {
		t.Errorf("TuiRender() result = %q, want fallback %q", result, tmpl)
	}
}

func TestTuiRender_NilContext(t *testing.T) {
	tmpl := "# Static Heading"

	result, err := TuiRender(tmpl, nil)
	if err != nil {
		t.Fatalf("TuiRender() failed with nil context: %v", err)
	}

	if !strings.Contains(result, "Static Heading") {
		t.Errorf("TuiRender() result = %q, want to contain 'Static Heading'", result)
	}
}
