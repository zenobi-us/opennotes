package cmd

import (
	"testing"
)

func TestNotesAdd_TypeFlagRegistered(t *testing.T) {
	flag := notesAddCmd.Flags().Lookup("type")
	if flag == nil {
		t.Fatalf("notes add should define --type flag")
	}
	if flag.Shorthand != "T" {
		t.Fatalf("expected --type shorthand to be 'T', got %q", flag.Shorthand)
	}
	if flag.DefValue != "" {
		t.Fatalf("expected --type default value to be empty, got %q", flag.DefValue)
	}
}

func TestNotesAdd_TypeFlagParses(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "long form --type",
			args:     []string{"--type", "task"},
			expected: "task",
		},
		{
			name:     "short form -T",
			args:     []string{"-T", "meeting"},
			expected: "meeting",
		},
		{
			name:     "type with equals sign",
			args:     []string{"--type=project"},
			expected: "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			_ = notesAddCmd.Flags().Set("type", "")

			if err := notesAddCmd.Flags().Parse(tt.args); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			got, err := notesAddCmd.Flags().GetString("type")
			if err != nil {
				t.Fatalf("failed to get type flag: %v", err)
			}

			if got != tt.expected {
				t.Fatalf("expected type %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestNotesAdd_TypeFlagCombinesWithNotebook(t *testing.T) {
	// Reset flags
	_ = notesAddCmd.Flags().Set("type", "")
	_ = notesAddCmd.Flags().Set("template", "")

	// Parse --type together with other flags (--notebook is a persistent root flag, so test with --template)
	args := []string{"--type", "task", "--template", "default"}
	if err := notesAddCmd.Flags().Parse(args); err != nil {
		t.Fatalf("failed to parse combined flags: %v", err)
	}

	noteType, err := notesAddCmd.Flags().GetString("type")
	if err != nil {
		t.Fatalf("failed to get type flag: %v", err)
	}
	if noteType != "task" {
		t.Fatalf("expected type 'task', got %q", noteType)
	}

	template, err := notesAddCmd.Flags().GetString("template")
	if err != nil {
		t.Fatalf("failed to get template flag: %v", err)
	}
	if template != "default" {
		t.Fatalf("expected template 'default', got %q", template)
	}
}

func TestNotesAdd_OtherFlagsStillWork(t *testing.T) {
	// Ensure adding --type didn't break existing flags
	if notesAddCmd.Flags().Lookup("template") == nil {
		t.Fatalf("notes add should define --template flag")
	}
	if notesAddCmd.Flags().Lookup("title") == nil {
		t.Fatalf("notes add should define --title flag")
	}
	if notesAddCmd.Flags().Lookup("data") == nil {
		t.Fatalf("notes add should define --data flag")
	}
}

func TestNotesAdd_NoInteractiveFlagRegistered(t *testing.T) {
	flag := notesAddCmd.Flags().Lookup("no-interactive")
	if flag == nil {
		t.Fatalf("notes add should define --no-interactive flag")
	}
	if flag.DefValue != "false" {
		t.Fatalf("expected --no-interactive default value to be 'false', got %q", flag.DefValue)
	}
}

func TestNotesAdd_NoInteractiveFlagParses(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "flag not set",
			args:     []string{},
			expected: false,
		},
		{
			name:     "flag set",
			args:     []string{"--no-interactive"},
			expected: true,
		},
		{
			name:     "flag set to true explicitly",
			args:     []string{"--no-interactive=true"},
			expected: true,
		},
		{
			name:     "flag set to false explicitly",
			args:     []string{"--no-interactive=false"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each test
			_ = notesAddCmd.Flags().Set("no-interactive", "false")

			if err := notesAddCmd.Flags().Parse(tt.args); err != nil {
				t.Fatalf("failed to parse flags: %v", err)
			}

			got, err := notesAddCmd.Flags().GetBool("no-interactive")
			if err != nil {
				t.Fatalf("failed to get no-interactive flag: %v", err)
			}

			if got != tt.expected {
				t.Fatalf("expected no-interactive=%v, got %v", tt.expected, got)
			}
		})
	}
}
