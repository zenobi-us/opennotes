---
id: 5a6b7c8d
title: Workflow integration points and persistence risks
epic_id: b2f4e6a8
related_task_id: 1d2e3f4a
created_at: 2026-02-28T17:40:00+10:30
updated_at: 2026-02-28T17:40:00+10:30
status: completed
---

# Workflow integration points and persistence risks

## Research Questions
1. Where should workflow logic live in Jot architecture?
2. Which existing command/service boundaries are reusable for workflow features?
3. What config/migration risks block safe rollout?

## Summary
Workflow logic should live in `internal/services` with thin command orchestration in `cmd/`. Existing `NotebookService` is the canonical notebook config read/write boundary and was dropping unknown config fields during save operations. That made future `workflows` definitions lossy. A minimal vertical slice was implemented to preserve and expose `workflows` metadata through notebook config load/save and notebook info JSON output.

## Findings
- **Command boundary (thin):**
  - `cmd/root.go` initializes services only.
  - `cmd/notebook_info.go` delegates formatting/rendering only.
  - `cmd/output_format.go` is current payload boundary for machine-readable notebook info.
- **Service boundary (fat):**
  - `internal/services/notebook.go` owns `.jot.json` load/save semantics.
  - `StoredNotebookConfig` is the source of truth for preserved notebook config fields.
- **Migration alignment:**
  - `internal/services/migration.go` already provides versioned migration framework; workflow schema evolution should hook here, not ad-hoc command transforms.
- **Risk identified:**
  - Before this slice, unknown keys in `.jot.json` could be silently dropped on `SaveConfig()`.
  - This would have destroyed workflow definitions once written.
- **Mitigation implemented in slice:**
  - Added `workflows` field to `StoredNotebookConfig` and preservation through load/save.
  - Added payload exposure in notebook info JSON for discovery/debugging.

## References
- `cmd/root.go`
- `cmd/notebook_info.go`
- `cmd/output_format.go`
- `internal/services/notebook.go`
- `internal/services/migration.go`
- codemapper outputs (`cm stats`, `cm map`, `cm query`) from this session
