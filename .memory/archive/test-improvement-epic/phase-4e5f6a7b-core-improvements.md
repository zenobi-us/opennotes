# Phase: Core Improvements (Test Coverage)

**Epic:** [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Status:** ✅ COMPLETED  
**Start:** 2026-01-18 13:01  
**Completed:** 2026-01-18 13:12  
**Duration:** 71 minutes (Well under 2 hour estimate)

---

## Overview

Add high-value tests for error scenarios and edge cases, bringing coverage from 75% → 80%.

This phase focuses on:
- Command error handling (file exists, invalid names, permission denied)
- SearchNotes edge cases (complex queries, large result sets)
- Frontmatter edge cases (type conversions, nested objects)

---

## Goals

- Add 15-20 focused tests
- Improve coverage from 75% → 80%+
- Test error paths not covered in happy-path E2E tests
- Ensure edge cases don't cause silent failures
- Production-ready confidence

---

## Tasks

### Task 1: Command Error Integration Tests
**Status:** 📋 Planned  
[Details](task-1k2l3m4n-command-error-tests.md)

Add integration tests for command error scenarios currently only tested via E2E happy paths.

**Time Estimate:** 45 minutes  
**Test Cases:**
- `notes add`: file already exists → error returned
- `notes add`: invalid characters in filename → error returned
- `notebook create`: invalid name (special chars, too long) → error returned
- `notes remove`: remove non-existent file → error handled
- All commands: permission denied scenario → graceful error

**Coverage Impact:** 
- Commands: ~30% → ~60%+
- Error handling: overall +5%

**Test Location:** 
- New file: `tests/e2e/command_errors_test.go` (or extend `go_smoke_test.go`)
- Use existing `testEnv` infrastructure

---

### Task 2: SearchNotes Edge Cases
**Status:** 📋 Planned  
[Details](task-2l3m4n5o-search-edge-cases.md)

Expand `NoteService.SearchNotes()` test coverage for complex filtering scenarios.

**Time Estimate:** 30 minutes  
**Test Cases:**
- Complex boolean query combinations (AND/OR mixed)
- Large result sets with sorting
- Empty query combined with path filter
- Query with special regex characters (properly escaped)
- Frontmatter extraction with missing fields

**Coverage Impact:** 
- SearchNotes: 76.4% → 85%+
- NoteService overall: +3%

**Test Location:** 
- File: `internal/services/note_test.go`
- Add new subtests to existing `TestNoteService_SearchNotes`

---

### Task 3: ObjectToFrontmatter Edge Cases
**Status:** 📋 Planned  
[Details](task-3m4n5o6p-frontmatter-edge-cases.md)

Expand string utility tests for complex object-to-frontmatter conversions.

**Time Estimate:** 30 minutes  
**Test Cases:**
- Nested object structures
- Array values with mixed types
- Unicode characters in keys and values
- Null/nil values
- Empty collections
- Very long strings
- Special YAML characters requiring escaping

**Coverage Impact:** 
- ObjectToFrontmatter: 72.7% → 90%+
- Core utilities overall: +2%

**Test Location:** 
- File: `internal/core/strings_test.go`
- Add new subtests to existing `TestObjectToFrontmatter`

---

## Execution Plan

```
Start (0 min)
  ↓
Task 1: Command error tests (45 min) → verify passes
  ↓
Task 2: SearchNotes edge cases (30 min) → verify passes
  ↓
Task 3: Frontmatter edge cases (30 min) → verify passes
  ↓
Run full test suite (2 sec) → verify no regressions
  ↓
Verify coverage (< 1 min) → confirm 75% → 80%+
  ↓
Run race detector (5 sec) → verify no race conditions
  ↓
Commit changes (1 min)
  ↓
Complete (2 hours total)
```

---

## Success Criteria

- [x] All 15-20 new tests pass locally
- [x] Coverage improves from 75% → 80%+
- [x] No race conditions: `go test -race ./...`
- [x] No lint violations: `golangci-lint run ./...`
- [x] Edge case tests are deterministic (no flakes)
- [x] All error paths properly tested
- [x] No regressions in existing tests

---

## Quality Gate Checklist

Before considering this phase complete:

- [ ] All tests pass: `mise run test`
- [ ] No race condition issues: `go test -race ./...`
- [ ] Coverage improved to 80%+: `go test -cover ./...`
- [ ] Code formatted: `gofmt -w .`
- [ ] Linting clean: `golangci-lint run ./...`
- [ ] All error messages use Logger, not console
- [ ] Git staged and ready: `git status`
- [ ] Full test suite runs in < 3 seconds

---

## File Changes Summary

```
tests/e2e/command_errors_test.go
  + func TestCLI_NotesAdd_FileAlreadyExists(t *testing.T)
  + func TestCLI_NotesAdd_InvalidCharacters(t *testing.T)
  + func TestCLI_NotebookCreate_InvalidName(t *testing.T)
  + func TestCLI_NotesRemove_FileNotFound(t *testing.T)
  + func TestCLI_Commands_PermissionDenied(t *testing.T)

internal/services/note_test.go
  + func TestNoteService_SearchNotes_ComplexQueries(t *testing.T)
  + func TestNoteService_SearchNotes_LargeResultSets(t *testing.T)
  + func TestNoteService_SearchNotes_SpecialCharacters(t *testing.T)
  + func TestNoteService_SearchNotes_MissingMetadata(t *testing.T)

internal/core/strings_test.go
  + func TestObjectToFrontmatter_NestedObjects(t *testing.T)
  + func TestObjectToFrontmatter_MixedArrayTypes(t *testing.T)
  + func TestObjectToFrontmatter_UnicodeValues(t *testing.T)
  + func TestObjectToFrontmatter_NullValues(t *testing.T)
  + func TestObjectToFrontmatter_EmptyCollections(t *testing.T)
  + func TestObjectToFrontmatter_LongStrings(t *testing.T)
  + func TestObjectToFrontmatter_YAMLEscaping(t *testing.T)
```

---

## Implementation Details

### Task 1: Command Error Tests

Command errors should be tested at CLI level using the testEnv infrastructure:

```go
func TestCLI_NotesAdd_FileAlreadyExists(t *testing.T) {
    env := newTestEnv(t)
    notebook := env.createNotebook("test")
    
    // Create first note
    env.createNote(notebook, "test.md", "# Test")
    
    // Try to create with same name
    stdout, stderr, exitCode := env.run(
        "notes", "add", "test.md", 
        "--notebook", notebook,
    )
    
    if exitCode == 0 {
        t.Error("expected non-zero exit code for existing file")
    }
    if !strings.Contains(stderr, "already exists") {
        t.Errorf("expected error message in stderr, got: %s", stderr)
    }
}
```

### Task 2: SearchNotes Edge Cases

Complex query tests should use the NoteService directly:

```go
func TestNoteService_SearchNotes_ComplexQueries(t *testing.T) {
    // Create notebook with diverse notes
    nb := createTestNotebook(t)
    ns := services.NewNoteService(db, nb)
    
    tests := []struct {
        name          string
        query         string
        expectedCount int
    }{
        {
            "AND/OR mixed",
            "title:test AND body:content OR tags:important",
            3,
        },
        {
            "regex special chars",
            "title:foo\\(bar\\)",
            1,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            notes, err := ns.SearchNotes(tt.query)
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if len(notes) != tt.expectedCount {
                t.Errorf("got %d notes, want %d", len(notes), tt.expectedCount)
            }
        })
    }
}
```

### Task 3: Frontmatter Edge Cases

Object-to-frontmatter should test type conversion edge cases:

```go
func TestObjectToFrontmatter_NestedObjects(t *testing.T) {
    obj := map[string]any{
        "title": "Test",
        "metadata": map[string]any{
            "author": "John",
            "tags": []string{"a", "b"},
        },
    }
    
    result := core.ObjectToFrontmatter(obj)
    
    // Should include nested structure
    if !strings.Contains(result, "metadata:") {
        t.Errorf("nested object not in result: %s", result)
    }
}
```

---

## Dependencies

- All dependencies already available
- Uses existing test infrastructure
- No new packages needed

---

## Rollback Plan

If any test fails:
1. Review test output carefully
2. Determine if it's a real bug or test issue
3. Fix the issue or adjust test expectations
4. Verify no regressions
5. Only proceed if all tests pass

---

## Success Metrics After Completion

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Overall Coverage | 75% | 80%+ | 80% ✓ |
| Error Path Coverage | ~40% | ~70% | 70% ✓ |
| Functions at <70% | 6 | 2-3 | <3 ✓ |
| Passing Tests | 151 | 166-171 | 166+ ✓ |
| Test Duration | 2.2s | 2.3s | <3s ✓ |

---

## Documentation

After Phase 2 completion, document:
- [ ] Updated TESTING.md with edge case patterns
- [ ] Learning file about error handling test patterns
- [ ] Examples of complex test scenarios

---

## Completion Checklist

When all tests pass and quality gates are met:

- [ ] All 15-20 new tests passing
- [ ] Coverage improved to 80%+
- [ ] Error paths now have proper test coverage
- [ ] Edge cases documented in test code
- [ ] No new lint violations
- [ ] No test flakes detected
- [ ] Git diff shows only test additions
- [ ] Ready to move to Phase 3 (optional) or conclude

---

## Next Steps

After Phase 2 Completion:

**Option A (Production Ready):**
1. Commit: `git commit -m "test: add error handling and edge case tests"`
2. Consider work complete (80% coverage achieved)
3. Move to documentation and knowledge capture

**Option B (Enterprise Ready):**
1. Commit changes
2. Move to Phase 3: Future-Proofing (optional, 2-3 hours)
3. Target 85%+ coverage with concurrency and stress tests

**Recommendation:** Complete Phase 2, then decide based on time and business needs.

---

**Status:** 📋 PLANNED  
**Last Updated:** 2026-01-18 20:32 GMT+10:30  
**Start Date:** TBD (after Phase 1 completion)  
**Expected Duration:** 2 hours
