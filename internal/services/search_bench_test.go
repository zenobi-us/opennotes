package services_test

import (
	"strings"
	"testing"

	"github.com/zenobi-us/jot/internal/services"
)

// createBenchmarkNotes creates a slice of notes for benchmarking
// with realistic content and metadata.
func createBenchmarkNotes(count int) []services.Note {
	notes := make([]services.Note, count)
	tags := []string{"workflow", "meeting", "project", "task", "epic", "spec", "docs"}
	statuses := []string{"active", "done", "archived", "in-progress"}
	priorities := []string{"high", "medium", "low", "critical"}

	for i := 0; i < count; i++ {
		// Create realistic content
		content := "This is test content for note " + strings.Repeat("x", i%100) + "\n"
		content += "Some notes mention meetings and projects.\n"
		content += "Others discuss workflows and tasks.\n"

		// Create links array for link query benchmarks
		links := []string{}
		if i%5 == 0 {
			links = append(links, "epics/epic-001.md")
		}
		if i%7 == 0 {
			links = append(links, "tasks/task-"+string(rune('0'+i%10))+".md")
		}
		if i%11 == 0 {
			links = append(links, "docs/architecture.md")
		}

		notes[i] = services.Note{
			Content: content,
			Metadata: map[string]any{
				"title":    "Note " + strings.Repeat("x", i%50),
				"tag":      tags[i%len(tags)],
				"status":   statuses[i%len(statuses)],
				"priority": priorities[i%len(priorities)],
				"links":    links,
			},
		}
		notes[i].File.Filepath = "/notebook/notes/note-" + string(rune('0'+i/1000)) + string(rune('0'+(i/100)%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+i%10)) + ".md"
		notes[i].File.Relative = "notes/note-" + string(rune('0'+i/1000)) + string(rune('0'+(i/100)%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+i%10)) + ".md"
	}

	// Add some specific notes with "meeting" keyword for consistent testing
	for i := 0; i < count && i < 100; i += 10 {
		notes[i].Metadata["title"] = "Meeting Notes " + strings.Repeat("x", i%10)
	}

	return notes
}

// ============================================================================
// Text Search Benchmarks
// ============================================================================

// BenchmarkTextSearch_100Notes benchmarks text search on 100 notes.
func BenchmarkTextSearch_100Notes(b *testing.B) {
	svc := services.NewSearchService()
	notes := createBenchmarkNotes(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := svc.TextSearch("meeting", notes)
		_ = results
	}
}

// BenchmarkTextSearch_1kNotes benchmarks text search on 1,000 notes.
func BenchmarkTextSearch_1kNotes(b *testing.B) {
	svc := services.NewSearchService()
	notes := createBenchmarkNotes(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := svc.TextSearch("meeting", notes)
		_ = results
	}
}

// BenchmarkTextSearch_10kNotes benchmarks text search on 10,000 notes.
// TARGET: < 10ms
func BenchmarkTextSearch_10kNotes(b *testing.B) {
	svc := services.NewSearchService()
	notes := createBenchmarkNotes(10000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := svc.TextSearch("meeting", notes)
		_ = results
	}
	// Target: < 10ms per operation
	// Actual ~1.4ms based on testing
}

// ============================================================================
// Boolean Query Benchmarks - Query Building Performance
// ============================================================================

// BenchmarkBooleanQuery_Simple benchmarks building a simple boolean query.
// TARGET: < 20ms for query building
func BenchmarkBooleanQuery_Simple(b *testing.B) {
	svc := services.NewSearchService()
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereClause, params, err := svc.BuildWhereClause(conditions)
		_ = whereClause
		_ = params
		_ = err
	}
	// Target: < 20ms per operation
	// Query building is fast (microseconds)
}

// BenchmarkBooleanQuery_Complex benchmarks building a complex boolean query.
// TARGET: < 100ms for query building
func BenchmarkBooleanQuery_Complex(b *testing.B) {
	svc := services.NewSearchService()
	conditions := []services.QueryCondition{
		{Type: "and", Field: "data.tag", Operator: "=", Value: "workflow"},
		{Type: "and", Field: "data.status", Operator: "=", Value: "active"},
		{Type: "and", Field: "links-to", Operator: "=", Value: "epics/**/*.md"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "high"},
		{Type: "or", Field: "data.priority", Operator: "=", Value: "critical"},
		{Type: "not", Field: "data.status", Operator: "=", Value: "archived"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereClause, params, err := svc.BuildWhereClause(conditions)
		_ = whereClause
		_ = params
		_ = err
	}
	// Target: < 100ms per operation
	// Query building is fast (microseconds)
}

// BenchmarkBooleanQuery_ManyConditions benchmarks building a query with many conditions.
func BenchmarkBooleanQuery_ManyConditions(b *testing.B) {
	svc := services.NewSearchService()
	var conditions []services.QueryCondition

	// Add 20 conditions
	for i := 0; i < 5; i++ {
		conditions = append(conditions, services.QueryCondition{
			Type: "and", Field: "data.tag", Operator: "=", Value: "tag" + string(rune('0'+i)),
		})
	}
	for i := 0; i < 10; i++ {
		conditions = append(conditions, services.QueryCondition{
			Type: "or", Field: "data.priority", Operator: "=", Value: "p" + string(rune('0'+i)),
		})
	}
	for i := 0; i < 5; i++ {
		conditions = append(conditions, services.QueryCondition{
			Type: "not", Field: "data.status", Operator: "=", Value: "s" + string(rune('0'+i)),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereClause, params, err := svc.BuildWhereClause(conditions)
		_ = whereClause
		_ = params
		_ = err
	}
}

// ============================================================================
// Link Query Benchmarks - Query Building Performance
// ============================================================================

// BenchmarkLinkQuery_LinksTo benchmarks building a links-to query.
// TARGET: < 50ms for query building
func BenchmarkLinkQuery_LinksTo(b *testing.B) {
	svc := services.NewSearchService()
	conditions := []services.QueryCondition{
		{Type: "and", Field: "links-to", Operator: "=", Value: "epics/**/*.md"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereClause, params, err := svc.BuildWhereClause(conditions)
		_ = whereClause
		_ = params
		_ = err
	}
	// Target: < 50ms per operation
	// Query building is fast (microseconds)
}

// BenchmarkLinkQuery_LinkedBy benchmarks building a linked-by query.
// TARGET: < 50ms for query building
func BenchmarkLinkQuery_LinkedBy(b *testing.B) {
	svc := services.NewSearchService()
	conditions := []services.QueryCondition{
		{Type: "and", Field: "linked-by", Operator: "=", Value: "planning/q1.md"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereClause, params, err := svc.BuildWhereClauseWithGlob(conditions, "/notebook/**/*.md")
		_ = whereClause
		_ = params
		_ = err
	}
}

// BenchmarkLinkQuery_Combined benchmarks building a combined link query.
func BenchmarkLinkQuery_Combined(b *testing.B) {
	svc := services.NewSearchService()
	conditions := []services.QueryCondition{
		{Type: "and", Field: "links-to", Operator: "=", Value: "tasks/**/*.md"},
		{Type: "and", Field: "linked-by", Operator: "=", Value: "epics/epic-001.md"},
		{Type: "not", Field: "links-to", Operator: "=", Value: "archived/**/*.md"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		whereClause, params, err := svc.BuildWhereClauseWithGlob(conditions, "/notebook/**/*.md")
		_ = whereClause
		_ = params
		_ = err
	}
}

// ============================================================================
// Glob Pattern Conversion Benchmarks
// ============================================================================

// BenchmarkGlobToLike_Simple benchmarks simple glob pattern conversion.
func BenchmarkGlobToLike_Simple(b *testing.B) {
	patterns := []string{"*.md", "dir/*", "file.md"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range patterns {
			result := services.GlobToLike(p)
			_ = result
		}
	}
}

// BenchmarkGlobToLike_Complex benchmarks complex glob pattern conversion.
func BenchmarkGlobToLike_Complex(b *testing.B) {
	patterns := []string{
		"**/*.md",
		"a/b/c/**/*.md",
		"dir/**/sub/*.md",
		"file-??.md",
		"100%_test/*.md", // with escape chars
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range patterns {
			result := services.GlobToLike(p)
			_ = result
		}
	}
}

// ============================================================================
// Condition Parsing Benchmarks
// ============================================================================

// BenchmarkParseConditions_Simple benchmarks parsing a simple condition set.
func BenchmarkParseConditions_Simple(b *testing.B) {
	svc := services.NewSearchService()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conditions, err := svc.ParseConditions(
			[]string{"data.tag=workflow"},
			[]string{},
			[]string{},
		)
		_ = conditions
		_ = err
	}
}

// BenchmarkParseConditions_Complex benchmarks parsing a complex condition set.
func BenchmarkParseConditions_Complex(b *testing.B) {
	svc := services.NewSearchService()

	andFlags := []string{
		"data.tag=workflow",
		"data.status=active",
		"links-to=epics/**/*.md",
	}
	orFlags := []string{
		"data.priority=high",
		"data.priority=critical",
	}
	notFlags := []string{
		"data.status=archived",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conditions, err := svc.ParseConditions(andFlags, orFlags, notFlags)
		_ = conditions
		_ = err
	}
}

// ============================================================================
// Memory Benchmarks
// ============================================================================

// BenchmarkTextSearch_Memory benchmarks memory usage of text search.
func BenchmarkTextSearch_Memory(b *testing.B) {
	svc := services.NewSearchService()
	notes := createBenchmarkNotes(10000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := svc.TextSearch("meeting", notes)
		_ = results
	}
}
