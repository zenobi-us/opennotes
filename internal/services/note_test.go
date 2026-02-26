package services_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/jot/internal/services"
	"github.com/zenobi-us/jot/internal/testutil"
)

func TestNoteService_SearchNotes_NoNotebookSelected(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")
	svc := services.NewNoteService(cfg, nil, "")

	notes, err := svc.SearchNotes(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, notes)
	assert.Contains(t, err.Error(), "no notebook selected")
}

func TestNoteService_SearchNotes_FindsAllNotes(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create a test notebook with notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# Note 1\n\nFirst note content.")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Note 2\n\nSecond note content.")
	testutil.CreateTestNote(t, notebookDir, "note3.md", "# Note 3\n\nThird note content.")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	assert.Len(t, notes, 3)
}

func TestNoteService_SearchNotes_FiltersByQuery(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create a test notebook with notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "apple.md", "# Apple\n\nThis is about apples.")
	testutil.CreateTestNote(t, notebookDir, "banana.md", "# Banana\n\nThis is about bananas.")
	testutil.CreateTestNote(t, notebookDir, "cherry.md", "# Cherry\n\nThis is about cherries.")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Search for "apple"
	notes, err := svc.SearchNotes(ctx, "apple")
	require.NoError(t, err)

	assert.Len(t, notes, 1)
	assert.Contains(t, notes[0].File.Filepath, "apple.md")
}

func TestNoteService_SearchNotes_FiltersByQueryCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "mixed.md", "# UPPERCASE content\n\nSome text.")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Search with lowercase should match uppercase content
	notes, err := svc.SearchNotes(ctx, "uppercase")
	require.NoError(t, err)

	assert.Len(t, notes, 1)
}

func TestNoteService_SearchNotes_FiltersByFilepath(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "project-ideas.md", "# Ideas\n\nSome ideas.")
	testutil.CreateTestNote(t, notebookDir, "daily-notes.md", "# Daily\n\nDaily notes.")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Search by filename pattern
	notes, err := svc.SearchNotes(ctx, "project")
	require.NoError(t, err)

	assert.Len(t, notes, 1)
	assert.Contains(t, notes[0].File.Filepath, "project-ideas.md")
}

func TestNoteService_SearchNotes_EmptyNotebook(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create empty notebook (no notes)
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "empty-notebook")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Empty notebook should return empty list without error
	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)
	assert.Empty(t, notes)
}

func TestNoteService_SearchNotes_ExtractsMetadata(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create note with frontmatter metadata
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "with-meta.md",
		map[string]interface{}{
			"title": "Test Title",
			"tags":  "[tag1, tag2]",
		},
		"# Test Note\n\nThis is content with frontmatter.",
	)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	// Metadata should be populated (DuckDB returns it as a map)
	assert.NotNil(t, notes[0].Metadata)
}

func TestNoteService_SearchNotes_SetsRelativePath(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "my-note.md", "# My Note\n\nContent here.")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Equal(t, "notes/my-note.md", notes[0].File.Relative)
}

func TestNoteService_Count_NoNotebookSelected(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")
	svc := services.NewNoteService(cfg, nil, "")

	// Count returns 0 when no notebook is selected (not an error)
	count, err := svc.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestNoteService_Count_ReturnsCorrectCount(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# Note 1")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Note 2")
	testutil.CreateTestNote(t, notebookDir, "note3.md", "# Note 3")
	testutil.CreateTestNote(t, notebookDir, "note4.md", "# Note 4")
	testutil.CreateTestNote(t, notebookDir, "note5.md", "# Note 5")

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	count, err := svc.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestNoteService_Count_EmptyNotebook(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "empty-notebook")

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Empty notebook should return 0 without error
	count, err := svc.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestNoteService_SearchNotes_MultipleQueryMatches(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# First Note\n\nContains the word golang.")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Second Note\n\nAlso mentions golang here.")
	testutil.CreateTestNote(t, notebookDir, "note3.md", "# Third Note\n\nNo match in this one.")

	idx := testutil.CreateTestIndex(t, notebookDir)

	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "golang")
	require.NoError(t, err)

	assert.Len(t, notes, 2)
}

func TestNoteService_SearchNotes_ContentHasText(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	expectedContent := "# My Note Title\n\nThis is the body content."
	testutil.CreateTestNote(t, notebookDir, "note.md", expectedContent)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Contains(t, notes[0].Content, "My Note Title")
	assert.Contains(t, notes[0].Content, "body content")
}

func TestNewNoteService(t *testing.T) {
	t.Cleanup(func() {
	})

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")

	svc := services.NewNoteService(cfg, nil, "/test/notebook/path")

	assert.NotNil(t, svc)
}

func TestNoteService_SearchNotes_DisplayNameWithTitle(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create note with title in frontmatter
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "my-file.md",
		map[string]interface{}{
			"title": "My Custom Title",
		},
		"# Note\n\nContent here.",
	)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Equal(t, "My Custom Title", notes[0].DisplayName())
}

func TestNoteService_SearchNotes_DisplayNameSlugifyFilename(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create note without title - should slugify filename
	testutil.CreateTestNote(t, notebookDir, "Hello World.md", "# Hello\n\nContent here.")

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Equal(t, "hello-world", notes[0].DisplayName())
}

func TestNoteService_SearchNotes_DisplayNameMultipleNotes(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create notes with mixed title/no title
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note1.md",
		map[string]interface{}{"title": "First Note"},
		"Content",
	)
	testutil.CreateTestNote(t, notebookDir, "note2.md", "Content")
	testutil.CreateTestNoteWithFrontmatter(t, notebookDir, "note3.md",
		map[string]interface{}{"title": "Third Note"},
		"Content",
	)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 3)

	// Verify display names
	displayNames := make([]string, len(notes))
	for i, note := range notes {
		displayNames[i] = note.DisplayName()
	}

	// Check that we have expected display names (order may vary)
	assert.Contains(t, displayNames, "First Note")
	assert.Contains(t, displayNames, "note2")
	assert.Contains(t, displayNames, "Third Note")
}

// === SearchNotes Edge Case Tests ===

func TestNoteService_SearchNotes_ComplexQueries(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create test notebook with diverse content
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "complex-search-test")

	// Create notes with varied content for complex searching
	testutil.CreateTestNote(t, notebookDir, "golang-tips.md", "# Golang Tips\n\nUseful golang programming patterns.")
	testutil.CreateTestNote(t, notebookDir, "javascript-tricks.md", "# JavaScript Tricks\n\nSome javascript and golang comparisons.")
	testutil.CreateTestNote(t, notebookDir, "python-guide.md", "# Python Guide\n\nPython programming fundamentals.")
	testutil.CreateTestNote(t, notebookDir, "mixed-content.md", "# Mixed Content\n\nThis mentions golang, python, and javascript.")

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	tests := []struct {
		name          string
		query         string
		expectedCount int
		description   string
	}{
		{
			"case_insensitive_search",
			"GOLANG",
			3, // golang-tips.md, javascript-tricks.md, mixed-content.md
			"Search should be case-insensitive",
		},
		{
			"partial_word_match",
			"java",
			2, // javascript-tricks.md, mixed-content.md
			"Should find partial word matches",
		},
		{
			"filename_search",
			"tips",
			1, // golang-tips.md
			"Should search in filename as well",
		},
		{
			"common_word_search",
			"programming",
			2, // golang-tips.md, python-guide.md
			"Should find notes with common programming terms",
		},
		{
			"no_matches",
			"nonexistent",
			0,
			"Should return no results for non-matching query",
		},
		{
			"empty_query",
			"",
			4,
			"Empty query should return all notes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notes, err := svc.SearchNotes(ctx, tt.query)
			require.NoError(t, err, tt.description)
			assert.Len(t, notes, tt.expectedCount,
				"Expected %d notes for query '%s', got %d",
				tt.expectedCount, tt.query, len(notes))
		})
	}
}

func TestNoteService_SearchNotes_SpecialCharacters(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "special-chars-test")

	// Create notes with special characters
	testutil.CreateTestNote(t, notebookDir, "unicode-test.md", "# Unicode Test\n\nCafé, naïve, résumé")
	testutil.CreateTestNote(t, notebookDir, "symbols.md", "# Symbols\n\nC++ programming, @mentions, #hashtags")
	testutil.CreateTestNote(t, notebookDir, "quotes.md", "# Quotes\n\n\"Double quotes\" and 'single quotes'")
	testutil.CreateTestNote(t, notebookDir, "math.md", "# Math\n\n2 + 2 = 4, x² + y² = z²")

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	tests := []struct {
		name          string
		query         string
		expectedCount int
	}{
		{
			"unicode_search",
			"café",
			1,
		},
		{
			"plus_signs",
			"C++",
			1,
		},
		{
			"at_symbol",
			"@mentions",
			1,
		},
		{
			"hashtag",
			"#hashtags",
			1,
		},
		{
			"quotes_double",
			"\"Double quotes\"",
			1,
		},
		{
			"quotes_single",
			"'single quotes'",
			1,
		},
		{
			"math_equation",
			"2 + 2",
			1,
		},
		{
			"superscript_unicode",
			"x²",
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notes, err := svc.SearchNotes(ctx, tt.query)
			require.NoError(t, err)
			assert.Len(t, notes, tt.expectedCount,
				"Expected %d notes for query '%s'", tt.expectedCount, tt.query)
		})
	}
}

func TestNoteService_SearchNotes_LargeResultSets(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "large-test")

	// Create many notes with shared content
	commonWord := "shared"
	for i := 1; i <= 25; i++ {
		content := fmt.Sprintf("# Note %d\n\nThis note contains the %s keyword and unique content %d.",
			i, commonWord, i)
		testutil.CreateTestNote(t, notebookDir, fmt.Sprintf("note%03d.md", i), content)
	}

	// Create some notes without the shared word
	for i := 1; i <= 5; i++ {
		content := fmt.Sprintf("# Different %d\n\nThis note has different content without the keyword.", i)
		testutil.CreateTestNote(t, notebookDir, fmt.Sprintf("different%03d.md", i), content)
	}

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Test large result set
	notes, err := svc.SearchNotes(ctx, commonWord)
	require.NoError(t, err)
	assert.Len(t, notes, 25, "Should find all notes with shared keyword")

	// Test all notes (empty query)
	allNotes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)
	assert.Len(t, allNotes, 30, "Should find all 30 notes")

	// Verify note structure is correct for all notes
	for _, note := range notes {
		assert.NotEmpty(t, note.Content, "Note should have content")
		assert.NotEmpty(t, note.File.Filepath, "Note should have filepath")
		assert.NotEmpty(t, note.File.Relative, "Note should have relative path")
		assert.Contains(t, note.Content, commonWord, "Note should contain search term")
	}
}

func TestNoteService_SearchNotes_FrontmatterEdgeCases(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "frontmatter-test")

	// Note with no frontmatter
	testutil.CreateTestNote(t, notebookDir, "no-frontmatter.md", "# No Frontmatter\n\nJust content here.")

	// Note with complex frontmatter
	complexFrontmatter := `---
title: "Complex Note"
tags: ["test", "complex", "frontmatter"]
metadata:
  author: "Test Author"
  date: 2024-01-15
  nested:
    value: 42
    enabled: true
categories: null
---

# Complex Note

Content with complex frontmatter.`
	testutil.CreateTestNote(t, notebookDir, "complex-frontmatter.md", complexFrontmatter)

	// Note with empty frontmatter
	emptyFrontmatter := `---
---

# Empty Frontmatter

Content with empty frontmatter.`
	testutil.CreateTestNote(t, notebookDir, "empty-frontmatter.md", emptyFrontmatter)

	// Note with malformed frontmatter (should still work)
	malformedFrontmatter := `---
title: Malformed
missing_colon_value
tags: [unclosed list
---

# Malformed

Content despite frontmatter issues.`
	testutil.CreateTestNote(t, notebookDir, "malformed-frontmatter.md", malformedFrontmatter)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Test that all notes are found regardless of frontmatter quality
	allNotes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)
	assert.Len(t, allNotes, 4, "Should find all notes regardless of frontmatter")

	// Test searching content works even with frontmatter issues
	contentSearch, err := svc.SearchNotes(ctx, "Content")
	require.NoError(t, err)
	assert.Len(t, contentSearch, 4, "Content search should work despite frontmatter variations")

	// Verify metadata is populated where possible
	for _, note := range allNotes {
		assert.NotNil(t, note.Metadata, "Metadata map should exist even if empty")

		// Check specific notes
		switch {
		case strings.Contains(note.File.Relative, "complex-frontmatter"):
			// Complex frontmatter should have some metadata
			assert.NotEmpty(t, note.Metadata, "Complex frontmatter should have metadata")
		case strings.Contains(note.File.Relative, "no-frontmatter"):
			// No frontmatter note might have empty or minimal metadata
			assert.NotNil(t, note.Metadata, "Even no-frontmatter notes should have metadata map")
		}
	}
}

func TestNoteService_SearchNotes_ErrorConditions(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Test with empty/non-existent notebook
	svc := services.NewNoteService(cfg, nil, "")

	notes, err := svc.SearchNotes(ctx, "test")
	assert.Error(t, err, "Should error when no notebook selected")
	assert.Nil(t, notes, "Notes should be nil on error")
	assert.Contains(t, err.Error(), "no notebook selected", "Error should mention no notebook")

	// Test with non-existent notebook path
	nonExistentPath := filepath.Join(tmpDir, "nonexistent-notebook")
	svc2 := services.NewNoteService(cfg, nil, nonExistentPath)

	// This might not error immediately since DuckDB might handle empty globs gracefully
	notes2, err := svc2.SearchNotes(ctx, "test")
	if err != nil {
		// If it errors, that's fine - means validation exists
		assert.Nil(t, notes2)
	} else {
		// If no error, should return empty result set
		assert.Empty(t, notes2, "Non-existent notebook should return empty results")
	}
}

// ============================================================================
// SearchWithConditions Tests
// ============================================================================

func TestNoteService_SearchWithConditions_SimpleAnd(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create test notebook with notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "workflow1.md", `---
tag: workflow
status: active
---
# Workflow 1
Active workflow note.
`)
	testutil.CreateTestNote(t, notebookDir, "workflow2.md", `---
tag: workflow
status: done
---
# Workflow 2
Completed workflow.
`)
	testutil.CreateTestNote(t, notebookDir, "meeting.md", `---
tag: meeting
status: active
---
# Meeting Notes
Team meeting.
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Single AND condition
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should find at least 2 workflow notes")

	// Verify all results have workflow tag
	for _, note := range results {
		tag, ok := note.Metadata["tag"]
		if ok {
			assert.Equal(t, "workflow", tag, "All results should have workflow tag")
		}
	}
}

func TestNoteService_SearchWithConditions_MultipleAnd(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "active-workflow.md", `---
tag: workflow
status: active
---
# Active Workflow
`)
	testutil.CreateTestNote(t, notebookDir, "done-workflow.md", `---
tag: workflow
status: done
---
# Done Workflow
`)
	testutil.CreateTestNote(t, notebookDir, "active-meeting.md", `---
tag: meeting
status: active
---
# Active Meeting
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Multiple AND conditions - both must match
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
		{Type: "and", Field: "data.status", Operator: "=", Value: "active"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "Should find active workflow")

	// Verify all results match both conditions
	for _, note := range results {
		tag := note.Metadata["tag"]
		status := note.Metadata["status"]
		assert.Equal(t, "workflow", tag)
		assert.Equal(t, "active", status)
	}
}

func TestNoteService_SearchWithConditions_OrConditions(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "high-priority.md", `---
priority: high
---
# High Priority
`)
	testutil.CreateTestNote(t, notebookDir, "critical-priority.md", `---
priority: critical
---
# Critical Priority
`)
	testutil.CreateTestNote(t, notebookDir, "low-priority.md", `---
priority: low
---
# Low Priority
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// OR conditions - any can match
	conditions := []services.QueryCondition{
		{Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should find high OR critical priority")

	// Verify no low priority notes
	for _, note := range results {
		priority := note.Metadata["priority"]
		assert.NotEqual(t, "low", priority, "Should not include low priority")
	}
}

func TestNoteService_SearchWithConditions_NotCondition(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "epic1.md", `---
tag: epic
status: active
---
# Epic 1
`)
	testutil.CreateTestNote(t, notebookDir, "epic2.md", `---
tag: epic
status: archived
---
# Epic 2
`)
	testutil.CreateTestNote(t, notebookDir, "epic3.md", `---
tag: epic
status: done
---
# Epic 3
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// NOT condition - exclude archived
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "epic"},
		{Type: "not", Field: "data.status", Operator: "=", Value: "archived"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should find non-archived epics")

	// Verify no archived notes
	for _, note := range results {
		status := note.Metadata["status"]
		assert.NotEqual(t, "archived", status, "Should not include archived notes")
	}
}

func TestNoteService_SearchWithConditions_PathGlob(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create flat structure (all notes in notes/ directory)
	testutil.CreateTestNote(t, notebookDir, "epic1.md", `---
title: Epic 1
---
# Epic 1
`)
	testutil.CreateTestNote(t, notebookDir, "epic2.md", `---
title: Epic 2
---
# Epic 2
`)
	testutil.CreateTestNote(t, notebookDir, "task1.md", `---
title: Task 1
---
# Task 1
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Path glob pattern matching "epic*.md" files
	conditions := []services.QueryCondition{
		{Type: "and", Field: "path", Operator: "=", Value: "epic*.md"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should find epic notes")

	// Verify all results match the pattern
	for _, note := range results {
		assert.Contains(t, note.File.Relative, "epic", "All results should contain 'epic'")
	}
}

func TestNoteService_SearchWithConditions_NoResults(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "meeting.md", `---
tag: meeting
---
# Meeting
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Search for non-existent tag
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "nonexistent"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err, "Should not error on no results")
	assert.Empty(t, results, "Should return empty results")
}

func TestNoteService_SearchWithConditions_NoNotebook(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create service without notebook path
	svc := services.NewNoteService(cfg, nil, "")

	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "test"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.Error(t, err, "Should error when no notebook selected")
	assert.Nil(t, results, "Results should be nil on error")
	assert.Contains(t, err.Error(), "no notebook selected", "Error should mention no notebook")
}

func TestNoteService_SearchWithConditions_ComplexQuery(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() {
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "match.md", `---
tag: workflow
status: active
priority: high
---
# Should Match
`)
	testutil.CreateTestNote(t, notebookDir, "wrong-status.md", `---
tag: workflow
status: archived
priority: high
---
# Wrong Status
`)
	testutil.CreateTestNote(t, notebookDir, "wrong-priority.md", `---
tag: workflow
status: active
priority: low
---
# Wrong Priority
`)
	testutil.CreateTestNote(t, notebookDir, "wrong-tag.md", `---
tag: meeting
status: active
priority: high
---
# Wrong Tag
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	// Complex query: AND + OR
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
		{Type: "and", Field: "data.status", Operator: "=", Value: "active"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
	}

	results, err := svc.SearchWithConditions(ctx, conditions)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 1, "Should find matching note")

	// Verify the match
	for _, note := range results {
		tag := note.Metadata["tag"]
		status := note.Metadata["status"]
		priority := note.Metadata["priority"]
		assert.Equal(t, "workflow", tag)
		assert.Equal(t, "active", status)
		assert.True(t, priority == "high" || priority == "critical", "Priority should be high or critical")
	}
}

// TestParseDataFlags tests the data flag parsing functionality
func TestParseDataFlags(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		want     map[string]interface{}
		wantErr  bool
		errMatch string
	}{
		{
			name:  "empty flags",
			flags: []string{},
			want:  map[string]interface{}{},
		},
		{
			name:  "single field",
			flags: []string{"tag=meeting"},
			want:  map[string]interface{}{"tag": "meeting"},
		},
		{
			name:  "multiple different fields",
			flags: []string{"tag=meeting", "priority=high", "status=draft"},
			want: map[string]interface{}{
				"tag":      "meeting",
				"priority": "high",
				"status":   "draft",
			},
		},
		{
			name:  "repeated field creates array",
			flags: []string{"tag=meeting", "tag=sprint", "tag=planning"},
			want: map[string]interface{}{
				"tag": []interface{}{"meeting", "sprint", "planning"},
			},
		},
		{
			name:  "mixed single and repeated fields",
			flags: []string{"tag=meeting", "priority=high", "tag=sprint"},
			want: map[string]interface{}{
				"tag":      []interface{}{"meeting", "sprint"},
				"priority": "high",
			},
		},
		{
			name:     "invalid format no equals",
			flags:    []string{"tagmeeting"},
			wantErr:  true,
			errMatch: "invalid --data format",
		},
		{
			name:     "invalid format empty field",
			flags:    []string{"=value"},
			wantErr:  true,
			errMatch: "field name cannot be empty",
		},
		{
			name:  "invalid format empty value",
			flags: []string{"field="},
			want:  map[string]interface{}{"field": ""},
		},
		{
			name:  "field with special characters in value",
			flags: []string{"description=Meeting notes: Q1 planning (2024)"},
			want:  map[string]interface{}{"description": "Meeting notes: Q1 planning (2024)"},
		},
		{
			name:  "field with equals in value",
			flags: []string{"equation=x=y+1"},
			want:  map[string]interface{}{"equation": "x=y+1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := services.ParseDataFlags(tt.flags)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMatch != "" {
					assert.Contains(t, err.Error(), tt.errMatch)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestResolvePath tests the path resolution functionality
func TestResolvePath(t *testing.T) {
	tests := []struct {
		name           string
		notebookRoot   string
		inputPath      string
		slugifiedTitle string
		want           string
	}{
		{
			name:           "no path uses root and slugified title",
			notebookRoot:   "/notebook",
			inputPath:      "",
			slugifiedTitle: "my-note",
			want:           "/notebook/my-note.md",
		},
		{
			name:           "folder path ending with slash",
			notebookRoot:   "/notebook",
			inputPath:      "meetings/",
			slugifiedTitle: "sprint-planning",
			want:           "/notebook/meetings/sprint-planning.md",
		},
		{
			name:           "full filepath with extension",
			notebookRoot:   "/notebook",
			inputPath:      "meetings/2024-01-20.md",
			slugifiedTitle: "meeting-notes",
			want:           "/notebook/meetings/2024-01-20.md",
		},
		{
			name:           "filepath without extension",
			notebookRoot:   "/notebook",
			inputPath:      "meetings/2024-01-20",
			slugifiedTitle: "meeting-notes",
			want:           "/notebook/meetings/2024-01-20.md",
		},
		{
			name:           "nested folder path",
			notebookRoot:   "/notebook",
			inputPath:      "work/meetings/",
			slugifiedTitle: "standup",
			want:           "/notebook/work/meetings/standup.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.ResolvePath(tt.notebookRoot, tt.inputPath, tt.slugifiedTitle)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNoteService_SearchWithFindOpts_Basic(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create test notebook with notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", `---
tag: work
status: active
---
# Note 1
Work related content.
`)
	testutil.CreateTestNote(t, notebookDir, "note2.md", `---
tag: personal
status: done
---
# Note 2
Personal note.
`)
	testutil.CreateTestNote(t, notebookDir, "note3.md", `---
tag: work
status: done
---
# Note 3
Another work note.
`)

	idx := testutil.CreateTestIndex(t, notebookDir)
	svc := services.NewNoteService(cfg, idx, notebookDir)

	t.Run("empty opts returns all notes", func(t *testing.T) {
		results, err := svc.SearchWithFindOpts(ctx, testutil.NewFindOpts())
		assert.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("with limit", func(t *testing.T) {
		results, err := svc.SearchWithFindOpts(ctx, testutil.NewFindOpts().WithLimit(2))
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})
}

func TestNoteService_SearchWithFindOpts_ErrorCases(t *testing.T) {
	ctx := context.Background()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	t.Run("no notebook selected returns error", func(t *testing.T) {
		svc := services.NewNoteService(cfg, nil, "")
		_, err := svc.SearchWithFindOpts(ctx, testutil.NewFindOpts())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no notebook selected")
	})

	t.Run("nil index returns error", func(t *testing.T) {
		svc := services.NewNoteService(cfg, nil, "/some/path")
		_, err := svc.SearchWithFindOpts(ctx, testutil.NewFindOpts())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "index not initialized")
	})
}
