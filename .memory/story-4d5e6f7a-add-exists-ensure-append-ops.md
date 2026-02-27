---
id: 4d5e6f7a
title: Add exists ensure append ops
epic_id: a7c3d9e1
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T01:34:00+10:30
status: completed
priority: high
story_points: 5
---

# Add exists ensure append ops

## User Story
As an agent, I want utility operations (`exists`, `ensure`, `append`) so that planning/state notes can be maintained without shell utilities.

## Acceptance Criteria
- [x] `jot notes exists <path>` provides reliable exit codes.
- [x] `jot notes ensure <path>` creates missing notes deterministically.
- [x] `jot notes append <path>` supports stdin append.
- [x] All operations are root-bounded and script-friendly.

## Context
Agent loops frequently check/create/update summary/todo/team notes.

## Out of Scope
Line-level structured patch syntax.

## Tasks
- [task-b4d5e6f7](task-b4d5e6f7-implement-exists-ensure-append-ops.md) — Implement exists ensure append ops

## Notes
Design should minimize command count for common agent flows.
