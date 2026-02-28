---
id: s4d5e6f7
title: Notebook Config Validation Command
created_at: 2026-03-01T01:45:00+10:30
updated_at: 2026-03-01T01:45:00+10:30
status: proposed
epic_id: 7c631839
phase_id: null
priority: low
story_points: 3
test_coverage: none
---

# Notebook Config Validation Command

## User Story

As a **user**, I want a `jot notebook validate` command so that I can check my configuration for errors before they cause problems at runtime.

## Acceptance Criteria

- [ ] AC1: `jot notebook validate` checks current notebook's `.jot.json` against schema
- [ ] AC2: Valid config produces exit code 0 and success message
- [ ] AC3: Invalid config produces exit code 1 and lists all validation errors
- [ ] AC4: Validation errors include field path, expected type, and actual value
- [ ] AC5: Supports `--json` flag for machine-readable error output
- [ ] AC6: Works with `--notebook` flag to validate specific notebook

## Context

Currently, invalid `.jot.json` configs are only detected at runtime when a command fails. Explicit validation helps users debug configuration issues proactively, especially when hand-editing configs or migrating between versions.

This complements IDE validation by providing CLI-based verification for:
- CI pipelines
- Pre-commit hooks
- Users without IDE schema support

## Out of Scope

- Automatic config migration/fixing
- Validation of note content (only config validation)

## Tasks

_To be populated during task breakdown_

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| AC1 | Validates current notebook | `notebook_validate_test.go` / `TestValidateCurrentNotebook` | ❌ |
| AC2 | Exit 0 on valid | `notebook_validate_test.go` / `TestValidateExitCodeSuccess` | ❌ |
| AC3 | Exit 1 on invalid | `notebook_validate_test.go` / `TestValidateExitCodeFailure` | ❌ |
| AC5 | JSON output | `notebook_validate_test.go` / `TestValidateJSONOutput` | ❌ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are created_

## Notes

- Use `github.com/xeipuuv/gojsonschema` or similar for validation
- Consider loading schema from embedded resource vs external file
- Error format should match workflow validation (`WF_*` error codes pattern)
- May want `--fix` flag in future to auto-correct simple issues (out of scope now)
