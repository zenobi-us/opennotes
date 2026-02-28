---
id: f3a4b5c7
title: Define and implement workflow assignment resolution
epic_id: b2f4e6a8
story_id: d1e9f2a3
created_at: 2026-02-28T20:33:26+10:30
updated_at: 2026-02-28T20:57:40+10:30
status: completed
assigned_to: "3219612528192551"
---

# Define and implement workflow assignment resolution

## Objective
Provide deterministic workflow selection for a note based on matching groups and `workflow_id` assignment.

## Related Story
[story-d1e9f2a3](story-d1e9f2a3-workflow-assignment-and-lifecycle-hook-enforcement.md)

## Steps
1. Define conflict policy for multiple matched groups with differing workflow IDs.
2. Implement group matching + workflow assignment resolver in service layer.
3. Return machine-readable diagnostics for no-assignment, unknown-workflow, and conflict cases.
4. Add unit tests for single-match, multi-match-same-ID, and conflict scenarios.

## Expected Outcome
Assignment resolver is deterministic, tested, and produces automation-safe diagnostics.

## Actual Outcome
Implemented deterministic assignment resolver in `internal/services/workflow_assignment.go` with machine-readable diagnostics for:
- no assignment (`workflow.assignment_not_found`)
- conflicting group assignments (`workflow.assignment_conflict`)
- unknown workflow reference (`workflow.assignment_unknown_workflow`)
- invalid workflow payload (`workflow.invalid_definition`)
- missing workflow field selector (`workflow.missing_field_selector`)

Added resolver tests in `internal/services/workflow_assignment_test.go` for single-match, same-ID multi-match, conflict, no-assignment, and unknown-workflow scenarios.

## Lessons Learned
Fail-fast assignment diagnostics make lifecycle enforcement predictable and easier to integrate into CLI error surfaces.
