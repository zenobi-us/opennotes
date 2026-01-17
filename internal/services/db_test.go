package services

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDbService_GetDB_ReturnsConnection(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db, err := svc.GetDB(ctx)
	require.NoError(t, err)
	assert.NotNil(t, db)
}

func TestDbService_GetDB_LoadsMarkdownExtension(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db, err := svc.GetDB(ctx)
	require.NoError(t, err)

	// Verify markdown extension is loaded by checking for the function
	rows, err := db.QueryContext(ctx, "SELECT extension_name FROM duckdb_extensions() WHERE extension_name = 'markdown' AND loaded = true")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	})

	// Should find the markdown extension
	assert.True(t, rows.Next(), "markdown extension should be loaded")
}

func TestDbService_GetDB_LazyInit(t *testing.T) {
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Before GetDB, db should be nil
	assert.Nil(t, svc.db)

	// After GetDB, db should be initialized
	ctx := context.Background()
	_, err := svc.GetDB(ctx)
	require.NoError(t, err)
	assert.NotNil(t, svc.db)
}

func TestDbService_GetDB_ReturnsSameConnection(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db1, err := svc.GetDB(ctx)
	require.NoError(t, err)

	db2, err := svc.GetDB(ctx)
	require.NoError(t, err)

	// Should return the same connection
	assert.Same(t, db1, db2)
}

func TestDbService_Query_SimpleSQL(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	results, err := svc.Query(ctx, "SELECT 1 as value, 'hello' as message")
	require.NoError(t, err)

	require.Len(t, results, 1)
	assert.Equal(t, int32(1), results[0]["value"])
	assert.Equal(t, "hello", results[0]["message"])
}

func TestDbService_Query_ResultMapping(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Query with multiple rows
	results, err := svc.Query(ctx, `
		SELECT * FROM (VALUES (1, 'a'), (2, 'b'), (3, 'c')) AS t(id, letter)
	`)
	require.NoError(t, err)

	require.Len(t, results, 3)

	// Verify column names and values
	assert.Equal(t, int32(1), results[0]["id"])
	assert.Equal(t, "a", results[0]["letter"])
	assert.Equal(t, int32(2), results[1]["id"])
	assert.Equal(t, "b", results[1]["letter"])
	assert.Equal(t, int32(3), results[2]["id"])
	assert.Equal(t, "c", results[2]["letter"])
}

func TestDbService_Query_ReadMarkdown(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Create a temporary markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `---
title: Test Note
tags: [test, sample]
---

# Test Note

This is test content.
`
	err := os.WriteFile(mdFile, []byte(content), 0644)
	require.NoError(t, err)

	// Query using read_markdown
	results, err := svc.Query(ctx, "SELECT * FROM read_markdown(?)", mdFile)
	require.NoError(t, err)

	require.Len(t, results, 1)

	// Verify markdown metadata was extracted (returns duckdb.Map)
	metadata := results[0]["metadata"]
	assert.NotNil(t, metadata)

	// Verify content is present
	mdContent := results[0]["content"]
	assert.NotNil(t, mdContent)
	assert.Contains(t, mdContent, "# Test Note")
}

func TestDbService_Query_EmptyResult(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	results, err := svc.Query(ctx, "SELECT 1 WHERE 1=0")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestDbService_Query_InvalidSQL(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	_, err := svc.Query(ctx, "INVALID SQL SYNTAX")
	assert.Error(t, err)
}

func TestDbService_Close(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()

	// Initialize the connection
	_, err := svc.GetDB(ctx)
	require.NoError(t, err)

	// Close should succeed
	err = svc.Close()
	assert.NoError(t, err)
}

func TestDbService_Close_NilDB(t *testing.T) {
	svc := NewDbService()

	// Close on uninitialized service should not error
	err := svc.Close()
	assert.NoError(t, err)
}

func TestDbService_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Run multiple goroutines calling GetDB concurrently
	const numGoroutines = 10
	var wg sync.WaitGroup
	errs := make(chan error, numGoroutines)
	dbs := make(chan interface{}, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := svc.GetDB(ctx)
			if err != nil {
				errs <- err
				return
			}
			dbs <- db
		}()
	}

	wg.Wait()
	close(errs)
	close(dbs)

	// No errors should have occurred
	for err := range errs {
		t.Errorf("concurrent GetDB failed: %v", err)
	}

	// All goroutines should have received the same DB instance
	var firstDB interface{}
	for db := range dbs {
		if firstDB == nil {
			firstDB = db
		} else {
			assert.Same(t, firstDB, db)
		}
	}
}

func TestDbService_Query_WithArgs(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	results, err := svc.Query(ctx, "SELECT ? as value, ? as name", 42, "test")
	require.NoError(t, err)

	require.Len(t, results, 1)
	// DuckDB returns int64 for integer parameters
	assert.Equal(t, int64(42), results[0]["value"])
	assert.Equal(t, "test", results[0]["name"])
}

// Tests for GetReadOnlyDB

func TestDbService_GetReadOnlyDB_ReturnsConnection(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)
	assert.NotNil(t, db)
}

func TestDbService_GetReadOnlyDB_LoadsMarkdownExtension(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	// Verify markdown extension is loaded
	rows, err := db.QueryContext(ctx, "SELECT extension_name FROM duckdb_extensions() WHERE extension_name = 'markdown' AND loaded = true")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	})

	// Should find the markdown extension
	assert.True(t, rows.Next(), "markdown extension should be loaded on read-only connection")
}

func TestDbService_GetReadOnlyDB_LazyInit(t *testing.T) {
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Before GetReadOnlyDB, readOnly should be nil
	assert.Nil(t, svc.readOnly)

	// After GetReadOnlyDB, readOnly should be initialized
	ctx := context.Background()
	_, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)
	assert.NotNil(t, svc.readOnly)
}

func TestDbService_GetReadOnlyDB_ReturnsSameConnection(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db1, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	db2, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	// Should return the same connection
	assert.Same(t, db1, db2)
}

func TestDbService_GetReadOnlyDB_IsSeparateFromMainDB(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db, err := svc.GetDB(ctx)
	require.NoError(t, err)

	roDb, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	// Should be different connections
	assert.NotSame(t, db, roDb)
}

func TestDbService_GetReadOnlyDB_ExecutesQuery(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	db, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	// Should be able to execute a simple query
	rows, err := db.QueryContext(ctx, "SELECT 1 as value")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	})

	assert.True(t, rows.Next())
	var value int
	err = rows.Scan(&value)
	require.NoError(t, err)
	assert.Equal(t, 1, value)
}

func TestDbService_Close_BothConnections(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()

	// Initialize both connections
	_, err := svc.GetDB(ctx)
	require.NoError(t, err)

	_, err = svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	// Close should close both
	err = svc.Close()
	assert.NoError(t, err)
}

func TestDbService_GetReadOnlyDB_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Run multiple goroutines calling GetReadOnlyDB concurrently
	const numGoroutines = 10
	var wg sync.WaitGroup
	errs := make(chan error, numGoroutines)
	dbs := make(chan interface{}, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db, err := svc.GetReadOnlyDB(ctx)
			if err != nil {
				errs <- err
				return
			}
			dbs <- db
		}()
	}

	wg.Wait()
	close(errs)
	close(dbs)

	// No errors should have occurred
	for err := range errs {
		t.Errorf("concurrent GetReadOnlyDB failed: %v", err)
	}

	// All goroutines should have received the same DB instance
	var firstDB interface{}
	for db := range dbs {
		if firstDB == nil {
			firstDB = db
		} else {
			assert.Same(t, firstDB, db)
		}
	}
}

func TestDbService_GetReadOnlyDB_ReadMarkdown(t *testing.T) {
	ctx := context.Background()
	svc := NewDbService()
	t.Cleanup(func() {
		if err := svc.Close(); err != nil {
			t.Logf("warning: failed to close db: %v", err)
		}
	})

	// Create a temporary markdown file
	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := `# Read Only Test

This should be readable from read-only connection.
`
	err := os.WriteFile(mdFile, []byte(content), 0644)
	require.NoError(t, err)

	// Query using read-only connection
	db, err := svc.GetReadOnlyDB(ctx)
	require.NoError(t, err)

	rows, err := db.QueryContext(ctx, "SELECT * FROM read_markdown(?)", mdFile)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := rows.Close(); err != nil {
			t.Logf("warning: failed to close rows: %v", err)
		}
	})

	// Should be able to read the markdown file
	assert.True(t, rows.Next(), "read-only connection should be able to read markdown files")
}
