---
id: a4b5c6d8
title: Wire lifecycle hooks into note create/edit with diagnostics
epic_id: b2f4e6a8
story_id: d1e9f2a3
created_at: 2026-02-28T20:33:26+10:30
updated_at: 2026-02-28T20:57:40+10:30
status: in-progress
assigned_to: "3219612528192551"
---

# Wire lifecycle hooks into note create/edit with diagnostics

## Objective
Integrate workflow evaluation into note create/edit lifecycle so validation is automatic and hard-enforced.

## Related Story
[story-d1e9f2a3](story-d1e9f2a3-workflow-assignment-and-lifecycle-hook-enforcement.md)

## Steps
1. Add RED tests in command/service flows for create/edit lifecycle enforcement.
2. Resolve workflow assignment and metadata field state (`workflows.<id>.field`) at mutation points.
3. Invoke workflow evaluator and block invalid apply path.
4. Ensure diagnostics surfaced in CLI outputs (`list/json`) with stable codes.
5. Add e2e coverage for create/edit success/failure and no-workflow cases.

## Expected Outcome
Create/edit operations enforce workflow transitions automatically with deterministic diagnostic behavior.

## Actual Outcome
TBD

## Lessons Learned
TBD
