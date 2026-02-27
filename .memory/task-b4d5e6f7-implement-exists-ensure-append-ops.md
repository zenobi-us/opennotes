---
id: b4d5e6f7
title: Implement exists ensure append ops
type: "task"
created_at: 2026-02-27T08:45:00+10:30
updated_at: 2026-02-28T01:34:00+10:30
status: completed
epic_id: a7c3d9e1
story_id: 4d5e6f7a
assigned_to: 3219612528192551
---

# Implement exists ensure append ops

## Objective
Add script-safe utility operations for common agent loops: `exists`, `ensure`, and `append`.

## Steps
1. Add `notes exists <path>` with reliable boolean exit semantics.
2. Add `notes ensure <path>` with deterministic create-if-missing behavior.
3. Add `notes append <path>` for stdin/file append workflows.
4. Add tests for exit codes, idempotency, and root-bound checks.

## Exit Criteria
- Utility commands reduce shell-level file checks/creation logic.
- Exit codes are stable for automation.
- Path traversal outside notebook root is blocked.
- Tests cover success/failure automation paths.

## Actual Outcome
- Added `jot notes exists <path>` (`cmd/notes_exists.go`) with stable not-found semantics (`ExitCodeNotFound`) and deterministic `list|json` output.
- Added `jot notes ensure <path>` (`cmd/notes_ensure.go`) with deterministic create-if-missing behavior and idempotent second-run semantics.
- Added `jot notes append <path>` (`cmd/notes_append.go`) supporting stdin or `--input`, plus optional `--create` when target is missing.
- Reused root-bound path validation and `.md` normalization from shared path resolution logic.
- Added tests in `cmd/notes_ops_test.go` covering exit codes, idempotency, append behavior, and command flag wiring.

## Lessons Learned
- Small utility commands become reliable building blocks when they share path/format/error conventions with existing commands.
- Using typed exit codes for automation-critical paths makes shell orchestration deterministic.

## Notes
- @TODO(epic:c5d7e9b1): group-aware defaults for ensure/append deferred.
- @TODO(epic:b2f4e6a8): workflow-aware note lifecycle policies deferred.
