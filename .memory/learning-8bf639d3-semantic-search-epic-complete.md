---
id: 8bf639d3
title: Semantic Search Enhancement Epic - Complete Learnings
type: "learning"
created_at: 2026-02-17T18:40:00+10:30
updated_at: 2026-02-17T18:40:00+10:30
status: completed
tags:
  - lessons-learned
  - best-practices
  - technical-insights
  - semantic-search
  - hybrid-retrieval
---

# Semantic Search Enhancement Epic - Complete Learnings

## Summary
The Semantic Search Enhancement epic (7c9d2e1f) successfully added optional semantic search capabilities to OpenNotes, augmenting the existing Bleve full-text search with vector-based relevance. This learning captures insights from all four phases: Research & Discovery, Architecture & Design, Implementation & Testing, and Documentation & Release.

## Details

### Technical Architecture Insights

**1. chromem-go as Semantic Backend**
- Pure Go implementation eliminates CGO dependencies
- Local embedding generation (all-MiniLM-L6-v2) preserves privacy
- Single-file storage (`.opennotes.semantic.gob`) simplifies lifecycle
- Noop fallback pattern allows graceful degradation when embeddings unavailable

**2. Hybrid Retrieval with RRF Merge**
- Reciprocal Rank Fusion (RRF) provides deterministic, parameter-light merging
- k=60 constant balances relevance across keyword and semantic results
- Source labels (`[keyword]`, `[semantic]`, `[both]`) build user trust
- Default hybrid mode maximizes recall without mode selection complexity

**3. Query DSL Parity**
- Semantic path preserves all existing filters: `tag:`, `notebook:`, `is:archived`
- Post-filtering strategy ensures backward compatibility
- Query-string fallback when pure semantic returns sparse results

### Process Insights

**1. Contract-First Phase 2 Approach**
- Producing research artifacts as "contracts" (hybrid retrieval, explainability, mode controls) before implementation reduced ambiguity
- Clear acceptance criteria enabled fast validation during Phase 3
- Phase boundaries created natural review checkpoints with human oversight

**2. Incremental Task Breakdown**
- Phase 1: 5 discovery tasks → Phase 2: 5 architecture tasks → Phase 3: 5 implementation tasks → Phase 4: 3 documentation tasks
- Small task granularity enabled frequent commits and progress visibility
- Each task had clear deliverables and verification steps

**3. Documentation as Validation**
- Writing `docs/semantic-search-guide.md` revealed edge cases not covered in implementation
- README feature section forced clarity on user-facing value proposition
- Cross-linking docs/INDEX.md ensured discoverability

### Performance Insights

**1. Benchmark Harness Design**
- Deterministic test corpus (100 notes, fixed seed) enables reproducible benchmarks
- P50/P95 latency metrics provide operational targets
- Threshold checks integrated into test suite prevent regression

**2. Latency Targets**
- P95 < 500ms for typical notebook sizes (< 1000 notes)
- First-query cold start acceptable with user guidance
- Index lifecycle cost amortized across queries

### User Experience Insights

**1. Explainability via `--explain`**
- Match snippets with highlight markers build trust in semantic relevance
- Score display helps users calibrate expectations
- Mode labels clarify result provenance in hybrid mode

**2. Mode Controls**
- Dedicated `semantic` subcommand avoids polluting default search
- `--mode keyword|semantic|hybrid` provides explicit control
- Hybrid default balances recall with minimal user configuration

## Implications

### For Future Enhancements
1. **Model Upgrade Path**: chromem-go supports model swaps; monitor embedding model landscape
2. **Incremental Indexing**: Current full-reindex approach may need optimization for large notebooks
3. **Batch Embedding**: Consider background embedding generation for better UX on large imports

### For Project Process
1. **Contract-driven design phases** work well for optional features with clear boundaries
2. **Small task granularity** (5 tasks per phase) keeps momentum and enables checkpointing
3. **Documentation as final phase** catches gaps and validates user-facing clarity
4. **Human review gates** at phase boundaries ensure alignment without blocking velocity

### For Architecture
1. **Noop fallback pattern** simplifies integration of optional features
2. **Source labeling** for hybrid results increases transparency
3. **Post-filter strategy** preserves backward compatibility when adding search layers

## Related Artifacts

### Phase Learnings (Consolidated)
- [learning-b7c2d9e1-semantic-phase1-discovery.md](learning-b7c2d9e1-semantic-phase1-discovery.md)
- [learning-c9e4b1a2-semantic-phase2-architecture-contracts.md](learning-c9e4b1a2-semantic-phase2-architecture-contracts.md)

### Research Artifacts
- [research-3f2a9c1b-semantic-search.md](research-3f2a9c1b-semantic-search.md) - Technical synthesis
- [research-f2d6a8c0-hybrid-retrieval-contract.md](research-f2d6a8c0-hybrid-retrieval-contract.md)
- [research-a4e1b7c3-dsl-parity-contract-semantic.md](research-a4e1b7c3-dsl-parity-contract-semantic.md)
- [research-d2f5a7c9-explainability-output-contract.md](research-d2f5a7c9-explainability-output-contract.md)
- [research-e1c3a5d7-retrieval-mode-controls-contract.md](research-e1c3a5d7-retrieval-mode-controls-contract.md)

### Epic & Phases (Archived)
- [epic-7c9d2e1f-semantic-search.md](archive/epic-7c9d2e1f-semantic-search.md)
- [phase-52a9f0b3-semantic-search-research.md](archive/phase-52a9f0b3-semantic-search-research.md)
- [phase-b2f4c8d1-architecture-integration-design.md](archive/phase-b2f4c8d1-architecture-integration-design.md)
- [phase-91d3f6a2-implementation-testing.md](archive/phase-91d3f6a2-implementation-testing.md)
- [phase-4e8a3c2d-documentation-release.md](archive/phase-4e8a3c2d-documentation-release.md)

## Metrics

| Metric | Value |
|--------|-------|
| Epic Duration | ~15 days (Feb 2 - Feb 17, 2026) |
| Total Phases | 4 |
| Total Tasks | 18 |
| Commits | ~30 (implementation-specific) |
| Files Changed | ~20 (cmd, internal, docs) |
| Test Coverage | Existing coverage maintained |
| Binary Size Impact | Minimal (chromem-go pure Go) |
