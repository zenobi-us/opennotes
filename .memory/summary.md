# Jot Project Summary

## Current Focus
- Active execution includes epic `9b7c2a4d` (Unified Search DSL Deprecation).
- Completed `task-1f2e3d4c`: `notes search` now routes colon-based queries (e.g., `type:epic`) through DSL execution.
- Completed `task-5a6b7c8d`: legacy `search query` now emits runtime deprecation warning and docs include unified DSL migration guidance.
- Next milestone: `task-9c0d1e2f` (remove legacy `search query` command in next minor).
- Verification passed with `mise run test`.

## Archived (Cleanup)
- [epic-8361d3a2](archive/rename-to-jot-8361d3a2/epic-8361d3a2-rename-to-jot.md) — Rename Project from OpenNotes to Jot
- [task-b66d2263](archive/rename-to-jot-8361d3a2/task-b66d2263-identify-rename-locations.md)
- [task-8281af6b](archive/rename-to-jot-8361d3a2/task-8281af6b-notebook-migrate-versioned-framework.md)
- [task-6c834006](archive/cleanup-2026-02-22/task-6c834006-view-save-delete-cli-flags.md)
- [task-ddbaa84b](archive/cleanup-2026-02-22/task-ddbaa84b-view-cli-override-flags.md)
- [task-a82526f2](archive/cleanup-2026-02-22/task-a82526f2-dsl-or-syntax-support.md)
- [task-b66d2263](archive/cleanup-2026-02-22/task-b66d2263-identify-rename-locations.md)

## Parked Work
- [task-9c4a2f8d](task-9c4a2f8d-github-actions-moonrepo-releases.md) — GitHub Actions CI/CD
- [plan-b4e2f7a1](plan-b4e2f7a1-dsl-views-implementation.md) — DSL views implementation
- [phase-4adb81db](phase-4adb81db-dsl-views-deferred-followups.md) — Deferred DSL views follow-up phase (active)

## Open Tasks (Phase 4adb81db)
- [x] [task-684a9a73](task-684a9a73-view-parameter-substitution.md) — `{{param_name}}` substitution
- [x] [task-e19963c7](task-e19963c7-global-views-config-support.md) — global views in user config

## Learnings
- Added: [learning-f1c2d3e4](learning-f1c2d3e4-dsl-views-cleanup-archive-pattern.md) documenting cleanup/archive decisions and implications.
- Parameterized views require strict execution order to avoid unresolved template values and false type failures.
