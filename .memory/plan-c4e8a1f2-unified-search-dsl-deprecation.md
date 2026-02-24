---
id: c4e8a1f2
title: Unified Search DSL Deprecation Plan
type: "plan"
created_at: 2026-02-23T23:15:45+10:30
updated_at: 2026-02-23T23:15:45+10:30
status: ready
epic_id: 9b7c2a4d
assigned_to: unassigned
---

# Unified Search DSL Deprecation Plan

## Scope
Deprecate and remove legacy `jot notes search query --and/--or/--not` with minimal disruption and clear user migration path.

## Execution Order
1. Ship DSL-first behavior and examples where search entry points are documented.
2. Add explicit runtime deprecation warnings and docs migration section for legacy command.
3. Remove legacy command in next minor release and keep compatibility notes in changelog.

## Guardrails
- Keep existing tests passing during warning period.
- Avoid behavior regressions for `notes search semantic` filter flags.
- Ensure release notes include migration snippets.

## Deliverables
- Deprecation metadata in code.
- Reminder test gate for removal target + epic linkage.
- Milestone tasks completed and reflected in epic status.
