package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type noteMoveResult struct {
	Status      string `json:"status"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Overwritten bool   `json:"overwritten"`
	Action      string `json:"action"`
	Error       string `json:"error,omitempty"`
}

var notesMvCmd = &cobra.Command{
	Use:     "mv <from> <to>",
	Aliases: []string{"move"},
	Short:   "Move a note within the current notebook",
	Long: `Move a note from one path to another within the notebook root.

By default destination overwrite is blocked. Use --force to replace an existing destination.

Examples:
  jot notes mv tasks/todo.md archive/tasks/todo.md
  jot notes mv tasks/todo archive/tasks/todo --force --format json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		force, _ := cmd.Flags().GetBool("force")
		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		result, err := moveNoteFile(nb.Config.Root, args[0], args[1], force)
		if err != nil {
			renderNoteMoveFailure(format, args[0], args[1], err.Error())
			return err
		}

		return emitNoteMoveResult(result, format)
	},
}

func init() {
	notesMvCmd.Flags().Bool("force", false, "Overwrite destination if it already exists")
	notesMvCmd.Flags().String("format", "list", "Output format: list or json")
	notesCmd.AddCommand(notesMvCmd)
}

func resolveMovePath(root, inputPath string) (absolutePath, relativePath string, err error) {
	normalized := filepath.Clean(inputPath)
	if !strings.HasSuffix(normalized, ".md") {
		normalized += ".md"
	}

	absolutePath = filepath.Join(root, normalized)
	relPath, err := filepath.Rel(root, absolutePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve note path: %w", err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", "", fmt.Errorf("note path is outside notebook root: %s", inputPath)
	}

	return absolutePath, filepath.ToSlash(relPath), nil
}

func moveNoteFile(root, sourceInput, destinationInput string, force bool) (noteMoveResult, error) {
	sourceAbsolute, sourceRelative, err := resolveMovePath(root, sourceInput)
	if err != nil {
		return noteMoveResult{}, err
	}
	destinationAbsolute, destinationRelative, err := resolveMovePath(root, destinationInput)
	if err != nil {
		return noteMoveResult{}, err
	}

	if _, err := os.Stat(sourceAbsolute); err != nil {
		if os.IsNotExist(err) {
			return noteMoveResult{}, withExitCode(ExitCodeNotFound, fmt.Errorf("source note not found: %s", sourceRelative))
		}
		return noteMoveResult{}, fmt.Errorf("failed to inspect source note: %w", err)
	}

	overwritten := false
	if _, err := os.Stat(destinationAbsolute); err == nil {
		if !force {
			return noteMoveResult{}, withExitCode(ExitCodeConflict, fmt.Errorf("destination already exists: %s", destinationRelative))
		}
		overwritten = true
		if err := os.Remove(destinationAbsolute); err != nil {
			return noteMoveResult{}, fmt.Errorf("failed to overwrite destination note: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return noteMoveResult{}, fmt.Errorf("failed to inspect destination note: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(destinationAbsolute), 0o755); err != nil {
		return noteMoveResult{}, fmt.Errorf("failed to create destination directory: %w", err)
	}

	if err := os.Rename(sourceAbsolute, destinationAbsolute); err != nil {
		return noteMoveResult{}, fmt.Errorf("failed to move note: %w", err)
	}

	return noteMoveResult{
		Status:      "success",
		Source:      sourceRelative,
		Destination: destinationRelative,
		Overwritten: overwritten,
		Action:      "moved",
	}, nil
}

func emitNoteMoveResult(result noteMoveResult, format string) error {
	switch format {
	case "json":
		return renderJSON(result)
	case "list":
		fmt.Printf("status=%s action=%s source=%q destination=%q overwritten=%t\n", result.Status, result.Action, result.Source, result.Destination, result.Overwritten)
		return nil
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func renderNoteMoveFailure(format, source, destination, message string) {
	result := noteMoveResult{
		Status:      "failure",
		Source:      source,
		Destination: destination,
		Error:       message,
	}

	if format == "json" {
		_ = renderJSON(result)
		return
	}

	fmt.Printf("status=%s source=%q destination=%q error=%q\n", result.Status, result.Source, result.Destination, result.Error)
}
