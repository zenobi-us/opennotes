---
id: e1a2b3c4
title: Implement notes update command
type: "task"
created_at: 2026-02-27T08:45:00+10:30
updated_at: 2026-02-27T23:59:00+10:30
status: completed
epic_id: a7c3d9e1
story_id: 1a2b3c4d
assigned_to: 3219612528192551
---

# Implement notes update command

## Objective
Add `jot notes update <path>` with alias `put`, script-safe output, and replace-only default behavior.

## Steps
1. Add command scaffold and alias in `cmd/`.
2. Implement stdin/file input handling and `--create` semantics.
3. Enforce root-bound path checks and deterministic output (`list|json`).
4. Add unit/integration tests for success and failure paths.

## Exit Criteria
- `notes update` works with alias `put`.
- Missing target fails without `--create`.
- JSON output and exit codes are machine-friendly.
- Tests cover traversal, missing file, create/replace behavior.

## Actual Outcome
- Added `jot notes update <path>` with alias `put` in `cmd/notes_update.go`.
- Implemented input handling for stdin and `--input <file>` with conflict detection.
- Enforced replace-only default behavior; missing target now fails unless `--create` is supplied.
- Added root-bound path resolution and traversal protection.
- Added deterministic machine output modes (`--format list|json`) for success and failure paths.
- Exit behavior is machine-usable: success returns exit code `0`, failures return non-zero via command error propagation.
- Added tests in `cmd/notes_update_test.go` covering traversal rejection, replace/create semantics, missing-target failure, input source handling, alias registration, and deterministic output shapes.

## Lessons Learned
- Path traversal checks are easiest to keep correct by resolving against notebook root then validating `filepath.Rel` prefix.
- Explicit source conflict validation (`stdin` vs `--input`) avoids non-deterministic CLI behavior in automation scripts.
- Emitting deterministic failure payloads before returning errors preserves script-friendly output while keeping non-zero exit signaling.

## Notes
- @TODO(epic:c5d7e9b1): group-aware enforcement deferred.
- @TODO(epic:b2f4e6a8): workflow policy hooks deferred.
