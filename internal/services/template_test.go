package services

import (
	"errors"
	"strings"
	"testing"
)

func TestTemplateEngine_Render(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:     "Basic variable substitution",
			template: "{{ .title }}",
			data:     map[string]interface{}{"title": "my title"},
			expected: "my title",
			wantErr:  false,
		},
		{
			name:     "Slug pipe function",
			template: "{{ .title | slug }}",
			data:     map[string]interface{}{"title": "my title"},
			expected: "my-title",
			wantErr:  false,
		},
		{
			name:     "Default filename format",
			template: "{{ .title | slug }}.md",
			data:     map[string]interface{}{"title": "my title"},
			expected: "my-title.md",
			wantErr:  false,
		},
		{
			name:     "Multiple variables",
			template: "{{ .prefix }}-{{ .title | slug }}",
			data:     map[string]interface{}{"prefix": "note", "title": "Hello World"},
			expected: "note-hello-world",
			wantErr:  false,
		},
		{
			name:     "Lower function",
			template: "{{ .title | lower }}",
			data:     map[string]interface{}{"title": "HELLO"},
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "Upper function",
			template: "{{ .title | upper }}",
			data:     map[string]interface{}{"title": "hello"},
			expected: "HELLO",
			wantErr:  false,
		},
		{
			name:     "Trim function",
			template: "{{ .title | trim }}",
			data:     map[string]interface{}{"title": "  hello  "},
			expected: "hello",
			wantErr:  false,
		},
		{
			name:     "Chained functions",
			template: "{{ .title | trim | slug }}",
			data:     map[string]interface{}{"title": "  Hello World  "},
			expected: "hello-world",
			wantErr:  false,
		},
		{
			name:     "Empty template",
			template: "",
			data:     map[string]interface{}{"title": "test"},
			expected: "",
			wantErr:  false,
		},
		{
			name:     "Static text only",
			template: "static-filename.md",
			data:     map[string]interface{}{},
			expected: "static-filename.md",
			wantErr:  false,
		},
		{
			name:     "Invalid template syntax - unclosed brace",
			template: "{{ .title",
			data:     map[string]interface{}{"title": "test"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Invalid template syntax - unknown function",
			template: "{{ .title | unknownfunc }}",
			data:     map[string]interface{}{"title": "test"},
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Missing variable returns empty",
			template: "prefix-{{ .missing }}-suffix",
			data:     map[string]interface{}{"title": "test"},
			expected: "prefix-<no value>-suffix",
			wantErr:  false,
		},
		{
			name:     "Unicode in title with slug",
			template: "{{ .title | slug }}.md",
			data:     map[string]interface{}{"title": "会议 Meeting"},
			expected: "hui-yi-meeting.md",
			wantErr:  false,
		},
		{
			name:     "Emoji handling with slug",
			template: "{{ .title | slug }}.md",
			data:     map[string]interface{}{"title": "🎉 Party Notes"},
			expected: "party-notes.md",
			wantErr:  false,
		},
	}

	te := NewTemplateEngine()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := te.Render(tt.template, tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("Render(%q) = %q, want %q", tt.template, result, tt.expected)
			}
		})
	}
}

func TestTemplateEngine_RegisterFunc(t *testing.T) {
	te := NewTemplateEngine()

	// Register a custom function
	te.RegisterFunc("double", func(s string) string {
		return s + s
	})

	result, err := te.Render("{{ .value | double }}", map[string]interface{}{"value": "abc"})
	if err != nil {
		t.Errorf("Render() unexpected error: %v", err)
	}

	expected := "abcabc"
	if result != expected {
		t.Errorf("Render() = %q, want %q", result, expected)
	}
}

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		title    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Simple slug format",
			format:   "{{ .title | slug }}.md",
			title:    "My Note Title",
			expected: "my-note-title.md",
			wantErr:  false,
		},
		{
			name:     "Plain title without slug",
			format:   "{{ .title }}.txt",
			title:    "My Note",
			expected: "My Note.txt",
			wantErr:  false,
		},
		{
			name:     "With prefix",
			format:   "note-{{ .title | slug }}.md",
			title:    "Hello World",
			expected: "note-hello-world.md",
			wantErr:  false,
		},
		{
			name:     "Empty title becomes untitled",
			format:   "{{ .title | slug }}.md",
			title:    "",
			expected: "untitled.md",
			wantErr:  false,
		},
		{
			name:     "Special characters in title",
			format:   "{{ .title | slug }}.md",
			title:    "Hello & World!",
			expected: "hello-and-world.md",
			wantErr:  false,
		},
		{
			name:     "Invalid template format",
			format:   "{{ .title | slug",
			title:    "Test",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "Complex format with multiple operations",
			format:   "docs/{{ .title | slug }}/index.md",
			title:    "Getting Started",
			expected: "docs/getting-started/index.md",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateFilename(tt.format, tt.title)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.expected {
				t.Errorf("GenerateFilename(%q, %q) = %q, want %q",
					tt.format, tt.title, result, tt.expected)
			}
		})
	}
}

func TestNewTemplateEngine_BuiltinFunctions(t *testing.T) {
	te := NewTemplateEngine()

	// Verify all expected built-in functions are registered
	builtins := []string{"slug", "slugmax", "lower", "upper", "trim", "replace"}

	for _, fn := range builtins {
		if te.funcs[fn] == nil {
			t.Errorf("Expected built-in function %q to be registered", fn)
		}
	}
}

func TestTemplateEngine_SlugMaxFunction(t *testing.T) {
	te := NewTemplateEngine()

	// Test slugmax with max length parameter
	result, err := te.Render(`{{ slugmax .title 10 }}`, map[string]interface{}{
		"title": "This is a very long title",
	})

	if err != nil {
		t.Errorf("Render() unexpected error: %v", err)
	}

	if len(result) > 10 {
		t.Errorf("slugmax result length %d exceeds max 10: %q", len(result), result)
	}
}

func TestTemplateEngine_ReplaceFunction(t *testing.T) {
	te := NewTemplateEngine()

	result, err := te.Render(`{{ replace .title " " "_" }}`, map[string]interface{}{
		"title": "hello world test",
	})

	if err != nil {
		t.Errorf("Render() unexpected error: %v", err)
	}

	expected := "hello_world_test"
	if result != expected {
		t.Errorf("Render() = %q, want %q", result, expected)
	}
}

func TestGenerateContent(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		contains []string
		wantErr  bool
	}{
		{
			name: "Basic template renders correctly",
			template: `---
title: {{ .title }}
---

# {{ .title }}
`,
			data: map[string]interface{}{
				"title": "My Test Note",
			},
			contains: []string{
				"title: My Test Note",
				"# My Test Note",
			},
			wantErr: false,
		},
		{
			name: "Template with filename and group",
			template: `---
title: {{ .title }}
filename: {{ .filename }}
group: {{ .group }}
---
`,
			data: map[string]interface{}{
				"title":    "Test",
				"filename": "test.md",
				"group":    "tasks",
			},
			contains: []string{
				"title: Test",
				"filename: test.md",
				"group: tasks",
			},
			wantErr: false,
		},
		{
			name: "Template with jot.Now function",
			template: `---
title: {{ .title }}
created_at: {{ jot.Now "2006-01-02" }}
---
`,
			data: map[string]interface{}{
				"title": "Test",
			},
			contains: []string{
				"title: Test",
				"created_at: 20", // Partial check for date format
			},
			wantErr: false,
		},
		{
			name: "Template with jot.Slug function",
			template: `---
slug: {{ jot.Slug .title }}
---
`,
			data: map[string]interface{}{
				"title": "Hello World",
			},
			contains: []string{
				"slug: hello-world",
			},
			wantErr: false,
		},
		{
			name: "Template with pipe functions",
			template: `{{ .title | lower }}`,
			data: map[string]interface{}{
				"title": "UPPERCASE TITLE",
			},
			contains: []string{
				"uppercase title",
			},
			wantErr: false,
		},
		{
			name:     "Invalid template syntax",
			template: `{{ .title`,
			data:     map[string]interface{}{"title": "Test"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateContent(tt.template, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateContent() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GenerateContent() unexpected error: %v", err)
				return
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("GenerateContent() result missing expected content %q, got:\n%s", expected, result)
				}
			}
		})
	}
}

func TestGenerateContent_DefaultContentTemplate(t *testing.T) {
	// Test that DefaultContentTemplate renders correctly
	data := map[string]interface{}{
		"title": "My Note Title",
	}

	result, err := GenerateContent(DefaultContentTemplate, data)
	if err != nil {
		t.Fatalf("GenerateContent() with DefaultContentTemplate failed: %v", err)
	}

	// Check YAML frontmatter is present
	if !strings.HasPrefix(result, "---\n") {
		t.Errorf("DefaultContentTemplate result should start with YAML frontmatter, got: %s", result)
	}

	// Check title is in frontmatter
	if !strings.Contains(result, "title: My Note Title") {
		t.Errorf("DefaultContentTemplate result should contain title in frontmatter, got: %s", result)
	}

	// Check created_at is present (with ISO 8601 format)
	if !strings.Contains(result, "created_at: 20") {
		t.Errorf("DefaultContentTemplate result should contain created_at timestamp, got: %s", result)
	}

	// Check heading is present
	if !strings.Contains(result, "# My Note Title") {
		t.Errorf("DefaultContentTemplate result should contain title heading, got: %s", result)
	}
}

func TestTemplateEngine_Validate(t *testing.T) {
	te := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "Valid simple template",
			template: "{{ .title }}",
			wantErr:  false,
		},
		{
			name:     "Valid template with pipe",
			template: "{{ .title | slug }}.md",
			wantErr:  false,
		},
		{
			name:     "Valid empty template",
			template: "",
			wantErr:  false,
		},
		{
			name:     "Valid static text",
			template: "static-text.md",
			wantErr:  false,
		},
		{
			name:     "Invalid unclosed brace",
			template: "{{ .title",
			wantErr:  true,
		},
		{
			name:     "Invalid unknown function",
			template: "{{ .title | nonexistent }}",
			wantErr:  true,
		},
		{
			name:     "Invalid malformed template",
			template: "{{ {{ }}",
			wantErr:  true,
		},
		{
			name:     "Valid complex template",
			template: DefaultContentTemplate,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := te.Validate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateEngine_Validate_ReturnsTemplateSyntaxError(t *testing.T) {
	te := NewTemplateEngine()

	err := te.Validate("{{ .title | nonexistent }}")
	if err == nil {
		t.Fatal("Validate() expected error for invalid template")
	}

	var syntaxErr *TemplateSyntaxError
	if !errors.As(err, &syntaxErr) {
		t.Errorf("Validate() error should be TemplateSyntaxError, got %T", err)
	}

	// Check error contains useful information
	if syntaxErr.Template != "{{ .title | nonexistent }}" {
		t.Errorf("TemplateSyntaxError.Template = %q, want %q",
			syntaxErr.Template, "{{ .title | nonexistent }}")
	}

	if syntaxErr.OriginalErr == nil {
		t.Error("TemplateSyntaxError.OriginalErr should not be nil")
	}

	if syntaxErr.Message == "" {
		t.Error("TemplateSyntaxError.Message should not be empty")
	}

	// Check error message is human-readable
	errorMsg := syntaxErr.Error()
	if !strings.Contains(errorMsg, "template syntax error") {
		t.Errorf("Error message should contain 'template syntax error', got: %s", errorMsg)
	}
}

func TestTemplateEngine_Render_WrapsErrors(t *testing.T) {
	te := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
	}{
		{
			name:     "Parse error - unclosed brace",
			template: "{{ .title",
			data:     map[string]interface{}{"title": "test"},
		},
		{
			name:     "Parse error - unknown function",
			template: "{{ .title | unknownfunc }}",
			data:     map[string]interface{}{"title": "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := te.Render(tt.template, tt.data)
			if err == nil {
				t.Fatal("Render() expected error")
			}

			var syntaxErr *TemplateSyntaxError
			if !errors.As(err, &syntaxErr) {
				t.Errorf("Render() error should be TemplateSyntaxError, got %T", err)
			}

			// Verify we can unwrap to get the original error
			if syntaxErr.Unwrap() == nil {
				t.Error("TemplateSyntaxError.Unwrap() should return original error")
			}
		})
	}
}

func TestValidateGroupTemplates(t *testing.T) {
	tests := []struct {
		name       string
		groups     []NotebookGroup
		wantErrors int
	}{
		{
			name: "All valid templates",
			groups: []NotebookGroup{
				{
					Name:           "Tasks",
					Template:       "# {{ .title }}",
					FilenameFormat: "{{ .title | slug }}.md",
				},
				{
					Name:           "Notes",
					Template:       "---\ntitle: {{ .title }}\n---",
					FilenameFormat: "{{ .title | lower }}.md",
				},
			},
			wantErrors: 0,
		},
		{
			name: "Empty templates are valid",
			groups: []NotebookGroup{
				{
					Name:           "Default",
					Template:       "", // Uses default
					FilenameFormat: "", // Uses default
				},
			},
			wantErrors: 0,
		},
		{
			name: "Invalid content template",
			groups: []NotebookGroup{
				{
					Name:           "Bad",
					Template:       "{{ .title | nonexistent }}",
					FilenameFormat: "{{ .title | slug }}.md",
				},
			},
			wantErrors: 1,
		},
		{
			name: "Invalid filename format",
			groups: []NotebookGroup{
				{
					Name:           "Bad",
					Template:       "# {{ .title }}",
					FilenameFormat: "{{ .title",
				},
			},
			wantErrors: 1,
		},
		{
			name: "Both templates invalid",
			groups: []NotebookGroup{
				{
					Name:           "Bad",
					Template:       "{{ {{",
					FilenameFormat: "{{ .title | nonexistent }}",
				},
			},
			wantErrors: 2,
		},
		{
			name: "Multiple groups with mixed validity",
			groups: []NotebookGroup{
				{
					Name:           "Good",
					Template:       "# {{ .title }}",
					FilenameFormat: "{{ .title | slug }}.md",
				},
				{
					Name:           "Bad",
					Template:       "{{ .title | nonexistent }}",
					FilenameFormat: "{{ .title | slug }}.md",
				},
			},
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateGroupTemplates(tt.groups)
			if len(errs) != tt.wantErrors {
				t.Errorf("ValidateGroupTemplates() returned %d errors, want %d", len(errs), tt.wantErrors)
				for _, err := range errs {
					t.Logf("  error: %v", err)
				}
			}
		})
	}
}

func TestValidateGroupTemplates_ErrorContainsGroupName(t *testing.T) {
	groups := []NotebookGroup{
		{
			Name:     "MyGroup",
			Template: "{{ .title | nonexistent }}",
		},
	}

	errs := ValidateGroupTemplates(groups)
	if len(errs) != 1 {
		t.Fatalf("ValidateGroupTemplates() returned %d errors, want 1", len(errs))
	}

	errorMsg := errs[0].Error()
	if !strings.Contains(errorMsg, "MyGroup") {
		t.Errorf("Error message should contain group name 'MyGroup', got: %s", errorMsg)
	}

	// Should also contain TemplateSyntaxError in the chain
	var syntaxErr *TemplateSyntaxError
	if !errors.As(errs[0], &syntaxErr) {
		t.Error("Error should wrap TemplateSyntaxError")
	}
}

func TestNotebookGroup_GetTemplate_Fallback(t *testing.T) {
	tests := []struct {
		name     string
		group    NotebookGroup
		expected string
	}{
		{
			name:     "Empty template uses default",
			group:    NotebookGroup{Name: "Test", Template: ""},
			expected: DefaultContentTemplate,
		},
		{
			name:     "Custom template used when specified",
			group:    NotebookGroup{Name: "Test", Template: "# {{ .title }}"},
			expected: "# {{ .title }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.group.GetTemplate()
			if result != tt.expected {
				t.Errorf("GetTemplate() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestNotebookGroup_GetFilenameFormat_Fallback(t *testing.T) {
	tests := []struct {
		name     string
		group    NotebookGroup
		expected string
	}{
		{
			name:     "Empty format uses default",
			group:    NotebookGroup{Name: "Test", FilenameFormat: ""},
			expected: DefaultFilenameFormat,
		},
		{
			name:     "Custom format used when specified",
			group:    NotebookGroup{Name: "Test", FilenameFormat: "custom-{{ .title | slug }}.md"},
			expected: "custom-{{ .title | slug }}.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.group.GetFilenameFormat()
			if result != tt.expected {
				t.Errorf("GetFilenameFormat() = %q, want %q", result, tt.expected)
			}
		})
	}
}
