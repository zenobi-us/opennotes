---
id: 9c0d1e2f
title: Remove legacy query command in next minor
type: "task"
created_at: 2026-02-23T23:15:45+10:30
updated_at: 2026-02-23T23:15:45+10:30
status: todo
epic_id: 9b7c2a4d
plan_id: c4e8a1f2
assigned_to: unassigned
---

# Remove legacy query command in next minor

## Objective
Delete legacy `jot notes search query` implementation after deprecation window closes.

## Steps
1. Remove command registration and implementation files.
2. Update tests/docs/changelog to reflect removal.
3. Validate migration instructions remain for one release cycle.

## Exit Criteria
- Legacy command no longer appears in CLI help.
- Unified DSL path is the only supported search surface.
- Release notes include breaking-change migration example.
