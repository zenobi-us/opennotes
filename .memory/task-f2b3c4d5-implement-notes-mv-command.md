---
id: f2b3c4d5
title: Implement notes mv command
type: "task"
created_at: 2026-02-27T08:45:00+10:30
updated_at: 2026-02-28T01:26:00+10:30
status: completed
epic_id: a7c3d9e1
story_id: 2b3c4d5e
assigned_to: 3219612528192551
---

# Implement notes mv command

## Objective
Add `jot notes mv <from> <to>` with safe default no-overwrite behavior and explicit `--force` override.

## Steps
1. Add command scaffold with `--force` and `--format` support.
2. Implement source/destination normalization and root-bound path checks.
3. Implement move behavior with conflict detection and overwrite guard.
4. Add tests for source missing, destination exists, and force overwrite paths.

## Exit Criteria
- `notes mv` moves notes within notebook root safely.
- Destination overwrite is blocked unless `--force`.
- Exit codes distinguish not-found vs conflict.
- Tests cover traversal and conflict behavior.

## Actual Outcome
- Implemented `jot notes mv <from> <to>` with alias `move` in `cmd/notes_mv.go`.
- Added root-bound source/destination normalization (`.md` auto-suffix) and traversal rejection.
- Added safe default conflict behavior: destination overwrite is blocked unless `--force`.
- Added explicit conflict/not-found exit-code mapping via `ExitCodeConflict` and `ExitCodeNotFound`.
- Added machine-readable output (`--format list|json`) for both success and failure flows.
- Added tests in `cmd/notes_mv_test.go` for traversal rejection, missing source, destination conflict, and force-overwrite with content preservation.

## Lessons Learned
- Shared path resolution patterns across `update`, `get`, and `mv` reduce boundary-check drift.
- Conflict/nonexistent conditions should be encoded explicitly for automation-friendly behavior.

## Notes
- @TODO(epic:b2f4e6a8): workflow state-transition policies deferred.
