package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/zenobi-us/jot/internal/services"
	"gopkg.in/yaml.v3"
)

// enforceWorkflowForCreate checks workflow rules before note creation.
// Returns nil if allowed, or an error (with ExitCodeWorkflowBlocked) if blocked.
func enforceWorkflowForCreate(
	nb *services.Notebook,
	notePath string,
	metadata map[string]any,
) error {
	if len(nb.Config.Groups) == 0 || len(nb.Config.Workflows) == 0 {
		return nil
	}

	relPath, err := filepath.Rel(nb.Config.Root, notePath)
	if err != nil {
		relPath = notePath
	}
	relPath = filepath.ToSlash(relPath)

	result := services.EnforceWorkflowOnMutation(
		relPath,
		nb.Config.Groups,
		nb.Config.Workflows,
		nil, // no existing metadata for create
		metadata,
		true, // isCreate
	)

	if !result.Enforced || result.Allowed {
		return nil
	}

	return withExitCode(ExitCodeWorkflowBlocked,
		fmt.Errorf("workflow %s blocked note creation: %s", result.WorkflowID, formatDiagnostics(result.Diagnostics)))
}

// enforceWorkflowForEdit checks workflow rules before note edit/update.
// existingContent is the raw bytes of the current note file.
// newMetadata is the frontmatter metadata of the incoming content.
// Returns nil if allowed, or an error (with ExitCodeWorkflowBlocked) if blocked.
func enforceWorkflowForEdit(
	nb *services.Notebook,
	notePath string,
	existingContent []byte,
	newMetadata map[string]any,
) error {
	if len(nb.Config.Groups) == 0 || len(nb.Config.Workflows) == 0 {
		return nil
	}

	relPath, err := filepath.Rel(nb.Config.Root, notePath)
	if err != nil {
		relPath = notePath
	}
	relPath = filepath.ToSlash(relPath)

	existingMeta := extractFrontmatterMetadata(existingContent)

	result := services.EnforceWorkflowOnMutation(
		relPath,
		nb.Config.Groups,
		nb.Config.Workflows,
		existingMeta,
		newMetadata,
		false, // isCreate
	)

	if !result.Enforced || result.Allowed {
		return nil
	}

	return withExitCode(ExitCodeWorkflowBlocked,
		fmt.Errorf("workflow %s blocked note update: %s", result.WorkflowID, formatDiagnostics(result.Diagnostics)))
}

// extractFrontmatterMetadata parses frontmatter from raw note bytes.
// Returns an empty map if no frontmatter found.
func extractFrontmatterMetadata(content []byte) map[string]any {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return map[string]any{}
	}

	rest := s[4:]
	endIdx := strings.Index(rest, "\n---\n")
	if endIdx == -1 {
		// Try end-of-file variant: "---\n" followed by "---" at EOF
		if strings.HasSuffix(strings.TrimRight(rest, "\n"), "---") {
			endIdx = strings.LastIndex(rest, "\n---")
			if endIdx == -1 {
				return map[string]any{}
			}
		} else {
			return map[string]any{}
		}
	}

	fmBytes := rest[:endIdx]
	// Use yaml for parsing, imported as gopkg.in/yaml.v3
	var metadata map[string]any
	if err := yaml.Unmarshal([]byte(fmBytes), &metadata); err != nil {
		return map[string]any{}
	}

	return metadata
}

// formatDiagnostics formats workflow diagnostics into a human-readable string.
func formatDiagnostics(diagnostics []services.WorkflowDiagnostic) string {
	if len(diagnostics) == 0 {
		return "no details"
	}

	parts := make([]string, 0, len(diagnostics))
	for _, d := range diagnostics {
		parts = append(parts, fmt.Sprintf("[%s] %s", d.Code, d.Message))
	}
	return strings.Join(parts, "; ")
}
