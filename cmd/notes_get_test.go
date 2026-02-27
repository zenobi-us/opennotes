package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zenobi-us/jot/internal/services"
)

func TestLoadNoteByPath_ReadsMarkdownFile(t *testing.T) {
	root := t.TempDir()
	notesDir := filepath.Join(root, "docs")
	if err := os.MkdirAll(notesDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	notePath := filepath.Join(notesDir, "plan.md")
	content := "---\ntitle: Planning\nstatus: active\n---\n\n# Plan\n"
	if err := os.WriteFile(notePath, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	nb := &services.Notebook{}
	nb.Config.Root = root

	note, err := loadNoteByPath(nb, "docs/plan")
	if err != nil {
		t.Fatalf("loadNoteByPath returned error: %v", err)
	}

	if note.File.Relative != "docs/plan.md" {
		t.Fatalf("expected relative path docs/plan.md, got %q", note.File.Relative)
	}
	if note.Metadata["title"] != "Planning" {
		t.Fatalf("expected title metadata, got %v", note.Metadata["title"])
	}
}

func TestLoadNoteByPath_RejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	nb := &services.Notebook{}
	nb.Config.Root = root

	_, err := loadNoteByPath(nb, "../outside")
	if err == nil {
		t.Fatalf("expected path traversal to fail")
	}
}

func TestNotesGetCommand_HasRawFlag(t *testing.T) {
	rawFlag := notesGetCmd.Flags().Lookup("raw")
	if rawFlag == nil {
		t.Fatalf("expected --raw flag to be defined")
	}
	if rawFlag.DefValue != "false" {
		t.Fatalf("expected --raw default false, got %q", rawFlag.DefValue)
	}
}

func TestValidateGetOutputFlags_RawIncompatibleWithJSON(t *testing.T) {
	err := validateGetOutputFlags("json", true)
	if err == nil {
		t.Fatalf("expected raw+json to fail")
	}
}

func TestLoadRawNoteByPath_ReturnsExactBytes(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "docs", "raw.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	raw := []byte("---\ntitle: Raw\n---\n\nBody line 1\nBody line 2\n")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	nb := &services.Notebook{}
	nb.Config.Root = root

	got, err := loadRawNoteByPath(nb, "docs/raw")
	if err != nil {
		t.Fatalf("loadRawNoteByPath returned error: %v", err)
	}

	if string(got) != string(raw) {
		t.Fatalf("expected exact raw bytes match")
	}
}
