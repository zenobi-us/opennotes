package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/zenobi-us/jot/internal/services"
)

func TestValidateOutputFormat_AllowsListAndJSON(t *testing.T) {
	for _, format := range []string{"list", "json"} {
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			t.Fatalf("validateOutputFormat(%q) returned error: %v", format, err)
		}
	}
}

func TestValidateOutputFormat_RejectsInvalidValue(t *testing.T) {
	err := validateOutputFormat("table", "list", "json")
	if err == nil {
		t.Fatalf("expected error for unsupported format")
	}
}

func TestRenderNotesByFormat_JSON(t *testing.T) {
	note := services.Note{Content: "hello", Metadata: map[string]any{"title": "Hello"}}
	note.File.Relative = "hello.md"
	note.File.Filepath = "hello.md"

	output := captureStdout(t, func() {
		if err := renderNotesByFormat([]services.Note{note}, "json"); err != nil {
			t.Fatalf("renderNotesByFormat returned error: %v", err)
		}
	})

	var decoded []map[string]any
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("expected valid json output, got error: %v; output=%q", err, output)
	}
	if len(decoded) != 1 {
		t.Fatalf("expected 1 note in output, got %d", len(decoded))
	}
}

func TestNotesGetCommand_Registered(t *testing.T) {
	found := false
	for _, sub := range notesCmd.Commands() {
		if sub.Use == "get <path>" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected notes get command to be registered")
	}
}

func TestNotesSearchCommand_HasFormatFlag(t *testing.T) {
	flag := notesSearchCmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatalf("expected notes search to have --format flag")
	}
	if flag.DefValue != "list" {
		t.Fatalf("expected default format=list, got %q", flag.DefValue)
	}
}

func TestNotebookCommand_HasFormatFlag(t *testing.T) {
	flag := notebookCmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatalf("expected notebook to have --format flag")
	}
	if flag.DefValue != "list" {
		t.Fatalf("expected default format=list, got %q", flag.DefValue)
	}
}

func TestNotebookInfoCommand_Registered(t *testing.T) {
	found := false
	for _, sub := range notebookCmd.Commands() {
		if sub.Use == "info" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected notebook info command to be registered")
	}
}

func TestNotebookInfoCommand_HasFormatFlag(t *testing.T) {
	flag := notebookInfoCmd.Flags().Lookup("format")
	if flag == nil {
		t.Fatalf("expected notebook info to have --format flag")
	}
	if flag.DefValue != "list" {
		t.Fatalf("expected default format=list, got %q", flag.DefValue)
	}
}

func TestNotebookInfoPayload_IncludesConfigPath(t *testing.T) {
	nb := &services.Notebook{}
	nb.Config.Name = "Work"
	nb.Config.Path = "/tmp/work/.jot.json"
	nb.Config.Root = "/tmp/work/notes"

	payload := notebookInfoPayload(nb)

	if payload["config_path"] != "/tmp/work/.jot.json" {
		t.Fatalf("expected config_path in notebook payload")
	}
}

func TestNotebookInfoPayload_IncludesWorkflows(t *testing.T) {
	nb := &services.Notebook{}
	nb.Config.Workflows = map[string]services.WorkflowDefinition{
		"project_story": {InitialState: "proposed"},
	}

	payload := notebookInfoPayload(nb)
	workflows, ok := payload["workflows"].(map[string]services.WorkflowDefinition)
	if !ok {
		t.Fatalf("expected workflows map in payload")
	}
	if _, exists := workflows["project_story"]; !exists {
		t.Fatalf("expected project_story workflow in payload")
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close write pipe failed: %v", err)
	}
	os.Stdout = original

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close read pipe failed: %v", err)
	}

	return buf.String()
}
