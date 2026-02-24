---
id: e19963c7
title: Add global views support in user config hierarchy
type: "task"
created_at: 2026-02-20T19:14:00+10:30
updated_at: 2026-02-22T22:53:00+10:30
status: done
epic_id: f661c068
phase_id: b4e2f7a1-plan
assigned_to: null
---

# Add global views support in user config hierarchy

## Objective

Enable global custom views in `~/.config/jot/config.json` with clear precedence versus notebook-local and built-in views.

## Related Story

- Plan deferred item #5 in [plan-b4e2f7a1-dsl-views-implementation.md](plan-b4e2f7a1-dsl-views-implementation.md)

## Steps

1. Implement/confirm global config loading for views in `ViewService`.
2. Define precedence contract and conflict resolution (notebook > global > built-in).
3. Add tests covering lookup, list output origin, and override behavior.
4. Add migration-safe handling for missing/invalid global config.
5. Document global views configuration and examples.

## Expected Outcome

Users can define once in global config and run the same view in any notebook, with documented precedence behavior.

## Actual Outcome

Implemented with TDD:
- added RED coverage for list-origin precedence and malformed global config fallback
- updated `ListAllViews()` to apply effective precedence (`notebook > global > built-in`) when duplicate names exist
- verified global config parse failures do not block built-in view lookup
- updated views guide config examples and list precedence wording
- full test suite passed via `mise run test`

## Lessons Learned

`GetView()` and `ListAllViews()` can silently diverge in precedence behavior unless both are covered by explicit collision tests.
