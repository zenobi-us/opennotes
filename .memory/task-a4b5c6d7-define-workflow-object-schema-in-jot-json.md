---
id: a4b5c6d7
title: Define workflow object schema in jot json
epic_id: b2f4e6a8
story_id: 7a8b9c0d
created_at: 2026-02-28T01:32:00+10:30
updated_at: 2026-02-28T01:49:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Define workflow object schema in jot json

## Objective
Specify the canonical on-disk workflow definition shape in `.jot.json` as keyed objects.

## Related Story
[story-7a8b9c0d](story-7a8b9c0d-workflow-spec-metadata-dsl.md)

## Steps
1. Define top-level workflow object layout and required keys.
2. Define per-state structure in `states` map (`schema`, `transitions`).
3. Specify constraints needed for migration and validation.

## Expected Outcome
A precise schema contract for workflow storage in `.jot.json`.

## Actual Outcome
Schema decisions captured in [research-d4e5f6a7](research-d4e5f6a7-workflow-dsl-schema-contract-v1.md), including keyed-object storage, free-form workflow keys, `states` map shape, required top-level fields (`description`, `initial_state`, `mode`, `states`), and strict unknown-field rejection.

## Lessons Learned
Locking schema shape choices early (key model, state container, strictness) prevents ambiguity leaking into execution-engine planning.
