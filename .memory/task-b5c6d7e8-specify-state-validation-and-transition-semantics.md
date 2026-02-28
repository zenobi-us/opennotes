---
id: b5c6d7e8
title: Specify state validation and transition semantics
epic_id: b2f4e6a8
story_id: 7a8b9c0d
created_at: 2026-02-28T01:32:00+10:30
updated_at: 2026-02-28T02:08:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Specify state validation and transition semantics

## Objective
Define deterministic rules for metadata validation and allowed state changes per workflow step.

## Related Story
[story-7a8b9c0d](story-7a8b9c0d-workflow-spec-metadata-dsl.md)

## Steps
1. Define validation order and failure behavior.
2. Specify transition evaluation against per-state `transitions`.
3. Define semantics for self-transitions and custom states (e.g., cancelled).

## Expected Outcome
Unambiguous transition/validation behavior suitable for implementation and tests.

## Actual Outcome
Transition and validation semantics were finalized for V1:
- Transition legality check runs first.
- Invalid transition fails immediately.
- Target-state schema validation runs only when transition is valid.
- Self-transition is allowed only when explicitly listed.
- Custom states are allowed as long as they are declared in `states`.
- `initial_state` and all transition targets must resolve to existing `states` keys.

## Decision Log (Questions, Options, Answers)

### Q1: Transition validation order on state change request
**Options presented:**
- **A)** Validate transition first, then validate target-state schema
- **B)** Validate target-state schema first, then transition
- **C)** Run both and return combined errors

**Answer:**
- "yeah run both, but do A fail if needed then B"

**Recorded interpretation:**
- Transition check first, fail fast on invalid transition, then schema validation when transition is valid.

### Q2: Self-transition behavior (`planned -> planned`)
**Options presented:**
- **A)** Allow only if explicitly listed in `transitions`
- **B)** Always allow self-transition implicitly
- **C)** Disallow self-transition always

**Answer:**
- "a, yeah not sure why we're discussing this. if its in the workflow state transitions targets then :shrug: ... allow it"

**Recorded interpretation:**
- Self-transition allowed only when explicitly declared.

### Q3: Consolidated default semantic rule set
**Proposed defaults:**
- Any state key under `states` is valid (custom states allowed)
- `initial_state` must exist in `states`
- Every transition target must exist in `states`
- Transition check first; schema check second
- Self-transition only if explicitly listed

**Answer:**
- "yes"

## Lessons Learned
Capturing explicit transition semantics in the planning phase removes ambiguity before diagnostics and engine implementation work begins.
