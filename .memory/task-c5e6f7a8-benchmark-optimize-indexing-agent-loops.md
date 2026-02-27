---
id: c5e6f7a8
title: Benchmark and optimize indexing for agent loops
type: "task"
created_at: 2026-02-27T08:45:00+10:30
updated_at: 2026-02-28T01:58:00+10:30
status: completed
epic_id: a7c3d9e1
story_id: 5e6f7a8b
assigned_to: 3219612528192551
---

# Benchmark and optimize indexing for agent loops

## Objective
Measure notebook open/index/search costs and validate at least one low-risk optimization path for repeated agent runs.

## Steps
1. Capture baseline timings for notebook open + note search across representative note counts.
2. Identify bottlenecks in index creation and document options.
3. Implement one low-risk optimization and re-benchmark.
4. Record before/after metrics and operational tradeoffs.

## Exit Criteria
- Baseline and post-change metrics are documented.
- At least one optimization path is validated empirically.
- No regression in correctness/safety behavior.

## Actual Outcome
Implemented notebook open/index caching in `NotebookService` with file-state invalidation:
- `NotebookService.Open()` now computes a notebook markdown file state fingerprint.
- If unchanged, it reuses the previously created in-memory notebook/index instance.
- If changed (path/size/modtime), it rebuilds the index to preserve correctness.

Added verification tests:
- `TestNotebookService_Open_PersistsIndexStateAndReusesWhenUnchanged`
- `TestNotebookService_Open_RebuildsIndexStateWhenNotesChange`

Added benchmark:
- `BenchmarkNotebookService_Open_AgentLoopIndexing`
  - latest verification run (2026-02-28):
    - `cold-open-reindex`: `146798405 ns/op`, `146471817 B/op`, `2778810 allocs/op`
    - `warm-open-cached`: `1295740 ns/op`, `280921 B/op`, `2037 allocs/op`

Observed impact for repeated loop opens in-process: ~113x faster (`~146.8ms` → `~1.30ms`).

## Lessons Learned
- The dominant cost for agent-loop opens is repeated full index construction, not search execution.
- A low-risk cache+invalidator approach gives major wins without changing search contracts.
- Persistence-based index reuse is higher-risk in this codebase due index lifecycle/lock semantics; cache reuse was safer for first pass.

## Notes
- Keep architecture changes minimal in first pass; prioritize measurable wins.

## Verification
- `mise run test` → PASS
- `mise exec -- go test -run '^$' -bench BenchmarkNotebookService_Open_AgentLoopIndexing -benchmem ./internal/services` → PASS with benchmark metrics above
