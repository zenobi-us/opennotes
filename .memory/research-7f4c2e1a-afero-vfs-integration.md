---
id: 7f4c2e1a
title: Afero Virtual File System Integration Research
type: "research"
created_at: 2026-01-23T21:15:00+10:30
updated_at: 2026-01-23T21:15:00+10:30
status: completed
epic_id: 3e01c563
tags:
  - testing
  - architecture
  - vfs
  - storage-abstraction
---

# Afero Virtual File System Integration Research

## Executive Summary

Research into using `github.com/spf13/afero` for test file system isolation in OpenNotes. This addresses the critical issue where tests modify `~/.config/opennotes/config.json`.

**Recommendation**: Adopt afero with a storage abstraction layer supporting pluggable backends (OsFs for production, MemMapFs for tests).

---

## Problem Statement

**Current Issue**:
- Tests directly read/write `~/.config/opennotes/config.json`
- Test pollution: User's actual config file gets modified
- Non-deterministic test behavior
- CI/CD risk: Tests can fail based on user environment

**Impact**:
- User experience: Tests corrupt personal config during development
- Reliability: Tests not reproducible across machines
- Isolation: No way to run tests in parallel safely

---

## Why Afero?

### Advantages
1. ✅ **Actively Maintained** - spf13 (Cobra/Viper creator) maintains it
2. ✅ **Production-Tested** - Used by many real projects
3. ✅ **Minimal API** - Uses familiar `os.File` interface patterns
4. ✅ **Pluggable** - Multiple backends available
5. ✅ **Lightweight** - No heavy dependencies
6. ✅ **Clean Integration** - `afero.Fs` interface-based design

### Alternative Considered: blang/vfs
- ❌ Less maintained
- ❌ Smaller ecosystem
- ❌ Not used by Cobra/Viper ecosystem
- ✅ Smaller library (but unnecessary feature parity)

---

## Afero Architecture

### Core Interfaces

```go
// Main filesystem abstraction
type Fs interface {
    Create(name string) (File, error)
    Mkdir(name string, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    Open(name string) (File, error)
    OpenFile(name string, flag int, perm os.FileMode) (File, error)
    Remove(name string) error
    RemoveAll(path string) error
    Rename(oldname, newname string) error
    Stat(name string) (os.FileInfo, error)
    // ... other methods
}

// File abstraction
type File interface {
    Name() string
    Read(b []byte) (n int, err error)
    ReadAt(b []byte, off int64) (n int, err error)
    Write(b []byte) (n int, err error)
    WriteAt(b []byte, off int64) (n int, err error)
    Close() error
    // ... other methods
}
```

### Available Implementations

| Backend | Type | Use Case | Memory | Concurrency |
|---------|------|----------|--------|-------------|
| **OsFs** | Real FS | Production | Disk | Safe |
| **MemMapFs** | In-Memory | Testing | RAM | Safe |
| **UnionFs** | Layered | Overlays | Composite | Safe |
| **ReadOnlyFs** | Wrapper | Read-only views | N/A | Safe |
| **BasePathFs** | Wrapper | Chroot-like | N/A | Safe |

---

## Proposed Storage Abstraction Layer

### Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                    OpenNotes                        │
├─────────────────────────────────────────────────────┤
│  Commands (cmd/)                                    │
│  - Pass afero.NewOsFs() to services                 │
├─────────────────────────────────────────────────────┤
│  Services (internal/services/)                      │
│  - ConfigService(fs afero.Fs)                       │
│  - NotebookService(fs afero.Fs)                     │
│  - All use fs.Open, afero.ReadFile, etc.           │
├─────────────────────────────────────────────────────┤
│  Storage Layer (internal/storage/) [NEW]            │
│  - StorageProvider interface (optional layer)       │
│  - OsProvider: wraps afero.NewOsFs()               │
│  - MemProvider: wraps afero.NewMemMapFs()          │
│  - Easy to add: CloudProvider, etc.                │
├─────────────────────────────────────────────────────┤
│  Filesystem Backends (afero)                        │
│  - afero.OsFs (real filesystem)                     │
│  - afero.MemMapFs (in-memory)                       │
│  - afero.UnionFs (layered/overlays)                │
└─────────────────────────────────────────────────────┘

Tests:
- Create fresh afero.MemMapFs() for each test
- Setup test files in memory
- No side effects on real filesystem
- Fast, deterministic, reproducible
```

### Layer Design

**Option 1: Direct afero.Fs** (Simpler, recommended for now)
```go
type ConfigService struct {
    fs afero.Fs
    // ...
}

// Usage
cfg := services.NewConfigService(afero.NewOsFs())  // Production
cfg := services.NewConfigService(afero.NewMemMapFs())  // Tests
```

**Option 2: StorageProvider abstraction** (Future-proof)
```go
type StorageProvider interface {
    GetFs() afero.Fs
    GetConfig() map[string]string
}

type OsProvider struct {}
func (p *OsProvider) GetFs() afero.Fs { return afero.NewOsFs() }

type MemProvider struct { fs afero.Fs }
func (p *MemProvider) GetFs() afero.Fs { return p.fs }
```

**Recommendation**: Start with Option 1 (direct afero.Fs), evolve to Option 2 if needed for cloud storage, S3, etc.

---

## Integration Points

### Services That Need Changes

1. **ConfigService** (`internal/services/config.go`)
   - Currently uses: `os.Open`, `ioutil.ReadFile`, etc.
   - Change to: `fs.Open`, `afero.ReadFile`, etc.
   - Constructor: Add `fs afero.Fs` parameter

2. **NotebookService** (`internal/services/notebook.go`)
   - File discovery (glob patterns)
   - Config loading from `.opennotes.json`
   - Constructor: Add `fs afero.Fs` parameter

3. **DbService** (`internal/services/db.go`)
   - Database file creation/paths
   - Maybe not needed immediately (DuckDB handles its own FS)

4. **Commands** (`cmd/`)
   - Construct services with `afero.NewOsFs()`
   - No business logic changes

### Test Changes

1. **Test Helpers** (`internal/testing/helpers.go` - NEW)
   - `SetupTestFS()` - Create MemMapFs with sample files
   - `CreateTestConfig(fs afero.Fs, data map[string]interface{})` - Write test config
   - `CreateTestNotebook(fs afero.Fs, name string)` - Setup notebook
   - `CreateTestFiles(fs afero.Fs, paths []string)` - Create test markdown files

2. **Test Files** (`*_test.go`)
   - Use `testFs := afero.NewMemMapFs()`
   - Setup test data with helpers
   - No cleanup needed (tests are isolated)

---

## Migration Strategy

### Phase 1: Infrastructure
- [ ] Add afero to go.mod: `go get github.com/spf13/afero`
- [ ] Create test helpers (`internal/testing/helpers.go`)
- [ ] Create storage abstraction layer docs

### Phase 2: Core Services
- [ ] ConfigService: Accept afero.Fs parameter
- [ ] NotebookService: Accept afero.Fs parameter
- [ ] Commands: Pass afero.NewOsFs() to services

### Phase 3: Test Migration (Incremental)
- [ ] Config-related tests: Use MemMapFs
- [ ] Notebook-related tests: Use MemMapFs
- [ ] Command integration tests: Use MemMapFs
- [ ] View system tests: Use MemMapFs

### Phase 4: Cleanup
- [ ] Remove any hardcoded paths
- [ ] Verify no ~/.config access during tests
- [ ] Performance benchmarks

---

## Code Examples

### Before (Current)

```go
// ConfigService
type ConfigService struct {
    configPath string
    // ...
}

func NewConfigService() *ConfigService {
    configPath := filepath.Join(os.Getenv("HOME"), ".config", "opennotes", "config.json")
    data, err := ioutil.ReadFile(configPath)
    // ...
}
```

### After (With Afero)

```go
// ConfigService
type ConfigService struct {
    fs afero.Fs
    configPath string
    // ...
}

func NewConfigService(fs afero.Fs) *ConfigService {
    configPath := filepath.Join("/.config/opennotes/config.json")  // Relative to fs root
    data, err := afero.ReadFile(fs, configPath)
    // ...
}

// Production
cfg := services.NewConfigService(afero.NewOsFs())

// Tests
testFs := afero.NewMemMapFs()
afero.WriteFile(testFs, "/.config/opennotes/config.json", []byte(`{...}`), 0644)
cfg := services.NewConfigService(testFs)
```

---

## Benefits Checklist

- ✅ **Test Isolation**: Each test gets fresh MemMapFs
- ✅ **No Pollution**: ~/.config/opennotes never touched during tests
- ✅ **Reproducible**: Same results every test run
- ✅ **Fast**: In-memory operations (no disk I/O)
- ✅ **Parallel-Safe**: Each test has isolated filesystem
- ✅ **Pluggable**: Easy to add cloud storage later
- ✅ **Production-Ready**: Uses proven afero library
- ✅ **Minimal Breaking Changes**: Mostly constructor changes

---

## Recommended Next Steps

1. **Create this epic**: "Storage Abstraction Layer with Afero"
2. **Break into phases**:
   - Phase 1: Add afero dependency and test helpers
   - Phase 2: Migrate ConfigService and NotebookService
   - Phase 3: Convert tests incrementally
   - Phase 4: Verify and cleanup
3. **Estimate effort**: ~3-4 hours total
4. **Risk level**: Low (isolated to services and tests)
5. **Compatibility**: 100% backward compatible

---

## References

- **Afero GitHub**: https://github.com/spf13/afero
- **Afero Docs**: https://pkg.go.dev/github.com/spf13/afero
- **Cobra (uses afero)**: https://github.com/spf13/cobra
- **Example**: https://github.com/spf13/afero/tree/master/examples

---

## Conclusion

Afero is the right choice for OpenNotes. It provides a clean, production-tested way to abstract filesystem operations while maintaining minimal code changes. The storage abstraction layer approach allows for future extensibility (cloud storage, S3, etc.) without major refactoring.

**Action**: Proceed with implementation planning for storage abstraction epic.
