---
id: 7a8b9c0d
title: Workflow spec metadata dsl
epic_id: b2f4e6a8
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T13:12:00+10:30
status: completed
priority: high
story_points: 5
---

# Workflow spec metadata dsl

## User Story
As a notebook owner, I want workflow definitions that express metadata rules and transitions so that note state is enforceable.

## Acceptance Criteria
- [x] Workflow definition format is specified. (see [research-d4e5f6a7](research-d4e5f6a7-workflow-dsl-schema-contract-v1.md))
- [x] Metadata validation rules and transition model are defined. (see [task-b5c6d7e8](task-b5c6d7e8-specify-state-validation-and-transition-semantics.md))
- [x] Error/reporting contract is defined for automation. (see [task-c6d7e8f9](task-c6d7e8f9-define-format-aware-diagnostics-contract.md))

## Context
Workflows should supersede the simplistic miniproject flow chart.

## Out of Scope
Execution engine implementation.

## Deliverables
- `.jot.json` workflow object schema contract (keyed objects, step layout, JSON Schema embedding).
- Transition + validation semantics specification (state-scoped schema + allowed transitions).
- Format-aware diagnostics contract aligned with `--format` output behavior.

## Tasks
- [task-a4b5c6d7](task-a4b5c6d7-define-workflow-object-schema-in-jot-json.md)
- [task-b5c6d7e8](task-b5c6d7e8-specify-state-validation-and-transition-semantics.md)
- [task-c6d7e8f9](task-c6d7e8f9-define-format-aware-diagnostics-contract.md)

## Notes
Prefer explicit, testable contracts over implicit behavior.
Reuse existing config migration framework for new workflow schema fields.
