package cmd

import "github.com/spf13/cobra"

var notebookInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show information about the current notebook",
	Long: `Show information about the current notebook.

Examples:
  jot notebook info
  jot notebook info --format json
  jot notebook info --notebook /path/to/notebook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		return renderNotebookInfoByFormat(nb, format)
	},
}

func init() {
	notebookInfoCmd.Flags().String("format", "list", "Output format: list or json")
	notebookCmd.AddCommand(notebookInfoCmd)
}
