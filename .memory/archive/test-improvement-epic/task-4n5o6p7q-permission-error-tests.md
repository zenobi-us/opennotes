# Task: Permission/Filesystem Error Tests

**Epic**: [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Phase**: [Phase 3: Future-Proofing](phase-5g6h7i8j-future-proofing.md)  
**Status**: ✅ COMPLETED  
**Actual Time**: 30 minutes  
**Results**: 8 comprehensive filesystem error tests implemented
- All tests passing with graceful error handling
- Cross-platform compatibility verified  
- Clear error messages validated
- Permission denied scenarios properly handled  
**Priority**: Phase 3 Task 1

---

## Objective

Test graceful error handling for filesystem permission issues and edge cases that can occur in real-world environments.

## Success Criteria

- [x] 5-8 new test cases covering permission scenarios
- [x] All tests demonstrate graceful error handling
- [x] Clear, actionable error messages returned
- [x] No crashes or panics on permission denied
- [x] Appropriate logging for troubleshooting

## Test Cases to Implement

### 1. Read-only Notebook Directory
**Test**: `TestNotebookService_ReadOnlyDirectory`
- Create notebook directory with read-only permissions
- Attempt to create new note → Should return permission error
- Verify error message is user-friendly

### 2. Permission Denied on Config File
**Test**: `TestConfigService_PermissionDeniedWrite`
- Create config directory with no write permissions
- Attempt to save config → Should return permission error
- Verify fallback behavior (read-only mode)

### 3. Disk Full Simulation (where possible)
**Test**: `TestNotebookService_DiskFull`
- Mock or simulate disk full scenario
- Attempt to write note → Should handle gracefully
- Clean error message about disk space

### 4. Stale File Handle (NFS/Network drives)
**Test**: `TestNoteService_StaleFileHandle`
- Mock file that exists but becomes inaccessible
- Attempt to read → Should detect and report
- Recoverable error handling

### 5. Symlink Handling
**Test**: `TestNotebookService_SymlinkHandling`
- Create notebook with symlinked directories
- Verify notes can be read/written through symlinks
- Handle broken symlinks gracefully

### 6. Very Long Paths (OS limits)
**Test**: `TestNotebookService_LongPaths`
- Create deeply nested directory structure
- Test if OS path length limits are handled
- Verify error messages are helpful

### 7. Invalid Characters in Filenames
**Test**: `TestNoteService_InvalidCharacters`
- Attempt to create notes with OS-invalid characters
- Verify sanitization or clear rejection
- Cross-platform compatibility

### 8. Concurrent File Access
**Test**: `TestNoteService_ConcurrentFileAccess`
- Multiple goroutines accessing same file
- File locking behavior
- Clean error handling for lock conflicts

---

## Implementation Strategy

### Test File Location
`tests/e2e/filesystem_errors_test.go`

### Mocking Strategy
- Use `internal/testutil` helper for temp directories
- Mock filesystem calls where needed
- Test actual OS permission behavior in temp directories

### Error Message Validation
Each test should verify:
- Error is not nil
- Error message contains helpful context
- Error type is appropriate (permission, not found, etc.)

---

## Expected Implementation

```go
func TestNotebookService_ReadOnlyDirectory(t *testing.T) {
    // Create temp notebook with read-only permissions
    tempDir := testutil.CreateTempDir(t)
    defer os.RemoveAll(tempDir)
    
    // Make directory read-only
    err := os.Chmod(tempDir, 0444)
    require.NoError(t, err)
    
    // Try to create note
    notebookService := services.NewNotebookService(...)
    err = notebookService.CreateNote(tempDir, "test.md", "content")
    
    // Verify appropriate error
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "permission denied")
}
```

---

## Quality Gates

- [ ] All new tests pass consistently
- [ ] Error messages are user-friendly (not just system errors)
- [ ] Cross-platform compatibility verified
- [ ] No test flakes or timing dependencies
- [ ] Proper cleanup in all test cases
- [ ] Tests run in < 5 seconds total

---

## Technical Notes

### Cross-Platform Considerations
- Windows vs Unix permission models differ
- Use `filepath` package for path handling
- Test symlink behavior may differ by OS

### Cleanup Strategy
- Always use `defer` for temp directory cleanup
- Reset permissions before cleanup if modified
- Handle cleanup errors gracefully

### Performance Impact
- These tests should be fast (< 1 second each)
- Use temp directories in memory if possible
- Avoid actual network filesystem testing

---

## Implementation Checklist

- [ ] Create `tests/e2e/filesystem_errors_test.go`
- [ ] Implement 8 test functions listed above
- [ ] Add helper functions for permission setup
- [ ] Verify error message quality
- [ ] Test cross-platform compatibility
- [ ] Add documentation for edge cases discovered
- [ ] Run full test suite to ensure no regressions

---

## Completion Criteria

1. All 8 test functions implemented and passing
2. Error messages are clear and actionable
3. No crashes or panics in error scenarios  
4. Tests complete in under 30 minutes development time
5. Zero regressions in existing test suite
6. Cross-platform compatibility verified

---

**Created**: 2026-01-18  
**Next Task**: [Concurrency Tests](task-5o6p7q8r-concurrency-tests.md)
