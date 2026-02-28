---
id: e2f9a3b4
title: Update config schema for group workflow_id and workflow field
epic_id: b2f4e6a8
story_id: d1e9f2a3
created_at: 2026-02-28T20:33:26+10:30
updated_at: 2026-02-28T20:57:40+10:30
status: completed
assigned_to: "3219612528192551"
---

# Update config schema for group workflow_id and workflow field

## Objective
Refactor notebook config model to use `groups[].workflow_id` and `workflows.<name>.field`, removing deprecated group workflow object shape.

## Related Story
[story-d1e9f2a3](story-d1e9f2a3-workflow-assignment-and-lifecycle-hook-enforcement.md)

## Steps
1. Update notebook/group/workflow service structs to new shape.
2. Add parser/serialization tests for new fields.
3. Add backward-compat migration/normalization tests for legacy workflow object keys.
4. Ensure save/load preserve new shape.

## Expected Outcome
Config schema supports group assignment by `workflow_id` and workflow-level `field` definition with test coverage.

## Actual Outcome
Implemented and verified config schema changes:
- Added `groups[].workflow_id` support in notebook config model.
- Added legacy compatibility normalization from `groups[].workflow.id` to `workflow_id`.
- Added workflow definition `field` support in service contract and fixtures.
- Save path now emits canonical `workflow_id` (no legacy `workflow` object).

Added tests in `internal/services/notebook_test.go` and `internal/services/workflow_validation_test.go` and verified with `mise run test`.

## Lessons Learned
Custom JSON unmarshal on `NotebookGroup` is the safest bridge for legacy compatibility without polluting canonical write format.
