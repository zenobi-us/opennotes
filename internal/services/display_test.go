package services

import (
	"testing"
)

func TestNewDisplay(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	if display == nil {
		t.Fatal("NewDisplay() returned nil display")
	}

	if display.renderer == nil {
		t.Fatal("NewDisplay() returned display with nil renderer")
	}
}

func TestDisplay_Render_BasicMarkdown(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "heading",
			input:    "# Hello World",
			contains: "Hello World",
		},
		{
			name:     "bullet list",
			input:    "- Item 1\n- Item 2",
			contains: "Item 1",
		},
		{
			name:     "bold text",
			input:    "**bold text**",
			contains: "bold",
		},
		{
			name:     "plain text",
			input:    "Just plain text",
			contains: "Just plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := display.Render(tt.input)
			if err != nil {
				t.Fatalf("Render() failed: %v", err)
			}

			if result == "" {
				t.Error("Render() returned empty string")
			}

			// Check that the result contains expected text
			if tt.contains != "" && !containsString(result, tt.contains) {
				t.Errorf("Render() result = %q, want to contain %q", result, tt.contains)
			}
		})
	}
}

func TestDisplay_Render_EmptyString(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	result, err := display.Render("")
	if err != nil {
		t.Fatalf("Render() failed: %v", err)
	}

	// Empty input should produce some output (glamour may add whitespace)
	// We just check it doesn't error
	_ = result
}

func TestDisplay_RenderTemplate_ValidTemplate(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tmpl := "# {{ .Title }}\n\nWelcome, {{ .Name }}!"
	ctx := map[string]string{
		"Title": "Greeting",
		"Name":  "User",
	}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() failed: %v", err)
	}

	if !containsString(result, "Greeting") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "Greeting")
	}

	if !containsString(result, "User") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "User")
	}
}

func TestDisplay_RenderTemplate_WithStruct(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	type Data struct {
		Title string
		Count int
	}

	tmpl := "# {{ .Title }}\n\nItems: {{ .Count }}"
	ctx := Data{Title: "My List", Count: 42}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() failed: %v", err)
	}

	if !containsString(result, "My List") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "My List")
	}

	if !containsString(result, "42") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "42")
	}
}

func TestDisplay_RenderTemplate_InvalidTemplate_Fallback(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Invalid template syntax (unclosed braces)
	tmpl := "# {{ .Title"
	ctx := map[string]string{"Title": "Test"}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() should not fail on invalid template: %v", err)
	}

	// Should fallback to returning template as-is
	if result != tmpl {
		t.Errorf("RenderTemplate() result = %q, want fallback %q", result, tmpl)
	}
}

func TestDisplay_RenderTemplate_ExecutionError_Fallback(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Valid template but missing field will cause execution error
	tmpl := "# {{ .MissingField }}"
	ctx := map[string]string{"Title": "Test"}

	result, err := display.RenderTemplate(tmpl, ctx)
	if err != nil {
		t.Fatalf("RenderTemplate() should not fail on execution error: %v", err)
	}

	// The result should be rendered (template execution doesn't fail on missing keys in maps)
	// But if ctx was a struct, it would fail - let's test with struct
	type Data struct {
		Title string
	}

	ctx2 := Data{Title: "Test"}
	tmpl2 := "# {{ .MissingField }}"

	result2, err := display.RenderTemplate(tmpl2, ctx2)
	if err != nil {
		t.Fatalf("RenderTemplate() should not fail on execution error: %v", err)
	}

	// Should fallback to template as-is
	if result2 != tmpl2 {
		t.Errorf("RenderTemplate() with missing struct field result = %q, want fallback %q", result2, tmpl2)
	}

	_ = result // suppress unused warning
}

func TestDisplay_RenderTemplate_NilContext(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	tmpl := "# Static Heading"

	result, err := display.RenderTemplate(tmpl, nil)
	if err != nil {
		t.Fatalf("RenderTemplate() failed with nil context: %v", err)
	}

	if !containsString(result, "Static Heading") {
		t.Errorf("RenderTemplate() result = %q, want to contain %q", result, "Static Heading")
	}
}

func TestDisplay_RenderTemplate_EmptyTemplate(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	result, err := display.RenderTemplate("", nil)
	if err != nil {
		t.Fatalf("RenderTemplate() failed with empty template: %v", err)
	}

	// Empty template produces empty or whitespace result
	_ = result
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
