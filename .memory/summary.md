# Jot Project Summary

## Current Focus
- **Active epic**: [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md) (`status: proposed`, queued for activation)
- **Completed epic**: [epic-b2f4e6a8](epic-b2f4e6a8-jot-workflows.md) (`status: completed`, pending human review before archival)
- **NEW proposed epic**: [epic-7c631839](epic-7c631839-json-schema-generation.md) (`status: proposed`)

## Latest Milestone — Jot Workflows Epic Complete ✅
Epic `b2f4e6a8` delivered a full workflow enforcement system across 4 stories and 15 tasks:
1. **Discovery & Requirements** — use cases, constraints, non-goals, codebase integration map
2. **Workflow Spec & Metadata DSL** — `.jot.json` schema, transition semantics, diagnostics contract
3. **Execution & Validation Engine** — `EvaluateWorkflow()` with apply/dry-run, machine-readable output
4. **Assignment & Lifecycle Hooks** — group→workflow resolution, `notes add`/`notes update` enforcement

Key deliverables: `ExitCodeWorkflowBlocked` (exit 4), 7 stable `WF_*` error codes, 33+ new tests, zero regressions.
Learnings distilled: [learning-w4k9f2m1](learning-w4k9f2m1-jot-workflows-epic-complete.md)

## New Epic: JSON Schema Generation (`7c631839`)
Addresses tech debt from commit `3e8fcbf` where schema generation was planned but never implemented.
- **Goal**: Auto-generate `jot.schema.json` from Go structs to prevent drift
- **4 Stories**: Generation tooling, CI drift detection, init schema copy, validate command
- **Drivers**: Schema already drifted (missing workflow fields), manual maintenance is error-prone

## Next Steps
1. **Human review** of epic `b2f4e6a8` completion before archival.
2. **Prioritize** between `7c631839` (Schema Generation) and `c5d7e9b1` (Groups Verification).
3. Begin research phase for selected epic.

## Active Planning Epics
- [epic-7c631839](epic-7c631839-json-schema-generation.md) — JSON Schema Generation (proposed)
- [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md) — Jot Groups Verification Analysis (queued)
- [epic-6e1f2a9c](epic-6e1f2a9c-cli-config-normalization-layer.md) — CLI Config Normalization Layer (proposed/parked)

## Parked Work
- [task-9c4a2f8d](task-9c4a2f8d-github-actions-moonrepo-releases.md) — GitHub Actions CI/CD
- [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) — DSL views implementation
- [phase-4adb81db](phase-4adb81db-dsl-views-deferred-followups.md) — Deferred DSL views follow-up phase
