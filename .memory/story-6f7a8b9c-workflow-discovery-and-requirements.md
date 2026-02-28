---
id: 6f7a8b9c
title: Workflow discovery and requirements
epic_id: b2f4e6a8
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T18:00:00+10:30
status: completed
priority: critical
story_points: 3
---

# Workflow discovery and requirements

## User Story
As a product team, we want a structured discovery discussion for Jot workflows so that design choices match real use cases.

## Acceptance Criteria
- [x] Discovery session scope and questions are documented. (see [research-a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md))
- [x] Primary workflow use cases are captured and prioritized. (see [research-a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md))
- [x] Non-goals and constraints are explicitly recorded. (see [research-a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md))
- [x] Codebase integration points for workflow implementation are documented at command/service boundaries. (see [research-5a6b7c8d](research-5a6b7c8d-workflow-integration-points.md))
- [x] Validation harness scope and red-test matrix are defined and linked from task artifacts. (see [task-3f4a5b6c](task-3f4a5b6c-workflow-validation-harness-foundation.md))

## Context
You requested first task to be discussion + design framing.

## Out of Scope
Implementation in this story.

## Deliverables
- Discovery brief with scope, assumptions, and question set.
- Prioritized workflow use-case list (P0/P1/P2).
- Constraint/non-goal register to bound downstream DSL and engine work.

## Tasks
- [task-d1e2f3a4](task-d1e2f3a4-workflow-discovery-brief-and-question-set.md)
- [task-e2f3a4b5](task-e2f3a4b5-capture-and-prioritize-workflow-use-cases.md)
- [task-f3a4b5c6](task-f3a4b5c6-record-workflow-constraints-and-non-goals.md)
- [task-1d2e3f4a](task-1d2e3f4a-workflow-codebase-discovery-and-ownership-map.md)
- [task-2e3f4a5b](task-2e3f4a5b-workflow-config-persistence-and-introspection-slice.md)
- [task-3f4a5b6c](task-3f4a5b6c-workflow-validation-harness-foundation.md)

## Notes
This story is the gate before technical design for stories `7a8b9c0d` and `8b9c0d1e`.
