# Phase 1: Core Infrastructure

## Overview
Set up the foundational Go project structure and core services required for the OpenNotes CLI.

## Status: COMPLETE

## Tasks

### 1.1 Project Setup
- [x] Create Go module structure (`go.mod` with `github.com/zenobi-us/opennotes`)
- [x] Set up directory structure: `cmd/`, `internal/services/`, `internal/middleware/`, `internal/core/`
- [x] Add required dependencies to `go.mod`:
  - `github.com/spf13/cobra` (CLI framework)
  - `github.com/knadh/koanf/v2` (configuration)
  - `github.com/duckdb/duckdb-go/v2` (database)
  - `github.com/rs/zerolog` (logging)
  - `github.com/charmbracelet/glamour` (markdown rendering)

### 1.2 LoggerService
- [x] Create `internal/services/logger.go`
- [x] Implement `InitLogger()` function
- [x] Support DEBUG env var for debug level
- [x] Support LOG_LEVEL env var for custom levels
- [x] Configure zerolog console writer for pretty output

### 1.3 ConfigService
- [x] Create `internal/services/config.go`
- [x] Define `Config` struct with `Notebooks []string` and `NotebookPath string`
- [x] Implement `GlobalConfigFile()` to return `~/.config/opennotes/config.json`
- [x] Implement `NewConfigService()` with priority: env vars > config file > defaults
- [x] Implement `Write()` method for persisting config
- [x] Support `OPENNOTES_*` environment variable prefix

### 1.4 DbService
- [x] Create `internal/services/db.go`
- [x] Implement `DbService` struct with sync.Once for lazy initialization
- [x] Implement `GetDB()` returning DuckDB connection
- [x] Install and load markdown extension on first connection
- [x] Implement `Query()` method for raw SQL queries
- [x] Implement `Close()` method for cleanup

### 1.5 Root Command
- [x] Create `cmd/root.go`
- [x] Define `rootCmd` with Cobra
- [x] Implement `PersistentPreRunE` to initialize services (interceptor pattern)
- [x] Implement `PersistentPostRun` for cleanup
- [x] Add global `--notebook` flag
- [x] Create `main.go` entry point with `cmd.Execute()`

## Dependencies
None - this is the foundation phase.

## Acceptance Criteria
- [x] `go build` succeeds
- [x] Running `./opennotes --help` shows help text
- [x] Logger outputs to stderr with timestamps
- [x] Config loads from file and env vars correctly
- [x] DuckDB connection initializes with markdown extension
