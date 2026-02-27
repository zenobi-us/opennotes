---
id: 2b3c4d5e
title: Add note move and archive operations
epic_id: a7c3d9e1
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T01:26:00+10:30
status: completed
priority: critical
story_points: 3
---

# Add note move and archive operations

## User Story
As an agent, I want to move notes between active and archive locations using Jot so that archival workflows stay inside Jot.

## Acceptance Criteria
- [x] `jot notes mv <from> <to>` exists.
- [x] Move operation preserves note content and metadata.
- [x] Root boundary checks block path escape.
- [x] Default is no-overwrite; destination overwrite requires explicit `--force`.

## Context
Miniproject-style archival needs safe move semantics.

## Out of Scope
Cross-notebook copy/move in first iteration.

## Use Stories
1. As an agent, I archive completed tasks from `tasks/` to `archive/tasks/` without leaving Jot.
2. As concurrent automation, I want destination collisions to fail by default so a run never silently overwrites another run's artifact.
3. As a maintenance operator, I intentionally replace a stale archive entry using `--force`.

## Tasks
- [task-f2b3c4d5](task-f2b3c4d5-implement-notes-mv-command.md) — Implement notes mv command

## Notes
- Include clear errors for missing source and destination conflicts.
- @TODO(epic:b2f4e6a8): Add workflow-state policy checks (for example active->archive transition rules).
