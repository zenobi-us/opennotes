package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type noteUpdateResult struct {
	Status  string `json:"status"`
	Path    string `json:"path"`
	Created bool   `json:"created"`
	Action  string `json:"action"`
	Error   string `json:"error,omitempty"`
}

var notesUpdateCmd = &cobra.Command{
	Use:     "update <path>",
	Aliases: []string{"put"},
	Short:   "Replace note content from stdin or file input",
	Long: `Replace note content for an existing note.

Default behavior is replace-only. Use --create to allow creating a missing target.

Examples:
  echo "# Updated" | jot notes update docs/plan.md
  jot notes update docs/plan.md --input /tmp/content.md
  jot notes put docs/new.md --input /tmp/content.md --create --format json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		targetPath := args[0]
		create, _ := cmd.Flags().GetBool("create")
		inputFile, _ := cmd.Flags().GetString("input")
		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		stdinAvailable, err := isStdinPiped(os.Stdin)
		if err != nil {
			renderNoteUpdateFailure(format, targetPath, fmt.Sprintf("failed to inspect stdin: %v", err))
			return err
		}

		content, err := readUpdateContent(inputFile, os.Stdin, stdinAvailable)
		if err != nil {
			renderNoteUpdateFailure(format, targetPath, err.Error())
			return err
		}

		result, err := updateNoteFile(nb.Config.Root, targetPath, content, create)
		if err != nil {
			renderNoteUpdateFailure(format, targetPath, err.Error())
			return err
		}

		return emitNoteUpdateResult(result, format)
	},
}

func init() {
	notesUpdateCmd.Flags().Bool("create", false, "Create note when target does not exist")
	notesUpdateCmd.Flags().String("input", "", "Read content from file path instead of stdin")
	notesUpdateCmd.Flags().String("format", "list", "Output format: list or json")
	notesCmd.AddCommand(notesUpdateCmd)
}

func isStdinPiped(stdin *os.File) (bool, error) {
	stat, err := stdin.Stat()
	if err != nil {
		return false, err
	}
	return (stat.Mode() & os.ModeCharDevice) == 0, nil
}

func readUpdateContent(inputFile string, stdin io.Reader, stdinAvailable bool) ([]byte, error) {
	if inputFile != "" && stdinAvailable {
		return nil, fmt.Errorf("input source conflict: use either --input or stdin")
	}

	if inputFile != "" {
		content, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read input file: %w", err)
		}
		return content, nil
	}

	if !stdinAvailable {
		return nil, fmt.Errorf("no input provided: pipe stdin or use --input <file>")
	}

	content, err := io.ReadAll(stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	return content, nil
}

func resolveUpdateTargetPath(root, inputPath string) (absolutePath, relativePath string, err error) {
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

func updateNoteFile(root, targetPath string, content []byte, create bool) (noteUpdateResult, error) {
	absolutePath, relativePath, err := resolveUpdateTargetPath(root, targetPath)
	if err != nil {
		return noteUpdateResult{}, err
	}

	_, statErr := os.Stat(absolutePath)
	created := false
	if statErr != nil {
		if !os.IsNotExist(statErr) {
			return noteUpdateResult{}, fmt.Errorf("failed to inspect target note: %w", statErr)
		}
		if !create {
			return noteUpdateResult{}, fmt.Errorf("target note does not exist (use --create): %s", relativePath)
		}
		created = true
		if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
			return noteUpdateResult{}, fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	if err := os.WriteFile(absolutePath, content, 0o644); err != nil {
		return noteUpdateResult{}, fmt.Errorf("failed to write note: %w", err)
	}

	action := "updated"
	if created {
		action = "created"
	}

	return noteUpdateResult{
		Status:  "success",
		Path:    relativePath,
		Created: created,
		Action:  action,
	}, nil
}

func emitNoteUpdateResult(result noteUpdateResult, format string) error {
	switch format {
	case "json":
		return renderJSON(result)
	case "list":
		fmt.Printf("status=%s action=%s path=%q created=%t\n", result.Status, result.Action, result.Path, result.Created)
		return nil
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func renderNoteUpdateFailure(format, targetPath, message string) {
	result := noteUpdateResult{
		Status: "failure",
		Path:   targetPath,
		Error:  message,
	}

	if format == "json" {
		_ = renderJSON(result)
		return
	}

	fmt.Printf("status=%s path=%q error=%q\n", result.Status, result.Path, result.Error)
}
