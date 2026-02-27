package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveMovePath_RejectsTraversal(t *testing.T) {
	root := t.TempDir()

	_, _, err := resolveMovePath(root, "../outside")
	if err == nil {
		t.Fatalf("expected traversal path to fail")
	}
}

func TestMoveNoteFile_SourceMissingReturnsNotFoundExitCode(t *testing.T) {
	root := t.TempDir()

	_, err := moveNoteFile(root, "docs/missing", "archive/missing", false)
	if err == nil {
		t.Fatalf("expected missing source to fail")
	}
	if ExitCode(err) != ExitCodeNotFound {
		t.Fatalf("expected not-found exit code %d, got %d", ExitCodeNotFound, ExitCode(err))
	}
}

func TestMoveNoteFile_DestinationExistsWithoutForceReturnsConflictExitCode(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "docs", "task.md")
	destinationPath := filepath.Join(root, "archive", "task.md")

	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatalf("mkdir source failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		t.Fatalf("mkdir destination failed: %v", err)
	}
	if err := os.WriteFile(sourcePath, []byte("source"), 0o644); err != nil {
		t.Fatalf("write source failed: %v", err)
	}
	if err := os.WriteFile(destinationPath, []byte("destination"), 0o644); err != nil {
		t.Fatalf("write destination failed: %v", err)
	}

	_, err := moveNoteFile(root, "docs/task", "archive/task", false)
	if err == nil {
		t.Fatalf("expected destination conflict to fail")
	}
	if ExitCode(err) != ExitCodeConflict {
		t.Fatalf("expected conflict exit code %d, got %d", ExitCodeConflict, ExitCode(err))
	}
}

func TestMoveNoteFile_ForceOverwritesAndPreservesContentAndMetadata(t *testing.T) {
	root := t.TempDir()
	sourcePath := filepath.Join(root, "tasks", "done.md")
	destinationPath := filepath.Join(root, "archive", "done.md")
	sourceContent := "---\ntitle: Done Task\nstatus: complete\n---\n\n# Done\n"

	if err := os.MkdirAll(filepath.Dir(sourcePath), 0o755); err != nil {
		t.Fatalf("mkdir source failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		t.Fatalf("mkdir destination failed: %v", err)
	}
	if err := os.WriteFile(sourcePath, []byte(sourceContent), 0o644); err != nil {
		t.Fatalf("write source failed: %v", err)
	}
	if err := os.WriteFile(destinationPath, []byte("old"), 0o644); err != nil {
		t.Fatalf("write destination failed: %v", err)
	}

	result, err := moveNoteFile(root, "tasks/done", "archive/done", true)
	if err != nil {
		t.Fatalf("moveNoteFile returned error: %v", err)
	}
	if !result.Overwritten {
		t.Fatalf("expected overwritten=true")
	}
	if result.Source != "tasks/done.md" {
		t.Fatalf("expected normalized source path, got %q", result.Source)
	}
	if result.Destination != "archive/done.md" {
		t.Fatalf("expected normalized destination path, got %q", result.Destination)
	}

	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Fatalf("expected source to be moved away")
	}

	movedContent, err := os.ReadFile(destinationPath)
	if err != nil {
		t.Fatalf("read destination failed: %v", err)
	}
	if string(movedContent) != sourceContent {
		t.Fatalf("expected destination content to match source")
	}
}

func TestNotesCommand_RegistersMoveSubcommand(t *testing.T) {
	found := false
	for _, sub := range notesCmd.Commands() {
		if sub == notesMvCmd {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected notes mv command to be registered")
	}
}

func TestNotesMvCommand_HasForceAndFormatFlags(t *testing.T) {
	forceFlag := notesMvCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Fatalf("expected --force flag to be defined")
	}

	formatFlag := notesMvCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatalf("expected --format flag to be defined")
	}
	if formatFlag.DefValue != "list" {
		t.Fatalf("expected --format default to be list, got %q", formatFlag.DefValue)
	}
}

func TestRenderNoteMoveFailure_List(t *testing.T) {
	output := captureStdout(t, func() {
		renderNoteMoveFailure("list", "tasks/a.md", "archive/a.md", "source note not found")
	})

	if !strings.Contains(output, "status=failure") {
		t.Fatalf("expected failure status output, got %q", output)
	}
	if !strings.Contains(output, "source=\"tasks/a.md\"") {
		t.Fatalf("expected source in output, got %q", output)
	}
}
