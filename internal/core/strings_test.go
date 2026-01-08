package core

import (
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
