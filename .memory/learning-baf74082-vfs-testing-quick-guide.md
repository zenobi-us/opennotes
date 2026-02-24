---
type: "learning"
---

# Virtual File System Testing - Quick Reference Guide

## TL;DR: What Should I Use Right Now?

**For OpenNotes**: Keep doing what you're doing!
- ✅ Use `t.TempDir()` for file system isolation
- ✅ Use `t.Setenv()` for environment variables
- ✅ Use `testify/assert` for assertions
- ❌ Don't introduce afero unless tests become slow

---

## Quick Decision Tree

```
Need to test file system behavior?
├─ Yes, quickly & simply?
│  └─ → Use t.TempDir() + testify (CURRENT APPROACH) ✅
│
├─ Yes, with OS error simulation?
│  └─ → Use afero (NOT NOW, maybe in 6+ months) 🟡
│
├─ Testing environment variables?
│  └─ → Use t.Setenv() (ALREADY DOING) ✅
│
└─ Performance critical (1000+ tests/sec)?
   └─ → Use afero + memory FS (NOT YET) 🟡
```

---

## Current OpenNotes Test Pattern (CORRECT)

### Pattern 1: Config File Testing
```go
func TestConfigService_LoadFromFile(t *testing.T) {
    // 1. Create isolated temp directory (auto-cleanup)
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.json")
    
    // 2. Create test file
    createTestConfigFile(t, configPath, testConfig)
    
    // 3. Load and test
    svc, err := NewConfigServiceWithPath(configPath)
    require.NoError(t, err)
    assert.Equal(t, testConfig, svc.Store)
    
    // 4. Auto-cleanup (no manual cleanup needed)
}
```

**Why this works**:
- Temp dir is unique per test
- Files exist in real filesystem
- Auto-cleanup after test
- Clear, readable code

### Pattern 2: Environment Variable Testing
```go
func TestConfigService_EnvVarOverride(t *testing.T) {
    // 1. Override env var (auto-reset after test)
    t.Setenv("OPENNOTES_NOTEBOOKPATH", "/test/path")
    
    // 2. Test
    svc, err := NewConfigServiceWithPath(configPath)
    require.NoError(t, err)
    assert.Equal(t, "/test/path", svc.Store.NotebookPath)
    
    // 3. Auto-reset (no cleanup needed)
}
```

**Why this works**:
- Safe environment variable override
- Auto-reset, no side effects
- Thread-safe per test
- Clear what's being tested

### Pattern 3: Table-Driven Tests
```go
func TestNotebookService_Discovery(t *testing.T) {
    tests := []struct {
        name     string
        setup    func(tmpDir string)
        expected int
    }{
        {
            name: "finds notebooks",
            setup: func(tmpDir string) {
                createTestNotebook(t, tmpDir, "nb1")
                createTestNotebook(t, tmpDir, "nb2")
            },
            expected: 2,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()
            tt.setup(tmpDir)
            
            // Test
        })
    }
}
```

**Why this works**:
- Each test case gets fresh tmpDir
- Clear, parameterized test logic
- Easy to add new test cases

---

## Common Pitfalls & How to Avoid

### ❌ Pitfall 1: Forgetting to use t.TempDir()
```go
// BAD: Leaves files on disk
func TestSomething(t *testing.T) {
    tmpDir := "/tmp/my-test"  // ❌ Leftover after test
    os.MkdirAll(tmpDir, 0755)
    // ... test code ...
    // Manual cleanup often forgotten
}

// GOOD: Auto-cleanup
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()  // ✅ Cleaned up automatically
    // ... test code ...
}
```

### ❌ Pitfall 2: Assuming file system state between tests
```go
// BAD: Depends on global state
var testDir = "/tmp/shared-test-dir"

func TestA(t *testing.T) {
    createFile(testDir, "file1.txt")
}

func TestB(t *testing.T) {
    // ❌ Assumes file1.txt exists from TestA
    // But tests run in unpredictable order!
    content, _ := os.ReadFile(filepath.Join(testDir, "file1.txt"))
}

// GOOD: Each test is independent
func TestA(t *testing.T) {
    tmpDir := t.TempDir()
    createFile(tmpDir, "file1.txt")
}

func TestB(t *testing.T) {
    tmpDir := t.TempDir()
    createFile(tmpDir, "file1.txt")  // ✅ Independent setup
}
```

### ❌ Pitfall 3: Manually managing cleanup with t.Cleanup()
```go
// AVOID: Manual cleanup is error-prone
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    f, _ := os.Create(filepath.Join(tmpDir, "file.txt"))
    
    t.Cleanup(func() {
        f.Close()  // ❌ Easy to miss, forget, or do wrong
    })
}

// GOOD: t.TempDir() handles it
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    f, _ := os.Create(filepath.Join(tmpDir, "file.txt"))
    defer f.Close()  // ✅ Standard Go pattern
    // tmpDir auto-cleaned
}
```

---

## Copy-Paste Templates

### Template 1: Simple File Test
```go
func TestMyService_LoadConfig(t *testing.T) {
    tmpDir := t.TempDir()
    configPath := filepath.Join(tmpDir, "config.json")
    
    // Create test file
    testConfig := MyConfig{Field: "value"}
    data, _ := json.MarshalIndent(testConfig, "", "  ")
    os.MkdirAll(filepath.Dir(configPath), 0755)
    os.WriteFile(configPath, data, 0644)
    
    // Test
    svc, err := NewService(configPath)
    require.NoError(t, err)
    assert.Equal(t, testConfig.Field, svc.Config.Field)
}
```

### Template 2: Environment Variable Test
```go
func TestMyService_ConfigFromEnv(t *testing.T) {
    t.Setenv("MY_ENV_VAR", "test_value")
    
    svc, err := NewService()
    require.NoError(t, err)
    assert.Equal(t, "test_value", svc.Config.Value)
}
```

### Template 3: Table-Driven File Tests
```go
func TestMyService_Scenarios(t *testing.T) {
    tests := []struct {
        name    string
        create  func(tmpDir string) string  // Create setup, return data
        expect  string
    }{
        {
            name: "basic scenario",
            create: func(tmpDir string) string {
                // Setup files
                return "expected_result"
            },
            expect: "expected_result",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()
            result := tt.create(tmpDir)
            assert.Equal(t, tt.expect, result)
        })
    }
}
```

### Template 4: Helper Function (Like OpenNotes Does)
```go
// Helper function in *_test.go
func createTestFile(t *testing.T, dir, name string, content string) string {
    t.Helper()  // Important: tells Go this is a helper
    
    path := filepath.Join(dir, name)
    os.MkdirAll(filepath.Dir(path), 0755)
    os.WriteFile(path, []byte(content), 0644)
    
    return path
}

// Usage
func TestSomething(t *testing.T) {
    tmpDir := t.TempDir()
    path := createTestFile(t, tmpDir, "config.json", "{}")
    // ... test
}
```

---

## Performance Expectations

| Test Type | Count | Time | Speed |
|-----------|-------|------|-------|
| Simple unit (no FS) | 100 | 0.5s | ⚡⚡⚡ Fast |
| With t.TempDir() | 100 | 2-3s | ⚡⚡ Acceptable |
| With actual files | 100 | 2-3s | ⚡⚡ Acceptable |
| Current OpenNotes | 161 | ~4s | ✅ Good |
| Large suite needed? | 500+ | 10-20s | 🐢 Slow |

**When to optimize**: If suite > 500 tests and runs > 10 seconds

---

## When to Reconsider This Approach

### ✅ Signal 1: Test Suite Becomes Slow
```bash
$ mise run test
# Takes 20+ seconds for 500+ tests
# → Consider afero for ~10x speedup
```

### ✅ Signal 2: Need to Simulate OS Errors
```go
// Current approach can't easily test:
// - Permission denied
// - Disk full
// - File in use (on Windows)
// → May want afero for error simulation
```

### ✅ Signal 3: Many Tests Setting Up Same Files
```go
// If 50+ tests do similar setup:
tests := []struct {
    setup func()  // Repeated across many tests
}
// → Table-driven with afero might be cleaner
```

---

## "Should I Add afero Right Now?"

**Answer: No, not yet.**

### Why:
1. **Current approach works fine** - 161 tests, 4 seconds ✅
2. **No OS error testing needed** - Filesystem works correctly ✅
3. **Low refactoring burden now** - Would be high refactoring cost ✅
4. **No performance issue** - 4 seconds is acceptable ✅

### When to revisit:
- ⏱️ **6 months from now** if test count > 500
- ⏱️ **When test suite > 10 seconds**
- ⏱️ **When need to test permission errors, disk full, etc.**

---

## FAQ

**Q: Why not use afero now?**  
A: Too much refactoring for no immediate gain. Current approach is simpler and works well.

**Q: Will t.TempDir() ever cause issues?**  
A: No, it's standard Go and well-tested. Cleanup is guaranteed by test framework.

**Q: Can I mix t.TempDir() and afero in same project?**  
A: Yes! You could migrate services one at a time. But not necessary for OpenNotes.

**Q: What if a test needs specific file permissions?**  
A: With afero, hard to test. With real FS, you can use `os.Chmod()`. Use real FS.

**Q: Is there a middle ground between current and afero?**  
A: Not really. Either use real FS (simple) or virtual FS (complex). Current is best middle ground.

**Q: How do I test the config loading code reliably?**  
A: Current approach: Create real config file in t.TempDir(), load it, verify. Perfect. ✅

---

## Next Steps

1. **Keep current testing approach** ✅
2. **Monitor test performance** - Record in CI/CD
3. **Document this in AGENTS.md** - Add testing section
4. **Schedule 6-month review** - Revisit if needed
5. **If tests slow down**: Start migration plan

---

## References

- [Go testing best practices](https://pkg.go.dev/testing)
- [t.TempDir() documentation](https://pkg.go.dev/testing#T.TempDir)
- [t.Setenv() documentation](https://pkg.go.dev/testing#T.Setenv)
- [testify documentation](https://github.com/stretchr/testify)
- [spf13/afero](https://github.com/spf13/afero) (for future reference)

---

**Last Updated**: 2025-01-23  
**Status**: OpenNotes testing approach is optimal ✅
