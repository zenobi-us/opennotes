---
id: s2b3c4d5
title: CI Schema Drift Detection
created_at: 2026-03-01T01:45:00+10:30
updated_at: 2026-03-01T01:45:00+10:30
status: proposed
epic_id: 7c631839
phase_id: null
priority: high
story_points: 3
test_coverage: none
---

# CI Schema Drift Detection

## User Story

As a **maintainer**, I want CI to fail if the committed schema doesn't match what would be generated so that schema drift is caught before merge.

## Acceptance Criteria

- [ ] AC1: GitHub Actions workflow includes a schema verification step
- [ ] AC2: CI runs `mise run schema:generate` and compares output to committed `jot.schema.json`
- [ ] AC3: CI fails with clear error message if generated schema differs from committed
- [ ] AC4: Error message shows the diff between expected and actual
- [ ] AC5: CI passes when schema is up-to-date

## Context

Even with automated generation, developers might forget to regenerate the schema after modifying Go structs. CI enforcement ensures the committed schema always matches the source of truth.

This follows the same pattern as `go generate` checks in other Go projects — generate, diff, fail if different.

## Out of Scope

- Automatic schema commits by CI (we want human review)
- Schema publishing to external URL

## Tasks

_To be populated during task breakdown_

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| AC2 | CI generates and compares | Manual CI verification | ❌ |
| AC3 | CI fails on drift | Manual CI verification | ❌ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are created_

## Notes

- Pattern: `mise run schema:generate && git diff --exit-code jot.schema.json`
- Should be added to existing CI workflow, not a separate workflow
- Consider using `--check` flag pattern: `mise run schema:check` that fails on diff
