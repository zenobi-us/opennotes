package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zenobi-us/opennotes/internal/services"
)

// generateRandomNote creates a note with realistic content for stress testing
func generateRandomNote(index int) string {
	content := fmt.Sprintf(`---
title: "Stress Test Note %06d"
tags: ["stress%d", "test%d", "category%d"]
date: "%s"
priority: %d
status: "active"
---

# Stress Test Note %06d

This is a comprehensive stress test note designed to validate system performance 
under load. The content includes various markdown elements and metadata.

## Key Information
- **Note ID**: %06d
- **Generation Time**: %s
- **Test Category**: %d
- **Priority Level**: %d

## Sample Content
This note contains searchable content with keywords: test%d, stress%d, 
performance%d, scalability%d, benchmark%d.

### Lorem Ipsum Section
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor 
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis 
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

### Unicode Content
Here's some Unicode content for testing: 🚀 📊 ⭐ 🎯 💡 
Chinese: 测试内容 Japanese: テスト内容 Arabic: محتوى الاختبار

### Code Block
` + "```go\n" +
`func ExampleFunction() {
    fmt.Printf("Note %d generated at %s", %d, "%s")
}
` + "```\n" +
`
### Links and References
- [External Link](https://example.com/note/%d)
- Internal reference: [[Note %06d]]
- Tag reference: #test%d

---
Generated for OpenNotes stress testing - note %d of large dataset.
`,
		// title, tags, date, priority
		index, index%10, index%100, index%20, time.Now().Format(time.RFC3339), index%5+1,
		// header
		index,
		// key info
		index, time.Now().Format(time.RFC3339Nano), index%20, index%5+1,
		// sample content keywords
		index, index, index, index, index,
		// code block
		index, time.Now().Format(time.RFC3339), index, time.Now().Format(time.RFC3339),
		// links
		index, index, index%10,
		// footer
		index)

	// Add extra content for larger files on some notes
	if index%100 == 0 {
		content += "\n\n## Extended Content for Large File Testing\n\n"
		for i := 0; i < 50; i++ {
			content += fmt.Sprintf("Paragraph %d: This is additional content to create larger files "+
				"for testing purposes. The content includes repeated text with variations "+
				"to test search and processing performance with substantial file sizes.\n\n", i)
		}
	}

	return content
}

// generateStressNotebook creates a notebook with specified number of notes
func generateStressNotebook(t *testing.T, numNotes int, depth int) (string, *services.Notebook) {
	t.Helper()

	tempDir := t.TempDir()
	
	// Create notebook config
	configContent := fmt.Sprintf(`{
		"name": "Stress Test Notebook (%d notes)",
		"root": ".",
		"contexts": ["%s"]
	}`, numNotes, tempDir)
	
	configPath := filepath.Join(tempDir, ".opennotes.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create notes in various directory structures
	for i := 0; i < numNotes; i++ {
		var notePath string
		
		if depth > 1 && i%10 == 0 {
			// Create some nested structure
			subdir := fmt.Sprintf("category_%d/subcategory_%d", i%5, i%25)
			dirPath := filepath.Join(tempDir, subdir)
			err := os.MkdirAll(dirPath, 0755)
			require.NoError(t, err)
			notePath = filepath.Join(dirPath, fmt.Sprintf("note_%06d.md", i))
		} else {
			notePath = filepath.Join(tempDir, fmt.Sprintf("note_%06d.md", i))
		}
		
		content := generateRandomNote(i)
		err := os.WriteFile(notePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create notebook service and open the notebook
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	
	dbService := services.NewDbService()
	t.Cleanup(func() { dbService.Close() })
	
	notebookService := services.NewNotebookService(configService, dbService)
	notebook, err := notebookService.Open(tempDir)
	require.NoError(t, err)

	return tempDir, notebook
}

// TestNoteService_LargeNotebook tests performance with 1000+ notes
func TestNoteService_LargeNotebook(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large notebook stress test in short mode")
	}

	const numNotes = 1000 // Start with 1K for reasonable test time
	
	t.Logf("Generating %d notes for stress test...", numNotes)
	start := time.Now()
	_, notebook := generateStressNotebook(t, numNotes, 2)
	generationTime := time.Since(start)
	t.Logf("Note generation completed in %v", generationTime)

	// Test search performance
	searchStart := time.Now()
	results, err := notebook.Notes.SearchNotes(context.Background(), "test")
	searchTime := time.Since(searchStart)
	require.NoError(t, err)
	
	t.Logf("Search completed in %v, found %d results", searchTime, len(results))
	
	// Performance targets
	assert.Less(t, searchTime.Seconds(), 2.0, "Search should complete in < 2 seconds")
	assert.Greater(t, len(results), numNotes/2, "Search should find substantial results")

	// Test count performance
	countStart := time.Now()
	count, err := notebook.Notes.Count(context.Background())
	countTime := time.Since(countStart)
	require.NoError(t, err)
	
	t.Logf("Count operation completed in %v, counted %d notes", countTime, count)
	assert.Equal(t, numNotes, count, "Count should match generated notes")
	assert.Less(t, countTime, 500*time.Millisecond, "Count should complete in < 500ms")
}

// TestNoteService_DeepNesting tests performance with deep directory structures
func TestNoteService_DeepNesting(t *testing.T) {
	tempDir := t.TempDir()
	
	// Create deep directory structure
	deepPath := tempDir
	const maxDepth = 20 // Reasonable depth for testing
	
	for i := 0; i < maxDepth; i++ {
		deepPath = filepath.Join(deepPath, fmt.Sprintf("level_%02d", i))
		err := os.MkdirAll(deepPath, 0755)
		require.NoError(t, err)
		
		// Add a note at this level
		noteContent := fmt.Sprintf(`---
title: "Deep Note Level %d"
depth: %d
---

# Note at Depth %d

This note is at directory depth %d for testing deep structure handling.
`, i, i, i, i)
		
		notePath := filepath.Join(deepPath, fmt.Sprintf("note_depth_%02d.md", i))
		err = os.WriteFile(notePath, []byte(noteContent), 0644)
		require.NoError(t, err)
	}
	
	// Create notebook config at root
	configContent := `{
		"name": "Deep Nesting Test",
		"root": ".",
		"contexts": []
	}`
	configPath := filepath.Join(tempDir, ".opennotes.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Test notebook discovery and operations
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	
	dbService := services.NewDbService()
	defer dbService.Close()
	
	notebookService := services.NewNotebookService(configService, dbService)
	
	// Test discovery performance
	discoveryStart := time.Now()
	hasNotebook := notebookService.HasNotebook(tempDir)
	discoveryTime := time.Since(discoveryStart)
	
	assert.True(t, hasNotebook, "Should discover notebook with deep structure")
	assert.Less(t, discoveryTime, 100*time.Millisecond, "Discovery should complete in < 100ms")
	t.Logf("Deep structure discovery completed in %v", discoveryTime)
	
	// Test opening and searching deep structure
	notebook, err := notebookService.Open(tempDir)
	require.NoError(t, err)
	
	searchStart := time.Now()
	results, err := notebook.Notes.SearchNotes(context.Background(), "depth")
	searchTime := time.Since(searchStart)
	require.NoError(t, err)
	
	t.Logf("Deep structure search completed in %v, found %d results", searchTime, len(results))
	assert.Equal(t, maxDepth, len(results), "Should find notes at all depths")
	assert.Less(t, searchTime.Seconds(), 1.0, "Deep search should complete in < 1 second")
}

// TestNoteService_LargeFiles tests handling of large individual notes
func TestNoteService_LargeFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tempDir := t.TempDir()
	
	// Create notebook config
	configContent := `{
		"name": "Large File Test",
		"root": ".",
		"contexts": []
	}`
	configPath := filepath.Join(tempDir, ".opennotes.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Create large note (10MB content)
	largeContent := `---
title: "Large File Test"
size: "large"
---

# Large File Performance Test

This file contains a large amount of content to test handling of substantial markdown files.

`
	
	// Add 10MB of content
	const targetSizeMB = 10
	baseSizeKB := len(largeContent)
	additionalBytes := (targetSizeMB * 1024 * 1024) - baseSizeKB
	
	padding := strings.Repeat("A", additionalBytes/100) // Reduce memory usage
	for i := 0; i < 100; i++ {
		largeContent += fmt.Sprintf("\n## Section %d\n%s\n", i, padding)
	}
	
	largeFilePath := filepath.Join(tempDir, "large_note.md")
	
	writeStart := time.Now()
	err = os.WriteFile(largeFilePath, []byte(largeContent), 0644)
	writeTime := time.Since(writeStart)
	require.NoError(t, err)
	
	t.Logf("Large file written in %v", writeTime)

	// Test opening notebook with large file
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	
	dbService := services.NewDbService()
	defer dbService.Close()
	
	notebookService := services.NewNotebookService(configService, dbService)
	
	openStart := time.Now()
	notebook, err := notebookService.Open(tempDir)
	openTime := time.Since(openStart)
	require.NoError(t, err)
	
	t.Logf("Notebook with large file opened in %v", openTime)

	// Test search performance with large file
	searchStart := time.Now()
	results, err := notebook.Notes.SearchNotes(context.Background(), "Large")
	searchTime := time.Since(searchStart)
	require.NoError(t, err)
	
	t.Logf("Large file search completed in %v, found %d results", searchTime, len(results))
	assert.Greater(t, len(results), 0, "Should find the large file")
	assert.Less(t, searchTime.Seconds(), 5.0, "Large file search should complete in < 5 seconds")
}

// TestNoteService_UnicodeAtScale tests Unicode handling with many files
func TestNoteService_UnicodeAtScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Unicode scale test in short mode")
	}

	const numNotes = 500
	tempDir := t.TempDir()
	
	// Create notebook config
	configContent := `{
		"name": "Unicode Scale Test",
		"root": ".",
		"contexts": []
	}`
	configPath := filepath.Join(tempDir, ".opennotes.json")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	// Unicode test content samples
	unicodeContent := []string{
		"测试内容", "テスト", "محتوى الاختبار", "测试", "🚀📊⭐",
		"Тест", "тест", "δοκιμή", "परीक्षण", "প্রোগ্রামে",
		"한국어", "ไทย", "中文测试", "עברית", "🎯💡🔥",
	}

	// Generate Unicode-heavy notes
	for i := 0; i < numNotes; i++ {
		unicodeSample := unicodeContent[i%len(unicodeContent)]
		content := fmt.Sprintf(`---
title: "Unicode Test %s %d"
language: "multi"
unicode_test: true
---

# Unicode Test Note %d

Content: %s

Unicode characters: %s %s %s

Mixed content: Test%d with %s and 🚀 emoji.

Search terms: unicode%d, test%d, %s

`, unicodeSample, i, i, unicodeSample, 
   unicodeContent[i%len(unicodeContent)], 
   unicodeContent[(i+1)%len(unicodeContent)], 
   unicodeContent[(i+2)%len(unicodeContent)],
   i, unicodeSample, i, i, unicodeSample)

		notePath := filepath.Join(tempDir, fmt.Sprintf("unicode_note_%03d.md", i))
		err := os.WriteFile(notePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Test Unicode search performance
	configService, err := services.NewConfigService()
	require.NoError(t, err)
	
	dbService := services.NewDbService()
	defer dbService.Close()
	
	notebookService := services.NewNotebookService(configService, dbService)
	notebook, err := notebookService.Open(tempDir)
	require.NoError(t, err)

	// Test Unicode search
	for _, searchTerm := range []string{"测试", "🚀", "test", "unicode"} {
		searchStart := time.Now()
		results, err := notebook.Notes.SearchNotes(context.Background(), searchTerm)
		searchTime := time.Since(searchStart)
		require.NoError(t, err)
		
		t.Logf("Unicode search for '%s' completed in %v, found %d results", 
			searchTerm, searchTime, len(results))
		
		if searchTerm == "test" || searchTerm == "unicode" {
			assert.Greater(t, len(results), numNotes/2, "Common terms should find many results")
		}
		assert.Less(t, searchTime, 500*time.Millisecond, "Unicode search should be fast")
	}
}

// TestNoteService_MemoryUsageScale tests memory usage patterns with increasing dataset sizes
func TestNoteService_MemoryUsageScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory scale test in short mode")
	}

	sizes := []int{100, 250, 500}
	
	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d_notes", size), func(t *testing.T) {
			var m1, m2 runtime.MemStats
			
			runtime.GC()
			runtime.ReadMemStats(&m1)

			_, notebook := generateStressNotebook(t, size, 1)
			
			// Perform operations that load data into memory
			_, err := notebook.Notes.SearchNotes(context.Background(), "test")
			require.NoError(t, err)
			
			count, err := notebook.Notes.Count(context.Background())
			require.NoError(t, err)
			assert.Equal(t, size, count)

			runtime.GC()
			runtime.ReadMemStats(&m2)

			memUsed := m2.Alloc - m1.Alloc
			memPerNote := memUsed / uint64(size)

			t.Logf("Size: %d notes, Memory: %s, Per note: %s",
				size, formatBytes(memUsed), formatBytes(memPerNote))

			// Memory per note should be reasonable (< 50KB per note including metadata)
			assert.Less(t, memPerNote, uint64(50*1024),
				"Memory per note should be < 50KB")
				
			// Total memory should be reasonable (< 50MB for 500 notes)
			if size >= 500 {
				assert.Less(t, memUsed, uint64(50*1024*1024),
					"Total memory for 500 notes should be < 50MB")
			}
		})
	}
}

// TestNoteService_SearchPerformanceScale tests search performance across different dataset sizes
func TestNoteService_SearchPerformanceScale(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping search performance scale test in short mode")
	}

	sizes := []int{100, 500, 1000}
	
	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d_notes", size), func(t *testing.T) {
			_, notebook := generateStressNotebook(t, size, 1)
			
			// Test different search scenarios
			searchTests := []struct {
				term     string
				expected string
			}{
				{"test", "common term"},
				{"stress", "specific term"},
				{"note_000001", "exact match"},
				{"nonexistent", "no results"},
			}
			
			for _, test := range searchTests {
				start := time.Now()
				results, err := notebook.Notes.SearchNotes(context.Background(), test.term)
				searchTime := time.Since(start)
				require.NoError(t, err)
				
				t.Logf("Search '%s' (%s) in %d notes: %v, found %d results", 
					test.term, test.expected, size, searchTime, len(results))
				
				// Performance targets scale with dataset size
				maxTime := time.Duration(size/100+1) * time.Second
				assert.Less(t, searchTime, maxTime, 
					"Search time should scale reasonably with dataset size")
			}
		})
	}
}

// formatBytes converts bytes to human-readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}