package services

import (
	"strings"
	"testing"
)

func TestNote_DisplayName_WithTitle(t *testing.T) {
	note := Note{}
	note.File.Relative = "notes/my-note.md"
	note.File.Filepath = "/path/to/notes/my-note.md"
	note.Metadata = map[string]any{
		"title": "My Custom Title",
	}

	expected := "My Custom Title"
	if result := note.DisplayName(); result != expected {
		t.Errorf("DisplayName() = %q, want %q", result, expected)
	}
}

func TestNote_DisplayName_WithoutTitle(t *testing.T) {
	note := Note{}
	note.File.Relative = "notes/my-note-name.md"
	note.File.Filepath = "/path/to/notes/my-note-name.md"
	note.Metadata = map[string]any{}

	expected := "my-note-name"
	if result := note.DisplayName(); result != expected {
		t.Errorf("DisplayName() = %q, want %q", result, expected)
	}
}

func TestNote_DisplayName_EmptyTitle(t *testing.T) {
	note := Note{}
	note.File.Relative = "notes/hello-world.md"
	note.File.Filepath = "/path/to/notes/hello-world.md"
	note.Metadata = map[string]any{
		"title": "",
	}

	expected := "hello-world"
	if result := note.DisplayName(); result != expected {
		t.Errorf("DisplayName() = %q, want %q", result, expected)
	}
}

func TestNote_DisplayName_SpecialCharacters(t *testing.T) {
	note := Note{}
	note.File.Relative = "notes/My Special Note!.md"
	note.File.Filepath = "/path/to/notes/My Special Note!.md"
	note.Metadata = map[string]any{}

	expected := "my-special-note"
	if result := note.DisplayName(); result != expected {
		t.Errorf("DisplayName() = %q, want %q", result, expected)
	}
}

func TestTuiRender_NoteList_WithNotes(t *testing.T) {
	// Create notes with the anonymous struct format
	note1 := Note{}
	note1.File.Relative = "notes/test1.md"
	note1.File.Filepath = "/path/to/notes/test1.md"
	note1.Metadata = map[string]any{
		"title": "Test Note One",
	}

	note2 := Note{}
	note2.File.Relative = "notes/test2.md"
	note2.File.Filepath = "/path/to/notes/test2.md"
	note2.Metadata = map[string]any{}

	ctx := map[string]any{
		"Notes": []Note{note1, note2},
	}

	result, err := TuiRender("note-list", ctx)
	if err != nil {
		t.Fatalf("TuiRender() failed: %v", err)
	}

	// Check for note count in header
	if !strings.Contains(result, "(2)") {
		t.Errorf("TuiRender() result = %q, want to contain '(2)' in header", result)
	}

	// Check for note 1 with title
	if !strings.Contains(result, "Test Note One") {
		t.Errorf("TuiRender() result = %q, want to contain 'Test Note One'", result)
	}

	if !strings.Contains(result, "notes/test1.md") {
		t.Errorf("TuiRender() result = %q, want to contain 'notes/test1.md'", result)
	}

	// Check for note 2 with slugified name
	if !strings.Contains(result, "test2") {
		t.Errorf("TuiRender() result = %q, want to contain 'test2'", result)
	}

	if !strings.Contains(result, "notes/test2.md") {
		t.Errorf("TuiRender() result = %q, want to contain 'notes/test2.md'", result)
	}
}

func TestTuiRender_NoteList_Empty(t *testing.T) {
	ctx := map[string]any{
		"Notes": []Note{},
	}

	result, err := TuiRender("note-list", ctx)
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

	result, err := TuiRender("notebook-info", ctx)
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

	result, err := TuiRender("notebook-info", ctx)
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

	result, err := TuiRender("notebook-list", ctx)
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

	result, err := TuiRender("notebook-list", ctx)
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

	result, err := TuiRender("note-detail", ctx)
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

func TestTemplates_Loaded(t *testing.T) {
	// Ensure all templates are loaded
	templateNames := []string{"note-list", "note-detail", "notebook-info", "notebook-list"}

	for _, name := range templateNames {
		t.Run(name, func(t *testing.T) {
			if _, ok := loadedTemplates[name]; !ok {
				t.Errorf("Template %q not loaded", name)
			}
		})
	}
}

func TestTuiRender_InvalidTemplateName_Error(t *testing.T) {
	// Invalid template name should return error
	ctx := map[string]string{"Title": "Test"}

	_, err := TuiRender("nonexistent-template", ctx)
	if err == nil {
		t.Fatalf("TuiRender() should fail on invalid template name")
	}
}

func TestTuiRender_NilContext(t *testing.T) {
	// TuiRender with a known template and nil context
	// The note-list template doesn't use context, so this should work
	result, err := TuiRender("note-list", map[string]any{"Notes": []Note{}})
	if err != nil {
		t.Fatalf("TuiRender() failed with empty notes: %v", err)
	}

	if !strings.Contains(result, "No notes found") {
		t.Errorf("TuiRender() result = %q, want to contain 'No notes found'", result)
	}
}

func TestTuiRender_TemplateNotFound(t *testing.T) {
	ctx := map[string]string{"Name": "World"}
	output, err := TuiRender("nonexistent-template", ctx)

	if err == nil {
		t.Error("expected error for nonexistent template")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error message should contain 'not found', got: %v", err)
	}
	if output != "" {
		t.Errorf("expected empty output on error, got: %s", output)
	}
}

func TestTuiRender_ValidTemplate_ValidContext(t *testing.T) {
	// Test successful case with valid template and context
	// Uses the existing note-list template
	note := Note{}
	note.File.Relative = "notes/test.md"
	note.File.Filepath = "/path/to/notes/test.md"
	note.Metadata = map[string]any{"title": "Test Note"}

	ctx := map[string]any{"Notes": []Note{note}}

	output, err := TuiRender("note-list", ctx)

	if err != nil {
		t.Errorf("TuiRender() unexpected error: %v", err)
	}
	if output == "" {
		t.Error("expected non-empty output")
	}
	if !strings.Contains(output, "Test Note") && !strings.Contains(output, "test") {
		t.Errorf("output should contain note reference, got: %s", output)
	}
}
