---
id: 1f2e3d4c
title: Introduce unified search DSL behavior
type: "task"
created_at: 2026-02-23T23:15:45+10:30
updated_at: 2026-02-25T09:10:00+10:30
status: completed
epic_id: 9b7c2a4d
plan_id: c4e8a1f2
assigned_to: unassigned
---

# Introduce unified search DSL behavior

## Objective
Promote unified DSL-based search syntax as the primary search UX, while preserving temporary compatibility for legacy path.

## Steps
1. Ensure `notes search` docs and help examples prioritize DSL query usage.
2. Validate semantic search references DSL filter compatibility clearly.
3. Keep legacy command operational during deprecation period.

## Exit Criteria
- DSL appears as default recommendation in help/docs.
- No functional regressions in current search commands.

## Actual Outcome (2026-02-25)
- `notes search` now routes colon-based queries (for example `type:epic`) through the DSL search path.
- Parser grammar now supports `type:<value>` as a first-class field query.
- Bleve field normalization maps `type` to `metadata.type` for indexed filtering.
- Added/updated tests:
  - `cmd/notes_search_test.go` (`isDSLStyleQuery`)
  - `internal/search/parser/parser_test.go` (`type:epic` field parsing)
  - `internal/search/bleve/query_test.go` (metadata field normalization)
- Verification:
  - `mise run test`
  - `go run . notes search "type:epic"` (returns expected epic notes)
