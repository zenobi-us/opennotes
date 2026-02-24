---
id: 9b7c2a4d
title: Unified Search DSL Deprecation
type: "epic"
created_at: 2026-02-23T23:15:45+10:30
updated_at: 2026-02-23T23:15:45+10:30
status: planned
---

# Unified Search DSL Deprecation

## Goal
Retire legacy `jot notes search query --and/--or/--not` in favor of a single unified search DSL surface, with explicit migration messaging and a scheduled removal in the next minor release window.

## Why
- Current dual search UX splits user guidance and docs.
- Legacy boolean flags are harder to compose than DSL query expressions.
- A single search path reduces maintenance and support burden.

## Milestones
1. Introduce unified DSL behavior as primary guidance and compatibility path.
2. Add deprecation warning and docs migration notes for legacy command.
3. Remove legacy command in the next minor release after warning period.

## Tracking
- Plan: `.memory/plan-c4e8a1f2-unified-search-dsl-deprecation.md`
- Task 1: `.memory/task-1f2e3d4c-introduce-unified-search-dsl-behavior.md`
- Task 2: `.memory/task-5a6b7c8d-deprecate-legacy-query-warnings-and-docs.md`
- Task 3: `.memory/task-9c0d1e2f-remove-legacy-query-command-next-minor.md`
