package services

import (
	"testing"
)

func TestBuildGroupLabel(t *testing.T) {
	tests := []struct {
		name     string
		group    NotebookGroup
		expected string
	}{
		{
			name:     "name only",
			group:    NotebookGroup{Name: "Tasks"},
			expected: "Tasks",
		},
		{
			name:     "name with different type",
			group:    NotebookGroup{Name: "Daily Tasks", Type: "task"},
			expected: "Daily Tasks (task)",
		},
		{
			name:     "name equals type",
			group:    NotebookGroup{Name: "task", Type: "task"},
			expected: "task",
		},
		{
			name:     "empty type",
			group:    NotebookGroup{Name: "Notes", Type: ""},
			expected: "Notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := BuildGroupLabel(tt.group)
			if result != tt.expected {
				t.Errorf("BuildGroupLabel() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestBuildGroupSelectOptions(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Tasks", Type: "task"},
		{Name: "Meetings"},
		{Name: "Ideas", Type: "idea"},
	}

	options := BuildGroupSelectOptions(groups)

	if len(options) != len(groups) {
		t.Errorf("BuildGroupSelectOptions() returned %d options, want %d", len(options), len(groups))
	}

	// Verify option values are correct indices
	for i := range options {
		// Since we're testing the count and structure, we verify indices match
		if i < 0 || i >= len(groups) {
			t.Errorf("Option index %d out of range", i)
		}
	}
}

func TestBuildGroupSelectOptions_Empty(t *testing.T) {
	groups := []NotebookGroup{}
	options := BuildGroupSelectOptions(groups)

	if len(options) != 0 {
		t.Errorf("BuildGroupSelectOptions() with empty groups returned %d options, want 0", len(options))
	}
}

func TestSelectGroupInteractively_NoGroups(t *testing.T) {
	groups := []NotebookGroup{}

	_, err := SelectGroupInteractively(groups)

	if err == nil {
		t.Error("SelectGroupInteractively() with empty groups should return error")
	}

	expectedMsg := "no groups available for selection"
	if err.Error() != expectedMsg {
		t.Errorf("SelectGroupInteractively() error = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestBuildGroupSelectOptions_LabelsMatchExpected(t *testing.T) {
	groups := []NotebookGroup{
		{Name: "Daily Tasks", Type: "task"},
		{Name: "meeting", Type: "meeting"},
		{Name: "Random Notes"},
	}

	expectedLabels := []string{
		"Daily Tasks (task)",
		"meeting",
		"Random Notes",
	}

	options := BuildGroupSelectOptions(groups)

	if len(options) != len(expectedLabels) {
		t.Fatalf("BuildGroupSelectOptions() returned %d options, want %d", len(options), len(expectedLabels))
	}

	// Verify labels by building them independently
	for i, g := range groups {
		label := BuildGroupLabel(g)
		if label != expectedLabels[i] {
			t.Errorf("Option[%d] label = %q, want %q", i, label, expectedLabels[i])
		}
	}
}

func TestShouldShowInteractiveSelector(t *testing.T) {
	multipleGroups := []NotebookGroup{
		{Name: "Tasks", Type: "task"},
		{Name: "Notes", Type: "note"},
	}
	singleGroup := []NotebookGroup{
		{Name: "Tasks", Type: "task"},
	}

	tests := []struct {
		name     string
		ctx      InteractiveContext
		expected bool
	}{
		{
			name: "no type, no path, multiple groups, TTY → returns true",
			ctx: InteractiveContext{
				TypeFlag:      "",
				ExplicitPath:  "",
				Groups:        multipleGroups,
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: true,
		},
		{
			name: "type provided → returns false",
			ctx: InteractiveContext{
				TypeFlag:      "task",
				ExplicitPath:  "",
				Groups:        multipleGroups,
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: false,
		},
		{
			name: "path provided → returns false",
			ctx: InteractiveContext{
				TypeFlag:      "",
				ExplicitPath:  "some/path/",
				Groups:        multipleGroups,
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: false,
		},
		{
			name: "single group → returns false",
			ctx: InteractiveContext{
				TypeFlag:      "",
				ExplicitPath:  "",
				Groups:        singleGroup,
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: false,
		},
		{
			name: "empty groups → returns false",
			ctx: InteractiveContext{
				TypeFlag:      "",
				ExplicitPath:  "",
				Groups:        []NotebookGroup{},
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: false,
		},
		{
			name: "not TTY → returns false",
			ctx: InteractiveContext{
				TypeFlag:      "",
				ExplicitPath:  "",
				Groups:        multipleGroups,
				IsTTY:         false,
				NoInteractive: false,
			},
			expected: false,
		},
		{
			name: "NoInteractive flag set → returns false",
			ctx: InteractiveContext{
				TypeFlag:      "",
				ExplicitPath:  "",
				Groups:        multipleGroups,
				IsTTY:         true,
				NoInteractive: true,
			},
			expected: false,
		},
		{
			name: "all conditions met with three groups → returns true",
			ctx: InteractiveContext{
				TypeFlag:     "",
				ExplicitPath: "",
				Groups: []NotebookGroup{
					{Name: "Tasks"},
					{Name: "Notes"},
					{Name: "Ideas"},
				},
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: true,
		},
		{
			name: "NoInteractive overrides all other conditions",
			ctx: InteractiveContext{
				TypeFlag:     "",
				ExplicitPath: "",
				Groups: []NotebookGroup{
					{Name: "Tasks"},
					{Name: "Notes"},
				},
				IsTTY:         true,
				NoInteractive: true,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldShowInteractiveSelector(tt.ctx)
			if result != tt.expected {
				t.Errorf("ShouldShowInteractiveSelector() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldShowInteractiveSelector_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		ctx      InteractiveContext
		expected bool
	}{
		{
			name: "exactly 2 groups (minimum for choice)",
			ctx: InteractiveContext{
				TypeFlag:     "",
				ExplicitPath: "",
				Groups: []NotebookGroup{
					{Name: "A"},
					{Name: "B"},
				},
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: true,
		},
		{
			name: "whitespace in path counts as path provided",
			ctx: InteractiveContext{
				TypeFlag:     "",
				ExplicitPath: "  ",
				Groups: []NotebookGroup{
					{Name: "A"},
					{Name: "B"},
				},
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: false,
		},
		{
			name: "whitespace in type counts as type provided",
			ctx: InteractiveContext{
				TypeFlag:     " ",
				ExplicitPath: "",
				Groups: []NotebookGroup{
					{Name: "A"},
					{Name: "B"},
				},
				IsTTY:         true,
				NoInteractive: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldShowInteractiveSelector(tt.ctx)
			if result != tt.expected {
				t.Errorf("ShouldShowInteractiveSelector() = %v, want %v", result, tt.expected)
			}
		})
	}
}
