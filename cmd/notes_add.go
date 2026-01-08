package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/core"
)

var notesAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new note to the notebook",
	Long:  `Creates a new markdown note in the current notebook with optional template support.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		template, _ := cmd.Flags().GetString("template")
		title, _ := cmd.Flags().GetString("title")

		// Determine filename
		var filename string
		if len(args) > 0 {
			filename = args[0]
		} else {
			// Generate filename from title or timestamp
			if title != "" {
				filename = core.Slugify(title) + ".md"
			} else {
				filename = time.Now().Format("2006-01-02-150405") + ".md"
			}
		}

		// Ensure .md extension
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}

		// Validate note name
		if err := core.ValidateNoteName(filename); err != nil {
			return err
		}

		// Full path to the note
		notePath := filepath.Join(nb.Config.Root, filename)

		// Check if file already exists
		if _, err := os.Stat(notePath); err == nil {
			return fmt.Errorf("note already exists: %s", notePath)
		}

		// Create directories if needed
		noteDir := filepath.Dir(notePath)
		if err := os.MkdirAll(noteDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Generate content
		content := generateNoteContent(title, template, nb.Config.Templates)

		// Write the file
		if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		fmt.Printf("Created note: %s\n", notePath)
		return nil
	},
}

func init() {
	notesAddCmd.Flags().StringP("template", "t", "", "Template to use")
	notesAddCmd.Flags().String("title", "", "Note title")
	notesCmd.AddCommand(notesAddCmd)
}

// generateNoteContent creates the initial note content.
func generateNoteContent(title, templateName string, templates map[string]string) string {
	var content strings.Builder

	// If a template is specified and exists, use it
	if templateName != "" && templates != nil {
		if tmplContent, ok := templates[templateName]; ok {
			// Replace {{title}} placeholder if present
			if title != "" {
				return strings.ReplaceAll(tmplContent, "{{title}}", title)
			}
			return tmplContent
		}
	}

	// Default content with frontmatter
	content.WriteString("---\n")
	if title != "" {
		content.WriteString(fmt.Sprintf("title: %s\n", title))
	}
	content.WriteString(fmt.Sprintf("created: %s\n", time.Now().Format(time.RFC3339)))
	content.WriteString("---\n\n")

	if title != "" {
		content.WriteString(fmt.Sprintf("# %s\n\n", title))
	}

	return content.String()
}
