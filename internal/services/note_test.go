package services_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/services"
	"github.com/zenobi-us/opennotes/internal/testutil"
)

func TestNoteService_SearchNotes_NoNotebookSelected(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")
	svc := services.NewNoteService(cfg, db, "")

	notes, err := svc.SearchNotes(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, notes)
	assert.Contains(t, err.Error(), "no notebook selected")
}

func TestNoteService_SearchNotes_FindsAllNotes(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create a test notebook with notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# Note 1\n\nFirst note content.")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Note 2\n\nSecond note content.")
	testutil.CreateTestNote(t, notebookDir, "note3.md", "# Note 3\n\nThird note content.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	assert.Len(t, notes, 3)
}

func TestNoteService_SearchNotes_FiltersByQuery(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create a test notebook with notes
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "apple.md", "# Apple\n\nThis is about apples.")
	testutil.CreateTestNote(t, notebookDir, "banana.md", "# Banana\n\nThis is about bananas.")
	testutil.CreateTestNote(t, notebookDir, "cherry.md", "# Cherry\n\nThis is about cherries.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Search for "apple"
	notes, err := svc.SearchNotes(ctx, "apple")
	require.NoError(t, err)

	assert.Len(t, notes, 1)
	assert.Contains(t, notes[0].File.Filepath, "apple.md")
}

func TestNoteService_SearchNotes_FiltersByQueryCaseInsensitive(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "mixed.md", "# UPPERCASE content\n\nSome text.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Search with lowercase should match uppercase content
	notes, err := svc.SearchNotes(ctx, "uppercase")
	require.NoError(t, err)

	assert.Len(t, notes, 1)
}

func TestNoteService_SearchNotes_FiltersByFilepath(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "project-ideas.md", "# Ideas\n\nSome ideas.")
	testutil.CreateTestNote(t, notebookDir, "daily-notes.md", "# Daily\n\nDaily notes.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Search by filename pattern
	notes, err := svc.SearchNotes(ctx, "project")
	require.NoError(t, err)

	assert.Len(t, notes, 1)
	assert.Contains(t, notes[0].File.Filepath, "project-ideas.md")
}

func TestNoteService_SearchNotes_EmptyNotebook(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Create empty notebook (no notes)
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "empty-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Note: DuckDB's read_markdown errors when no files match the glob.
	// This tests the current behavior - the service returns an error for empty notebooks.
	notes, err := svc.SearchNotes(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, notes)
	assert.Contains(t, err.Error(), "File or directory does not exist")
}

func TestNoteService_SearchNotes_ExtractsMetadata(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

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

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	// Metadata should be populated (DuckDB returns it as a map)
	assert.NotNil(t, notes[0].Metadata)
}

func TestNoteService_SearchNotes_SetsRelativePath(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "my-note.md", "# My Note\n\nContent here.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Equal(t, "notes/my-note.md", notes[0].File.Relative)
}

func TestNoteService_Count_NoNotebookSelected(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")
	svc := services.NewNoteService(cfg, db, "")

	// Count returns 0 when no notebook is selected (not an error)
	count, err := svc.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestNoteService_Count_ReturnsCorrectCount(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# Note 1")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Note 2")
	testutil.CreateTestNote(t, notebookDir, "note3.md", "# Note 3")
	testutil.CreateTestNote(t, notebookDir, "note4.md", "# Note 4")
	testutil.CreateTestNote(t, notebookDir, "note5.md", "# Note 5")

	svc := services.NewNoteService(cfg, db, notebookDir)

	count, err := svc.Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, count)
}

func TestNoteService_Count_EmptyNotebook(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "empty-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Note: DuckDB's read_markdown errors when no files match the glob.
	// This tests the current behavior - Count returns an error for empty notebooks.
	count, err := svc.Count(ctx)
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Contains(t, err.Error(), "File or directory does not exist")
}

func TestNoteService_Query_ExecutesSQL(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")
	svc := services.NewNoteService(cfg, db, "")

	// Query method delegates to DbService.Query
	results, err := svc.Query(ctx, "SELECT 42 as answer")
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, int32(42), results[0]["answer"])
}

func TestNoteService_Query_ReturnsResults(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "test.md", "# Test\n\nTest content.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Use raw SQL to query notes - use fmt.Sprintf since Query doesn't take args
	glob := notebookDir + "/**/*.md"
	query := fmt.Sprintf("SELECT * FROM read_markdown('%s')", glob)
	results, err := svc.Query(ctx, query)
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.NotNil(t, results[0]["content"])
}

func TestNoteService_SearchNotes_MultipleQueryMatches(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# First Note\n\nContains the word golang.")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Second Note\n\nAlso mentions golang here.")
	testutil.CreateTestNote(t, notebookDir, "note3.md", "# Third Note\n\nNo match in this one.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "golang")
	require.NoError(t, err)

	assert.Len(t, notes, 2)
}

func TestNoteService_SearchNotes_ContentHasText(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	defer db.Close()

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	expectedContent := "# My Note Title\n\nThis is the body content."
	testutil.CreateTestNote(t, notebookDir, "note.md", expectedContent)

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Contains(t, notes[0].Content, "My Note Title")
	assert.Contains(t, notes[0].Content, "body content")
}

func TestNewNoteService(t *testing.T) {
	db := services.NewDbService()
	defer db.Close()

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")

	svc := services.NewNoteService(cfg, db, "/test/notebook/path")

	assert.NotNil(t, svc)
}
