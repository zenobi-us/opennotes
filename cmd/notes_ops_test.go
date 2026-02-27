package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckNoteExists_ReturnsNotFoundExitCode(t *testing.T) {
	root := t.TempDir()

	_, err := checkNoteExists(root, "missing")
	if err == nil {
		t.Fatalf("expected missing note error")
	}
	if ExitCode(err) != ExitCodeNotFound {
		t.Fatalf("expected not-found exit code %d, got %d", ExitCodeNotFound, ExitCode(err))
	}
}

func TestCheckNoteExists_ReturnsSuccessWhenPresent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "a.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	result, err := checkNoteExists(root, "docs/a")
	if err != nil {
		t.Fatalf("checkNoteExists returned error: %v", err)
	}
	if !result.Exists {
		t.Fatalf("expected exists=true")
	}
}

func TestEnsureNoteFile_IsIdempotent(t *testing.T) {
	root := t.TempDir()

	first, err := ensureNoteFile(root, "docs/task")
	if err != nil {
		t.Fatalf("first ensure failed: %v", err)
	}
	if !first.Created {
		t.Fatalf("expected first ensure to create file")
	}

	second, err := ensureNoteFile(root, "docs/task")
	if err != nil {
		t.Fatalf("second ensure failed: %v", err)
	}
	if second.Created {
		t.Fatalf("expected second ensure to be non-creating")
	}
}

func TestAppendToNoteFile_AppendsContent(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "log.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte("line1\n"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := appendToNoteFile(root, "docs/log", []byte("line2\n"), false)
	if err != nil {
		t.Fatalf("appendToNoteFile failed: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "line1\nline2\n" {
		t.Fatalf("unexpected content after append: %q", string(got))
	}
}

func TestAppendToNoteFile_MissingWithoutCreateReturnsNotFoundExitCode(t *testing.T) {
	root := t.TempDir()

	_, err := appendToNoteFile(root, "docs/missing", []byte("x"), false)
	if err == nil {
		t.Fatalf("expected append without create to fail")
	}
	if ExitCode(err) != ExitCodeNotFound {
		t.Fatalf("expected not-found exit code %d, got %d", ExitCodeNotFound, ExitCode(err))
	}
}

func TestNotesOpsCommands_RegisteredAndFlagged(t *testing.T) {
	if notesExistsCmd.Flags().Lookup("format") == nil {
		t.Fatalf("notes exists should define --format")
	}
	if notesEnsureCmd.Flags().Lookup("format") == nil {
		t.Fatalf("notes ensure should define --format")
	}
	if notesAppendCmd.Flags().Lookup("input") == nil {
		t.Fatalf("notes append should define --input")
	}
	if notesAppendCmd.Flags().Lookup("create") == nil {
		t.Fatalf("notes append should define --create")
	}
}
