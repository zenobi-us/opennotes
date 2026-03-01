---
id: a4b5c6d8
title: Wire lifecycle hooks into note create/edit with diagnostics
epic_id: b2f4e6a8
story_id: d1e9f2a3
created_at: 2026-02-28T20:33:26+10:30
updated_at: 2026-02-28T21:51:34+10:30
status: completed
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
Implemented lifecycle enforcement across three layers:

### Service layer (`internal/services/workflow_lifecycle.go`)
- `EnforceWorkflowOnMutation()` — resolves workflow assignment, determines from/to state, evaluates transition.
- Handles create (initial_state → target) and edit (existing → new) semantics.
- Self-transitions and no-state-change edits pass without evaluation.
- Assignment errors (conflict, unknown workflow) block mutations.
- No-match paths pass through with `enforced=false`.
- Produces `workflow.lifecycle_blocked` diagnostic on failure.

### CMD layer (`cmd/workflow_enforce.go`)
- `enforceWorkflowForCreate()` / `enforceWorkflowForEdit()` — thin wrappers that resolve relative paths and call service.
- `extractFrontmatterMetadata()` — parses YAML frontmatter from raw note bytes.
- `formatDiagnostics()` — human-readable diagnostic string.
- Returns `ExitCodeWorkflowBlocked` (exit code 4) on enforcement failure.

### Integration points
- `cmd/notes_add.go` — calls `enforceWorkflowForCreate()` before file write.
- `cmd/notes_update.go` — `updateNoteFileWithWorkflow()` calls `enforceWorkflowForCreate()` or `enforceWorkflowForEdit()` based on create/exists.

### Test coverage (23 new tests)
- 10 service-layer tests (`workflow_lifecycle_test.go`): create/edit/no-match/conflict/missing-metadata.
- 13 cmd-layer tests (`workflow_enforce_test.go`): end-to-end create/edit/block/allow/no-workflow/frontmatter-parsing.
- Full suite passes with zero regressions.

## Lessons Learned
- Keeping the lifecycle enforcement as a pure function with no I/O (service layer) makes it trivially testable.
- The cmd-layer wrapper only needs to handle path resolution and frontmatter extraction from raw bytes.
- Self-transition on create with initial_state is the key heuristic for allowing notes to be created with the default state.
