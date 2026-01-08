package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

var notesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes",
	Long: `Searches notes by content or filename using DuckDB SQL.

The query searches both file names and content of markdown files.

Examples:
  # Search for notes containing "meeting"
  opennotes notes search "meeting"

  # Search with specific notebook
  opennotes notes search "todo" --notebook ~/notes`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		notes, err := nb.Notes.SearchNotes(context.Background(), args[0])
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", err)
		}

		if len(notes) == 0 {
			fmt.Printf("No notes found matching '%s'\n", args[0])
			return nil
		}

		fmt.Printf("Found %d note(s) matching '%s':\n\n", len(notes), args[0])
		return displayNoteList(notes)
	},
}

func init() {
	notesCmd.AddCommand(notesSearchCmd)
}
