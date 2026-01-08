package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notebookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notebooks",
	Long:  `Lists all registered notebooks and notebooks found in ancestor directories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		notebooks, err := notebookService.List("")
		if err != nil {
			return err
		}

		if len(notebooks) == 0 {
			fmt.Println("No notebooks found.")
			fmt.Println("")
			fmt.Println("Create one with:")
			fmt.Println("  opennotes notebook create --name \"My Notebook\"")
			return nil
		}

		return displayNotebookList(notebooks)
	},
}

func init() {
	notebookCmd.AddCommand(notebookListCmd)
}

func displayNotebookList(notebooks []*services.Notebook) error {
	output, err := services.TuiRender(services.Templates.NotebookList, map[string]any{
		"Notebooks": notebooks,
	})
	if err != nil {
		// Fallback to simple output
		fmt.Printf("Found %d notebook(s):\n\n", len(notebooks))
		for _, nb := range notebooks {
			fmt.Printf("  %s\n", nb.Config.Name)
			fmt.Printf("    Path: %s\n", nb.Config.Path)
			fmt.Printf("    Root: %s\n", nb.Config.Root)
			if len(nb.Config.Contexts) > 0 {
				fmt.Printf("    Contexts: %v\n", nb.Config.Contexts)
			}
			fmt.Println()
		}
		return nil
	}

	fmt.Print(output)
	return nil
}
