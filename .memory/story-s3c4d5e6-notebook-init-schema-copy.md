---
id: s3c4d5e6
title: Notebook Init Includes Schema Reference
created_at: 2026-03-01T01:45:00+10:30
updated_at: 2026-03-01T01:45:00+10:30
status: proposed
epic_id: 7c631839
phase_id: null
priority: medium
story_points: 2
test_coverage: none
---

# Notebook Init Includes Schema Reference

## User Story

As a **user**, I want `jot init` to set up my `.jot.json` with a proper `$schema` reference so that my IDE provides autocomplete and validation immediately.

## Acceptance Criteria

- [ ] AC1: `jot init` creates `.jot.json` with `$schema` field pointing to the schema
- [ ] AC2: Schema reference works for IDE validation (VSCode, IntelliJ)
- [ ] AC3: User can choose between local schema file or remote URL (flag or prompt)
- [ ] AC4: If local schema chosen, schema file is copied to notebook directory
- [ ] AC5: Documentation updated to explain schema usage

## Context

Currently `.jot.json` files created by `jot init` don't include a `$schema` reference, so users miss out on IDE autocomplete. The original plan was to "update notebook create command to generate schema" but this was never implemented.

Two options for schema reference:
1. **Local file**: Copy `jot.schema.json` to notebook directory, reference as `"./$jot.schema.json"`
2. **Remote URL**: Reference `https://jot.dev/schema/v1/notebook.json` (requires hosting)

For now, local file is simpler and doesn't require infrastructure.

## Out of Scope

- Hosting schema at jot.dev (future infrastructure work)
- Auto-updating local schema files

## Tasks

_To be populated during task breakdown_

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| AC1 | Init creates $schema field | `init_test.go` / `TestInitCreatesSchemaReference` | ❌ |
| AC4 | Local schema copied | `init_test.go` / `TestInitCopiesSchemaFile` | ❌ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are created_

## Notes

- Consider embedding schema in binary via `go:embed` for portability
- If embedded, `jot init` extracts schema to notebook directory
- VSCode JSON Schema association also possible via `.vscode/settings.json`
