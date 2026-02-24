---
id: b7c2d9e1
title: Semantic Search Phase 1 Discovery Learnings
type: "learning"
created_at: 2026-02-14T22:03:00+10:30
updated_at: 2026-02-14T22:03:00+10:30
status: completed
tags:
  - semantic-search
  - discovery
  - phase-1
---

# Semantic Search Phase 1 Discovery Learnings

## Summary
Phase 1 confirmed semantic search should stay optional, default to hybrid retrieval, and reuse existing query DSL semantics to avoid user-facing inconsistency.

## Details
- Hybrid retrieval should use Reciprocal Rank Fusion (RRF) for robust merge behavior without fragile score normalization.
- Result trust improves when each hit is labeled (`Exact match`, `Semantic match`, `Hybrid`) and optional explainability output is available.
- Phase 1 should produce concrete mode controls early (`hybrid`, `keyword`, `semantic`) to make behavior testable and debuggable.
- Query DSL compatibility is a non-negotiable requirement: exclusions and `--not` semantics must apply to both candidate sets before merge.
- Practical latency target for initial implementation remains P50 <= 250ms / P95 <= 750ms for notebooks up to 50k notes.

## Implications
- Phase 2 should focus on architecture decisions that preserve current DSL parser flow and apply filtering uniformly pre-merge.
- Implementation tasks should include benchmark harness work from day one to keep latency budgets visible.
- UX/docs should treat semantic search as an enhancement layer over Bleve, not a replacement.