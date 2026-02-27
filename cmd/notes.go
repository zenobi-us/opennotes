package cmd

import (
	"github.com/spf13/cobra"
)

var notesCmd = &cobra.Command{
	Use:   "notes",
	Short: "Manage notes - list, search, add, remove",
	Long: `Commands for managing notes - list, search, add, and remove notes.

Notes are markdown files stored in the notebook's notes directory.
The notebook is automatically discovered from the current directory,
or can be specified with the --notebook flag.

POWER USER FEATURES:
  🔍 Advanced DSL Filters: jot notes search "path:projects/*.md"
  🤖 JSON Output for Automation: Results automatically JSON-formatted for jq and tool integration
  💾 Large Notebook Support: Efficiently search thousands of notes in seconds

DOCUMENTATION:
  📚 Search Guide: https://github.com/zenobi-us/jot/blob/main/docs/commands/notes-search.md

Examples:
  # List all notes
  jot notes list

  # Add a new note with title
  jot notes add --title "Meeting Notes"

  # Search notes by content
  jot notes search "project deadline"

  # Filter with DSL operators
  jot notes search "path:**/*.md NOT path:archive/*"

  # Remove a note
  jot notes remove my-note.md`,
}

func init() {
	rootCmd.AddCommand(notesCmd)
}
