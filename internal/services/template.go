package services

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

// TemplateEngine provides template rendering with custom functions.
// Uses Go's text/template with registered helper functions.
type TemplateEngine struct {
	funcs template.FuncMap
}

// NewTemplateEngine creates a new TemplateEngine with built-in functions.
// Includes the slug function for filename-safe string conversion.
func NewTemplateEngine() *TemplateEngine {
	te := &TemplateEngine{
		funcs: template.FuncMap{},
	}

	// Register built-in functions
	te.RegisterFunc("slug", Slug)
	te.RegisterFunc("slugmax", SlugWithMax)
	te.RegisterFunc("lower", strings.ToLower)
	te.RegisterFunc("upper", strings.ToUpper)
	te.RegisterFunc("trim", strings.TrimSpace)
	te.RegisterFunc("replace", strings.ReplaceAll)

	// Register jot namespace for {{ jot.Slug "text" }} syntax
	te.RegisterFunc("jot", JotFuncs)

	return te
}

// RegisterFunc registers a custom function for use in templates.
// Functions can be called using {{ .value | funcName }} syntax.
func (te *TemplateEngine) RegisterFunc(name string, fn interface{}) {
	te.funcs[name] = fn
}

// Validate checks if a template string has valid syntax.
// Returns a TemplateSyntaxError if parsing fails, nil if valid.
func (te *TemplateEngine) Validate(tmpl string) error {
	_, err := template.New("validate").Funcs(te.funcs).Parse(tmpl)
	if err != nil {
		return &TemplateSyntaxError{
			Template:    tmpl,
			OriginalErr: err,
			Message:     fmt.Sprintf("failed to parse template: %v", err),
		}
	}
	return nil
}

// Render processes a template string with the provided data.
// Returns the rendered string or a TemplateSyntaxError if parsing/execution fails.
func (te *TemplateEngine) Render(tmpl string, data map[string]interface{}) (string, error) {
	t, err := template.New("template").Funcs(te.funcs).Parse(tmpl)
	if err != nil {
		return "", &TemplateSyntaxError{
			Template:    tmpl,
			OriginalErr: err,
			Message:     fmt.Sprintf("failed to parse template: %v", err),
		}
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", &TemplateSyntaxError{
			Template:    tmpl,
			OriginalErr: err,
			Message:     fmt.Sprintf("failed to execute template: %v", err),
		}
	}

	return buf.String(), nil
}

// GenerateFilename renders a filename format template with the given title.
// The title is available as .title in the template.
// Example format: "{{ .title | slug }}.md" -> "my-title.md"
func GenerateFilename(format string, title string) (string, error) {
	te := NewTemplateEngine()
	data := map[string]interface{}{
		"title": title,
	}
	return te.Render(format, data)
}

// GenerateContent renders a content template with the provided data.
// Template data should include: title, filename, group (optional).
// All jot functions (jot.Now, jot.Slug, etc.) are available in the template.
// Example template:
//
//	---
//	title: {{ .title }}
//	created_at: {{ jot.Now "2006-01-02T15:04:05Z07:00" }}
//	---
//
//	# {{ .title }}
func GenerateContent(tmpl string, data map[string]interface{}) (string, error) {
	te := NewTemplateEngine()
	return te.Render(tmpl, data)
}

// ValidateGroupTemplates validates all template strings in a slice of groups.
// Returns a slice of errors, one for each invalid template.
// Empty templates are valid (they will use defaults at runtime).
func ValidateGroupTemplates(groups []NotebookGroup) []error {
	te := NewTemplateEngine()
	var errs []error

	for _, group := range groups {
		// Validate content template if specified
		if group.Template != "" {
			if err := te.Validate(group.Template); err != nil {
				errs = append(errs, fmt.Errorf("group %q content template: %w", group.Name, err))
			}
		}

		// Validate filename format if specified
		if group.FilenameFormat != "" {
			if err := te.Validate(group.FilenameFormat); err != nil {
				errs = append(errs, fmt.Errorf("group %q filename format: %w", group.Name, err))
			}
		}
	}

	return errs
}
