# Task: Add --sql Flag to Search Command

**Spec**: [SQL Flag Specification](spec-a1b2c3d4-sql-flag.md)
**Story**: Story 1 - Core Functionality (MVP)
**Priority**: HIGH
**Complexity**: Low
**Estimated Time**: 30 minutes

## Objective

Add the `--sql` flag to the search command that accepts a SQL query string and integrates it with the core SQL execution and display components.

## Context

This is the CLI integration task that ties together:
- Query validation (task-d4548dcd)
- Read-only DB connection (task-4f209693)
- SQL execution (task-bef53880)
- Result display (task-c7fc4f57)

The search command should check for `--sql` flag early and bypass normal search logic if present.

## Steps to Take

1. **Locate search command file**
   - File: `cmd/search.go` (or similar, check project structure)

2. **Add --sql flag definition**
   - Add string flag using command framework
   - Flag name: "sql"
   - Description: "Execute custom SQL query against notes"
   - Should accept arbitrary SQL string

3. **Add early check in command handler**
   - Check if `--sql` flag is provided (not empty)
   - If provided, run SQL path instead of normal search

4. **Implement SQL path**
   - Get `NoteService` from context (already available)
   - Get `DisplayService` from context (already available)
   - Call `noteService.ExecuteSQLSafe(ctx, query)`
   - If error, return with "SQL query failed: <error>"
   - Call `displayService.RenderSQLResults(results)`
   - Return (skip normal search logic)

5. **Update help text**
   - Ensure flag shows in `--help` output
   - Add usage example in help

6. **Add error handling**
   - Wrap query errors with context
   - Display errors appropriately

## Expected Outcomes

- [ ] `--sql` flag added to search command
- [ ] Flag accepts query string
- [ ] Flag bypasses normal search logic
- [ ] Calls ExecuteSQLSafe() with user query
- [ ] Renders results with RenderSQLResults()
- [ ] Errors are handled gracefully
- [ ] Help text is updated

## Acceptance Criteria

- [x] Flag accepts string parameter
- [x] Flag is optional (normal search works without it)
- [x] When provided, executes SQL path
- [x] SQL path calls NoteService.ExecuteSQLSafe()
- [x] SQL path calls DisplayService.RenderSQLResults()
- [x] Errors are wrapped with "SQL query failed: ..."
- [x] Help text includes --sql flag
- [x] Flag works end-to-end
- [x] No breaking changes to existing search functionality
- [x] Code compiles and lints

## Implementation Notes

### Flag Definition Pattern

```go
// In command definition or handler setup
sqlQuery := cmd.Flags().String(
    "sql",
    "",
    "Execute custom SQL query against notes",
)
```

### Command Handler Pattern

```go
func (ctx) {
    // Early check for SQL flag
    if *sqlQuery != "" {
        // Execute SQL path
        results, err := noteService.ExecuteSQLSafe(ctx, *sqlQuery)
        if err != nil {
            return fmt.Errorf("SQL query failed: %w", err)
        }
        return displayService.RenderSQLResults(results)
    }
    
    // Normal search path continues...
}
```

### Service Access

- `ctx.store.noteService` - NoteService instance
- `ctx.store.displayService` - DisplayService instance (if available)
- Check Clerc documentation for exact context structure in project

## Dependencies

- ✅ [task-4f209693-add-readonly-db.md](task-4f209693-add-readonly-db.md) - GetReadOnlyDB()
- ✅ [task-d4548dcd-sql-validation.md](task-d4548dcd-sql-validation.md) - validateSQLQuery()
- ✅ [task-bef53880-execute-sql-safe.md](task-bef53880-execute-sql-safe.md) - ExecuteSQLSafe()
- ✅ [task-c7fc4f57-render-sql-results.md](task-c7fc4f57-render-sql-results.md) - RenderSQLResults()

## Blockers

- Need to verify exact flag definition pattern for project's CLI framework

## Time Estimate

- Implementation: 20 minutes
- Testing: 5 minutes
- Review: 5 minutes
- Total: 30 minutes

## Definition of Done

- [ ] Flag added to search command
- [ ] SQL path implemented and working
- [ ] All dependencies available
- [ ] Code compiles and passes linting
- [ ] Manual testing confirms working
- [ ] Ready for unit tests

---

**Created**: 2026-01-17
**Status**: Awaiting Start
**Links**:
- [Add ReadOnly DB](task-4f209693-add-readonly-db.md)
- [SQL Validation](task-d4548dcd-sql-validation.md)
- [Execute SQL Safe](task-bef53880-execute-sql-safe.md)
- [Render SQL Results](task-c7fc4f57-render-sql-results.md)
- [SQL Unit Tests](task-a1e4fa4c-sql-unit-tests.md)
