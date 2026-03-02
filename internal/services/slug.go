package services

import (
	"github.com/gosimple/slug"
)

func init() {
	// Configure slug settings globally
	slug.Lowercase = true
}

// SlugWithMax converts a title to a filesystem-safe slug with a maximum length.
// It handles Unicode characters (including CJK), removes emojis,
// and normalizes accented characters.
// If the slug exceeds max length, it truncates at the last word boundary (hyphen)
// before max. Falls back to hard truncate if no suitable word boundary exists
// (hyphen must be > max/2 to be considered).
// Returns "untitled" if the result would be empty (e.g., all-emoji input).
func SlugWithMax(title string, max int) string {
	result := slug.Make(title)
	if result == "" {
		return "untitled"
	}

	// If within max length, return as-is
	if len(result) <= max {
		return result
	}

	// Try to truncate at word boundary (last hyphen before max)
	truncated := result[:max]
	lastHyphen := -1
	for i := len(truncated) - 1; i >= 0; i-- {
		if truncated[i] == '-' {
			lastHyphen = i
			break
		}
	}

	// Use word boundary if hyphen is past halfway point (> max/2)
	if lastHyphen > max/2 {
		return truncated[:lastHyphen]
	}

	// Fall back to hard truncate
	return truncated
}

// Slug converts a title to a filesystem-safe slug.
// It handles Unicode characters (including CJK), removes emojis,
// and normalizes accented characters.
// Returns "untitled" if the result would be empty (e.g., all-emoji input).
// Default maximum length is 50 characters.
func Slug(title string) string {
	return SlugWithMax(title, 50)
}
