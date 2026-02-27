package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zenobi-us/jot/internal/services"
)

var notesSearchSemanticCmd = &cobra.Command{
	Use:   "semantic [query]",
	Short: "Search notes using semantic, keyword, or hybrid retrieval modes",
	Long: `Search notes through the semantic search command surface.

Modes:
  --mode hybrid   Default. Merges keyword and semantic retrieval
  --mode keyword  Keyword-only retrieval (for diagnostics)
  --mode semantic Semantic-only retrieval

Optional condition flags are supported:
  --and/--or/--not field=value

Examples:
  jot notes search semantic "meeting notes"
  jot notes search semantic "workflow" --mode keyword --and data.status=active
  jot notes search semantic "project triage" --mode hybrid --not data.status=archived
  jot notes search semantic "architecture" --explain`,
	Args: cobra.MaximumNArgs(1),
	RunE: notesSearchSemanticRunE,
}

func init() {
	notesSearchCmd.AddCommand(notesSearchSemanticCmd)

	notesSearchSemanticCmd.Flags().String("mode", "hybrid", "Retrieval mode: hybrid|keyword|semantic")
	notesSearchSemanticCmd.Flags().StringArray("and", []string{}, "AND condition (field=value) - all must match")
	notesSearchSemanticCmd.Flags().StringArray("or", []string{}, "OR condition (field=value) - any can match")
	notesSearchSemanticCmd.Flags().StringArray("not", []string{}, "NOT condition (field=value) - excludes matches")
	notesSearchSemanticCmd.Flags().Int("top-k", 100, "Maximum candidates per retrieval source before merge")
	notesSearchSemanticCmd.Flags().Bool("explain", false, "Show per-result match label and why snippet")
}

func notesSearchSemanticRunE(cmd *cobra.Command, args []string) error {
	var query string
	if len(args) > 0 {
		query = args[0]
	}

	modeRaw, err := cmd.Flags().GetString("mode")
	if err != nil {
		return fmt.Errorf("failed to parse --mode: %w", err)
	}

	mode, err := services.ParseRetrievalMode(modeRaw)
	if err != nil {
		return err
	}

	explain, err := cmd.Flags().GetBool("explain")
	if err != nil {
		return fmt.Errorf("failed to parse --explain: %w", err)
	}

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

	topK, err := cmd.Flags().GetInt("top-k")
	if err != nil {
		return fmt.Errorf("failed to parse --top-k: %w", err)
	}

	searchService := services.NewSearchService()
	conditions, err := searchService.ParseConditions(andFlags, orFlags, notFlags)
	if err != nil {
		return fmt.Errorf("invalid condition: %w", err)
	}

	nb, err := requireNotebook(cmd)
	if err != nil {
		return err
	}

	hits, meta, err := nb.Notes.SearchSemanticDetailed(context.Background(), query, conditions, mode, topK)
	if err != nil {
		if errors.Is(err, services.ErrSemanticUnavailable) {
			fmt.Println("Semantic backend unavailable. Try --mode keyword or --mode hybrid.")
			return nil
		}
		return fmt.Errorf("semantic search failed: %w", err)
	}

	if meta.SemanticFallback {
		fmt.Println("Warning: semantic backend unavailable, showing keyword-mode results.")
	}

	if len(hits) == 0 {
		switch mode {
		case services.RetrievalModeKeyword:
			fmt.Println("No keyword-mode results. Try --mode hybrid or --mode semantic.")
		case services.RetrievalModeSemantic:
			fmt.Println("No semantic-mode results. Try --mode hybrid or --mode keyword.")
		default:
			if query != "" {
				fmt.Printf("No notes found matching '%s'\n", query)
			} else {
				fmt.Println("No notes found")
			}
		}
		return nil
	}

	if explain {
		fmt.Printf("Found %d note(s) using %s mode (explain):\n\n", len(hits), mode)
		return displaySemanticSearchHits(hits, true)
	}

	notes := make([]services.Note, len(hits))
	for i, hit := range hits {
		notes[i] = hit.Note
	}

	fmt.Printf("Found %d note(s) using %s mode:\n\n", len(notes), mode)
	return displayNoteList(notes)
}

func displaySemanticSearchHits(hits []services.SemanticSearchHit, explain bool) error {
	output, err := services.TuiRender("note-search-semantic", map[string]any{
		"Hits":    hits,
		"Explain": explain,
	})
	if err != nil {
		for _, hit := range hits {
			fmt.Printf("- [%s] %s (%s)\n", hit.Note.DisplayName(), hit.Note.File.Relative, hit.MatchType)
			if explain {
				reason := hit.Explain
				if reason == "" {
					reason = "No snippet available"
				}
				fmt.Printf("  Why: %s\n", reason)
			}
		}
		return nil
	}

	fmt.Print(output)
	return nil
}
