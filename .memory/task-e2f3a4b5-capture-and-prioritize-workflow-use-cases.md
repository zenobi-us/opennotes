---
id: e2f3a4b5
title: Capture and prioritize workflow use cases
epic_id: b2f4e6a8
story_id: 6f7a8b9c
created_at: 2026-02-28T00:53:14+10:30
updated_at: 2026-02-28T01:10:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Capture and prioritize workflow use cases

## Objective
Produce a prioritized set of workflow use cases that determines DSL and execution engine scope.

## Related Story
[story-6f7a8b9c](story-6f7a8b9c-workflow-discovery-and-requirements.md)

## Steps
1. Collect candidate use cases from Jot operational workflows.
2. Normalize into a common format (trigger, actors, state transitions, expected outputs).
3. Prioritize into P0/P1/P2 with rationale.

## Expected Outcome
A ranked workflow use-case list with enough detail to derive acceptance criteria for story `7a8b9c0d`.

## Actual Outcome
Prioritized use-case ranking captured in [research-a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md):
- P0: Authoring governance
- P0: Agent safety/compliance
- P1: Execution orchestration
- P2: State rollback/reopen patterns

## Lessons Learned
Prioritizing by enforcement risk first (P0) gives a stable contract foundation for orchestration features.
