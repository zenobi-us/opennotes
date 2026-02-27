package cmd

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/search"
	"github.com/zenobi-us/jot/internal/search/parser"
	"github.com/zenobi-us/jot/internal/services"
)

var notesSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes with text and DSL filter/directive syntax",
	Long: `Search notes using multiple methods: text search, boolean queries, or DSL with pipe syntax.

SEARCH METHODS:

  1. Default Fieldless Search: title-only DSL normalization
     jot notes search "meeting"        # normalized to title:meeting

  2. DSL Filter + Directives Syntax
     jot notes search "tag:work | sort:modified:desc limit:10"

FIELDLESS SEARCH EXAMPLES (title-only):
  jot notes search "meeting"              # Search title for "meeting"
  jot notes search "todo" --notebook ~/n  # Search title in specific notebook
  jot notes search                        # List all notes
  jot notes search "body:meeting"         # Explicit body search

DSL PIPE SYNTAX EXAMPLES:
  jot notes search "tag:work | sort:modified:desc"
  jot notes search "status:todo | sort:created:asc limit:20"
  jot notes search "| sort:title:asc"     # All notes, sorted by title

  DSL Filter:
  - tag:<value>      Notes with specific tag
  - status:<value>   Notes with status field
  - title:<text>     Search in title
  - path:<prefix>    Notes in path prefix
  - created:>date    Created after date
  - modified:<date   Modified before date

  Directives (after |):
  - sort:<field>:<dir>  Sort by field (modified, created, title, path)
                        Direction: asc or desc (default: asc)
  - limit:<n>           Return at most n results
  - offset:<n>          Skip first n results (for pagination)

DSL FILTER EXAMPLES:
  jot notes search "tag:workflow status:active"
  jot notes search "tag:epic NOT status:archived"
  jot notes search "priority:high OR priority:critical"

DOCUMENTATION:
  📖 Command Reference: docs/commands/notes-search.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var searchTerm string
		if len(args) > 0 {
			searchTerm = args[0]
		}

		format, _ := cmd.Flags().GetString("format")
		if err := validateOutputFormat(format, "list", "json"); err != nil {
			return err
		}

		nb, err := requireNotebook(cmd)
		if err != nil {
			return err
		}

		if isDSLStyleQuery(searchTerm) {
			return runSearchWithPipeSyntax(cmd.Context(), nb, searchTerm, format)
		}

		normalized := normalizeFieldlessToTitleQuery(searchTerm)
		if normalized != "" {
			return runSearchWithPipeSyntax(cmd.Context(), nb, normalized, format)
		}

		notes, err := nb.Notes.SearchNotes(context.Background(), searchTerm)
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", err)
		}

		if format == "json" {
			return renderJSON(notes)
		}

		if len(notes) == 0 {
			if searchTerm != "" {
				fmt.Printf("No notes found matching '%s'\n", searchTerm)
			} else {
				fmt.Println("No notes found")
			}
			return nil
		}

		if searchTerm != "" {
			fmt.Printf("Found %d note(s) matching '%s':\n\n", len(notes), searchTerm)
		} else {
			fmt.Printf("Found %d note(s):\n\n", len(notes))
		}

		return displayNoteList(notes)
	},
}

func init() {
	notesSearchCmd.Flags().String("format", "list", "Output format: list or json")
	notesCmd.AddCommand(notesSearchCmd)
}

// isDSLStyleQuery detects when a search query is intended for DSL parsing.
// Current heuristic: field/directive style tokens (:) or explicit pipe directives (|).
func isDSLStyleQuery(query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return false
	}
	return strings.Contains(q, ":") || strings.Contains(q, "|")
}

// normalizeFieldlessToTitleQuery converts non-DSL user input into a title-only DSL query.
// Single-token input becomes title:token, and multi-word input becomes title:"phrase".
func normalizeFieldlessToTitleQuery(query string) string {
	q := strings.TrimSpace(query)
	if q == "" {
		return ""
	}

	if strings.ContainsAny(q, " \t\n\"") {
		return "title:" + strconv.Quote(q)
	}

	return "title:" + q
}

// runSearchWithPipeSyntax executes a search using pipe syntax (filter | directives).
// This allows DSL-based search with sort, limit, and other options.
// Example: "tag:work | sort:modified:desc limit:10"
func runSearchWithPipeSyntax(ctx context.Context, nb *services.Notebook, query string, format string) error {
	// Split query into filter and directives
	filterPart, directivesPart := services.SplitViewQuery(query)

	// Parse directives
	directives, err := services.ParseDirectives(directivesPart)
	if err != nil {
		return fmt.Errorf("failed to parse directives: %w", err)
	}

	// Build FindOpts from directives
	opts := search.FindOpts{
		Limit:  directives.Limit,
		Offset: directives.Offset,
	}

	// Parse filter DSL if present
	if filterPart != "" {
		p := parser.New()
		parsedQuery, err := p.Parse(filterPart)
		if err != nil {
			return fmt.Errorf("failed to parse filter: %w", err)
		}
		opts.Query = parsedQuery
		opts.RawQuery = filterPart
	}

	// Set sort from directives
	if directives.SortField != "" {
		opts.Sort = directiveToSortSpec(directives.SortField, directives.SortDirection)
	}

	// Execute search using the new method
	notes, err := nb.Notes.SearchWithFindOpts(ctx, opts)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if format == "json" {
		return renderJSON(notes)
	}

	if len(notes) == 0 {
		fmt.Printf("No notes found matching query\n")
		return nil
	}

	fmt.Printf("Found %d note(s):\n\n", len(notes))
	return displayNoteList(notes)
}

// directiveToSortSpec converts directive sort parameters to search.SortSpec
func directiveToSortSpec(field, direction string) search.SortSpec {
	var sortDirection search.SortDirection
	if direction == "desc" {
		sortDirection = search.SortDesc
	} else {
		sortDirection = search.SortAsc
	}

	var sortField search.SortField
	switch field {
	case "modified":
		sortField = search.SortByModified
	case "created":
		sortField = search.SortByCreated
	case "title":
		sortField = search.SortByTitle
	case "path":
		sortField = search.SortByPath
	default:
		sortField = search.SortByRelevance
	}

	return search.SortSpec{Field: sortField, Direction: sortDirection}
}
