package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple lowercase",
			input:    "Hello World",
			expected: "hello-world",
		},
		{
			name:     "with special characters",
			input:    "Hello, World!",
			expected: "hello-world",
		},
		{
			name:     "with multiple spaces",
			input:    "Hello   World",
			expected: "hello-world",
		},
		{
			name:     "with newlines",
			input:    "Hello\nWorld",
			expected: "hello-world",
		},
		{
			name:     "with leading/trailing hyphens",
			input:    "-Hello World-",
			expected: "hello-world",
		},
		{
			name:     "with numbers",
			input:    "Note 123",
			expected: "note-123",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDedent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no indent",
			input:    "Hello\nWorld",
			expected: "Hello\nWorld",
		},
		{
			name:     "uniform indent",
			input:    "    Hello\n    World",
			expected: "Hello\nWorld",
		},
		{
			name:     "mixed indent",
			input:    "    Hello\n        World",
			expected: "Hello\n    World",
		},
		{
			name:     "empty lines",
			input:    "    Hello\n\n    World",
			expected: "Hello\n\nWorld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Dedent(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestObjectToFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
	}{
		{
			name: "simple values",
			input: map[string]any{
				"title": "Hello",
				"count": 42,
			},
			contains: []string{"title: Hello", "count: 42"},
		},
		{
			name: "array values",
			input: map[string]any{
				"tags": []string{"one", "two"},
			},
			contains: []string{"tags:", "  - one", "  - two"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObjectToFrontmatter(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// === ObjectToFrontmatter Edge Case Tests ===

func TestObjectToFrontmatter_NestedObjects(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
		note     string
	}{
		{
			name: "nested map becomes string representation",
			input: map[string]any{
				"title": "Test",
				"metadata": map[string]any{
					"author": "John",
					"date":   "2024-01-01",
				},
			},
			contains: []string{"title: Test", "metadata: map[author:John date:2024-01-01]"},
			note:     "Nested maps are converted to string representation",
		},
		{
			name: "deeply nested objects",
			input: map[string]any{
				"config": map[string]any{
					"database": map[string]any{
						"host": "localhost",
						"port": 5432,
					},
				},
			},
			contains: []string{"config: map[database:map[host:localhost port:5432]]"},
			note:     "Deep nesting becomes flattened string",
		},
		{
			name: "mixed types in nested object",
			input: map[string]any{
				"settings": map[string]any{
					"enabled": true,
					"count":   10,
					"name":    "test",
				},
			},
			contains: []string{"settings: map["},
			note:     "Mixed types in nested objects work",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObjectToFrontmatter(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Test: %s", tt.note)
			}
		})
	}
}

func TestObjectToFrontmatter_ArrayValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
		note     string
	}{
		{
			name: "mixed type array",
			input: map[string]any{
				"items": []any{"string", 42, true, nil},
			},
			contains: []string{"items:", "  - string", "  - 42", "  - true", "  - <nil>"},
			note:     "Mixed types in any array are converted appropriately",
		},
		{
			name: "empty string array",
			input: map[string]any{
				"tags": []string{},
			},
			contains: []string{"tags:"},
			note:     "Empty arrays create header only",
		},
		{
			name: "empty any array",
			input: map[string]any{
				"items": []any{},
			},
			contains: []string{"items:"},
			note:     "Empty any arrays work the same",
		},
		{
			name: "array with special characters",
			input: map[string]any{
				"commands": []string{"git commit", "echo \"hello\"", "rm -rf /"},
			},
			contains: []string{"commands:", "  - git commit", "  - echo \"hello\"", "  - rm -rf /"},
			note:     "Special characters in array items are preserved",
		},
		{
			name: "array with unicode",
			input: map[string]any{
				"names": []string{"José", "François", "München"},
			},
			contains: []string{"names:", "  - José", "  - François", "  - München"},
			note:     "Unicode in arrays works correctly",
		},
		{
			name: "nested arrays in any array",
			input: map[string]any{
				"matrix": []any{[]string{"a", "b"}, []int{1, 2}},
			},
			contains: []string{"matrix:", "  - [a b]", "  - [1 2]"},
			note:     "Nested arrays become string representations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObjectToFrontmatter(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Test: %s", tt.note)
			}
		})
	}
}

func TestObjectToFrontmatter_SpecialValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
		note     string
	}{
		{
			name: "nil values",
			input: map[string]any{
				"title":       "Test",
				"description": nil,
				"count":       0,
				"enabled":     false,
			},
			contains: []string{"title: Test", "description: <nil>", "count: 0", "enabled: false"},
			note:     "Nil and zero values are handled correctly",
		},
		{
			name: "boolean values",
			input: map[string]any{
				"published": true,
				"draft":     false,
			},
			contains: []string{"published: true", "draft: false"},
			note:     "Boolean values work correctly",
		},
		{
			name: "numeric types",
			input: map[string]any{
				"int":     42,
				"int64":   int64(1000000000),
				"float32": float32(3.14),
				"float64": 2.71828,
			},
			contains: []string{"int: 42", "int64: 1000000000", "float32: 3.14", "float64: 2.71828"},
			note:     "Different numeric types work",
		},
		{
			name: "empty string",
			input: map[string]any{
				"title": "",
				"body":  "content",
			},
			contains: []string{"title: ", "body: content"},
			note:     "Empty strings are preserved",
		},
		{
			name: "whitespace strings",
			input: map[string]any{
				"spaces": "   ",
				"tabs":   "\t\t",
				"mixed":  " \t \n ",
			},
			contains: []string{"spaces:    ", "tabs: \t\t", "mixed:  \t \n "},
			note:     "Whitespace-only strings are preserved",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObjectToFrontmatter(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Test: %s", tt.note)
			}
		})
	}
}

func TestObjectToFrontmatter_UnicodeAndEscaping(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
		note     string
	}{
		{
			name: "unicode characters",
			input: map[string]any{
				"title":  "Café Notes",
				"author": "José María González",
				"city":   "München",
				"symbol": "£€¥₹",
			},
			contains: []string{"title: Café Notes", "author: José María González", "city: München", "symbol: £€¥₹"},
			note:     "Unicode characters are preserved correctly",
		},
		{
			name: "special characters that might need escaping",
			input: map[string]any{
				"quote":     "She said \"hello\"",
				"colon":     "key: value format",
				"newline":   "line1\nline2",
				"backslash": "path\\to\\file",
			},
			contains: []string{
				"quote: She said \"hello\"",
				"colon: key: value format",
				"newline: line1\nline2",
				"backslash: path\\to\\file",
			},
			note: "Special characters are included as-is (YAML escaping handled by consumer)",
		},
		{
			name: "emoji and symbols",
			input: map[string]any{
				"status": "✅ Complete",
				"mood":   "😊 Happy",
				"math":   "x² + y² = z²",
			},
			contains: []string{"status: ✅ Complete", "mood: 😊 Happy", "math: x² + y² = z²"},
			note:     "Emoji and mathematical symbols work",
		},
		{
			name: "very long string",
			input: map[string]any{
				"long": strings.Repeat("Lorem ipsum ", 100),
			},
			contains: []string{"long: " + strings.Repeat("Lorem ipsum ", 10)}, // Check beginning
			note:     "Very long strings are handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObjectToFrontmatter(tt.input)
			for _, expected := range tt.contains {
				assert.Contains(t, result, expected, "Test: %s", tt.note)
			}
		})
	}
}

func TestObjectToFrontmatter_ErrorConditions(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]any
		contains []string
		note     string
	}{
		{
			name: "empty map",
			input: map[string]any{},
			contains: []string{},
			note:     "Empty map should produce empty result",
		},
		{
			name: "complex unsupported types",
			input: map[string]any{
				"channel": make(chan int),
				"func":    func() {},
			},
			contains: []string{"channel:", "func:"},
			note:     "Complex types get converted to string representation",
		},
		{
			name: "keys with special characters",
			input: map[string]any{
				"normal-key":    "value1",
				"key with space": "value2",
				"key:with:colon": "value3",
				"key\"quote":     "value4",
			},
			contains: []string{
				"normal-key: value1",
				"key with space: value2", 
				"key:with:colon: value3",
				"key\"quote: value4",
			},
			note: "Keys with special characters are preserved as-is",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ObjectToFrontmatter(tt.input)
			
			if len(tt.contains) == 0 {
				assert.Empty(t, result, "Test: %s", tt.note)
			} else {
				for _, expected := range tt.contains {
					assert.Contains(t, result, expected, "Test: %s", tt.note)
				}
			}
		})
	}
}
