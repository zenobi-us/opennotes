---
id: c9e4b1a2
title: Semantic Search Phase 2 Architecture Contract Learnings
type: "learning"
created_at: 2026-02-14T23:33:00+10:30
updated_at: 2026-02-14T23:33:00+10:30
status: completed
tags:
  - semantic-search
  - architecture
  - phase2
  - benchmarks
---

# Semantic Search Phase 2 Architecture Contract Learnings

## Summary
Phase 2 produced implementation-ready architecture contracts for hybrid retrieval, DSL filter parity, explainability output, retrieval mode controls, and benchmark validation. The key outcome is a low-risk integration path that preserves existing search behavior while adding semantic capabilities behind an explicit command surface.

## Details
- **Hybrid merge stability**: RRF + deterministic tie-breaks (coverage then path) is required for reproducible UX and testing.
- **Single query AST contract**: using one condition parsing/building path for both retrieval branches is essential to avoid semantic/keyword drift.
- **Explainability UX**: trust signals require dedicated output model/template; generic note list rendering is insufficient.
- **Mode controls**: isolating `--mode` semantics to semantic subcommand protects existing text/fuzzy command behavior.
- **Performance governance**: percentile targets only work with controlled run protocol, fixed corpus, and hardware metadata.

## Implications
- Phase 3 can begin implementation with clear interfaces and acceptance criteria.
- Existing CLI users should experience no regressions because semantic behavior is additive.
- Benchmark instrumentation should be built early in Phase 3 to catch integration regressions before final docs/release work.
