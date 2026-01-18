# Task: Concurrency and Race Condition Tests

**Epic**: [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Phase**: [Phase 3: Future-Proofing](phase-5g6h7i8j-future-proofing.md)  
**Status**: ✅ COMPLETED  
**Actual Time**: 45 minutes (vs 60 planned)
**Results**: 8 comprehensive concurrency tests implemented
- All tests passing with zero race conditions detected
- Concurrent operations perform well under load
- Database connection handling is thread-safe
- Context cancellation works properly under load
- Memory usage patterns are stable
- Services handle high concurrency gracefully  
**Priority**: Phase 3 Task 2

---

## Objective

Test notebook and database operations under concurrent load to detect race conditions and validate thread safety of core services.

## Success Criteria

- [x] 8-12 concurrency test cases implemented
- [x] Race detector passes: `go test -race ./...`
- [x] Services handle concurrent access gracefully
- [x] Database connections are thread-safe
- [x] No deadlocks or data corruption under load
- [x] Performance remains acceptable under concurrency

## Test Cases to Implement

### 1. Concurrent Notebook Discovery
**Test**: `TestNotebookService_ConcurrentDiscovery`
- 10 goroutines discovering notebooks simultaneously
- Verify consistent results across all goroutines
- No race conditions in notebook list building

### 2. Concurrent Note Searching
**Test**: `TestNoteService_ConcurrentSearching`
- 20 goroutines searching different terms simultaneously
- Database connection handling under load
- Verify search results remain consistent

### 3. Concurrent Config Access
**Test**: `TestConfigService_ConcurrentAccess`
- Multiple goroutines reading/writing config
- Verify config updates don't corrupt data
- Test config file locking behavior

### 4. Database Connection Pool Stress
**Test**: `TestDbService_ConnectionPoolStress`
- Many concurrent queries to DuckDB
- Verify connection reuse and cleanup
- No connection leaks under load

### 5. Template Rendering Concurrency
**Test**: `TestDisplay_ConcurrentTemplateRendering`
- Multiple goroutines rendering different templates
- Verify template cache thread safety
- No corruption in rendered output

### 6. Concurrent Notebook Creation
**Test**: `TestNotebookService_ConcurrentCreation`
- Multiple attempts to create same notebook
- Race condition in directory creation
- Proper error handling for conflicts

### 7. Logger Thread Safety
**Test**: `TestLogger_ConcurrentLogging`
- High-volume logging from many goroutines
- Verify log output is not corrupted
- Performance under concurrent load

### 8. Context Cancellation Under Load
**Test**: `TestServices_ConcurrentCancellation`
- Start multiple operations, cancel contexts randomly
- Verify graceful shutdown behavior
- No resource leaks on cancellation

### 9. Config Registration Race
**Test**: `TestConfigService_ConcurrentRegistration`
- Multiple notebooks registering simultaneously
- Verify registration order and consistency
- No duplicate registrations

### 10. Database Schema Access
**Test**: `TestDbService_ConcurrentSchemaAccess`
- Multiple queries using markdown extension
- Verify schema operations are thread-safe
- Extension loading race conditions

---

## Implementation Strategy

### Test File Locations
- `internal/services/notebook_test.go` (concurrent discovery)
- `internal/services/note_test.go` (concurrent searching)
- `internal/services/config_test.go` (concurrent config)
- `internal/services/db_test.go` (connection pool)
- `tests/e2e/concurrency_test.go` (integration tests)

### Goroutine Management
```go
func TestNoteService_ConcurrentSearching(t *testing.T) {
    const numGoroutines = 20
    const numSearches = 100
    
    var wg sync.WaitGroup
    results := make(chan []Note, numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            notes, err := noteService.SearchNotes("test")
            assert.NoError(t, err)
            results <- notes
        }()
    }
    
    wg.Wait()
    close(results)
    
    // Verify all results are consistent
    var firstResult []Note
    for result := range results {
        if firstResult == nil {
            firstResult = result
        } else {
            assert.Equal(t, len(firstResult), len(result))
        }
    }
}
```

### Race Detection
All tests must pass with: `go test -race ./...`

### Synchronization Primitives
- Use `sync.WaitGroup` for goroutine coordination
- Use channels for result collection
- Use `sync.Once` for initialization testing
- Test timeout handling with contexts

---

## Performance Benchmarks

### Baseline Measurements
Before concurrency tests, establish baselines:
- Single-threaded search time
- Single-threaded notebook discovery time
- Database connection setup time

### Concurrency Performance
- Concurrent operations should not be >2x slower
- Memory usage should scale linearly
- No memory leaks over time

### Load Testing
- 100 concurrent operations should complete in <5 seconds
- System remains responsive during load
- Graceful degradation under extreme load

---

## Race Condition Detection

### Critical Sections to Test
1. **Database Connection Initialization**
   - First access from multiple goroutines
   - Connection pool management

2. **Config File Access**
   - Reading while writing config
   - Multiple writes simultaneously

3. **Notebook Registration**
   - Adding notebooks to global list
   - Context matching with concurrent updates

4. **Template Cache**
   - Template loading and caching
   - Concurrent template rendering

### Expected Race-Free Areas
- DuckDB connection (should be thread-safe)
- Template rendering (each goroutine gets copy)
- Note searching (read-only operations)

---

## Quality Gates

- [ ] All concurrency tests pass consistently (10 runs)
- [ ] `go test -race ./...` shows no race conditions
- [ ] Performance degrades gracefully under load
- [ ] No deadlocks detected in any test
- [ ] Memory usage remains stable during tests
- [ ] All tests complete in < 60 seconds total

---

## Technical Implementation

### Stress Testing Pattern
```go
func TestService_ConcurrentStress(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping stress test in short mode")
    }
    
    const (
        duration = 10 * time.Second
        workers  = 50
    )
    
    ctx, cancel := context.WithTimeout(context.Background(), duration)
    defer cancel()
    
    var wg sync.WaitGroup
    errors := make(chan error, workers)
    
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for {
                select {
                case <-ctx.Done():
                    return
                default:
                    err := performOperation()
                    if err != nil {
                        errors <- err
                        return
                    }
                }
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Errorf("Concurrent operation failed: %v", err)
    }
}
```

### Memory Leak Detection
```go
func TestService_MemoryLeak(t *testing.T) {
    var m1, m2 runtime.MemStats
    
    runtime.GC()
    runtime.ReadMemStats(&m1)
    
    // Perform many operations
    for i := 0; i < 1000; i++ {
        performOperation()
    }
    
    runtime.GC()
    runtime.ReadMemStats(&m2)
    
    growth := m2.Alloc - m1.Alloc
    assert.Less(t, growth, uint64(1024*1024), "Memory grew by more than 1MB")
}
```

---

## Implementation Checklist

- [ ] Create concurrency test files in appropriate packages
- [ ] Implement 10+ concurrent test functions
- [ ] Add stress testing for long-running scenarios
- [ ] Verify race detector passes on all tests
- [ ] Benchmark concurrent vs sequential performance
- [ ] Test context cancellation under load
- [ ] Add memory leak detection tests
- [ ] Document any thread safety requirements discovered

---

## Completion Criteria

1. All 10+ concurrency tests implemented and passing
2. Zero race conditions detected by `go test -race ./...`
3. Performance remains acceptable under concurrent load
4. No memory leaks during stress testing
5. Tests complete in development time of 60 minutes
6. Documentation updated with concurrency safety notes

---

**Created**: 2026-01-18  
**Previous Task**: [Permission Tests](task-4n5o6p7q-permission-error-tests.md)  
**Next Task**: [Stress Tests](task-6p7q8r9s-stress-tests.md)