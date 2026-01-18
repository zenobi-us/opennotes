# Task: Add DbService.GetReadOnlyDB() Method

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)  
**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 1 - Core Functionality (MVP)
**Priority**: HIGH
**Complexity**: Medium
**Estimated Time**: 45 minutes

## Objective

Add a new `GetReadOnlyDB()` method to `DbService` that creates a read-only database connection for executing user SQL queries safely.

## Context

The SQL flag feature requires a separate read-only database connection to execute user-provided queries without risk of data modification. The existing `DbService` singleton maintains the primary connection; this method creates a new independent connection.

## Steps to Take

1. **Open the DbService file**
   - File: `internal/services/db.go`
   - Locate the `DbService` struct

2. **Implement GetReadOnlyDB() method**
   - Create new method with signature: `func (d *DbService) GetReadOnlyDB(ctx context.Context) (*sql.DB, error)`
   - Use `sql.Open("duckdb", "?access_mode=READ_ONLY")` to create connection
   - Load markdown extension with `LOAD markdown`
   - Add appropriate logging at DEBUG level
   - Handle errors properly with context

3. **Add error handling**
   - Wrap errors with context using `fmt.Errorf()`
   - Close connection if markdown extension fails to load

4. **Test the implementation**
   - Verify connection opens successfully
   - Verify markdown extension loads
   - Verify connection is read-only (write attempt should fail)

## Expected Outcomes

- [ ] New `GetReadOnlyDB(ctx context.Context)` method added to DbService
- [ ] Method creates independent read-only connection
- [ ] Markdown extension loads properly
- [ ] Errors include useful context
- [ ] Method follows existing code patterns and naming conventions

## Acceptance Criteria

- [x] Method signature matches specification
- [x] Returns `*sql.DB` and `error`
- [x] Accepts `context.Context` parameter for cancellation
- [x] Opens with `access_mode=READ_ONLY` flag
- [x] Loads markdown extension without INSTALL
- [x] Wraps errors with descriptive messages
- [x] Uses existing logging service
- [x] Follows Go error handling patterns
- [x] Code compiles without warnings
- [x] Passes linting checks

## Implementation Notes

### Connection String
```
?access_mode=READ_ONLY
```

### Markdown Load
```go
"LOAD markdown"
```

### Error Wrapping Pattern
```go
if err != nil {
    return nil, fmt.Errorf("failed to [action]: %w", err)
}
```

### Logging Pattern
```go
d.log.Debug().Msg("creating read-only database connection")
```

## Related Code

**Existing DbService pattern** (reference for style):
- See `internal/services/db.go` for logging and error patterns
- See `DbService.Query()` method for connection usage
- See `DbService.Init()` for markdown extension loading

**DuckDB Go docs**:
- https://duckdb.org/docs/stable/clients/go

## Dependencies

- ✅ DbService logging already initialized
- ✅ DuckDB markdown extension already researched
- ✅ Context cancellation support available

## Blockers

- None identified

## Time Estimate

- Implementation: 30 minutes
- Testing: 15 minutes
- Total: 45 minutes

## Definition of Done

- [ ] Code written and compiles
- [ ] Passes `mise run lint`
- [ ] Passes `mise run format`
- [ ] Ready for unit testing
- [ ] Linked to unit test task

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Links**: [SQL Unit Tests](task-a1e4fa4c-sql-unit-tests.md)
