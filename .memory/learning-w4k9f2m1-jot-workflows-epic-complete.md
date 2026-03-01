---
id: w4k9f2m1
title: Jot Workflows epic complete
created_at: 2026-02-28T22:36:00+10:30
updated_at: 2026-02-28T22:36:00+10:30
status: completed
tags:
  - epic-learnings
  - workflows
  - architecture
  - best-practices
  - lessons-learned
---

# Jot Workflows Epic Complete — Distilled Learnings

## Summary

Epic `b2f4e6a8` (Jot Workflows) delivered a complete workflow enforcement system across 4 stories, 15 tasks, and 3 research artifacts. The system spans config schema, assignment resolution, validation engine, and lifecycle hooks — all with machine-readable diagnostics.

## Key Architectural Decisions

### 1. Workflow storage in `.jot.json` (keyed objects)
- Workflows live alongside notebook config as `workflows.<name>` keyed objects.
- No separate file (`workflows.yaml`) or notes-as-config. This keeps discovery trivial and leverages the existing config migration framework.
- Free-form workflow key names in V1 (no naming constraints).

### 2. Group-to-workflow assignment via `workflow_id`
- Groups reference workflows by `workflow_id: string`, not by embedding workflow definitions.
- This separation means workflows are reusable across groups and independently testable.
- Conflict policy: fail-fast when multiple matching groups resolve to different `workflow_id` values.

### 3. Three-layer architecture
- **Service layer** (`workflow_validation.go`, `workflow_assignment.go`, `workflow_lifecycle.go`): Pure functions, no I/O, fully unit-testable.
- **CMD layer** (`workflow_enforce.go`): Thin wrappers for path resolution and frontmatter extraction.
- **Integration points** (`notes_add.go`, `notes_update.go`): Single enforcement call before mutation.

### 4. Hard-enforce-only in V1
- `mode: warn` is accepted in config but treated as enforce for apply-mode gating.
- This eliminates a class of "warn-only" bugs and simplifies the first implementation slice.
- Warn mode deferred to future iteration with explicit acceptance criteria.

### 5. Canonical diagnostics payload
- Single payload schema for all output formats (JSON, text, table).
- Stable error codes (`WF_TRANSITION_NOT_ALLOWED`, `WF_SCHEMA_VALIDATION_FAILED`, etc.).
- Table output uses `details_json` column for lossless rendering of structured remainder.
- This prevents format-specific drift and makes automation reliable.

### 6. Exit code 4 for workflow enforcement failures
- Dedicated `ExitCodeWorkflowBlocked` separates workflow failures from general errors.
- Enables automation agents to distinguish "blocked by workflow" from "crashed".

## Implementation Patterns That Worked Well

### Contract-first TDD
- Every story started with RED tests defining the exact output contract.
- Tests were written against undefined symbols, verified to fail, then implementation filled gaps.
- This prevented accidental overreach and kept scope disciplined.

### Persistence-first vertical slices
- Before building the validation engine, a minimal slice ensured `.jot.json` preserves unknown `workflows` fields through load/save.
- This eliminated a data-loss risk that would have silently destroyed workflow definitions.

### Fixture-backed test harness
- Workflow definitions stored as JSON fixtures in `testdata/workflows/`.
- Reusable diagnostics assertion helpers check for stable `code` values.
- Fixture-first approach made test evolution across stories frictionless.

### Layered enforcement (pure function + thin wrapper)
- `EnforceWorkflowOnMutation()` is a pure function: input metadata, output decision + diagnostics.
- CMD layer only handles path resolution and frontmatter byte parsing.
- This split made both layers independently testable and kept the cmd thin.

### Existing framework reuse
- Config migration uses the existing `internal/services/migration.go` versioned framework.
- Legacy `groups[].workflow.id` normalized to `workflow_id` via custom JSON unmarshal, not migration step.
- Transition validator baseline was reused for execution engine without duplication.

## Gotchas and Edge Cases

### Self-transition on create
- When creating a note, the "from" state is the workflow's `initial_state` and the "to" state is also `initial_state`.
- This self-transition must be implicitly allowed for create to work, even when self-transitions are not explicitly listed.
- Key heuristic: `initial_state → initial_state` is always allowed on create.

### No-state-change edits
- Editing a note without changing the state field should pass without workflow evaluation.
- If from_state == to_state and it's not a create operation, skip evaluation entirely.

### Frontmatter extraction from raw bytes
- CMD layer must parse YAML frontmatter from raw note bytes before the note is written.
- This is the only I/O-adjacent operation in the enforcement path.

### Config field silent dropping
- Before the persistence slice, `NotebookService.SaveConfig()` would drop unknown keys.
- Adding `workflows` to `StoredNotebookConfig` was a prerequisite for everything else.
- **Lesson**: Always verify persistence round-trip before building features that depend on persisted data.

### Multiple group matches
- A note can match multiple groups. If they all reference the same `workflow_id`, proceed normally.
- If they reference different `workflow_id` values, fail with `workflow.assignment_conflict`.
- No priority/precedence system — conflict is an error, not a resolution.

## Reusable Patterns for Future Epics

### 1. Discovery → Schema → Engine → Integration pipeline
- This four-story progression worked well for any "new subsystem" epic:
  1. Discovery & requirements (scope, non-goals, use cases)
  2. Schema/contract definition (data shapes, error codes)
  3. Engine/logic implementation (pure functions, TDD)
  4. Integration/lifecycle hooks (wiring into existing flows)

### 2. Decision log in task artifacts
- Recording verbatim Q&A (question, options, answer, interpretation) in task files creates an audit trail.
- Prevents revisiting settled decisions and gives future sessions exact rationale.

### 3. Diagnostics-first design
- Define error codes and payload shapes before implementation.
- Every failure path should be actionable and machine-parseable.
- Table output should never lose information (use serialized remainder columns).

### 4. Thin command / fat service
- Commands should never contain business logic, only orchestration.
- This pattern scales well and keeps tests fast (no CLI bootstrapping for logic tests).

### 5. Legacy compatibility via custom unmarshal
- When evolving config schemas, implement backward compatibility in the deserialization layer.
- Write path always emits canonical format — no legacy shapes on output.
- This prevents schema drift without requiring migration steps for every field rename.

## Metrics

| Metric | Value |
|--------|-------|
| Stories completed | 4 |
| Tasks completed | 15 |
| Research artifacts | 3 |
| New test files | 3 (`workflow_validation_test.go`, `workflow_assignment_test.go`, `workflow_lifecycle_test.go`, `workflow_enforce_test.go`) |
| New tests added | ~33+ (10 service, 13 cmd, 10+ validation) |
| Exit codes added | 1 (`ExitCodeWorkflowBlocked = 4`) |
| Error codes defined | 7 (`WF_*` canonical codes) |
| Config schema fields added | `groups[].workflow_id`, `workflows.<name>.field` |
| Regressions | 0 |

## Links

- **Epic**: [epic-b2f4e6a8](epic-b2f4e6a8-jot-workflows.md)
- **Stories**: [6f7a8b9c](story-6f7a8b9c-workflow-discovery-and-requirements.md), [7a8b9c0d](story-7a8b9c0d-workflow-spec-metadata-dsl.md), [8b9c0d1e](story-8b9c0d1e-workflow-execution-and-validation-engine.md), [d1e9f2a3](story-d1e9f2a3-workflow-assignment-and-lifecycle-hook-enforcement.md)
- **Research**: [a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md), [5a6b7c8d](research-5a6b7c8d-workflow-integration-points.md), [d4e5f6a7](research-d4e5f6a7-workflow-dsl-schema-contract-v1.md)
