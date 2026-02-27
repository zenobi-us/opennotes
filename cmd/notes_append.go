package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

type noteAppendResult struct {
	Status string `json:"status"`
	Path   string `json:"path"`
	Action string `json:"action"`
	Error  string `json:"error,omitempty"`
}

var notesAppendCmd = &cobra.Command{
	Use:   "append <path>",
	Short: "Append content to a note from stdin or file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		create, _ := cmd.Flags().GetBool("create")
		inputFile, _ := cmd.Flags().GetString("input")
		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		stdinAvailable, err := isStdinPiped(os.Stdin)
		if err != nil {
			renderNoteAppendFailure(format, args[0], fmt.Sprintf("failed to inspect stdin: %v", err))
			return err
		}

		content, err := readUpdateContent(inputFile, os.Stdin, stdinAvailable)
		if err != nil {
			renderNoteAppendFailure(format, args[0], err.Error())
			return err
		}

		result, err := appendToNoteFile(nb.Config.Root, args[0], content, create)
		if err != nil {
			renderNoteAppendFailure(format, args[0], err.Error())
			return err
		}

		return emitNoteAppendResult(result, format)
	},
}

func init() {
	notesAppendCmd.Flags().Bool("create", false, "Create note when target does not exist")
	notesAppendCmd.Flags().String("input", "", "Read content from file path instead of stdin")
	notesAppendCmd.Flags().String("format", "list", "Output format: list or json")
	notesCmd.AddCommand(notesAppendCmd)
}

func appendToNoteFile(root, notePath string, content []byte, create bool) (noteAppendResult, error) {
	absolutePath, relativePath, err := resolveUpdateTargetPath(root, notePath)
	if err != nil {
		return noteAppendResult{}, err
	}

	if _, err := os.Stat(absolutePath); err != nil {
		if !os.IsNotExist(err) {
			return noteAppendResult{}, fmt.Errorf("failed to inspect target note: %w", err)
		}
		if !create {
			return noteAppendResult{}, withExitCode(ExitCodeNotFound, fmt.Errorf("target note does not exist (use --create): %s", relativePath))
		}
		if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
			return noteAppendResult{}, fmt.Errorf("failed to create target directory: %w", err)
		}
		if err := os.WriteFile(absolutePath, []byte{}, 0o644); err != nil {
			return noteAppendResult{}, fmt.Errorf("failed to create note: %w", err)
		}
	}

	f, err := os.OpenFile(absolutePath, os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return noteAppendResult{}, fmt.Errorf("failed to open note for append: %w", err)
	}

	if _, err := f.Write(content); err != nil {
		if cerr := f.Close(); cerr != nil {
			return noteAppendResult{}, fmt.Errorf("failed to append note: %w (close error: %v)", err, cerr)
		}
		return noteAppendResult{}, fmt.Errorf("failed to append note: %w", err)
	}

	if err := f.Close(); err != nil {
		return noteAppendResult{}, fmt.Errorf("failed to close note file: %w", err)
	}

	return noteAppendResult{Status: "success", Path: relativePath, Action: "appended"}, nil
}

func emitNoteAppendResult(result noteAppendResult, format string) error {
	switch format {
	case "json":
		return renderJSON(result)
	case "list":
		fmt.Printf("status=%s action=%s path=%q\n", result.Status, result.Action, result.Path)
		return nil
	default:
		return validateOutputFormat(format, "list", "json")
	}
}

func renderNoteAppendFailure(format, path, message string) {
	result := noteAppendResult{Status: "failure", Path: path, Error: message}

	if format == "json" {
		_ = renderJSON(result)
		return
	}

	fmt.Printf("status=%s path=%q error=%q\n", result.Status, result.Path, result.Error)
}
