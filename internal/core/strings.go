package core

import (
	"fmt"
	"regexp"
	"strings"
)

// Slugify converts text to lowercase, removes special chars, replaces spaces with hyphens.
func Slugify(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Replace newlines with spaces
	text = strings.ReplaceAll(text, "\n", " ")

	// Remove special characters except alphanumeric, spaces, and hyphens
	re := regexp.MustCompile(`[^a-z0-9\s-]`)
	text = re.ReplaceAllString(text, "")

	// Replace spaces with hyphens
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, "-")

	// Trim leading/trailing hyphens
	text = strings.Trim(text, "-")

	return text
}

// Dedent removes leading whitespace from multiline strings based on the minimum indent.
func Dedent(text string) string {
	lines := strings.Split(text, "\n")

	// Find minimum indent from non-empty lines
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		indent := 0
		for _, ch := range line {
			if ch == ' ' || ch == '\t' {
				indent++
			} else {
				break
			}
		}

		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return text
	}

	// Remove minimum indent from each line
	result := make([]string, len(lines))
	for i, line := range lines {
		if len(line) >= minIndent {
			result[i] = line[minIndent:]
		} else {
			result[i] = strings.TrimLeft(line, " \t")
		}
	}

	return strings.Join(result, "\n")
}

// ObjectToFrontmatter converts a map to YAML-style frontmatter.
func ObjectToFrontmatter(obj map[string]any) string {
	var lines []string

	for key, value := range obj {
		switch v := value.(type) {
		case []string:
			lines = append(lines, fmt.Sprintf("%s:", key))
			for _, item := range v {
				lines = append(lines, fmt.Sprintf("  - %s", item))
			}
		case []any:
			lines = append(lines, fmt.Sprintf("%s:", key))
			for _, item := range v {
				lines = append(lines, fmt.Sprintf("  - %v", item))
			}
		default:
			lines = append(lines, fmt.Sprintf("%s: %v", key, v))
		}
	}

	return strings.Join(lines, "\n")
}
