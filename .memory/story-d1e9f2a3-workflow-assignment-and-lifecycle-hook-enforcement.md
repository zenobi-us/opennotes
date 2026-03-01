---
id: d1e9f2a3
title: Workflow assignment and lifecycle hook enforcement
epic_id: b2f4e6a8
created_at: 2026-02-28T20:33:26+10:30
updated_at: 2026-02-28T21:51:34+10:30
status: completed
priority: high
story_points: 8
---

# Workflow assignment and lifecycle hook enforcement

## User Story
As a notebook owner, I want workflows assigned by group and enforced during note create/edit so workflow compliance happens automatically.

## Acceptance Criteria
- [x] `.jot.json` group schema supports `workflow_id: string` and removes group-level workflow object shape.
- [x] `.jot.json` workflow schema requires `workflows.<name>.field` as the metadata field used for state evaluation.
- [x] Migration/normalization behavior is defined for legacy configs using group workflow object (`workflow.id`, `workflow.field`, `on_create`, `on_edit`).
- [x] Note create/edit lifecycle invokes workflow validation automatically when a matched group has `workflow_id`.
- [x] Lifecycle enforcement uses `workflows[workflow_id].field` to read current/target state from note metadata and applies hard-enforce semantics.
- [x] Invalid workflow transitions/metadata block note mutation with machine-readable diagnostics.
- [x] Conflict behavior is deterministic when multiple matching groups reference different workflow IDs.

## Context
Story `8b9c0d1e` delivered service-level evaluation primitives. This story wires assignment model + lifecycle integration.

## Out of Scope
- Warn-mode semantics
- Distributed orchestration or async workflow execution engines

## Tasks
- [task-e2f9a3b4](task-e2f9a3b4-update-config-schema-for-group-workflow-id-and-workflow-field.md) ✅
- [task-f3a4b5c7](task-f3a4b5c7-define-and-implement-workflow-assignment-resolution.md) ✅
- [task-a4b5c6d8](task-a4b5c6d8-wire-lifecycle-hooks-into-note-create-edit-with-diagnostics.md) ✅

## Notes
- Conflict policy: fail fast if multiple matched groups resolve to different `workflow_id` values.
- Self-transition on create (initial_state → initial_state) always allowed.
- No-state-change edits pass without evaluation.
- Exit code 4 (`ExitCodeWorkflowBlocked`) for workflow enforcement failures.
