---
id: f2d6a8c0
title: Hybrid Retrieval Orchestration Contract
type: "research"
created_at: 2026-02-14T23:05:00+10:30
updated_at: 2026-02-14T23:05:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
related_task_id: a1c4e7f2
---

# Hybrid Retrieval Orchestration Contract

## Research Questions
1. Where should hybrid orchestration integrate with current command/service flow?
2. What are candidate set boundaries for keyword and semantic retrieval?
3. How should RRF merge be deterministic and stable?
4. How should match labels and empty-source fallback work?

## Summary
Define hybrid retrieval as an orchestration layer in `internal/services` that combines existing keyword retrieval from `search.Index` with semantic retrieval from a new semantic index adapter, then merges results with deterministic RRF + path tie-break ordering.

## Findings

### 1) Integration points in current flow
Current keyword flow:
- `cmd/notes_search.go` → `NoteService.SearchNotes()` (text/fuzzy in-memory)
- `cmd/notes_search_query.go` → `NoteService.SearchWithConditions()` (AST query to Bleve index)

Hybrid integration contract:
- Add a semantic search command path (dedicated semantic subcommand) that calls new service method:
  - `NoteService.SearchSemantic(ctx, query string, conditions []QueryCondition, mode RetrievalMode, opts SemanticFindOpts)`
- Keep `SearchWithConditions()` unchanged for current behavior compatibility.
- Reuse `SearchService.ParseConditions()` and `BuildQuery()` for DSL parity and shared filtering semantics.

### 2) Candidate set boundaries
- Keyword candidate list source: `index.Find()` using `search.Query` + existing sort/relevance behavior.
- Semantic candidate list source: new semantic adapter `SemanticIndex.FindSimilar()`.
- Default candidate limits before merge:
  - `keywordTopK = 100`
  - `semanticTopK = 100`
- Filters/conditions must be applied to both sources before merge.

### 3) Deterministic RRF merge
For each unique path `d`:
- `score(d) = Σ 1/(k + rank_i(d))` for each source list `i` where document appears.
- Default `k = 60`.

Deterministic ordering rules:
1. Descending RRF score
2. Descending source coverage (appears in both lists > one list)
3. Ascending normalized path (lexicographic)

This guarantees stable ordering across runs when source rankings are unchanged.

### 4) Match labels and fallback
Match labeling:
- `Exact match`: present only in keyword list
- `Semantic match`: present only in semantic list
- `Hybrid`: present in both lists

Fallback behavior:
- If one list is empty, return the other list unchanged.
- If both are empty, return empty with mode-appropriate guidance from CLI layer.
- If semantic backend unavailable, hybrid mode degrades to keyword-only with warning.

### 5) Suggested types/contracts (Phase 3 implementation target)
```go
type RetrievalMode string

const (
    RetrievalModeHybrid   RetrievalMode = "hybrid"
    RetrievalModeKeyword  RetrievalMode = "keyword"
    RetrievalModeSemantic RetrievalMode = "semantic"
)

type MatchType string

const (
    MatchExact    MatchType = "exact"
    MatchSemantic MatchType = "semantic"
    MatchHybrid   MatchType = "hybrid"
)

type SemanticResult struct {
    Document search.Document
    Score    float64
}

type SemanticIndex interface {
    FindSimilar(ctx context.Context, query string, topK int) ([]SemanticResult, error)
}
```

## References
- [cmd/notes_search.go](../cmd/notes_search.go)
- [cmd/notes_search_query.go](../cmd/notes_search_query.go)
- [internal/services/note.go](../internal/services/note.go)
- [internal/services/search.go](../internal/services/search.go)
- [task-a1c4e7f2-hybrid-retrieval-orchestration.md](task-a1c4e7f2-hybrid-retrieval-orchestration.md)
