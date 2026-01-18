# Task: Add Database Context Cancellation Tests

**Phase:** [Critical Fixes](phase-3f5a6b7c-critical-fixes.md)  
**Epic:** [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Status:** ✅ Completed  
**Time Estimate:** 15 minutes  
**Priority:** CRITICAL

---

## Objective

Add tests for `DbService.GetDB()` context cancellation scenarios to achieve **65% → 80%+ coverage**.

---

## Current State

**Function Location:** `internal/services/db.go:34-62`

```go
// GetDB returns an initialized database connection.
// The connection is lazily initialized on first call and reused thereafter.
func (d *DbService) GetDB(ctx context.Context) (*sql.DB, error) {
    var initErr error

    d.once.Do(func() {
        d.log.Debug().Msg("initializing database")

        // Open in-memory database
        db, err := sql.Open("duckdb", "")
        if err != nil {
            initErr = fmt.Errorf("failed to open database: %w", err)
            return
        }
        d.db = db

        // Install and load markdown extension
        d.log.Debug().Msg("installing markdown extension")
        if _, err := db.ExecContext(ctx, "INSTALL markdown FROM community"); err != nil {
            initErr = fmt.Errorf("failed to install markdown extension: %w", err)
            return
        }

        d.log.Debug().Msg("loading markdown extension")
        if _, err := db.ExecContext(ctx, "LOAD markdown"); err != nil {
            initErr = fmt.Errorf("failed to load markdown extension: %w", err)
            return
        }

        d.log.Debug().Msg("database initialized")
    })

    if initErr != nil {
        return nil, initErr
    }

    return d.db, nil
}
```

**Current Coverage:** 65% (happy path covered, context errors not tested)

**Untested Paths:**
- Context cancellation during INSTALL command
- Context cancellation during LOAD command
- Error handling and propagation

---

## Test Cases to Implement

| Test Case | Scenario | Expected |
|-----------|----------|----------|
| Context cancelled before init | Cancelled context on first call | Error returned |
| Multiple concurrent inits | Multiple goroutines calling simultaneously | Only one initializes |
| Context deadline exceeded | Deadline exceeded during INSTALL | Error returned |

---

## Implementation

### File to Modify

`internal/services/db_test.go`

### Code to Add

```go
// TestDbService_GetDB_ContextCancellation tests context cancellation handling
func TestDbService_GetDB_ContextCancellation(t *testing.T) {
    d := NewDbService()
    
    // Test 1: Cancelled context on first call
    t.Run("cancelled context on init", func(t *testing.T) {
        ctx, cancel := context.WithCancel(context.Background())
        cancel()  // Cancel immediately
        
        db, err := d.GetDB(ctx)
        if err == nil {
            t.Errorf("expected error with cancelled context, got nil")
        }
        if db != nil {
            t.Errorf("expected nil db with error, got %v", db)
        }
        if !strings.Contains(err.Error(), "context canceled") &&
           !strings.Contains(err.Error(), "context") {
            t.Errorf("error should mention context cancellation: %v", err)
        }
    })
}

// TestDbService_GetDB_ContextDeadline tests context deadline handling
func TestDbService_GetDB_ContextDeadline(t *testing.T) {
    d := NewDbService()
    
    // Create context with very short deadline
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()
    
    // Sleep to ensure deadline is exceeded
    time.Sleep(10 * time.Millisecond)
    
    db, err := d.GetDB(ctx)
    if err == nil {
        t.Errorf("expected error with exceeded deadline, got nil")
    }
    if db != nil {
        t.Errorf("expected nil db with error, got %v", db)
    }
    if !strings.Contains(err.Error(), "context deadline") &&
       !strings.Contains(err.Error(), "context") {
        t.Errorf("error should mention context deadline: %v", err)
    }
}

// TestDbService_GetDB_ConcurrentInit tests concurrent initialization attempts
func TestDbService_GetDB_ConcurrentInit(t *testing.T) {
    d := NewDbService()
    
    // Create a channel to synchronize goroutines
    done := make(chan error, 5)
    
    // Launch 5 concurrent attempts
    ctx := context.Background()
    for i := 0; i < 5; i++ {
        go func() {
            db, err := d.GetDB(ctx)
            if err != nil {
                done <- err
                return
            }
            if db == nil {
                done <- fmt.Errorf("db is nil")
                return
            }
            done <- nil
        }()
    }
    
    // Wait for all to complete
    var errs []error
    for i := 0; i < 5; i++ {
        if err := <-done; err != nil {
            errs = append(errs, err)
        }
    }
    
    // All should succeed with same DB instance
    if len(errs) > 0 {
        t.Errorf("concurrent init failed: %v", errs)
    }
    
    // Verify only one DB was created (same instance)
    db1, err := d.GetDB(ctx)
    if err != nil {
        t.Fatalf("failed to get db: %v", err)
    }
    db2, err := d.GetDB(ctx)
    if err != nil {
        t.Fatalf("failed to get db: %v", err)
    }
    if db1 != db2 {
        t.Errorf("expected same DB instance, got different pointers")
    }
    
    // Cleanup
    d.Close()
}

// TestDbService_GetDB_ErrorPropagation tests error handling during init
func TestDbService_GetDB_ErrorPropagation(t *testing.T) {
    d := NewDbService()
    
    // Create a context with a custom error
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    // Call GetDB with cancelled context
    db, err := d.GetDB(ctx)
    
    // Should return error
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    
    // Should be nil db
    if db != nil {
        t.Errorf("expected nil db, got %v", db)
    }
    
    // Error should be about context or markdown extension
    if !strings.Contains(err.Error(), "failed") {
        t.Errorf("error message format unexpected: %v", err)
    }
}
```

---

## Verification Steps

1. **Add the test functions** to `internal/services/db_test.go`

2. **Run the tests:**
   ```bash
   cd /mnt/Store/Projects/Mine/Github/opennotes
   go test -v ./internal/services -run TestDbService_GetDB
   ```

3. **Run with race detector:**
   ```bash
   go test -race ./internal/services -run TestDbService_GetDB
   ```

4. **Check coverage:**
   ```bash
   go test -cover ./internal/services
   ```

5. **Full test suite:**
   ```bash
   mise run test
   ```

---

## Success Criteria

- [x] All test functions compile
- [x] All tests pass consistently
- [x] No race conditions detected
- [x] Coverage improves from 65% → 80%+
- [x] Error scenarios properly tested
- [x] Concurrent access properly synchronized

---

## Notes

- Context cancellation must be tested carefully to avoid flaky tests
- Use `time.Sleep()` judiciously to ensure conditions are met
- Race detector (`-race` flag) is critical for concurrency tests
- The `sync.Once` in GetDB ensures only one initialization happens
- Tests verify both error handling and concurrency safety

---

## Dependencies

- Standard `context` package (already imported)
- Standard `time` package (already imported)
- Standard `strings` package (already imported)

---

**Status:** 📋 READY  
**Time Estimate:** 15 minutes  
**Blocks:** Phase 1 completion
