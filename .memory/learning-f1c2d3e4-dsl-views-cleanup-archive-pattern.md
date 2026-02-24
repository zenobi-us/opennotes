---
id: f1c2d3e4
title: DSL views cleanup archive pattern
type: "learning"
created_at: 2026-02-22T19:28:46+10:30
updated_at: 2026-02-22T19:28:46+10:30
status: completed
tags:
  - lessons-learned
  - project-hygiene
  - dsl-views
---

# DSL views cleanup archive pattern

## Summary
Completed task files for phase `4adb81db` should be archived as soon as they are closed, while active/deferred tasks remain in root `.memory/` and tracked in `todo.md`.

## Details
- Archived completed follow-up tasks into `archive/cleanup-2026-02-22/`:
  - `task-6c834006`
  - `task-ddbaa84b`
  - `task-a82526f2`
  - `task-b66d2263` (duplicate completed task still present in root)
- Kept research and learning files unarchived.
- Left phase `4adb81db` unarchived because phase status is `planning` and open tasks remain (`task-684a9a73`, `task-e19963c7`).

## Implications
- Reduced stale completed-task noise in root `.memory/`.
- Preserved searchable completion history by grouping archived tasks in a dated cleanup folder.
- Future cleanup should follow the same rule: archive only completed task/phase artifacts, never archive learning/research.
