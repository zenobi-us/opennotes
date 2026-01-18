# Epic: Test Coverage Improvement & Error Handling

**Status:** In Progress  
**Timeline:** Q1 2026 (1-2 weeks)  
**Owner:** Code Quality Team  
**Created:** 2026-01-18 20:32 GMT+10:30

---

## Vision

Improve test coverage from 73.3% to 80%+ by adding strategic tests for error handling, edge cases, and critical paths. This epic ensures production confidence in error scenarios and prevents user-facing bugs.

---

## Success Criteria

- [x] Coverage: 73.3% → 80%+ (internal packages)
- [ ] All critical gaps fixed (ValidatePath, templates, context errors)
- [ ] Error handling tested (file exists, permission denied, invalid inputs)
- [ ] Edge cases covered (unicode, special characters, malformed data)
- [ ] No test flakes or brittleness
- [ ] All new tests follow project patterns
- [ ] Documentation updated with testing guidelines
- [ ] Zero linting/format errors

---

## Metrics

| Metric | Current | Target | Gap |
|--------|---------|--------|-----|
| Overall Coverage | 73.3% | 80% | 6.7% |
| Services | 83.5% | 85%+ | 1.5% |
| Core | 82.6% | 85%+ | 2.4% |
| Error Paths | 30-40% | 80%+ | 40-50% |
| Functions at 0% | 1 | 0 | 1 |
| Functions at <70% | 6 | 0 | 6 |

---

## Phases

### Phase 1: Critical Fixes (Immediate) ⏳
**Duration:** 30 minutes  
**Status:** Not Started  
[Phase Details](./phase-3f5a6b7c-critical-fixes.md)

Quick wins to fix the most important gaps:
- Fix ValidatePath() (0% coverage)
- Test template error paths
- Test DB context cancellation

**Expected Outcome:**
- 3 critical gaps fixed
- Coverage: 73% → 75%
- Foundation for Phase 2

---

### Phase 2: Core Improvements 🔄
**Duration:** 2 hours  
**Status:** Not Started  
[Phase Details](phase-4e5f6a7b-core-improvements.md)

High-value additions focusing on error scenarios and edge cases:
- Command error integration tests
- SearchNotes edge cases
- ObjectToFrontmatter edge cases

**Expected Outcome:**
- Coverage: 75% → 80%
- Production-ready confidence
- Error handling validated

---

### Phase 3: Future-Proofing (Optional) ⏳
**Duration:** 2-3 hours  
**Status:** Not Started  
[Phase Details](phase-5g6h7i8j-future-proofing.md)

Stress and concurrency tests for enterprise readiness:
- Permission/filesystem error scenarios
- Concurrency tests for race conditions
- Stress tests (large notebooks)

**Expected Outcome:**
- Coverage: 80% → 85%
- Enterprise-ready confidence
- Performance validated

---

## Current Test State Analysis

### What's Being Tested Well ✅

- **Core validation** (82-83% coverage)
  - Notebook name validation
  - Note name validation
  - String utilities (slugify, dedent)

- **Service layer** (83.5% coverage)
  - Database operations with markdown extension
  - Note searching and filtering
  - Notebook discovery and management
  - Logger functionality
  - Display/rendering

- **E2E workflows** (28 tests)
  - Creating notebooks and notes
  - Listing and searching
  - SQL queries
  - Config management

### Critical Gaps ❌

| Function | Coverage | Issue | Priority |
|----------|----------|-------|----------|
| schema.ValidatePath | 0% | No tests at all | CRITICAL |
| db.GetDB | 65% | Context errors not covered | HIGH |
| templates.TuiRender | 60% | Fallback paths untested | HIGH |
| note.SearchNotes | 76.4% | Some edge cases missing | MEDIUM |
| Command errors | ~30% | Only happy paths tested | HIGH |
| Special characters | ~50% | Unicode/emoji untested | MEDIUM |
| Permissions | ~10% | Filesystem errors untested | MEDIUM |

---

## Implementation Strategy

### By Risk Level

**CRITICAL (Fix Now):**
1. ValidatePath() - security implications, only 4 lines
2. Command error handling - silent failures in production
3. Template error paths - fallback behavior unknown

**HIGH (Add Soon):**
4. DB context cancellation - concurrent initialization risks
5. SearchNotes edge cases - incorrect results possible
6. Frontmatter edge cases - type conversion errors

**MEDIUM (Nice to Have):**
7. Special character handling - unicode filename edge cases
8. Permission errors - graceful failure modes
9. Stress tests - performance under load

### Test Organization Pattern

Each phase will add tests organized by:
- **Unit tests** in existing `*_test.go` files
- **Integration tests** as separate test files when needed
- **E2E tests** in `tests/e2e/` for workflow validation
- **Helpers** in `internal/testutil/` for shared fixtures

### Quality Gates

Before considering a test complete:
- [ ] Test passes consistently (no flakes)
- [ ] Follows project naming conventions
- [ ] Uses established fixtures and helpers
- [ ] Covers both success and failure paths
- [ ] Error messages are clear and actionable
- [ ] Code review approved
- [ ] Coverage targets met

---

## Dependencies

- Go 1.21+ (already in place)
- Testify v1.11.1 (already in use)
- DuckDB driver (already in place)
- No external test dependencies needed

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Test false positives | LOW | Use deterministic tests, proper cleanup |
| Test performance regression | LOW | Current suite runs in 2 seconds |
| Coverage metrics misleading | LOW | Focus on error paths, not just line coverage |
| Breaking existing tests | LOW | Run full suite after each change |
| Time overruns | MEDIUM | Phase 1 is fixed 30 min, Phase 2 is 2 hours |

---

## Related Documentation

- **Architecture Review**: [learning-8f6a2e3c-architecture-review-sql-flag.md](.memory/learning-8f6a2e3c-architecture-review-sql-flag.md)
- **Codebase Architecture**: [learning-5e4c3f2a-codebase-architecture.md](.memory/learning-5e4c3f2a-codebase-architecture.md)
- **Review Analysis**: [review-cleanup-report.md](.memory/review-cleanup-report.md)

---

## Success Indicators

### By Phase

- **Phase 1 (30 min):** 
  - ✓ ValidatePath() at 100% coverage
  - ✓ Template errors tested
  - ✓ All 3 tests pass

- **Phase 2 (2 hours):**
  - ✓ Coverage reaches 80%
  - ✓ 10-15 new tests added
  - ✓ All error paths tested

- **Phase 3 (2-3 hours):**
  - ✓ Coverage reaches 85%
  - ✓ Concurrency tests pass
  - ✓ Stress tests validate performance

### Overall

- [ ] No test flakes in CI
- [ ] Code coverage dashboard shows improvement
- [ ] Error logs indicate better error handling
- [ ] Team confidence in edge case handling increases
- [ ] Fewer customer-reported bugs related to tested areas

---

## Timeline & Milestones

```
Day 1:   Phase 1 (Critical Fixes) - 30 min
         Phase 2 (Core Improvements) - 2 hours
         Ready for production
         
Day 2-3: Phase 3 (Optional) - 2-3 hours
         Enterprise-ready
         
Day 4-5: Code review, refinement, documentation
```

---

## Team

| Role | Assigned | Status |
|------|----------|--------|
| Lead | TBD | Needs Assignment |
| QA | TBD | Needs Assignment |
| Code Review | TBD | Needs Assignment |

---

## Files & Structure

```
.memory/
├── epic-7a2b3c4d-test-improvement.md          # This file
├── phase-3f5a6b7c-critical-fixes.md           # Phase 1
├── phase-4e5f6a7b-core-improvements.md        # Phase 2
├── phase-5g6h7i8j-future-proofing.md          # Phase 3
├── task-8h9i0j1k-validate-path-tests.md       # Phase 1 Task 1
├── task-9i0j1k2l-template-error-tests.md      # Phase 1 Task 2
├── task-0j1k2l3m-db-context-tests.md          # Phase 1 Task 3
├── task-1k2l3m4n-command-error-tests.md       # Phase 2 Task 1
├── task-2l3m4n5o-search-edge-cases.md         # Phase 2 Task 2
├── task-3m4n5o6p-frontmatter-edge-cases.md    # Phase 2 Task 3
└── task-4n5o6p7q-permission-error-tests.md    # Phase 3 Task 1
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-18 | Initial epic creation with phases and assessment |

---

## Next Steps

1. ✅ Create this epic file
2. ⏳ Create Phase 1 details file
3. ⏳ Create Phase 2 details file
4. ⏳ Create Phase 3 details file
5. ⏳ Create individual task files for all tests
6. ⏳ Update todo.md with test improvement tasks
7. ⏳ Begin Phase 1 implementation

---

**Last Updated:** 2026-01-18 20:32 GMT+10:30  
**Epic Status:** 🟡 CREATED - Ready for Phase 1 kickoff
