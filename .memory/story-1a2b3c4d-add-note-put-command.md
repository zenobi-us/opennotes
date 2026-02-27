---
id: 1a2b3c4d
title: Add note put command
epic_id: a7c3d9e1
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T02:02:00+10:30
status: completed
priority: critical
story_points: 5
---

# Add note put command

## User Story
As an agent, I want to update or replace an existing note through Jot so that I do not need direct filesystem writes.

## Acceptance Criteria
- [x] `jot notes update <path>` exists with alias `jot notes put <path>` and supports stdin/file input.
- [x] Default behavior is replace-only; create-on-missing requires explicit `--create`.
- [x] Returns machine-usable success/failure output and exit code.
- [x] Root-bound path safety is enforced.

## Context
Current flow has create/delete but no direct update command.

## Out of Scope
Bulk batch editing.

## Use Stories
1. As an agent, I update `.memory/summary.md` each loop and want a typo like `summmary.md` to fail fast instead of creating a new file.
2. As an agent, I bootstrap a missing state file intentionally with `--create` during first-run setup.
3. As CI automation, I replace note content from generated artifacts using deterministic JSON/exit-code behavior.

## Tasks
- [task-e1a2b3c4](task-e1a2b3c4-implement-notes-update-command.md) — Implement notes update command

## Notes
- Command should be optimized for automation and agent loops.
- @TODO(epic:c5d7e9b1): Add group-aware validation/enforcement during update.
- @TODO(epic:b2f4e6a8): Add workflow-aware policy hooks for state transitions.
