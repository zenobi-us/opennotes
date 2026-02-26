package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testEnv holds shared test environment state.
type testEnv struct {
	binaryPath string
	tmpDir     string
	t          *testing.T
}

// newTestEnv creates a new test environment.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()

	// Use pre-built binary from dist/jot or build if not exists
	binaryPath := filepath.Join(getRootDir(), "dist", "jot")

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("Binary not found at %s - run 'mise run go-build' first", binaryPath)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "jot-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		_ = os.RemoveAll(tmpDir)
	})

	return &testEnv{
		binaryPath: binaryPath,
		tmpDir:     tmpDir,
		t:          t,
	}
}

// getRootDir returns the project root directory.
func getRootDir() string {
	// Navigate from tests/e2e to project root
	dir, err := os.Getwd()
	if err != nil {
		return "../.."
	}

	// Try to find project root by looking for go.mod
	for i := 0; i < 5; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}

	return "../.."
}

// run executes the CLI with given args.
func (e *testEnv) run(args ...string) (stdout, stderr string, exitCode int) {
	return e.runInDir(e.tmpDir, args...)
}

// runInDir executes the CLI with given args in a specific directory.
func (e *testEnv) runInDir(dir string, args ...string) (stdout, stderr string, exitCode int) {
	e.t.Helper()

	cmd := exec.Command(e.binaryPath, args...)
	cmd.Dir = dir

	// Use isolated config directory
	configDir := filepath.Join(e.tmpDir, ".config")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("HOME=%s", e.tmpDir),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", configDir),
	)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			e.t.Logf("Error running command: %v", err)
			exitCode = -1
		}
	}

	return stdoutBuf.String(), stderrBuf.String(), exitCode
}

// createNotebook creates a test notebook directory.
func (e *testEnv) createNotebook(name string) string {
	e.t.Helper()

	notebookDir := filepath.Join(e.tmpDir, name)
	if err := os.MkdirAll(notebookDir, 0755); err != nil {
		e.t.Fatalf("failed to create notebook dir: %v", err)
	}

	// Create notes subdirectory (mimics what 'notebook create' does)
	notesDir := filepath.Join(notebookDir, ".notes")
	if err := os.MkdirAll(notesDir, 0755); err != nil {
		e.t.Fatalf("failed to create notes dir: %v", err)
	}

	// Create .jot.json with relative root path (as the CLI does)
	config := map[string]interface{}{
		"name":     name,
		"root":     ".notes", // Relative path!
		"contexts": []string{notebookDir},
		"groups": []map[string]interface{}{
			{
				"name":     "Default",
				"globs":    []string{"**/*.md"},
				"metadata": map[string]interface{}{},
			},
		},
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		e.t.Fatalf("failed to marshal config: %v", err)
	}

	configPath := filepath.Join(notebookDir, ".jot.json")
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		e.t.Fatalf("failed to write config: %v", err)
	}

	return notebookDir
}

// createNote creates a markdown note in the notebook's .notes directory.
func (e *testEnv) createNote(notebookDir, filename, content string) string {
	e.t.Helper()

	// Notes are stored in the .notes subdirectory
	notesDir := filepath.Join(notebookDir, ".notes")
	notePath := filepath.Join(notesDir, filename)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(notePath), 0755); err != nil {
		e.t.Fatalf("failed to create note dir: %v", err)
	}

	if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
		e.t.Fatalf("failed to write note: %v", err)
	}

	return notePath
}

// createNoteWithFrontmatter creates a note with YAML frontmatter.
func (e *testEnv) createNoteWithFrontmatter(notebookDir, filename string, frontmatter map[string]interface{}, body string) string {
	e.t.Helper()

	var content strings.Builder
	content.WriteString("---\n")
	for key, value := range frontmatter {
		switch v := value.(type) {
		case string:
			content.WriteString(fmt.Sprintf("%s: %q\n", key, v))
		case []string:
			content.WriteString(fmt.Sprintf("%s: [%s]\n", key, strings.Join(v, ", ")))
		default:
			data, _ := json.Marshal(v)
			content.WriteString(fmt.Sprintf("%s: %s\n", key, string(data)))
		}
	}
	content.WriteString("---\n\n")
	content.WriteString(body)

	return e.createNote(notebookDir, filename, content.String())
}

// === Init Command Tests ===

func TestCLI_Init_CreatesConfig(t *testing.T) {
	env := newTestEnv(t)

	stdout, stderr, exitCode := env.run("init")

	if exitCode != 0 {
		t.Errorf("init failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "initialized") {
		t.Errorf("expected 'initialized' in output, got: %s", stdout)
	}

	// Verify config file was created
	configPath := filepath.Join(env.tmpDir, ".config", "jot", "config.json")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("config file not created at %s", configPath)
	}
}

// === Notebook Command Tests ===

func TestCLI_NotebookCreate_CreatesNotebook(t *testing.T) {
	env := newTestEnv(t)

	notebookPath := filepath.Join(env.tmpDir, "my-notebook")
	// Path is positional arg, not --path flag
	stdout, stderr, exitCode := env.run("notebook", "create", notebookPath, "--name", "My Test Notebook")

	if exitCode != 0 {
		t.Errorf("notebook create failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "My Test Notebook") {
		t.Errorf("expected notebook name in output, got: %s", stdout)
	}

	// Verify notebook directory structure
	if _, err := os.Stat(filepath.Join(notebookPath, ".jot.json")); os.IsNotExist(err) {
		t.Error("notebook config file not created")
	}

	// CLI creates .notes directory, not "notes"
	if _, err := os.Stat(filepath.Join(notebookPath, ".notes")); os.IsNotExist(err) {
		t.Error(".notes directory not created")
	}
}

func TestCLI_NotebookList_ShowsNotebooks(t *testing.T) {
	env := newTestEnv(t)

	// First create a notebook with --register so it appears in list
	notebookPath := filepath.Join(env.tmpDir, "list-test-notebook")
	stdout, stderr, exitCode := env.run("notebook", "create", notebookPath, "--name", "List Test", "--register")
	if exitCode != 0 {
		t.Fatalf("failed to create notebook for list test: stderr=%s stdout=%s", stderr, stdout)
	}

	// Now list notebooks
	stdout, stderr, exitCode = env.run("notebook", "list")

	if exitCode != 0 {
		t.Errorf("notebook list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "List Test") {
		t.Errorf("expected 'List Test' in output, got: %s", stdout)
	}
}

func TestCLI_NotebookRegister_RegistersNotebook(t *testing.T) {
	env := newTestEnv(t)

	// Create a notebook manually
	notebookDir := env.createNotebook("register-test")

	// Register it
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notebook", "register")

	if exitCode != 0 {
		t.Errorf("notebook register failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Output says "Registered notebook" with capital R
	if !strings.Contains(stdout, "Registered") || !strings.Contains(stdout, "register-test") {
		t.Errorf("expected registration confirmation, got: %s", stdout)
	}

	// Verify it shows in list
	stdout, _, exitCode = env.run("notebook", "list")
	if exitCode != 0 {
		t.Error("notebook list failed after registration")
	}

	if !strings.Contains(stdout, "register-test") {
		t.Errorf("registered notebook not in list, got: %s", stdout)
	}
}

func TestCLI_NotebookAddContext_AddsContext(t *testing.T) {
	env := newTestEnv(t)

	// Create a notebook
	notebookDir := env.createNotebook("context-test")
	contextPath := filepath.Join(env.tmpDir, "my-project")
	_ = os.MkdirAll(contextPath, 0755)

	// Add context - path is positional arg, not --path flag
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notebook", "add-context", contextPath)

	if exitCode != 0 {
		t.Errorf("add-context failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Output says "Added context" not "Context added"
	if !strings.Contains(stdout, "Added context") {
		t.Errorf("expected 'Added context' in output, got: %s", stdout)
	}

	// Verify context in config
	configPath := filepath.Join(notebookDir, ".jot.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	if !strings.Contains(string(data), contextPath) {
		t.Errorf("context path not in config, got: %s", string(data))
	}
}

func TestCLI_Notebook_DisplaysInfo(t *testing.T) {
	env := newTestEnv(t)

	// Create a notebook
	notebookDir := env.createNotebook("info-test")
	env.createNote(notebookDir, "note1.md", "# Note 1\n\nContent")
	env.createNote(notebookDir, "note2.md", "# Note 2\n\nContent")

	// Display notebook info
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notebook")

	if exitCode != 0 {
		t.Errorf("notebook info failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "info-test") {
		t.Errorf("expected notebook name in output, got: %s", stdout)
	}
}

// === Notes Command Tests ===

func TestCLI_NotesList_ShowsNotes(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with notes
	notebookDir := env.createNotebook("notes-list-test")
	env.createNote(notebookDir, "note1.md", "# Note 1\n\nFirst note")
	env.createNote(notebookDir, "note2.md", "# Note 2\n\nSecond note")
	env.createNote(notebookDir, "note3.md", "# Note 3\n\nThird note")

	// List notes
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "list")

	if exitCode != 0 {
		t.Errorf("notes list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	noteCount := strings.Count(stdout, ".md")
	if noteCount < 3 {
		t.Errorf("expected at least 3 notes in output, got %d: %s", noteCount, stdout)
	}
}

func TestCLI_NotesList_EmptyNotebook(t *testing.T) {
	env := newTestEnv(t)

	// Create empty notebook
	notebookDir := env.createNotebook("empty-notebook")

	// List notes - currently DuckDB fails on empty directory (no matching files)
	// This is a known limitation: read_markdown() requires at least one file
	_, stderr, exitCode := env.runInDir(notebookDir, "notes", "list")

	// Currently this returns error when no files exist
	// TODO: Consider handling empty directory gracefully in the CLI
	if exitCode == 0 {
		// If this starts passing, the CLI was improved to handle empty notebooks
		t.Log("notes list handles empty notebook gracefully now")
	} else {
		// Expected behavior: fails with "File or directory does not exist"
		if !strings.Contains(stderr, "does not exist") && !strings.Contains(stderr, "no notes") {
			t.Errorf("unexpected error for empty notebook, stderr: %s", stderr)
		}
	}
}

func TestCLI_NotesSearch_FiltersNotes(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with notes
	notebookDir := env.createNotebook("search-test")
	env.createNote(notebookDir, "apple.md", "# Apple\n\nThis is about apples")
	env.createNote(notebookDir, "banana.md", "# Banana\n\nThis is about bananas")
	env.createNote(notebookDir, "cherry.md", "# Cherry\n\nThis is about cherries")

	// Search for body content explicitly (fieldless search is title-only)
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "search", "body:apple")

	if exitCode != 0 {
		t.Errorf("notes search failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "apple.md") {
		t.Errorf("expected apple.md in output, got: %s", stdout)
	}

	// Should not contain other notes
	if strings.Contains(stdout, "banana.md") {
		t.Errorf("unexpected banana.md in output: %s", stdout)
	}
}

func TestCLI_NotesAdd_CreatesNote(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("add-note-test")

	// Add a note
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "add", "--title", "My New Note")

	if exitCode != 0 {
		t.Errorf("notes add failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Created note") {
		t.Errorf("expected 'Created note' message, got: %s", stdout)
	}

	// Verify note was created in .notes directory
	notesDir := filepath.Join(notebookDir, ".notes")
	entries, err := os.ReadDir(notesDir)
	if err != nil {
		t.Fatalf("failed to read notes dir: %v", err)
	}

	found := false
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".md") {
			found = true
			break
		}
	}

	if !found {
		t.Error("note file not created")
	}
}

func TestCLI_NotesRemove_RemovesNote(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with a note
	notebookDir := env.createNotebook("remove-test")
	notePath := env.createNote(notebookDir, "to-delete.md", "# Delete Me\n\nThis will be deleted")

	// Verify note exists
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		t.Fatal("note was not created")
	}

	// Remove the note with --force
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "remove", "--force", "to-delete.md")

	if exitCode != 0 {
		t.Errorf("notes remove failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Removed note") {
		t.Errorf("expected 'Removed note' message, got: %s", stdout)
	}

	// Verify note was removed
	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Error("note was not deleted")
	}
}

// === Advanced Scenarios ===

func TestCLI_NestedMarkdownFiles(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with nested notes
	notebookDir := env.createNotebook("nested-test")
	env.createNote(notebookDir, "root-note.md", "# Root Note\n\nAt root level")
	env.createNote(notebookDir, "folder/nested.md", "# Nested Note\n\nIn a folder")
	env.createNote(notebookDir, "folder/deep/deeper.md", "# Deep Note\n\nDeep nested")

	// List notes
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "list")

	if exitCode != 0 {
		t.Errorf("notes list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Should find nested note
	if !strings.Contains(stdout, "nested.md") {
		t.Errorf("expected nested.md in output, got: %s", stdout)
	}
}

func TestCLI_NotesWithFrontmatter(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with frontmatter note
	notebookDir := env.createNotebook("frontmatter-test")
	env.createNoteWithFrontmatter(notebookDir, "tagged.md",
		map[string]interface{}{
			"title": "Tagged Note",
			"tags":  []string{"test", "smoke"},
			"date":  "2024-01-08",
		},
		"# Tagged Note\n\nThis has frontmatter.")

	// List notes
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "list")

	if exitCode != 0 {
		t.Errorf("notes list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "tagged.md") {
		t.Errorf("expected tagged.md in output, got: %s", stdout)
	}
}

func TestCLI_SpecialCharactersInFilenames(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with various filenames
	notebookDir := env.createNotebook("special-chars-test")
	env.createNote(notebookDir, "note-with-dashes.md", "# Dashes\n\nContent")
	env.createNote(notebookDir, "note_with_underscores.md", "# Underscores\n\nContent")
	env.createNote(notebookDir, "note.multiple.dots.md", "# Dots\n\nContent")

	// List notes
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "list")

	if exitCode != 0 {
		t.Errorf("notes list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "note-with-dashes.md") {
		t.Errorf("expected note-with-dashes.md in output, got: %s", stdout)
	}

	if !strings.Contains(stdout, "note_with_underscores.md") {
		t.Errorf("expected note_with_underscores.md in output, got: %s", stdout)
	}
}

func TestCLI_LargeNotebook(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with many notes
	notebookDir := env.createNotebook("large-notebook-test")
	for i := 1; i <= 20; i++ {
		env.createNote(notebookDir, fmt.Sprintf("note%02d.md", i),
			fmt.Sprintf("# Note %d\n\nContent for note %d", i, i))
	}

	// List notes
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "list")

	if exitCode != 0 {
		t.Errorf("notes list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Count notes in output
	noteCount := strings.Count(stdout, ".md")
	if noteCount < 20 {
		t.Errorf("expected 20 notes in output, got %d", noteCount)
	}
}

func TestCLI_NotebookFlag(t *testing.T) {
	env := newTestEnv(t)

	// Create two notebooks
	nb1 := env.createNotebook("notebook-1")
	nb2 := env.createNotebook("notebook-2")

	env.createNote(nb1, "note-in-nb1.md", "# In NB1\n\nContent")
	env.createNote(nb2, "note-in-nb2.md", "# In NB2\n\nContent")

	// List notes from specific notebook using --notebook flag
	stdout, stderr, exitCode := env.run("notes", "list", "--notebook", nb2)

	if exitCode != 0 {
		t.Errorf("notes list with --notebook failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "note-in-nb2.md") {
		t.Errorf("expected note-in-nb2.md in output, got: %s", stdout)
	}
}

// === Error Handling Tests ===

func TestCLI_NoNotebookFound(t *testing.T) {
	env := newTestEnv(t)

	// Run notes list without a notebook (should error)
	emptyDir := filepath.Join(env.tmpDir, "empty")
	_ = os.MkdirAll(emptyDir, 0755)

	_, stderr, exitCode := env.runInDir(emptyDir, "notes", "list")

	// Should fail since no notebook is available
	if exitCode == 0 {
		t.Error("expected error when no notebook found")
	}

	if !strings.Contains(stderr, "notebook") {
		t.Errorf("expected notebook-related error, got: %s", stderr)
	}
}

func TestCLI_InvalidNotebookPath(t *testing.T) {
	env := newTestEnv(t)

	// Try to list notes from non-existent notebook
	_, stderr, exitCode := env.run("notes", "list", "--notebook", "/nonexistent/path")

	if exitCode == 0 {
		t.Error("expected error with invalid notebook path")
	}

	if stderr == "" {
		t.Error("expected error message in stderr")
	}
}

func TestCLI_HelpCommands(t *testing.T) {
	env := newTestEnv(t)

	// Test --help flag
	stdout, _, exitCode := env.run("--help")
	if exitCode != 0 {
		t.Error("--help should exit 0")
	}
	if !strings.Contains(stdout, "jot") {
		t.Errorf("help should contain 'jot', got: %s", stdout)
	}

	// Test subcommand help
	stdout, _, exitCode = env.run("notebook", "--help")
	if exitCode != 0 {
		t.Error("notebook --help should exit 0")
	}
	if !strings.Contains(stdout, "notebook") {
		t.Errorf("notebook help should contain 'notebook', got: %s", stdout)
	}

	// Test notes help
	stdout, _, exitCode = env.run("notes", "--help")
	if exitCode != 0 {
		t.Error("notes --help should exit 0")
	}
	if !strings.Contains(stdout, "notes") {
		t.Errorf("notes help should contain 'notes', got: %s", stdout)
	}
}

// === Command Aliases Tests ===

func TestCLI_NotebookAlias(t *testing.T) {
	env := newTestEnv(t)

	// Create a notebook
	notebookDir := env.createNotebook("alias-test")

	// Use 'nb' alias
	stdout, stderr, exitCode := env.runInDir(notebookDir, "nb")

	if exitCode != 0 {
		t.Errorf("nb alias failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "alias-test") {
		t.Errorf("expected notebook info, got: %s", stdout)
	}
}

func TestCLI_NotesListAlias(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with notes
	notebookDir := env.createNotebook("ls-alias-test")
	env.createNote(notebookDir, "test.md", "# Test\n\nContent")

	// Use 'notes ls' alias
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "ls")

	if exitCode != 0 {
		t.Errorf("notes ls alias failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "test.md") {
		t.Errorf("expected test.md in output, got: %s", stdout)
	}
}

func TestCLI_NotesRemoveAlias(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook with a note
	notebookDir := env.createNotebook("rm-alias-test")
	notePath := env.createNote(notebookDir, "to-remove.md", "# Remove\n\nContent")

	// Verify note exists
	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		t.Fatal("note was not created")
	}

	// Use 'notes rm' alias with --force
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "rm", "--force", "to-remove.md")

	if exitCode != 0 {
		t.Errorf("notes rm alias failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	if !strings.Contains(stdout, "Removed note") {
		t.Errorf("expected 'Removed note' message, got: %s", stdout)
	}

	// Verify note was removed
	if _, err := os.Stat(notePath); !os.IsNotExist(err) {
		t.Error("note was not deleted")
	}
}

// === SQL Flag Tests ===

// NOTE: TestCLI_SQLFlag_* tests removed as part of Phase 5 SQL-to-Bleve migration.
// The --sql flag has been replaced with structured query commands.
// See cmd/notes_search_query.go for the new boolean query interface.

// TestCLI_ViewDiscovery_EmptyCommandListsViews tests that running "notes view" without arguments lists views
func TestCLI_ViewDiscovery_EmptyCommandListsViews(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-discovery-test")

	// Run "notes view" with no arguments
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view")

	// Should succeed
	if exitCode != 0 {
		t.Errorf("view command with no args failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Should list views
	if !strings.Contains(stdout, "AVAILABLE VIEWS") {
		t.Errorf("expected 'AVAILABLE VIEWS' header in output, got: %s", stdout)
	}

	// Should show built-in views
	if !strings.Contains(stdout, "today") {
		t.Errorf("expected 'today' view in output, got: %s", stdout)
	}

	if !strings.Contains(stdout, "recent") {
		t.Errorf("expected 'recent' view in output, got: %s", stdout)
	}

	// Should show descriptions
	if !strings.Contains(stdout, "Notes created or updated today") {
		t.Errorf("expected view description in output, got: %s", stdout)
	}
}

// TestCLI_ViewDiscovery_ListFlag tests that --list flag shows all views
func TestCLI_ViewDiscovery_ListFlag(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-list-flag-test")

	// Run "notes view --list"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view", "--list")

	// Should succeed
	if exitCode != 0 {
		t.Errorf("view --list failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Should list views
	if !strings.Contains(stdout, "AVAILABLE VIEWS") {
		t.Errorf("expected 'AVAILABLE VIEWS' header, got: %s", stdout)
	}

	// Should show all built-in views
	builtinViews := []string{"today", "recent", "kanban", "untagged", "orphans", "broken-links"}
	for _, view := range builtinViews {
		if !strings.Contains(stdout, view) {
			t.Errorf("expected view '%s' in output, got: %s", view, stdout)
		}
	}
}

// TestCLI_ViewDiscovery_ListFlagJSON tests that --list --format json outputs valid JSON
func TestCLI_ViewDiscovery_ListFlagJSON(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-list-json-test")

	// Run "notes view --list --format json"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view", "--list", "--format", "json")

	// Should succeed
	if exitCode != 0 {
		t.Errorf("view --list --format json failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Parse JSON output
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("invalid JSON output: %v, output: %s", err, stdout)
	}

	// Should have "views" key
	views, ok := result["views"].([]interface{})
	if !ok {
		t.Errorf("expected 'views' array in JSON, got: %v", result)
	}

	// Should have at least 6 views (built-in)
	if len(views) < 6 {
		t.Errorf("expected at least 6 views, got %d", len(views))
	}

	// First view should have required fields
	firstView, ok := views[0].(map[string]interface{})
	if !ok {
		t.Errorf("view should be an object, got: %v", views[0])
	}

	if _, ok := firstView["name"]; !ok {
		t.Errorf("view should have 'name' field, got: %v", firstView)
	}

	if _, ok := firstView["origin"]; !ok {
		t.Errorf("view should have 'origin' field, got: %v", firstView)
	}

	if _, ok := firstView["description"]; !ok {
		t.Errorf("view should have 'description' field, got: %v", firstView)
	}
}

// TestCLI_ViewDiscovery_OriginInfo tests that views show origin information
func TestCLI_ViewDiscovery_OriginInfo(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-origin-test")

	// Add a custom view to notebook config using DSL query format
	configPath := filepath.Join(notebookDir, ".jot.json")
	configData := map[string]interface{}{
		"name": "Test Notebook",
		"root": "notes",
		"views": map[string]interface{}{
			"my-custom-view": map[string]interface{}{
				"name":        "my-custom-view",
				"description": "A custom notebook view",
				"query":       "| sort:modified:desc", // DSL format
			},
		},
	}

	configJSON, _ := json.Marshal(configData)
	if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Run "notes view --list --format json"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view", "--list", "--format", "json")

	if exitCode != 0 {
		t.Errorf("view --list --format json failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("invalid JSON output: %v", err)
	}

	views, _ := result["views"].([]interface{})

	// Check origins
	origins := make(map[string]int)
	for _, v := range views {
		view, _ := v.(map[string]interface{})
		origin, _ := view["origin"].(string)
		origins[origin]++
	}

	// Should have built-in views
	if origins["built-in"] == 0 {
		t.Errorf("expected at least one built-in view, got: %v", origins)
	}

	// Should have notebook views
	if origins["notebook"] == 0 {
		t.Errorf("expected at least one notebook view, got: %v", origins)
	}
}

// TestCLI_ViewDiscovery_ParameterDisplay tests that views show description in listings
// NOTE: DSL-based views use directives (e.g., "group:status") instead of parameters
func TestCLI_ViewDiscovery_ParameterDisplay(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-params-test")

	// Run "notes view"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view")

	if exitCode != 0 {
		t.Errorf("view command failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// kanban view should be present
	if !strings.Contains(stdout, "kanban") {
		t.Errorf("expected 'kanban' view in output")
	}

	// Should show kanban description
	if !strings.Contains(stdout, "status") {
		t.Errorf("expected 'status' in kanban view description")
	}

	// Should show "Built-in Views" header
	if !strings.Contains(stdout, "Built-in Views") {
		t.Errorf("expected 'Built-in Views' header in output")
	}

	// Should show all 6 builtin views
	builtins := []string{"today", "recent", "kanban", "untagged", "orphans", "broken-links"}
	for _, name := range builtins {
		if !strings.Contains(stdout, name) {
			t.Errorf("expected builtin view '%s' in output", name)
		}
	}
}

// TestCLI_ViewDiscovery_JSONParametersComplete tests that JSON output includes all view details
// NOTE: DSL-based views use directives instead of parameters, so parameters may be empty
func TestCLI_ViewDiscovery_JSONParametersComplete(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-json-params-test")

	// Run "notes view --list --format json"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view", "--list", "--format", "json")

	if exitCode != 0 {
		t.Errorf("view --list --format json failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("invalid JSON output: %v", err)
	}

	views, _ := result["views"].([]interface{})

	// Should have at least 6 builtin views
	if len(views) < 6 {
		t.Errorf("expected at least 6 builtin views, got: %d", len(views))
	}

	// Find kanban view
	var kanbanView map[string]interface{}
	for _, v := range views {
		view, _ := v.(map[string]interface{})
		if name, _ := view["name"].(string); name == "kanban" {
			kanbanView = view
			break
		}
	}

	if kanbanView == nil {
		t.Errorf("could not find kanban view in output")
		return
	}

	// Check required fields exist on kanban view
	if name, _ := kanbanView["name"].(string); name != "kanban" {
		t.Errorf("expected view name 'kanban', got: %s", name)
	}

	if origin, _ := kanbanView["origin"].(string); origin != "built-in" {
		t.Errorf("expected kanban origin 'built-in', got: %s", origin)
	}

	if desc, _ := kanbanView["description"].(string); desc == "" {
		t.Errorf("expected kanban view to have a description")
	}

	// DSL-based views may have empty parameters array, which is valid
	// The directive "group:status" handles grouping instead of parameters
}

// TestCLI_ViewDiscovery_Sorting tests that views are sorted by origin
func TestCLI_ViewDiscovery_Sorting(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("view-sorting-test")

	// Add custom views to both global config and notebook config
	configDir := filepath.Join(env.tmpDir, ".config", "jot")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	globalConfigPath := filepath.Join(configDir, "config.json")
	globalConfig := map[string]interface{}{
		"notebooks": []interface{}{},
		"views": map[string]interface{}{
			"global-view": map[string]interface{}{
				"name":        "global-view",
				"description": "A global view",
				"query": map[string]interface{}{
					"order_by": "updated DESC",
				},
			},
		},
	}

	globalJSON, _ := json.Marshal(globalConfig)
	if err := os.WriteFile(globalConfigPath, globalJSON, 0644); err != nil {
		t.Fatalf("failed to write global config: %v", err)
	}

	// Add notebook view
	notebookConfigPath := filepath.Join(notebookDir, ".jot.json")
	notebookConfig := map[string]interface{}{
		"name": "Test Notebook",
		"root": "notes",
		"views": map[string]interface{}{
			"notebook-view": map[string]interface{}{
				"name":        "notebook-view",
				"description": "A notebook view",
				"query": map[string]interface{}{
					"order_by": "updated DESC",
				},
			},
		},
	}

	notebookJSON, _ := json.Marshal(notebookConfig)
	if err := os.WriteFile(notebookConfigPath, notebookJSON, 0644); err != nil {
		t.Fatalf("failed to write notebook config: %v", err)
	}

	// Run "notes view --list --format json"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "view", "--list", "--format", "json")

	if exitCode != 0 {
		t.Errorf("view --list --format json failed with exit code %d, stderr: %s", exitCode, stderr)
	}

	// Parse JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Errorf("invalid JSON output: %v", err)
	}

	views, _ := result["views"].([]interface{})

	// Track origin order
	originOrder := []string{}
	lastOriginIndex := map[string]int{"built-in": 0, "global": 1, "notebook": 2}

	for _, v := range views {
		view, _ := v.(map[string]interface{})
		origin, _ := view["origin"].(string)

		// Skip if we've already seen this origin
		if len(originOrder) == 0 || originOrder[len(originOrder)-1] != origin {
			originOrder = append(originOrder, origin)
		}
	}

	// Verify ordering: built-in before global before notebook
	for i := 0; i < len(originOrder)-1; i++ {
		if lastOriginIndex[originOrder[i]] >= lastOriginIndex[originOrder[i+1]] {
			t.Errorf("views not properly sorted by origin, got order: %v", originOrder)
		}
	}
}

func TestCLI_ViewSaveDelete_SaveOverwriteDelete(t *testing.T) {
	env := newTestEnv(t)
	notebookDir := env.createNotebook("view-save-delete-test")

	stdout, stderr, exitCode := env.runInDir(
		notebookDir,
		"notes", "view", "--save", "work-inbox", "tag:work status:todo | sort:created:desc", "--description", "Work queue",
	)
	if exitCode != 0 {
		t.Fatalf("save view failed with exit code %d, stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "Saved notebook view 'work-inbox'") {
		t.Fatalf("expected save confirmation, got: %s", stdout)
	}

	stdout, stderr, exitCode = env.runInDir(
		notebookDir,
		"notes", "view", "--save", "work-inbox", "tag:work | sort:modified:desc",
	)
	if exitCode != 0 {
		t.Fatalf("overwrite view failed with exit code %d, stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "Updated notebook view 'work-inbox'") {
		t.Fatalf("expected overwrite confirmation, got: %s", stdout)
	}

	stdout, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--list", "--format", "json")
	if exitCode != 0 {
		t.Fatalf("view list failed with exit code %d, stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "work-inbox") {
		t.Fatalf("expected saved view in list output, got: %s", stdout)
	}

	configPath := filepath.Join(notebookDir, ".jot.json")
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed reading notebook config: %v", err)
	}
	if !strings.Contains(string(configBytes), "tag:work | sort:modified:desc") {
		t.Fatalf("expected overwritten query persisted, got: %s", string(configBytes))
	}

	stdout, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--delete", "work-inbox")
	if exitCode != 0 {
		t.Fatalf("delete view failed with exit code %d, stderr: %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "Deleted notebook view 'work-inbox'") {
		t.Fatalf("expected delete confirmation, got: %s", stdout)
	}

	stdout, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--list", "--format", "json")
	if exitCode != 0 {
		t.Fatalf("view list after delete failed with exit code %d, stderr: %s", exitCode, stderr)
	}
	if strings.Contains(stdout, "work-inbox") {
		t.Fatalf("expected deleted view to be absent, got: %s", stdout)
	}
}

func TestCLI_ViewSaveDelete_FailureModes(t *testing.T) {
	env := newTestEnv(t)
	notebookDir := env.createNotebook("view-save-delete-failure-test")

	_, stderr, exitCode := env.runInDir(notebookDir, "notes", "view", "--save", "work-inbox")
	if exitCode == 0 {
		t.Fatalf("expected save without query to fail")
	}
	if !strings.Contains(stderr, "--save requires exactly one query argument") {
		t.Fatalf("expected save usage error, got: %s", stderr)
	}

	_, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--save", "work-inbox", "tag:work", "--list")
	if exitCode == 0 {
		t.Fatalf("expected save/list conflict to fail")
	}
	if !strings.Contains(stderr, "cannot combine --save with --list") {
		t.Fatalf("expected conflict error, got: %s", stderr)
	}

	_, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--delete", "work-inbox", "extra")
	if exitCode == 0 {
		t.Fatalf("expected delete with positional args to fail")
	}
	if !strings.Contains(stderr, "--delete does not accept positional arguments") {
		t.Fatalf("expected delete argument error, got: %s", stderr)
	}

	_, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "today", "--description", "x")
	if exitCode == 0 {
		t.Fatalf("expected description without save to fail")
	}
	if !strings.Contains(stderr, "--description can only be used with --save") {
		t.Fatalf("expected description usage error, got: %s", stderr)
	}

	_, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--delete")
	if exitCode == 0 {
		t.Fatalf("expected delete without view name to fail")
	}
	if !strings.Contains(stderr, "flag needs an argument: --delete") {
		t.Fatalf("expected missing delete argument error, got: %s", stderr)
	}

	_, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--delete", "missing-view")
	if exitCode == 0 {
		t.Fatalf("expected deleting missing view to fail")
	}
	if !strings.Contains(stderr, "does not exist in notebook config") {
		t.Fatalf("expected missing view behavior text, got: %s", stderr)
	}

	_, stderr, exitCode = env.runInDir(notebookDir, "notes", "view", "--save", "bad", "tag:work | limit:not-a-number")
	if exitCode == 0 {
		t.Fatalf("expected invalid query save to fail")
	}
	if !strings.Contains(stderr, "invalid directives") {
		t.Fatalf("expected DSL validation failure, got: %s", stderr)
	}
}
