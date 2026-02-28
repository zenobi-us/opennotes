---
id: f3a4b5c6
title: Record workflow constraints and non-goals
epic_id: b2f4e6a8
story_id: 6f7a8b9c
created_at: 2026-02-28T00:53:14+10:30
updated_at: 2026-02-28T01:28:00+10:30
status: completed
assigned_to: "3219612528192551"
---

# Record workflow constraints and non-goals

## Objective
Document hard constraints and explicit non-goals to prevent scope bleed across DSL and engine stories.

## Related Story
[story-6f7a8b9c](story-6f7a8b9c-workflow-discovery-and-requirements.md)

## Steps
1. Identify technical constraints (metadata model, notebook scope, compatibility boundaries).
2. Define explicit non-goals for this epic.
3. Publish constraint register linked from epic and story artifacts.

## Expected Outcome
A clear constraint/non-goal register referenced by planning and implementation artifacts.

## Actual Outcome
Constraints and non-goals recorded in [research-a9b8c7d6](research-a9b8c7d6-workflow-discovery-brief.md), including:
- Notebook-scoped workflows in `.jot.json` keyed objects
- Full JSON Schema per state
- Explicit per-state transitions
- V1 non-goals (A/B/C/D/F)
- Migration handled by existing config migration framework

## Lessons Learned
Treating migration as an existing platform capability, not a new workflow feature, removes unnecessary debate and keeps discovery focused.
