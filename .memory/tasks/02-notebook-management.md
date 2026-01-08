# Phase 2: Notebook Management

## Overview
Implement the NotebookService and all notebook-related commands.

## Status: COMPLETE

## Tasks

### 2.1 NotebookService Types
- [x] Create `internal/services/notebook.go`
- [x] Define `NotebookGroup` struct (Name, Globs, Metadata, Template)
- [x] Define `StoredNotebookConfig` struct (Root, Name, Contexts, Templates, Groups)
- [x] Define `NotebookConfig` struct (embeds StoredNotebookConfig + Path)
- [x] Define `Notebook` struct (Config, Notes *NoteService)
- [x] Define `NotebookService` struct (configService, dbService, log)

### 2.2 NotebookService Core Methods
- [x] Implement `NewNotebookService()`
- [x] Implement `configFilePath()` helper (returns `.opennotes.json` path)
- [x] Implement `HasNotebook()` - checks if path contains notebook config
- [x] Implement `LoadConfig()` - reads and parses `.opennotes.json`
- [x] Implement `Open()` - loads notebook with NoteService

### 2.3 NotebookService Discovery
- [x] Implement `Infer()` - discovers notebook from context
  - Priority 1: Declared NotebookPath from config
  - Priority 2: Context matching from registered notebooks
  - Priority 3: Ancestor directory search
- [x] Implement `List()` - returns all known notebooks
  - From registered paths in config
  - From ancestor directories (deduplicated)

### 2.4 Notebook Methods
- [x] Implement `Notebook.MatchContext()` - checks if path matches contexts
- [x] Implement `Notebook.AddContext()` - adds context path to notebook
- [x] Implement `Notebook.SaveConfig()` - writes config to disk with optional global registration

### 2.5 NotebookService CRUD
- [x] Implement `Create()` - creates new notebook
  - Creates `.notes/` directory
  - Creates `.opennotes.json` config
  - Optionally registers globally

### 2.6 Require Notebook Middleware
- [x] Implement `requireNotebook()` function (in `cmd/notes_list.go`)
  - Checks `--notebook` flag first
  - Falls back to `notebookService.Infer()`
  - Returns error with helpful create message if none found

Note: Middleware implemented inline in cmd package rather than separate `internal/middleware/` - consider refactoring if needed.

### 2.7 Notebook Commands
- [x] Create `cmd/notebook.go` - parent command (shows current notebook info)
- [x] Create `cmd/notebook_list.go` - lists all notebooks
- [x] Create `cmd/notebook_create.go` - creates notebook
  - Flags: `--name` (required), `--path`, `--global`
- [x] Create `cmd/notebook_register.go` - registers existing notebook globally
- [x] Create `cmd/notebook_addcontext.go` - adds context path to notebook

## Dependencies
- Phase 1: Core Infrastructure (ConfigService, DbService, Logger) - COMPLETE

## Acceptance Criteria
- [x] `opennotes notebook create --name "Test"` creates a notebook
- [x] `opennotes notebook list` shows all notebooks
- [x] `opennotes notebook` shows current notebook info or prompts to create
- [x] Notebook discovery works via ancestor search
- [x] Context matching works for registered notebooks
