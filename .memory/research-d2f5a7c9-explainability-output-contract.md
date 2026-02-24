---
id: d2f5a7c9
title: Explainability Output Contract for Semantic Search
type: "research"
created_at: 2026-02-14T23:15:00+10:30
updated_at: 2026-02-14T23:15:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
related_task_id: c5f8b2d4
---

# Explainability Output Contract for Semantic Search

## Research Questions
1. What CLI output shape should `--explain` produce?
2. How should snippets be chosen for keyword vs semantic matches?
3. What fallback and truncation rules keep output readable?
4. What template/display constraints affect implementation?

## Summary
`--explain` should emit compact per-result context with match label + one snippet line, while preserving the current non-explain output path. Because current `note-list` template lacks explain fields, semantic explain mode should use a dedicated render template and a typed result DTO.

## Findings

### 1) Output contract
For semantic search list output (`--explain` enabled), each result row includes:
- display title
- relative path
- match label (`Exact match`, `Semantic match`, `Hybrid`)
- explain snippet (single line)

Recommended structure (logical):
```text
- [Title] path/to/note.md  (Hybrid)
  Why: "...best matching snippet..."
```

Without `--explain`, preserve current concise note list behavior.

### 2) Snippet selection rules
Keyword-first rules:
1. Use top snippet from keyword hit ranges if available.
2. Highlight matched terms with bracket markers (e.g., `[term]`).
3. If no range/snippet available, fallback to first sentence containing query token.

Semantic rules:
1. Use semantic best-sentence candidate from semantic scorer context.
2. If sentence unavailable, fallback to note lead.
3. If lead unavailable, fallback to title only and reason text `"No snippet available"`.

### 3) Truncation and ordering
- Default snippet length: 160 chars.
- Hard cap: 220 chars after highlighting.
- Add ellipsis when truncated.
- Keep result order as produced by retrieval merge; explain mode must not reorder.

### 4) Rendering constraints in current code
Current state:
- `displayNoteList()` renders `note-list.gotmpl` with only `.DisplayName` + `.File.Relative`.
- No fields for match type or snippet in existing `services.Note` output.

Implementation contract:
- Introduce a semantic-specific view model, e.g. `SearchResultView` with `MatchType`, `Explain`, `Score` (optional hidden), and embedded note fields.
- Add dedicated template, e.g. `internal/services/templates/note-search-semantic.gotmpl`.
- Route explain output through semantic command path, not general `displayNoteList()`.

### 5) Acceptance-test scenarios
1. `--explain` + keyword-only result shows highlighted keyword snippet.
2. `--explain` + semantic-only result shows semantic sentence fallback.
3. `--explain` + hybrid result shows `Hybrid` label and snippet.
4. no `--explain` uses standard concise note list.
5. empty snippet paths render safe fallback message without panic.

## References
- [cmd/notes_list.go](../cmd/notes_list.go)
- [cmd/notes_search.go](../cmd/notes_search.go)
- [internal/services/display.go](../internal/services/display.go)
- [internal/services/templates/note-list.gotmpl](../internal/services/templates/note-list.gotmpl)
- [task-c5f8b2d4-explainability-output-contract.md](task-c5f8b2d4-explainability-output-contract.md)
