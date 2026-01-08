package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var notebookRegisterCmd = &cobra.Command{
	Use:   "register [path]",
	Short: "Register an existing notebook globally",
	Long: `Registers an existing notebook directory in the global configuration.

This adds the notebook to ~/.config/opennotes/config.json so it can be
discovered from anywhere using context paths.

Examples:
  # Register current directory
  opennotes notebook register

  # Register specific path
  opennotes notebook register /path/to/notebook`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		} else {
			path, _ = os.Getwd()
		}

		// Verify it's a valid notebook
		if !notebookService.HasNotebook(path) {
			return fmt.Errorf("no notebook found at %s", path)
		}

		nb, err := notebookService.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open notebook: %w", err)
		}

		// Register it
		if err := nb.SaveConfig(true, cfgService); err != nil {
			return fmt.Errorf("failed to register notebook: %w", err)
		}

		fmt.Printf("Registered notebook '%s' at %s\n", nb.Config.Name, path)
		return nil
	},
}

func init() {
	notebookCmd.AddCommand(notebookRegisterCmd)
}
