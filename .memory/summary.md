# OpenNotes - Project Memory

## Project Overview

OpenNotes is a CLI tool for managing markdown-based notes organized in notebooks. It uses DuckDB for SQL-powered search and supports templates.

## Current Status

- **Phase**: Phase 5 ALL COMPLETE
- **Last Updated**: 2026-01-09
- **Next Action**: Commit all test work

## Active Task: Go Rewrite

### Progress Review

#### Phase 1: Core Infrastructure - COMPLETE

| Task              | Status | Notes                                                       |
| ----------------- | ------ | ----------------------------------------------------------- |
| 1.1 Project Setup | Done   | go.mod, directory structure, all deps                       |
| 1.2 LoggerService | Done   | `internal/services/logger.go` with DEBUG/LOG_LEVEL support  |
| 1.3 ConfigService | Done   | `internal/services/config.go` with koanf, env vars, Write() |
| 1.4 DbService     | Done   | `internal/services/db.go` with markdown extension           |
| 1.5 Root Command  | Done   | `cmd/root.go` with interceptor pattern, `main.go`           |

#### Phase 2: Notebook Management - COMPLETE

| Task                      | Status | Notes                                         |
| ------------------------- | ------ | --------------------------------------------- |
| 2.1 NotebookService Types | Done   | All types defined in `notebook.go`            |
| 2.2 Core Methods          | Done   | HasNotebook, LoadConfig, Open                 |
| 2.3 Discovery             | Done   | Infer (3-step priority), List                 |
| 2.4 Notebook Methods      | Done   | MatchContext, AddContext, SaveConfig          |
| 2.5 CRUD                  | Done   | Create with notes dir and config              |
| 2.6 Require Middleware    | Done   | `requireNotebook()` in `notes_list.go`        |
| 2.7 Commands              | Done   | notebook, list, create, register, add-context |

#### Phase 3: Note Operations - COMPLETE

| Task                  | Status | Notes                                                                   |
| --------------------- | ------ | ----------------------------------------------------------------------- |
| 3.1 NoteService Types | Done   | Note struct in `note.go`                                                |
| 3.2 Query Methods     | Done   | SearchNotes, Count                                                      |
| 3.3 Raw Query         | Done   | Query() method                                                          |
| 3.4 Notes List Cmd    | Done   | `cmd/notes_list.go`                                                     |
| 3.5 Notes Search Cmd  | Done   | `cmd/notes_search.go`                                                   |
| 3.6 Notes Add Cmd     | Done   | `cmd/notes_add.go` with template support                                |
| 3.7 Notes Remove Cmd  | Done   | `cmd/notes_remove.go` with --force flag                                 |
| 3.8 Display Service   | Done   | `internal/services/display.go` with glamour                             |
| 3.9 Display Templates | Done   | `internal/services/templates.go` - NoteList, NotebookList, NotebookInfo |

#### Phase 4: Polish - COMPLETE

| Task                     | Status  | Notes                                                             |
| ------------------------ | ------- | ----------------------------------------------------------------- |
| 4.1 Error Handling       | Done    | Using fmt.Errorf with context                                     |
| 4.2 Input Validation     | Done    | `internal/core/schema.go` with validators                         |
| 4.3 String Utilities     | Done    | `internal/core/strings.go` (Slugify, Dedent, ObjectToFrontmatter) |
| 4.4 Help Text            | Done    | Comprehensive help with examples, env vars, aliases               |
| 4.5 Init Command         | Done    | `cmd/init.go`                                                     |
| 4.6 Integration Tests    | Partial | Core tests passing, service tests in Phase 5                      |
| 4.7 Build Config         | Done    | mise tasks: go-build, go-test, go-lint                            |
| 4.8 Config Compatibility | Done    | Go reads/writes same format as TypeScript                         |
| 4.9 Feature Parity       | Done    | All commands match, added aliases (nb, ls, rm)                    |

#### Phase 5: Testing Tasks - ALL COMPLETE

See `.memory/tasks/05-testing-tasks.md` for detailed task breakdown.  
See `.memory/research/testing-gaps.md` for gap analysis.

**Progress:**
| Task | Status | Tests Added |
|------|--------|-------------|
| 5.0 Setup (testutil) | ✅ Complete | - |
| 5.1 ConfigService | ✅ Complete | 10 tests |
| 5.2 DbService | ✅ Complete | 14 tests |
| 5.3 NotebookService | ✅ Complete | 28 tests |
| 5.4 NoteService | ✅ Complete | 16 tests |
| 5.5 Display Service | ✅ Complete | 13 tests (incl. subtests) |
| 5.6 Templates | ✅ Complete | 10 tests |
| 5.7 Logger | ✅ Complete | 17 tests (incl. subtests) |
| 5.8 E2E tests | ✅ Complete | 23 tests |

**Total: 131 tests across all packages**

### Build Status

- `go build` succeeds
- `./opennotes --help` works
- All commands functional with glamour output
- All Go tests passing

### Testing Coverage Matrix

| Package           | Files | Test Files | Test Count | Coverage                 |
| ----------------- | ----- | ---------- | ---------- | ------------------------ |
| internal/core     | 2     | 2          | 27         | Good                     |
| internal/services | 7     | 7          | 108        | Good (P0/P1/P2 complete) |
| internal/testutil | 2     | 0          | (helpers)  | N/A                      |
| tests/e2e         | 1     | 1          | 23         | Good (E2E complete)      |
| cmd               | 12    | 0          | 0          | Covered by E2E tests     |

### Uncommitted Work

New files (untracked):

- `internal/services/note_test.go` - NoteService tests (16 tests)
- `internal/services/display_test.go` - Display tests (13 tests)
- `internal/services/templates_test.go` - Templates tests (10 tests)
- `internal/services/logger_test.go` - Logger tests (17 tests)
- `tests/e2e/go_smoke_test.go` - E2E CLI tests (23 tests)

### Remaining Work

**Immediate:**

1. Commit all testing work

**Optional Future:**

1. Performance optimization
2. Documentation updates

### Key Files

```
cmd/
  init.go                   # init command
  notebook.go               # notebook parent command
  notebook_addcontext.go    # add-context subcommand
  notebook_create.go        # create subcommand + displayNotebookInfo
  notebook_list.go          # list subcommand + displayNotebookList
  notebook_register.go      # register subcommand
  notes.go                  # notes parent command
  notes_add.go              # add with template support (NEW)
  notes_list.go             # list + requireNotebook + displayNoteList
  notes_remove.go           # remove with confirmation (NEW)
  notes_search.go           # search subcommand
  root.go                   # root with service init

internal/core/
  schema.go                 # Validation helpers (NEW)
  schema_test.go            # Validation tests (NEW) - 14 cases
  strings.go                # Slugify, Dedent, ObjectToFrontmatter (NEW)
  strings_test.go           # String utility tests (NEW) - 13 cases

internal/services/
  config.go                 # ConfigService with koanf
  config_test.go            # ConfigService tests (10 tests) ✅
  db.go                     # DbService with DuckDB
  db_test.go                # DbService tests (14 tests) ✅
  display.go                # Display service with glamour
  display_test.go           # Display tests (13 tests) ✅
  logger.go                 # zerolog logging
  logger_test.go            # Logger tests (17 tests) ✅
  notebook.go               # NotebookService (complete)
  notebook_test.go          # NotebookService tests (28 tests) ✅
  note.go                   # NoteService (complete)
  note_test.go              # NoteService tests (16 tests) ✅
  templates.go              # TUI templates
  templates_test.go         # Templates tests (10 tests) ✅

internal/testutil/
  config.go                 # Test helper utilities (NEW)
  notebook.go               # Test helper utilities (NEW)

.mise/tasks/
  go-build                  # Build Go binary (NEW)
  go-test                   # Run Go tests (NEW)
  go-lint                   # Lint Go code (NEW)

main.go                     # Entry point
```

### Memory Structure

```
.memory/
  summary.md                # This file - project overview
  spec.md                   # Original specification
  tasks/
    01-core-infrastructure.md   # Phase 1 tasks (COMPLETE)
    02-notebook-management.md   # Phase 2 tasks (COMPLETE)
    03-note-operations.md       # Phase 3 tasks (COMPLETE)
    04-polish.md                # Phase 4 tasks (PARTIAL)
    05-testing-tasks.md         # Phase 5 tasks (PLANNED)
  research/
    testing-gaps.md             # Testing gap analysis
```

### Principles

- Preserve Node.js version (don't remove src/)
- Match same structure and patterns as Node code
- Config priority: env vars > config file > defaults
- Module: github.com/zenobi-us/opennotes
