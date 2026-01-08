package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var notebookAddContextCmd = &cobra.Command{
	Use:   "add-context [path]",
	Short: "Add a context path to the current notebook",
	Long: `Adds a directory path as a context for the current notebook.

When working in a context directory (or any subdirectory), the notebook
will be automatically selected. This is useful for associating project
directories with specific notebooks.

Examples:
  # Add current directory as context
  opennotes notebook add-context

  # Add specific path as context
  opennotes notebook add-context ~/projects/myapp`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		contextPath := ""
		if len(args) > 0 {
			contextPath = args[0]
		} else {
			contextPath, _ = os.Getwd()
		}

		// Get the current notebook
		nb, err := notebookService.Infer("")
		if err != nil {
			return err
		}

		if nb == nil {
			return fmt.Errorf("no notebook found. Create one first with: opennotes notebook create --name \"My Notebook\"")
		}

		if err := nb.AddContext(contextPath, cfgService); err != nil {
			return fmt.Errorf("failed to add context: %w", err)
		}

		fmt.Printf("Added context '%s' to notebook '%s'\n", contextPath, nb.Config.Name)
		return nil
	},
}

func init() {
	notebookCmd.AddCommand(notebookAddContextCmd)
}
