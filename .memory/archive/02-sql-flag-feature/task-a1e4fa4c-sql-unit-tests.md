# Task: Write SQL Unit Tests

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)  
**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 1 - Core Functionality (MVP)
**Priority**: HIGH
**Complexity**: High
**Estimated Time**: 90 minutes

## Objective

Write comprehensive unit tests for all SQL functionality to achieve >80% code coverage, ensuring query validation, execution, and display work correctly across all scenarios.

## Context

This is the testing task for the entire MVP feature. Tests must cover:
1. Query validation (valid/invalid queries)
2. Read-only DB connection
3. SQL safe execution (success/error/timeout)
4. Result display (empty/single/multiple rows)

Target coverage: >80% of new code

## Steps to Take

1. **Create test file**
   - File: `internal/services/note_test.go` (add to existing file if present)
   - Use Vitest pattern from project

2. **Test validateSQLQuery()**
   ```
   ✓ SELECT query allowed
   ✓ WITH (CTE) query allowed
   ✓ Lowercase select allowed
   ✓ INSERT query blocked
   ✓ UPDATE query blocked
   ✓ DELETE query blocked
   ✓ DROP query blocked
   ✓ ALTER query blocked
   ✓ CREATE query blocked
   ✓ TRUNCATE query blocked
   ✓ REPLACE query blocked
   ✓ ATTACH query blocked
   ✓ DETACH query blocked
   ✓ PRAGMA query blocked
   ✓ Error message includes keyword
   ✓ Case insensitive validation
   ```

3. **Test GetReadOnlyDB()**
   ```
   ✓ Returns valid connection
   ✓ Connection is read-only (write fails)
   ✓ Markdown extension loads
   ✓ Error handling for connection failure
   ✓ Properly closes on error
   ✓ Context cancellation works
   ```

4. **Test ExecuteSQLSafe()**
   ```
   ✓ Valid SELECT succeeds
   ✓ WITH query succeeds
   ✓ Invalid query blocked (validation error)
   ✓ Query execution error wrapped
   ✓ Timeout works (sets 30s limit)
   ✓ Returns correct structure []map[string]any
   ✓ Empty result set returns empty slice
   ✓ Single row handled correctly
   ✓ Multiple rows handled correctly
   ✓ Different data types handled
   ✓ nil values handled
   ✓ Connection closed on error
   ✓ Connection closed on success
   ```

5. **Test RenderSQLResults()**
   ```
   ✓ Empty results shows "No results"
   ✓ Single row displayed correctly
   ✓ Multiple rows displayed correctly
   ✓ Column headers shown
   ✓ Separator line shown
   ✓ Columns aligned properly
   ✓ Row count shown
   ✓ Different column widths handled
   ✓ Wide content handled
   ✓ Different data types formatted
   ✓ nil values formatted
   ✓ Stdout output is readable
   ```

6. **Add integration tests** (if time permits)
   ```
   ✓ End-to-end: SQL to display
   ✓ Real query against markdown files
   ✓ Error message quality
   ```

## Expected Outcomes

- [ ] All unit tests written and passing
- [ ] >80% code coverage for new methods
- [ ] Tests follow project patterns
- [ ] Tests are readable and maintainable
- [ ] Both positive and negative cases covered
- [ ] Edge cases handled
- [ ] Ready for CI/CD pipeline

## Acceptance Criteria

- [x] Test file created with proper structure
- [x] Tests for validateSQLQuery() - 16+ cases
- [x] Tests for GetReadOnlyDB() - 5+ cases
- [x] Tests for ExecuteSQLSafe() - 13+ cases
- [x] Tests for RenderSQLResults() - 10+ cases
- [x] All tests passing locally
- [x] Code coverage >80%
- [x] Tests follow project patterns
- [x] Proper assertions (expect, deep equal, etc.)
- [x] No test pollution (cleanup after each test)
- [x] Descriptive test names

## Implementation Notes

### Test File Structure
```go
import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestValidateSQLQuery(t *testing.T) {
    tests := []struct {
        name    string
        query   string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "SELECT allowed",
            query:   "SELECT * FROM table",
            wantErr: false,
        },
        {
            name:    "INSERT blocked",
            query:   "INSERT INTO table VALUES (1)",
            wantErr: true,
            errMsg:  "INSERT",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateSQLQuery(tt.query)
            if tt.wantErr {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Test Helpers Needed

```go
// Create test database for integration tests
func setupTestDB(t *testing.T) *DbService { ... }

// Create test markdown files
func createTestNotebook(t *testing.T) string { ... }

// Mock DisplayService output
func captureOutput(fn func()) string { ... }
```

### Coverage Measurement
```bash
# Run tests with coverage
mise run test -- --coverage

# Check specific file coverage
go test -coverprofile=coverage.out ./internal/services/
go tool cover -html=coverage.out
```

## Dependencies

- ✅ [task-4f209693-add-readonly-db.md](task-4f209693-add-readonly-db.md)
- ✅ [task-d4548dcd-sql-validation.md](task-d4548dcd-sql-validation.md)
- ✅ [task-bef53880-execute-sql-safe.md](task-bef53880-execute-sql-safe.md)
- ✅ [task-c7fc4f57-render-sql-results.md](task-c7fc4f57-render-sql-results.md)
- ✅ Testing framework (Vitest)
- ✅ Assertion library (testify)

## Blockers

- None identified (tests can be written in parallel with implementation)

## Time Estimate

- Test infrastructure setup: 20 minutes
- validateSQLQuery tests: 15 minutes
- GetReadOnlyDB tests: 15 minutes
- ExecuteSQLSafe tests: 25 minutes
- RenderSQLResults tests: 15 minutes
- Integration tests: 15 minutes
- Coverage review and tweaks: 10 minutes
- Total: 90 minutes

## Definition of Done

- [ ] All test cases implemented
- [ ] All tests passing locally
- [ ] Coverage >80% verified
- [ ] Tests follow project conventions
- [ ] No flaky tests
- [ ] Ready for CI/CD

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Links**:
- [Add ReadOnly DB](task-4f209693-add-readonly-db.md)
- [SQL Validation](task-d4548dcd-sql-validation.md)
- [Execute SQL Safe](task-bef53880-execute-sql-safe.md)
- [Render SQL Results](task-c7fc4f57-render-sql-results.md)
