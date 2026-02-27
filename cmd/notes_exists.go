package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type noteExistsResult struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Error  string `json:"error,omitempty"`
}

var notesExistsCmd = &cobra.Command{
	Use:   "exists <path>",
	Short: "Check whether a note exists",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		result, err := checkNoteExists(nb.Config.Root, args[0])
		if err != nil {
			renderNoteExistsFailure(format, args[0], err.Error())
			return err
		}

		return emitNoteExistsResult(result, format)
	},
}

func init() {
	notesExistsCmd.Flags().String("format", "list", "Output format: list or json")
	notesCmd.AddCommand(notesExistsCmd)
}

func checkNoteExists(root, notePath string) (noteExistsResult, error) {
	absolutePath, relativePath, err := resolveUpdateTargetPath(root, notePath)
	if err != nil {
		return noteExistsResult{}, err
	}

	if _, err := os.Stat(absolutePath); err != nil {
		if os.IsNotExist(err) {
			return noteExistsResult{}, withExitCode(ExitCodeNotFound, fmt.Errorf("note not found: %s", relativePath))
		}
		return noteExistsResult{}, fmt.Errorf("failed to inspect note: %w", err)
	}

	return noteExistsResult{Status: "success", Path: relativePath, Exists: true}, nil
}

func emitNoteExistsResult(result noteExistsResult, format string) error {
	switch format {
	case "json":
		return renderJSON(result)
	case "list":
		fmt.Printf("status=%s path=%q exists=%t\n", result.Status, result.Path, result.Exists)
		return nil
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func renderNoteExistsFailure(format, path, message string) {
	result := noteExistsResult{Status: "failure", Path: path, Exists: false, Error: message}

	if format == "json" {
		_ = renderJSON(result)
		return
	}

	fmt.Printf("status=%s path=%q exists=%t error=%q\n", result.Status, result.Path, result.Exists, result.Error)
}
