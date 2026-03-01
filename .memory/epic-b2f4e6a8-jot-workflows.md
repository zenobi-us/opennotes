---
id: b2f4e6a8
title: Jot Workflows
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T22:36:00+10:30
status: completed
---

# Jot Workflows

## Vision/Goal
Introduce native Jot workflows to express multi-step operational flows, metadata-based routing, and validation gates with machine-readable execution outcomes.

## Scope
- Define workflow model from real Jot usage patterns.
- Specify metadata/transition DSL contract for notebook-scoped workflows.
- Deliver execution + validation engine contract (dry-run/apply, diagnostics, status output).

## Success Criteria
- Workflow definition schema is documented with required fields and examples.
- Validation semantics (metadata requirements, transition constraints, failure handling) are explicit and testable.
- Execution contract includes: entrypoint, dry-run/apply behavior, machine-readable output shape, and failure diagnostics.
- Story-level outputs are linked and traceable from this epic.

## Phased Plan
1. **Discovery & Requirements** (`story-6f7a8b9c`)  
   Capture use cases, constraints, and non-goals; produce prioritized requirement baseline.
2. **Workflow Spec & Metadata DSL** (`story-7a8b9c0d`)  
   Define workflow definition format, transition model, and validation/error contracts.
3. **Execution & Validation Engine** (`story-8b9c0d1e`)  
   Implement evaluator behavior and machine-readable workflow status/diagnostics interface.

## Dependencies
- Metadata conventions across note groups/types.
- Alignment with current search/query and notebook configuration contracts.
- Decision on workflow definition storage/authoring location.

## Linked Stories
- [story-6f7a8b9c](story-6f7a8b9c-workflow-discovery-and-requirements.md) ✅
- [story-7a8b9c0d](story-7a8b9c0d-workflow-spec-metadata-dsl.md) ✅
- [story-8b9c0d1e](story-8b9c0d1e-workflow-execution-and-validation-engine.md) ✅
- [story-d1e9f2a3](story-d1e9f2a3-workflow-assignment-and-lifecycle-hook-enforcement.md) ✅

## Distilled Learnings
- [learning-w4k9f2m1](learning-w4k9f2m1-jot-workflows-epic-complete.md) — Epic-level learnings covering architecture decisions, patterns, gotchas, and reusable templates.
