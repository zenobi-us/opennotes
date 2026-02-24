---
id: f6b2d1a9
title: Semantic Latency Benchmark Harness Plan
type: "research"
created_at: 2026-02-14T23:27:00+10:30
updated_at: 2026-02-14T23:27:00+10:30
status: completed
epic_id: 7c9d2e1f
phase_id: b2f4c8d1
related_task_id: e9b2d4f6
---

# Semantic Latency Benchmark Harness Plan

## Research Questions
1. What datasets and query corpus are needed for reliable semantic latency validation?
2. How should latency be measured for keyword/semantic/hybrid modes?
3. Which reproducibility controls avoid noisy results?
4. How should pass/fail and release reporting be defined?

## Summary
Implement a deterministic benchmark harness that runs fixed corpora against three retrieval modes and reports P50/P95 latency. Use notebook scales of 1k/10k/50k with repeated runs, warmup passes, and environment capture to validate the `P95 <= 750ms` target for <=50k notes.

## Findings

### 1) Dataset generation plan
Benchmark notebook sizes:
- small: 1,000 notes
- medium: 10,000 notes
- target upper bound: 50,000 notes

Dataset characteristics:
- realistic markdown bodies and frontmatter metadata (`tag`, `status`, `priority`, `title`)
- mixed document lengths (short/medium/long)
- deterministic seeded generation for reproducibility

Existing helpers from search benchmarks can be adapted (see `internal/services/search_bench_test.go`).

### 2) Fixed query corpus
Run 20 canonical queries per dataset:
- 5 exact keyword queries
- 5 paraphrase/conceptual queries
- 5 metadata-filtered queries (AND/NOT combinations)
- 5 mixed difficult queries (long phrase + exclusions)

For each query run 5 iterations after warmup; aggregate per-mode latencies:
- mode `keyword`
- mode `semantic`
- mode `hybrid`

### 3) Measurement and percentile contract
Per mode and dataset, collect:
- min, max, mean
- P50 latency
- P95 latency
- total query count

Timing scope:
- start immediately before retrieval orchestration call
- end after results returned (before rendering)

Output format (machine + human):
- JSON artifact for CI comparison
- markdown table for release notes and docs

### 4) Reproducibility controls
- fixed random seed for synthetic corpus
- 3 warmup runs per mode before measured runs
- disable unrelated background indexing activity
- capture hardware/OS metadata (CPU model, cores, memory class)
- keep benchmark notebook on local disk (no network FS)

### 5) Pass/fail thresholds
Primary acceptance (from story):
- for <=50k notes: `P95 <= 750ms`

Secondary guardrails:
- for <=50k notes: `P50 <= 250ms`
- no single mode regresses by >20% vs previous baseline without explicit note

### 6) Release report format
Release notes section must include:
- dataset size(s)
- mode-wise P50/P95 table
- hardware profile
- whether thresholds passed/failed
- caveats (cold cache, semantic backend unavailability fallback)

### 7) Risk register (Phase 2 requirement)
1. **Hardware variance risk**: laptop differences can skew percentile values.
   - Mitigation: include hardware profile and compare like-for-like runs.
2. **Warm-cache bias**: repeated runs may underestimate first-query behavior.
   - Mitigation: report cold-start sample separately.
3. **Synthetic-data bias**: generated notes may not reflect real notebooks.
   - Mitigation: add optional real-notebook benchmark profile in Phase 3.

## References
- [internal/services/search_bench_test.go](../internal/services/search_bench_test.go)
- [internal/search/bleve/index_bench_test.go](../internal/search/bleve/index_bench_test.go)
- [task-e9b2d4f6-benchmark-harness-latency-validation.md](task-e9b2d4f6-benchmark-harness-latency-validation.md)
