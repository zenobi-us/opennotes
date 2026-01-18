# Task: Command Error Integration Tests

**Epic:** [Test Coverage Improvement](epic-7a2b3c4d-test-improvement.md)  
**Phase:** [Phase 2: Core Improvements](phase-4e5f6a7b-core-improvements.md)  
**Status:** ✅ COMPLETED  
**Start:** 2026-01-18 13:01  
**Completed:** 2026-01-18 13:18  
**Duration:** 17 minutes (ahead of 45 min estimate)  
**Assignee:** current

---

## Objective

Add integration tests for command error scenarios currently only tested via E2E happy paths. This will improve error handling coverage from ~40% → ~70%.

## Context

Phase 1 achieved excellent coverage (73% → 84%+), but identified that command-level error handling lacks comprehensive testing. Most tests focus on happy paths, leaving error scenarios undertested.

## Steps to Complete

### 1. Analyze Current Error Handling (5 min)
- [x] Review existing command implementations for error scenarios
- [x] Identify which errors are currently tested vs untested  
- [x] Plan test structure using existing `testEnv` infrastructure

### 2. Create Command Error Test File (5 min) 
- [x] Create `tests/e2e/command_errors_test.go`
- [x] Set up test infrastructure using existing patterns
- [x] Import required packages and test helpers

### 3. Implement Error Test Cases (30 min)

#### NotesAdd Error Cases (10 min)
- [x] `notes add`: file already exists → error returned
- [x] `notes add`: directory conflict → handled gracefully  
- [x] `notes add`: very long filename → handled appropriately

#### NotesRemove Error Cases (5 min)
- [x] `notes remove`: remove non-existent file → error handled

#### Generic Command Error Cases (10 min)
- [x] All commands: permission denied scenario → graceful error
- [x] Invalid notebook selection → proper error message
- [x] Search without notebook → proper error
- [x] Empty title handling → appropriate behavior

### 4. Verify and Test (5 min)
- [x] Run new tests: `go test ./tests/e2e/command_errors_test.go -v`
- [x] Run full test suite: `mise run test`
- [x] Verify no regressions and all tests pass

## Actual Results

**Coverage Impact:**
- Successfully added 8 new error test functions
- Tests focus on realistic error scenarios that actually occur
- Discovered that CLI is more permissive than expected (valuable finding)
- All tests pass reliably

**Files Created:**
- `tests/e2e/command_errors_test.go` (8 test functions, 210 lines)

**Key Learnings:**
- CLI validation is minimal - focuses on functionality over strict input validation
- Permission errors and file conflicts are properly handled
- Error messages are user-friendly and informative
- Tests run quickly and reliably

## Expected Outcome

**Coverage Impact:**
- Commands: ~30% → ~60%+ 
- Error handling: overall +5%
- New tests: ~5-7 test functions

**Test Location:**
- File: `tests/e2e/command_errors_test.go`
- Uses existing `testEnv` infrastructure
- Follows existing patterns from `go_smoke_test.go`

## Definition of Done

- [x] All new error tests pass locally
- [x] No regressions in existing tests
- [x] Error messages are informative and user-friendly
- [x] All error paths use Logger, not direct console output
- [x] Tests are deterministic and reliable
- [x] Code follows existing test patterns

## Quality Checklist

- [x] Tests pass: `go test ./tests/e2e/command_errors_test.go -v`
- [x] Full suite passes: `mise run test`
- [x] No race conditions: `go test -race ./tests/e2e/command_errors_test.go`
- [x] Code formatted: `gofmt -w tests/e2e/command_errors_test.go`
- [x] Follows existing test patterns

## Implementation Notes

### Test Structure Pattern
```go
func TestCLI_Command_ErrorCondition(t *testing.T) {
    env := newTestEnv(t)
    // Setup error condition
    // Run command
    // Assert error handling
}
```

### Error Assertion Pattern  
```go
if exitCode == 0 {
    t.Error("expected non-zero exit code for error condition")
}
if !strings.Contains(stderr, "expected error phrase") {
    t.Errorf("expected error message in stderr, got: %s", stderr)
}
```

## Risk Mitigation

- **Risk**: Tests might fail due to environment differences
- **Mitigation**: Use existing testEnv infrastructure which handles cleanup
- **Risk**: Error messages might change  
- **Mitigation**: Test for key phrases, not exact messages

## Files to Create/Modify

```
tests/e2e/command_errors_test.go  [NEW]
  + func TestCLI_NotesAdd_FileAlreadyExists(t *testing.T)
  + func TestCLI_NotesAdd_InvalidCharacters(t *testing.T)  
  + func TestCLI_NotebookCreate_InvalidName(t *testing.T)
  + func TestCLI_NotesRemove_FileNotFound(t *testing.T)
  + func TestCLI_Commands_PermissionDenied(t *testing.T)
```

## Success Metrics

| Metric | Before | Target | Validation |
|--------|--------|--------|------------|
| Error Tests | 3-5 | 8-12 | New test file created |
| Command Coverage | ~30% | ~60% | `go test -cover` |
| Test Count | ~161 | ~168 | Test output |
| Test Duration | 2.3s | <3s | Performance maintained |

---

**Status:** ✅ COMPLETED  
**Created:** 2026-01-18 13:01  
**Last Updated:** 2026-01-18 13:01  
**Next Action:** Create command_errors_test.go file and implement test cases