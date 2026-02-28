---
id: ae2f3a4b
title: Apply-mode enforce implementation and validation
epic_id: b2f4e6a8
story_id: 8b9c0d1e
created_at: 2026-02-28T18:07:18+10:30
updated_at: 2026-02-28T18:09:30+10:30
status: completed
assigned_to: "3219612528192551"
---

# Apply-mode enforce implementation and validation

## Objective
Implement minimal workflow evaluation logic to satisfy red tests, enforce apply-mode validity, and provide stable machine-readable output.

## Related Story
[story-8b9c0d1e](story-8b9c0d1e-workflow-execution-and-validation-engine.md)

## Steps
1. Implement evaluation request/result structures in service layer.
2. Wire apply and dry-run paths to existing transition validator.
3. Ensure actionable diagnostic output and stable status fields.
4. Run full test suite and update memory artifacts.

## Expected Outcome
Green tests, stable output contract, and documented completion in memory artifacts.

## Actual Outcome
Implemented `EvaluateWorkflow(...)` plus request/result contracts in `internal/services/workflow_validation.go` with stable machine-readable fields:
`valid`, `allowed`, `applied`, `mode`, `from_state`, `to_state`, `result_state`, `diagnostics`.

Behavior delivered:
- Apply mode validates transition/metadata and blocks invalid requests.
- Dry-run validates without applying state mutation.
- Config `mode: warn` still enforces apply-mode gating in this first slice.
- Invalid execution mode returns diagnostic code `workflow.invalid_execution_mode`.

Full suite verified green with `mise run test`.

## Lessons Learned
Reusing the transition validator baseline kept the first execution slice minimal and deterministic.
