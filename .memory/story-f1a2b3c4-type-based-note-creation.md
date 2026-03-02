---
id: f1a2b3c4
title: Type-based Note Creation
created_at: 2026-03-02T13:43:00+10:30
updated_at: 2026-03-02T13:43:00+10:30
status: proposed
epic_id: c5d7e9b1
priority: high
story_points: 5
test_coverage: none
---

# Type-based Note Creation

## User Story

**As a** Jot user  
**I want to** run `jot notes add --type task "Fix the bug"`  
**So that** a note is created in the right folder without specifying the path

## Acceptance Criteria

- [ ] `--type` flag added to `notes add` command
- [ ] Type maps to group via `type` or `aliases` field in group config
- [ ] Note created in group's directory with auto-generated filename
- [ ] Error with helpful message if type not found
- [ ] Works with existing `--notebook` flag

## Context

Currently, users must know the exact path or directory structure to create notes. This creates friction for users who think in terms of note "types" (task, meeting, idea) rather than filesystem locations.

Groups already define collections of related notes. This story extends that by allowing users to reference groups by a semantic type name, enabling intent-level commands that Jot resolves to the correct location.

## Out of Scope

- Multi-notebook type resolution (type must be unambiguous within selected notebook)
- Type aliases across notebooks
- Creating new groups via the `--type` flag

## Tasks

- [task-c3d4e5f6](task-c3d4e5f6-add-type-flag.md) — Add --type flag to notes add (Phase 1)
- [task-d4e5f6a7](task-d4e5f6a7-type-to-group-resolver.md) — Type-to-group resolver (Phase 1)

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| 1 | `--type` flag added to `notes add` command | notes_add_type_test.go / TestTypeFlag | ⬜ |
| 2 | Type maps to group via `type` or `aliases` | notes_add_type_test.go / TestTypeMapping | ⬜ |
| 3 | Note created in group's directory | notes_add_type_test.go / TestNoteInGroupDir | ⬜ |
| 4 | Error with helpful message if type not found | notes_add_type_test.go / TestTypeNotFoundError | ⬜ |
| 5 | Works with existing `--notebook` flag | notes_add_type_test.go / TestTypeWithNotebookFlag | ⬜ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are completed_

## Notes

- Type lookup order: exact `type` field match → `aliases` array match → error
- Consider case-insensitive matching for better UX
- Error message should list available types in current notebook
