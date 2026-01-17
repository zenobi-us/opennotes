package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/opennotes/internal/services"
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
  opennotes notes search "todo" --notebook ~/notes

  # Execute custom SQL query
  opennotes notes search --sql "SELECT * FROM markdown LIMIT 10"`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get --sql flag if provided
		sqlQuery, _ := cmd.Flags().GetString("sql")

		// If --sql flag is provided, run SQL mode
		if sqlQuery != "" {
			nb, err := requireNotebook(cmd)
			if err != nil {
				return err
			}

			// Execute the SQL query using NoteService
			results, err := nb.Notes.ExecuteSQLSafe(context.Background(), sqlQuery)
			if err != nil {
				return fmt.Errorf("SQL query failed: %w", err)
			}

			// Create display service and render results
			display, err := services.NewDisplay()
			if err != nil {
				return fmt.Errorf("failed to create display: %w", err)
			}

			return display.RenderSQLResults(results)
		}

		// Normal search mode - require a query argument
		if len(args) == 0 {
			return fmt.Errorf("query argument required (or use --sql flag)")
		}

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

	// Add --sql flag for custom SQL queries
	notesSearchCmd.Flags().String(
		"sql",
		"",
		"Execute custom SQL query against notes (bypasses normal search)",
	)
}
