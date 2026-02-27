package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
	"gopkg.in/yaml.v3"
)

var notesGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get a single note by path",
	Long: `Get a single note and display either list or JSON output.

Examples:
  jot notes get project/plan.md
  jot notes get project/plan --format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		raw, _ := cmd.Flags().GetBool("raw")
		if err := validateGetOutputFlags(format, raw); err != nil {
			return err
		}

		if raw {
			content, err := loadRawNoteByPath(nb, args[0])
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(content)
			return err
		}

		note, err := loadNoteByPath(nb, args[0])
		if err != nil {
			return err
		}

		return renderSingleNoteByFormat(note, format)
	},
}

func init() {
	notesGetCmd.Flags().String("format", "list", "Output format: list or json")
	notesGetCmd.Flags().Bool("raw", false, "Emit exact raw file bytes")
	notesCmd.AddCommand(notesGetCmd)
}

func validateGetOutputFlags(format string, raw bool) error {
	if err := validateOutputFormat(format, "list", "json"); err != nil {
		return err
	}
	if raw && format == "json" {
		return fmt.Errorf("--raw cannot be used with --format json")
	}
	return nil
}

func resolveGetPath(root, notePathArg string) (string, string, error) {
	normalized := filepath.Clean(notePathArg)
	if !strings.HasSuffix(normalized, ".md") {
		normalized += ".md"
	}

	absolutePath := filepath.Join(root, normalized)
	relPath, err := filepath.Rel(root, absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve note path: %w", err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("note path is outside notebook root: %s", notePathArg)
	}

	return absolutePath, filepath.ToSlash(relPath), nil
}

func loadRawNoteByPath(nb *services.Notebook, notePathArg string) ([]byte, error) {
	absolutePath, _, err := resolveGetPath(nb.Config.Root, notePathArg)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("note not found: %s", absolutePath)
		}
		return nil, fmt.Errorf("failed to read note: %w", err)
	}

	return content, nil
}

func loadNoteByPath(nb *services.Notebook, notePathArg string) (services.Note, error) {
	absolutePath, relPath, err := resolveGetPath(nb.Config.Root, notePathArg)
	if err != nil {
		return services.Note{}, err
	}

	content, err := os.ReadFile(absolutePath)
	if err != nil {
		if os.IsNotExist(err) {
			return services.Note{}, fmt.Errorf("note not found: %s", absolutePath)
		}
		return services.Note{}, fmt.Errorf("failed to read note: %w", err)
	}

	metadata, body := parseNoteFrontmatter(content)

	note := services.Note{
		Content:  body,
		Metadata: metadata,
	}
	note.File.Relative = filepath.ToSlash(relPath)
	note.File.Filepath = note.File.Relative

	return note, nil
}

func parseNoteFrontmatter(content []byte) (map[string]any, string) {
	if !strings.HasPrefix(string(content), "---\n") {
		return map[string]any{}, string(content)
	}

	rest := content[4:]
	endIdx := strings.Index(string(rest), "\n---\n")
	if endIdx == -1 {
		return map[string]any{}, string(content)
	}

	frontmatterBytes := rest[:endIdx]
	bodyBytes := rest[endIdx+5:]

	metadata := make(map[string]any)
	if err := yaml.Unmarshal(frontmatterBytes, &metadata); err != nil {
		return map[string]any{}, string(content)
	}

	return metadata, string(bodyBytes)
}

func displayNoteDetail(note services.Note) error {
	title := note.DisplayName()
	if title == "" {
		title = note.File.Relative
	}

	output, err := services.TuiRender("note-detail", map[string]any{
		"Title":    title,
		"File":     note.File,
		"Metadata": note.Metadata,
		"Content":  note.Content,
	})
	if err != nil {
		fmt.Printf("# %s\n\n", title)
		fmt.Printf("File: %s\n\n", note.File.Relative)
		if len(note.Metadata) > 0 {
			fmt.Println("Metadata:")
			for key, value := range note.Metadata {
				fmt.Printf("- %s: %v\n", key, value)
			}
			fmt.Println()
		}
		fmt.Print(note.Content)
		if !strings.HasSuffix(note.Content, "\n") {
			fmt.Println()
		}
		return nil
	}

	fmt.Print(output)
	return nil
}
