---
id: 8b9c0d1e
title: Workflow execution and validation engine
epic_id: b2f4e6a8
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T18:09:30+10:30
status: completed
priority: high
story_points: 8
---

# Workflow execution and validation engine

## User Story
As an automation agent, I want Jot to execute workflow steps and validations so that process compliance is machine-checkable.

## Acceptance Criteria
- [x] Service-layer workflow evaluation entrypoint exists and accepts explicit execution mode (`dry-run` or `apply`).
- [x] In `apply` mode, invalid transitions and missing required metadata are hard-blocked (no warn-only fallback), reusing baseline transition validator diagnostics.
- [x] Validation failures return actionable, machine-readable diagnostics with stable `code` values and paths suitable for automation.
- [x] Evaluation result includes stable machine-readable state/status fields (`valid`, `allowed`, `applied`, `from_state`, `to_state`, `result_state`, `diagnostics`).
- [x] Initial slice does **not** implement warn mode semantics; any configured `mode: warn` is treated as enforce for apply-mode gating.

## Context
Enables expressive process orchestration in Jot-native workflows.

## Out of Scope
Advanced distributed orchestration.

## Tasks
- [task-8c0d1e2f](task-8c0d1e2f-story8-first-slice-acceptance-and-task-breakdown.md)
- [task-9d1e2f3a](task-9d1e2f3a-workflow-evaluation-red-tests-and-contract.md)
- [task-ae2f3a4b](task-ae2f3a4b-apply-mode-enforce-implementation-and-validation.md)

## Notes
Should integrate cleanly with note metadata and group semantics.
