---
id: a1b2c3d4
title: Parallel Research Methodology for Technology Decisions
type: "learning"
created_at: 2026-02-19T08:11:00+10:30
updated_at: 2026-02-19T08:11:00+10:30
status: active
tags: [research, methodology, decision-making, technology-evaluation]
epic_id: f661c068
---

# Parallel Research Methodology for Technology Decisions

## Summary

When replacing a core dependency (DuckDB → Bleve), parallel subtopic research across 4 independent domains produced high-confidence decisions with zero regrets. The structured approach—researching alternatives, existing implementations, query design, and performance baselines simultaneously—compressed 2+ weeks of ad-hoc research into 4 focused hours.

## Details

### The Approach

The DuckDB removal epic (f661c068) used a parallel research structure with 4 independent subtopics:

| Subtopic | Focus | Key Output |
|----------|-------|------------|
| 1. ZK Search Architecture | Study existing pure-Go note tool | Interface patterns to adopt (not implementation) |
| 2. Vector RAG Libraries | Evaluate semantic search options | chromem-go selected, deferred to future epic |
| 3. Query DSL Design | Research query syntax patterns | Gmail-style DSL confirmed as gold standard |
| 4. Performance Baseline | Profile current DuckDB overhead | 29% wasted overhead quantified; targets set |

### Why It Worked

1. **Independence**: Each subtopic could be researched without blocking others
2. **Cross-verification**: Findings from one subtopic validated others (e.g., zk's query syntax confirmed DSL research)
3. **Decision matrix**: Aggregated findings into clear technology choices with rationale
4. **Confidence levels**: Each finding tagged with confidence (HIGH/MEDIUM/LOW) based on source count

### The Decision Matrix Pattern

For each technology choice, evaluate:
- **Pure Go?** (eliminates CGO complexity)
- **Maturity** (years in production, community size)
- **Performance characteristics** (benchmarks, not marketing)
- **Interface compatibility** (can we adopt patterns from proven tools?)
- **Deployment simplicity** (single binary? external services?)

### Key Insight: Study Interfaces, Not Implementations

The zk research (subtopic 1) was pivotal. Instead of copying zk's SQLite+FTS5 implementation (which requires CGO), we adopted its **interface design** (`NoteIndex`, `FileStorage`, `NoteFindOpts`) and implemented them with pure Go alternatives. This gave us battle-tested API design without the dependency baggage.

### Quantifying the Problem First

Subtopic 4 (performance profiling) proved DuckDB was **pure overhead** for search:
- Search was already in-memory; DuckDB added 29% CGO overhead
- DuckDB contributed 37.8MB to binary size
- Extension downloads failed 30-50% in CI

This data made the "remove entirely" decision obvious—no migration period needed.

## Implications

- **For future epics**: Always research in parallel subtopics before implementation. 4 hours of structured research prevents weeks of wrong-direction work.
- **For technology evaluation**: Study existing tools' interfaces, not their implementations. Adopt proven API patterns with your own backend.
- **For decision confidence**: Quantify the problem (profiling, size analysis) before choosing solutions. Data eliminates debate.
- **Related learnings**: [archive/duckdb-removal-f661c068/research-f410e3ba-search-replacement-synthesis.md](archive/duckdb-removal-f661c068/research-f410e3ba-search-replacement-synthesis.md)
