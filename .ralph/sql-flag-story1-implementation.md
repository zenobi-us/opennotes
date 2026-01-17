# SQL Flag Feature - Story 1 Implementation

## Goals
Implement core SQL flag functionality for OpenNotes search command. Add `--sql` flag allowing custom SQL queries on notebook markdown data with proper security, timeout, and result limits.

## Story Tasks

### Task 1: DbService.GetReadOnlyDB() (45 min) ✅ COMPLETE
Add method to create read-only database connection for safe query execution

**Acceptance Criteria**:
- [x] New method `GetReadOnlyDB()` in DbService
- [x] Returns separate *sql.DB connection
- [x] Uses same markdown extension setup
- [x] Proper error handling
- [x] Unit tests pass (8 tests)

**File**: `.memory/task-4f209693-add-readonly-db.md`

### Task 2: ValidateSQL() (30 min) ✅ COMPLETE
Add SQL validation: SELECT/WITH only, no DDL/DML

**Acceptance Criteria**:
- [x] Validates query syntax
- [x] Rejects DROP/DELETE/UPDATE/CREATE/INSERT
- [x] Allows SELECT and WITH
- [x] Handles empty queries
- [x] Unit tests pass (25 tests)

**File**: `.memory/task-d4548dcd-sql-validation.md`

### Task 3: ExecuteSQLSafe() (60 min) ✅ COMPLETE
Execute validated queries with timeout and result conversion

**Acceptance Criteria**:
- [x] Accepts query string
- [x] Uses read-only connection (Task 1)
- [x] Validates query (Task 2)
- [x] 30-second timeout context
- [x] Returns []map[string]interface{}
- [x] Uses rowsToMaps() converter
- [x] Error handling for timeouts
- [x] Unit tests pass (10 tests)

**File**: `.memory/task-bef53880-execute-sql-safe.md`

### Task 4: RenderSQLResults() (45 min) 🔄 IN PROGRESS
Format query results for terminal display

**Acceptance Criteria**:
- [ ] Accept []map[string]interface{}
- [ ] Generate table format output
- [ ] Handle empty results
- [ ] Truncate long cells (DisplayService)
- [ ] Column alignment
- [ ] Unit tests pass

**File**: `.memory/task-c7fc4f57-render-sql-results.md`

### Task 5: CLI --sql Flag (30 min)
Add --sql flag to search command, integrate all components

**Acceptance Criteria**:
- [ ] Flag added to searchCmd
- [ ] Takes SQL query as string
- [ ] Calls ExecuteSQLSafe()
- [ ] Renders output with RenderSQLResults()
- [ ] Help text updated
- [ ] Integration tests pass

**File**: `.memory/task-710bd5bd-sql-flag-cli.md`

### Task 6: Unit Tests (90 min)
Comprehensive test coverage for all components

**Acceptance Criteria**:
- [ ] Tests for GetReadOnlyDB()
- [ ] Tests for ValidateSQL() (including edge cases)
- [ ] Tests for ExecuteSQLSafe()
- [ ] Tests for RenderSQLResults()
- [ ] Tests for CLI integration
- [ ] >80% coverage
- [ ] All tests pass

**File**: `.memory/task-a1e4fa4c-sql-unit-tests.md`

## Implementation Notes

### Key Patterns
- Use global services pattern (DbService, NotebookService)
- Follow existing Go conventions in codebase
- Use testify for assertions
- Import from internal/services/

### Critical References
- **CLI Context**: Global variables in cmd/root.go
- **rowsToMaps()**: internal/services/db.go (ready to use)
- **DisplayService**: internal/services/display.go (Render method)
- **Logger**: services.Log("context")

### Dependencies
None - all infrastructure already exists!

## Success Criteria

- [ ] All 6 tasks completed
- [ ] 123 existing tests still pass
- [ ] >80% test coverage
- [ ] Zero linting errors
- [ ] Manual smoke test: `opennotes notes search --sql "SELECT * FROM markdown LIMIT 5"`
- [ ] Help text accurate

## Timeline
- **Total**: ~5 hours
- **Per task**: See task breakdown above
- **Parallel**: No dependencies between tasks except Task 1 → Task 3

## Verification Checklist
- [ ] Pre-start verification complete (see verification-pre-start-checklist.md)
- [ ] All blockers resolved
- [ ] Code editor open with Go support
- [ ] Test suite baseline verified (123 tests passing)
- [ ] Ready to write first test