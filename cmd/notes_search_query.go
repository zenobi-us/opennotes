package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
)

const (
	legacyQueryDeprecationSince         = "TODO(vX.Y.Z)"
	legacyQueryDeprecationRemovalTarget = "TODO(vX.(Y+1).0)"
	legacyQueryDeprecationTrackingEpic  = ".memory/epic-9b7c2a4d-unified-search-dsl-deprecation.md"
)

// @deprecated Legacy boolean flag query path (--and/--or/--not).
// Since: TODO(vX.Y.Z).
// Removal target: TODO(vX.(Y+1).0).
// Tracking: .memory/epic-9b7c2a4d-unified-search-dsl-deprecation.md
var notesSearchQueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Search notes with boolean AND/OR/NOT operators",
	Long: `Search notes using structured boolean queries with AND/OR/NOT operators.

BOOLEAN OPERATORS:
  --and field=value    All AND conditions must match (intersection)
  --or field=value     Any OR condition can match (union)  
  --not field=value    Excludes notes matching this condition

OPERATOR PRECEDENCE:
  1. AND conditions evaluated first
  2. OR conditions combined
  3. NOT conditions applied as exclusions

SUPPORTED FIELDS:

  Metadata fields (data.*):
    data.tag        - Note tags           --and data.tag=workflow
    data.tags       - Note tags (alias)   --and data.tags=meeting
    data.status     - Note status         --and data.status=active
    data.priority   - Priority level      --and data.priority=high
    data.assignee   - Assigned person     --and data.assignee=alice
    data.author     - Note author         --and data.author=bob
    data.type       - Note type           --and data.type=epic
    data.category   - Category            --and data.category=docs
    data.project    - Project name        --and data.project=alpha
    data.sprint     - Sprint identifier   --and data.sprint=s23

  Path and title:
    path            - File path (globs)   --and path=projects/*.md
    title           - Note title          --and title=Meeting

  Link queries (DAG):
    links-to        - Notes linking TO target    --and links-to=epics/*.md
    linked-by       - Notes linked FROM source   --and linked-by=plan.md

GLOB PATTERNS:
  *      Any characters (single level)   docs/*.md
  **     Any path depth                  **/*.md
  ?      Single character                task?.md

EXAMPLES:

  Basic filtering:
    jot notes search query --and data.tag=workflow
    jot notes search query --and data.tag=workflow --and data.status=active
    jot notes search query --or data.priority=high --or data.priority=critical
    jot notes search query --and data.tag=epic --not data.status=archived

  Path filtering:
    jot notes search query --and path=projects/*
    jot notes search query --and path=**/*.md --not path=archive/*

  Link queries:
    # Find notes that link TO architecture.md
    jot notes search query --and links-to=docs/architecture.md

    # Find notes that planning.md links TO
    jot notes search query --and linked-by=planning/q1.md

    # Find epics linking to any task
    jot notes search query --and data.tag=epic --and links-to=tasks/**/*.md

  Complex queries:
    jot notes search query \
      --and data.tag=workflow \
      --and data.status=active \
      --or data.priority=high \
      --not data.assignee=bob

SECURITY:
  - Only whitelisted fields can be queried
  - Values limited to 1000 characters

DOCUMENTATION:
  📖 Full reference: docs/commands/notes-search.md`,
	RunE: notesSearchQueryRunE,
}

func init() {
	notesSearchCmd.AddCommand(notesSearchQueryCmd)

	notesSearchQueryCmd.Flags().StringArray("and", []string{}, "AND condition (field=value) - all must match")
	notesSearchQueryCmd.Flags().StringArray("or", []string{}, "OR condition (field=value) - any must match")
	notesSearchQueryCmd.Flags().StringArray("not", []string{}, "NOT condition (field=value) - excludes matches")
}

// @deprecated Legacy execution path for `jot notes search query --and/--or/--not`.
// Since: TODO(vX.Y.Z).
// Removal target: TODO(vX.(Y+1).0).
// Tracking: .memory/epic-9b7c2a4d-unified-search-dsl-deprecation.md
func notesSearchQueryRunE(cmd *cobra.Command, args []string) error {
	// Parse flags
	andFlags, err := cmd.Flags().GetStringArray("and")
	if err != nil {
		return fmt.Errorf("failed to parse --and flags: %w", err)
	}

	orFlags, err := cmd.Flags().GetStringArray("or")
	if err != nil {
		return fmt.Errorf("failed to parse --or flags: %w", err)
	}

	notFlags, err := cmd.Flags().GetStringArray("not")
	if err != nil {
		return fmt.Errorf("failed to parse --not flags: %w", err)
	}

	// Check that at least one condition is provided
	if len(andFlags) == 0 && len(orFlags) == 0 && len(notFlags) == 0 {
		return fmt.Errorf("at least one condition is required (use --and, --or, or --not)")
	}

	// Get search service to parse conditions
	searchService := services.NewSearchService()

	// Parse and validate conditions
	conditions, err := searchService.ParseConditions(andFlags, orFlags, notFlags)
	if err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	// Get notebook
	nb, err := requireNotebook(cmd)
	if err != nil {
		return err
	}

	// Execute query
	notes, err := nb.Notes.SearchWithConditions(context.Background(), conditions)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Display results
	if len(notes) == 0 {
		fmt.Println("No notes found matching the specified conditions.")
		return nil
	}

	fmt.Printf("Found %d note(s) matching conditions:\n\n", len(notes))
	return displayNoteList(notes)
}
