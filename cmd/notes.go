package cmd

import (
	"github.com/spf13/cobra"
)

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "Manage notes",
	Long: `Commands for managing notes - list, search, add, and remove notes.

Notes are markdown files stored in the notebook's notes directory.
The notebook is automatically discovered from the current directory,
or can be specified with the --notebook flag.

Examples:
  # List all notes
  opennotes notes list

  # Add a new note with title
  opennotes notes add --title "Meeting Notes"

  # Search notes by content
  opennotes notes search "project deadline"

  # Remove a note
  opennotes notes remove my-note.md`,
}

func init() {
	rootCmd.AddCommand(notesCmd)
}
