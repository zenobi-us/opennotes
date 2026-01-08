package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var notesRemoveCmd = &cobra.Command{
	Use:     "remove <note>",
	Aliases: []string{"rm"},
	Short:   "Remove a note from the notebook",
	Long: `Removes a markdown note from the current notebook.

Prompts for confirmation unless --force is used. The .md extension
is optional when specifying the note name.

Examples:
  # Remove with confirmation
  opennotes notes remove my-note

  # Remove without confirmation
  opennotes notes remove my-note.md --force`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		noteName := args[0]
		force, _ := cmd.Flags().GetBool("force")

		// Ensure .md extension
		if !strings.HasSuffix(noteName, ".md") {
			noteName += ".md"
		}

		// Build full path
		notePath := filepath.Join(nb.Config.Root, noteName)

		// Check if file exists
		if _, err := os.Stat(notePath); os.IsNotExist(err) {
			return fmt.Errorf("note not found: %s", notePath)
		}

		// Confirm deletion unless --force is used
		if !force {
			fmt.Printf("Remove note '%s'? [y/N]: ", noteName)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Remove the file
		if err := os.Remove(notePath); err != nil {
			return fmt.Errorf("failed to remove note: %w", err)
		}

		fmt.Printf("Removed note: %s\n", notePath)
		return nil
	},
}

func init() {
	notesRemoveCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	notesCmd.AddCommand(notesRemoveCmd)
}
