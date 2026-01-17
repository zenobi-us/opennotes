# Task: Add NoteService.ExecuteSQLSafe() Method

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 1 - Core Functionality (MVP)
**Priority**: HIGH
**Complexity**: Medium
**Estimated Time**: 60 minutes

## Objective

Create the core `ExecuteSQLSafe()` method in NoteService that orchestrates validated SQL query execution with timeout protection, returning results as maps.

## Context

This is the main orchestrator method that:
1. Validates the user query using `validateSQLQuery()`
2. Gets a read-only connection from DbService
3. Executes with 30-second timeout
4. Converts rows to maps using existing `rowsToMaps()` helper
5. Returns results or error

This method is the bridge between CLI and database layers.

## Steps to Take

1. **Add ExecuteSQLSafe() method to NoteService**
   - File: `internal/services/note.go`
   - Signature: `func (s *NoteService) ExecuteSQLSafe(ctx context.Context, query string) ([]map[string]any, error)`

2. **Implement execution flow**
   - Call `validateSQLQuery(query)` - return error if invalid
   - Call `s.dbService.GetReadOnlyDB(ctx)` - get read-only connection
   - Defer connection close
   - Create new context with 30-second timeout
   - Execute query with `db.QueryContext(ctx, query)`
   - Defer rows close
   - Call `rowsToMaps(rows)` to convert results
   - Return results or error

3. **Extract or create rowsToMaps() helper**
   - Check if already exists in DbService
   - If not, create in NoteService (or shared utility)
   - Function should convert `*sql.Rows` to `[]map[string]interface{}`
   - Handle column scanning properly

4. **Add error handling**
   - Wrap validation errors
   - Wrap database errors with context
   - Wrap timeout errors properly
   - Close connection on error

5. **Add logging**
   - Log query execution at DEBUG level
   - Log errors at appropriate level

## Expected Outcomes

- [ ] ExecuteSQLSafe() method implemented
- [ ] Validates query before execution
- [ ] Gets read-only connection safely
- [ ] Executes with 30-second timeout
- [ ] Converts rows to maps correctly
- [ ] Returns clean error messages
- [ ] Properly closes all resources

## Acceptance Criteria

- [x] Method signature matches specification
- [x] Accepts query string and context
- [x] Returns `[]map[string]any` and error
- [x] Validates query (blocks dangerous queries)
- [x] Creates read-only connection
- [x] Sets 30-second timeout
- [x] Handles context cancellation
- [x] Closes connection on error
- [x] Converts rows to maps correctly
- [x] All resources properly deferred/cleaned up
- [x] Error messages are user-friendly
- [x] Query logging at DEBUG level
- [x] Handles empty result sets
- [x] Code compiles and lints

## Implementation Notes

### Method Flow
```go
func (s *NoteService) ExecuteSQLSafe(ctx context.Context, query string) ([]map[string]any, error) {
    // 1. Validate
    if err := validateSQLQuery(query); err != nil {
        return nil, err
    }
    
    // 2. Get connection
    db, err := s.dbService.GetReadOnlyDB(ctx)
    if err != nil {
        return nil, err
    }
    defer db.Close()
    
    // 3. Create timeout context
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // 4. Execute query
    rows, err := db.QueryContext(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("query execution failed: %w", err)
    }
    defer rows.Close()
    
    // 5. Convert to maps
    return rowsToMaps(rows)
}
```

### Error Cases to Handle

```
- Query validation fails → return validation error
- Read-only connection fails → wrap and return
- Query execution fails → wrap with "query execution failed: ..."
- Timeout occurs → context.DeadlineExceeded → user sees timeout
- Row scanning fails → wrap and return
- Empty result set → return empty slice, no error
```

### Testing Considerations

This method will be heavily tested in unit tests task. Consider:
- Valid SELECT query
- Invalid query (blocked keyword)
- Timeout scenario
- Error scenario
- Empty results
- Multiple rows with different types

## Dependencies

- ✅ [task-4f209693-add-readonly-db.md](task-4f209693-add-readonly-db.md) - GetReadOnlyDB() method
- ✅ [task-d4548dcd-sql-validation.md](task-d4548dcd-sql-validation.md) - validateSQLQuery() function
- ✅ rowsToMaps() helper (check if exists)

## Blockers

- Need to verify rowsToMaps() exists or can be extracted from DbService

## Time Estimate

- Implementation: 35 minutes
- Testing preparation: 15 minutes
- Review: 10 minutes
- Total: 60 minutes

## Definition of Done

- [ ] Method implemented and compiles
- [ ] Passes linting and formatting
- [ ] Handles all error cases properly
- [ ] Resources properly cleaned up
- [ ] Ready for unit testing
- [ ] Linked to unit test task

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Links**: 
- [Add ReadOnly DB](task-4f209693-add-readonly-db.md)
- [SQL Validation](task-d4548dcd-sql-validation.md)
- [SQL Unit Tests](task-a1e4fa4c-sql-unit-tests.md)
