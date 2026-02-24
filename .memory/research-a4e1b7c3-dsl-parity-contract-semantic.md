---
id: a4e1b7c3
title: DSL Filter Parity Contract for Semantic Search
type: "research"
created_at: 2026-02-14T23:10:00+10:30
updated_at: 2026-02-14T23:10:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
related_task_id: b3d6a9c1
---

# DSL Filter Parity Contract for Semantic Search

## Research Questions
1. How should query conditions be parsed and applied across keyword and semantic paths?
2. How do we preserve existing `--and/--or/--not` behavior exactly?
3. What is the supported-field contract for semantic mode today?
4. What parity tests are required before implementation?

## Summary
Semantic and hybrid paths must reuse existing `SearchService.ParseConditions()` and AST construction (`BuildQuery()`) so boolean semantics remain identical. Filtering is applied pre-merge to both candidate sources. Current unsupported link fields (`links-to`, `linked-by`) remain hard errors in semantic/hybrid mode until link graph support exists.

## Findings

### 1) Source-of-truth parsing and AST order
Use existing pipeline as source of truth:
1. CLI flags (`--and/--or/--not`) → `ParseConditions()`
2. Conditions → `BuildQuery()` AST
3. Apply query to keyword retrieval and semantic candidate filter stage

Current semantics from `BuildQuery()`:
- all `and` expressions are required
- all `or` expressions are grouped into one OR tree
- all `not` expressions are wrapped as NOT and applied as exclusions

### 2) Pre-merge filtering contract
For hybrid mode:
- Build one canonical `search.Query` from conditions.
- Keyword branch: execute `index.Find()` with query directly.
- Semantic branch: retrieve candidate set then apply same query-compatible filter contract before merge.
- Merge only documents passing filter contract.

Rationale: prevents post-merge leakage where excluded docs reappear from the other branch.

### 3) Unsupported filter behavior
Current service behavior already rejects link fields in `BuildQuery()` with explicit error.

Contract for Phase 3:
- `links-to`, `linked-by`: fail fast with clear unsupported message in semantic/hybrid mode.
- Recommended fallback remains keyword query mode where applicable, with message consistency.

### 4) Supported filter parity matrix
Must behave identically in keyword, semantic, and hybrid modes for:
- `data.*` fields (`tag`, `status`, `priority`, `assignee`, `author`, `type`, `category`, `project`, `sprint`)
- alias `data.tags` → `data.tag`
- `path` exact/prefix/wildcard
- `title`
- boolean composition: AND/OR/NOT

### 5) Required test matrix for implementation phase
1. Same-condition equivalence: keyword vs hybrid-filtered keyword-only baseline.
2. NOT exclusion parity with mixed AND/OR/NOT combinations.
3. Alias parity (`data.tags` equals `data.tag`).
4. Path wildcard parity (`*`, `**`, `?`).
5. Unsupported link-field errors in semantic/hybrid return same class/message pattern.

## References
- [internal/services/search.go](../internal/services/search.go)
- [internal/services/search_test.go](../internal/services/search_test.go)
- [cmd/notes_search_query.go](../cmd/notes_search_query.go)
- [task-b3d6a9c1-dsl-filter-parity-semantic.md](task-b3d6a9c1-dsl-filter-parity-semantic.md)
