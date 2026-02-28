---
id: d4e5f6a7
title: Workflow DSL schema contract v1
epic_id: b2f4e6a8
related_task_id: a4b5c6d7
created_at: 2026-02-28T01:40:00+10:30
updated_at: 2026-02-28T01:40:00+10:30
status: in-progress
---

# Workflow DSL schema contract v1

## Research Questions
1. What is the canonical `.jot.json` structure for workflows?
2. What key constraints and naming rules apply?
3. What minimum fields are required per state definition?

## Summary
Initial DSL contract decisions are being captured for `story-7a8b9c0d`. Workflows remain notebook-scoped in `.jot.json` using keyed objects. Workflow key naming is explicitly free-form for V1.

## Findings
- Workflows are stored in `.jot.json` as keyed objects.
- Workflow keys are **free-form** in V1 (no kebab/snake restriction).
- Existing migration framework covers the added workflow schema.
- Canonical state container is `states` map:
  - `states.<stateName>.schema` (full JSON Schema)
  - `states.<stateName>.transitions` (allowed next states)
- Required top-level workflow fields (V1 minimum):
  - `description`
  - `initial_state`
  - `mode` (`enforce|warn`)
  - `states`
- Unknown fields policy: strict reject (fail fast).

## Open Decisions
- None (schema shape decisions for V1 are closed).

## References
- [research-a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md)
- Interactive planning session with project owner (Q), 2026-02-28.
