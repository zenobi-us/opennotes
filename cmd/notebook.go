package cmd

import (
	"github.com/spf13/cobra"
)

var notebookCmd = &cobra.Command{
	Use:     "notebook",
	Aliases: []string{"nb"},
	Short:   "Manage notebooks - create, list, auto-discovery",
	Long: `Commands for managing notebooks - create, list, register, and configure notebooks.

A notebook is a directory containing markdown notes with a .jot.json config file.
When run without a subcommand, displays info about the current notebook.

QUICK START WITH EXISTING MARKDOWN:
  1. Import: jot notebook create "My Notes" --path ~/my-notes
  2. Verify: jot notes list
  3. Search: jot notes search "meeting"

AUTO-DISCOVERY:
  - Notebooks are discovered by looking for .jot.json in current directory or ancestors
  - This allows context-aware workflows: cd ~/work/notes && jot notes list (uses work notebook)
  - Perfect for multi-project setups where each project has its own notes

DOCUMENTATION:
  📋 Notebook Discovery & Management: https://github.com/zenobi-us/jot/blob/main/docs/notebook-discovery.md

Examples:
  # Show current notebook info
  jot notebook

  # List all notebooks
  jot notebook list

  # Create a notebook with existing markdown files
  jot notebook create "My Notes" --path ~/my-notes

  # Create new empty notebook
  jot notebook create --name "Work Notes"

  # Register existing notebook globally
  jot notebook register /path/to/notebook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		// Default: show current notebook info
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		return renderNotebookInfoByFormat(nb, format)
	},
}

func init() {
	notebookCmd.Flags().String("format", "list", "Output format: list or json")
	rootCmd.AddCommand(notebookCmd)
}
