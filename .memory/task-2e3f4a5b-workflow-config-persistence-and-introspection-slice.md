---
id: 2e3f4a5b
title: Workflow config persistence and introspection slice
epic_id: b2f4e6a8
story_id: 6f7a8b9c
created_at: 2026-02-28T17:42:00+10:30
updated_at: 2026-02-28T17:42:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Workflow config persistence and introspection slice

## Objective
Implement a minimal vertical slice that preserves workflow config and exposes it for inspection.

## Related Story
[story-6f7a8b9c](story-6f7a8b9c-workflow-discovery-and-requirements.md)

## Steps
1. Add failing tests for workflow preservation in notebook config load/save.
2. Implement minimal config model changes to preserve `workflows` in `.jot.json`.
3. Expose workflows in notebook info JSON payload.
4. Run full test suite.

## Expected Outcome
Workflow definitions are not dropped when notebook config is saved; workflow metadata is visible via notebook info output.

## Actual Outcome
Completed. Tests added and passing for config preservation and payload exposure.

## Lessons Learned
Small persistence-first slices reduce rollout risk for later workflow execution engine work.
