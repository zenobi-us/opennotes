package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateNotebookName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple name",
			input:   "My Notebook",
			wantErr: false,
		},
		{
			name:    "valid with numbers",
			input:   "Notes 2024",
			wantErr: false,
		},
		{
			name:    "valid with hyphen",
			input:   "My-Notes",
			wantErr: false,
		},
		{
			name:    "valid with underscore",
			input:   "My_Notes",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "invalid special chars",
			input:   "My Notes!@#",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotebookName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateNoteName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid simple name",
			input:   "my-note.md",
			wantErr: false,
		},
		{
			name:    "valid without extension",
			input:   "my-note",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "path traversal",
			input:   "../secret.md",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNoteName(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationErrors(t *testing.T) {
	errors := ValidationErrors{
		{Path: "user.name", Message: "is required"},
		{Path: "user.email", Message: "must be valid email"},
	}

	// Test Error()
	errStr := errors.Error()
	assert.Contains(t, errStr, "user.name")
	assert.Contains(t, errStr, "user.email")

	// Test PrettyPrint()
	pretty := errors.PrettyPrint()
	assert.Contains(t, pretty, "- user.name")
	assert.Contains(t, pretty, "  - is required")
}

func TestValidator(t *testing.T) {
	v := NewValidator()

	assert.False(t, v.HasErrors())

	v.AddError("root error")
	assert.True(t, v.HasErrors())

	nested := v.WithPath("field")
	nested.AddError("field error")

	errors := v.Errors()
	assert.Len(t, errors, 2)
}
