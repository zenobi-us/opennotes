# Phase 3: Note Operations

## Overview
Implement the NoteService and all note-related commands with DuckDB markdown queries.

## Status: COMPLETE (9/9 tasks complete)

## Tasks

### 3.1 NoteService Types
- [x] Create `internal/services/note.go`
- [x] Define `Note` struct:
  - `File` struct with `Filepath` and `Relative`
  - `Content` string
  - `Metadata` map[string]any
- [x] Define `NoteService` struct (configService, dbService, notebookPath, log)

### 3.2 NoteService Query Methods
- [x] Implement `NewNoteService()`
- [x] Implement `SearchNotes(query string)` - returns all notes
  - Uses `read_markdown()` DuckDB function with `include_filepath:=true`
  - Glob pattern: `notebookPath/**/*.md`
  - Returns Note slice with file info and metadata
- [x] Implement `Count()` - returns note count
  - Uses `COUNT(*)` with `read_markdown()`

### 3.3 NoteService Raw Query
- [x] Implement `Query(sql string)` - executes raw SQL
  - Returns `[]map[string]any` for flexibility
  - Useful for custom queries

### 3.4 Notes List Command
- [x] Create `cmd/notes.go` - parent command
- [x] Create `cmd/notes_list.go`
  - Calls `requireNotebook()` middleware
  - Lists all notes via `nb.Notes.SearchNotes("")`
  - Displays with glamour-rendered templates

### 3.5 Notes Search Command
- [x] Create `cmd/notes_search.go`
  - Takes query argument
  - Calls `requireNotebook()` middleware
  - Searches notes via `nb.Notes.SearchNotes(query)`
  - Displays results

### 3.6 Notes Add Command
- [x] Create `cmd/notes_add.go`
  - Creates new markdown file in notebook
  - Supports `--template` flag for template selection
  - Supports `--title` flag for note title
  - Auto-generates filename from title or timestamp
  - Uses `core.Slugify` for filename generation
  - Uses `core.ValidateNoteName` for input validation

### 3.7 Notes Remove Command
- [x] Create `cmd/notes_remove.go`
  - Removes note file from notebook
  - Confirmation prompt (use `--force` to skip)
  - Validates note exists before removal

### 3.8 Display Service
- [x] Create `internal/services/display.go`
- [x] Implement `NewDisplay()` - creates glamour renderer
- [x] Implement `Render(markdown)` - renders plain markdown
- [x] Implement `RenderTemplate(tmpl, ctx)` - combines Go templates with glamour
- [x] Implement `TuiRender()` convenience function
- [x] Configure glamour with auto style and word wrap (100 chars)

### 3.9 Display Templates
- [x] Create `Templates.NoteList` - lists notes with count and links
- [x] Create `Templates.NoteDetail` - shows note with metadata and content
- [x] Create `Templates.NotebookInfo` - shows notebook config details
- [x] Create `Templates.NotebookList` - lists all notebooks with info

## Dependencies
- Phase 1: Core Infrastructure (DbService for queries) - COMPLETE
- Phase 2: Notebook Management (NotebookService, middleware) - COMPLETE

## Acceptance Criteria
- [x] `opennotes notes list` shows all notes in current notebook
- [x] `opennotes notes search "query"` filters notes
- [x] `opennotes notes add` creates a new note
- [x] `opennotes notes remove <note>` removes a note
- [x] Output is beautifully formatted via glamour
