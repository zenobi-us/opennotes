package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type noteEnsureResult struct {
	Status  string `json:"status"`
	Path    string `json:"path"`
	Created bool   `json:"created"`
	Action  string `json:"action"`
	Error   string `json:"error,omitempty"`
}

var notesEnsureCmd = &cobra.Command{
	Use:   "ensure <path>",
	Short: "Create a note if it does not exist",
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

		result, err := ensureNoteFile(nb.Config.Root, args[0])
		if err != nil {
			renderNoteEnsureFailure(format, args[0], err.Error())
			return err
		}

		return emitNoteEnsureResult(result, format)
	},
}

func init() {
	notesEnsureCmd.Flags().String("format", "list", "Output format: list or json")
	notesCmd.AddCommand(notesEnsureCmd)
}

func ensureNoteFile(root, notePath string) (noteEnsureResult, error) {
	absolutePath, relativePath, err := resolveUpdateTargetPath(root, notePath)
	if err != nil {
		return noteEnsureResult{}, err
	}

	if _, err := os.Stat(absolutePath); err == nil {
		return noteEnsureResult{Status: "success", Path: relativePath, Created: false, Action: "exists"}, nil
	} else if !os.IsNotExist(err) {
		return noteEnsureResult{}, fmt.Errorf("failed to inspect note: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return noteEnsureResult{}, fmt.Errorf("failed to create target directory: %w", err)
	}
	if err := os.WriteFile(absolutePath, []byte{}, 0o644); err != nil {
		return noteEnsureResult{}, fmt.Errorf("failed to create note: %w", err)
	}

	return noteEnsureResult{Status: "success", Path: relativePath, Created: true, Action: "created"}, nil
}

func emitNoteEnsureResult(result noteEnsureResult, format string) error {
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

func renderNoteEnsureFailure(format, path, message string) {
	result := noteEnsureResult{Status: "failure", Path: path, Error: message}

	if format == "json" {
		_ = renderJSON(result)
		return
	}

	fmt.Printf("status=%s path=%q error=%q\n", result.Status, result.Path, result.Error)
}
