---
id: 8281af6b
title: Plan Versioned `jot notebook migrate` Framework
type: "task"
created_at: 2026-02-19T21:03:59+10:30
updated_at: 2026-02-19T21:03:59+10:30
status: planning
epic_id: 8361d3a2
phase_id: phase-4-external-updates
assigned_to: planning-2026-02-19
---

# Plan Versioned `jot notebook migrate` Framework

## Objective

Define an initial implementation vision for evolving `jot notebook migrate` from a one-off rename helper into a versioned, extensible migration framework that can safely support future notebook/config/index schema changes.

## Related Story

N/A (planning-first task; story can be created after review).

## User-Approved Core Principles

1. **Code-only migrations**
   - Migrations live in code files only (example naming: `internal/migrations/00001_<name>.go`).
2. **Notebook config stores only version**
   - Notebook config (JSON/TOML/YAML) contains only migration schema version metadata for migration state (`config_version`).
3. **Each migration declares target version**
   - Each migration explicitly describes the version it upgrades to.
4. **Every migration supports up/down**
   - Each migration provides both `Up` and `Down`.

## Design Constraints Added During Review

5. **Explicit from/to contracts**
   - Each migration should declare both `from_version` and `to_version` to prevent ambiguous chains.
6. **Strict linear chain validation**
   - No gaps, no duplicate versions, deterministic ordering.
7. **Dry-run + idempotency guarantees**
   - Migration planning mode is required; operations should be safe on repeated runs (or clearly detect already-applied state).

## Steps

1. Define migration interface and registry contract in `internal/migrations/`:
   - file layout and naming convention
   - migration metadata (`id`, `from`, `to`, `description`)
   - `Up()` and `Down()` signatures
2. Define config version contract in notebook config schemas:
   - `config_version` field behavior
   - bootstrap/default version behavior for existing notebooks
3. Define migration planner/executor behavior:
   - build migration path from current to target
   - support both upward and downward traversal
   - enforce chain validation before execution
4. Define CLI behavior for `jot notebook migrate`:
   - default: migrate to latest
   - optional explicit target version
   - `--dry-run`, `--json`, and controlled `--force` semantics
5. Define safety and recoverability model:
   - preflight checks
   - transactional boundaries for file operations where possible
   - rollback/down strategy expectations
6. Produce implementation-ready breakdown for next session:
   - concrete stories/tasks
   - test matrix (unit + integration + failure/rollback paths)

## Expected Outcome

A reviewed and approved planning artifact describing:
- migration architecture boundaries,
- CLI UX and safety guarantees,
- and a clear next-session implementation path.

## Actual Outcome

Initial vision drafted and queued for user review/edits before implementation session.

## Lessons Learned

- The rename epic already needs migration guidance; converting this into a versioned framework now reduces future one-off migration debt.
- Keeping migration logic in services (not command handlers) aligns with the project’s thin-command/fat-service architecture.
