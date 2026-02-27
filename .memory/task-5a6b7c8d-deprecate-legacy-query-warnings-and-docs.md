---
id: 5a6b7c8d
title: Deprecate legacy query warnings and docs
type: "task"
created_at: 2026-02-23T23:15:45+10:30
updated_at: 2026-02-26T08:30:34+10:30
status: completed
epic_id: 9b7c2a4d
plan_id: c4e8a1f2
assigned_to: 4133247916144172
---

# Deprecate legacy query warnings and docs

## Objective
Add explicit deprecation metadata and user-facing migration guidance for legacy `--and/--or/--not` command path.

## Steps
1. Mark legacy command code paths with `@deprecated` including since/removal placeholders.
2. Add runtime warning and docs section that points users to unified DSL equivalents.
3. Ensure references include tracking epic for release management.

## Exit Criteria
- Code contains deprecation markers and tracking link.
- Docs mention migration path and removal timeline placeholder.

## Actual Outcome
- Added runtime warning emission for legacy `jot notes search query` path in `cmd/notes_search_query.go`.
- Added `legacyQueryDeprecationWarningMessage()` with timeline placeholders, DSL migration example, and tracking epic link.
- Added test coverage in `cmd/notes_search_query_deprecation_test.go` for migration guidance content.
- Updated `docs/commands/notes-search.md` to mark `query` subcommand as deprecated and provide unified DSL migration examples.
- Verification: `mise run test` passed.
