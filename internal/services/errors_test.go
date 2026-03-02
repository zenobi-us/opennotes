package services_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/services"
)

func TestFilenameCollisionError_Error(t *testing.T) {
	err := &services.FilenameCollisionError{
		Path:       "/path/to/note.md",
		Filename:   "note.md",
		Suggestion: "Use a different title",
	}

	msg := err.Error()
	assert.Contains(t, msg, "filename collision")
	assert.Contains(t, msg, "note.md")
	assert.Contains(t, msg, "Use a different title")
}

func TestFilenameCollisionError_TypeAssertion(t *testing.T) {
	var err error = &services.FilenameCollisionError{
		Path:       "/path/to/note.md",
		Filename:   "note.md",
		Suggestion: "Use a different title",
	}

	var collisionErr *services.FilenameCollisionError
	assert.True(t, errors.As(err, &collisionErr))
	assert.Equal(t, "/path/to/note.md", collisionErr.Path)
	assert.Equal(t, "note.md", collisionErr.Filename)
}

func TestCheckFilenameCollision_FileExists(t *testing.T) {
	// Create a temporary directory and file
	tmpDir := t.TempDir()
	existingFile := filepath.Join(tmpDir, "existing-note.md")
	require.NoError(t, os.WriteFile(existingFile, []byte("# Existing Note"), 0644))

	// Check should return FilenameCollisionError
	err := services.CheckFilenameCollision(existingFile)
	require.Error(t, err)

	var collisionErr *services.FilenameCollisionError
	require.True(t, errors.As(err, &collisionErr))
	assert.Equal(t, existingFile, collisionErr.Path)
	assert.Equal(t, "existing-note.md", collisionErr.Filename)
	assert.Contains(t, collisionErr.Suggestion, "jot.NanoID")
}

func TestCheckFilenameCollision_FileDoesNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	nonExistingFile := filepath.Join(tmpDir, "non-existing-note.md")

	// Check should return nil
	err := services.CheckFilenameCollision(nonExistingFile)
	assert.NoError(t, err)
}

func TestCheckFilenameCollision_DirectoryExists(t *testing.T) {
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "notes")
	require.NoError(t, os.MkdirAll(existingDir, 0755))

	// Check should return FilenameCollisionError for directories too
	err := services.CheckFilenameCollision(existingDir)
	require.Error(t, err)

	var collisionErr *services.FilenameCollisionError
	require.True(t, errors.As(err, &collisionErr))
}

func TestFilenameCollisionError_SuggestionIncludes_NanoID(t *testing.T) {
	err := services.CheckFilenameCollision(t.TempDir())

	// TempDir exists, so we should get an error
	var collisionErr *services.FilenameCollisionError
	require.True(t, errors.As(err, &collisionErr))
	assert.Contains(t, collisionErr.Suggestion, "NanoID")
}

func TestTemplateSyntaxError_Error(t *testing.T) {
	err := &services.TemplateSyntaxError{
		Template:    "{{ .title | bad }}",
		OriginalErr: errors.New("function bad not defined"),
		Message:     "failed to parse template: function bad not defined",
	}

	msg := err.Error()
	assert.Contains(t, msg, "template syntax error")
	assert.Contains(t, msg, "failed to parse template")
}

func TestTemplateSyntaxError_TypeAssertion(t *testing.T) {
	originalErr := errors.New("underlying error")
	var err error = &services.TemplateSyntaxError{
		Template:    "{{ .title",
		OriginalErr: originalErr,
		Message:     "failed to parse template",
	}

	var syntaxErr *services.TemplateSyntaxError
	assert.True(t, errors.As(err, &syntaxErr))
	assert.Equal(t, "{{ .title", syntaxErr.Template)
	assert.Equal(t, originalErr, syntaxErr.OriginalErr)
}

func TestTemplateSyntaxError_Unwrap(t *testing.T) {
	originalErr := errors.New("underlying error")
	err := &services.TemplateSyntaxError{
		Template:    "{{ bad",
		OriginalErr: originalErr,
		Message:     "parse error",
	}

	unwrapped := err.Unwrap()
	assert.Equal(t, originalErr, unwrapped)

	// Test that errors.Unwrap works
	assert.Equal(t, originalErr, errors.Unwrap(err))
}

func TestTemplateSyntaxError_ContainsUsefulInfo(t *testing.T) {
	err := &services.TemplateSyntaxError{
		Template:    "{{ .title | nonexistent }}",
		OriginalErr: errors.New("function nonexistent not defined"),
		Message:     "failed to parse template: function nonexistent not defined",
	}

	// Template should be preserved for debugging
	assert.Equal(t, "{{ .title | nonexistent }}", err.Template)

	// Original error should be accessible
	assert.NotNil(t, err.OriginalErr)
	assert.Contains(t, err.OriginalErr.Error(), "nonexistent")

	// Message should be human-readable
	assert.Contains(t, err.Message, "failed to parse")
}
