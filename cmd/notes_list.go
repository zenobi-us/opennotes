package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
)

var notesListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all notes in the notebook",
	Long: `Lists all markdown notes in the current notebook.

Shows all .md files in the notebook's notes directory with metadata.

Examples:
  # List notes in current notebook
  jot notes list

  # List notes from specific notebook
  jot notes list --notebook /path/to/notebook`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		notes, err := nb.Notes.SearchNotes(context.Background(), "")
		if err != nil {
			// DuckDB returns an error when the glob pattern matches no files
			// Treat this as an empty notebook
			return displayNoteList([]services.Note{})
		}

		return displayNoteList(notes)
	},
}

func init() {
	notesCmd.AddCommand(notesListCmd)
}

func displayNoteList(notes []services.Note) error {
	output, err := services.TuiRender("note-list", map[string]any{
		"Notes": notes,
	})
	if err != nil {
		// Fallback to simple output
		if len(notes) == 0 {
			fmt.Println("No notes found.")
			return nil
		}
		fmt.Printf("Found %d note(s):\n\n", len(notes))
		for _, note := range notes {
			fmt.Printf("  %s\n", note.File.Relative)
		}
		return nil
	}

	fmt.Print(output)
	return nil
}

// requireNotebook is a helper to get the current notebook or return an error.
// Resolution order (first wins):
// 1. JOT_NOTEBOOK envvar
// 2. --notebook flag
// 3. .jot.json in current directory
// 4. context match (registered notebooks)
// 5. ancestor search
func requireNotebook(cmd *cobra.Command) (*services.Notebook, error) {
	// Step 1: Check JOT_NOTEBOOK envvar
	if envNotebook := os.Getenv("JOT_NOTEBOOK"); envNotebook != "" {
		return notebookService.Open(envNotebook)
	}

	// Step 2: Check --notebook flag
	notebookPath, _ := cmd.Flags().GetString("notebook")
	if notebookPath != "" {
		return notebookService.Open(notebookPath)
	}

	// Step 3-5: Use Infer() for auto-detection
	nb, err := notebookService.Infer("")
	if err != nil {
		return nil, err
	}

	if nb == nil {
		return nil, fmt.Errorf("no notebook found. Set JOT_NOTEBOOK, use --notebook flag, or create one with: jot notebook create --name \"My Notebook\"")
	}

	return nb, nil
}
