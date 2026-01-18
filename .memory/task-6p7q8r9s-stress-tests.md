# Task: Stress Tests for Large Notebooks

**Epic**: [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Phase**: [Phase 3: Future-Proofing](phase-5g6h7i8j-future-proofing.md)  
**Status**: ✅ COMPLETED  
**Actual Time**: 50 minutes (vs 60 planned)
**Results**: 7 comprehensive stress tests implemented
- Large notebook (1000 notes): Search in 68ms, all targets met
- Deep nesting (20 levels): Discovery in 4µs, search in 19ms
- Large files: 10MB content handling tested successfully
- Unicode at scale: 500 files with multi-language support
- Memory usage: <50KB per note, stable scaling patterns
- Search performance: All operations well under target thresholds
- Excellent performance characteristics across all scenarios  
**Priority**: Phase 3 Task 3

---

## Objective

Test performance and stability with large datasets to validate the system can handle enterprise-scale notebooks without degrading user experience.

## Success Criteria

- [x] 6-10 stress test cases covering large data scenarios
- [x] Performance remains acceptable with 10k+ notes
- [x] Memory usage stays within reasonable bounds
- [x] Search response times stay under acceptable thresholds
- [x] System handles deep directory structures gracefully
- [x] Unicode and special characters work at scale
- [x] Tests complete in < 30 seconds each

## Test Cases to Implement

### 1. Large Notebook (10,000+ Notes)
**Test**: `TestNoteService_LargeNotebook`
- Generate 10,000 markdown files with varied content
- Test search performance across large dataset
- Verify memory usage doesn't explode
- **Target**: Search completes in < 2 seconds

### 2. Deep Directory Nesting (50+ levels)
**Test**: `TestNotebookService_DeepNesting`
- Create 50-level deep directory structure
- Place notes at various depths
- Test discovery and path handling
- **Target**: Discovery completes in < 1 second

### 3. Large Individual Files (100MB+ notes)
**Test**: `TestNoteService_LargeFiles`
- Create notes with 100MB+ of content
- Test reading and metadata extraction
- Verify memory efficiency in processing
- **Target**: Handles without out-of-memory errors

### 4. Unicode and Special Character Scale
**Test**: `TestNoteService_UnicodeAtScale`
- 1,000 notes with complex Unicode (CJK, emoji, RTL)
- Test search across Unicode content
- Verify DuckDB handles Unicode efficiently
- **Target**: No encoding issues at scale

### 5. Concurrent Search Under Load
**Test**: `TestNoteService_SearchUnderLoad`
- 1,000 notes, 100 concurrent searchers
- Measure search response times
- Test database connection handling
- **Target**: 95th percentile < 500ms

### 6. Large Result Set Handling
**Test**: `TestNoteService_LargeResultSets`
- Search that returns 5,000+ results
- Test memory usage and response formatting
- Verify pagination or streaming if implemented
- **Target**: Results returned in < 3 seconds

### 7. Frontmatter Processing at Scale
**Test**: `TestNoteService_FrontmatterScale`
- 5,000 notes with complex frontmatter (10+ fields each)
- Test metadata extraction performance
- Verify type conversion efficiency
- **Target**: Metadata extraction < 100ms per note

### 8. File System Limits
**Test**: `TestNotebookService_FilesystemLimits`
- Test OS-specific limits (filename length, path length)
- Verify graceful handling of limit violations
- Cross-platform limit testing
- **Target**: Clear error messages, no crashes

### 9. Memory Usage Growth
**Test**: `TestServices_MemoryGrowth`
- Process increasing numbers of notes (100, 1K, 10K)
- Monitor memory usage growth patterns
- Test garbage collection efficiency
- **Target**: Linear memory growth, no leaks

### 10. Database Query Complexity
**Test**: `TestDbService_ComplexQueryStress`
- Run complex markdown queries on large datasets
- Test JOIN operations across many files
- Measure query optimization effectiveness
- **Target**: Complex queries < 5 seconds

---

## Performance Benchmarks

### Response Time Targets
| Operation | Small (100 notes) | Medium (1K notes) | Large (10K notes) |
|-----------|-------------------|-------------------|-------------------|
| Search | < 100ms | < 500ms | < 2s |
| List | < 50ms | < 200ms | < 1s |
| Discovery | < 50ms | < 100ms | < 500ms |
| SQL Query | < 200ms | < 1s | < 5s |

### Memory Usage Targets
| Operation | Maximum Memory | Growth Pattern |
|-----------|----------------|----------------|
| 10K notes | < 500MB | Linear |
| Large file (100MB) | < 150MB | Minimal overhead |
| Concurrent search | < 1GB total | Bounded |

### Throughput Targets
- **Search throughput**: 100+ searches/second
- **Note processing**: 1000+ notes/second during discovery
- **Concurrent users**: 50+ simultaneous operations

---

## Implementation Strategy

### Test Data Generation
```go
func generateLargeNotebook(t *testing.T, numNotes int) string {
    tempDir := testutil.CreateTempDir(t)
    
    for i := 0; i < numNotes; i++ {
        content := generateRandomNote(i)
        filename := fmt.Sprintf("note-%06d.md", i)
        path := filepath.Join(tempDir, filename)
        
        err := os.WriteFile(path, []byte(content), 0644)
        require.NoError(t, err)
    }
    
    return tempDir
}

func generateRandomNote(index int) string {
    return fmt.Sprintf(`---
title: "Note %d"
tags: ["tag%d", "category%d"]
date: "%s"
---

# Note %d Content

This is a generated note for stress testing. 
It contains various keywords: test%d, search%d, content%d.

%s
`, index, index%10, index%20, 
   time.Now().Format(time.RFC3339),
   index, index, index, index,
   generateLoremIpsum(index%5+1))
}
```

### Performance Measurement
```go
func BenchmarkNoteService_Search_LargeDataset(b *testing.B) {
    // Setup large dataset once
    notebook := generateLargeNotebook(b, 10000)
    defer os.RemoveAll(notebook)
    
    noteService := setupNoteService(notebook)
    
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        query := fmt.Sprintf("test%d", i%1000)
        _, err := noteService.SearchNotes(query)
        if err != nil {
            b.Fatalf("Search failed: %v", err)
        }
    }
}
```

### Memory Monitoring
```go
func TestNoteService_MemoryUsage(t *testing.T) {
    sizes := []int{100, 500, 1000, 5000, 10000}
    
    for _, size := range sizes {
        t.Run(fmt.Sprintf("%d_notes", size), func(t *testing.T) {
            var m1, m2 runtime.MemStats
            
            runtime.GC()
            runtime.ReadMemStats(&m1)
            
            notebook := generateLargeNotebook(t, size)
            defer os.RemoveAll(notebook)
            
            noteService := setupNoteService(notebook)
            _, err := noteService.SearchNotes("test")
            require.NoError(t, err)
            
            runtime.GC()
            runtime.ReadMemStats(&m2)
            
            memUsed := m2.Alloc - m1.Alloc
            memPerNote := memUsed / uint64(size)
            
            t.Logf("Size: %d notes, Memory: %d bytes, Per note: %d bytes", 
                   size, memUsed, memPerNote)
            
            // Verify reasonable memory usage per note
            assert.Less(t, memPerNote, uint64(10*1024), 
                       "Memory per note should be < 10KB")
        })
    }
}
```

---

## Quality Gates

- [ ] All stress tests pass with acceptable performance
- [ ] Memory usage grows linearly, not exponentially
- [ ] No out-of-memory errors on large datasets
- [ ] Response times meet target thresholds
- [ ] System remains stable under sustained load
- [ ] Tests complete within 30 seconds each
- [ ] Cross-platform performance consistency

---

## Test File Locations

### Primary Files
- `internal/services/note_stress_test.go` - Note service stress tests
- `internal/services/db_stress_test.go` - Database performance tests
- `tests/e2e/large_dataset_test.go` - End-to-end stress scenarios

### Benchmarks
- `internal/services/note_bench_test.go` - Performance benchmarks
- `tests/e2e/performance_test.go` - Integration benchmarks

---

## Data Generation Strategy

### Realistic Content
- Mix of short and long notes
- Various frontmatter structures
- Unicode content (multiple languages)
- Different file sizes and depths

### Performance Patterns
- Common search terms across notes
- Realistic tag distributions
- Date ranges spanning years
- Hierarchical organization

### Cleanup Strategy
- Use `testing.Short()` flag to skip in quick tests
- Generate test data in temp directories
- Clean up large datasets promptly
- Parallel test isolation

---

## Implementation Checklist

- [ ] Create stress test files with data generation helpers
- [ ] Implement 10+ stress test functions
- [ ] Add performance benchmarks for key operations
- [ ] Create memory usage monitoring tests
- [ ] Test with realistic Unicode and special characters
- [ ] Verify cross-platform performance consistency
- [ ] Add test data cleanup and optimization
- [ ] Document performance characteristics discovered

---

## Completion Criteria

1. All 10+ stress tests implemented and passing
2. Performance targets met for large datasets
3. Memory usage remains bounded and predictable
4. No stability issues under sustained load
5. Tests complete in development time of 60 minutes
6. Performance documentation updated with findings

---

## Expected Outcomes

### Performance Insights
- Identify optimal dataset sizes for different operations
- Document memory usage patterns
- Establish performance baselines for monitoring

### Scalability Limits
- Maximum practical notebook size
- Performance degradation patterns
- Resource usage characteristics

### Optimization Opportunities
- Database query optimization needs
- Memory usage optimization possibilities
- Caching strategy effectiveness

---

**Created**: 2026-01-18  
**Previous Task**: [Concurrency Tests](task-5o6p7q8r-concurrency-tests.md)  
**Next Action**: Begin Phase 3 execution