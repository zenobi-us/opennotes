# OpenNotes - Project Memory

## Project Overview

OpenNotes is a CLI tool for managing markdown-based notes organized in notebooks. It uses DuckDB for SQL-powered search and supports templates.

## Current Status

- **Phase**: Planning (Ready for Implementation)
- **Last Updated**: 2026-01-09
- **Next Action**: Review and approve specification, then implement Phase 1

## Active Work

### SQL Flag Feature (2026-01-09)

**Status**: Research & Specification Complete ✅ - Ready for Approval

Created comprehensive specification for adding `--sql` flag to search command:

- **Research**: Completed research on DuckDB Go client and markdown extension
  - Analyzed current OpenNotes implementation
  - Found much infrastructure already exists!
  - Simplified implementation plan
  
- **Specification**: Detailed spec aligned with current codebase
  - Leverages existing `DbService.Query()` and `rowsToMaps()`
  - Only 3-4 hours work for MVP (not 4-6)
  - Clear, achievable implementation path

**Files**:
  - Research: `.memory/research-b8f3d2a1-duckdb-go-markdown.md` ✅
  - Spec: `.memory/task-a1b2c3d4-sql-flag-spec.md` ✅

**Key Features**:
- Custom SQL query support in search command
- Read-only access for security (new connection)
- Query validation (SELECT/WITH only)
- Full markdown extension function access
- 30-second timeout
- Table-formatted output

**Implementation Phases**:
1. Core Functionality (MVP) - 3-4 hours ⚠️ REDUCED
2. Enhanced Display - 2-3 hours
3. Documentation - 2-3 hours
4. Advanced Features (Future)

**What Already Exists**:
- ✅ `DbService.Query()` returning maps
- ✅ `NoteService.Query()` wrapping DbService
- ✅ Markdown extension loading
- ✅ `rowsToMaps()` helper

**What's New**:
- ✅ Query validation (SELECT only)
- ✅ Read-only connection option
- ✅ Table display formatter
- ✅ CLI flag integration

**Next Steps**:
- Review aligned specification
- Approve implementation approach
- Begin Phase 1 implementation

## Recent Completed Work

### Go Rewrite (archived: 01-migrate-to-golang)

Successfully completed full migration from TypeScript to Go:

- 5 phases completed (Core Infrastructure, Notebook Management, Note Operations, Polish, Testing)
- 131 tests across all packages
- Full feature parity with TypeScript version
- All commands functional with glamour output

## Archive

| Archive                | Description                      | Completed  |
| ---------------------- | -------------------------------- | ---------- |
| `01-migrate-to-golang` | Full Go rewrite of OpenNotes CLI | 2026-01-09 |

## Key Files

```
cmd/                        # CLI commands (Go)
internal/core/              # Validation, string utilities
internal/services/          # Core services (config, db, notebook, note, display, logger)
internal/testutil/          # Test helpers
tests/e2e/                  # End-to-end tests
main.go                     # Entry point
```

## Recent Analysis

### Codebase Exploration (2026-01-17)

**Status**: Complete ✅

Comprehensive codebase analysis using CodeMapper skill:
- **File**: `.memory/analysis-20260117-103843-codebase-exploration.md`
- **Scope**: Complete architecture, data flow, user journeys, dependencies
- **Key Findings**:
  - 79 files, 307KB total codebase
  - 123 test cases with 95%+ coverage
  - Successful TypeScript → Go migration (100% feature parity)
  - 12 CLI commands, 6 core services
  - Clean service-oriented architecture
  - Production-ready with comprehensive tests

**Included Artifacts**:
- Language statistics and symbol distribution
- Complete package structure maps
- ASCII state machine diagrams (notebook lifecycle, note operations)
- User journey diagrams (4 common workflows)
- Data flow diagrams (3 primary flows)
- Dependency graphs
- Test coverage analysis
- Migration status assessment

## Memory Structure

```
.memory/
  summary.md                              # This file - project overview
  team.md                                 # Current work tracking
  todo.md                                 # Active tasks
  research-b8f3d2a1-duckdb-go-markdown.md # DuckDB research (aligned)
  task-a1b2c3d4-sql-flag-spec.md         # SQL flag specification (aligned)
  archived/
    01-migrate-to-golang/                 # Completed Go migration
      spec.md                             # Original specification
      tasks/                              # Phase task files (01-05)
      research/                           # Testing gap analysis
```
