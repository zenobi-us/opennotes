---
id: e1c3a5d7
title: Retrieval Mode Controls and Validation Contract
type: "research"
created_at: 2026-02-14T23:20:00+10:30
updated_at: 2026-02-14T23:20:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
related_task_id: d7a1c3e5
---

# Retrieval Mode Controls and Validation Contract

## Research Questions
1. Where should mode controls live in CLI surface?
2. How should mode flags interact with existing search flags?
3. What validation and warning behavior is required?
4. How do we preserve backward compatibility?

## Summary
Mode controls should be explicit and centralized on semantic search command path via `--mode {hybrid|keyword|semantic}` with default `hybrid`. Invalid values fail fast. Existing `notes search` text/fuzzy behavior remains unchanged for backward compatibility.

## Findings

### 1) Command surface contract
Recommended command path:
- `opennotes notes search semantic [query] --mode hybrid|keyword|semantic`

Default behavior:
- if `--mode` omitted → `hybrid`

Reason:
- keeps current `notes search` and `--fuzzy` behavior stable
- isolates semantic rollout behind dedicated subcommand (story requirement)

### 2) Flag interaction contract
On semantic subcommand:
- allowed: `--mode`, `--explain`, boolean condition flags (`--and`, `--or`, `--not`)
- disallowed: `--fuzzy` (text/fuzzy-only concern)

Validation rules:
1. `--mode` must be one of `hybrid`, `keyword`, `semantic`
2. unknown mode returns Cobra validation error and non-zero exit
3. any incompatible flag combo (`--fuzzy` on semantic command) returns clear error and usage hint

### 3) No-result warning contract
If search returns zero results:
- `--mode keyword`:
  - warning: `No keyword-mode results. Try --mode hybrid or --mode semantic.`
- `--mode semantic`:
  - warning: `No semantic-mode results. Try --mode hybrid or --mode keyword.`
- `--mode hybrid`:
  - no mode-switch warning; normal "No notes found" messaging

Warnings are advisory and should not change exit code (still success when query executes correctly).

### 4) Shared search pathway behavior
Service contract should accept mode enum and return metadata about executed branches so CLI can render accurate warning text.

Suggested service API direction:
```go
SearchSemantic(ctx context.Context, query string, conditions []QueryCondition, mode RetrievalMode, explain bool) ([]SemanticSearchResult, SearchExecutionMeta, error)
```

Where `SearchExecutionMeta` reports:
- selected mode
- keyword branch attempted?
- semantic branch attempted?
- semantic backend availability/fallback

### 5) Migration and compatibility impact
- No behavior change for existing commands:
  - `opennotes notes search`
  - `opennotes notes search --fuzzy`
  - `opennotes notes search query ...`
- Semantic mode controls are additive and opt-in.

## References
- [cmd/notes_search.go](../cmd/notes_search.go)
- [cmd/notes_search_query.go](../cmd/notes_search_query.go)
- [docs/commands/notes-search.md](../docs/commands/notes-search.md)
- [task-d7a1c3e5-retrieval-mode-controls-cli.md](task-d7a1c3e5-retrieval-mode-controls-cli.md)
