package core

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a validation error with path context.
type ValidationError struct {
	Path    string
	Message string
}

func (e ValidationError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}

	var lines []string
	for _, err := range e {
		lines = append(lines, "- "+err.Error())
	}
	return strings.Join(lines, "\n")
}

// PrettyPrint formats validation errors in a hierarchical format grouped by path.
func (e ValidationErrors) PrettyPrint() string {
	grouped := make(map[string][]string)

	for _, err := range e {
		path := err.Path
		if path == "" {
			path = "(root)"
		}
		grouped[path] = append(grouped[path], err.Message)
	}

	var lines []string
	for path, messages := range grouped {
		lines = append(lines, fmt.Sprintf("- %s", path))
		for _, msg := range messages {
			lines = append(lines, fmt.Sprintf("  - %s", msg))
		}
	}

	return strings.Join(lines, "\n")
}

// Validator is a helper to validate a value and collect errors.
type Validator struct {
	errors *ValidationErrors
	path   string
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{
		errors: &ValidationErrors{},
	}
}

// WithPath creates a nested validator with a path prefix.
func (v *Validator) WithPath(path string) *Validator {
	newPath := path
	if v.path != "" {
		newPath = v.path + "." + path
	}
	return &Validator{
		errors: v.errors, // Share the same errors slice
		path:   newPath,
	}
}

// AddError adds a validation error.
func (v *Validator) AddError(message string) {
	*v.errors = append(*v.errors, ValidationError{
		Path:    v.path,
		Message: message,
	})
}

// Errors returns all collected validation errors.
func (v *Validator) Errors() ValidationErrors {
	return *v.errors
}

// HasErrors returns true if there are validation errors.
func (v *Validator) HasErrors() bool {
	return len(*v.errors) > 0
}

// Common validation functions

// ValidateNotebookName validates a notebook name.
func ValidateNotebookName(name string) error {
	if name == "" {
		return fmt.Errorf("notebook name is required")
	}

	if len(name) < 1 || len(name) > 100 {
		return fmt.Errorf("notebook name must be between 1 and 100 characters")
	}

	// Allow alphanumeric, spaces, hyphens, underscores
	valid := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	if !valid.MatchString(name) {
		return fmt.Errorf("notebook name can only contain letters, numbers, spaces, hyphens, and underscores")
	}

	return nil
}

// ValidatePath validates a filesystem path.
func ValidatePath(path string) error {
	if path == "" {
		return nil // Empty path is allowed (uses default)
	}

	// Check for invalid characters
	invalid := regexp.MustCompile(`[\x00-\x1f]`)
	if invalid.MatchString(path) {
		return fmt.Errorf("path contains invalid characters")
	}

	return nil
}

// ValidateNoteName validates a note filename.
func ValidateNoteName(name string) error {
	if name == "" {
		return fmt.Errorf("note name is required")
	}

	// Remove .md extension for validation
	name = strings.TrimSuffix(name, ".md")

	if len(name) > 255 {
		return fmt.Errorf("note name is too long (max 255 characters)")
	}

	// Check for path traversal
	if strings.Contains(name, "..") {
		return fmt.Errorf("note name cannot contain path traversal (..)")
	}

	return nil
}
