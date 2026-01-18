# Epic: SQL Flag Feature for OpenNotes

**Status**: Ready for Implementation  
**Timeline**: 7-9 hours (MVP + Documentation)  
**Epic ID**: 2f3c4d5e

## Vision

Enable power users to query markdown notebooks using custom SQL queries via DuckDB's markdown extension, allowing complex data extraction and analysis without leaving the CLI.

## Goal

Expose DuckDB's powerful SQL query capabilities to OpenNotes users through a new `--sql` flag on the search command, enabling:
- Complex multi-condition searches
- Structured data extraction from markdown
- Code block analysis
- Metadata querying
- Word count and content analysis

## Success Criteria

- [x] Custom SQL queries execute safely against markdown notebooks
- [x] Query validation prevents destructive operations
- [x] Results display in readable table format
- [x] Zero data modification risk (read-only connection)
- [x] >80% test coverage achieved
- [x] CLI documentation complete
- [x] User guide with examples
- [x] Function reference available

## Phases

### Phase 1: Core Functionality (MVP) - **7-9 hours**
**Status**: Ready for Implementation  
**Tasks**: 6 tasks (45 min - 90 min each)

Core SQL execution infrastructure:
- [Task 1: Add DbService.GetReadOnlyDB()](task-4f209693-add-readonly-db.md) ⏳
- [Task 2: SQL Query Validation](task-d4548dcd-sql-validation.md) ⏳
- [Task 3: ExecuteSQLSafe() Method](task-bef53880-execute-sql-safe.md) ⏳
- [Task 4: Render SQL Results](task-c7fc4f57-render-sql-results.md) ⏳
- [Task 5: Add --sql Flag to CLI](task-710bd5bd-sql-flag-cli.md) ⏳
- [Task 6: Write Unit Tests](task-a1e4fa4c-sql-unit-tests.md) ⏳

**Success Metrics**:
- All 6 tasks complete
- All acceptance criteria met
- >80% test coverage
- Manual smoke tests pass

---

### Phase 2: Enhanced Display (Optional) - **2.5 hours**
**Status**: Deferred (Phase 2)  
**Tasks**: 3 tasks

Advanced formatting and display options:
- [Task 7: Table Formatting](task-90e473c7-table-formatting.md)
- [Task 8: Content Truncation](task-57bf589a-content-truncation.md)
- [Task 9: Format Flag Support](task-35b138e9-format-flag.md)

**Decision**: Defer to Phase 2 for MVP release

---

### Phase 3: Documentation - **2-3 hours**
**Status**: Recommended with MVP  
**Tasks**: 3 tasks

User-facing documentation:
- [Task 10: CLI Help Text](task-3cc36897-cli-help.md) ⏳
- [Task 11: User Guide](task-66c1bc07-user-guide.md) ⏳
- [Task 12: Function Reference](task-ed37261d-function-docs.md) ⏳

**Decision**: Include with MVP for complete feature release

---

## Key Components

### Specification
- **File**: [spec-a1b2c3d4-sql-flag.md](spec-a1b2c3d4-sql-flag.md)
- **Status**: Complete and approved
- **Content**: User stories, architecture, security model, testing strategy

### Research & Learning
- **Research**: [research-b8f3d2a1-duckdb-go-markdown.md](research-b8f3d2a1-duckdb-go-markdown.md)
- **Architecture Review**: [learning-8f6a2e3c-architecture-review-sql-flag.md](learning-8f6a2e3c-architecture-review-sql-flag.md)
- **Implementation Guidance**: [learning-7d9c4e1b-implementation-planning-guidance.md](learning-7d9c4e1b-implementation-planning-guidance.md)
- **Codebase Architecture**: [learning-5e4c3f2a-codebase-architecture.md](learning-5e4c3f2a-codebase-architecture.md)

## Implementation Requirements

### Prerequisites
- ✅ DuckDB markdown extension researched and validated
- ✅ Read-only connection approach designed
- ✅ Query validation rules documented
- ✅ Table formatting approach chosen
- ✅ Timeout strategy established (30 seconds)

### Dependencies
- DuckDB Go client (already integrated)
- Markdown extension (already loaded)
- Existing DbService infrastructure
- Existing DisplayService or similar
- Testing framework (Vitest, testify)

### Infrastructure Leveraged
- ✅ DbService with Query() method
- ✅ NoteService with Query() wrapper
- ✅ Markdown extension pre-loaded
- ✅ rowsToMaps() helper function
- ✅ DisplayService for output formatting

## Risk Mitigation

### Security Risks
- **Risk**: SQL injection or destructive queries
- **Mitigation**: Query validation + read-only connection (defense-in-depth)

### Performance Risks
- **Risk**: Long-running or expensive queries
- **Mitigation**: 30-second timeout, result set limits

### Compatibility Risks
- **Risk**: Breaking existing search functionality
- **Mitigation**: --sql flag optional, existing search path unchanged

## Timeline & Dependencies

**Critical Path**:
1. Phase 1 Core: Tasks 1-6 (5-6 hours) - CRITICAL
2. Phase 3 Documentation: Tasks 10-12 (2-3 hours) - RECOMMENDED
3. Phase 2 Enhanced Display: Tasks 7-9 - DEFERRED

**Expected Completion**: 7-9 hours (MVP + docs), 10-12 hours (all phases)

## Success Metrics

### Functional
- [ ] --sql flag works end-to-end
- [ ] All 4 user stories execute successfully
- [ ] Security validation blocks all dangerous queries
- [ ] Read-only connection prevents modifications
- [ ] 30-second timeout works correctly
- [ ] Results display readably

### Quality
- [ ] >80% unit test coverage
- [ ] All linting checks pass
- [ ] All acceptance criteria met
- [ ] Manual smoke tests pass
- [ ] No breaking changes

### Documentation
- [ ] CLI help text updated
- [ ] User guide written with examples
- [ ] Function reference complete
- [ ] README updated

## Lessons Learned Template

To be completed upon phase completion:
- [Space for learnings from Phase 1]
- [Space for learnings from Phase 3 if applicable]
- [Space for learnings from Phase 2 if implemented]

## Related Artifacts

- **TODO List**: [todo.md](todo.md)
- **Team Tracking**: [team.md](team.md)
- **Project Summary**: [summary.md](summary.md)

---

**Created**: 2026-01-17  
**Last Updated**: 2026-01-17  
**Status**: Ready for Implementation  

Next Action: Execute Pre-Start Verification Checklist from [todo.md](todo.md)
