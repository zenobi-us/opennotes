---
id: h3c4d5e6
title: Group Filename Patterns
created_at: 2026-03-02T13:43:00+10:30
updated_at: 2026-03-02T13:43:00+10:30
status: proposed
epic_id: c5d7e9b1
priority: high
story_points: 8
test_coverage: none
---

# Group Filename Patterns

## User Story

**As a** Jot user  
**I want** groups to define filename patterns like `task-{{nanoid 8}}-{{slug .title}}.md`  
**So that** I get consistent, unique filenames automatically

## Acceptance Criteria

- [ ] `filename_format` field added to group schema
- [ ] Processed via gomplate with `jot` namespace functions
- [ ] Available functions: `jot.Slug`, `jot.NanoID`, `jot.Timestamp`, `jot.DatePath`, `jot.UUID`, `jot.Now`
- [ ] Top-level aliases work: `{{ .title | slug }}`
- [ ] Fallback to `{{ slug .title }}.md` when no format specified
- [ ] Filename collisions detected and error raised

## Context

Different note types benefit from different naming conventions. Tasks might want `task-<id>-<title>.md` for uniqueness, while meeting notes prefer `2026-03-02-standup.md` for chronological sorting. This story enables groups to define their own filename patterns using gomplate templating.

Research in [research-c5d7e9b1](research-c5d7e9b1-gomplate-custom-functions.md) established the gomplate approach and custom function namespace design.

## Out of Scope

- Directory path templating (only filename, not full path)
- Custom user-defined functions
- Filename pattern inheritance between groups

## Tasks

- [task-e5f6a7b8](task-e5f6a7b8-filename-format-schema.md) — Add filename_format to group schema (Phase 2)
- [task-f6a7b8c9](task-f6a7b8c9-gomplate-integration.md) — Gomplate template engine integration (Phase 2)
- [task-a7b8c9d0](task-a7b8c9d0-jot-namespace-funcs.md) — Implement jot namespace functions (Phase 2)
- [task-b8c9d0e1](task-b8c9d0e1-filename-collision-detection.md) — Filename collision detection (Phase 2)

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| 1 | `filename_format` field in group schema | group_schema_test.go / TestFilenameFormatField | ⬜ |
| 2 | Processed via gomplate with jot namespace | filename_pattern_test.go / TestGomplateProcessing | ⬜ |
| 3 | All jot namespace functions available | filename_pattern_test.go / TestJotFunctions | ⬜ |
| 4 | Top-level aliases work | filename_pattern_test.go / TestTopLevelAliases | ⬜ |
| 5 | Fallback when no format specified | filename_pattern_test.go / TestDefaultFallback | ⬜ |
| 6 | Filename collisions detected | filename_pattern_test.go / TestCollisionDetection | ⬜ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are completed_

## Notes

- gomplate functions in `jot` namespace:
  - `jot.Slug` - Unicode-safe slugification (depends on story-g2b3c4d5)
  - `jot.NanoID` - Generate nanoid of specified length
  - `jot.Timestamp` - Unix timestamp
  - `jot.DatePath` - Date-based path segment (e.g., "2026/03/02")
  - `jot.UUID` - UUID v4
  - `jot.Now` - Current time with format string
- Collision handling: error immediately vs auto-suffix approach (recommend error)
- Consider validation of filename_format at notebook load time
