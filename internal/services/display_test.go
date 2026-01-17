package services

import (
	"bytes"
	"io"
	"os"
	"strings"
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

// Tests for RenderSQLResults

// captureOutput captures stdout during function execution
func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestDisplay_RenderSQLResults_EmptyResults(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults([]map[string]interface{}{})
	})

	if !strings.Contains(output, "No results") {
		t.Errorf("RenderSQLResults() with empty results = %q, want to contain 'No results'", output)
	}
}

func TestDisplay_RenderSQLResults_SingleRow(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{
			"name":  "John",
			"email": "john@example.com",
			"age":   30,
		},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Check headers
	if !strings.Contains(output, "age") {
		t.Errorf("RenderSQLResults() output missing column 'age'")
	}
	if !strings.Contains(output, "email") {
		t.Errorf("RenderSQLResults() output missing column 'email'")
	}
	if !strings.Contains(output, "name") {
		t.Errorf("RenderSQLResults() output missing column 'name'")
	}

	// Check data
	if !strings.Contains(output, "John") {
		t.Errorf("RenderSQLResults() output missing data 'John'")
	}
	if !strings.Contains(output, "30") {
		t.Errorf("RenderSQLResults() output missing data '30'")
	}

	// Check row count
	if !strings.Contains(output, "1 row") {
		t.Errorf("RenderSQLResults() output missing '1 row' summary")
	}
}

func TestDisplay_RenderSQLResults_MultipleRows(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Check headers
	if !strings.Contains(output, "id") {
		t.Errorf("RenderSQLResults() output missing column 'id'")
	}
	if !strings.Contains(output, "name") {
		t.Errorf("RenderSQLResults() output missing column 'name'")
	}

	// Check data
	if !strings.Contains(output, "Alice") {
		t.Errorf("RenderSQLResults() output missing 'Alice'")
	}
	if !strings.Contains(output, "Bob") {
		t.Errorf("RenderSQLResults() output missing 'Bob'")
	}
	if !strings.Contains(output, "Charlie") {
		t.Errorf("RenderSQLResults() output missing 'Charlie'")
	}

	// Check row count
	if !strings.Contains(output, "3 rows") {
		t.Errorf("RenderSQLResults() output missing '3 rows' summary")
	}
}

func TestDisplay_RenderSQLResults_ColumnAlignment(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"short": "a", "verylongname": "value1"},
		{"short": "abcdef", "verylongname": "v2"},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Should have header, separator, 2 data rows, blank line, summary = 6 lines
	if len(lines) < 5 {
		t.Errorf("RenderSQLResults() output has %d lines, want at least 5", len(lines))
	}

	// Header and data rows should have consistent structure
	headerLine := lines[0]
	separatorLine := lines[1]
	dataLine1 := lines[2]

	// All should contain the columns
	if !strings.Contains(headerLine, "short") {
		t.Error("Header missing 'short' column")
	}
	if !strings.Contains(separatorLine, "-") {
		t.Error("Separator line should contain dashes")
	}
	if !strings.Contains(dataLine1, "a") || !strings.Contains(dataLine1, "value1") {
		t.Error("Data line missing expected values")
	}
}

func TestDisplay_RenderSQLResults_DifferentTypes(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{
			"string_col": "text",
			"int_col":    42,
			"float_col":  3.14,
			"bool_col":   true,
		},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// All values should be present and formatted
	if !strings.Contains(output, "text") {
		t.Error("RenderSQLResults() missing string value")
	}
	if !strings.Contains(output, "42") {
		t.Error("RenderSQLResults() missing int value")
	}
	if !strings.Contains(output, "3.14") {
		t.Error("RenderSQLResults() missing float value")
	}
	if !strings.Contains(output, "true") {
		t.Error("RenderSQLResults() missing bool value")
	}
}

func TestDisplay_RenderSQLResults_ColumnSorting(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	// Create results with columns in non-alphabetical order
	results := []map[string]interface{}{
		{"zebra": 1, "apple": 2, "middle": 3},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	lines := strings.Split(strings.TrimSpace(output), "\n")
	headerLine := lines[0]

	// Find positions of columns in header
	applePos := strings.Index(headerLine, "apple")
	middlePos := strings.Index(headerLine, "middle")
	zebraPos := strings.Index(headerLine, "zebra")

	// Columns should be in alphabetical order
	if applePos == -1 || middlePos == -1 || zebraPos == -1 {
		t.Fatal("RenderSQLResults() missing expected columns")
	}

	if !(applePos < middlePos && middlePos < zebraPos) {
		t.Errorf("RenderSQLResults() columns not sorted: apple@%d, middle@%d, zebra@%d",
			applePos, middlePos, zebraPos)
	}
}

func TestDisplay_RenderSQLResults_NilValues(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	results := []map[string]interface{}{
		{"col1": "value", "col2": nil},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should handle nil values gracefully
	if !strings.Contains(output, "col1") {
		t.Error("RenderSQLResults() missing column with value")
	}
	if !strings.Contains(output, "col2") {
		t.Error("RenderSQLResults() missing column with nil value")
	}
}

func TestDisplay_RenderSQLResults_LargeValues(t *testing.T) {
	display, err := NewDisplay()
	if err != nil {
		t.Fatalf("NewDisplay() failed: %v", err)
	}

	longString := strings.Repeat("x", 100)
	results := []map[string]interface{}{
		{"short": "s", "long": longString},
	}

	output := captureOutput(func() {
		_ = display.RenderSQLResults(results)
	})

	// Should contain the long string even if column is wide
	if !strings.Contains(output, "long") {
		t.Error("RenderSQLResults() missing 'long' column header")
	}
	if !strings.Contains(output, "x") {
		t.Error("RenderSQLResults() missing long string data")
	}
}
