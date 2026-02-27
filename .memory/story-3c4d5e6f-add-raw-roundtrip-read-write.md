---
id: 3c4d5e6f
title: Add raw roundtrip read write
epic_id: a7c3d9e1
created_at: 2026-02-27T08:01:43+10:30
updated_at: 2026-02-28T01:26:00+10:30
status: completed
priority: high
story_points: 3
---

# Add raw roundtrip read write

## User Story
As an agent, I want lossless raw read/write for notes so that automation can preserve exact markdown and frontmatter formatting.

## Acceptance Criteria
- [x] `jot notes get <path> --raw` outputs exact file bytes as text.
- [x] Paired with `update` (alias `put`), roundtrip does not alter content unless requested.
- [x] Existing parsed output mode remains available.

## Context
Parsed read is useful for humans but raw mode is needed for deterministic automation.

## Out of Scope
Automatic markdown normalization.

## Use Stories
1. As an agent, I read a note in raw mode, transform it, and write it back without losing frontmatter/layout fidelity.
2. As a script, I need exact markdown bytes for downstream processors (not parsed/rendered output).
3. As a human user, I keep default `notes get` behavior for readable output while automation uses `--raw`.

## Tasks
- [task-a3c4d5e6](task-a3c4d5e6-implement-notes-get-raw-mode.md) — Implement notes get raw mode

## Notes
- Raw mode should be explicit and safe for scripts.
- @TODO(epic:c5d7e9b1): Evaluate whether group context should influence raw/parsed read helpers.
