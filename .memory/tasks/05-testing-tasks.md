# Phase 5: Testing Tasks

## Overview
Prioritized and grouped tasks to address testing gaps in the Go implementation.

**Research:** See `.memory/research/testing-gaps.md` for detailed gap analysis.

---

## Task Checklist

### Setup Tasks (Before Starting)
- [ ] 5.0a Create `internal/testutil/notebook.go` helpers
- [ ] 5.0b Create `internal/testutil/config.go` helpers

### Group A: Foundation Services (P0 - Critical)
- [ ] 5.1 ConfigService Unit Tests (`config_test.go`) - 2-3 hrs
- [ ] 5.2 DbService Integration Tests (`db_test.go`) - 2-3 hrs
- [ ] 5.3 NotebookService Unit Tests (`notebook_test.go`) - 4-5 hrs

### Group B: Data Services (P1 - Important)
- [ ] 5.4 NoteService Integration Tests (`note_test.go`) - 2-3 hrs

### Group C: Display Layer (P2 - Nice to Have)
- [ ] 5.5 Display Service Tests (`display_test.go`) - 1-2 hrs
- [ ] 5.6 Templates Tests (`templates_test.go`) - 1 hr
- [ ] 5.7 Logger Tests (`logger_test.go`) - 30 min

### Group D: E2E/Integration (P2 - Future)
- [ ] 5.8 CLI E2E Tests (`tests/e2e/go_smoke_test.go`) - 4-6 hrs

---

## Group A: Foundation Services (P0 - Critical)

These must be completed first as other services depend on them.

### Task 5.1: ConfigService Unit Tests
**Priority:** P0 - Critical  
**Estimated Effort:** 2-3 hours  
**Dependencies:** None  
**File:** `internal/services/config_test.go`

**Justification:**  
ConfigService is the first service initialized and used by all other services. Bugs here cascade through the entire application. The 3-tier config loading (defaults → file → env) is complex logic that needs verification.

**Test Cases to Implement:**
```go
// Test default config values when no file/env exists
TestNewConfigService_Defaults

// Test config file loading (valid JSON)
TestNewConfigService_LoadFromFile

// Test config file error handling (invalid JSON)
TestNewConfigService_InvalidFile

// Test env var override: OPENNOTES_NOTEBOOKPATH
TestNewConfigService_EnvVarOverride

// Test env var priority over file values
TestNewConfigService_EnvVarPriorityOverFile

// Test Write() creates directory if needed
TestConfigService_Write_CreatesDirectory

// Test Write() persists config correctly
TestConfigService_Write_PersistsConfig

// Test GlobalConfigFile() returns expected path
TestGlobalConfigFile
```

**Implementation Notes:**
- Use `t.TempDir()` for isolated filesystem tests
- Set/unset env vars with `t.Setenv()`
- Consider `testify/suite` for setup/teardown

---

### Task 5.2: DbService Integration Tests
**Priority:** P0 - Critical  
**Estimated Effort:** 2-3 hours  
**Dependencies:** None (uses real DuckDB)  
**File:** `internal/services/db_test.go`

**Justification:**  
DbService wraps DuckDB with the markdown extension. Testing ensures the extension loads correctly and queries work. This is an integration test because mocking DuckDB defeats the purpose.

**Test Cases to Implement:**
```go
// Test GetDB() returns valid connection
TestDbService_GetDB_ReturnsConnection

// Test GetDB() loads markdown extension
TestDbService_GetDB_LoadsMarkdownExtension

// Test GetDB() lazy initialization (sync.Once)
TestDbService_GetDB_LazyInit

// Test Query() with simple SQL
TestDbService_Query_SimpleSQL

// Test Query() returns maps correctly
TestDbService_Query_ResultMapping

// Test Close() properly cleans up
TestDbService_Close

// Test concurrent GetDB() calls are safe
TestDbService_ConcurrentAccess
```

**Implementation Notes:**
- Test with real DuckDB (in-memory mode is fast)
- Use `read_markdown()` with temp markdown files to verify extension
- Use `sync.WaitGroup` for concurrency tests

---

### Task 5.3: NotebookService Unit Tests
**Priority:** P0 - Critical  
**Estimated Effort:** 4-5 hours  
**Dependencies:** ConfigService, DbService (mock or real)  
**File:** `internal/services/notebook_test.go`

**Justification:**  
NotebookService contains the most complex business logic: notebook discovery (3-step priority), context matching, CRUD operations. This is the core of the application and affects every user-facing command.

**Test Cases to Implement:**
```go
// HasNotebook tests
TestNotebookService_HasNotebook_ExistsTrue
TestNotebookService_HasNotebook_NotExistsFalse
TestNotebookService_HasNotebook_EmptyPath

// LoadConfig tests
TestNotebookService_LoadConfig_ValidConfig
TestNotebookService_LoadConfig_InvalidJSON
TestNotebookService_LoadConfig_MissingFile
TestNotebookService_LoadConfig_CreatesRootIfMissing

// Open tests
TestNotebookService_Open_Success
TestNotebookService_Open_LoadsNoteService

// Create tests
TestNotebookService_Create_CreatesDirectories
TestNotebookService_Create_WritesConfig
TestNotebookService_Create_RegistersGlobally

// Infer tests (3-step priority)
TestNotebookService_Infer_DeclaredPathPriority
TestNotebookService_Infer_ContextMatchPriority
TestNotebookService_Infer_AncestorSearchPriority
TestNotebookService_Infer_NoneFound

// List tests
TestNotebookService_List_FromRegistered
TestNotebookService_List_FromAncestors
TestNotebookService_List_Deduplicated
TestNotebookService_List_Empty

// Notebook method tests
TestNotebook_MatchContext_Match
TestNotebook_MatchContext_NoMatch
TestNotebook_AddContext_NewContext
TestNotebook_AddContext_DuplicateIgnored
TestNotebook_SaveConfig_LocalOnly
TestNotebook_SaveConfig_WithRegistration
```

**Implementation Notes:**
- Use `t.TempDir()` for filesystem isolation
- Create helper function `createTestNotebook(t, path, name)` for setup
- Mock ConfigService for isolation, or use real ConfigService with temp config file

---

## Group B: Data Services (P1 - Important)

These depend on Group A services being stable.

### Task 5.4: NoteService Integration Tests
**Priority:** P1 - Important  
**Estimated Effort:** 2-3 hours  
**Dependencies:** DbService (real), NotebookService (can be minimal)  
**File:** `internal/services/note_test.go`

**Justification:**  
NoteService handles all note queries via DuckDB. Testing ensures the SQL queries work correctly with real markdown files and the result mapping is accurate.

**Test Cases to Implement:**
```go
// SearchNotes tests
TestNoteService_SearchNotes_FindsAllNotes
TestNoteService_SearchNotes_FiltersByQuery
TestNoteService_SearchNotes_EmptyNotebook
TestNoteService_SearchNotes_ExtractsMetadata
TestNoteService_SearchNotes_NoNotebookSelected

// Count tests
TestNoteService_Count_ReturnsCorrectCount
TestNoteService_Count_EmptyNotebook
TestNoteService_Count_NoNotebookSelected

// Query tests
TestNoteService_Query_ExecutesSQL
TestNoteService_Query_ReturnsResults
```

**Implementation Notes:**
- Create temp directory with test markdown files
- Include files with frontmatter to test metadata extraction
- Test with various glob patterns

---

## Group C: Display Layer (P2 - Nice to Have)

These are low priority as they're output-only.

### Task 5.5: Display Service Tests
**Priority:** P2 - Nice to Have  
**Estimated Effort:** 1-2 hours  
**Dependencies:** None  
**File:** `internal/services/display_test.go`

**Justification:**  
Display service is pure output transformation. Failures are visible but not data-affecting. Tests ensure template fallbacks work correctly.

**Test Cases to Implement:**
```go
// Render tests
TestDisplay_Render_BasicMarkdown
TestDisplay_Render_EmptyString

// RenderTemplate tests
TestDisplay_RenderTemplate_ValidTemplate
TestDisplay_RenderTemplate_InvalidTemplate_Fallback
TestDisplay_RenderTemplate_ExecutionError_Fallback
```

**Implementation Notes:**
- Test that fallbacks return sensible output on errors
- Don't test glamour's rendering itself (external library)

---

### Task 5.6: Templates Tests
**Priority:** P2 - Nice to Have  
**Estimated Effort:** 1 hour  
**Dependencies:** None  
**File:** `internal/services/templates_test.go`

**Justification:**  
Templates are static strings. Tests ensure they handle edge cases (empty data, nil values) without panicking.

**Test Cases to Implement:**
```go
// TuiRender tests
TestTuiRender_NoteList_WithNotes
TestTuiRender_NoteList_Empty
TestTuiRender_NotebookInfo_AllFields
TestTuiRender_NotebookList_WithNotebooks
TestTuiRender_NotebookList_Empty
```

---

### Task 5.7: Logger Tests
**Priority:** P2 - Nice to Have  
**Estimated Effort:** 30 minutes  
**Dependencies:** None  
**File:** `internal/services/logger_test.go`

**Justification:**  
Logger is simple initialization. Tests ensure env vars are respected.

**Test Cases to Implement:**
```go
// InitLogger tests
TestInitLogger_DefaultLevel
TestInitLogger_DEBUG_EnvVar
TestInitLogger_LOG_LEVEL_EnvVar
TestInitLogger_LOG_LEVEL_Precedence

// Log tests
TestLog_ReturnsNamespacedLogger
```

---

## Group D: E2E/Integration (Future)

### Task 5.8: CLI E2E Tests
**Priority:** P2 - Future  
**Estimated Effort:** 4-6 hours  
**Dependencies:** All services stable  
**File:** `tests/e2e/go_smoke_test.go` or similar

**Justification:**  
End-to-end tests verify the entire CLI works as expected from a user perspective. Should be implemented after unit/integration tests are stable.

**Test Cases to Implement:**
```go
// Init command
TestCLI_Init_CreatesConfig

// Notebook commands
TestCLI_NotebookCreate_CreatesNotebook
TestCLI_NotebookList_ShowsNotebooks
TestCLI_NotebookRegister_RegistersNotebook
TestCLI_NotebookAddContext_AddsContext

// Notes commands
TestCLI_NotesList_ShowsNotes
TestCLI_NotesSearch_FiltersNotes
TestCLI_NotesAdd_CreatesNote
TestCLI_NotesRemove_RemovesNote
```

**Implementation Notes:**
- Use `exec.Command` to run the compiled binary
- Create temp HOME directory for isolated config
- Create temp notebooks for testing

---

## Summary by Priority

| Priority | Tasks | Estimated Total Effort |
|----------|-------|------------------------|
| P0 - Critical | 5.1, 5.2, 5.3 | 8-11 hours |
| P1 - Important | 5.4 | 2-3 hours |
| P2 - Nice to Have | 5.5, 5.6, 5.7, 5.8 | 6-10 hours |

## Recommended Implementation Order

```
1. Task 5.0: Setup test utilities (enables all tests)
      ↓
2. Task 5.1: ConfigService tests (no deps, enables 5.3)
      ↓
3. Task 5.2: DbService tests (no deps, enables 5.4)
      ↓
4. Task 5.3: NotebookService tests (uses 5.1, core logic)
      ↓
5. Task 5.4: NoteService tests (uses 5.2)
      ↓
6. Tasks 5.5-5.7: Display/Templates/Logger (parallel, low priority)
      ↓
7. Task 5.8: E2E tests (after all unit tests stable)
```

## Testing Utilities to Create

Before implementing tests, create these helpers in `internal/testutil/`:

```go
// testutil/notebook.go
func CreateTestNotebook(t *testing.T, dir, name string) string
func CreateTestNote(t *testing.T, notebookDir, filename, content string) string

// testutil/config.go  
func CreateTestConfig(t *testing.T, cfg services.Config) string
func SetupTestEnv(t *testing.T, envVars map[string]string)
```

This reduces boilerplate and makes tests more readable.
