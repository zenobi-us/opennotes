package services

import (
	"strings"
	"testing"
)

func TestSlugWithMax(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		max      int
		expected string
	}{
		{
			name:     "Short title unchanged",
			title:    "hello",
			max:      50,
			expected: "hello",
		},
		{
			name:     "Exactly max length unchanged",
			title:    strings.Repeat("a", 50),
			max:      50,
			expected: strings.Repeat("a", 50),
		},
		{
			name:     "Long title truncates at word boundary",
			title:    "this-is-a-very-long-title-that-exceeds-the-maximum-length-allowed",
			max:      50,
			expected: "this-is-a-very-long-title-that-exceeds-the",
		},
		{
			name:     "Long single word hard truncates",
			title:    "supercalifragilisticexpialidocioussupercalifragilisticexpialidocious",
			max:      50,
			expected: "supercalifragilisticexpialidocioussupercalifragili",
		},
		{
			name:     "Empty input returns untitled",
			title:    "",
			max:      50,
			expected: "untitled",
		},
		{
			name:     "Hyphen too early falls back to hard truncate",
			title:    "ab-" + strings.Repeat("c", 60),
			max:      50,
			expected: "ab-" + strings.Repeat("c", 47),
		},
		{
			name:     "Hyphen exactly at half uses word boundary",
			title:    strings.Repeat("a", 26) + "-" + strings.Repeat("b", 30),
			max:      50,
			expected: strings.Repeat("a", 26),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SlugWithMax(tt.title, tt.max)
			if result != tt.expected {
				t.Errorf("SlugWithMax(%q, %d) = %q (len=%d), want %q (len=%d)",
					tt.title, tt.max, result, len(result), tt.expected, len(tt.expected))
			}
		})
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Chinese characters transliteration",
			title:    "会议",
			expected: "hui-yi",
		},
		{
			name:     "Emoji stripping",
			title:    "🎉 Party",
			expected: "party",
		},
		{
			name:     "Accent normalization",
			title:    "café",
			expected: "cafe",
		},
		{
			name:     "ASCII preservation",
			title:    "hello-world",
			expected: "hello-world",
		},
		{
			name:     "Empty fallback for all emoji",
			title:    "🎉🎊🎈",
			expected: "untitled",
		},
		{
			name:     "Mixed content",
			title:    "Meeting Notes 2024",
			expected: "meeting-notes-2024",
		},
		{
			name:     "Spaces to hyphens",
			title:    "my awesome note",
			expected: "my-awesome-note",
		},
		{
			name:     "Special characters handled",
			title:    "hello & world!",
			expected: "hello-and-world",
		},
		{
			name:     "Empty string fallback",
			title:    "",
			expected: "untitled",
		},
		{
			name:     "Whitespace only fallback",
			title:    "   ",
			expected: "untitled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slug(tt.title)
			if result != tt.expected {
				t.Errorf("Slug(%q) = %q, want %q", tt.title, result, tt.expected)
			}
		})
	}
}
