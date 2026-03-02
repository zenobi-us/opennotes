# Jot Project Summary

## Current Focus
- **Active epic**: [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md) (`status: in-progress`, 3 phases planned)
- **Completed epic**: [epic-b2f4e6a8](epic-b2f4e6a8-jot-workflows.md) (`status: completed`, pending human review before archival)
- **Proposed epic**: [epic-7c631839](epic-7c631839-json-schema-generation.md) (`status: proposed`)

## Latest Milestone — Jot Workflows Epic Complete ✅
Epic `b2f4e6a8` delivered a full workflow enforcement system across 4 stories and 15 tasks:
1. **Discovery & Requirements** — use cases, constraints, non-goals, codebase integration map
2. **Workflow Spec & Metadata DSL** — `.jot.json` schema, transition semantics, diagnostics contract
3. **Execution & Validation Engine** — `EvaluateWorkflow()` with apply/dry-run, machine-readable output
4. **Assignment & Lifecycle Hooks** — group→workflow resolution, `notes add`/`notes update` enforcement

Key deliverables: `ExitCodeWorkflowBlocked` (exit 4), 7 stable `WF_*` error codes, 33+ new tests, zero regressions.
Learnings distilled: [learning-w4k9f2m1](learning-w4k9f2m1-jot-workflows-epic-complete.md)

## Active Epic: Jot Groups Verification Analysis (`c5d7e9b1`)

**Goal**: Verify groups can handle intent-level commands (e.g., "create a task titled X") with automatic naming/path/metadata conventions.

### Phase Structure (3 phases, inline in epic)

| Phase | Focus | Stories | Status |
|-------|-------|---------|--------|
| Phase 1: Foundation | Slugify, type flag | g2b3c4d5, f1a2b3c4 | planned |
| Phase 2: Schema & Resolution | Filename patterns, intent mapping | h3c4d5e6, 9c0d1e2a, ad1e2f3b | planned |
| Phase 3: User Experience | Templates, interactive selection | i4d5e6f7, j5e6f7a8 | planned |

### Stories (7 total, phase-agnostic)
- 2 Foundation stories (P0)
- 3 Schema & Resolution stories (P1)  
- 2 User Experience stories (P2)

## New Epic: JSON Schema Generation (`7c631839`)
Addresses tech debt from commit `3e8fcbf` where schema generation was planned but never implemented.
- **Goal**: Auto-generate `jot.schema.json` from Go structs to prevent drift
- **5 Stories**: Type separation, generation tooling, CI drift detection, init schema copy, validate command
- **Drivers**: Schema already drifted (missing workflow fields), manual maintenance is error-prone

## Next Steps
1. **Human review** of epic `b2f4e6a8` completion before archival.
2. Begin task breakdown for epic `c5d7e9b1` Phase 1 (Foundation).
3. Continue research if needed: [research-r1a2b3c4](research-r1a2b3c4-jot-groups-intent-resolution.md)

## Active Planning Epics
- [epic-c5d7e9b1](epic-c5d7e9b1-jot-groups-verification-analysis.md) — Jot Groups Verification Analysis (active, refactored)
- [epic-7c631839](epic-7c631839-json-schema-generation.md) — JSON Schema Generation (proposed)
- [epic-6e1f2a9c](epic-6e1f2a9c-cli-config-normalization-layer.md) — CLI Config Normalization Layer (proposed/parked)

## Parked Work
- [task-9c4a2f8d](task-9c4a2f8d-github-actions-moonrepo-releases.md) — GitHub Actions CI/CD
- [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) — DSL views implementation
- [phase-4adb81db](phase-4adb81db-dsl-views-deferred-followups.md) — Deferred DSL views follow-up phase

## Memory Structure Note
Epic `c5d7e9b1` refactored to new miniproject model (2026-03-02):
- Phases are INLINE sections in epic files (not separate files)
- Stories are phase-agnostic (define WHAT, linked to epic only)
- Tasks link to both story (WHAT) and phase (WHEN)
