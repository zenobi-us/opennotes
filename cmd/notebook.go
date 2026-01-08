package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var notebookCmd = &cobra.Command{
	Use:     "notebook",
	Aliases: []string{"nb"},
	Short:   "Manage notebooks",
	Long: `Commands for managing notebooks - create, list, register, and configure notebooks.

A notebook is a directory containing markdown notes with a .opennotes.json config file.
When run without a subcommand, displays info about the current notebook.

Examples:
  # Show current notebook info
  opennotes notebook

  # List all notebooks
  opennotes notebook list

  # Create a new notebook
  opennotes notebook create --name "Work Notes"

  # Register existing notebook globally
  opennotes notebook register /path/to/notebook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: show current notebook info
		nb, err := notebookService.Infer("")
		if err != nil {
			return err
		}

		if nb == nil {
			fmt.Println("No notebook found.")
			fmt.Println("")
			fmt.Println("Create one with:")
			fmt.Println("  opennotes notebook create --name \"My Notebook\"")
			return nil
		}

		return displayNotebookInfo(nb)
	},
}

func init() {
	rootCmd.AddCommand(notebookCmd)
}
