package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
)

var notesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all notes in the notebook",
	Long:  `Lists all markdown notes in the current notebook.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		notes, err := nb.Notes.SearchNotes(context.Background(), "")
		if err != nil {
			return fmt.Errorf("failed to list notes: %w", err)
		}

		return displayNoteList(notes)
	},
}

func init() {
	notesCmd.AddCommand(notesListCmd)
}

func displayNoteList(notes []services.Note) error {
	output, err := services.TuiRender(services.Templates.NoteList, map[string]any{
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
func requireNotebook(cmd *cobra.Command) (*services.Notebook, error) {
	// Check --notebook flag first
	notebookPath, _ := cmd.Flags().GetString("notebook")

	if notebookPath != "" {
		return notebookService.Open(notebookPath)
	}

	// Try to infer from context
	nb, err := notebookService.Infer("")
	if err != nil {
		return nil, err
	}

	if nb == nil {
		return nil, fmt.Errorf("no notebook found. Create one with: opennotes notebook create --name \"My Notebook\"")
	}

	return nb, nil
}
