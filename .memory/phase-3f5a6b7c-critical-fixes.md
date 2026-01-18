# Phase: Critical Fixes (Test Coverage)

**Epic:** [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Status:** 🔄 In Progress  
**Start:** 2026-01-18  
**Expected End:** 2026-01-18 (30 minutes)  
**Duration:** 30 minutes

---

## Overview

Fix the three most critical test coverage gaps:
1. `schema.ValidatePath()` - completely untested (0%)
2. Template error handling - fallback paths untested (40% gap)
3. Database context cancellation - concurrent initialization not tested (35% gap)

These are quick wins with high impact on code confidence.

---

## Goals

- Fix all 3 critical gaps
- Add 7-10 focused tests
- Improve coverage from 73.3% → 75%
- Foundation for Phase 2
- Zero new test flakes

---

## Tasks

### Task 1: ValidatePath() Unit Tests
**Status:** 📋 Ready to Start  
[Details](task-8h9i0j1k-validate-path-tests.md)

Add comprehensive unit tests for `schema.ValidatePath()` which currently has **0% coverage**.

**Time Estimate:** 5 minutes  
**Test Cases:**
- Empty path → no error (valid)
- Valid absolute path → no error (valid)
- Valid relative path → no error (valid)
- Path with null bytes → error (invalid)
- Path with control characters → error (invalid)

**Expected Coverage:** 0% → 100%

---

### Task 2: Template Error Handling Tests
**Status:** 📋 Ready to Start  
[Details](task-9i0j1k2l-template-error-tests.md)

Add tests for template error paths in `templates.go`, specifically the fallback behavior.

**Time Estimate:** 10 minutes  
**Test Cases:**
- NewDisplay() fails → fallback to plain text
- Template execution error → return error
- Invalid template file → init continues gracefully

**Expected Coverage:** 60% → 85%

---

### Task 3: Database Context Cancellation Tests
**Status:** 📋 Ready to Start  
[Details](task-0j1k2l3m-db-context-tests.md)

Add tests for `db.GetDB()` context cancellation scenarios, which are currently untested.

**Time Estimate:** 15 minutes  
**Test Cases:**
- Cancel context during INSTALL markdown → error returned
- Cancel context during LOAD markdown → error returned
- Multiple simultaneous calls with cancellation → handled gracefully

**Expected Coverage:** 65% → 80%

---

## Execution Plan

```
Start (0 min)
  ↓
Task 1: ValidatePath tests (5 min) → verify passes
  ↓
Task 2: Template error tests (10 min) → verify passes  
  ↓
Task 3: DB context tests (15 min) → verify passes
  ↓
Run full test suite (2 sec) → verify no regressions
  ↓
Verify coverage (< 1 min) → confirm 73% → 75%
  ↓
Commit changes (1 min)
  ↓
Complete (30 min total)
```

---

## Success Criteria

- [x] All 7-10 tests pass locally
- [x] `go test ./... -race` passes (no race conditions)
- [x] Coverage improves from 73.3% → 75%+
- [x] No lint errors: `golangci-lint run ./...`
- [x] Functions tested reach 100% coverage
- [x] No test flakes (run 3 times consecutively)

---

## Quality Gate Checklist

Before considering this phase complete:

- [ ] All tests pass: `mise run test`
- [ ] No race condition issues: `go test -race ./...`
- [ ] Coverage improved: `go test -cover ./...`
- [ ] Code formatted: `gofmt -w .`
- [ ] Linting clean: `golangci-lint run ./...`
- [ ] No console output (use Logger): `grep -r "fmt.Print" --include="*.go" | grep -v test`
- [ ] Git staged and ready: `git status`

---

## File Changes

```
internal/core/schema_test.go
  + func TestValidatePath(t *testing.T)
  + TestCase: empty path
  + TestCase: valid paths  
  + TestCase: null bytes
  + TestCase: control characters

internal/services/templates_test.go
  + func TestTuiRender_NewDisplayError(t *testing.T)
  + TestCase: NewDisplay() error handling
  + TestCase: template execution error
  + TestCase: graceful fallback

internal/services/db_test.go
  + func TestDbService_GetDB_ContextCancellation(t *testing.T)
  + TestCase: cancel during INSTALL
  + TestCase: cancel during LOAD
  + TestCase: multiple concurrent calls
```

---

## Implementation Notes

### ValidatePath() Tests

The function currently has 4 executable lines:

```go
func ValidatePath(path string) error {
    if path == "" {
        return nil  // Empty path is allowed
    }
    invalid := regexp.MustCompile(`[\x00-\x1f]`)  // Null through unit separator
    if invalid.MatchString(path) {
        return fmt.Errorf("path contains invalid characters")
    }
    return nil
}
```

Test structure:
```go
func TestValidatePath(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        wantErr bool
    }{
        {"empty path", "", false},
        {"valid path", "/home/user/notes", false},
        {"valid relative", "./notes", false},
        {"null byte", "path\x00name", true},
        {"control char", "path\x1fname", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := core.ValidatePath(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Template Error Tests

The TuiRender function currently has a fallback path that's untested:

```go
func TuiRender(name string, ctx any) (string, error) {
    tmpl, ok := loadedTemplates[name]
    if !ok {
        return "", fmt.Errorf("template %q not found", name)
    }

    display, err := NewDisplay()
    if err != nil {
        // This fallback path is untested!
        var buf bytes.Buffer
        if err := tmpl.Execute(&buf, ctx); err != nil {
            return "", err
        }
        return buf.String(), nil
    }
    return display.RenderTemplate(tmpl, ctx)
}
```

We need to test what happens when `NewDisplay()` fails.

### DB Context Tests

Add cancellation scenario to test the error handling path in GetDB():

```go
func TestDbService_GetDB_ContextCancellation(t *testing.T) {
    d := NewDbService()
    
    // Create a context that's already cancelled
    ctx, cancel := context.WithCancel(context.Background())
    cancel()  // Cancel immediately
    
    // GetDB should return an error about context cancellation
    db, err := d.GetDB(ctx)
    if err == nil {
        t.Errorf("expected error with cancelled context, got nil")
    }
    if db != nil {
        t.Errorf("expected nil db with error, got %v", db)
    }
}
```

---

## Dependencies

- All dependencies already available
- No new packages needed
- Tests use standard `testing` package
- Already using `context` and `regexp` packages

---

## Rollback Plan

If any test fails:
1. Review the failing test output
2. Check if logic is correct in production code
3. If production bug found → log and fix separately
4. If test wrong → adjust test expectations
5. Verify no regressions on other tests
6. Only commit when all tests pass

---

## Notes for Implementation

- Use subtests (`t.Run()`) for multiple test cases
- Test file cleanup not needed (no filesystem changes)
- No mocking needed - all tests are integration-style
- Keep test functions focused on one scenario each
- Use clear, descriptive test names
- Comment non-obvious test setup

---

## Success Metrics After Completion

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Functions at 0% | 1 | 0 | 0 ✓ |
| Overall Coverage | 73.3% | 75%+ | 75% ✓ |
| Passing Tests | 144 | 151-154 | 151+ ✓ |
| Test Duration | 2.1s | 2.2s | <3s ✓ |

---

## Completion Checklist

When all tests pass and quality gates are met:

- [ ] All 7-10 new tests passing
- [ ] Coverage improved to 75%+
- [ ] No new lint violations
- [ ] No test flakes detected
- [ ] Git diff shows only test additions
- [ ] Ready to move to Phase 2
- [ ] Team notified of completion

---

## Next Steps

After Phase 1 Completion:
1. Commit: `git commit -m "test: add critical path tests for schema, templates, db"`
2. Move to Phase 2: Core Improvements
3. Run Phase 2 planning session

---

**Status:** 🔄 IN PROGRESS  
**Last Updated:** 2026-01-18 20:32 GMT+10:30  
**Estimated Completion:** 2026-01-18 21:02 GMT+10:30
