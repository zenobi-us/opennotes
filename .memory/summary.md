# OpenNotes - Project Memory

## Project Overview
OpenNotes is a CLI tool for managing markdown-based notes organized in notebooks. It uses DuckDB for SQL-powered search and supports templates.

## Current Status
- **Phase**: Phase 1-4 (Partial), Phase 5 Planned
- **Last Updated**: 2026-01-09
- **Next Action**: Implement P0 testing tasks (ConfigService, DbService, NotebookService)

## Active Task: Go Rewrite

### Progress Review

#### Phase 1: Core Infrastructure - COMPLETE
| Task | Status | Notes |
|------|--------|-------|
| 1.1 Project Setup | Done | go.mod, directory structure, all deps |
| 1.2 LoggerService | Done | `internal/services/logger.go` with DEBUG/LOG_LEVEL support |
| 1.3 ConfigService | Done | `internal/services/config.go` with koanf, env vars, Write() |
| 1.4 DbService | Done | `internal/services/db.go` with markdown extension |
| 1.5 Root Command | Done | `cmd/root.go` with interceptor pattern, `main.go` |

#### Phase 2: Notebook Management - COMPLETE
| Task | Status | Notes |
|------|--------|-------|
| 2.1 NotebookService Types | Done | All types defined in `notebook.go` |
| 2.2 Core Methods | Done | HasNotebook, LoadConfig, Open |
| 2.3 Discovery | Done | Infer (3-step priority), List |
| 2.4 Notebook Methods | Done | MatchContext, AddContext, SaveConfig |
| 2.5 CRUD | Done | Create with notes dir and config |
| 2.6 Require Middleware | Done | `requireNotebook()` in `notes_list.go` |
| 2.7 Commands | Done | notebook, list, create, register, add-context |

#### Phase 3: Note Operations - COMPLETE
| Task | Status | Notes |
|------|--------|-------|
| 3.1 NoteService Types | Done | Note struct in `note.go` |
| 3.2 Query Methods | Done | SearchNotes, Count |
| 3.3 Raw Query | Done | Query() method |
| 3.4 Notes List Cmd | Done | `cmd/notes_list.go` |
| 3.5 Notes Search Cmd | Done | `cmd/notes_search.go` |
| 3.6 Notes Add Cmd | Done | `cmd/notes_add.go` with template support |
| 3.7 Notes Remove Cmd | Done | `cmd/notes_remove.go` with --force flag |
| 3.8 Display Service | Done | `internal/services/display.go` with glamour |
| 3.9 Display Templates | Done | `internal/services/templates.go` - NoteList, NotebookList, NotebookInfo |

#### Phase 4: Polish - PARTIAL
| Task | Status | Notes |
|------|--------|-------|
| 4.1 Error Handling | Done | Using fmt.Errorf with context |
| 4.2 Input Validation | Done | `internal/core/schema.go` with validators |
| 4.3 String Utilities | Done | `internal/core/strings.go` (Slugify, Dedent, ObjectToFrontmatter) |
| 4.4 Help Text | TODO | Need comprehensive help text |
| 4.5 Init Command | Done | `cmd/init.go` |
| 4.6 Integration Tests | Partial | Core tests passing, service tests in Phase 5 |
| 4.7 Build Config | Done | mise tasks: go-build, go-test, go-lint |
| 4.8 Config Compatibility | TODO | Need verification |
| 4.9 Feature Parity | TODO | Need comparison |

#### Phase 5: Testing Tasks - PLANNED
See `.memory/tasks/05-testing-tasks.md` for detailed task breakdown.  
See `.memory/research/testing-gaps.md` for gap analysis.

**Priority Summary:**
| Priority | Tasks | Files | Est. Effort |
|----------|-------|-------|-------------|
| P0 Critical | 5.1-5.3 | config_test.go, db_test.go, notebook_test.go | 8-11 hrs |
| P1 Important | 5.4 | note_test.go | 2-3 hrs |
| P2 Nice to Have | 5.5-5.8 | display_test.go, templates_test.go, logger_test.go, e2e | 6-10 hrs |

### Build Status
- `go build` succeeds
- `./opennotes --help` works
- All commands functional with glamour output
- Core tests passing (`internal/core` tests: 27 cases)
- Service tests: **MISSING** (0 test files)

### Testing Coverage Matrix

| Package | Files | Test Files | Coverage |
|---------|-------|------------|----------|
| internal/core | 2 | 2 | Good |
| internal/services | 7 | 0 | **None** |
| cmd | 12 | 0 | **None** |

### Uncommitted Work
Files modified but not committed:
- `cmd/notebook_create.go` - Updated display integration
- `cmd/notebook_list.go` - Updated display integration
- `cmd/notes_list.go` - Updated display integration
- `go.mod` / `go.sum` - Dependency updates
- `internal/services/note.go` - Minor fixes

New files (untracked):
- `.mise/tasks/go-build` - Build Go binary task
- `.mise/tasks/go-lint` - Lint Go code task
- `.mise/tasks/go-test` - Run Go tests task
- `cmd/notes_add.go` - Add notes with template support
- `cmd/notes_remove.go` - Remove notes with --force flag
- `internal/core/` - Validation and string utilities
- `internal/services/display.go` - Display service with glamour
- `internal/services/templates.go` - TUI templates (separated from display)

### Remaining Work

**Immediate (This Sprint):**
1. Commit uncommitted Go work
2. Task 5.1: ConfigService tests (P0)
3. Task 5.2: DbService tests (P0)
4. Task 5.3: NotebookService tests (P0)

**Next Sprint:**
1. Task 5.4: NoteService tests (P1)
2. Task 4.4: Help text improvements
3. Task 4.8: Config compatibility verification

**Future:**
1. Tasks 5.5-5.7: Display/Templates/Logger tests
2. Task 5.8: E2E tests
3. Task 4.9: Feature parity checklist

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
  config.go                 # ConfigService with koanf - NEEDS TESTS
  db.go                     # DbService with DuckDB - NEEDS TESTS
  display.go                # Display service with glamour (NEW) - NEEDS TESTS
  logger.go                 # zerolog logging - NEEDS TESTS
  notebook.go               # NotebookService (complete) - NEEDS TESTS
  note.go                   # NoteService (complete) - NEEDS TESTS
  templates.go              # TUI templates (NEW) - NEEDS TESTS

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
