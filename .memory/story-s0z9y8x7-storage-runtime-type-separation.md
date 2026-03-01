---
id: s0z9y8x7
title: Separate Storage Types from Runtime Types
created_at: 2026-03-01T12:20:00+10:30
updated_at: 2026-03-01T12:20:00+10:30
status: proposed
epic_id: 7c631839
phase_id: null
priority: critical
story_points: 5
test_coverage: none
---

# Separate Storage Types from Runtime Types

## User Story

As a **maintainer**, I want storage types (what's persisted to JSON) clearly separated from runtime types (what the application uses internally) so that schema generation has an unambiguous target and the codebase has clear architectural boundaries.

## Acceptance Criteria

- [ ] AC1: All `.jot.json` storage types are in a dedicated package (e.g., `internal/schema/`)
- [ ] AC2: All `config.json` (global) storage types are in the same dedicated package
- [ ] AC3: Runtime types embed storage types (composition, not duplication)
- [ ] AC4: Runtime-only fields are clearly marked and NOT in storage types
- [ ] AC5: Existing tests pass without modification (behavioral compatibility)
- [ ] AC6: Schema generator can target the storage package exclusively
- [ ] AC7: Package has `// Code generated` guard if types will be schema-generated

## Context

Currently, storage and runtime concerns are mixed:

**Notebook config** (partial separation):
```go
// Storage type - good
type StoredNotebookConfig struct { ... }

// Runtime type embeds storage - good
type NotebookConfig struct {
    StoredNotebookConfig
    Path string `json:"-"` // runtime-only
}
```

**Global config** (no separation):
```go
// Same struct for storage AND runtime - bad
type Config struct {
    Notebooks    []string
    NotebookPath string
}
```

**Problem**: Schema generator would include runtime fields if it reflects on `NotebookConfig`. We need a package containing ONLY storage types.

### Proposed Structure

```
internal/
├── schema/                    # NEW: Storage types only
│   ├── notebook.go            # StoredNotebookConfig, NotebookGroup, etc.
│   ├── workflow.go            # WorkflowDefinition, WorkflowState
│   ├── global.go              # StoredGlobalConfig (new)
│   └── doc.go                 # Package docs explaining purpose
├── services/
│   ├── notebook.go            # NotebookConfig embeds schema.StoredNotebookConfig
│   ├── config.go              # ConfigService embeds schema.StoredGlobalConfig
│   └── ...
```

## Out of Scope

- Actual schema generation (separate story)
- CI integration (separate story)
- Changing JSON field names or structure (purely internal refactor)

## Tasks

_To be populated during task breakdown_

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| AC5 | All existing tests pass | `mise run test` | ❌ |
| AC6 | Schema types importable | Manual verification | ❌ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are created_

## Notes

- This is a **refactor with no behavioral change** — all tests must pass as-is
- Consider using `go:generate` comment in schema package to trigger schema generation
- Import cycle prevention: schema package must have zero dependencies on services
- Alternative: Could use build tags instead of separate package, but package is cleaner
- Related: [knowledge-config-transformation.md](knowledge-config-transformation.md) documents current architecture
