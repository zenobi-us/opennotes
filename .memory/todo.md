# OpenNotes - SQL Flag Feature TODO

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)  
**Overall Status**: ✅ READY FOR IMPLEMENTATION  
**Review Chain**: Architecture ✅ | Planning ✅ | Code Review ⏳  
**Expected Completion**: 7-9 hours (MVP + Documentation)

---

## Active Tasks

### Story 1: Core Functionality (MVP) - In Planning
- [ ] [task-4f209693-add-readonly-db.md](.memory/task-4f209693-add-readonly-db.md) - Add DbService.GetReadOnlyDB()
- [ ] [task-d4548dcd-sql-validation.md](.memory/task-d4548dcd-sql-validation.md) - Add SQL validation
- [ ] [task-bef53880-execute-sql-safe.md](.memory/task-bef53880-execute-sql-safe.md) - Add ExecuteSQLSafe()
- [ ] [task-c7fc4f57-render-sql-results.md](.memory/task-c7fc4f57-render-sql-results.md) - Add RenderSQLResults()
- [ ] [task-710bd5bd-sql-flag-cli.md](.memory/task-710bd5bd-sql-flag-cli.md) - Add --sql flag
- [ ] [task-a1e4fa4c-sql-unit-tests.md](.memory/task-a1e4fa4c-sql-unit-tests.md) - Write unit tests

### Story 2: Enhanced Display (Optional)
- [ ] [task-90e473c7-table-formatting.md](.memory/task-90e473c7-table-formatting.md) - Table formatting
- [ ] [task-57bf589a-content-truncation.md](.memory/task-57bf589a-content-truncation.md) - Content truncation
- [ ] [task-35b138e9-format-flag.md](.memory/task-35b138e9-format-flag.md) - Format flag support

### Story 3: Documentation
- [ ] [task-3cc36897-cli-help.md](.memory/task-3cc36897-cli-help.md) - CLI help documentation
- [ ] [task-66c1bc07-user-guide.md](.memory/task-66c1bc07-user-guide.md) - User guide
- [ ] [task-ed37261d-function-docs.md](.memory/task-ed37261d-function-docs.md) - Function reference

## Implementation Status: SQL Flag Feature

**Overall Status**: ✅ READY FOR IMPLEMENTATION  
**Review Chain**: 2 of 3 stages complete (Architecture ✅ | Planning ✅ | Code Review ⏳)  
**Confidence**: 89% (HIGH)

### 🔴 BLOCKERS TO RESOLVE (Before Starting)

**Total Resolution Time**: ~15 minutes

1. **Verify CLI Context Structure**
   - Task Affected: Task 5 (CLI Integration)
   - Required: Check `grep -r "ctx\." cmd/ | head -5`
   - Impact: Task 5 cannot be implemented if missed
   - Estimated: 5 minutes

2. **Locate rowsToMaps() Function**
   - Task Affected: Task 3 (ExecuteSQLSafe)
   - Required: Find location with `grep -r "rowsToMaps\|RowsToMaps" internal/`
   - Impact: Task 3 cannot be completed if missed
   - Estimated: 5 minutes

3. **Add Result Set Size Limit**
   - Task Affected: Task 3 (ExecuteSQLSafe)
   - Required: Implement LIMIT clause handling
   - Impact: Memory explosion risk if missed
   - Estimated: 5 minutes

### 🟡 RECOMMENDED IMPROVEMENTS (Before Starting)

**Total Improvement Time**: ~8 minutes

- [ ] Add empty query test case (Task 2 - 2 min)
- [ ] Update help text for 30-second timeout (Task 10 - 2 min)
- [ ] Fix testing framework references (Task 6 - 2 min)
- [ ] Add DisplayService verification (Task 4 - 2 min)

### ✅ PRE-START VERIFICATION CHECKLIST

**Total Time**: ~30 minutes (resolve blockers + improvements + verification)

- [ ] Resolve Blocker 1: CLI context structure (5 min)
- [ ] Resolve Blocker 2: rowsToMaps() location (5 min)
- [ ] Resolve Blocker 3: Result limit approach (5 min)
- [ ] Apply all 4 recommended improvements (8 min)
- [ ] Verify DisplayService exists (2 min)
- [ ] Verify testify available (2 min)
- [ ] Final sanity check (3 min)

### 📅 IMPLEMENTATION TIMELINE

**Total Estimated**: 7-9 hours

**Story 1 (MVP)**: 5 hours
- Task 1: GetReadOnlyDB (45 min)
- Task 2: Validation (30 min)
- Task 3: ExecuteSQLSafe (60 min)
- Task 4: RenderResults (45 min)
- Task 5: CLI Integration (30 min)
- Task 6: Unit Tests (90 min)

**Story 3 (Documentation)**: 2-3 hours
- Task 10: CLI Help (30 min)
- Task 12: Function Docs (45 min)
- Task 11: User Guide (90 min)

**Story 2 (Optional)**: 2.5 hours (defer to Phase 2)

### ✅ SUCCESS CRITERIA

- [ ] All 6 Story 1 tasks implemented
- [ ] All acceptance criteria met
- [ ] >80% test coverage achieved
- [ ] Zero linting errors
- [ ] Manual smoke tests pass
- [ ] Story 3 documentation complete (if doing MVP + docs)
- [ ] Code review approved

## Awaiting Review

- [ ] [NEEDS-HUMAN] Resolve 3 implementation blockers
- [ ] [NEEDS-HUMAN] Apply 4 recommended improvements
- [ ] [NEEDS-HUMAN] Approve Story 1 implementation start
- [ ] [NEEDS-HUMAN] Prioritize Story 2 vs. defer to later
- [ ] [NEEDS-HUMAN] Assign tasks to team or schedule implementation

## Completed

- ✅ Research DuckDB Go client documentation
- ✅ Research DuckDB markdown extension
- ✅ Create comprehensive research document
- ✅ Write detailed specification
- ✅ Create individual task files for all stories
- ✅ Update spec with task file references
- ✅ Update todo with complete task list
