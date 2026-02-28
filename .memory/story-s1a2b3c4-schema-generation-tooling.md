---
id: s1a2b3c4
title: Automated Schema Generation from Go Structs
created_at: 2026-03-01T01:45:00+10:30
updated_at: 2026-03-01T01:45:00+10:30
status: proposed
epic_id: 7c631839
phase_id: null
priority: high
story_points: 5
test_coverage: none
---

# Automated Schema Generation from Go Structs

## User Story

As a **maintainer**, I want the JSON schema to be automatically generated from Go struct definitions so that I don't have to manually keep the schema in sync with code changes.

## Acceptance Criteria

- [ ] AC1: A mise task `mise run schema:generate` generates `jot.schema.json` from Go structs
- [ ] AC2: Generated schema includes all fields from `StoredNotebookConfig`, `NotebookGroup`, and `NotebookGroupWorkflow`
- [ ] AC3: Generated schema preserves existing features: `$schema`, `$id`, `title`, `description`
- [ ] AC4: Generated schema includes `examples` for key fields (root, name, globs)
- [ ] AC5: Generated schema includes `pattern` validation for glob patterns
- [ ] AC6: Running the generator twice produces identical output (deterministic)
- [ ] AC7: Generated schema is valid JSON Schema Draft-07

## Context

The current `jot.schema.json` was hand-crafted and has already drifted from the actual Go structs. The workflows epic added `NotebookGroupWorkflow` which is not reflected in the schema. Manual maintenance is error-prone and burdensome.

Go JSON schema libraries like `github.com/invopop/jsonschema` can reflect on struct tags and generate compliant schemas. We need to evaluate options and select one that supports:
- JSON struct tags
- Custom descriptions via `jsonschema` tags
- Examples and patterns
- Handling of `map[string]any` types

## Out of Scope

- Runtime validation (separate story)
- CI integration (separate story)
- Embedding schema in binary (future consideration)

## Tasks

_To be populated during task breakdown_

## Test Specification

### E2E Tests

| AC# | Criterion | Test file/case | Status |
|-----|-----------|----------------|--------|
| AC1 | mise task generates schema | `schema_generation_test.go` / `TestSchemaGenerationTask` | ❌ |
| AC6 | Deterministic output | `schema_generation_test.go` / `TestSchemaGenerationDeterministic` | ❌ |
| AC7 | Valid JSON Schema Draft-07 | `schema_generation_test.go` / `TestSchemaValidDraft07` | ❌ |

### Unit Test Coverage (via Tasks)

_To be populated as tasks are created_

## Notes

- Consider `go generate` directive vs mise task — mise task is preferred for consistency
- May need custom reflector configuration to handle `map[string]any` as `additionalProperties: true`
- Existing schema has pattern `^.*\\.md$|^\\*\\*/\\*\\.md$` for globs — need to preserve or improve
