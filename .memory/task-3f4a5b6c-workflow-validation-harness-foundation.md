---
id: 3f4a5b6c
title: Workflow validation harness foundation
epic_id: b2f4e6a8
story_id: 6f7a8b9c
created_at: 2026-02-28T17:42:00+10:30
updated_at: 2026-02-28T18:00:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Workflow validation harness foundation

## Objective
Create the initial test harness and fixture strategy for workflow state/schema validation behavior.

## Related Story
[story-6f7a8b9c](story-6f7a8b9c-workflow-discovery-and-requirements.md)

## Steps
1. Define fixture layout for workflow definitions and sample note metadata.
2. Add RED tests for valid transition, invalid transition, and missing required metadata.
3. Define diagnostics assertion helpers for stable machine-readable output.
4. Add TODO marker for apply-mode semantics deferred to `story-8b9c0d1e`.

## Expected Outcome
A repeatable test harness that gates workflow validation behavior before engine implementation.

## Actual Outcome
Completed with a concrete harness in `internal/services/workflow_validation_test.go` and fixture-backed workflow definitions in `internal/services/testdata/workflows/project_story.json`.

Delivered:
- RED tests for valid transition, invalid transition, and missing required metadata.
- Reusable diagnostics assertion helper for machine-readable diagnostic code checks.
- Deferred apply-mode semantics explicitly captured via skipped TODO test tied to `story-8b9c0d1e`.

## Lessons Learned
A minimal transition validator and fixture-first tests establish stable diagnostic contracts early, reducing risk before execution engine wiring.
