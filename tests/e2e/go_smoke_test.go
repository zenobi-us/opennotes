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

	// Use pre-built binary from dist/opennotes or build if not exists
	binaryPath := filepath.Join(getRootDir(), "dist", "opennotes")

	// Verify binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Skipf("Binary not found at %s - run 'mise run go-build' first", binaryPath)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "opennotes-e2e-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
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

	// Create .opennotes.json with relative root path (as the CLI does)
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

	configPath := filepath.Join(notebookDir, ".opennotes.json")
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
	configPath := filepath.Join(env.tmpDir, ".config", "opennotes", "config.json")
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
	if _, err := os.Stat(filepath.Join(notebookPath, ".opennotes.json")); os.IsNotExist(err) {
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
	os.MkdirAll(contextPath, 0755)

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
	configPath := filepath.Join(notebookDir, ".opennotes.json")
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

	// Search for "apple"
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "search", "apple")

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
	os.MkdirAll(emptyDir, 0755)

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
	if !strings.Contains(stdout, "opennotes") {
		t.Errorf("help should contain 'opennotes', got: %s", stdout)
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
