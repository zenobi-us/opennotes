package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/core"
	"github.com/zenobi-us/jot/internal/services"
	"gopkg.in/yaml.v3"
)

var notesAddCmd = &cobra.Command{
	Use:   "add <title> [path]",
	Short: "Add a new note to the notebook",
	Long: `Creates a new markdown note in the current notebook with optional metadata and template support.

SYNTAX:
  jot notes add <title> [path] [flags]          # New style (recommended)
  jot notes add [path] --title "Title" [flags]  # Old style (deprecated)

EXAMPLES:
  # Create note in root
  jot notes add "Quick Thought"
  
  # Create note in folder
  jot notes add "Meeting Notes" meetings/
  
  # Create note with metadata
  jot notes add "Sprint Planning" meetings/ \
    --data tag=meeting --data priority=high
  
  # Pipe content from stdin
  echo "# Content" | jot notes add "My Note"
  
  # Use template
  jot notes add "Bug Report" bugs/ --template bug`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		template, _ := cmd.Flags().GetString("template")
		titleFlag, _ := cmd.Flags().GetString("title")
		titleFlagProvided := cmd.Flags().Changed("title")
		dataFlags, _ := cmd.Flags().GetStringArray("data")

		// Parse arguments (title and optional path)
		title, pathArg, err := parseArguments(args, titleFlag, titleFlagProvided)
		if err != nil {
			return err
		}

		// Show deprecation warning if --title flag used
		if titleFlagProvided {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: --title flag is deprecated, use positional argument instead. Will be removed in v2.0.0\n")
		}

		// Parse --data flags
		customData, err := services.ParseDataFlags(dataFlags)
		if err != nil {
			return fmt.Errorf("parsing --data flags: %w", err)
		}

		// Determine filename
		var notePath string
		if pathArg != "" {
			// If path is provided, use path resolution
			slugifiedTitle := core.Slugify(title)
			if slugifiedTitle == "" && title != "" {
				return fmt.Errorf("title produces empty filename after slugification")
			}
			notePath = services.ResolvePath(nb.Config.Root, pathArg, slugifiedTitle)
		} else if title != "" {
			// If only title is provided, slugify it
			slugifiedTitle := core.Slugify(title)
			if slugifiedTitle == "" {
				return fmt.Errorf("title produces empty filename after slugification")
			}
			notePath = services.ResolvePath(nb.Config.Root, "", slugifiedTitle)
		} else {
			// No title and no path - generate timestamp-based name
			timestamp := time.Now().Format("2006-01-02-150405")
			notePath = filepath.Join(nb.Config.Root, timestamp+".md")
		}

		// Check if file already exists
		if _, err := os.Stat(notePath); err == nil {
			return fmt.Errorf("note already exists: %s", notePath)
		}

		// Enforce workflow rules before creating note
		if err := enforceWorkflowForCreate(nb, notePath, customData); err != nil {
			return err
		}

		// Create directories if needed
		noteDir := filepath.Dir(notePath)
		if err := os.MkdirAll(noteDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Check for stdin content
		stdinContent, err := readStdin()
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}

		// Generate content (stdin > template > default)
		var content string
		if stdinContent != "" {
			content = stdinContent
		} else if template != "" {
			content = generateNoteContent(title, template, nb.Config.Templates)
		} else {
			if title != "" {
				content = fmt.Sprintf("# %s\n\n", title)
			} else {
				content = "\n"
			}
		}

		// Generate frontmatter with custom data
		frontmatter := generateFrontmatter(title, customData)

		// Construct final content with frontmatter
		finalContent := fmt.Sprintf("---\n%s---\n\n%s", frontmatter, content)

		// Write the file
		if err := os.WriteFile(notePath, []byte(finalContent), 0644); err != nil {
			return fmt.Errorf("failed to create note: %w", err)
		}

		fmt.Printf("Created note: %s\n", notePath)
		return nil
	},
}

func init() {
	notesAddCmd.Flags().StringP("template", "t", "", "Template to use")
	notesAddCmd.Flags().String("title", "", "Note title (DEPRECATED: use positional argument)")
	notesAddCmd.Flags().StringArray("data", []string{}, "Set frontmatter field (repeatable, format: field=value)")
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

// parseArguments parses command arguments to extract title and path
func parseArguments(args []string, titleFlag string, titleFlagProvided bool) (title, path string, err error) {
	if titleFlagProvided {
		// Old style: --title flag was used, args[0] is path (if provided)
		title = titleFlag
		if len(args) > 0 {
			path = args[0]
		}
		// Error if more than 1 positional arg when using --title
		if len(args) > 1 {
			return "", "", fmt.Errorf("too many arguments: when using --title flag, only one path argument is allowed")
		}
	} else {
		// New style: no --title flag
		if len(args) > 0 {
			title = args[0]
		}
		if len(args) > 1 {
			path = args[1]
		}
		// Error if more than 2 positional args
		if len(args) > 2 {
			return "", "", fmt.Errorf("too many arguments: expected <title> [path]")
		}
	}

	return title, path, nil
}

// readStdin reads content from stdin if available
func readStdin() (string, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	// Check if stdin is piped
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return "", nil // No stdin
	}

	bytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// generateFrontmatter creates frontmatter with title and custom data
func generateFrontmatter(title string, customData map[string]interface{}) string {
	fm := map[string]interface{}{
		"created": time.Now().Format(time.RFC3339),
	}

	// Add title if not empty
	if title != "" {
		fm["title"] = title
	}

	// Merge custom data
	for k, v := range customData {
		if k == "title" && title != "" {
			fmt.Fprintf(os.Stderr, "⚠️  Warning: title field in --data is redundant (use positional argument instead)\n")
		}
		fm[k] = v
	}

	// Serialize to YAML
	fmBytes, err := yaml.Marshal(fm)
	if err != nil {
		// Fallback to simple format if YAML fails
		if title != "" {
			return fmt.Sprintf("title: %s\ncreated: %s\n", title, time.Now().Format(time.RFC3339))
		}
		return fmt.Sprintf("created: %s\n", time.Now().Format(time.RFC3339))
	}

	return string(fmBytes)
}
