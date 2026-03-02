package services

import (
	"fmt"
	"os"
	"path/filepath"
)

// TemplateSyntaxError is returned when a template has invalid syntax.
// Wraps the underlying parse/execute error with context about the template.
type TemplateSyntaxError struct {
	Template    string // The template string that failed
	OriginalErr error  // The underlying parse/execute error
	Message     string // Human-readable explanation
}

// Error implements the error interface.
func (e *TemplateSyntaxError) Error() string {
	return fmt.Sprintf("template syntax error: %s", e.Message)
}

// Unwrap allows errors.Unwrap and errors.As to access the original error.
func (e *TemplateSyntaxError) Unwrap() error {
	return e.OriginalErr
}

// FilenameCollisionError is returned when a generated filename already exists.
type FilenameCollisionError struct {
	Path       string
	Filename   string
	Suggestion string
}

// Error implements the error interface.
func (e *FilenameCollisionError) Error() string {
	return fmt.Sprintf("filename collision: %q already exists. %s",
		e.Filename, e.Suggestion)
}

// CheckFilenameCollision checks if a file already exists at the given path.
// Returns a FilenameCollisionError if the file exists, nil otherwise.
func CheckFilenameCollision(fullPath string) error {
	if _, err := os.Stat(fullPath); err == nil {
		return &FilenameCollisionError{
			Path:       fullPath,
			Filename:   filepath.Base(fullPath),
			Suggestion: "Use a different title or add {{ jot.NanoID 8 }} to filename_format for uniqueness",
		}
	}
	return nil
}
