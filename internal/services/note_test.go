package services_test

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/services"
	"github.com/zenobi-us/opennotes/internal/testutil"
)

func TestNoteService_SearchNotes_NoNotebookSelected(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

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
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	cfg, _ := services.NewConfigServiceWithPath(t.TempDir() + "/config.json")

	svc := services.NewNoteService(cfg, db, "/test/notebook/path")

	assert.NotNil(t, svc)
}

// Tests for ValidateSQL

func TestValidateSQL_SelectQuery(t *testing.T) {
	tests := []string{
		"SELECT * FROM markdown",
		"SELECT id, title FROM markdown WHERE id > 5",
		"  SELECT  *  FROM  markdown  ",
		"select * from markdown",
		"SeLeCt * FrOm markdown",
	}

	for _, query := range tests {
		t.Run(fmt.Sprintf("valid_select_%s", query[:10]), func(t *testing.T) {
			err := services.ValidateSQL(query)
			assert.NoError(t, err, "valid SELECT query should pass: %s", query)
		})
	}
}

func TestValidateSQL_WithQuery(t *testing.T) {
	tests := []string{
		"WITH cte AS (SELECT * FROM markdown) SELECT * FROM cte",
		"  WITH  cte  AS  (SELECT * FROM markdown) SELECT * FROM cte  ",
		"with cte as (select * from markdown) select * from cte",
	}

	for _, query := range tests {
		t.Run(fmt.Sprintf("valid_with_%s", query[:10]), func(t *testing.T) {
			err := services.ValidateSQL(query)
			assert.NoError(t, err, "valid WITH (CTE) query should pass: %s", query)
		})
	}
}

func TestValidateSQL_EmptyQuery(t *testing.T) {
	tests := []string{
		"",
		"   ",
		"\n\t",
	}

	for _, query := range tests {
		t.Run("empty_query", func(t *testing.T) {
			err := services.ValidateSQL(query)
			assert.Error(t, err, "empty query should fail")
			assert.Contains(t, err.Error(), "empty")
		})
	}
}

func TestValidateSQL_InvalidQueryType(t *testing.T) {
	tests := []string{
		"INSERT INTO markdown VALUES (...)",
		"UPDATE markdown SET col = val",
		"DELETE FROM markdown",
		"SHOW TABLES",
		"DESCRIBE markdown",
	}

	for _, query := range tests {
		t.Run(fmt.Sprintf("invalid_type_%s", query[:10]), func(t *testing.T) {
			err := services.ValidateSQL(query)
			assert.Error(t, err, "non-SELECT query should fail: %s", query)
			assert.Contains(t, err.Error(), "only SELECT queries are allowed")
		})
	}
}

func TestValidateSQL_BlockedKeyword_Drop(t *testing.T) {
	err := services.ValidateSQL("SELECT * FROM markdown DROP TABLE temp")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DROP")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Delete(t *testing.T) {
	err := services.ValidateSQL("SELECT * DELETE FROM markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DELETE")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Update(t *testing.T) {
	err := services.ValidateSQL("SELECT * UPDATE markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "UPDATE")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Insert(t *testing.T) {
	err := services.ValidateSQL("SELECT * INSERT INTO markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "INSERT")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Alter(t *testing.T) {
	err := services.ValidateSQL("SELECT * ALTER TABLE markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ALTER")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Create(t *testing.T) {
	err := services.ValidateSQL("SELECT * CREATE TABLE markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "CREATE")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Truncate(t *testing.T) {
	err := services.ValidateSQL("SELECT * TRUNCATE TABLE x")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "TRUNCATE")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Replace(t *testing.T) {
	err := services.ValidateSQL("SELECT * REPLACE INTO markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "REPLACE")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Attach(t *testing.T) {
	err := services.ValidateSQL("SELECT * ATTACH DATABASE 'file.db'")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ATTACH")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Detach(t *testing.T) {
	err := services.ValidateSQL("SELECT * DETACH DATABASE db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DETACH")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_BlockedKeyword_Pragma(t *testing.T) {
	err := services.ValidateSQL("SELECT * PRAGMA table_info")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "PRAGMA")
	assert.Contains(t, err.Error(), "not allowed")
}

func TestValidateSQL_CaseInsensitive(t *testing.T) {
	// Same queries in different cases should all be blocked
	tests := []string{
		"drop table markdown",
		"DROP TABLE markdown",
		"DrOp TaBlE markdown",
	}

	for _, query := range tests {
		t.Run(fmt.Sprintf("case_insensitive_%s", query[:10]), func(t *testing.T) {
			err := services.ValidateSQL(query)
			assert.Error(t, err, "DROP in any case should be blocked: %s", query)
		})
	}
}

func TestValidateSQL_KeywordInStrings(t *testing.T) {
	// With word boundary checking, keywords in string literals are now allowed
	// This is better UX - users can search for content containing these words
	err := services.ValidateSQL("SELECT 'DROP' as dangerous FROM markdown")
	assert.NoError(t, err, "keywords in string literals should be allowed")

	// But if the keyword appears outside strings, it should still be blocked
	err = services.ValidateSQL("SELECT 'drop' as text DROP TABLE markdown")
	assert.Error(t, err, "actual DROP keyword should be blocked")
	assert.Contains(t, err.Error(), "DROP")
}

func TestValidateSQL_ComplexValidQuery(t *testing.T) {
	query := `
	WITH recent_notes AS (
		SELECT id, title, content
		FROM markdown
		WHERE published_date > NOW() - INTERVAL 7 DAY
	),
	tagged_notes AS (
		SELECT id, title, 'recent' as tag
		FROM recent_notes
	)
	SELECT * FROM tagged_notes
	ORDER BY id DESC
	LIMIT 100
	`

	err := services.ValidateSQL(query)
	assert.NoError(t, err, "complex valid CTE query should pass")
}

func TestValidateSQL_SubqueryValid(t *testing.T) {
	query := `
	SELECT * FROM (
		SELECT id, COUNT(*) as cnt
		FROM markdown
		GROUP BY id
		HAVING COUNT(*) > 5
	) AS subquery
	`

	err := services.ValidateSQL(query)
	assert.NoError(t, err, "subquery should pass validation")
}

func TestValidateSQL_JoinValid(t *testing.T) {
	query := `
	SELECT m1.id, m2.id
	FROM markdown m1
	JOIN markdown m2 ON m1.id = m2.id
	WHERE m1.content LIKE '%test%'
	`

	err := services.ValidateSQL(query)
	assert.NoError(t, err, "JOIN should be allowed")
}

func TestValidateSQL_UnionValid(t *testing.T) {
	query := `
	SELECT id FROM markdown WHERE id > 10
	UNION
	SELECT id FROM markdown WHERE title LIKE '%test%'
	`

	err := services.ValidateSQL(query)
	assert.NoError(t, err, "UNION should be allowed")
}

// Tests for ExecuteSQLSafe

func TestNoteService_ExecuteSQLSafe_ValidSelect(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")
	testutil.CreateTestNote(t, notebookDir, "note1.md", "# Note 1\n\nContent with numbers 42 and 100.")
	testutil.CreateTestNote(t, notebookDir, "note2.md", "# Note 2\n\nAnother note.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	results, err := svc.ExecuteSQLSafe(ctx, "SELECT 1 as value, 'test' as message")
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, int32(1), results[0]["value"])
	assert.Equal(t, "test", results[0]["message"])
}

func TestNoteService_ExecuteSQLSafe_InvalidQuery(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Try to execute DROP query
	_, err := svc.ExecuteSQLSafe(ctx, "DROP TABLE markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
}

func TestNoteService_ExecuteSQLSafe_EmptyResult(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Query that returns no results
	results, err := svc.ExecuteSQLSafe(ctx, "SELECT 1 WHERE 1=0")
	require.NoError(t, err)

	assert.Empty(t, results)
	assert.Len(t, results, 0)
}

func TestNoteService_ExecuteSQLSafe_MultipleRows(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Query that returns multiple rows
	results, err := svc.ExecuteSQLSafe(ctx, `
		SELECT * FROM (
			VALUES (1, 'a'), (2, 'b'), (3, 'c')
		) AS t(id, letter)
	`)
	require.NoError(t, err)

	require.Len(t, results, 3)
	assert.Equal(t, int32(1), results[0]["id"])
	assert.Equal(t, "a", results[0]["letter"])
	assert.Equal(t, int32(2), results[1]["id"])
	assert.Equal(t, "b", results[1]["letter"])
	assert.Equal(t, int32(3), results[2]["id"])
	assert.Equal(t, "c", results[2]["letter"])
}

func TestNoteService_ExecuteSQLSafe_WithClause(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// WITH (CTE) query
	results, err := svc.ExecuteSQLSafe(ctx, `
		WITH cte AS (
			SELECT 1 as num, 'first' as label
			UNION ALL
			SELECT 2 as num, 'second' as label
		)
		SELECT * FROM cte WHERE num > 1
	`)
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, int32(2), results[0]["num"])
	assert.Equal(t, "second", results[0]["label"])
}

func TestNoteService_ExecuteSQLSafe_InvalidSyntax(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Invalid SQL syntax
	_, err := svc.ExecuteSQLSafe(ctx, "SELECT * INVALID SYNTAX HERE")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query execution failed")
}

func TestNoteService_ExecuteSQLSafe_ContextCancellation(t *testing.T) {
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Create cancelled context
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should fail due to cancelled context
	_, err := svc.ExecuteSQLSafe(cancelledCtx, "SELECT 1")
	assert.Error(t, err)
}

func TestNoteService_ExecuteSQLSafe_TypeConversions(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Query with various types
	results, err := svc.ExecuteSQLSafe(ctx, `
		SELECT 
			42 as int_val,
			3.14 as float_val,
			'text' as str_val,
			true as bool_val,
			NULL as null_val
	`)
	require.NoError(t, err)

	require.Len(t, results, 1)
	row := results[0]

	// Check type conversions
	assert.NotNil(t, row["int_val"])
	assert.NotNil(t, row["str_val"])
	assert.Equal(t, true, row["bool_val"])
	assert.Nil(t, row["null_val"])
}

func TestNoteService_ExecuteSQLSafe_ComplexQuery(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	// Complex query with joins, aggregation, filtering
	results, err := svc.ExecuteSQLSafe(ctx, `
		WITH numbered AS (
			SELECT 1 as n, 'a' as letter
			UNION ALL
			SELECT 2 as n, 'b' as letter
			UNION ALL
			SELECT 3 as n, 'c' as letter
		)
		SELECT 
			n,
			letter,
			LENGTH(letter) as letter_len
		FROM numbered
		WHERE n >= 2
		ORDER BY n DESC
	`)
	require.NoError(t, err)

	require.Len(t, results, 2)
	
	// First result (n=3)
	assert.Equal(t, int32(3), results[0]["n"])
	assert.Equal(t, "c", results[0]["letter"])
	
	// Second result (n=2)
	assert.Equal(t, int32(2), results[1]["n"])
	assert.Equal(t, "b", results[1]["letter"])
}

func TestNoteService_ExecuteSQLSafe_ReadOnlyEnforcement(t *testing.T) {
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")
	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	svc := services.NewNoteService(cfg, db, notebookDir)

	ctx := context.Background()

	// Even if validation didn't catch it, read-only connection prevents writes
	// This is caught by ValidateSQL, but the read-only connection is a defense-in-depth layer
	_, err := svc.ExecuteSQLSafe(ctx, "SELECT 1")
	require.NoError(t, err)
	
	// DELETE would be caught by validation before reaching the DB
	_, err = svc.ExecuteSQLSafe(ctx, "DELETE FROM markdown")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
}

func TestNoteService_SearchNotes_DisplayNameWithTitle(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
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

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Equal(t, "My Custom Title", notes[0].DisplayName())
}

func TestNoteService_SearchNotes_DisplayNameSlugifyFilename(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "test-notebook")

	// Create note without title - should slugify filename
	testutil.CreateTestNote(t, notebookDir, "Hello World.md", "# Hello\n\nContent here.")

	svc := services.NewNoteService(cfg, db, notebookDir)

	notes, err := svc.SearchNotes(ctx, "")
	require.NoError(t, err)

	require.Len(t, notes, 1)
	assert.Equal(t, "hello-world", notes[0].DisplayName())
}

func TestNoteService_SearchNotes_DisplayNameMultipleNotes(t *testing.T) {
	ctx := context.Background()
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
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

	svc := services.NewNoteService(cfg, db, notebookDir)

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
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
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

	svc := services.NewNoteService(cfg, db, notebookDir)

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
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	notebookDir := testutil.CreateTestNotebook(t, tmpDir, "special-chars-test")
	
	// Create notes with special characters
	testutil.CreateTestNote(t, notebookDir, "unicode-test.md", "# Unicode Test\n\nCafé, naïve, résumé")
	testutil.CreateTestNote(t, notebookDir, "symbols.md", "# Symbols\n\nC++ programming, @mentions, #hashtags")
	testutil.CreateTestNote(t, notebookDir, "quotes.md", "# Quotes\n\n\"Double quotes\" and 'single quotes'")
	testutil.CreateTestNote(t, notebookDir, "math.md", "# Math\n\n2 + 2 = 4, x² + y² = z²")

	svc := services.NewNoteService(cfg, db, notebookDir)

	tests := []struct {
		name  string
		query string
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
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
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

	svc := services.NewNoteService(cfg, db, notebookDir)

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
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
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

	svc := services.NewNoteService(cfg, db, notebookDir)

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
	db := services.NewDbService()
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	tmpDir := t.TempDir()
	cfg, _ := services.NewConfigServiceWithPath(tmpDir + "/config.json")

	// Test with empty/non-existent notebook
	svc := services.NewNoteService(cfg, db, "")

	notes, err := svc.SearchNotes(ctx, "test")
	assert.Error(t, err, "Should error when no notebook selected")
	assert.Nil(t, notes, "Notes should be nil on error")
	assert.Contains(t, err.Error(), "no notebook selected", "Error should mention no notebook")

	// Test with non-existent notebook path
	nonExistentPath := filepath.Join(tmpDir, "nonexistent-notebook")
	svc2 := services.NewNoteService(cfg, db, nonExistentPath)

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
