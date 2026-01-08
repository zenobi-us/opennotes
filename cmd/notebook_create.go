package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notebookCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new notebook",
	Long:  `Creates a new notebook in the specified directory (or current directory).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		path, _ := cmd.Flags().GetString("path")
		global, _ := cmd.Flags().GetBool("global")

		nb, err := notebookService.Create(name, path, global)
		if err != nil {
			return fmt.Errorf("failed to create notebook: %w", err)
		}

		fmt.Printf("Created notebook '%s'\n", nb.Config.Name)
		fmt.Printf("  Config: %s\n", nb.Config.Path)
		fmt.Printf("  Notes:  %s\n", nb.Config.Root)

		if global {
			fmt.Println("  Registered globally")
		}

		return nil
	},
}

func init() {
	notebookCreateCmd.Flags().StringP("name", "n", "", "Notebook name (required)")
	notebookCreateCmd.Flags().StringP("path", "p", "", "Notebook path (default: current directory)")
	notebookCreateCmd.Flags().BoolP("global", "g", false, "Register globally")
	notebookCreateCmd.MarkFlagRequired("name")
	notebookCmd.AddCommand(notebookCreateCmd)
}

func displayNotebookInfo(nb *services.Notebook) error {
	output, err := services.TuiRender(services.Templates.NotebookInfo, nb)
	if err != nil {
		// Fallback to simple output
		fmt.Printf("Notebook: %s\n", nb.Config.Name)
		fmt.Printf("  Config: %s\n", nb.Config.Path)
		fmt.Printf("  Root:   %s\n", nb.Config.Root)

		if len(nb.Config.Contexts) > 0 {
			fmt.Printf("  Contexts:\n")
			for _, ctx := range nb.Config.Contexts {
				fmt.Printf("    - %s\n", ctx)
			}
		}

		if len(nb.Config.Groups) > 0 {
			fmt.Printf("  Groups:\n")
			for _, g := range nb.Config.Groups {
				fmt.Printf("    - %s (%v)\n", g.Name, g.Globs)
			}
		}
		return nil
	}

	fmt.Print(output)
	return nil
}
