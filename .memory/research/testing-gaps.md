# Testing Gaps Analysis (Research)

## Overview
Analysis of testing coverage gaps in the Go implementation, categorized by priority and grouped by dependency.

**Status:** Complete  
**Date:** 2026-01-09  
**Related Tasks:** See `.memory/tasks/05-testing-tasks.md` for implementation tasks.

---

## Research Checklist

### Current State
- [x] Review tested code (`internal/core/`)
- [x] Identify untested services (`internal/services/`)
- [x] Identify untested commands (`cmd/`)

### Gap Analysis
- [x] Analyze ConfigService gaps (P0)
- [x] Analyze DbService gaps (P0)
- [x] Analyze NotebookService gaps (P0)
- [x] Analyze NoteService gaps (P1)
- [x] Analyze Display Service gaps (P2)
- [x] Analyze Templates gaps (P2)
- [x] Analyze Logger gaps (P2)

### Documentation
- [x] Document testing dependencies
- [x] Define priority classification
- [x] Recommend testing strategy

---

## Current Testing State

### What's Tested (internal/core/)
- ✅ `schema_test.go` - Validation helpers (4 test functions, 14 test cases)
- ✅ `strings_test.go` - String utilities (3 test functions, 13 test cases)

### What's NOT Tested (internal/services/)
- ❌ `config.go` - ConfigService (0 tests)
- ❌ `db.go` - DbService (0 tests)
- ❌ `notebook.go` - NotebookService (0 tests)
- ❌ `note.go` - NoteService (0 tests)
- ❌ `display.go` - Display service (0 tests)
- ❌ `templates.go` - TuiRender function (0 tests)
- ❌ `logger.go` - Logger initialization (0 tests)

### What's NOT Tested (cmd/)
- ❌ Command handlers (all 12 command files)
- ❌ Integration/E2E tests

## Gap Analysis by Service

### 1. ConfigService (`config.go`) - HIGH PRIORITY
**Lines of Code:** 129  
**Complexity:** Medium  
**Untested Functions:**
- `GlobalConfigFile()` - Returns platform config path
- `NewConfigService()` - Configuration loading with priority chain
- `Write()` - Config persistence
- `Path()` - Simple getter

**Critical Gaps:**
- No test for env var override (OPENNOTES_* prefix)
- No test for config file loading/parsing
- No test for default values
- No test for Write() file operations
- No test for error handling on malformed config

**Risk:** High - Config errors silently fall back to defaults, could mask real issues

### 2. DbService (`db.go`) - HIGH PRIORITY
**Lines of Code:** 132  
**Complexity:** High (external dependency)  
**Untested Functions:**
- `NewDbService()` - Service initialization
- `GetDB()` - Lazy DB initialization with markdown extension
- `Query()` - Raw SQL execution
- `Close()` - Cleanup
- `rowsToMaps()` - Internal helper

**Critical Gaps:**
- No test for DuckDB connection
- No test for markdown extension loading
- No test for Query() with various SQL
- No test for lazy initialization (sync.Once behavior)
- No test for concurrent access (mutex behavior)
- No test for Close() cleanup

**Risk:** High - Database layer is critical path; failures would break all note operations

### 3. NotebookService (`notebook.go`) - HIGH PRIORITY
**Lines of Code:** 318  
**Complexity:** High  
**Untested Functions:**
- `NewNotebookService()` - Service initialization
- `configFilePath()` - Path helper
- `HasNotebook()` - Existence check
- `LoadConfig()` - Config parsing and path resolution
- `Open()` - Notebook loading
- `Create()` - Notebook creation with directory structure
- `Infer()` - Discovery with 3-step priority
- `List()` - All notebooks enumeration
- `MatchContext()` - Context path matching
- `AddContext()` - Context addition
- `SaveConfig()` - Config persistence with optional registration

**Critical Gaps:**
- No test for 3-step notebook inference priority
- No test for context matching logic
- No test for notebook creation (directory + config file)
- No test for config file parsing/validation
- No test for global registration
- No test for deduplication in List()
- No test for ancestor directory walking

**Risk:** Very High - Core business logic, notebook discovery is used by every command

### 4. NoteService (`note.go`) - MEDIUM PRIORITY
**Lines of Code:** 149  
**Complexity:** Medium  
**Untested Functions:**
- `NewNoteService()` - Service initialization
- `SearchNotes()` - Note query with DuckDB
- `Count()` - Note counting
- `Query()` - Raw SQL passthrough

**Critical Gaps:**
- No test for SearchNotes() result mapping
- No test for query filtering
- No test for empty notebook handling
- No test for metadata extraction from columns
- No test for glob pattern construction

**Risk:** Medium - Depends on DbService; errors would affect list/search commands

### 5. Display Service (`display.go`) - LOW PRIORITY
**Lines of Code:** 57  
**Complexity:** Low  
**Untested Functions:**
- `NewDisplay()` - Glamour renderer setup
- `Render()` - Plain markdown rendering
- `RenderTemplate()` - Template + glamour rendering

**Critical Gaps:**
- No test for template parsing fallback
- No test for glamour rendering
- No test for word wrap behavior

**Risk:** Low - Output formatting only; failures visible but not data-affecting

### 6. Templates (`templates.go`) - LOW PRIORITY
**Lines of Code:** 106  
**Complexity:** Low  
**Untested Functions:**
- `TuiRender()` - Convenience rendering function
- Template strings (NoteList, NoteDetail, NotebookInfo, NotebookList)

**Critical Gaps:**
- No test for template execution with various contexts
- No test for edge cases (empty data, nil values)
- No test for fallback behavior

**Risk:** Low - Templates are static; issues would be obvious in output

### 7. Logger (`logger.go`) - LOW PRIORITY
**Lines of Code:** 40  
**Complexity:** Low  
**Untested Functions:**
- `InitLogger()` - Logger setup
- `Log()` - Namespaced logger factory

**Critical Gaps:**
- No test for DEBUG env var handling
- No test for LOG_LEVEL env var handling
- No test for Log() namespace injection

**Risk:** Very Low - Logging is supplementary; failures would just reduce observability

## Testing Dependencies

```
                    ┌─────────────────┐
                    │   LoggerService │  (no deps, low priority)
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │  ConfigService  │  (depends on logger)
                    └────────┬────────┘
                             │
              ┌──────────────┼────────────────┐
              │              │                │
     ┌────────▼─────┐ ┌──────▼─────────┐      │
     │  DbService   │ │NotebookService │◄─────┘
     └────────┬─────┘ └──────┬─────────┘
              │              │ 
              │      ┌───────▼───────┐
              │      │  NoteService  │
              │      └───────┬───────┘
              │              │
              └──────────────┼──────────────┐
                             │              │
                    ┌────────▼────────┐     │
                    │  Display/TUI    │◄────┘
                    └─────────────────┘
```

## Priority Classification

### P0 - Critical (Must Have for Production)
1. **ConfigService tests** - Foundation of all config loading
2. **NotebookService tests** - Core discovery and CRUD
3. **DbService tests** - Database connectivity

### P1 - Important (Should Have)
4. **NoteService tests** - Query operations

### P2 - Nice to Have
5. **Display tests** - Output formatting
6. **Templates tests** - Template rendering
7. **Logger tests** - Observability

## Recommended Testing Strategy

### Unit Tests (Mock Dependencies)
- ConfigService: Mock filesystem operations
- NotebookService: Mock ConfigService and DbService
- NoteService: Mock DbService

### Integration Tests (Real Dependencies)
- DbService: Test with real DuckDB (in-memory)
- NotebookService + NoteService: Test with temp directories

### E2E Tests
- Full CLI command execution with temp notebooks
