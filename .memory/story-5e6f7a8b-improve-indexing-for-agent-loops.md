---
id: 5e6f7a8b
title: Improve indexing for agent loops
epic_id: a7c3d9e1
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T01:42:00+10:30
status: completed
priority: medium
story_points: 3
---

# Improve indexing for agent loops

## User Story
As an agent, I want low-latency notebook open/search cycles so that repeated automation steps remain fast.

## Acceptance Criteria
- [x] Baseline current indexing/open cost is measured.
- [x] Proposal for incremental/persistent strategy is documented.
- [x] At least one measurable optimization path is validated.

## Context
Current index build on open may become bottleneck for high-frequency loops.

## Out of Scope
Immediate full redesign of search backend.

## Tasks
- [task-c5e6f7a8](task-c5e6f7a8-benchmark-optimize-indexing-agent-loops.md) — Benchmark and optimize indexing for agent loops (completed)

## Notes
Low-risk optimization implemented: in-process notebook/index cache with file-state invalidation.
