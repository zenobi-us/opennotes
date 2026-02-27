---
id: a3c4d5e6
title: Implement notes get raw mode
type: "task"
created_at: 2026-02-27T08:45:00+10:30
updated_at: 2026-02-28T01:26:00+10:30
status: completed
epic_id: a7c3d9e1
story_id: 3c4d5e6f
assigned_to: 3219612528192551
---

# Implement notes get raw mode

## Objective
Extend `jot notes get` with `--raw` to emit exact file bytes for deterministic automation roundtrips.

## Steps
1. Add `--raw` flag and enforce incompatibility with `--format json`.
2. Implement raw output path that writes exact bytes to stdout.
3. Keep existing parsed/list/json behavior unchanged when `--raw` is absent.
4. Add tests for byte fidelity and flag validation.

## Exit Criteria
- `notes get --raw` outputs exact file bytes.
- `--raw` + `--format json` fails with clear error.
- Existing `notes get` outputs remain backward compatible.
- Tests verify raw roundtrip behavior.

## Actual Outcome
- Extended `jot notes get <path>` with `--raw` support in `cmd/notes_get.go`.
- Added validation to reject incompatible `--raw --format json` with a clear error.
- Added raw read path (`loadRawNoteByPath`) that emits exact file bytes to stdout.
- Kept existing parsed behavior unchanged for non-raw `list|json` output paths.
- Added tests in `cmd/notes_get_test.go` covering `--raw` flag registration, raw+json validation, exact-byte raw reads, and existing traversal protection.

## Lessons Learned
- Separating parsed read logic from raw byte reads avoids accidental normalization.
- Flag compatibility checks are best centralized to keep CLI behavior deterministic.

## Notes
- @TODO(epic:c5d7e9b1): evaluate optional group-context helpers for read flows.
