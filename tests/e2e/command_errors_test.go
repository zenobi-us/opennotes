package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// === Command Error Tests ===
//
// These tests focus on error scenarios to improve coverage of command error handling.
// They complement the existing happy-path tests in go_smoke_test.go.

func TestCLI_NotesAdd_FileAlreadyExists(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("add-error-test")

	// Create a note first
	env.createNote(notebookDir, "existing.md", "# Existing\n\nAlready exists")

	// Try to add a note with same name
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "add", "--title", "Test", "existing.md")

	if exitCode == 0 {
		t.Error("expected non-zero exit code when file already exists")
	}

	// Check for meaningful error message
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "already exists") && !strings.Contains(errOutput, "file exists") {
		t.Errorf("expected error about existing file, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_NotesAdd_DirectoryAsFilename(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("directory-error-test")

	// Create a subdirectory in .notes
	subDir := filepath.Join(notebookDir, ".notes", "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	// Try to add a note with the same name as the directory
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "add", "--title", "Test", "subdir")

	if exitCode == 0 {
		// This might actually succeed if CLI handles it gracefully
		t.Logf("CLI handled directory conflict gracefully")
		return
	}

	// Check for meaningful error message about directory conflict
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "directory") && !strings.Contains(errOutput, "exists") &&
	   !strings.Contains(errOutput, "conflict") {
		t.Errorf("expected error about directory conflict, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_NotesAdd_VeryLongFilename(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("long-filename-test")

	// Create a very long filename (most filesystems have limits around 255 chars)
	longName := strings.Repeat("a", 300) + ".md"

	// Try to add note with very long filename
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "add", "--title", "Test", longName)

	if exitCode == 0 {
		// If it succeeded, the filesystem and CLI can handle long names
		t.Logf("CLI and filesystem handle long filenames gracefully")
		return
	}

	// Check for error about filename length
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "long") && !strings.Contains(errOutput, "length") &&
	   !strings.Contains(errOutput, "filename") && !strings.Contains(errOutput, "name") {
		t.Errorf("expected error about filename length, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_NotesRemove_FileNotFound(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook (without the note we'll try to remove)
	notebookDir := env.createNotebook("remove-error-test")

	// Try to remove non-existent note
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "remove", "--force", "nonexistent.md")

	if exitCode == 0 {
		t.Error("expected non-zero exit code when removing non-existent file")
	}

	// Check for meaningful error message
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "not found") && !strings.Contains(errOutput, "does not exist") &&
	   !strings.Contains(errOutput, "no such file") {
		t.Errorf("expected error about file not found, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_Commands_PermissionDenied(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("permission-test")

	// Make the .notes directory read-only to simulate permission error
	notesDir := filepath.Join(notebookDir, ".notes")
	err := os.Chmod(notesDir, 0444) // read-only
	if err != nil {
		t.Skipf("Cannot change permissions for test: %v", err)
	}

	// Restore permissions after test
	t.Cleanup(func() {
		_ = os.Chmod(notesDir, 0755)
	})

	// Try to add note to read-only directory
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "add", "--title", "Test", "permission-test.md")

	if exitCode == 0 {
		t.Error("expected non-zero exit code when permission denied")
	}

	// Check for permission-related error
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "permission") && !strings.Contains(errOutput, "denied") &&
	   !strings.Contains(errOutput, "read-only") && !strings.Contains(errOutput, "cannot") {
		t.Errorf("expected permission error, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_NotebookFlag_InvalidPath(t *testing.T) {
	env := newTestEnv(t)

	// Try to use --notebook flag with non-existent path
	stdout, stderr, exitCode := env.run("notes", "list", "--notebook", "/nonexistent/path")

	if exitCode == 0 {
		t.Error("expected non-zero exit code for invalid notebook path")
	}

	// Check for meaningful error message
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "not found") && !strings.Contains(errOutput, "does not exist") &&
	   !strings.Contains(errOutput, "invalid") && !strings.Contains(errOutput, "notebook") {
		t.Errorf("expected error about invalid notebook, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_NotesSearch_WithoutNotebook(t *testing.T) {
	env := newTestEnv(t)

	// Try to search from a directory that's not a notebook (no .opennotes.json)
	emptyDir := filepath.Join(env.tmpDir, "empty-dir")
	err := os.MkdirAll(emptyDir, 0755)
	if err != nil {
		t.Fatalf("failed to create empty dir: %v", err)
	}

	// Try to search from non-notebook directory
	stdout, stderr, exitCode := env.runInDir(emptyDir, "notes", "search", "test")

	if exitCode == 0 {
		t.Error("expected non-zero exit code when not in notebook directory")
	}

	// Check for error about not being in notebook
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "notebook") && !strings.Contains(errOutput, "not found") &&
	   !strings.Contains(errOutput, "config") {
		t.Errorf("expected error about notebook not found, got stderr: %s, stdout: %s", stderr, stdout)
	}
}

func TestCLI_NotesAdd_EmptyTitle(t *testing.T) {
	env := newTestEnv(t)

	// Create notebook
	notebookDir := env.createNotebook("empty-title-test")

	// Try to add note with empty title
	stdout, stderr, exitCode := env.runInDir(notebookDir, "notes", "add", "--title", "", "test.md")

	if exitCode == 0 {
		// This might actually succeed if the CLI allows empty titles
		// Check if the note was created with empty content
		notePath := filepath.Join(notebookDir, ".notes", "test.md")
		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			t.Error("note should have been created even with empty title")
		}
		return // Test passed - empty title is allowed
	}

	// If it failed, check for meaningful error
	errOutput := stderr + stdout
	if !strings.Contains(errOutput, "title") && !strings.Contains(errOutput, "empty") &&
	   !strings.Contains(errOutput, "required") {
		t.Errorf("expected error about empty title, got stderr: %s, stdout: %s", stderr, stdout)
	}
}