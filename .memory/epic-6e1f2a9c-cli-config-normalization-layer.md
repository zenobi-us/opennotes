---
id: 6e1f2a9c
title: CLI Config Normalization Layer
created_at: 2026-02-27T08:28:00+10:30
updated_at: 2026-02-27T08:28:00+10:30
status: proposed
---

# CLI Config Normalization Layer

## Vision/Goal
Introduce a unified, deterministic configuration precedence model for Jot CLI options so env vars, flags, config file values, and defaults behave consistently across commands.

## Success Criteria
- Jot defines and documents one precedence contract for command options (for example: flag > env > config > default).
- Shared resolver layer is used by commands instead of ad-hoc per-command/env handling.
- `JOT_*` env support is explicit and consistent for supported options (including output format where relevant).
- Existing command behavior remains backward compatible unless explicitly versioned/deprecated.

## Phases
- Discovery: inventory current option resolution paths and conflicts.
- Contract: define precedence matrix and option ownership.
- Implementation: add shared resolver and migrate commands incrementally.
- Validation: regression tests for precedence and backward compatibility.

## Dependencies
- Cobra command/flag structure in `cmd/`.
- Existing config loading via `internal/services/config.go`.
- Agreement on supported global vs command-local options.

## Notes
This is intentionally parked as a future epic idea and not in the current execution slice.
