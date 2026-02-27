package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveUpdateTargetPath_RejectsTraversal(t *testing.T) {
	root := t.TempDir()

	_, _, err := resolveUpdateTargetPath(root, "../outside")
	if err == nil {
		t.Fatalf("expected traversal path to fail")
	}
}

func TestUpdateNoteFile_ReplaceExisting(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "docs", "plan.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(target, []byte("old"), 0o644); err != nil {
		t.Fatalf("seed write failed: %v", err)
	}

	result, err := updateNoteFile(root, "docs/plan", []byte("new"), false)
	if err != nil {
		t.Fatalf("updateNoteFile returned error: %v", err)
	}
	if result.Created {
		t.Fatalf("expected replace operation, got created=true")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("expected updated content, got %q", string(got))
	}
}

func TestUpdateNoteFile_MissingWithoutCreateFails(t *testing.T) {
	root := t.TempDir()

	_, err := updateNoteFile(root, "docs/missing", []byte("new"), false)
	if err == nil {
		t.Fatalf("expected missing target to fail without --create")
	}
}

func TestUpdateNoteFile_CreateWhenEnabled(t *testing.T) {
	root := t.TempDir()

	result, err := updateNoteFile(root, "docs/new-note", []byte("new"), true)
	if err != nil {
		t.Fatalf("updateNoteFile returned error: %v", err)
	}
	if !result.Created {
		t.Fatalf("expected create operation, got created=false")
	}

	target := filepath.Join(root, "docs", "new-note.md")
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("expected created content, got %q", string(got))
	}
}

func TestReadUpdateContent_FileInput(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "payload.md")
	if err := os.WriteFile(inputPath, []byte("from-file"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	content, err := readUpdateContent(inputPath, strings.NewReader(""), false)
	if err != nil {
		t.Fatalf("readUpdateContent returned error: %v", err)
	}
	if string(content) != "from-file" {
		t.Fatalf("expected file content, got %q", string(content))
	}
}

func TestReadUpdateContent_StdinInput(t *testing.T) {
	content, err := readUpdateContent("", strings.NewReader("from-stdin"), true)
	if err != nil {
		t.Fatalf("readUpdateContent returned error: %v", err)
	}
	if string(content) != "from-stdin" {
		t.Fatalf("expected stdin content, got %q", string(content))
	}
}

func TestReadUpdateContent_RejectsMixedSources(t *testing.T) {
	dir := t.TempDir()
	inputPath := filepath.Join(dir, "payload.md")
	if err := os.WriteFile(inputPath, []byte("from-file"), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	_, err := readUpdateContent(inputPath, strings.NewReader("from-stdin"), true)
	if err == nil {
		t.Fatalf("expected mixed sources to fail")
	}
}

func TestReadUpdateContent_RequiresSource(t *testing.T) {
	_, err := readUpdateContent("", strings.NewReader(""), false)
	if err == nil {
		t.Fatalf("expected error when no input source provided")
	}
}

func TestEmitNoteUpdateResult_JSON(t *testing.T) {
	result := noteUpdateResult{Status: "success", Path: "docs/plan.md", Created: false}

	output := captureStdout(t, func() {
		if err := emitNoteUpdateResult(result, "json"); err != nil {
			t.Fatalf("emitNoteUpdateResult returned error: %v", err)
		}
	})

	decoded := map[string]any{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("expected valid json output, got %v", err)
	}
	if decoded["status"] != "success" {
		t.Fatalf("expected status=success, got %v", decoded["status"])
	}
}

func TestEmitNoteUpdateResult_List(t *testing.T) {
	result := noteUpdateResult{Status: "success", Path: "docs/plan.md", Created: true}

	output := captureStdout(t, func() {
		if err := emitNoteUpdateResult(result, "list"); err != nil {
			t.Fatalf("emitNoteUpdateResult returned error: %v", err)
		}
	})

	if !strings.Contains(output, "status=success") {
		t.Fatalf("expected machine-list status output, got %q", output)
	}
	if !strings.Contains(output, "created=true") {
		t.Fatalf("expected created flag in output, got %q", output)
	}
}

func TestNotesUpdateCommand_HasPutAlias(t *testing.T) {
	found := false
	for _, alias := range notesUpdateCmd.Aliases {
		if alias == "put" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected notes update alias to include put")
	}
}

func TestNotesCommand_RegistersUpdateSubcommand(t *testing.T) {
	found := false
	for _, sub := range notesCmd.Commands() {
		if sub == notesUpdateCmd {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected notes update command to be registered")
	}
}

func TestRenderNoteUpdateFailure_JSON(t *testing.T) {
	output := captureStdout(t, func() {
		renderNoteUpdateFailure("json", "docs/missing.md", "note not found")
	})

	decoded := map[string]any{}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("expected valid json output, got %v", err)
	}
	if decoded["status"] != "failure" {
		t.Fatalf("expected status=failure, got %v", decoded["status"])
	}
}

func TestRenderNoteUpdateFailure_List(t *testing.T) {
	buf := &bytes.Buffer{}
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	os.Stdout = w

	renderNoteUpdateFailure("list", "docs/missing.md", "note not found")

	if err := w.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	os.Stdout = original
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("read failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "status=failure") {
		t.Fatalf("expected machine-list failure output, got %q", output)
	}
}
