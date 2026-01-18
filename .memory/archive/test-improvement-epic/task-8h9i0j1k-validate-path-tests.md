# Task: Add ValidatePath() Unit Tests

**Phase:** [Critical Fixes](phase-3f5a6b7c-critical-fixes.md)  
**Epic:** [Test Improvement](epic-7a2b3c4d-test-improvement.md)  
**Status:** ✅ Completed  
**Time Estimate:** 5 minutes  
**Priority:** CRITICAL

---

## Objective

Add comprehensive unit tests for `schema.ValidatePath()` to achieve **0% → 100% coverage**.

---

## Current State

**Function Location:** `internal/core/schema.go:128-138`

```go
// ValidatePath validates a filesystem path.
func ValidatePath(path string) error {
    if path == "" {
        return nil // Empty path is allowed (uses default)
    }

    // Check for invalid characters
    invalid := regexp.MustCompile(`[\x00-\x1f]`)
    if invalid.MatchString(path) {
        return fmt.Errorf("path contains invalid characters")
    }

    return nil
}
```

**Current Coverage:** 0% (completely untested)

**Why Untested:** Likely an oversight during initial implementation.

---

## Test Cases to Implement

| Test Case | Input | Expected | Rationale |
|-----------|-------|----------|-----------|
| Empty path | `""` | `nil` (no error) | Empty paths allowed per comment |
| Valid absolute path | `"/home/user/notes"` | `nil` | Standard absolute path |
| Valid relative path | `"./notes"` | `nil` | Relative path support |
| Path with null byte | `"path\x00name"` | Error | Security: null termination |
| Path with control char | `"path\x1fname"` | Error | Control characters invalid |

---

## Implementation

### File to Modify

`internal/core/schema_test.go`

### Code to Add

```go
func TestValidatePath(t *testing.T) {
    tests := []struct {
        name    string
        path    string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "empty path allowed",
            path:    "",
            wantErr: false,
        },
        {
            name:    "valid absolute path",
            path:    "/home/user/notes",
            wantErr: false,
        },
        {
            name:    "valid relative path",
            path:    "./notes",
            wantErr: false,
        },
        {
            name:    "path with null byte",
            path:    "path\x00name",
            wantErr: true,
            errMsg:  "invalid characters",
        },
        {
            name:    "path with control character",
            path:    "path\x1fname",
            wantErr: true,
            errMsg:  "invalid characters",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePath(tt.path)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
            }
            if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
                t.Errorf("ValidatePath(%q) error = %q, want to contain %q", tt.path, err.Error(), tt.errMsg)
            }
        })
    }
}
```

---

## Verification Steps

1. **Add the test function** to `internal/core/schema_test.go`

2. **Run the test:**
   ```bash
   cd /mnt/Store/Projects/Mine/Github/opennotes
   go test -v ./internal/core -run TestValidatePath
   ```
   
   Expected output:
   ```
   === RUN   TestValidatePath
   === RUN   TestValidatePath/empty_path_allowed
   === RUN   TestValidatePath/valid_absolute_path
   === RUN   TestValidatePath/valid_relative_path
   === RUN   TestValidatePath/path_with_null_byte
   === RUN   TestValidatePath/path_with_control_character
   --- PASS: TestValidatePath (0.00s)
   ```

3. **Check coverage increased:**
   ```bash
   go test -cover ./internal/core
   ```
   
   Should show `schema.go:128` → 100% now

4. **Run full test suite to ensure no regressions:**
   ```bash
   mise run test
   ```

---

## Success Criteria

- [x] Test function compiles without errors
- [x] All 5 test cases pass
- [x] Coverage goes from 0% → 100%
- [x] No new lint violations: `golangci-lint run ./internal/core`
- [x] No race conditions: `go test -race ./internal/core`
- [x] Existing tests still pass (no regressions)

---

## Acceptance Criteria

✓ Task is complete when:
1. All test cases pass locally
2. Coverage for ValidatePath() reaches 100%
3. No test flakes on 3 consecutive runs
4. Code review approved

---

## Notes

- The function is simple (4 executable lines) so tests should be straightforward
- Test uses table-driven approach for clarity and maintainability
- Control characters tested include: `\x00` (null), `\x1f` (unit separator)
- Empty path explicitly allowed per function documentation

---

## Time Breakdown

- Reading function: 1 min
- Writing test code: 2 min
- Running verification: 1 min
- Fixing any issues: 1 min
- **Total: 5 minutes**

---

**Status:** 📋 READY  
**Assigned To:** TBD  
**Due:** By end of Phase 1  
**Blocked By:** None  
**Blocks:** Phase 1 completion
