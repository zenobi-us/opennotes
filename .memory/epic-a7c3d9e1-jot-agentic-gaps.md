---
id: a7c3d9e1
title: Jot Agentic Gaps
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T01:45:00+10:30
status: completed
---

# Jot Agentic Gaps

## Vision/Goal
Close critical gaps so agentic workflows can use Jot as the single interface for query/search/read/write operations instead of direct filesystem manipulation.

## Success Criteria
- Jot supports full note lifecycle needed by agentic project workflows (create/read/update/append/move/archive/delete).
- Agent tasks can update `summary`, `todo`, and `team` notes without shell-level file write commands.
- Raw markdown round-trip is deterministic and lossless when required.
- Notebook operations are safe (root-bounded), machine-friendly (JSON/exit code), and scriptable.

## Phases
- Discovery and acceptance contract for agent-facing ops.
- CRUD+ops implementation and validation.
- Performance/safety hardening for iterative agent loops.

## Dependencies
- Existing notes/search/view command framework.
- Notebook root boundary and path resolution logic.
- Agreement on CLI contract for machine ops (`put/mv/cat/exists/ensure/append`).

## Critical Gaps (Captured)
1. No update/overwrite primitive for existing notes.
2. No move/archive primitive for workflow archiving.
3. No raw round-trip mode for deterministic automation.
4. No notebook-scoped utility ops (`exists`, `ensure`, append semantics).
5. In-memory reindexing on open may become expensive for agent loops.
