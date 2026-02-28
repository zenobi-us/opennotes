---
id: 9d1e2f3a
title: Workflow evaluation red tests and contract
epic_id: b2f4e6a8
story_id: 8b9c0d1e
created_at: 2026-02-28T18:07:18+10:30
updated_at: 2026-02-28T18:09:30+10:30
status: completed
assigned_to: "3219612528192551"
---

# Workflow evaluation red tests and contract

## Objective
Define and lock the workflow evaluation API/response contract through failing tests first.

## Related Story
[story-8b9c0d1e](story-8b9c0d1e-workflow-execution-and-validation-engine.md)

## Steps
1. Add tests for a service-level evaluation entrypoint and output shape.
2. Add tests for apply-mode hard enforcement on invalid transitions/metadata.
3. Add tests for dry-run behavior preserving current state while reporting validation.
4. Add test confirming no warn-only apply semantics in this slice.

## Expected Outcome
New tests fail against current code and describe the exact behavior required for implementation.

## Actual Outcome
Added failing tests first for workflow evaluation contract and behavior:
- apply-mode success path with machine-readable state fields
- apply-mode hard-block for invalid transitions
- apply-mode hard-block for missing metadata
- dry-run behavior (valid + allowed but not applied)
- enforce behavior maintained even when workflow config mode is `warn`

Verified RED state via `mise run test` with undefined symbol failures before implementation.

## Lessons Learned
Contract-first tests made the execution output shape explicit and prevented accidental overreach during implementation.
