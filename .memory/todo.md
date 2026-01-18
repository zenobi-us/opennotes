# OpenNotes - SQL Flag Feature TODO

**Epic**: [SQL Flag Feature](epic-2f3c4d5e-sql-flag-feature.md)  
**Overall Status**: ✅ READY FOR IMPLEMENTATION  
**Review Chain**: Architecture ✅ | Planning ✅ | Code Review ⏳  
**Expected Completion**: 7-9 hours (MVP + Documentation)

---

## Active Tasks

### Story 1: Core Functionality (MVP) - 50% Complete (3/6 Tasks)
- [x] [task-4f209693-add-readonly-db.md](.memory/task-4f209693-add-readonly-db.md) - Add DbService.GetReadOnlyDB() ✅
- [x] [task-d4548dcd-sql-validation.md](.memory/task-d4548dcd-sql-validation.md) - Add SQL validation ✅
- [x] [task-bef53880-execute-sql-safe.md](.memory/task-bef53880-execute-sql-safe.md) - Add ExecuteSQLSafe() ✅
- [ ] [task-c7fc4f57-render-sql-results.md](.memory/task-c7fc4f57-render-sql-results.md) - Add RenderSQLResults() (IN PROGRESS)
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

### ✅ BLOCKERS RESOLVED

**Total Resolution Time**: 15 minutes ✅

1. ✅ **CLI Context Structure** - Services via global variables (cmd/root.go)
2. ✅ **rowsToMaps() Function** - Located in internal/services/db.go (ready to use)
3. ✅ **Result Set Size Limit** - LIMIT injection strategy confirmed in validation

### ✅ RECOMMENDED IMPROVEMENTS VERIFIED

**Total Improvement Time**: 8 minutes ✅

- ✅ Empty query test case - Add to Task 6 (SQL Unit Tests)
- ✅ Help text for timeout - Update Task 10 (CLI Help)
- ✅ Testing framework references - testify v1.11.1 available
- ✅ DisplayService verification - display.go fully functional

### ✅ PRE-START VERIFICATION COMPLETE

**Total Time**: 30 minutes ✅

- ✅ Blocker 1: CLI context structure (5 min)
- ✅ Blocker 2: rowsToMaps() location (5 min)
- ✅ Blocker 3: Result limit approach (5 min)
- ✅ Applied all 4 recommended improvements (8 min)
- ✅ Verified DisplayService exists (2 min)
- ✅ Verified testify available (2 min)
- ✅ Test suite verification: 123 tests PASS, 95%+ coverage (5 sec)

**See**: [verification-pre-start-checklist.md](.memory/verification-pre-start-checklist.md) for full details

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

## Bug Reports & New Tasks

- [x] [task-b5c8a9f2-notes-list-format.md](.memory/task-b5c8a9f2-notes-list-format.md) - **✅ COMPLETE**: `opennotes notes list` now formats output with title and filepath
- [ ] [task-c03646d9-clipboard-filename-slugify.md](.memory/task-c03646d9-clipboard-filename-slugify.md) - **BUG**: Clipboard temp filename should be slugified (not UUID format)

## Ready for Implementation

✅ **ALL BLOCKERS RESOLVED** - READY TO START STORY 1

**Next Steps**:
1. Start implementing Story 1 (Core MVP - 6 tasks, ~5 hours)
2. Follow task order: Task 1 → Task 2 → Task 3 → Task 4 → Task 5 → Task 6
3. Tests should pass at each step
4. Update todo.md as tasks complete

**Optional**:
- [ ] Story 2 (Enhanced Display) - defer to Phase 2 unless time permits
- [ ] Story 3 (Documentation) - include with MVP if shipping

## Completed

- ✅ Research DuckDB Go client documentation
- ✅ Research DuckDB markdown extension
- ✅ Create comprehensive research document
- ✅ Write detailed specification
- ✅ Create individual task files for all stories
- ✅ Update spec with task file references
- ✅ Update todo with complete task list
